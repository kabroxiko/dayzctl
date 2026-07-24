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
				// Expect instance as first arg after rewrite in ExecuteWithArgs
				inst := ""
				if len(args) > 0 {
					inst = args[0]
					// shift command args so the actual command follows
					args = args[1:]
				}
				if inst == "" {
					inst = shared.GetInstanceNameFromCommandChain(cmd)
				}
				if inst == "" {
					inst = instanceName
				}
				if inst == "" {
					return fmt.Errorf("instance name required. Usage: dayzctl rcon <instance> send <command>")
				}

				instance, err := shared.GetInstance(inst)
				if err != nil {
					return err
				}
				if err := isInstanceRunning(instance.Name); err != nil {
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

// SendAction sends a raw RCON command to the named instance.
// args are the arguments following the instance (e.g., the command to send).
func SendAction(inst string, args []string) error {
	if inst == "" {
		return fmt.Errorf("instance name required. Usage: dayzctl rcon <instance> send <command>")
	}
	instance, err := shared.GetInstance(inst)
	if err != nil {
		return err
	}
	if err := isInstanceRunning(instance.Name); err != nil {
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
}
