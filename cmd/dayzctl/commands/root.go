package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/mods"
	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/rcon"
	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/shared"
	"github.com/kabroxiko/dayzctl/internal/config"
	"github.com/kabroxiko/dayzctl/internal/logger"
	"github.com/kabroxiko/dayzctl/internal/version"
	"github.com/spf13/cobra"
)

var (
	configPath string
	logLevel   string
	rootCmd    *cobra.Command
)

func init() {
	rootCmd = &cobra.Command{
		Use:   "dayzctl",
		Short: "DayZ Server Control Tool",
		Long:  fmt.Sprintf("DayZ server management tool\nVersion: %s\nBuild: %s", version.Version, version.BuildTime),
	}

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", config.DefaultConfigPath(), "Config file path")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if os.Geteuid() != 0 {
			return fmt.Errorf("dayzctl must be run as root")
		}
		var err error
		shared.Config, err = config.Load(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		for _, inst := range shared.Config.Instances {
			if inst.Name == "all" {
				return fmt.Errorf("instance name 'all' is reserved and cannot be used")
			}
		}

		logger.Init(logLevel)
		return nil
	}

	rootCmd.AddCommand(
		VersionCmd(),
		ValidateConfigCmd(),
		UpdateCmd(),
		SteamLoginCmd(),
		mods.ModsCmd(),
		rcon.RconCmd(),
		StartCmd(),
		StopCmd(),
		RestartCmd(),
		StatusCmd(),
		ListCmd(),
		ApplyCmd(),
		CheckSpaceCmd(),
		RenderConfigCmd(),
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// ExecuteWithArgs runs the cobra root command with the given args slice.
func ExecuteWithArgs(args []string) error {
	// Log the args received for debugging positional parsing issues
	if args == nil {
		args = []string{}
	}
	// If user called: dayzctl rcon <instance> <subcmd> ...
	// Cobra expects: dayzctl rcon <subcmd> <instance> ... so rewrite args accordingly.
	if len(args) >= 3 && args[0] == "rcon" {
		subcmds := map[string]struct{}{"send": {}, "players": {}, "kick": {}, "ban": {}, "say": {}}
		// args[1] is candidate instance, args[2] is candidate subcommand
		if _, isSub := subcmds[args[2]]; isSub {
			// swap positions 1 and 2 so cobra sees the subcommand first
			args[1], args[2] = args[2], args[1]
		}
	}
	log.Printf("ExecuteWithArgs called: %v\n", args)
	logger.Debug("ExecuteWithArgs called", "args", args)
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}
