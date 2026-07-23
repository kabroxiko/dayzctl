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
				instanceName := shared.GetInstanceNameFromParent(cmd)
				if instanceName == "" {
					return fmt.Errorf("instance name required")
				}
				instance, err := shared.GetInstance(instanceName)
				if err != nil {
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
