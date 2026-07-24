package rcon

import (
	"fmt"
	"strings"

	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/shared"
	"github.com/kabroxiko/dayzctl/internal/rcon"
	"github.com/spf13/cobra"
)

func SayCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "say [message]",
		Short: "Send a global chat message to all players",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			shared.RunCommand(func() error {
				inst := shared.GetInstanceNameFromCommandChain(cmd)
				if inst == "" {
					inst = instanceName
				}
				if inst == "" {
					return fmt.Errorf("instance name required. Usage: dayzctl rcon <instance> say <message>")
				}

				instance, err := shared.GetInstance(inst)
				if err != nil {
					return err
				}
				if !instance.RCON.Enabled {
					return fmt.Errorf("RCON is not enabled for instance: %s", instance.Name)
				}

				client := rcon.New(instance.RCON.Port, instance.RCON.Password)
				message := strings.Join(args, " ")
				response, err := client.Say(message)
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
