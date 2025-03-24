package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"sync"

	"github.com/calvinmclean/goblin/dns"
)

// Server runs the backend DNS server and IP allocation server
type Server struct {
	mgr    dns.Manager
	server *http.Server
	logger *slog.Logger
}

func New(mgr dns.Manager, addr string) Server {
	return Server{
		mgr,
		&http.Server{
			Addr: addr,
		},
		slog.Default(),
	}
}

func (s Server) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		<-ctx.Done()
		_ = s.server.Shutdown(context.Background())
		wg.Done()
	}()

	go func() {
		err := s.RunHTTP(ctx)
		if err != nil {
			log.Fatalf("failed to serve HTTP: %v", err)
		}
		wg.Done()
	}()

	go func() {
		err := s.mgr.RunDNS(ctx)
		if err != nil {
			log.Fatalf("failed to serve DNS: %v", err)
		}
		wg.Done()
	}()

	wg.Wait()

	return nil
}

func (s Server) RunHTTP(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /allocate/{subdomain}", s.allocateIPHandler)
	mux.HandleFunc("POST /register/{subdomain}", s.registerFallbackHandler)
	s.server.Handler = mux

	s.logger.Info("started local HTTP server", "addr", s.server.Addr)

	return s.server.ListenAndServe()
}

func (s Server) registerFallbackHandler(w http.ResponseWriter, r *http.Request) {
	err := s.registerFallback(w, r)
	if err != nil {
		s.logger.Error("error registering fallback", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s Server) registerFallback(w http.ResponseWriter, r *http.Request) error {
	subdomain := r.PathValue("subdomain")
	if subdomain == "" {
		return errors.New("missing required subdomain path variable")
	}

	address := r.URL.Query().Get("address")
	if address == "" {
		return errors.New("missing address")
	}

	s.mgr.RegisterFallback(subdomain, address)

	w.WriteHeader(http.StatusAccepted)
	return nil
}

func (s Server) allocateIPHandler(w http.ResponseWriter, r *http.Request) {
	err := s.allocateIP(w, r)
	if err != nil {
		s.logger.Error("error allocating IP", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s Server) allocateIP(w http.ResponseWriter, r *http.Request) error {
	subdomain := r.PathValue("subdomain")
	if subdomain == "" {
		return errors.New("missing required subdomain path variable")
	}

	ip, err := s.mgr.GetIP(r.Context(), subdomain)
	if err != nil {
		return fmt.Errorf("error getting IP: %w", err)
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		return errors.New("flush unsupported")
	}

	fmt.Fprintln(w, ip)
	flusher.Flush()

	// keep open until the context closes so the IP is still allocated
	<-r.Context().Done()
	return nil
}
