package cmd

import (
	"dns-plugin-thing/dns"
	"dns-plugin-thing/server"
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
	slog.SetLogLoggerLevel(slog.LevelDebug)

	dnsMgr := dns.New(dnsDomain)
	server := server.New(dnsMgr)

	err := server.Run(c.Context, grpcServerAddr, dnsServerAddr)
	if err != nil {
		log.Fatalf("error running GRPC server: %v", err)
	}

	return nil
}
