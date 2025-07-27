package copilot

import (
	"os"
	"testing"

	"github.com/kirmad/superopencode/internal/config"
)

func TestMergeConfig(t *testing.T) {
	tests := []struct {
		name   string
		input  *config.CopilotConfig
		expect func(*config.CopilotConfig) bool
	}{
		{
			name:  "nil config gets defaults",
			input: nil,
			expect: func(cfg *config.CopilotConfig) bool {
				return cfg.EnableCopilot == false && // Default is false
					cfg.FallbackToGopls == true && // Default is true
					cfg.LogLevel == "info" // Default log level
			},
		},
		{
			name: "partial config gets merged with defaults",
			input: &config.CopilotConfig{
				EnableCopilot: true,
			},
			expect: func(cfg *config.CopilotConfig) bool {
				return cfg.EnableCopilot == true && // User provided
					cfg.FallbackToGopls == true && // Default
					cfg.LogLevel == "info" // Default
			},
		},
		{
			name: "complete config preserved",
			input: &config.CopilotConfig{
				EnableCopilot:     true,
				ServerPath:        "/custom/path",
				FallbackToGopls:   false,
				ChatEnabled:       true,
				CompletionEnabled: true,
				LogLevel:          "debug",
			},
			expect: func(cfg *config.CopilotConfig) bool {
				return cfg.EnableCopilot == true &&
					cfg.ServerPath == "/custom/path" &&
					cfg.FallbackToGopls == false &&
					cfg.ChatEnabled == true &&
					cfg.CompletionEnabled == true &&
					cfg.LogLevel == "debug"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeConfig(tt.input)
			if !tt.expect(result) {
				t.Errorf("MergeConfig() result did not meet expectations: %+v", result)
			}
		})
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	// Save current env vars
	originalVars := make(map[string]string)
	envVars := []string{
		"OPENCODE_COPILOT_ENABLE",
		"OPENCODE_COPILOT_SERVER_PATH",
		"OPENCODE_COPILOT_AUTH_TOKEN",
		"OPENCODE_COPILOT_CHAT_ENABLED",
		"OPENCODE_COPILOT_COMPLETION_ENABLED",
		"OPENCODE_COPILOT_LOG_LEVEL",
		"OPENCODE_COPILOT_FALLBACK_TO_GOPLS",
		"OPENCODE_COPILOT_AUTO_INSTALL",
	}

	for _, envVar := range envVars {
		if val, exists := os.LookupEnv(envVar); exists {
			originalVars[envVar] = val
		}
		os.Unsetenv(envVar)
	}

	defer func() {
		// Restore original env vars
		for _, envVar := range envVars {
			os.Unsetenv(envVar)
			if val, exists := originalVars[envVar]; exists {
				os.Setenv(envVar, val)
			}
		}
	}()

	t.Run("no environment variables", func(t *testing.T) {
		cfg := &config.CopilotConfig{
			EnableCopilot: false,
			LogLevel:      "info",
		}
		
		result := LoadFromEnvironment(cfg)
		
		if result.EnableCopilot != false {
			t.Error("Expected EnableCopilot to remain false")
		}
		if result.LogLevel != "info" {
			t.Error("Expected LogLevel to remain 'info'")
		}
	})

	t.Run("with environment variables", func(t *testing.T) {
		os.Setenv("OPENCODE_COPILOT_ENABLE", "true")
		os.Setenv("OPENCODE_COPILOT_SERVER_PATH", "/env/path")
		os.Setenv("OPENCODE_COPILOT_AUTH_TOKEN", "env-token")
		os.Setenv("OPENCODE_COPILOT_CHAT_ENABLED", "true")
		os.Setenv("OPENCODE_COPILOT_COMPLETION_ENABLED", "false")
		os.Setenv("OPENCODE_COPILOT_LOG_LEVEL", "debug")
		os.Setenv("OPENCODE_COPILOT_FALLBACK_TO_GOPLS", "false")
		os.Setenv("OPENCODE_COPILOT_AUTO_INSTALL", "true")

		cfg := &config.CopilotConfig{
			EnableCopilot: false,
			LogLevel:      "info",
		}
		
		result := LoadFromEnvironment(cfg)
		
		if !result.EnableCopilot {
			t.Error("Expected EnableCopilot to be overridden to true")
		}
		if result.ServerPath != "/env/path" {
			t.Error("Expected ServerPath to be overridden")
		}
		if result.AuthToken != "env-token" {
			t.Error("Expected AuthToken to be set from env")
		}
		if !result.ChatEnabled {
			t.Error("Expected ChatEnabled to be overridden to true")
		}
		if result.CompletionEnabled {
			t.Error("Expected CompletionEnabled to be overridden to false")
		}
		if result.LogLevel != "debug" {
			t.Error("Expected LogLevel to be overridden to debug")
		}
		if result.FallbackToGopls {
			t.Error("Expected FallbackToGopls to be overridden to false")
		}
		if !result.AutoInstall {
			t.Error("Expected AutoInstall to be overridden to true")
		}
	})
}

func TestValidateConfig(t *testing.T) {
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
			wantErr: false, // Disabled is valid
		},
		{
			name: "enabled without server path",
			cfg: &config.CopilotConfig{
				EnableCopilot: true,
				ServerPath:    "",
			},
			wantErr: true,
		},
		{
			name: "enabled with server path",
			cfg: &config.CopilotConfig{
				EnableCopilot: true,
				ServerPath:    "/usr/local/bin/copilot-language-server",
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			cfg: &config.CopilotConfig{
				EnableCopilot: true,
				ServerPath:    "/usr/local/bin/copilot-language-server",
				LogLevel:      "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid log level",
			cfg: &config.CopilotConfig{
				EnableCopilot: true,
				ServerPath:    "/usr/local/bin/copilot-language-server",
				LogLevel:      "debug",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetConfigForProfile(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		expect  func(*config.CopilotConfig) bool
	}{
		{
			name:    "development profile",
			profile: "development",
			expect: func(cfg *config.CopilotConfig) bool {
				return cfg.LogLevel == "debug" &&
					cfg.FallbackToGopls == true
			},
		},
		{
			name:    "production profile",
			profile: "production",
			expect: func(cfg *config.CopilotConfig) bool {
				return cfg.LogLevel == "info" &&
					cfg.ReplaceGopls == true &&
					cfg.UseNativeBinary == true
			},
		},
		{
			name:    "testing profile",
			profile: "testing",
			expect: func(cfg *config.CopilotConfig) bool {
				return cfg.LogLevel == "warn" &&
					cfg.FallbackToGopls == true &&
					cfg.Performance.MaxCompletionTime == 1000
			},
		},
		{
			name:    "unknown profile returns defaults",
			profile: "unknown",
			expect: func(cfg *config.CopilotConfig) bool {
				return cfg.LogLevel == "info" &&
					cfg.FallbackToGopls == true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetConfigForProfile(tt.profile)
			if !tt.expect(result) {
				t.Errorf("GetConfigForProfile(%s) result did not meet expectations: %+v", tt.profile, result)
			}
		})
	}
}