package mods

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/shared"
	"github.com/kabroxiko/dayzctl/internal/config"
	"github.com/kabroxiko/dayzctl/internal/logger"
	"github.com/kabroxiko/dayzctl/internal/mods"
	"github.com/spf13/cobra"
)

func DeleteCmd() *cobra.Command {
	var deleteFiles bool

	cmd := &cobra.Command{
		Use:   "delete [mod_id] [mod_id...]",
		Short: "Delete mods from an instance (remove config, symlinks, and optionally files)",
		Long: `Delete one or more mods from an instance. This will remove the mods from the config,
delete the symlinks, and optionally delete the downloaded mod files.

Examples:
  dayzctl mods solo delete 3765278986
  dayzctl mods solo delete 1559212036 1564026768 --delete-files`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runDeleteMods(shared.GetInstanceNameFromParent(cmd), args, deleteFiles)
		},
	}

	cmd.Flags().BoolVar(&deleteFiles, "delete-files", false, "Delete the downloaded mod files from the workshop directory")

	return cmd
}

func runDeleteMods(instanceName string, modIDs []string, deleteFiles bool) {
	shared.RunCommand(func() error {
		if instanceName == "" {
			return fmt.Errorf("instance name required")
		}
		instance, err := shared.GetInstance(instanceName)
		if err != nil {
			return err
		}

		modManager := mods.New(shared.Config.GetInstallDir(), shared.Config.Paths.WorkshopDir)
		removedCount := 0

		for _, modID := range modIDs {
			found := false
			isServerMod := false
			var modRef config.ModRef

			var newMods []config.ModRef
			for _, m := range instance.Mods {
				if m.ID == modID {
					found = true
					isServerMod = false
					modRef = m
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
					modRef = m
					continue
				}
				newServerMods = append(newServerMods, m)
			}
			instance.ServerMods = newServerMods

			if !found {
				logger.Warn("Mod not found in instance", "mod", modID)
				continue
			}

			if err := modManager.RemoveMod(modRef, isServerMod); err != nil {
				logger.Warn("Failed to remove symlink", "mod", modID, "error", err)
			} else {
				logger.Info("Removed symlink", "mod", modID)
			}

			if deleteFiles {
				modPath := filepath.Join(shared.Config.GetWorkshopDir(), modID)
				logger.Info("Deleting mod files", "path", modPath)
				if err := os.RemoveAll(modPath); err != nil {
					logger.Warn("Failed to delete mod files", "error", err)
				} else {
					logger.Info("Mod files deleted", "path", modPath)
				}
			}

			modType := "client"
			if isServerMod {
				modType = "server"
			}
			logger.Info("Deleted mod from instance", "instance", instance.Name, "mod", modID, "type", modType)
			removedCount++
		}

		if removedCount == 0 {
			logger.Info("No mods were deleted")
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

		if deleteFiles {
			logger.Info("Mod files deleted from workshop")
		} else {
			logger.Info("Mod files preserved in workshop (use --delete-files to remove)")
		}
		return nil
	})
}
