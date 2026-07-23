package mods

import (
	"github.com/spf13/cobra"
)

func ModsCmd() *cobra.Command {
	modsCmd := &cobra.Command{
		Use:   "mods [instance|all]",
		Short: "Manage mods for an instance or all instances",
		Long: `Manage mods for DayZ server instances.

Commands:
  mods <instance> list                  List all installed mods
  mods <instance> add <id> [id...]     Add client mods
  mods <instance> add-server <id> [id...]  Add server-side mods
  mods <instance> remove <id> [id...]  Remove mods from config
  mods <instance> delete <id> [id...]  Delete mods (config + symlinks)
  mods <instance> sync                  Sync mods for an instance`,
		Args: cobra.ExactArgs(1),
	}

	modsCmd.AddCommand(
		ListCmd(),
		SyncCmd(),
		AddCmd(),
		AddServerCmd(),
		RemoveCmd(),
		DeleteCmd(),
	)

	return modsCmd
}
