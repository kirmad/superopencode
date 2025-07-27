package copilot

import (
	"context"
	"testing"
	"time"

	"github.com/kirmad/superopencode/internal/config"
)

func TestNewCopilotClient(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.CopilotConfig
		wantErr bool
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
		},
		{
			name: "disabled copilot",
			cfg: &config.CopilotConfig{
				EnableCopilot: false,
			},
			wantErr: true,
		},
		{
			name: "valid config without auto-install",
			cfg: &config.CopilotConfig{
				EnableCopilot:     true,
				AutoInstall:       false,
				FallbackToGopls:   true,
				ChatEnabled:       true,
				CompletionEnabled: true,
				ServerPath:        "/usr/local/bin/copilot-language-server",
				ServerArgs:        []string{"--stdio"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			client, err := NewCopilotClient(ctx, tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCopilotClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && client == nil {
				t.Error("NewCopilotClient() returned nil client without error")
			}

			if client != nil {
				defer client.Close()
			}
		})
	}
}

func TestCopilotClient_GetStatus(t *testing.T) {
	cfg := &config.CopilotConfig{
		EnableCopilot:     true,
		AutoInstall:       false,
		FallbackToGopls:   true,
		ChatEnabled:       true,
		CompletionEnabled: true,
		ServerPath:        "/usr/local/bin/copilot-language-server",
		ServerArgs:        []string{"--stdio"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := NewCopilotClient(ctx, cfg)
	if err != nil {
		t.Skipf("Cannot create copilot client: %v", err)
	}
	defer client.Close()

	status := client.GetStatus()
	
	if !status.Enabled {
		t.Error("Expected status.Enabled to be true")
	}
	
	if status.ChatEnabled != cfg.ChatEnabled {
		t.Errorf("Expected status.ChatEnabled = %v, got %v", cfg.ChatEnabled, status.ChatEnabled)
	}
	
	if status.CompletionEnabled != cfg.CompletionEnabled {
		t.Errorf("Expected status.CompletionEnabled = %v, got %v", cfg.CompletionEnabled, status.CompletionEnabled)
	}
}

func TestCopilotClient_EnableDisableFeatures(t *testing.T) {
	cfg := &config.CopilotConfig{
		EnableCopilot:     true,
		AutoInstall:       false,
		FallbackToGopls:   true,
		ChatEnabled:       false,
		CompletionEnabled: false,
		ServerPath:        "/usr/local/bin/copilot-language-server",
		ServerArgs:        []string{"--stdio"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := NewCopilotClient(ctx, cfg)
	if err != nil {
		t.Skipf("Cannot create copilot client: %v", err)
	}
	defer client.Close()

	// Test enabling chat
	client.EnableChat()
	status := client.GetStatus()
	if !status.ChatEnabled {
		t.Error("Expected chat to be enabled")
	}

	// Test disabling chat
	client.DisableChat()
	status = client.GetStatus()
	if status.ChatEnabled {
		t.Error("Expected chat to be disabled")
	}

	// Test enabling completion
	client.EnableCompletion()
	status = client.GetStatus()
	if !status.CompletionEnabled {
		t.Error("Expected completion to be enabled")
	}

	// Test disabling completion
	client.DisableCompletion()
	status = client.GetStatus()
	if status.CompletionEnabled {
		t.Error("Expected completion to be disabled")
	}
}

func TestCopilotClient_IsReady(t *testing.T) {
	cfg := &config.CopilotConfig{
		EnableCopilot:     true,
		AutoInstall:       false,
		FallbackToGopls:   true,
		ChatEnabled:       true,
		CompletionEnabled: true,
		ServerPath:        "/usr/local/bin/copilot-language-server",
		ServerArgs:        []string{"--stdio"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := NewCopilotClient(ctx, cfg)
	if err != nil {
		t.Skipf("Cannot create copilot client: %v", err)
	}
	defer client.Close()

	// Initially should not be ready (not authenticated and server not ready)
	if client.IsReady() {
		t.Error("Expected client to not be ready initially")
	}
}