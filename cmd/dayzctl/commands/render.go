package commands

import (
	"fmt"

	"github.com/kabroxiko/dayzctl/internal/logger"
	"github.com/spf13/cobra"
)

func RenderConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "render-config [instance]",
		Short: "Render server config for an instance",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			RunCommand(func() error {
				instance, err := GetInstance(args[0])
				if err != nil {
					return err
				}

				logger.Info("Rendering config for instance", "name", instance.Name)
				configPath := fmt.Sprintf("%s/serverDZ-%s.cfg", Config.GetInstallDir(), instance.Name)
				logger.Info("Config would be written to", "path", configPath)

				return nil
			})
		},
	}
}
