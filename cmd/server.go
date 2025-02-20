package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"

	"github.com/calvinmclean/goblin/dns"
	"github.com/calvinmclean/goblin/server"

	"github.com/urfave/cli/v3"
)

const (
	serverAddr    = "127.0.0.1:8080"
	dnsServerAddr = "127.0.0.1:5053"
)

var (
	topLevelDomain string
	ServerCmd      = &cli.Command{
		Name:        "server",
		Description: "run server",
		Action:      runServer,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "domain",
				Aliases:     []string{"d"},
				Value:       "goblin",
				Usage:       "top-level domain name to use",
				Destination: &topLevelDomain,
			},
		},
	}
)

func runServer(ctx context.Context, c *cli.Command) error {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	dnsMgr, err := dns.New(topLevelDomain, dnsServerAddr)
	if err != nil {
		var configErr dns.UserFixableError
		if errors.As(err, &configErr) {
			fmt.Println(configErr.Instructions)
		}
		return fmt.Errorf("error creating DNS Manager: %w", err)
	}

	server := server.New(dnsMgr, serverAddr)
	err = server.Run(ctx)
	if err != nil {
		log.Fatalf("error running GRPC server: %v", err)
	}

	return nil
}
