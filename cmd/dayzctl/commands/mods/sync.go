package mods

import (
	"fmt"

	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/shared"
	"github.com/kabroxiko/dayzctl/internal/logger"
	"github.com/kabroxiko/dayzctl/internal/mods"
	"github.com/spf13/cobra"
)

func SyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync mods for an instance",
		Run: func(cmd *cobra.Command, args []string) {
			shared.RunCommand(func() error {
				instanceName := shared.GetInstanceNameFromParent(cmd)
				if instanceName == "" {
					return fmt.Errorf("instance name required")
				}
				instance, err := shared.GetInstance(instanceName)
				if err != nil {
					return err
				}

				logger.Info("Syncing mods for instance", "name", instance.Name)
				modManager := mods.New(shared.Config.GetInstallDir(), shared.Config.Paths.WorkshopDir)

				if len(instance.Mods) == 0 && len(instance.ServerMods) == 0 {
					logger.Info("No mods configured for instance", "name", instance.Name)
					return nil
				}

				if err := modManager.SyncMods(instance.Mods, instance.ServerMods); err != nil {
					return err
				}

				if err := shared.UpdateServerConfig(instance); err != nil {
					logger.Warn("Failed to update server config", "error", err)
				}

				if err := shared.ApplyConfig(); err != nil {
					logger.Warn("Failed to apply config changes", "error", err)
				}

				logger.Info("Mods synced successfully", "count", len(instance.Mods)+len(instance.ServerMods))
				return nil
			})
		},
	}
}
