package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"path/filepath"

	"github.com/calvinmclean/goblin/dns"
	"github.com/calvinmclean/goblin/errors"
	"github.com/calvinmclean/goblin/server"

	"github.com/urfave/cli/v3"
)

const (
	defaultAddr       = "127.0.0.1"
	defaultServerPort = "8080"
	defaultDNSPort    = "5053"
)

var (
	portEnvVar = cli.EnvVar("GOBLIN_PORT")

	topLevelDomain, fallbackConfig, serverPort, dnsPort string
	ServerCmd                                           = &cli.Command{
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
			portFlag,
			&cli.StringFlag{
				Name:        "dns-port",
				Aliases:     []string{"s"},
				Value:       defaultDNSPort,
				Usage:       "port to run the DNS server on",
				Destination: &dnsPort,
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
  "subdomain": "remote-server.com"
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

	dnsMgr, err := dns.New(dns.Config{
		Domain:         topLevelDomain,
		Address:        net.JoinHostPort(defaultAddr, dnsPort),
		FallbackRoutes: fallbackRoutes,
	})
	if err != nil {
		errors.PrintUserFixableErrorInstruction(err)
		return fmt.Errorf("error creating DNS Manager: %w", err)
	}

	server := server.New(dnsMgr, net.JoinHostPort(defaultAddr, serverPort))
	err = server.Run(ctx)
	if err != nil {
		log.Fatalf("error running GRPC server: %v", err)
	}

	return nil
}
