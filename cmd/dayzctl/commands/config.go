package commands

import (
	"github.com/kabroxiko/dayzctl/internal/logger"
	"github.com/spf13/cobra"
)

func ValidateConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate-config",
		Short: "Validate the configuration file",
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("Config loaded successfully")
			logger.Info("Steam user", "user", Config.GetSteamUser())
			logger.Info("Install directory", "path", Config.GetInstallDir())
			logger.Info("Instances", "count", len(Config.Instances))
			for _, inst := range Config.Instances {
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
