package cmd

import (
	"dns-plugin-thing/dns"
	"dns-plugin-thing/server"
	"fmt"
	"log"
	"log/slog"

	"github.com/urfave/cli/v2"
)

const (
	grpcServerAddr = "127.0.0.1:50051"
	dnsServerAddr  = "127.0.0.1:5154"
	dnsDomain      = ".gotest."
)

var ServerCmd = &cli.Command{
	Name:        "serve",
	Description: "run server",
	Action:      runServer,
}

func runServer(c *cli.Context) error {
	dnsMgr := dns.New(dnsDomain, slog.LevelDebug)

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
