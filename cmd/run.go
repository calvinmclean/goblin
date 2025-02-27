package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/calvinmclean/goblin/dns"
	"github.com/calvinmclean/goblin/errors"
	"github.com/calvinmclean/goblin/plugins"

	"github.com/urfave/cli/v3"
)

var (
	pluginFilename, subdomain, ipEnvVar string
	isDir                               bool
	RunCmd                              = &cli.Command{
		Name:        "run",
		Description: "build and run a plugin",
		Action:      runPluginCmd,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "plugin",
				Aliases:     []string{"p"},
				Required:    true,
				TakesFile:   true,
				Usage:       "filename for *.so plugin or directory for building it",
				Destination: &pluginFilename,
				Validator: func(v string) error {
					if filepath.Ext(v) == ".so" {
						return nil
					}

					stat, err := os.Stat(v)
					if err != nil {
						return err
					}

					if !stat.IsDir() {
						return errors.New("plugin must be .so file or directory to build from")
					}
					isDir = true

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
			&cli.StringFlag{
				Name:    "env",
				Aliases: []string{"e"},
				Usage: "environment variable to communicate IP. Goblin will set this env var" +
					" with the allocated IP and run your application's main() function",
				Destination: &ipEnvVar,
			},
			&cli.StringFlag{
				Name:        "port",
				Value:       defaultServerPort,
				Usage:       "port to reach the API server running locally",
				Destination: &serverPort,
				Sources:     cli.ValueSourceChain{Chain: []cli.ValueSource{portEnvVar}},
			},
		},
	}
)

func runPluginCmd(ctx context.Context, c *cli.Command) error {
	client, err := dns.NewHTTPClient(net.JoinHostPort(defaultAddr, serverPort))
	if err != nil {
		return fmt.Errorf("error creating GRPC Client: %w", err)
	}
	return runPlugin(ctx, client, pluginFilename, subdomain, 0)
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

	if isDir {
		builtPlugin, err := plugins.Build(fname)
		if err != nil {
			errors.PrintUserFixableErrorInstruction(err)
			return fmt.Errorf("error building plugin: %w", err)
		}
		fname = builtPlugin
	}

	var err error
	var run plugins.RunFunc
	if ipEnvVar != "" {
		run, err = plugins.LoadMainWithIPEnvVar(fname, ipEnvVar)
	} else {
		run, err = plugins.Load(fname)
	}

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
