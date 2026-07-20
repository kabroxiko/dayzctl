package commands

import (
	"fmt"

	"github.com/kabroxiko/dayzctl/internal/generate"
	"github.com/kabroxiko/dayzctl/internal/logger"
	"github.com/kabroxiko/dayzctl/internal/systemd"
	"github.com/spf13/cobra"
)

func ApplyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "apply",
		Short: "Apply configuration and generate systemd units (does NOT restart services)",
		Long: `Apply configuration and generate systemd units.

This command:
  1. Generates server configs (serverDZ-*.cfg, BEServer.cfg)
  2. Generates systemd unit files (dayz@*.service, timers)
  3. Reloads systemd daemon
  4. Enables instances and timers

It does NOT start, stop, or restart any services. Use 'dayzctl instance start/stop/restart' for that.`,
		Run: func(cmd *cobra.Command, args []string) {
			RunCommand(func() error {
				// Generate server configs (serverDZ.cfg, BEServer.cfg)
				logger.Info("Generating server configs...")
				if err := generate.GenerateAll(Config); err != nil {
					return fmt.Errorf("failed to generate server configs: %w", err)
				}

				// Generate systemd units
				sysd := systemd.New()
				logger.Info("Generating systemd units...")
				if err := sysd.GenerateUnits(Config); err != nil {
					return fmt.Errorf("failed to generate units: %w", err)
				}
				if err := sysd.Reload(); err != nil {
					return fmt.Errorf("failed to reload systemd: %w", err)
				}

				// Enable instances (but don't start them)
				for _, instance := range Config.Instances {
					if instance.Enabled {
						logger.Info("Enabling instance", "name", instance.Name)
						if err := sysd.Enable("dayz@" + instance.Name); err != nil {
							logger.Warn("Failed to enable instance", "name", instance.Name, "error", err)
						}
					}
				}

				// Enable timers (but don't start them)
				if Config.Updates.Enabled {
					logger.Info("Enabling update timer...")
					if err := sysd.Enable("dayz-update.timer"); err != nil {
						logger.Warn("Failed to enable update timer", "error", err)
					}
				}

				logger.Info("Configuration applied successfully")
				logger.Info("Services are NOT started/stopped/restarted. Use 'dayzctl instance start/stop/restart' to control services.")
				return nil
			})
		},
	}
}
