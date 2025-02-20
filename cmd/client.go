package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/calvinmclean/goblin/dns"

	"github.com/urfave/cli/v2"
)

var ClientCmd = &cli.Command{
	Name:        "client",
	Description: "run client",
	Action:      runClient,
}

func runClient(c *cli.Context) error {
	ctx, cancel := context.WithTimeout(c.Context, 5*time.Second)
	defer cancel()

	client, err := dns.NewGRPC(grpcServerAddr)
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

	<-c.Context.Done()

	return err
}
