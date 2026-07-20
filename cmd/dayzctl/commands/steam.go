package commands

import (
	"github.com/kabroxiko/dayzctl/internal/steamcmd"
	"github.com/spf13/cobra"
)

func SteamLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "steam-login",
		Short: "Interactive Steam login as dayz user",
		Run: func(cmd *cobra.Command, args []string) {
			RunCommand(func() error {
				steam := steamcmd.New(Config.GetSteamUser(), Config.GetInstallDir(), Config.Paths.SteamcmdBin)
				return steam.InteractiveLogin()
			})
		},
	}
}
