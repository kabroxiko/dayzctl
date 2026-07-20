package commands

import (
	"fmt"

	"github.com/kabroxiko/dayzctl/internal/logger"
	"github.com/kabroxiko/dayzctl/internal/systemd"
	"github.com/spf13/cobra"
)

func InstanceCmd() *cobra.Command {
	instanceCmd := &cobra.Command{
		Use:   "instance",
		Short: "Manage server instances",
		Long: `Manage server instances.

Commands:
  start [instance]    Start a server instance
  stop [instance]     Stop a server instance
  restart [instance]  Restart a server instance
  status [instance]   Show status of a server instance
  list                List all instances`,
	}

	instanceCmd.AddCommand(
		instanceStartCmd(),
		instanceStopCmd(),
		instanceRestartCmd(),
		instanceStatusCmd(),
		instanceListCmd(),
	)

	return instanceCmd
}

func instanceStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start [instance]",
		Short: "Start a server instance",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			RunCommand(func() error {
				instanceName := args[0]
				if _, err := GetInstance(instanceName); err != nil {
					return err
				}
				sysd := systemd.New()
				if err := sysd.Start("dayz@" + instanceName); err != nil {
					return fmt.Errorf("failed to start %s: %w", instanceName, err)
				}
				logger.Info("Started instance", "name", instanceName)
				return nil
			})
		},
	}
}

func instanceStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop [instance]",
		Short: "Stop a server instance",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			RunCommand(func() error {
				instanceName := args[0]
				if _, err := GetInstance(instanceName); err != nil {
					return err
				}
				sysd := systemd.New()
				if err := sysd.Stop("dayz@" + instanceName); err != nil {
					return fmt.Errorf("failed to stop %s: %w", instanceName, err)
				}
				logger.Info("Stopped instance", "name", instanceName)
				return nil
			})
		},
	}
}

func instanceRestartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart [instance]",
		Short: "Restart a server instance",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			RunCommand(func() error {
				instanceName := args[0]
				if _, err := GetInstance(instanceName); err != nil {
					return err
				}
				sysd := systemd.New()
				if err := sysd.Restart("dayz@" + instanceName); err != nil {
					return fmt.Errorf("failed to restart %s: %w", instanceName, err)
				}
				logger.Info("Restarted instance", "name", instanceName)
				return nil
			})
		},
	}
}

func instanceStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status [instance]",
		Short: "Show status of a server instance",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			RunCommand(func() error {
				instanceName := args[0]
				if _, err := GetInstance(instanceName); err != nil {
					return err
				}
				sysd := systemd.New()
				status, err := sysd.Status("dayz@" + instanceName)
				if err != nil {
					return err
				}
				fmt.Println(status)
				return nil
			})
		},
	}
}

func instanceListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all instances",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Configured instances:\n")
			for _, inst := range Config.Instances {
				status := "disabled"
				if inst.Enabled {
					status = "enabled"
				}
				fmt.Printf("  %s: port=%d, mods=%d, status=%s\n",
					inst.Name, inst.Port, len(inst.Mods), status)
			}
		},
	}
}
