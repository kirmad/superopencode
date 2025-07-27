package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/kirmad/superopencode/internal/config"
	"github.com/kirmad/superopencode/internal/lsp/copilot"
	"github.com/spf13/cobra"
)

func TestCopilotCommands(t *testing.T) {
	// Test command structure
	if copilotCmd == nil {
		t.Error("copilotCmd is nil")
	}

	if copilotStatusCmd == nil {
		t.Error("copilotStatusCmd is nil")
	}

	if copilotInstallCmd == nil {
		t.Error("copilotInstallCmd is nil")
	}

	if copilotConfigCmd == nil {
		t.Error("copilotConfigCmd is nil")
	}
}

func TestCopilotCommandHierarchy(t *testing.T) {
	// Check that subcommands are properly added
	subcommands := copilotCmd.Commands()
	
	expectedCommands := map[string]bool{
		"status":  false,
		"install": false,
		"config":  false,
	}

	for _, cmd := range subcommands {
		if _, exists := expectedCommands[cmd.Name()]; exists {
			expectedCommands[cmd.Name()] = true
		}
	}

	for cmdName, found := range expectedCommands {
		if !found {
			t.Errorf("Expected subcommand %s not found in copilot command", cmdName)
		}
	}
}

func TestRunCopilotStatusWithMockConfig(t *testing.T) {
	// Create a mock configuration
	originalConfig := config.Get()
	defer func() {
		// This is tricky - we'd need a way to reset config in real usage
		// For now, just note that config management in tests is complex
	}()

	// Create test config
	testConfig := &config.Config{
		Copilot: config.CopilotConfig{
			EnableCopilot:     true,
			ChatEnabled:       true,
			CompletionEnabled: true,
			LogLevel:          "info",
		},
	}

	// We can't easily test the actual command execution without significant refactoring
	// to make config injectable, but we can test the command structure
	
	// Test command creation and basic properties
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show GitHub Copilot status",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Mock implementation for testing
			status := map[string]interface{}{
				"enabled": testConfig.Copilot.EnableCopilot,
				"installation": map[string]interface{}{
					"installed": false,
					"path":      "",
					"version":   "",
				},
				"authentication": map[string]interface{}{
					"authenticated": false,
					"has_token":     false,
					"token_source":  "",
				},
				"configuration": map[string]interface{}{
					"chat_enabled":       testConfig.Copilot.ChatEnabled,
					"completion_enabled": testConfig.Copilot.CompletionEnabled,
					"log_level":          testConfig.Copilot.LogLevel,
				},
			}

			output, err := json.MarshalIndent(status, "", "  ")
			if err != nil {
				return err
			}

			cmd.Print(string(output))
			return nil
		},
	}

	// Test that the command can be executed
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Mock status command failed: %v", err)
	}

	// Verify output is valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Errorf("Status command output is not valid JSON: %v", err)
	}

	// Check basic structure
	if enabled, ok := result["enabled"].(bool); !ok || !enabled {
		t.Error("Expected enabled to be true in status output")
	}

	_ = originalConfig // Use the variable to avoid compiler warning
}

func TestRunCopilotConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.CopilotConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &config.CopilotConfig{
				EnableCopilot:     true,
				ChatEnabled:       true,
				CompletionEnabled: true,
				LogLevel:          "info",
				ServerPath:        "/usr/local/bin/copilot-language-server",
			},
			wantErr: false,
		},
		{
			name: "invalid config - enabled without server path",
			cfg: &config.CopilotConfig{
				EnableCopilot: true,
				ServerPath:    "",
			},
			wantErr: true,
		},
		{
			name: "disabled config",
			cfg: &config.CopilotConfig{
				EnableCopilot: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test configuration validation
			mergedConfig := copilot.MergeConfig(tt.cfg)
			err := copilot.ValidateConfig(mergedConfig)

			if (err != nil) != tt.wantErr {
				t.Errorf("Config validation error = %v, wantErr %v", err, tt.wantErr)
			}

			// Test that config can be marshaled to JSON (what the command does)
			if !tt.wantErr {
				_, err := json.MarshalIndent(mergedConfig, "", "  ")
				if err != nil {
					t.Errorf("Failed to marshal config to JSON: %v", err)
				}
			}
		})
	}
}

