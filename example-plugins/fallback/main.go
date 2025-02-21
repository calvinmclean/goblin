package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
)

func main() {
	err := Run(context.Background(), "127.0.0.1")
	if err != nil {
		log.Fatal(err)
	}
}

func Run(ctx context.Context, ip string) error {
	addr := fmt.Sprintf("%s:8081", ip)
	log.Printf("starting server on http://%s", addr)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "I am running at %s\n", addr)
	})

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
