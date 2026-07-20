package commands

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func CheckSpaceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check-space",
		Short: "Check available disk space",
		Run: func(cmd *cobra.Command, args []string) {
			RunCommand(func() error {
				installDir := Config.GetInstallDir()

				execCmd := exec.Command("df", "-BG", installDir)
				output, err := execCmd.Output()
				if err != nil {
					return err
				}

				lines := strings.Split(string(output), "\n")
				if len(lines) < 2 {
					return fmt.Errorf("unexpected df output")
				}

				fields := strings.Fields(lines[1])
				if len(fields) < 4 {
					return fmt.Errorf("unexpected df output format")
				}

				available := fields[3]
				usePercent := fields[4]

				fmt.Printf("Disk usage for %s:\n", installDir)
				fmt.Printf("  Available: %s\n", available)
				fmt.Printf("  Used: %s\n", usePercent)

				return nil
			})
		},
	}
}