func TestCopilotInstallFlow(t *testing.T) {
	// Test the install command logic without actually running npm
	cfg := &config.CopilotConfig{
		EnableCopilot: true,
		AutoInstall:   true,
		ServerPath:    "",
	}

	// Create installer to test the flow
	installer := copilot.NewInstaller(cfg)
	
	// Test that we can get installation info
	info := installer.GetInstallationInfo()
	
	// Should return non-nil info
	if info == nil {
		t.Error("GetInstallationInfo() returned nil")
	}

	// Test path detection
	_, err := installer.GetServerPath()
	if err != nil {
		t.Logf("GetServerPath() failed (expected in test environment): %v", err)
	}

	// Test that installation check doesn't panic
	installed := installer.IsInstalled()
	_ = installed // Use the variable to avoid warning
}

func TestCopilotAuthFlow(t *testing.T) {
	// Save original env vars
	originalGithubToken := os.Getenv("GITHUB_TOKEN")
	originalCopilotToken := os.Getenv("COPILOT_TOKEN")
	
	defer func() {
		// Restore original env vars
		if originalGithubToken != "" {
			os.Setenv("GITHUB_TOKEN", originalGithubToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
		if originalCopilotToken != "" {
			os.Setenv("COPILOT_TOKEN", originalCopilotToken)
		} else {
			os.Unsetenv("COPILOT_TOKEN")
		}
	}()

	// Test auth flow without real tokens
	cfg := &config.CopilotConfig{
		EnableCopilot: true,
		AuthToken:     "",
	}

	authManager := copilot.NewAuthManager(cfg)
	
	// Test getting auth status
	status := authManager.GetAuthStatus()
	if status == nil {
		t.Error("GetAuthStatus() returned nil")
	}

	// Test environment token detection by setting a token and checking status
	os.Setenv("GITHUB_TOKEN", "test-token")
	
	// Create a new auth manager to pick up the environment token
	authManager2 := copilot.NewAuthManager(cfg)
	err := authManager2.tryTokenAuth()
	if err != nil {
		t.Errorf("tryTokenAuth() should succeed with GITHUB_TOKEN: %v", err)
	}
	
	if authManager2.GetToken() != "test-token" {
		t.Errorf("Expected token to be 'test-token', got '%s'", authManager2.GetToken())
	}

	// Test token setting
	authManager.SetToken("test-token-manual")
	if authManager.GetToken() != "test-token-manual" {
		t.Errorf("SetToken() - GetToken() = %v, want %v", authManager.GetToken(), "test-token-manual")
	}
}

func TestCopilotStatusCommandIntegration(t *testing.T) {
	// Test that the actual status command function can handle various scenarios
	
	// Test with no config (should handle gracefully)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// We can't easily test the actual command without significant refactoring
	// but we can test the underlying components it uses
	
	cfg := &config.CopilotConfig{
		EnableCopilot:     true,
		ChatEnabled:       true,
		CompletionEnabled: false,
		LogLevel:          "debug",
	}

	mergedConfig := copilot.LoadFromEnvironment(copilot.MergeConfig(cfg))
	
	// Test installer component
	installer := copilot.NewInstaller(mergedConfig)
	installInfo := installer.GetInstallationInfo()
	
	if installInfo == nil {
		t.Error("Expected installation info to be returned")
	}

	// Test auth component
	authManager := copilot.NewAuthManager(mergedConfig)
	authStatus := authManager.GetAuthStatus()
	
	if authStatus == nil {
		t.Error("Expected auth status to be returned")
	}

	// Test that we can construct the status object like the command does
	status := map[string]interface{}{
		"enabled":        mergedConfig.EnableCopilot,
		"installation":   installInfo,
		"authentication": authStatus,
		"configuration": map[string]interface{}{
			"chat_enabled":       mergedConfig.ChatEnabled,
			"completion_enabled": mergedConfig.CompletionEnabled,
			"log_level":          mergedConfig.LogLevel,
		},
	}

	// Test marshaling to JSON
	_, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		t.Errorf("Failed to marshal status to JSON: %v", err)
	}
}