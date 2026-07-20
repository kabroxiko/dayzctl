package generate

import (
	_ "embed"
	"fmt"

	"github.com/kabroxiko/dayzctl/internal/config"
)

//go:embed templates/serverDZ.cfg.tmpl
var serverDZTemplate string

//go:embed templates/BEServer.cfg.tmpl
var beserverTemplate string

// GenerateAll generates all server config files
func GenerateAll(cfg *config.ServerConfig) error {
	if err := GenerateServerConfig(cfg, serverDZTemplate); err != nil {
		return fmt.Errorf("failed to generate server configs: %w", err)
	}

	if err := GenerateBattlEyeConfig(cfg, beserverTemplate); err != nil {
		return fmt.Errorf("failed to generate battleye configs: %w", err)
	}

	return nil
}
