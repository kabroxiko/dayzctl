package mods

import (
	"fmt"

	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/shared"
	"github.com/kabroxiko/dayzctl/internal/logger"
	"github.com/spf13/cobra"
)

func AddServerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add-server [mod_id] [mod_id...]",
		Short: "Add server-side mods to an instance",
		Long: `Add one or more server-side mods to an instance. Server-side mods are NOT downloaded by players.
These are typically mods that only affect server behavior.

Examples:
  dayzctl mods solo add-server 3765278986
  dayzctl mods solo add-server 1559212036 1564026768`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			shared.RunCommand(func() error {
				instanceName := shared.GetInstanceNameFromParent(cmd)
				if instanceName == "" {
					return fmt.Errorf("instance name required")
				}
				instance, err := shared.GetInstance(instanceName)
				if err != nil {
					return err
				}
				modIDs := args

				var toAdd []string
				for _, modID := range modIDs {
					found := false
					for _, m := range instance.ServerMods {
						if m.ID == modID {
							logger.Info("Mod already in instance (server mods)", "mod", modID)
							found = true
							break
						}
					}
					if !found {
						for _, m := range instance.Mods {
							if m.ID == modID {
								logger.Warn("Mod is already in mods, not adding to servermods", "mod", modID)
								found = true
								break
							}
						}
					}
					if !found {
						toAdd = append(toAdd, modID)
					}
				}

				if len(toAdd) == 0 {
					logger.Info("No new mods to add")
					return nil
				}

				logger.Info("Adding server mods", "count", len(toAdd))

				for _, modID := range toAdd {
					if err := AddModToInstance(instance, modID, true); err != nil {
						logger.Warn("Failed to add server mod", "mod_id", modID, "error", err)
					}
				}

				return nil
			})
		},
	}
}
