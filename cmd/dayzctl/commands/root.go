package commands

import (
	"fmt"
	"os"

	"github.com/kabroxiko/dayzctl/internal/config"
	"github.com/kabroxiko/dayzctl/internal/logger"
	"github.com/kabroxiko/dayzctl/internal/version"
	"github.com/spf13/cobra"
)

var (
	configPath string
	logLevel   string
	Config     *config.ServerConfig
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
		Config, err = config.Load(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		logger.Init(logLevel)
		return nil
	}

	rootCmd.AddCommand(
		VersionCmd(),
		ValidateConfigCmd(),
		UpdateCmd(),
		SteamLoginCmd(),
		ModsCmd(),
		RconCmd(),
		InstanceCmd(),
		PlayersCmd(),
		StatusCmd(),
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
