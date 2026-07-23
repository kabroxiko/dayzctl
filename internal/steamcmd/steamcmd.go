package steamcmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/kabroxiko/dayzctl/internal/logger"
	"github.com/kabroxiko/dayzctl/internal/utils"
)

// Common errors
var (
	ErrRateLimited = errors.New("steam rate limited")
	ErrLoginFailed = errors.New("steam login failed")
)

// SteamCmd represents a SteamCMD instance
type SteamCmd struct {
	User         string
	InstallDir   string
	SteamCmdPath string
	WorkshopDir  string
	lastAttempt  time.Time
	mu           sync.Mutex
}

// New creates a new SteamCMD instance
func New(user, installDir, steamCmdPath string) *SteamCmd {
	workshop := filepath.Join(installDir, "workshop")
	return &SteamCmd{
		User:         user,
		InstallDir:   installDir,
		SteamCmdPath: steamCmdPath,
		WorkshopDir:  workshop,
	}
}

// runSteamCmd runs a steamcmd command as the dayz user
func (s *SteamCmd) runSteamCmd(args ...string) error {
	if s.SteamCmdPath == "" {
		return fmt.Errorf("steamcmd binary path not configured")
	}
	cmdStr := fmt.Sprintf("%s %s", s.SteamCmdPath, strings.Join(args, " "))
	logger.Debug("Executing steamcmd", "cmd", cmdStr, "user", s.User, "installDir", s.InstallDir)
	logger.Info("Running steamcmd command", "cmd", cmdStr)
	
	cmd := exec.Command("runuser", "-u", "dayz", "--", "sh", "-c", cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logger.Warn("steamcmd run failed", "cmd", cmdStr, "error", err)
		return err
	}
	return nil
}

