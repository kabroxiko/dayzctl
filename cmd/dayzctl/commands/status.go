package commands

import (
	"fmt"
	"strings"

	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/shared"
	"github.com/kabroxiko/dayzctl/internal/systemd"
	"github.com/spf13/cobra"
)

func StatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status [instance|all]",
		Short: "Show status of server instance(s)",
		Long: `Show status of a server instance or all instances.
		
If no argument is provided, shows overall server status.
Use 'all' to show status of all instances.`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			shared.RunCommand(func() error {
				sysd := systemd.New()

				if len(args) == 0 || args[0] == "all" {
					fmt.Print("\n=== DayZ Server Status ===\n\n")

					fmt.Printf("Configured instances: %v\n", shared.Config.GetInstanceNames())

					running, _ := sysd.ListRunningInstances()
					fmt.Printf("Running instances: %v\n", running)

					for _, inst := range shared.Config.Instances {
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
				} else {
					instance, err := shared.GetInstance(args[0])
					if err != nil {
						return err
					}
					status, err := sysd.Status("dayz@" + instance.Name)
					if err != nil {
						return err
					}
					fmt.Println(status)
				}
				return nil
			})
		},
	}
}
