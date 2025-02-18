package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"dns-plugin-thing/dns"
	"dns-plugin-thing/plugins"
)

// TODO:
// - Create `cmd/main.go`
// - Add subcommand to run plugin using .so path, optional hostname
// - Add server subcommand to run the DNS server and add simple HTTP endpoints to
//   get IP. Or jump straight to gRPC?
// - How do I handle closing the IPs? I would need to do it manually or keep an open connection
//   like SSE to still use the same context to close IPs. Or I could use a polling method since
//   the server will already know IPs, but then it's a bit trickier to do with port management
// -

func main() {
	dnsMgr := dns.New(".gotest.")

	go func() {
		err := runPlugin(dnsMgr, "./plugins/examples/hello-world.so", "helloworld", 10*time.Second)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		time.Sleep(5 * time.Second)

		err := runPlugin(dnsMgr, "./plugins/examples/howdy-world.so", "howdy", 10*time.Second)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		time.Sleep(15 * time.Second)

		err := runPlugin(dnsMgr, "./plugins/examples/howdy-world.so", "howdynew", 10*time.Second)
		if err != nil {
			panic(err)
		}
	}()

	err := dnsMgr.RunDNS(context.Background(), "127.0.0.1:5154")
	if err != nil {
		panic(err)
	}
}

func runPlugin(dnsMgr dns.Manager, fname, hostname string, timeout time.Duration) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	log.Printf("starting plugin: %q", hostname)
	run, err := plugins.Load(fname)
	if err != nil {
		return fmt.Errorf("error loading plugin: %w", err)
	}

	ip, err := dnsMgr.GetIP(ctx, hostname)
	if err != nil {
		return fmt.Errorf("error getting IP: %w", err)
	}

	err = run(ctx, ip)
	log.Printf("stopped plugin: %q", hostname)
	return err
}
