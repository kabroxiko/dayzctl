package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kabroxiko/dayzctl/internal/rcon"
	"github.com/spf13/cobra"
)

func RconCmd() *cobra.Command {
	rconCmd := &cobra.Command{
		Use:   "rcon",
		Short: "RCON commands",
	}

	rconCmd.AddCommand(
		rconSendCmd(),
		rconPlayersCmd(),
		rconKickCmd(),
		rconBanCmd(),
		rconSayCmd(),
	)

	return rconCmd
}

func rconSendCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "send [instance] [command]",
		Short: "Send raw RCON command to an instance",
		Args:  cobra.MinimumNArgs(2),
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
				response, err := client.Send(strings.Join(args[1:], " "))
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

func rconPlayersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "players [instance]",
		Short: "List players on an instance",
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

func rconKickCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "kick [instance] [player] [reason]",
		Short: "Kick a player by ID or name",
		Args:  cobra.MinimumNArgs(2),
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
				player := args[1]

				// Try to find player ID if name is provided
				playerID, err := strconv.Atoi(player)
				if err != nil {
					// Find by name
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
				if len(args) > 2 {
					reason = strings.Join(args[2:], " ")
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

func rconBanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ban [instance] [player] [minutes] [reason]",
		Short: "Ban a player by ID or name",
		Args:  cobra.MinimumNArgs(2),
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
				player := args[1]

				// Try to find player ID
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
				if len(args) > 2 {
					minutes, _ = strconv.Atoi(args[2])
				}

				reason := ""
				if len(args) > 3 {
					reason = strings.Join(args[3:], " ")
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

func rconSayCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "say [instance] [message]",
		Short: "Send a global chat message to all players",
		Args:  cobra.MinimumNArgs(2),
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
				message := strings.Join(args[1:], " ")
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
