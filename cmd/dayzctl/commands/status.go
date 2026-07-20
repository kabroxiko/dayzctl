package commands

import (
	"fmt"
	"strings"

	"github.com/kabroxiko/dayzctl/internal/systemd"
	"github.com/spf13/cobra"
)

func StatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show overall server status",
		Run: func(cmd *cobra.Command, args []string) {
			RunCommand(func() error {
				sysd := systemd.New()

				fmt.Print("\n=== DayZ Server Status ===\n\n")

				fmt.Printf("Configured instances: %v\n", Config.GetInstanceNames())

				running, _ := sysd.ListRunningInstances()
				fmt.Printf("Running instances: %v\n", running)

				for _, inst := range Config.Instances {
					if !inst.Enabled {
						continue
					}
					isRunning := false
					for _, r := range running {
						if r == inst.Name {
							isRunning = true
							break
						}
					}
					status := "stopped"
					if isRunning {
						status = "running"
					}
					fmt.Printf("\nInstance: %s\n", inst.Name)
					fmt.Printf("  Port: %d\n", inst.Port)
					fmt.Printf("  Status: %s\n", status)
					fmt.Printf("  Mods: %d\n", len(inst.Mods))
					if inst.RCON.Enabled {
						fmt.Printf("  RCON: enabled on port %d\n", inst.RCON.Port)
					}
				}

				fmt.Print("\n=== Timers ===\n")
				for _, timer := range []string{"dayz-update.timer", "dayz-prune.timer"} {
					status, _ := sysd.Status(timer)
					for _, line := range strings.Split(status, "\n") {
						if strings.Contains(line, "Active:") || strings.Contains(line, "Trigger:") {
							fmt.Println("  " + strings.TrimSpace(line))
						}
					}
				}
				fmt.Println()
				return nil
			})
		},
	}
}
