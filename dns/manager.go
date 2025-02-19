package dns

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"strings"
	"time"
)

const subnet = "10.0.0."

var (
	ErrNoAvailableIPs = errors.New("no available IPs")
	ErrSubdomainInUse = errors.New("subdomain already in-use")
)

type Manager struct {
	allocatedIPs map[string]*time.Time
	subdomains   map[string]string
	domain       string
	logger       *slog.Logger
}

func New(domain string) Manager {
	return Manager{
		allocatedIPs: map[string]*time.Time{},
		subdomains:   map[string]string{},
		domain:       domain,
		logger:       slog.Default(),
	}
}

func (m Manager) GetIP(ctx context.Context, subdomain string) (string, error) {
	ip, exists := m.subdomains[subdomain]
	if exists {
		removedAt := m.allocatedIPs[ip]
		if removedAt == nil {
			return "", ErrSubdomainInUse
		}

		m.allocateIP(ctx, ip, subdomain)
		return ip, nil
	}

	iface, err := net.InterfaceByName("lo0")
	if err != nil {
		return "", err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}

	unallocatedIPs := map[string]*time.Time{}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() || ipNet.IP.To4() == nil {
			continue
		}

		ip := ipNet.IP.String()
		if !strings.HasPrefix(ip, subnet) {
			continue
		}

		removedAt, inUse := m.allocatedIPs[ip]
		if inUse {
			// add to unallocatedIPs if allocation is closed so it can be used as backup
			if removedAt != nil {
				unallocatedIPs[ip] = removedAt
			}
			continue
		}

		m.allocateIP(ctx, ip, subdomain)
		return ip, nil
	}

	// if all unallocated IPs are exhausted, use the oldest removed IP
	resultIP := findOldestDeallocatedIP(unallocatedIPs)
	if resultIP != "" {
		m.allocateIP(ctx, resultIP, subdomain)
		return resultIP, nil
	}

	return "", ErrNoAvailableIPs
}

func findOldestDeallocatedIP(unallocatedIPs map[string]*time.Time) string {
	resultIP := ""
	for ip, removedAt := range unallocatedIPs {
		if resultIP == "" {
			resultIP = ip
			continue
		}
		if removedAt.Before(*unallocatedIPs[resultIP]) {
			resultIP = ip
		}
	}

	return resultIP
}

func (m Manager) allocateIP(ctx context.Context, ip, subdomain string) {
	m.allocatedIPs[ip] = nil
	m.subdomains[subdomain] = ip
	go m.removeIP(ctx, ip)

	m.logger.Debug("allocated IP", "ip", ip, "subdomain", subdomain)
}

func (m Manager) removeIP(ctx context.Context, ip string) {
	<-ctx.Done()
	now := time.Now()
	m.allocatedIPs[ip] = &now

	m.logger.Debug("removed IP", "ip", ip)
}
