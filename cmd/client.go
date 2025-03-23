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
	Description: "run client to demonstrate registering an IP for the lifecycle of a program",
	Action:      runClient,
	Flags: []cli.Flag{
		portFlag,
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
