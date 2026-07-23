package mods

import (
	"fmt"

	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/shared"
	"github.com/kabroxiko/dayzctl/internal/config"
	"github.com/kabroxiko/dayzctl/internal/logger"
	"github.com/spf13/cobra"
)

func RemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [mod_id] [mod_id...]",
		Short: "Remove mods from an instance (config only)",
		Long: `Remove one or more mods from an instance. This will remove the mods from the config only.
The symlinks and downloaded files are preserved in case other instances use them.

Examples:
  dayzctl mods solo remove 3765278986
  dayzctl mods solo remove 1559212036 1564026768`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runRemoveMods(shared.GetInstanceNameFromParent(cmd), args, false)
		},
	}

	return cmd
}

func runRemoveMods(instanceName string, modIDs []string, deleteFiles bool) {
	shared.RunCommand(func() error {
		if instanceName == "" {
			return fmt.Errorf("instance name required")
		}
		instance, err := shared.GetInstance(instanceName)
		if err != nil {
			return err
		}

		removedCount := 0

		for _, modID := range modIDs {
			found := false
			isServerMod := false

			var newMods []config.ModRef
			for _, m := range instance.Mods {
				if m.ID == modID {
					found = true
					isServerMod = false
					continue
				}
				newMods = append(newMods, m)
			}
			instance.Mods = newMods

			var newServerMods []config.ModRef
			for _, m := range instance.ServerMods {
				if m.ID == modID {
					found = true
					isServerMod = true
					continue
				}
				newServerMods = append(newServerMods, m)
			}
			instance.ServerMods = newServerMods

			if !found {
				logger.Warn("Mod not found in instance", "mod", modID)
				continue
			}

			modType := "client"
			if isServerMod {
				modType = "server"
			}
			logger.Info("Removed mod from instance (config only)", "instance", instance.Name, "mod", modID, "type", modType)
			removedCount++
		}

		if removedCount == 0 {
			logger.Info("No mods were removed")
			return nil
		}

		if err := shared.SaveConfig(); err != nil {
			logger.Warn("Failed to save config", "error", err)
		}

		if err := shared.UpdateServerConfig(instance); err != nil {
			logger.Warn("Failed to update server config", "error", err)
		}

		if err := shared.ApplyConfig(); err != nil {
			logger.Warn("Failed to apply config changes", "error", err)
		}

		logger.Info("Mods removed from config. Symlinks and files preserved.")
		return nil
	})
}
