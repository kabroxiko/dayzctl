package commands

import (
	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/shared"
	"github.com/kabroxiko/dayzctl/internal/logger"
	"github.com/kabroxiko/dayzctl/internal/systemd"
	"github.com/spf13/cobra"
)

func RestartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart [instance|all]",
		Short: "Restart a server instance or all instances",
		Long: `Restart a server instance or all instances.
		
Use 'all' to restart all running instances. 'all' is a reserved keyword and cannot be used as an instance name.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			shared.RunCommand(func() error {
				instances, err := shared.GetInstances(args[0])
				if err != nil {
					return err
				}

				sysd := systemd.New()
				for _, instance := range instances {
					if err := sysd.Restart("dayz@" + instance.Name); err != nil {
						logger.Warn("Failed to restart instance", "name", instance.Name, "error", err)
						continue
					}
					logger.Info("Restarted instance", "name", instance.Name)
				}
				return nil
			})
		},
	}
}
