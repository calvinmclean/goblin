package cmd

import (
	"dns-plugin-thing/dns"
	"dns-plugin-thing/server"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/urfave/cli/v2"
)

var ExampleCmd = &cli.Command{
	Name:        "example",
	Description: "run example",
	Action:      runExample,
}

func runExample(c *cli.Context) error {
	dnsMgr := dns.New(".gotest.", slog.LevelDebug)

	go func() {
		err := runPlugin(c.Context, dnsMgr, "./plugins/examples/hello-world.so", "helloworld", 0)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		time.Sleep(5 * time.Second)

		err := runPlugin(c.Context, dnsMgr, "./plugins/examples/howdy-world.so", "howdy", 0)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		time.Sleep(15 * time.Second)

		err := runPlugin(c.Context, dnsMgr, "./plugins/examples/howdy-world.so", "howdynew", 0)
		if err != nil {
			panic(err)
		}
	}()

	server := server.New(dnsMgr)
	go func() {
		err := server.Run(c.Context, grpcServerAddr)
		if err != nil {
			log.Fatalf("error running GRPC server: %v", err)
		}
	}()

	err := dnsMgr.RunDNS(c.Context, dnsServerAddr)
	if err != nil {
		return fmt.Errorf("error running DNS server: %w", err)
	}
	return nil
}
