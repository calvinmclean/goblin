package main

import (
	"context"
	"log"
	"os"

	"github.com/calvinmclean/goblin/cmd"

	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name: "goblin",
		Commands: []*cli.Command{
			cmd.ClientCmd,
			cmd.ServerCmd,
			cmd.ExampleCmd,
			cmd.PluginCmd,
		},
	}

	err := app.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
