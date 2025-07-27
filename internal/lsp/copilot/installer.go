package copilot

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kirmad/superopencode/internal/config"
	"github.com/kirmad/superopencode/internal/logging"
)

// Installer manages the installation of the GitHub Copilot Language Server
type Installer struct {
	config *config.CopilotConfig
}

// NewInstaller creates a new installer
func NewInstaller(cfg *config.CopilotConfig) *Installer {
	return &Installer{
		config: cfg,
	}
}

// EnsureInstalled ensures the Copilot server is installed
func (i *Installer) EnsureInstalled(ctx context.Context) error {
	// Check if server is already available
	if serverPath, err := i.GetServerPath(); err == nil {
		if i.verifyInstallation(serverPath) {
			logging.Info("Copilot server already installed", "path", serverPath)
			return nil
		}
	}
	
	// Install the server
	return i.Install(ctx)
}

// Install installs the GitHub Copilot Language Server
func (i *Installer) Install(ctx context.Context) error {
	logging.Info("Installing GitHub Copilot Language Server")
	
	// Check if npm is available
	if _, err := exec.LookPath("npm"); err != nil {
		return fmt.Errorf("npm is required but not found in PATH")
	}
	
	// Install the package globally
	cmd := exec.CommandContext(ctx, "npm", "install", "-g", "@github/copilot-language-server")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install copilot-language-server: %w", err)
	}
	
	// Verify installation
	serverPath, err := i.GetServerPath()
	if err != nil {
		return fmt.Errorf("installation succeeded but server not found: %w", err)
	}
	
	if !i.verifyInstallation(serverPath) {
		return fmt.Errorf("installation verification failed")
	}
	
	logging.Info("Successfully installed GitHub Copilot Language Server", "path", serverPath)
	return nil
}

// GetServerPath returns the path to the Copilot server executable
func (i *Installer) GetServerPath() (string, error) {
	// Use configured path if available
	if i.config.ServerPath != "" {
		if i.verifyInstallation(i.config.ServerPath) {
			return i.config.ServerPath, nil
		}
		return "", fmt.Errorf("configured server path is invalid: %s", i.config.ServerPath)
	}
	
	// Try to find in PATH
	if path, err := exec.LookPath("copilot-language-server"); err == nil {
		return path, nil
	}
	
	// Try common installation locations
	commonPaths := i.getCommonPaths()
	for _, path := range commonPaths {
		if i.verifyInstallation(path) {
			return path, nil
		}
	}
	
	return "", fmt.Errorf("copilot-language-server not found")
}

// getCommonPaths returns common installation paths for the Copilot server
func (i *Installer) getCommonPaths() []string {
	var paths []string
	
	// Get npm global prefix
	if npmPrefix, err := i.getNpmGlobalPrefix(); err == nil {
		binPath := filepath.Join(npmPrefix, "bin", "copilot-language-server")
		if runtime.GOOS == "windows" {
			binPath += ".cmd"
		}
		paths = append(paths, binPath)
	}
	
	// Platform-specific paths
	switch runtime.GOOS {
	case "darwin":
		paths = append(paths,
			"/usr/local/bin/copilot-language-server",
			"/opt/homebrew/bin/copilot-language-server",
		)
	case "linux":
		paths = append(paths,
			"/usr/local/bin/copilot-language-server",
			"/usr/bin/copilot-language-server",
		)
	case "windows":
		paths = append(paths,
			`C:\Program Files\nodejs\copilot-language-server.cmd`,
			`C:\Users\%USERNAME%\AppData\Roaming\npm\copilot-language-server.cmd`,
		)
	}
	
	// User-specific paths
	if homeDir, err := os.UserHomeDir(); err == nil {
		userPaths := []string{
			filepath.Join(homeDir, ".npm-global", "bin", "copilot-language-server"),
			filepath.Join(homeDir, "node_modules", ".bin", "copilot-language-server"),
		}
		if runtime.GOOS == "windows" {
			for i, path := range userPaths {
				userPaths[i] = path + ".cmd"
			}
		}
		paths = append(paths, userPaths...)
	}
	
	return paths
}

// getNpmGlobalPrefix gets the npm global prefix directory
func (i *Installer) getNpmGlobalPrefix() (string, error) {
	cmd := exec.Command("npm", "config", "get", "prefix")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	
	prefix := strings.TrimSpace(string(output))
	if prefix == "" {
		return "", fmt.Errorf("empty npm prefix")
	}
	
	return prefix, nil
}

// verifyInstallation verifies that the server at the given path is valid
func (i *Installer) verifyInstallation(path string) bool {
	// Check if file exists and is executable
	if info, err := os.Stat(path); err != nil || info.IsDir() {
		return false
	}
	
	// Try to run with --version flag
	cmd := exec.Command(path, "--version")
	if err := cmd.Run(); err != nil {
		return false
	}
	
	return true
}

// IsInstalled checks if the Copilot server is installed
func (i *Installer) IsInstalled() bool {
	_, err := i.GetServerPath()
	return err == nil
}

// GetVersion returns the version of the installed Copilot server
func (i *Installer) GetVersion() (string, error) {
	serverPath, err := i.GetServerPath()
	if err != nil {
		return "", err
	}
	
	cmd := exec.Command(serverPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}
	
	version := strings.TrimSpace(string(output))
	return version, nil
}

// Update updates the Copilot server to the latest version
func (i *Installer) Update(ctx context.Context) error {
	logging.Info("Updating GitHub Copilot Language Server")
	
	// Check if npm is available
	if _, err := exec.LookPath("npm"); err != nil {
		return fmt.Errorf("npm is required but not found in PATH")
	}
	
	// Update the package
	cmd := exec.CommandContext(ctx, "npm", "update", "-g", "@github/copilot-language-server")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update copilot-language-server: %w", err)
	}
	
	logging.Info("Successfully updated GitHub Copilot Language Server")
	return nil
}

// Uninstall removes the Copilot server
func (i *Installer) Uninstall(ctx context.Context) error {
	logging.Info("Uninstalling GitHub Copilot Language Server")
	
	// Check if npm is available
	if _, err := exec.LookPath("npm"); err != nil {
		return fmt.Errorf("npm is required but not found in PATH")
	}
	
	// Uninstall the package
	cmd := exec.CommandContext(ctx, "npm", "uninstall", "-g", "@github/copilot-language-server")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to uninstall copilot-language-server: %w", err)
	}
	
	logging.Info("Successfully uninstalled GitHub Copilot Language Server")
	return nil
}

// GetInstallationInfo returns information about the current installation
func (i *Installer) GetInstallationInfo() InstallationInfo {
	info := InstallationInfo{
		IsInstalled: false,
	}
	
	if serverPath, err := i.GetServerPath(); err == nil {
		info.IsInstalled = true
		info.Path = serverPath
		
		if version, err := i.GetVersion(); err == nil {
			info.Version = version
		}
	}
	
	return info
}

// InstallationInfo represents information about the Copilot server installation
type InstallationInfo struct {
	IsInstalled bool   `json:"is_installed"`
	Path        string `json:"path,omitempty"`
	Version     string `json:"version,omitempty"`
}