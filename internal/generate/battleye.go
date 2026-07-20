package generate

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/kabroxiko/dayzctl/internal/config"
)

// GenerateBattlEyeConfig generates BEServer.cfg for all instances
func GenerateBattlEyeConfig(cfg *config.ServerConfig, tmplContent string) error {
	tmpl, err := template.New("BEServer").Parse(tmplContent)
	if err != nil {
		return fmt.Errorf("failed to parse BEServer template: %w", err)
	}

	installDir := cfg.GetInstallDir()

	for _, instance := range cfg.GetEnabledInstances() {
		data := buildServerData(cfg, instance)

		if err := generateInstanceBattlEye(installDir, instance, data, tmpl); err != nil {
			return fmt.Errorf("failed to generate battleye config for %s: %w", instance.Name, err)
		}
	}

	return nil
}

// generateInstanceBattlEye generates BattlEye config for a single instance
func generateInstanceBattlEye(installDir string, instance config.Instance, data ServerConfigData, tmpl *template.Template) error {
	beDir := filepath.Join(installDir, "battleye-"+instance.Name)

	if err := os.MkdirAll(beDir, 0755); err != nil {
		return fmt.Errorf("failed to create battleye directory: %w", err)
	}

	// Generate config content
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to generate battleye config: %w", err)
	}
	content := buf.Bytes()

	// Write BEServer.cfg (Linux primary - case sensitive!)
	if err := os.WriteFile(filepath.Join(beDir, "BEServer.cfg"), content, 0644); err != nil {
		return fmt.Errorf("failed to write BEServer.cfg: %w", err)
	}

	// Write beserver_x64.cfg (for compatibility)
	if err := os.WriteFile(filepath.Join(beDir, "beserver_x64.cfg"), content, 0644); err != nil {
		return fmt.Errorf("failed to write beserver_x64.cfg: %w", err)
	}

	// Create symlink for .so file (beserver_x64.so - lowercase!)
	linkPath := filepath.Join(beDir, "beserver_x64.so")

	// Remove existing file/link if it exists
	if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing symlink: %w", err)
	}

	// Create symlink to main battleye directory
	if err := os.Symlink("../battleye/beserver_x64.so", linkPath); err != nil {
		return fmt.Errorf("failed to create symlink for beserver_x64.so: %w", err)
	}

	return nil
}
