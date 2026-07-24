package rcon

import (
	"github.com/spf13/cobra"
)

var instanceName string

func RconCmd() *cobra.Command {
	rconCmd := &cobra.Command{
		Use:   "rcon [instance]",
		Short: "RCON commands for an instance",
		Long: `RCON commands for a specific instance.

Usage:
  dayzctl rcon <instance> <command>

Commands:
  send <command>     Send raw RCON command
  players            List players
  kick <player>      Kick a player
  ban <player>       Ban a player
  say <message>      Send global message

Examples:
  dayzctl rcon solo players
  dayzctl rcon solo send status
  dayzctl rcon solo kick PlayerName`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// args[0] is the instance name
			if len(args) > 0 {
				instanceName = args[0]
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			// If no subcommand is given, show help
			cmd.Help()
		},
	}

	rconCmd.AddCommand(
		SendCmd(),
		PlayersCmd(),
		KickCmd(),
		BanCmd(),
		SayCmd(),
	)

	return rconCmd
}
