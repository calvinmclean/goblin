package cmd

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/calvinmclean/goblin/dns"
	"github.com/calvinmclean/goblin/server"

	"github.com/urfave/cli/v3"
)

var ExampleCmd = &cli.Command{
	Name:        "example",
	Description: "run example",
	Action:      runExample,
}

func runExample(ctx context.Context, c *cli.Command) error {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	dnsMgr, err := dns.New("goblin", dnsServerAddr, nil)
	if err != nil {
		return fmt.Errorf("error creating DNS Manager: %w", err)
	}

	go func() {
		err := runPlugin(ctx, dnsMgr, "./example-plugins/helloworld/cmd/hello/hello.so", "helloworld", 0)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		time.Sleep(5 * time.Second)

		err := runPlugin(ctx, dnsMgr, "./example-plugins/helloworld/cmd/hello/howdy.so", "howdy", 0)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		time.Sleep(15 * time.Second)

		err := runPlugin(ctx, dnsMgr, "./example-plugins/helloworld/cmd/hello/howdy.so", "howdynew", 0)
		if err != nil {
			panic(err)
		}
	}()

	server := server.New(dnsMgr, serverAddr)
	err = server.Run(ctx)
	if err != nil {
		log.Fatalf("error running GRPC server: %v", err)
	}

	return nil
}
