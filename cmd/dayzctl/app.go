package main

import (
	"log"
	"os"

	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/rcon"
	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/shared"
	"github.com/kabroxiko/dayzctl/internal/config"
	"github.com/kabroxiko/dayzctl/internal/logger"
	cli "github.com/urfave/cli/v2"
)

func NewApp() *cli.App {
	app := &cli.App{
		Name:  "dayzctl",
		Usage: "DayZ server management tool",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Value: config.DefaultConfigPath(), Usage: "Config file path"},
			&cli.StringFlag{Name: "log-level", Value: "info", Usage: "Log level (debug, info, warn, error)"},
		},
		Before: func(c *cli.Context) error {
			if os.Geteuid() != 0 {
				return cli.Exit("dayzctl must be run as root", 1)
			}
			cfgPath := c.String("config")
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return cli.Exit("failed to load config: "+err.Error(), 1)
			}
			shared.Config = cfg
			logger.Init(c.String("log-level"))
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:      "rcon",
				Usage:     "RCON commands for an instance",
				ArgsUsage: "<instance> <command>",
				Subcommands: []*cli.Command{
					{
						Name:  "players",
						Usage: "List players on an instance",
						Action: func(c *cli.Context) error {
							inst := c.Args().Get(0)
							if inst == "" {
								return cli.Exit("instance name required", 1)
							}
							return rcon.PlayersAction(inst, c.Args().Tail())
						},
					},
				},
				Action: func(c *cli.Context) error {
					// If no subcommand provided, show help
					return cli.ShowAppHelp(c)
				},
			},
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
