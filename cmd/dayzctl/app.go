package main

import (
	"log"
	"os"

	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands"
	cli "github.com/urfave/cli/v2"
)

func NewApp() *cli.App {
	app := &cli.App{
		Name:  "dayzctl",
		Usage: "DayZ server management tool",
		Commands: []*cli.Command{
			{
				Name:      "rcon",
				Usage:     "RCON commands for an instance",
				ArgsUsage: "<instance> <command>",
				Action: func(c *cli.Context) error {
					// Forward to cobra using the original argv so positional placement is preserved
					return commands.ExecuteWithArgs(os.Args[1:])
				},
			},
			{
				Name:  "mods",
				Usage: "Manage mods",
				Action: func(c *cli.Context) error {
					return commands.ExecuteWithArgs(os.Args[1:])
				},
			},
		},
            Action: func(c *cli.Context) error {
                // Forward the original argv (without program name) to cobra
                return commands.ExecuteWithArgs(os.Args[1:])
            },
	}
	return app
}

func RunApp() {
	app := NewApp()
	if err := app.Run(os.Args); err != nil {
		log.Fatalf("app error: %v", err)
	}
}
