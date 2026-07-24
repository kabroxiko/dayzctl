package rcon

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/shared"
	"github.com/kabroxiko/dayzctl/internal/rcon"
	"github.com/spf13/cobra"
)

func KickCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "kick [player] [reason]",
		Short: "Kick a player by ID or name",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			shared.RunCommand(func() error {
				// Expect instance as first arg after rewrite in ExecuteWithArgs
				inst := ""
				if len(args) > 0 {
					inst = args[0]
					// shift remaining args so player argument is first
					args = args[1:]
				}
				if inst == "" {
					inst = shared.GetInstanceNameFromCommandChain(cmd)
				}
				if inst == "" {
					inst = instanceName
				}
				if inst == "" {
					return fmt.Errorf("instance name required. Usage: dayzctl rcon <instance> kick <player>")
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
				player := args[0]

				playerID, err := strconv.Atoi(player)
				if err != nil {
					players, err := client.Players()
					if err != nil {
						return err
					}
					for _, p := range players {
						if strings.EqualFold(p.Name, player) {
							playerID = p.ID
							break
						}
					}
					if playerID == 0 {
						return fmt.Errorf("player not found: %s", player)
					}
				}

				reason := ""
				if len(args) > 1 {
					reason = strings.Join(args[1:], " ")
				}

				response, err := client.Kick(playerID, reason)
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

// KickAction kicks a player on the given instance. args are the args after the instance.
func KickAction(inst string, args []string) error {
	if inst == "" {
		return fmt.Errorf("instance name required. Usage: dayzctl rcon <instance> kick <player>")
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
	if len(args) == 0 {
		return fmt.Errorf("player required")
	}
	player := args[0]

	playerID, perr := strconv.Atoi(player)
	if perr != nil {
		players, err := client.Players()
		if err != nil {
			return err
		}
		for _, p := range players {
			if strings.EqualFold(p.Name, player) {
				playerID = p.ID
				break
			}
		}
		if playerID == 0 {
			return fmt.Errorf("player not found: %s", player)
		}
	}

	reason := ""
	if len(args) > 1 {
		reason = strings.Join(args[1:], " ")
	}

	response, err := client.Kick(playerID, reason)
	if err != nil {
		return err
	}
	if response != "" {
		fmt.Println(response)
	}
	return nil
}
