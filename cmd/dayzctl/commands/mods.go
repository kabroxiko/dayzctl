package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kabroxiko/dayzctl/internal/config"
	"github.com/kabroxiko/dayzctl/internal/generate"
	"github.com/kabroxiko/dayzctl/internal/logger"
	"github.com/kabroxiko/dayzctl/internal/mods"
	"github.com/kabroxiko/dayzctl/internal/steamcmd"
	"github.com/kabroxiko/dayzctl/internal/systemd"
	"github.com/spf13/cobra"
)

func ModsCmd() *cobra.Command {
	modsCmd := &cobra.Command{
		Use:   "mods",
		Short: "Manage mods",
		Long: `Manage mods for DayZ server instances.

Commands:
  list                  List all installed mods
  add <instance> <id> [id...]   Add client mods to an instance
  add-server <instance> <id> [id...]  Add server-side mods to an instance
  remove <instance> <id> [id...]      Remove mods from an instance (config only)
  delete <instance> <id> [id...]      Alias for remove
  sync <instance>       Sync mods for an instance`,
	}

	modsCmd.AddCommand(
		modsListCmd(),
		modsSyncCmd(),
		modsAddCmd(),
		modsAddServerCmd(),
		modsRemoveCmd(),
		modsDeleteCmd(),
	)

	return modsCmd
}

func modsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed mods with names",
		Run: func(cmd *cobra.Command, args []string) {
			RunCommand(func() error {
				modManager := mods.New(Config.GetInstallDir(), Config.Paths.WorkshopDir)
				installed, err := modManager.ListInstalled()
				if err != nil {
					return err
				}
				if len(installed) == 0 {
					fmt.Println("No mods installed")
					return nil
				}
				fmt.Printf("Installed mods (%d):\n", len(installed))
				for _, mod := range installed {
					fmt.Printf("  ID: %-10s Name: %s\n", mod.ID, mod.Name)
				}
				return nil
			})
		},
	}
}

func modsSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync [instance]",
		Short: "Sync mods for an instance",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			RunCommand(func() error {
				instance, err := GetInstance(args[0])
				if err != nil {
					return err
				}

				logger.Info("Syncing mods for instance", "name", instance.Name)
				modManager := mods.New(Config.GetInstallDir(), Config.Paths.WorkshopDir)

				if len(instance.Mods) == 0 && len(instance.ServerMods) == 0 {
					logger.Info("No mods configured for instance", "name", instance.Name)
					return nil
				}

				if err := modManager.SyncMods(instance.Mods, instance.ServerMods); err != nil {
					return err
				}

				if err := UpdateServerConfig(instance); err != nil {
					logger.Warn("Failed to update server config", "error", err)
				}

				if err := applyConfig(); err != nil {
					logger.Warn("Failed to apply config changes", "error", err)
				}

				logger.Info("Mods synced successfully", "count", len(instance.Mods)+len(instance.ServerMods))
				return nil
			})
		},
	}
}

func modsAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add [instance] [mod_id] [mod_id...]",
		Short: "Add client mods to an instance",
		Long: `Add one or more client mods to an instance. Client mods are downloaded by players when joining.

Examples:
  dayzctl mods add solo 3765278986
  dayzctl mods add solo 1559212036 1564026768 1646187754`,
		Args: cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			RunCommand(func() error {
				instance, err := GetInstance(args[0])
				if err != nil {
					return err
				}
				modIDs := args[1:]

				var toAdd []string
				for _, modID := range modIDs {
					found := false
					for _, m := range instance.Mods {
						if m.ID == modID {
							logger.Info("Mod already in instance (client mods)", "mod", modID)
							found = true
							break
						}
					}
					if !found {
						for _, m := range instance.ServerMods {
							if m.ID == modID {
								logger.Warn("Mod is already in servermods, not adding to mods", "mod", modID)
								found = true
								break
							}
						}
					}
					if !found {
						toAdd = append(toAdd, modID)
					}
				}

				if len(toAdd) == 0 {
					logger.Info("No new mods to add")
					return nil
				}

				logger.Info("Adding mods", "count", len(toAdd))

				for _, modID := range toAdd {
					if err := addModToInstance(instance, modID, false); err != nil {
						logger.Warn("Failed to add mod", "mod_id", modID, "error", err)
					}
				}

				return nil
			})
		},
	}
}

func modsAddServerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add-server [instance] [mod_id] [mod_id...]",
		Short: "Add server-side mods to an instance",
		Long: `Add one or more server-side mods to an instance. Server-side mods are NOT downloaded by players.
These are typically mods that only affect server behavior.

Examples:
  dayzctl mods add-server solo 3765278986
  dayzctl mods add-server solo 1559212036 1564026768`,
		Args: cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			RunCommand(func() error {
				instance, err := GetInstance(args[0])
				if err != nil {
					return err
				}
				modIDs := args[1:]

				var toAdd []string
				for _, modID := range modIDs {
					found := false
					for _, m := range instance.ServerMods {
						if m.ID == modID {
							logger.Info("Mod already in instance (server mods)", "mod", modID)
							found = true
							break
						}
					}
					if !found {
						for _, m := range instance.Mods {
							if m.ID == modID {
								logger.Warn("Mod is already in mods, not adding to servermods", "mod", modID)
								found = true
								break
							}
						}
					}
					if !found {
						toAdd = append(toAdd, modID)
					}
				}

				if len(toAdd) == 0 {
					logger.Info("No new mods to add")
					return nil
				}

				logger.Info("Adding server mods", "count", len(toAdd))

				for _, modID := range toAdd {
					if err := addModToInstance(instance, modID, true); err != nil {
						logger.Warn("Failed to add server mod", "mod_id", modID, "error", err)
					}
				}

				return nil
			})
		},
	}
}

func modsRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [instance] [mod_id] [mod_id...]",
		Short: "Remove mods from an instance (config only)",
		Long: `Remove one or more mods from an instance. This will remove the mods from the config only.
The symlinks and downloaded files are preserved in case other instances use them.

Examples:
  dayzctl mods remove solo 3765278986
  dayzctl mods remove solo 1559212036 1564026768`,
		Args: cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			runRemoveMods(args, false)
		},
	}

	return cmd
}

func modsDeleteCmd() *cobra.Command {
	var deleteFiles bool

	cmd := &cobra.Command{
		Use:   "delete [instance] [mod_id] [mod_id...]",
		Short: "Delete mods from an instance (remove config, symlinks, and optionally files)",
		Long: `Delete one or more mods from an instance. This will remove the mods from the config,
delete the symlinks, and optionally delete the downloaded mod files.

Examples:
  dayzctl mods delete solo 3765278986
  dayzctl mods delete solo 1559212036 1564026768 --delete-files`,
		Args: cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			runDeleteMods(args, deleteFiles)
		},
	}

	cmd.Flags().BoolVar(&deleteFiles, "delete-files", false, "Delete the downloaded mod files from the workshop directory")

	return cmd
}

// runRemoveMods removes mods from config only (preserves symlinks and files)
func runRemoveMods(args []string, deleteFiles bool) {
	RunCommand(func() error {
		instance, err := GetInstance(args[0])
		if err != nil {
			return err
		}
		modIDs := args[1:]

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

		if err := SaveConfig(); err != nil {
			logger.Warn("Failed to save config", "error", err)
		}

		if err := UpdateServerConfig(instance); err != nil {
			logger.Warn("Failed to update server config", "error", err)
		}

		if err := applyConfig(); err != nil {
			logger.Warn("Failed to apply config changes", "error", err)
		}

		logger.Info("Mods removed from config. Symlinks and files preserved.")
		return nil
	})
}

// runDeleteMods removes mods from config, deletes symlinks, and optionally deletes files
func runDeleteMods(args []string, deleteFiles bool) {
	RunCommand(func() error {
		instance, err := GetInstance(args[0])
		if err != nil {
			return err
		}
		modIDs := args[1:]

		modManager := mods.New(Config.GetInstallDir(), Config.Paths.WorkshopDir)
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

			// Delete the symlink
			if err := modManager.RemoveMod(modRef, isServerMod); err != nil {
				logger.Warn("Failed to remove symlink", "mod", modID, "error", err)
			} else {
				logger.Info("Removed symlink", "mod", modID)
			}

			// Delete the mod files from workshop if requested
			if deleteFiles {
				modPath := filepath.Join(Config.GetWorkshopDir(), modID)
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

		if err := SaveConfig(); err != nil {
			logger.Warn("Failed to save config", "error", err)
		}

		if err := UpdateServerConfig(instance); err != nil {
			logger.Warn("Failed to update server config", "error", err)
		}

		if err := applyConfig(); err != nil {
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

// addModToInstance is a helper function to add a mod to an instance
func addModToInstance(instance *config.Instance, modID string, isServerMod bool) error {
	modManager := mods.New(Config.GetInstallDir(), Config.Paths.WorkshopDir)

	logger.Info("Downloading mod", "mod_id", modID)
	steam := steamcmd.New(Config.GetSteamUser(), Config.GetInstallDir(), Config.GetSteamcmdBin())
	if err := steam.DownloadMod(modID); err != nil {
		logger.Warn("Failed to download mod", "mod_id", modID, "error", err)
	}

	modInfo, err := modManager.GetModInfo(modID)
	if err != nil {
		logger.Warn("Failed to get mod info, using ID as name", "error", err)
		modInfo = mods.Mod{ID: modID, Name: modID}
	}

	if isServerMod {
		instance.ServerMods = append(instance.ServerMods, config.ModRef{
			ID:   modID,
			Name: modInfo.Name,
		})
		logger.Info("Adding as server mod", "mod", modID, "name", modInfo.Name)
	} else {
		instance.Mods = append(instance.Mods, config.ModRef{
			ID:   modID,
			Name: modInfo.Name,
		})
		logger.Info("Adding as client mod", "mod", modID, "name", modInfo.Name)
	}

	if err := SaveConfig(); err != nil {
		logger.Warn("Failed to save config", "error", err)
	}

	if err := modManager.SyncMods(instance.Mods, instance.ServerMods); err != nil {
		logger.Warn("Failed to sync mods", "error", err)
	}

	if err := UpdateServerConfig(instance); err != nil {
		logger.Warn("Failed to update server config", "error", err)
	}

	if err := applyConfig(); err != nil {
		logger.Warn("Failed to apply config changes", "error", err)
	}

	modType := "client"
	if isServerMod {
		modType = "server"
	}
	logger.Info("Added mod to instance", "instance", instance.Name, "mod", modID, "name", modInfo.Name, "type", modType)
	return nil
}

// applyConfig applies the current configuration (generates server configs and systemd units)
func applyConfig() error {
	logger.Info("Applying configuration changes...")

	if err := generate.GenerateAll(Config); err != nil {
		return fmt.Errorf("failed to generate server configs: %w", err)
	}

	sysd := systemd.New()
	if err := sysd.GenerateUnits(Config); err != nil {
		return fmt.Errorf("failed to generate units: %w", err)
	}
	if err := sysd.Reload(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	logger.Info("Configuration applied successfully")
	return nil
}
