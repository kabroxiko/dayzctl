package commands

import (
	"fmt"

	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/shared"
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

It does NOT start, stop, or restart any services. Use 'dayzctl start/stop/restart' for that.`,
		Run: func(cmd *cobra.Command, args []string) {
			shared.RunCommand(func() error {
				logger.Info("Generating server configs...")
				if err := generate.GenerateAll(shared.Config); err != nil {
					return fmt.Errorf("failed to generate server configs: %w", err)
				}

				sysd := systemd.New()
				logger.Info("Generating systemd units...")
				if err := sysd.GenerateUnits(shared.Config); err != nil {
					return fmt.Errorf("failed to generate units: %w", err)
				}
				if err := sysd.Reload(); err != nil {
					return fmt.Errorf("failed to reload systemd: %w", err)
				}

				for _, instance := range shared.Config.Instances {
					if instance.Enabled {
						logger.Info("Enabling instance", "name", instance.Name)
						if err := sysd.Enable("dayz@" + instance.Name); err != nil {
							logger.Warn("Failed to enable instance", "name", instance.Name, "error", err)
						}
					}
				}

				if shared.Config.Updates.Enabled {
					logger.Info("Enabling update timer...")
					if err := sysd.Enable("dayz-update.timer"); err != nil {
						logger.Warn("Failed to enable update timer", "error", err)
					}
				}

				logger.Info("Configuration applied successfully")
				logger.Info("Services are NOT started/stopped/restarted. Use 'dayzctl start/stop/restart' to control services.")
				return nil
			})
		},
	}
}
