package commands

import (
	"fmt"

	"github.com/kabroxiko/dayzctl/internal/lock"
	"github.com/kabroxiko/dayzctl/internal/logger"
	"github.com/kabroxiko/dayzctl/internal/mods"
	"github.com/kabroxiko/dayzctl/internal/steamcmd"
	"github.com/kabroxiko/dayzctl/internal/systemd"
	"github.com/spf13/cobra"
)

func UpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update DayZ server (stops/restarts services only if update available)",
		Run: func(cmd *cobra.Command, args []string) {
			RunCommand(func() error {
				l, err := lock.New("/run/dayzctl.lock")
				if err != nil {
					return fmt.Errorf("failed to acquire lock: %w", err)
				}
				defer func() {
					if err := l.Release(); err != nil {
						logger.Warn("Failed to release lock", "error", err)
					}
				}()

				if !Config.Updates.Enabled {
					logger.Info("Updates are disabled in config")
					return nil
				}

				logger.Info("Starting update check...")
				steam := steamcmd.New(Config.GetSteamUser(), Config.GetInstallDir(), Config.GetSteamcmdBin())

				// Check current build ID
				buildID, err := steam.GetBuildID()
				if err != nil {
					if steamcmd.IsRateLimitError(err) {
						logger.Warn("Rate limit hit, please wait before retrying")
						return err
					}
					return fmt.Errorf("failed to check build: %w", err)
				}
				logger.Info("Current build", "build_id", buildID)

				// Check if update is needed
				needsUpdate, err := steam.NeedsUpdate()
				if err != nil {
					return fmt.Errorf("failed to check update status: %w", err)
				}
				if !needsUpdate {
					logger.Info("Server is already up to date - no update available")
					return nil
				}

				logger.Info("Update available! Proceeding with update...")

				// Get running instances BEFORE we stop them (to restart later)
				sysd := systemd.New()
				instances, err := sysd.ListRunningInstances()
				if err != nil {
					logger.Warn("Failed to list running instances", "error", err)
					instances = []string{}
				}

				// Stop running instances
				if len(instances) > 0 {
					logger.Info("Stopping running instances...")
					for _, instance := range instances {
						logger.Info("Stopping instance", "name", instance)
						if err := sysd.Stop("dayz@" + instance); err != nil {
							return fmt.Errorf("failed to stop %s: %w", instance, err)
						}
					}
				} else {
					logger.Info("No running instances to stop")
				}

				// Perform the update
				logger.Info("Updating DayZ server as dayz user...")
				if err := steam.Update(); err != nil {
					return fmt.Errorf("update failed: %w", err)
				}
				logger.Info("Update completed successfully")

				// Sync mods for all enabled instances
				modManager := mods.New(Config.GetInstallDir(), Config.Paths.WorkshopDir)
				for _, instance := range Config.Instances {
					if instance.Enabled {
						allMods := append(instance.Mods, instance.ServerMods...)
						if len(allMods) > 0 {
							logger.Info("Syncing mods for instance", "name", instance.Name, "count", len(allMods))
							if err := modManager.SyncMods(allMods, instance.ServerMods); err != nil {
								logger.Warn("Failed to sync mods", "instance", instance.Name, "error", err)
							}
						}
					}
				}

				// Restart previously running instances
				if len(instances) > 0 {
					logger.Info("Restarting previously running instances...")
					for _, instance := range instances {
						logger.Info("Starting instance", "name", instance)
						if err := sysd.Start("dayz@" + instance); err != nil {
							return fmt.Errorf("failed to start %s: %w", instance, err)
						}
					}
				} else {
					logger.Info("No instances to restart (were not running before update)")
				}

				logger.Info("Update process completed successfully")
				return nil
			})
		},
	}
}
