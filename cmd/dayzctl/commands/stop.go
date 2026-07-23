package commands

import (
	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/shared"
	"github.com/kabroxiko/dayzctl/internal/logger"
	"github.com/kabroxiko/dayzctl/internal/systemd"
	"github.com/spf13/cobra"
)

func StopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop [instance|all]",
		Short: "Stop a server instance or all instances",
		Long: `Stop a server instance or all instances.
		
Use 'all' to stop all running instances. 'all' is a reserved keyword and cannot be used as an instance name.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			shared.RunCommand(func() error {
				instances, err := shared.GetInstances(args[0])
				if err != nil {
					return err
				}
				
				sysd := systemd.New()
				for _, instance := range instances {
					if err := sysd.Stop("dayz@" + instance.Name); err != nil {
						logger.Warn("Failed to stop instance", "name", instance.Name, "error", err)
						continue
					}
					logger.Info("Stopped instance", "name", instance.Name)
				}
				return nil
			})
		},
	}
}
