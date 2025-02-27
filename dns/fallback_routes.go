package dns

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// FallbackRoutes maps a subdomain to the FallbackRoute that should be used
// when the local subdomain is not running on the server
type FallbackRoutes map[string]FallbackRoute

// FallbackRoute configures the remote Hostname and Ports that should be proxied
// in the case that the local subdomain is not available. A reverse proxy will run
// for each provided Port
type FallbackRoute struct {
	Hostname string
	Ports    []int
}

func createReverseProxy(hostname string, port int, ip net.IP) (*http.Server, error) {
	target, err := url.Parse(fmt.Sprintf("%s:%d", hostname, port))
	if err != nil {
		return nil, fmt.Errorf("error parsing target URL: %w", err)
	}
	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(target)
			r.Out.Host = target.Host
		},
	}

	return &http.Server{
		Addr:    fmt.Sprintf("%s:%d", ip.String(), port),
		Handler: proxy,
	}, nil
}

func stopServers(servers []*http.Server) error {
	errs := []error{}
	for _, server := range servers {
		err := server.Shutdown(context.Background())
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 1 {
		return errs[0]
	}
	if len(errs) > 1 {
		return fmt.Errorf("%d errors occurred: %v", len(errs), errs)
	}

	return nil
}

func (m Manager) handleFallbackRoutes(subdomain string) (*record, error) {
	if m.FallbackRoutes == nil {
		return nil, nil
	}

	fallback, ok := m.FallbackRoutes[subdomain]
	if !ok {
		return nil, nil
	}

	logger := m.logger.With(
		"subdomain", subdomain,
		"hostname", fallback.Hostname,
		"port", fallback.Ports,
	)

	logger.Info("found fallback proxy")

	rec, err := m.findOrCreateRecord(subdomain)
	if err != nil {
		return nil, fmt.Errorf("error finding IP: %w", err)
	}
	logger = logger.With("ip", rec.ip)

	// if reverse proxy is running, don't re-run it
	if rec.stop != nil {
		return rec, nil
	}

	servers := make([]*http.Server, len(fallback.Ports))
	for i, port := range fallback.Ports {
		server, err := createReverseProxy(fallback.Hostname, port, rec.ip)
		if err != nil {
			return nil, fmt.Errorf("error parsing target URL: %w", err)
		}
		servers[i] = server

		go func() {
			err = servers[i].ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				logger.Info("error running fallback proxy", "error", err)
			}
		}()
	}

	ctx, cancel := context.WithCancel(context.Background())
	rec.stop = func() error {
		cancel()
		return stopServers(servers)
	}

	m.allocateIPRecord(ctx, rec)

	m.logger.Info("started fallback proxy")

	return rec, nil
}
