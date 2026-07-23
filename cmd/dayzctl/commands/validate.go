package commands

import (
	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/shared"
	"github.com/kabroxiko/dayzctl/internal/logger"
	"github.com/spf13/cobra"
)

func ValidateConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate-config",
		Short: "Validate the configuration file",
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("Config loaded successfully")
			logger.Info("Steam user", "user", shared.Config.GetSteamUser())
			logger.Info("Base directory", "path", shared.Config.GetBaseDir())
			logger.Info("Install directory", "path", shared.Config.GetInstallDir())
			logger.Info("Instances", "count", len(shared.Config.Instances))
			for _, inst := range shared.Config.Instances {
				logger.Info("Instance",
					"name", inst.Name,
					"port", inst.Port,
					"enabled", inst.Enabled,
					"mods", len(inst.Mods),
					"servermods", len(inst.ServerMods))
			}
		},
	}
}
