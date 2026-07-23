package mods

import (
	"fmt"

	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands/shared"
	"github.com/kabroxiko/dayzctl/internal/config"
	"github.com/kabroxiko/dayzctl/internal/logger"
	"github.com/kabroxiko/dayzctl/internal/mods"
	"github.com/kabroxiko/dayzctl/internal/steamcmd"
)

func AddModToInstance(instance *config.Instance, modID string, isServerMod bool) error {
	modManager := mods.New(shared.Config.GetInstallDir(), shared.Config.Paths.WorkshopDir)

	logger.Info("Downloading mod", "mod_id", modID)
	if shared.Config.GetSteamcmdBin() == "" {
		return fmt.Errorf("steamcmd path not configured; set 'paths.steamcmd_bin' in the config or install SteamCMD via the installer")
	}
	steam := steamcmd.New(shared.Config.GetSteamUser(), shared.Config.GetInstallDir(), shared.Config.GetSteamcmdBin())
	if err := steam.DownloadMod(modID); err != nil {
		logger.Warn("Failed to download mod", "mod_id", modID, "error", err)
	}

	modInfo, err := modManager.GetModInfo(modID)
	if err != nil {
		logger.Warn("Failed to get mod info, using ID as name", "error", err)
		modInfo = mods.Mod{ID: modID, Name: modID}
	}

	if isServerMod {
		instance.ServerMods = append(instance.ServerMods, config.ModRef{
			ID:   modID,
			Name: modInfo.Name,
		})
		logger.Info("Adding as server mod", "mod", modID, "name", modInfo.Name)
	} else {
		instance.Mods = append(instance.Mods, config.ModRef{
			ID:   modID,
			Name: modInfo.Name,
		})
		logger.Info("Adding as client mod", "mod", modID, "name", modInfo.Name)
	}

	if err := shared.SaveConfig(); err != nil {
		logger.Warn("Failed to save config", "error", err)
	}

	if err := modManager.SyncMods(instance.Mods, instance.ServerMods); err != nil {
		logger.Warn("Failed to sync mods", "error", err)
	}

	if err := shared.UpdateServerConfig(instance); err != nil {
		logger.Warn("Failed to update server config", "error", err)
	}

	if err := shared.ApplyConfig(); err != nil {
		logger.Warn("Failed to apply config changes", "error", err)
	}

	modType := "client"
	if isServerMod {
		modType = "server"
	}
	logger.Info("Added mod to instance", "instance", instance.Name, "mod", modID, "name", modInfo.Name, "type", modType)
	return nil
}
