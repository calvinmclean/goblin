package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/calvinmclean/goblin/dns"
	"github.com/calvinmclean/goblin/errors"
	"github.com/calvinmclean/goblin/server"

	"github.com/urfave/cli/v3"
)

const (
	serverAddr    = "127.0.0.1:8080"
	dnsServerAddr = "127.0.0.1:5053"
)

var (
	topLevelDomain, fallbackConfig string
	ServerCmd                      = &cli.Command{
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
			&cli.StringFlag{
				Name:      "fallback-routes",
				Aliases:   []string{"r"},
				TakesFile: true,
				Validator: func(v string) error {
					if filepath.Ext(v) != ".json" {
						return errors.New("fallback-routes must be JSON file")
					}
					return nil
				},
				Usage: `path to a JSON file holding fallback route config in this format:
{
  "subdomain": {
    "Hostname": "remote-server.mydomain",
    "Ports": [8080]
  }
}`,
				Destination: &fallbackConfig,
			},
		},
	}
)

func runServer(ctx context.Context, c *cli.Command) error {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	var fallbackRoutes dns.FallbackRoutes
	if fallbackConfig != "" {
		data, err := os.ReadFile(fallbackConfig)
		if err != nil {
			return fmt.Errorf("error opening fallback routes config: %w", err)
		}

		err = json.Unmarshal(data, &fallbackRoutes)
		if err != nil {
			return fmt.Errorf("error parsing fallback routes config: %w", err)
		}
	}

	dnsMgr, err := dns.New(topLevelDomain, dnsServerAddr, fallbackRoutes)
	if err != nil {
		errors.PrintUserFixableErrorInstruction(err)
		return fmt.Errorf("error creating DNS Manager: %w", err)
	}

	server := server.New(dnsMgr, serverAddr)
	err = server.Run(ctx)
	if err != nil {
		log.Fatalf("error running GRPC server: %v", err)
	}

	return nil
}
