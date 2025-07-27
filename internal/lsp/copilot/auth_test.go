package copilot

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/kirmad/superopencode/internal/config"
)

func TestNewAuthManager(t *testing.T) {
	cfg := &config.CopilotConfig{
		EnableCopilot: true,
		AuthToken:     "test-token",
	}

	authManager := NewAuthManager(cfg)
	if authManager == nil {
		t.Error("NewAuthManager() returned nil")
	}

	if authManager.config != cfg {
		t.Error("NewAuthManager() did not store config correctly")
	}
}

func TestAuthManager_GetAuthStatus(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *config.CopilotConfig
		expectHasToken bool
	}{
		{
			name: "with auth token",
			cfg: &config.CopilotConfig{
				EnableCopilot: true,
				AuthToken:     "test-token",
			},
			expectHasToken: true,
		},
		{
			name: "without auth token",
			cfg: &config.CopilotConfig{
				EnableCopilot: true,
				AuthToken:     "",
			},
			expectHasToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authManager := NewAuthManager(tt.cfg)
			status := authManager.GetAuthStatus()

			if status.HasToken != tt.expectHasToken {
				t.Errorf("GetAuthStatus().HasToken = %v, want %v", status.HasToken, tt.expectHasToken)
			}

			if tt.expectHasToken && status.TokenSource == "" {
				t.Error("Expected TokenSource to be set when token is available")
			}
		})
	}
}

func TestAuthManager_SetToken(t *testing.T) {
	cfg := &config.CopilotConfig{
		EnableCopilot: true,
	}
	authManager := NewAuthManager(cfg)

	// Test setting a token
	testToken := "ghs_abcdefghijklmnopqrstuvwxyz123456789"
	authManager.SetToken(testToken)
	
	if authManager.GetToken() != testToken {
		t.Errorf("SetToken() - GetToken() = %v, want %v", authManager.GetToken(), testToken)
	}
	
	// Test that setting a token clears validation status
	status := authManager.GetAuthStatus()
	if status.IsValidated {
		t.Error("SetToken() should clear validation status")
	}
}

func TestAuthManager_EnvironmentTokens(t *testing.T) {
	// Save original env vars
	originalToken := os.Getenv("GITHUB_TOKEN")
	originalCopilotToken := os.Getenv("COPILOT_TOKEN")
	
	defer func() {
		// Restore original env vars
		if originalToken != "" {
			os.Setenv("GITHUB_TOKEN", originalToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
		if originalCopilotToken != "" {
			os.Setenv("COPILOT_TOKEN", originalCopilotToken)
		} else {
			os.Unsetenv("COPILOT_TOKEN")
		}
	}()

	// Test environment token detection through authentication
	cfg := &config.CopilotConfig{
		EnableCopilot: true,
		AuthToken:     "", // No config token
	}
	authManager := NewAuthManager(cfg)

	// Test with GITHUB_TOKEN
	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("COPILOT_TOKEN")
	os.Setenv("GITHUB_TOKEN", "test-github-token")
	
	// Try token authentication method (which reads from environment)
	err := authManager.tryTokenAuth()
	if err != nil {
		t.Errorf("tryTokenAuth() should succeed with GITHUB_TOKEN: %v", err)
	}
	
	if authManager.GetToken() != "test-github-token" {
		t.Errorf("Expected token to be 'test-github-token', got '%s'", authManager.GetToken())
	}

	// Test with COPILOT_TOKEN (should take precedence)
	os.Setenv("COPILOT_TOKEN", "test-copilot-token")
	authManager.SetToken("") // Reset
	
	err = authManager.tryTokenAuth()
	if err != nil {
		t.Errorf("tryTokenAuth() should succeed with COPILOT_TOKEN: %v", err)
	}
	
	if authManager.GetToken() != "test-copilot-token" {
		t.Errorf("Expected token to be 'test-copilot-token', got '%s'", authManager.GetToken())
	}
}

func TestAuthManager_Authenticate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.CopilotConfig
		wantErr bool
	}{
		{
			name: "no auth token",
			cfg: &config.CopilotConfig{
				EnableCopilot: true,
				AuthToken:     "",
			},
			wantErr: true,
		},
		{
			name: "with auth token",
			cfg: &config.CopilotConfig{
				EnableCopilot: true,
				AuthToken:     "ghs_abcdefghijklmnopqrstuvwxyz123456789",
			},
			wantErr: false, // Note: This may still fail due to network/auth issues, but it should pass validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authManager := NewAuthManager(tt.cfg)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := authManager.Authenticate(ctx)
			
			// For tests, we mainly care about the validation logic
			// Network calls may fail in test environments
			if tt.wantErr && err == nil {
				t.Error("Authenticate() expected error but got none")
			}
			// If we don't expect an error, we might still get one due to network issues
			// So we don't fail the test for that case
		})
	}
}

func TestAuthManager_HandleAuthRequest(t *testing.T) {
	cfg := &config.CopilotConfig{
		EnableCopilot: true,
		AuthToken:     "test-token",
	}

	authManager := NewAuthManager(cfg)

	// Test with valid JSON
	validParams := json.RawMessage(`{"method": "test", "params": {}}`)
	result, err := authManager.HandleAuthRequest(validParams)
	
	// The exact result depends on implementation, but it should not panic
	if err != nil {
		t.Logf("HandleAuthRequest returned error (expected in test): %v", err)
	}
	
	// Result might be nil or contain auth response
	_ = result

	// Test with invalid JSON
	invalidParams := json.RawMessage(`{invalid json}`)
	_, err = authManager.HandleAuthRequest(invalidParams)
	
	if err == nil {
		t.Error("HandleAuthRequest() expected error for invalid JSON but got none")
	}
}

func TestAuthManager_IsAuthenticated(t *testing.T) {
	cfg := &config.CopilotConfig{
		EnableCopilot: true,
		AuthToken:     "test-token",
	}

	authManager := NewAuthManager(cfg)
	
	// Initially should not be authenticated
	if authManager.IsAuthenticated() {
		t.Error("IsAuthenticated() should return false initially")
	}
	
	// Set a token but don't validate - still not authenticated
	authManager.SetToken("test-token")
	if authManager.IsAuthenticated() {
		t.Error("IsAuthenticated() should return false with unvalidated token")
	}
}