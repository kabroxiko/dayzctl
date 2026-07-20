package commands

import (
	"fmt"
	"strings"

	"github.com/kabroxiko/dayzctl/internal/rcon"
	"github.com/spf13/cobra"
)

func PlayersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "players [instance]",
		Short: "List players on a server instance",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			RunCommand(func() error {
				instance, err := GetInstance(args[0])
				if err != nil {
					return err
				}
				if !instance.RCON.Enabled {
					return fmt.Errorf("RCON is not enabled for instance: %s", instance.Name)
				}

				client := rcon.New(instance.RCON.Port, instance.RCON.Password)
				response, err := client.Send("players")
				if err != nil {
					return err
				}

				if response == "" || strings.Contains(response, "There are no players") {
					fmt.Println("No players online")
					return nil
				}

				lines := strings.Split(strings.TrimSpace(response), "\n")
				if len(lines) == 0 {
					fmt.Println("No players online")
					return nil
				}

				fmt.Printf("\n=== Players on %s ===\n", instance.Name)
				for _, line := range lines {
					if strings.TrimSpace(line) != "" {
						fmt.Println(line)
					}
				}
				fmt.Println()
				return nil
			})
		},
	}
}
