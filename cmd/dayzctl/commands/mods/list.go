package mods

import (
	"fmt"

	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/shared"
	"github.com/kabroxiko/dayzctl/internal/mods"
	"github.com/spf13/cobra"
)

func ListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed mods with names",
		Run: func(cmd *cobra.Command, args []string) {
			shared.RunCommand(func() error {
				instanceName := shared.GetInstanceNameFromParent(cmd)
				if instanceName == "" {
					return fmt.Errorf("instance name required")
				}
				_, err := shared.GetInstance(instanceName)
				if err != nil {
					return err
				}
				modManager := mods.New(shared.Config.GetInstallDir(), shared.Config.Paths.WorkshopDir)
				installed, err := modManager.ListInstalled()
				if err != nil {
					return err
				}
				if len(installed) == 0 {
					fmt.Println("No mods installed")
					return nil
				}
				fmt.Printf("Installed mods (%d):\n", len(installed))
				for _, mod := range installed {
					fmt.Printf("  ID: %-10s Name: %s\n", mod.ID, mod.Name)
				}
				return nil
			})
		},
	}
}
