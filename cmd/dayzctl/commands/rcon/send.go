package rcon

import (
	"fmt"
	"strings"

	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/shared"
	"github.com/kabroxiko/dayzctl/internal/rcon"
	"github.com/spf13/cobra"
)

func SendCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "send [command]",
		Short: "Send raw RCON command to an instance",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			shared.RunCommand(func() error {
				parent := cmd.Parent()
				if parent == nil {
					return fmt.Errorf("parent command not found")
				}
				parentArgs := parent.Flags().Args()
				if len(parentArgs) == 0 {
					return fmt.Errorf("instance name required. Usage: dayzctl rcon <instance> send <command>")
				}
				instanceName := parentArgs[0]

				instance, err := shared.GetInstance(instanceName)
				if err != nil {
					return err
				}
				if !instance.RCON.Enabled {
					return fmt.Errorf("RCON is not enabled for instance: %s", instance.Name)
				}

				client := rcon.New(instance.RCON.Port, instance.RCON.Password)
				response, err := client.Send(strings.Join(args, " "))
				if err != nil {
					return err
				}
				if response != "" {
					fmt.Println(response)
				}
				return nil
			})
		},
	}
}
