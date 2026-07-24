package rcon

import (
	"fmt"
	"os"

	"github.com/kabroxiko/dayzctl/internal/logger"

	"github.com/kabroxiko/dayzctl/internal/systemd"
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
			// debug log
			logger.Debug("rcon.PersistentPreRun args", "args", args, "osArgs", os.Args)
			// Prefer Cobra-provided args when available
			if len(args) > 0 {
				instanceName = args[0]
				return
			}

			// Fallback: parse os.Args to find the token after "rcon"
			// Skip known subcommand names so we don't confuse them with instance names
			known := map[string]struct{}{"send": {}, "players": {}, "kick": {}, "ban": {}, "say": {}, "help": {}, "-h": {}, "--help": {}}
			argv := os.Args
			for i := 0; i < len(argv); i++ {
				if argv[i] == "rcon" && i+1 < len(argv) {
					cand := argv[i+1]
					if _, ok := known[cand]; !ok && len(cand) > 0 && cand[0] != '-' {
						instanceName = cand
					}
					break
				}
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			// If no subcommand is given, show help
			if err := cmd.Help(); err != nil {
				logger.Error("failed to show help", "error", err)
			}
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

// isInstanceRunning checks systemd for the running instance and returns an error if it's not active.
func isInstanceRunning(name string) error {
	s := systemd.New()
	running, err := s.ListRunningInstances()
	if err != nil {
		// If we can't determine running instances, return a descriptive error
		return fmt.Errorf("failed to determine instance state: %w", err)
	}
	for _, n := range running {
		if n == name {
			return nil
		}
	}
	return fmt.Errorf("instance is not running: %s", name)
}
