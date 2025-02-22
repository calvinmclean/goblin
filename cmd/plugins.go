package cmd

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/calvinmclean/goblin/dns"
	"github.com/calvinmclean/goblin/errors"
	"github.com/calvinmclean/goblin/plugins"

	"github.com/urfave/cli/v3"
)

var (
	filename, subdomain string
	PluginCmd           = &cli.Command{
		Name:        "plugin",
		Description: "run a plugin",
		Action:      runPluginCmd,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "plugin",
				Aliases:     []string{"p"},
				Required:    true,
				TakesFile:   true,
				Usage:       "filename for *.so plugin",
				Destination: &filename,
				Validator: func(v string) error {
					if filepath.Ext(v) != ".so" {
						return errors.New("plugin must be .so file")
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:        "subdomain",
				Aliases:     []string{"d"},
				Usage:       "subdomain name",
				DefaultText: "plugin filename (without .so)",
				Destination: &subdomain,
			},
		},
	}
)

func runPluginCmd(ctx context.Context, c *cli.Command) error {
	client, err := dns.NewHTTPClient(serverAddr)
	if err != nil {
		return fmt.Errorf("error creating GRPC Client: %w", err)
	}
	return runPlugin(ctx, client, filename, subdomain, 0)
}

func runPlugin(ctx context.Context, dnsMgr plugins.IPGetter, fname, subdomain string, timeout time.Duration) error {
	if timeout != 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		ctx = timeoutCtx
	}

	if subdomain == "" {
		subdomain = strings.TrimSuffix(filepath.Base(fname), ".so")
	}

	run, err := plugins.Load(fname)
	if err != nil {
		errors.PrintUserFixableErrorInstruction(err)
		return fmt.Errorf("error loading plugin: %w", err)
	}

	log.Printf("starting plugin: %q", subdomain)
	err = plugins.Run(ctx, run, dnsMgr, subdomain)
	if err != nil {
		return fmt.Errorf("error running plugin: %w", err)
	}

	log.Printf("stopped plugin: %q", subdomain)
	return err
}
