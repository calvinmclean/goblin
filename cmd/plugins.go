package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/calvinmclean/goblin/dns"
	"github.com/calvinmclean/goblin/plugins"

	"github.com/urfave/cli/v2"
)

var (
	filename, subdomain string
	PluginCmd           = &cli.Command{
		Name:        "plugin",
		Description: "run a plugin",
		Action:      runPluginCmd,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "filename",
				Aliases:     []string{"f"},
				Required:    true,
				TakesFile:   true,
				Usage:       "filename for *.so plugin",
				Destination: &filename,
			},
			&cli.StringFlag{
				Name:        "subdomain",
				Aliases:     []string{"d"},
				Required:    true,
				Usage:       "subdomain name",
				Destination: &subdomain,
			},
		},
	}
)

func runPluginCmd(c *cli.Context) error {
	client, err := dns.NewGRPC(grpcServerAddr)
	if err != nil {
		return fmt.Errorf("error creating GRPC Client: %w", err)
	}
	return runPlugin(c.Context, client, filename, subdomain, 0)
}

func runPlugin(ctx context.Context, dnsMgr plugins.IPGetter, fname, subdomain string, timeout time.Duration) error {
	if timeout != 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		ctx = timeoutCtx
	}

	log.Printf("starting plugin: %q", subdomain)

	run, err := plugins.Load(fname)
	if err != nil {
		return fmt.Errorf("error loading plugin: %w", err)
	}

	err = plugins.Run(ctx, run, dnsMgr, subdomain)
	if err != nil {
		return fmt.Errorf("error running plugin: %w", err)
	}

	log.Printf("stopped plugin: %q", subdomain)
	return err
}
