package mods

import (
	"fmt"

	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/shared"
	"github.com/kabroxiko/dayzctl/internal/logger"
	"github.com/spf13/cobra"
)

func AddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add [mod_id] [mod_id...]",
		Short: "Add client mods to an instance",
		Long: `Add one or more client mods to an instance. Client mods are downloaded by players when joining.

Examples:
  dayzctl mods solo add 3765278986
  dayzctl mods solo add 1559212036 1564026768 1646187754`,
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
					for _, m := range instance.Mods {
						if m.ID == modID {
							logger.Info("Mod already in instance (client mods)", "mod", modID)
							found = true
							break
						}
					}
					if !found {
						for _, m := range instance.ServerMods {
							if m.ID == modID {
								logger.Warn("Mod is already in servermods, not adding to mods", "mod", modID)
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

				logger.Info("Adding mods", "count", len(toAdd))

				for _, modID := range toAdd {
					if err := AddModToInstance(instance, modID, false); err != nil {
						logger.Warn("Failed to add mod", "mod_id", modID, "error", err)
					}
				}

				return nil
			})
		},
	}
}
