package rcon

import (
	"github.com/spf13/cobra"
)

func RconCmd() *cobra.Command {
	rconCmd := &cobra.Command{
		Use:   "rcon [instance]",
		Short: "RCON commands for an instance",
		Long: `RCON commands for a specific instance.

Commands:
  rcon <instance> send <command>     Send raw RCON command
  rcon <instance> players            List players
  rcon <instance> kick <player>      Kick a player
  rcon <instance> ban <player>       Ban a player
  rcon <instance> say <message>      Send global message`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Store the instance name for subcommands
			cmd.SetContext(cmd.Context())
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
