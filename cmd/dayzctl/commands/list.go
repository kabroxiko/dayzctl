package commands

import (
	"fmt"

	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/shared"
	"github.com/spf13/cobra"
)

func ListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all instances",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Configured instances:\n")
			for _, inst := range shared.Config.Instances {
				status := "disabled"
				if inst.Enabled {
					status = "enabled"
				}
				fmt.Printf("  %s: port=%d, mods=%d, status=%s\n",
					inst.Name, inst.Port, len(inst.Mods), status)
			}
		},
	}
}
