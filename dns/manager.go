package dns

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"
)

const (
	defaultSubnet   = "10.0.0.0/8"
	resolverFileFmt = `nameserver %s
port %s`
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

	addr   string
	domain string
	subnet *net.IPNet

	logger *slog.Logger
}

func New(domain, dnsAddr string) (Manager, error) {
	_, subnet, err := net.ParseCIDR(defaultSubnet)
	if err != nil {
		return Manager{}, fmt.Errorf("error parsing subnet: %w", err)
	}

	manager := Manager{
		allocatedIPs: map[string]*record{},
		subdomains:   map[string]*record{},
		addr:         dnsAddr,
		subnet:       subnet,
		domain:       domain,
		logger:       slog.Default(),
	}

	err = checkResolverFile(domain, dnsAddr)
	if err != nil {
		return Manager{}, err
	}

	numIPs, err := manager.checkIPAliases()
	if err != nil {
		return Manager{}, err
	}

	manager.logger.Info("found IP aliases", "count", numIPs)

	return manager, nil
}

// ensure IP aliases exist in the system
func (m Manager) checkIPAliases() (int, error) {
	ipIter, err := m.getIPs()
	if err != nil {
		return 0, err
	}

	count := 0
	for range ipIter {
		count++
	}

	if count == 0 {
		return 0, NewUserFixableError(errors.New("no IP aliases configured"), ipAliasInstruction)
	}

	return count, nil
}

// ensure correct resolver config exists on the system
func checkResolverFile(domain, dnsAddr string) error {
	addrParts := strings.Split(dnsAddr, ":")
	if len(addrParts) != 2 {
		return errors.New("unexpected format for address")
	}

	addr := addrParts[0]
	port := addrParts[1]
	if addr == "" {
		addr = "127.0.0.1"
	}

	expected := fmt.Sprintf(resolverFileFmt, addr, port)

	fname := fmt.Sprintf("/etc/resolver/%s", domain)
	contents, err := os.ReadFile(fname)
	if err != nil {
		return NewUserFixableError(
			fmt.Errorf("error reading resolver file: %w", err),
			resolverFileInstructions(fname, expected),
		)
	}

	if strings.TrimSpace(string(contents)) != expected {
		return NewUserFixableError(
			errors.New("unexpected contents of resolver file"),
			resolverFileInstructions(fname, expected),
		)
	}

	return nil
}

// getIPs iterates through IPs in the subnet
func (m Manager) getIPs() (iter.Seq[net.IP], error) {
	iface, err := net.InterfaceByName("lo0")
	if err != nil {
		return nil, err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	return func(yield func(net.IP) bool) {
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.IsLoopback() || ipNet.IP.To4() == nil {
				continue
			}

			if !m.subnet.Contains(ipNet.IP) {
				continue
			}

			if !yield(ipNet.IP) {
				return
			}
		}
	}, nil
}

// GetIP allocates and returns an IP address. It will keep it open until the context is closed
func (m Manager) GetIP(ctx context.Context, subdomain string) (string, error) {
	rec := m.subdomains[subdomain]
	if rec != nil {
		if rec.removedAt == nil {
			return "", ErrSubdomainInUse
		}

		m.allocateIP(ctx, rec.ip, subdomain)
		return rec.ip, nil
	}

	ipIter, err := m.getIPs()
	if err != nil {
		return "", fmt.Errorf("error getting IPs from system: %w", err)
	}

	unallocatedIPs := []*record{}
	for ip := range ipIter {
		ipStr := ip.String()

		rec := m.allocatedIPs[ipStr]
		if rec != nil {
			// add to unallocatedIPs if allocation is closed so it can be used as backup
			if rec.removedAt != nil {
				unallocatedIPs = append(unallocatedIPs, rec)
			}
			continue
		}

		m.allocateIP(ctx, ipStr, subdomain)
		return ipStr, nil
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

	if resultIP == nil {
		return ""
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
