package rcon

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/shared"
	"github.com/kabroxiko/dayzctl/internal/rcon"
	"github.com/spf13/cobra"
)

func BanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ban [player] [minutes] [reason]",
		Short: "Ban a player by ID or name",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			shared.RunCommand(func() error {
				parent := cmd.Parent()
				if parent == nil {
					return fmt.Errorf("parent command not found")
				}
				parentArgs := parent.Flags().Args()
				if len(parentArgs) == 0 {
					return fmt.Errorf("instance name required. Usage: dayzctl rcon <instance> ban <player>")
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

				minutes := 0
				if len(args) > 1 {
					minutes, _ = strconv.Atoi(args[1])
				}

				reason := ""
				if len(args) > 2 {
					reason = strings.Join(args[2:], " ")
				}

				response, err := client.Ban(playerID, minutes, reason)
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
