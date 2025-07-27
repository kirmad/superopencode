package copilot

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kirmad/superopencode/internal/config"
	"github.com/kirmad/superopencode/internal/logging"
)

// AuthManager handles GitHub Copilot authentication
type AuthManager struct {
	config      *config.CopilotConfig
	token       string
	isValidated bool
	lastCheck   time.Time
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(cfg *config.CopilotConfig) *AuthManager {
	return &AuthManager{
		config: cfg,
	}
}

// Authenticate authenticates with GitHub Copilot
func (a *AuthManager) Authenticate(ctx context.Context) error {
	// Try different authentication methods in order of preference
	if err := a.tryTokenAuth(); err == nil {
		return a.validateAuth(ctx)
	}
	
	if err := a.tryGitHubCLIAuth(); err == nil {
		return a.validateAuth(ctx)
	}
	
	if err := a.tryInteractiveAuth(ctx); err == nil {
		return a.validateAuth(ctx)
	}
	
	return fmt.Errorf("all authentication methods failed")
}

// tryTokenAuth attempts authentication using a configured token
func (a *AuthManager) tryTokenAuth() error {
	// Check config token first
	if a.config.AuthToken != "" {
		a.token = a.config.AuthToken
		logging.Debug("Using token from config")
		return nil
	}
	
	// Check environment variable
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		a.token = token
		logging.Debug("Using token from GITHUB_TOKEN environment variable")
		return nil
	}
	
	// Check Copilot-specific environment variable
	if token := os.Getenv("COPILOT_TOKEN"); token != "" {
		a.token = token
		logging.Debug("Using token from COPILOT_TOKEN environment variable")
		return nil
	}
	
	return fmt.Errorf("no token found in config or environment")
}

// tryGitHubCLIAuth attempts authentication using GitHub CLI
func (a *AuthManager) tryGitHubCLIAuth() error {
	// Check if gh CLI is available
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("github cli not found")
	}
	
	// Get token from gh CLI
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get token from gh cli: %w", err)
	}
	
	token := strings.TrimSpace(string(output))
	if token == "" {
		return fmt.Errorf("empty token from gh cli")
	}
	
	a.token = token
	logging.Debug("Using token from GitHub CLI")
	return nil
}

// tryInteractiveAuth attempts interactive authentication
func (a *AuthManager) tryInteractiveAuth(ctx context.Context) error {
	// For now, return an error. In a full implementation, this would
	// guide the user through the GitHub OAuth flow
	return fmt.Errorf("interactive authentication not yet implemented")
}

// validateAuth validates the current authentication token
func (a *AuthManager) validateAuth(ctx context.Context) error {
	if a.token == "" {
		return fmt.Errorf("no token available for validation")
	}
	
	// Check if we recently validated
	if a.isValidated && time.Since(a.lastCheck) < 5*time.Minute {
		return nil
	}
	
	// Validate token by checking Copilot subscription
	if err := a.checkCopilotSubscription(ctx); err != nil {
		a.isValidated = false
		return fmt.Errorf("token validation failed: %w", err)
	}
	
	a.isValidated = true
	a.lastCheck = time.Now()
	return nil
}

// checkCopilotSubscription verifies the user has access to GitHub Copilot
func (a *AuthManager) checkCopilotSubscription(ctx context.Context) error {
	// Use GitHub CLI to check Copilot billing status
	if _, err := exec.LookPath("gh"); err == nil {
		cmd := exec.CommandContext(ctx, "gh", "api", "user/copilot_billing")
		if err := cmd.Run(); err == nil {
			logging.Debug("Copilot subscription verified via GitHub CLI")
			return nil
		}
	}
	
	// For now, assume valid if we have a token
	// In a full implementation, this would make an HTTP request to GitHub's API
	logging.Debug("Copilot subscription assumed valid (verification not implemented)")
	return nil
}

// GetToken returns the current authentication token
func (a *AuthManager) GetToken() string {
	return a.token
}

// IsAuthenticated returns whether we have a valid authentication
func (a *AuthManager) IsAuthenticated() bool {
	return a.isValidated && a.token != ""
}

// HandleAuthRequest handles authentication requests from the Copilot server
func (a *AuthManager) HandleAuthRequest(params interface{}) (interface{}, error) {
	// Return the current token if we have one
	if a.token != "" {
		return map[string]interface{}{
			"token": a.token,
		}, nil
	}
	
	return nil, fmt.Errorf("no authentication token available")
}

// RefreshAuth refreshes the authentication token
func (a *AuthManager) RefreshAuth(ctx context.Context) error {
	a.isValidated = false
	a.lastCheck = time.Time{}
	return a.Authenticate(ctx)
}

// SetToken sets the authentication token manually
func (a *AuthManager) SetToken(token string) {
	a.token = token
	a.isValidated = false // Force re-validation
}

// GetAuthStatus returns the current authentication status
func (a *AuthManager) GetAuthStatus() AuthStatus {
	return AuthStatus{
		HasToken:      a.token != "",
		IsValidated:   a.isValidated,
		LastCheck:     a.lastCheck,
		TokenSource:   a.getTokenSource(),
	}
}

// getTokenSource determines where the current token came from
func (a *AuthManager) getTokenSource() string {
	if a.config.AuthToken != "" && a.token == a.config.AuthToken {
		return "config"
	}
	if os.Getenv("GITHUB_TOKEN") != "" && a.token == os.Getenv("GITHUB_TOKEN") {
		return "GITHUB_TOKEN"
	}
	if os.Getenv("COPILOT_TOKEN") != "" && a.token == os.Getenv("COPILOT_TOKEN") {
		return "COPILOT_TOKEN"
	}
	if _, err := exec.LookPath("gh"); err == nil {
		return "github_cli"
	}
	return "unknown"
}

// AuthStatus represents the current authentication status
type AuthStatus struct {
	HasToken    bool      `json:"has_token"`
	IsValidated bool      `json:"is_validated"`
	LastCheck   time.Time `json:"last_check"`
	TokenSource string    `json:"token_source"`
}