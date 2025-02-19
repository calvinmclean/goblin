package main

import (
	"log"
	"os"

	"dns-plugin-thing/cmd"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "dns-plugin-thing",
		Commands: []*cli.Command{
			cmd.ClientCmd,
			cmd.ServerCmd,
			cmd.ExampleCmd,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
