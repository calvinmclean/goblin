package cmd

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/calvinmclean/goblin/dns"

	"github.com/urfave/cli/v3"
)

var (
	address     string
	RegisterCmd = &cli.Command{
		Name:        "register",
		Description: "register a fallback route with the server",
		Action:      runRegister,
		Flags: []cli.Flag{
			portFlag,
			&cli.StringFlag{
				Name:        "subdomain",
				Aliases:     []string{"d"},
				Usage:       "subdomain name",
				Destination: &subdomain,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "address",
				Aliases:     []string{"a"},
				Usage:       "fallback address to route to",
				Destination: &address,
				Required:    true,
			},
		},
	}
)

func runRegister(ctx context.Context, c *cli.Command) error {
	client, err := dns.NewHTTPClient(net.JoinHostPort(defaultAddr, serverPort))
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}

	err = client.RegisterFallback(subdomain, address)
	if err != nil {
		return fmt.Errorf("error registering fallback: %w", err)
	}
	log.Print("registered fallback")

	return err
}
