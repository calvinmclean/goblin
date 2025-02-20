package main

import (
	"log"
	"os"

	"github.com/calvinmclean/goblin/cmd"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "goblin",
		Commands: []*cli.Command{
			cmd.ClientCmd,
			cmd.ServerCmd,
			cmd.ExampleCmd,
			cmd.PluginCmd,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
