package dns

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/miekg/dns"
)

func (m Manager) RunDNS(ctx context.Context) error {
	dns.HandleFunc(".", m.handleDNSRequest)
	server := &dns.Server{
		Addr: m.dnsAddr,
		Net:  "udp",
	}
	m.logger.Info("starting local DNS server", "addr", server.Addr)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		<-ctx.Done()
		err := server.Shutdown()
		if err != nil {
			log.Fatalf("failed to stop DNS server: %v", err)
		}
	}()

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("failed to run DNS server: %v", err)
	}

	wg.Wait()

	return nil
}

func (m Manager) handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(r)

	m.logger.Debug("received DNS request")

	for _, q := range r.Question {
		if !strings.HasSuffix(q.Name, m.domain) {
			continue
		}

		subdomain := getSubdomain(q.Name)
		rec, ok := m.subdomains[subdomain]
		if !ok || rec.removedAt != nil {
			continue
		}
		m.logger.Debug("found ip for subdomain", "subdomain", subdomain, "ip", rec.ip)

		rr, _ := dns.NewRR(fmt.Sprintf("%s IN A %s", q.Name, rec.ip))
		rr.Header().Ttl = 0
		msg.Answer = append(msg.Answer, rr)
	}

	w.WriteMsg(msg)
}

func getSubdomain(d string) string {
	parts := strings.Split(d, ".")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}
