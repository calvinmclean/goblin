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
	ip        net.IP
	subdomain string
	removedAt *time.Time
}

func (r *record) isActive() bool {
	return r.removedAt == nil
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
			if !ok || ipNet.IP.IsLoopback() || ipNet.IP.To4() == nil || ipNet.IP.IsUnspecified() {
				continue
			}

			if !m.subnet.Contains(ipNet.IP) {
				continue
			}

			if !yield(ipNet.IP.To4()) {
				return
			}
		}
	}, nil
}

func (m Manager) getExistingRecord(subdomain string) (*record, error) {
	rec := m.subdomains[subdomain]
	if rec == nil {
		return nil, nil
	}

	if rec.isActive() {
		return nil, ErrSubdomainInUse
	}

	return rec, nil
}

func (m Manager) getNextAvailableIP() (net.IP, []net.IP, error) {
	ipIter, err := m.getIPs()
	if err != nil {
		return nil, nil, fmt.Errorf("error getting IPs from system: %w", err)
	}

	unallocatedIPs := []net.IP{}
	for ip := range ipIter {
		rec := m.allocatedIPs[ip.String()]
		// IP is not currently in-use so it can be used
		if rec == nil {
			return ip, nil, nil
		}

		// add to unallocatedRECs if allocation is closed so it can be used as backup
		if !rec.isActive() {
			// if !rec.isActive() {
			unallocatedIPs = append(unallocatedIPs, rec.ip)
		}
	}

	return nil, unallocatedIPs, nil
}

func (m Manager) findOrCreateRecord(subdomain string) (*record, error) {
	rec, err := m.getExistingRecord(subdomain)
	if err != nil {
		return nil, ErrSubdomainInUse
	}
	if rec != nil {
		return rec, nil
	}

	ip, unallocatedIPs, err := m.getNextAvailableIP()
	if err != nil {
		return nil, err
	}
	if ip != nil {
		return &record{ip, subdomain, nil}, nil
	}

	// if all unallocated IPs are exhausted, use the oldest removed IP
	rec = m.findOldestDeallocatedIP(unallocatedIPs)
	if rec != nil {
		return rec, nil
	}

	return nil, ErrNoAvailableIPs
}

// GetIP allocates and returns an IP address. It will keep it open until the context is closed
func (m Manager) GetIP(ctx context.Context, subdomain string) (string, error) {
	rec, err := m.findOrCreateRecord(subdomain)
	if err != nil {
		return "", err
	}

	m.allocateIPRecord(ctx, rec)
	return rec.ip.String(), nil
}

// find the oldest in a list of records that were de-allocated
func (m Manager) findOldestDeallocatedIP(unallocatedIPs []net.IP) *record {
	var result *record
	for _, ip := range unallocatedIPs {
		rec := m.allocatedIPs[ip.String()]

		if result == nil {
			result = rec
			continue
		}

		if rec.removedAt.Before(*result.removedAt) {
			result = rec
		}
	}

	if result == nil {
		return nil
	}

	return result
}

func (m Manager) allocateIPRecord(ctx context.Context, rec *record) {
	m.allocatedIPs[rec.ip.String()] = rec
	m.subdomains[rec.subdomain] = rec
	go m.removeIP(ctx, rec)

	m.logger.Debug("allocated IP", "ip", rec.ip, "subdomain", rec.subdomain)
}

func (m Manager) removeIP(ctx context.Context, rec *record) {
	<-ctx.Done()
	now := time.Now()
	rec.removedAt = &now

	m.logger.Debug("removed IP", "ip", rec.ip)
}
