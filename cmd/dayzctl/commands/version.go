package commands

import (
	"fmt"

	"github.com/kabroxiko/dayzctl/internal/version"
	"github.com/spf13/cobra"
)

func VersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("dayzctl version %s (built %s)\n", version.Version, version.BuildTime)
		},
	}
}
