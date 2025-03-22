package dns

import (
	"errors"
	"fmt"
	"net"
)

// FallbackRoutes maps a subdomain to an actual domain name that should be used
// when the local subdomain is not running on the server
type FallbackRoutes map[string]string

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
		"hostname", fallback,
	)

	logger.Debug("found fallback configuration")

	remoteIP, err := lookupIP(fallback)
	if err != nil {
		return nil, fmt.Errorf("error finding IP for remote address: %w", err)
	}

	m.logger.With("remote_ip", remoteIP.String()).Debug("found IP address for fallback route")

	return &record{
		ip:        remoteIP,
		subdomain: "Remote Address (no subdomain)",
	}, nil
}

func lookupIP(domain string) (net.IP, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, err
	}

	for _, ip := range ips {
		resultIP := ip.To4()
		if resultIP != nil {
			return resultIP, nil
		}
	}

	return nil, errors.New("no ip found for domain")
}