// runSteamCmdWithOutput runs a steamcmd command and returns output
func (s *SteamCmd) runSteamCmdWithOutput(args ...string) (string, error) {
	if s.SteamCmdPath == "" {
		return "", fmt.Errorf("steamcmd binary path not configured")
	}
	cmdStr := fmt.Sprintf("%s %s", s.SteamCmdPath, strings.Join(args, " "))
	logger.Debug("Executing steamcmd (with output)", "cmd", cmdStr, "user", s.User, "installDir", s.InstallDir)
	logger.Info("Running steamcmd with output", "cmd", cmdStr)
	
	cmd := exec.Command("runuser", "-u", "dayz", "--", "sh", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	outStr := string(output)
	
	lines := strings.Split(outStr, "\n")
	if len(lines) > 10 {
		logger.Debug("steamcmd output (first 10 lines)", "output", strings.Join(lines[:10], "\n"))
	} else {
		logger.Debug("steamcmd output", "output", outStr)
	}
	
	if err != nil {
		logger.Warn("steamcmd returned error", "cmd", cmdStr, "error", err, "output", outStr)
	} else {
		logger.Debug("steamcmd completed successfully", "cmd", cmdStr)
	}
	return outStr, err
}

// GetBuildID retrieves the current build ID from Steam
func (s *SteamCmd) GetBuildID() (string, error) {
	logger.Info("Fetching build ID from Steam...")
	output, err := s.runSteamCmdWithOutput(
		"+@sSteamCmdForcePlatformType", "linux",
		"+login", "anonymous",
		"+app_info_print", "223350",
		"+quit",
	)
	if err != nil {
		if strings.Contains(output, "Rate Limit Exceeded") {
			return "", ErrRateLimited
		}
		return "", fmt.Errorf("steamcmd failed: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, `"buildid"`) {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				buildID := strings.Trim(parts[1], `"`)
				logger.Info("Retrieved build ID", "build_id", buildID)
				return buildID, nil
			}
		}
	}
	return "", fmt.Errorf("buildid not found in output")
}

// getCurrentLocalBuildID reads the current local build ID
func (s *SteamCmd) getCurrentLocalBuildID() (string, error) {
	appManifestPath := filepath.Join(s.InstallDir, "steamapps", "appmanifest_223350.acf")
	logger.Debug("Checking local build ID", "path", appManifestPath)
	
	if _, err := os.Stat(appManifestPath); err == nil {
		data, err := os.ReadFile(appManifestPath)
		if err == nil {
			content := string(data)
			re := regexp.MustCompile(`"buildid"\s+"(\d+)"`)
			matches := re.FindStringSubmatch(content)
			if len(matches) > 1 {
				buildID := matches[1]
				logger.Info("Local build ID from manifest", "build_id", buildID)
				return buildID, nil
			}
		}
	} else {
		logger.Warn("App manifest not found, server not installed", "path", appManifestPath)
		return "", nil
	}
	
	return "", nil
}

// NeedsUpdate checks if the server needs an update
func (s *SteamCmd) NeedsUpdate() (bool, error) {
	logger.Info("Checking for updates...")
	
	// Check if server is installed FIRST
	localBuildID, err := s.getCurrentLocalBuildID()
	if err != nil {
		logger.Warn("Error reading local build ID", "error", err)
		return true, nil
	}
	
	// If no local build ID, server is not installed - force update
	if localBuildID == "" {
		logger.Info("Server not installed, update required")
		return true, nil
	}
	
	// Only get latest build ID if server is installed
	latestBuildID, err := s.GetBuildID()
	if err != nil {
		if errors.Is(err, ErrRateLimited) {
			return false, err
		}
		return false, err
	}
	
	logger.Debug("Comparing builds", "local", localBuildID, "latest", latestBuildID)
	needsUpdate := localBuildID != latestBuildID
	logger.Info("Update status", "needs_update", needsUpdate, "local_build", localBuildID, "latest_build", latestBuildID)
	return needsUpdate, nil
}

// Update updates the DayZ server
func (s *SteamCmd) Update() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRateLimited() {
		return ErrRateLimited
	}

	logger.Info("Starting server update...")
	logger.Info("SteamCMD path", "path", s.SteamCmdPath)
	logger.Info("Install directory", "path", s.InstallDir)
	logger.Info("Steam user", "user", s.User)
	
	err := s.runSteamCmd(
		"+@sSteamCmdForcePlatformType", "linux",
		"+force_install_dir", s.InstallDir,
		"+login", s.User,
		"+app_update", "223350", "validate",
		"+quit",
	)
	
	s.lastAttempt = time.Now()
	if err != nil {
		logger.Error("Update failed", "error", err)
		return fmt.Errorf("update failed: %w", err)
	}
	
	logger.Info("Server update completed successfully")
	return nil
}

// DownloadMod downloads a mod using anonymous login
func (s *SteamCmd) DownloadMod(modID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRateLimited() {
		return ErrRateLimited
	}

	logger.Info("Downloading mod", "mod_id", modID)
	logger.Debug("Mod download settings", "workshopDir", s.WorkshopDir, "installDir", s.InstallDir)

	if err := os.MkdirAll(s.WorkshopDir, 0755); err != nil {
		return fmt.Errorf("failed to create workshop directory: %w", err)
	}

	output, err := s.runSteamCmdWithOutput(
		"+@sSteamCmdForcePlatformType", "linux",
		"+login", "anonymous",
		"+force_install_dir", s.WorkshopDir,
		"+workshop_download_item", "221100", modID, "validate",
		"+quit",
	)

	s.lastAttempt = time.Now()
	if err != nil {
		if strings.Contains(output, "Rate Limit Exceeded") {
			return ErrRateLimited
		}
		return fmt.Errorf("mod download failed: %w\nOutput: %s", err, output)
	}

	if strings.Contains(output, "Success.") || strings.Contains(output, "already downloaded") {
		logger.Info("Mod downloaded successfully", "mod_id", modID)
		return s.linkMod(modID)
	}

	return fmt.Errorf("mod %s download failed - check output", modID)
}

// linkMod creates the symlink for a mod in the workshop directory
func (s *SteamCmd) linkMod(modID string) error {
	possiblePaths := []string{
		filepath.Join(s.WorkshopDir, "steamapps", "workshop", "content", "221100", modID),
		filepath.Join(s.WorkshopDir, "content", "221100", modID),
		filepath.Join(s.WorkshopDir, modID),
		filepath.Join("/home/dayz/Steam/steamapps/workshop/content/221100", modID),
		filepath.Join("/home/dayz/.steam/steamapps/workshop/content/221100", modID),
		filepath.Join(s.InstallDir, "steamapps", "workshop", "content", "221100", modID),
	}

	var srcPath string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			srcPath = path
			logger.Debug("Found mod at", "path", path)
			break
		}
	}

	if srcPath == "" {
		return fmt.Errorf("mod %s not found in any workshop location", modID)
	}

	targetPath := filepath.Join(s.WorkshopDir, modID)
	
	if _, err := os.Lstat(targetPath); err == nil {
		if err := os.Remove(targetPath); err != nil {
			return fmt.Errorf("failed to remove existing symlink: %w", err)
		}
	}

	if err := os.Symlink(srcPath, targetPath); err != nil {
		return fmt.Errorf("failed to create symlink from %s to %s: %w", targetPath, srcPath, err)
	}

	if err := utils.ChownSymlink(targetPath); err != nil {
		logger.Warn("Failed to chown workshop symlink", "path", targetPath, "error", err)
	}

	logger.Info("Mod linked successfully", "mod_id", modID, "path", targetPath)
	return nil
}

// InteractiveLogin performs an interactive Steam login
func (s *SteamCmd) InteractiveLogin() error {
	cmdStr := fmt.Sprintf("%s +login %s +quit", s.SteamCmdPath, s.User)
	logger.Info("Starting interactive Steam login", "cmd", cmdStr, "user", s.User)
	
	cmd := exec.Command(
		"runuser", "-u", "dayz", "--",
		"sh", "-c", cmdStr,
	)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logger.Warn("interactive login failed", "cmd", cmdStr, "error", err)
		return fmt.Errorf("interactive login failed: %w", err)
	}
	return nil
}

// isRateLimited checks if we're rate limited
func (s *SteamCmd) isRateLimited() bool {
	return time.Since(s.lastAttempt) < 5*time.Minute
}

// IsRateLimitError checks if an error is a rate limit error
func IsRateLimitError(err error) bool {
	return errors.Is(err, ErrRateLimited)
}
