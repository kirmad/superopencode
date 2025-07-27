package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/kirmad/superopencode/internal/config"
	"github.com/kirmad/superopencode/internal/lsp/copilot"
	"github.com/spf13/cobra"
)

var copilotCmd = &cobra.Command{
	Use:   "copilot",
	Short: "GitHub Copilot management commands",
	Long:  `Manage GitHub Copilot integration, including installation, authentication, and configuration.`,
}

var copilotStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show GitHub Copilot status",
	Long:  `Display the current status of GitHub Copilot integration including installation, authentication, and configuration.`,
	RunE:  runCopilotStatus,
}

var copilotInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install GitHub Copilot Language Server",
	Long:  `Install the GitHub Copilot Language Server using npm.`,
	RunE:  runCopilotInstall,
}

var copilotConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Show GitHub Copilot configuration",
	Long:  `Display the current GitHub Copilot configuration with merged defaults and environment variables.`,
	RunE:  runCopilotConfig,
}

func runCopilotStatus(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Load configuration (if not already loaded)
	cfg := config.Get()
	if cfg == nil {
		// Load config from current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		
		cfg, err = config.Load(cwd, false)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
	}

	// Merge with defaults and environment
	copilotConfig := copilot.LoadFromEnvironment(copilot.MergeConfig(&cfg.Copilot))

	// Create installer to check installation status
	installer := copilot.NewInstaller(copilotConfig)
	installInfo := installer.GetInstallationInfo()

	// Create auth manager to check authentication status
	authManager := copilot.NewAuthManager(copilotConfig)
	
	// Always try to authenticate first to load tokens
	_ = authManager.Authenticate(ctx) // Try to authenticate and validate
	authStatus := authManager.GetAuthStatus()

	// Prepare status output
	status := map[string]interface{}{
		"enabled":        copilotConfig.EnableCopilot,
		"installation":   installInfo,
		"authentication": authStatus,
		"configuration": map[string]interface{}{
			"chat_enabled":       copilotConfig.ChatEnabled,
			"completion_enabled": copilotConfig.CompletionEnabled,
			"auto_install":       copilotConfig.AutoInstall,
			"replace_gopls":      copilotConfig.ReplaceGopls,
			"fallback_to_gopls":  copilotConfig.FallbackToGopls,
			"log_level":          copilotConfig.LogLevel,
		},
	}

	// Output as JSON for now
	output, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

func runCopilotInstall(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Load configuration (if not already loaded)
	cfg := config.Get()
	if cfg == nil {
		// Load config from current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		
		cfg, err = config.Load(cwd, false)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
	}

	// Merge with defaults and environment
	copilotConfig := copilot.LoadFromEnvironment(copilot.MergeConfig(&cfg.Copilot))

	// Create installer
	installer := copilot.NewInstaller(copilotConfig)

	fmt.Println("Installing GitHub Copilot Language Server...")

	// Install
	if err := installer.Install(ctx); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	fmt.Println("GitHub Copilot Language Server installed successfully!")

	// Show installation info
	installInfo := installer.GetInstallationInfo()
	if installInfo.Version != "" {
		fmt.Printf("Version: %s\n", installInfo.Version)
	}
	if installInfo.Path != "" {
		fmt.Printf("Path: %s\n", installInfo.Path)
	}

	return nil
}

func runCopilotConfig(cmd *cobra.Command, args []string) error {
	// Load configuration (if not already loaded)
	cfg := config.Get()
	if cfg == nil {
		// Load config from current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		
		cfg, err = config.Load(cwd, false)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
	}

	// Merge with defaults and environment
	copilotConfig := copilot.LoadFromEnvironment(copilot.MergeConfig(&cfg.Copilot))

	// Validate configuration
	if err := copilot.ValidateConfig(copilotConfig); err != nil {
		fmt.Printf("Warning: Configuration validation failed: %v\n\n", err)
	}

	// Output configuration as JSON
	output, err := json.MarshalIndent(copilotConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

func init() {
	// Add subcommands
	copilotCmd.AddCommand(copilotStatusCmd)
	copilotCmd.AddCommand(copilotInstallCmd)
	copilotCmd.AddCommand(copilotConfigCmd)

	// Add to root command
	rootCmd.AddCommand(copilotCmd)
}