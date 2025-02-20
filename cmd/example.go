package cmd

import (
	"dns-plugin-thing/dns"
	"dns-plugin-thing/server"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/urfave/cli/v2"
)

var ExampleCmd = &cli.Command{
	Name:        "example",
	Description: "run example",
	Action:      runExample,
}

func runExample(c *cli.Context) error {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	dnsMgr, err := dns.New("gotest", dnsServerAddr)
	if err != nil {
		return fmt.Errorf("error creating DNS Manager: %w", err)
	}

	go func() {
		err := runPlugin(c.Context, dnsMgr, "./example-plugins/helloworld/cmd/hello/hello.so", "helloworld", 0)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		time.Sleep(5 * time.Second)

		err := runPlugin(c.Context, dnsMgr, "./example-plugins/helloworld/cmd/hello/howdy.so", "howdy", 0)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		time.Sleep(15 * time.Second)

		err := runPlugin(c.Context, dnsMgr, "./example-plugins/helloworld/cmd/hello/howdy.so", "howdynew", 0)
		if err != nil {
			panic(err)
		}
	}()

	server := server.New(dnsMgr)
	err = server.Run(c.Context, grpcServerAddr)
	if err != nil {
		log.Fatalf("error running GRPC server: %v", err)
	}

	return nil
}
