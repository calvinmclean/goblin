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

type record struct {
	ip        string
	subdomain string
	removedAt *time.Time
}

type Manager struct {
	// allocatedIPs and subdomains point to the same data but with IP or Subdomain as the key
	allocatedIPs map[string]*record
	subdomains   map[string]*record

	domain string
	logger *slog.Logger
}

func New(domain string) Manager {
	return Manager{
		allocatedIPs: map[string]*record{},
		subdomains:   map[string]*record{},
		domain:       domain,
		logger:       slog.Default(),
	}
}

func (m Manager) GetIP(ctx context.Context, subdomain string) (string, error) {
	rec := m.subdomains[subdomain]
	if rec != nil {
		if rec.removedAt == nil {
			return "", ErrSubdomainInUse
		}

		m.allocateIP(ctx, rec.ip, subdomain)
		return rec.ip, nil
	}

	iface, err := net.InterfaceByName("lo0")
	if err != nil {
		return "", err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}

	unallocatedIPs := []*record{}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() || ipNet.IP.To4() == nil {
			continue
		}

		ip := ipNet.IP.String()
		if !strings.HasPrefix(ip, subnet) {
			continue
		}

		rec := m.allocatedIPs[ip]
		if rec != nil {
			// add to unallocatedIPs if allocation is closed so it can be used as backup
			if rec.removedAt != nil {
				unallocatedIPs = append(unallocatedIPs, rec)
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

func findOldestDeallocatedIP(unallocatedIPs []*record) string {
	var resultIP *record
	for i := range unallocatedIPs {
		rec := unallocatedIPs[i]

		if resultIP == nil {
			resultIP = rec
			continue
		}

		if rec.removedAt.Before(*resultIP.removedAt) {
			resultIP = rec
		}
	}

	return resultIP.ip
}

func (m Manager) allocateIP(ctx context.Context, ip, subdomain string) {
	rec := &record{ip, subdomain, nil}

	m.allocatedIPs[ip] = rec
	m.subdomains[subdomain] = rec
	go m.removeIP(ctx, rec)

	m.logger.Debug("allocated IP", "ip", ip, "subdomain", subdomain)
}

func (m Manager) removeIP(ctx context.Context, rec *record) {
	<-ctx.Done()
	now := time.Now()
	rec.removedAt = &now

	m.logger.Debug("removed IP", "ip", rec.ip)
}
