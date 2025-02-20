package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/calvinmclean/goblin/dns"

	"github.com/urfave/cli/v3"
)

var ClientCmd = &cli.Command{
	Name:        "client",
	Description: "run client",
	Action:      runClient,
}

func runClient(ctx context.Context, c *cli.Command) error {
	client, err := dns.NewHTTPClient(serverAddr)
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}

	subdomain := c.Args().First()
	if subdomain == "" {
		subdomain = "test"
	}

	ip, err := client.GetIP(ctx, subdomain)
	if err != nil {
		return fmt.Errorf("error getting IP: %w", err)
	}
	log.Printf("Got IP: %s", ip)

	<-ctx.Done()

	return err
}
