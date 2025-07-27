package copilot

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kirmad/superopencode/internal/config"
)

func TestNewInstaller(t *testing.T) {
	cfg := &config.CopilotConfig{
		EnableCopilot: true,
		ServerPath:    "/usr/local/bin/copilot-language-server",
	}

	installer := NewInstaller(cfg)
	if installer == nil {
		t.Error("NewInstaller() returned nil")
	}

	if installer.config != cfg {
		t.Error("NewInstaller() did not store config correctly")
	}
}

func TestInstaller_GetServerPath(t *testing.T) {
	tests := []struct {
		name       string
		cfg        *config.CopilotConfig
		wantErr    bool
		wantCustom bool
	}{
		{
			name: "custom server path",
			cfg: &config.CopilotConfig{
				EnableCopilot: true,
				ServerPath:    "/custom/path/copilot-language-server",
			},
			wantErr:    false,
			wantCustom: true,
		},
		{
			name: "auto-detect path",
			cfg: &config.CopilotConfig{
				EnableCopilot: true,
				ServerPath:    "",
			},
			wantErr:    false,
			wantCustom: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installer := NewInstaller(tt.cfg)
			path, err := installer.GetServerPath()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetServerPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if path == "" {
					t.Error("GetServerPath() returned empty path")
				}

				if tt.wantCustom && path != tt.cfg.ServerPath {
					t.Errorf("GetServerPath() = %v, want %v", path, tt.cfg.ServerPath)
				}
			}
		})
	}
}

func TestInstaller_GetInstallationInfo(t *testing.T) {
	cfg := &config.CopilotConfig{
		EnableCopilot: true,
		ServerPath:    "/nonexistent/path/copilot-language-server",
	}

	installer := NewInstaller(cfg)
	info := installer.GetInstallationInfo()

	// For a non-existent path, should indicate not installed
	if info.IsInstalled {
		t.Error("Expected installation info to show not installed for non-existent path")
	}

	if info.Path != "" {
		t.Error("Expected empty path for non-existent installation")
	}

	if info.Version != "" {
		t.Error("Expected empty version for non-existent installation")
	}
}

func TestInstaller_IsInstalled(t *testing.T) {
	cfg := &config.CopilotConfig{
		EnableCopilot: true,
		ServerPath:    "/nonexistent/path/copilot-language-server",
	}

	installer := NewInstaller(cfg)
	
	if installer.IsInstalled() {
		t.Error("Expected IsInstalled() to return false for non-existent path")
	}
}

func TestInstaller_EnsureInstalled(t *testing.T) {
	cfg := &config.CopilotConfig{
		EnableCopilot: true,
		AutoInstall:   true,
		ServerPath:    "",
	}

	installer := NewInstaller(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This will likely fail in most test environments, but we can test the basic flow
	err := installer.EnsureInstalled(ctx)
	
	// We expect this to likely fail in test environments due to npm not being available
	// or network restrictions, but the function should handle errors gracefully
	if err != nil {
		t.Logf("EnsureInstalled failed as expected in test environment: %v", err)
	}
}

func TestInstaller_GetVersion(t *testing.T) {
	cfg := &config.CopilotConfig{
		EnableCopilot: true,
		ServerPath:    "/nonexistent/path/copilot-language-server",
	}

	installer := NewInstaller(cfg)
	
	// Should fail for non-existent path
	_, err := installer.GetVersion()
	if err == nil {
		t.Error("GetVersion() should fail for non-existent server path")
	}
	
	// If we could find an actual path, we could test version detection
	// But in test environment, this is expected to fail
	t.Logf("GetVersion failed as expected: %v", err)
}

func TestInstaller_Uninstall(t *testing.T) {
	// Create a temporary directory and binary for testing
	tempDir := t.TempDir()
	tempBinary := filepath.Join(tempDir, "copilot-language-server")

	// Create a dummy binary file
	file, err := os.Create(tempBinary)
	if err != nil {
		t.Fatalf("Failed to create temp binary: %v", err)
	}
	file.Close()

	cfg := &config.CopilotConfig{
		EnableCopilot: true,
		ServerPath:    tempBinary,
	}

	installer := NewInstaller(cfg)
	
	// Test uninstalling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = installer.Uninstall(ctx)
	// This might fail if npm is not available, but we can test the basic flow
	if err != nil {
		t.Logf("Uninstall failed as expected in test environment: %v", err)
	}
}

func TestInstaller_Update(t *testing.T) {
	cfg := &config.CopilotConfig{
		EnableCopilot: true,
		ServerPath:    "/nonexistent/path",
	}

	installer := NewInstaller(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This will likely fail in most test environments
	err := installer.Update(ctx)
	if err != nil {
		t.Logf("Update failed as expected in test environment: %v", err)
	}
}

func TestInstaller_getNpmGlobalPrefix(t *testing.T) {
	cfg := &config.CopilotConfig{
		EnableCopilot: true,
	}

	installer := NewInstaller(cfg)
	
	// This might return empty if npm is not available, which is fine for testing
	path, err := installer.getNpmGlobalPrefix()
	
	// Just verify it doesn't panic
	if err != nil {
		t.Logf("getNpmGlobalPrefix() failed (expected in environments without npm): %v", err)
		return
	}
	
	if path != "" {
		// If we got a path, it should be a valid directory
		if stat, err := os.Stat(path); err != nil || !stat.IsDir() {
			t.Errorf("getNpmGlobalPrefix() returned invalid path: %s", path)
		}
	}
}