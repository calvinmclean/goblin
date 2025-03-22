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
	fallback, ok := m.FallbackRoutes[subdomain]
	if !ok {
		return nil, nil
	}

	logger := m.logger.With(
		"subdomain", subdomain,
		"fallback", fallback,
	)

	logger.Debug("found fallback configuration")

	rec := &record{
		subdomain: "Remote Address (no subdomain)",
	}

	fallbackIP, ok := asIP(fallback)
	if ok {
		rec.ip = fallbackIP
		return rec, nil
	}

	var err error
	rec.ip, err = lookupIP(fallback)
	if err != nil {
		return nil, fmt.Errorf("error finding IP for remote address: %w", err)
	}

	m.logger.With("remote_ip", rec.ip.String()).Debug("found IP address for fallback route")

	return rec, nil
}

func asIP(fallback string) (net.IP, bool) {
	ip := net.ParseIP(fallback)
	if ip == nil {
		return nil, false
	}
	return ip.To4(), true
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

// RegisterFallback allows registering a fallback domain that will be used if a Goblin plugin is not running
func (m Manager) RegisterFallback(subdomain, address string) {
	m.FallbackRoutes[subdomain] = address
}
