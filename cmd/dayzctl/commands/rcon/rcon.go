package rcon

import (
	"context"

	"github.com/spf13/cobra"
)

// Context key for storing instance name
type contextKey string

const InstanceKey contextKey = "instance"

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
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Store the instance name in the context
			ctx := context.WithValue(cmd.Context(), InstanceKey, args[0])
			cmd.SetContext(ctx)
		},
		Run: func(cmd *cobra.Command, args []string) {
			// Show help if no subcommand is provided
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

// GetInstanceFromContext retrieves the instance name from the context
func GetInstanceFromContext(cmd *cobra.Command) string {
	if val := cmd.Context().Value(InstanceKey); val != nil {
		if instance, ok := val.(string); ok {
			return instance
		}
	}
	return ""
}
