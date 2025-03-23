package dns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
)

func (m Manager) RunDNS(ctx context.Context) error {
	conn, err := net.ListenPacket("udp", m.Address)
	if err != nil {
		return fmt.Errorf("failed to create UDP listener: %w", err)
	}

	m.logger.Info("started local DNS server", "addr", m.Address)

	buffer := make([]byte, 512)
	for {
		select {
		case <-ctx.Done():
			return conn.Close()
		default:
		}

		n, clientAddr, err := conn.ReadFrom(buffer)
		if err != nil {
			m.logger.Error("error reading buffer", "error", err)
			continue
		}

		err = m.handleDNSRequest(conn, clientAddr, buffer[:n])
		if err != nil {
			m.logger.Error("error handling DNS request", "error", err)
			continue
		}
	}
}

func getSubdomain(d string) string {
	parts := strings.Split(d, ".")
	if len(parts) == 0 {
		return ""
	}
	// Sometimes the web-browser throws a 'www.' in front of the domain
	if parts[0] == "www" {
		if len(parts) >= 2 {
			return parts[1]
		}
	}
	return parts[0]
}

func (m Manager) handleDNSRequest(conn net.PacketConn, clientAddr net.Addr, request []byte) error {
	if len(request) < 12 {
		return fmt.Errorf("invalid length for request: %d", len(request))
	}

	query := request[12:]
	domain := parseDomain(query)

	m.logger.Info("received DNS request", "domain", domain)

	if !strings.HasSuffix(domain, m.Domain) {
		// Ignore this DNS domain
		if domain == "_dns.resolver.arpa" {
			return nil
		}
		return fmt.Errorf("unexpected domain: %s", domain)
	}

	subdomain := getSubdomain(domain)

	rec, ok := m.subdomains[subdomain]
	if !ok || !rec.isActive() {
		// if a domain is not registered or is registered but un-allocated, check for fallback routes
		m.logger.Debug("checking for fallback routes")
		var err error
		rec, err = m.handleFallbackRoutes(subdomain)
		if err != nil {
			return fmt.Errorf("error handling fallback routes: %w", err)
		}
		if rec == nil {
			return fmt.Errorf("no record found for subdomain %q", subdomain)
		}
	}
	m.logger.Info("responding with ip", "subdomain", subdomain, "ip", rec.ip.String())

	response := m.createDNSResponse(request, rec.ip)
	if response == nil {
		return errors.New("unexpected empty response")
	}

	_, err := conn.WriteTo(response, clientAddr)
	if err != nil {
		return fmt.Errorf("error writing response: %w", err)
	}

	return nil
}

func (m Manager) createDNSResponse(request []byte, ip net.IP) []byte {
	// Create a DNS response based on the request
	response := make([]byte, len(request)+16)
	copy(response, request)

	// Set response flags: QR (response), RD, RA
	response[2] = 0x81
	response[3] = 0x80

	// Set Answer count to 1
	response[6] = 0x00
	response[7] = 0x01

	// Copy question section to response
	questionEnd := len(request)
	copy(response[12:], request[12:questionEnd])

	// Add Answer section
	offset := questionEnd
	response[offset] = 0xc0 // Pointer to domain name
	response[offset+1] = 0x0c
	response[offset+2] = 0x00 // Type A (IPv4)
	response[offset+3] = 0x01
	response[offset+4] = 0x00 // Class IN
	response[offset+5] = 0x01
	response[offset+6] = 0x00 // TTL
	response[offset+7] = 0x00
	response[offset+8] = 0x00
	response[offset+9] = 0x00  // 0 seconds TTL
	response[offset+10] = 0x00 // Data length
	response[offset+11] = 0x04
	response[offset+12] = ip[0]
	response[offset+13] = ip[1]
	response[offset+14] = ip[2]
	response[offset+15] = ip[3]

	return response
}

func ipToBytes(ip string) net.IP {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil
	}

	ipBytes := parsedIP.To4()
	if ipBytes == nil {
		return nil
	}

	return ipBytes
}

func parseDomain(query []byte) string {
	var domainParts []string
	i := 0
	for query[i] != 0 {
		length := query[i]
		i++
		part := string(query[i : i+int(length)])
		domainParts = append(domainParts, part)
		i += int(length)
	}
	return strings.Join(domainParts, ".")
}
