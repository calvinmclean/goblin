package cmd

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/calvinmclean/goblin/dns"

	"github.com/urfave/cli/v3"
)

var ClientCmd = &cli.Command{
	Name:        "client",
	Description: "run client",
	Action:      runClient,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "port",
			Value:       defaultServerPort,
			Usage:       "port to reach the API server running locally",
			Destination: &serverPort,
			Sources:     cli.ValueSourceChain{Chain: []cli.ValueSource{portEnvVar}},
		},
		&cli.StringFlag{
			Name:        "subdomain",
			Aliases:     []string{"d"},
			Usage:       "subdomain name",
			Destination: &subdomain,
			Required:    true,
		},
	},
}

func runClient(ctx context.Context, c *cli.Command) error {
	client, err := dns.NewHTTPClient(net.JoinHostPort(defaultAddr, serverPort))
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}

	ip, err := client.GetIP(ctx, subdomain)
	if err != nil {
		return fmt.Errorf("error getting IP: %w", err)
	}
	log.Printf("Got IP: %s", ip)

	<-ctx.Done()

	return err
}
