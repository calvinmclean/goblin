package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
)

const greeting = "Howdy"

func Run(ctx context.Context, ip string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s, World!\n", greeting)
		log.Printf("%s, World!", greeting)
	})

	addr := fmt.Sprintf("%s:8080", ip)
	log.Printf("starting server on http://%s", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		<-ctx.Done()
		err := server.Shutdown(context.Background())
		if err != nil {
			log.Fatalf("failed to stop server: %v", err)
		}
	}()

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to run server: %v", err)
	}

	wg.Wait()

	return nil
}
