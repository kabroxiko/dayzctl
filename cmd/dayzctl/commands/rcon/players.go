package rcon

import (
	"fmt"
	"strings"

	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/shared"
	"github.com/kabroxiko/dayzctl/internal/rcon"
	"github.com/spf13/cobra"
)

func PlayersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "players",
		Short: "List players on an instance",
		Run: func(cmd *cobra.Command, args []string) {
			shared.RunCommand(func() error {
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
				players, err := client.Players()
				if err != nil {
					return err
				}

				if len(players) == 0 {
					fmt.Println("No players online")
					return nil
				}

				fmt.Printf("\n=== Players on %s (%d total) ===\n", instance.Name, len(players))
				fmt.Printf("%-4s %-20s %-6s %-8s %-34s %s\n", "ID", "IP", "Port", "Ping", "GUID", "Name")
				fmt.Println(strings.Repeat("-", 80))
				for _, p := range players {
					fmt.Printf("%-4d %-20s %-6s %-8s %-34s %s\n",
						p.ID, p.IP, p.Port, p.Ping, p.GUID, p.Name)
				}
				fmt.Println()
				return nil
			})
		},
	}
}
