package main

import (
	"log"
	"os"

	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands"
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
				Action: func(c *cli.Context) error {
					// Preserve legacy UX: accept `rcon <instance> <command>`
					inst := c.Args().Get(0)
					if inst == "" {
						return cli.ShowCommandHelp(c, "rcon")
					}
					sub := c.Args().Get(1)
					if sub == "" {
						return cli.ShowCommandHelp(c, "rcon")
					}

					switch sub {
					case "players":
						return rcon.PlayersAction(inst, c.Args().Tail())
					case "send":
						return rcon.SendAction(inst, c.Args().Tail())
					case "kick":
						return rcon.KickAction(inst, c.Args().Tail())
					case "ban":
						return rcon.BanAction(inst, c.Args().Tail())
					case "say":
						return rcon.SayAction(inst, c.Args().Tail())
					default:
						// Forward unknown subcommands to Cobra-based handlers for now
						return commands.ExecuteWithArgs(os.Args[1:])
					}
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
