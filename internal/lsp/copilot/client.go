package copilot

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/kirmad/superopencode/internal/config"
	"github.com/kirmad/superopencode/internal/lsp"
	"github.com/kirmad/superopencode/internal/logging"
)

// CopilotClient extends the base LSP client with Copilot-specific functionality
type CopilotClient struct {
	*lsp.Client
	
	config       *config.CopilotConfig
	authManager  *AuthManager
	installer    *Installer
	migrator     *MigrationManager
	
	// State management
	isAuthenticated bool
	serverReady     bool
	mu              sync.RWMutex
	
	// Chat and completion state
	chatEnabled       bool
	completionEnabled bool
	
	// Performance monitoring
	requestCount    int64
	errorCount      int64
	averageLatency  float64
}

// NewCopilotClient creates a new Copilot LSP client
func NewCopilotClient(ctx context.Context, cfg *config.CopilotConfig) (*CopilotClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("copilot config is required")
	}
	
	if !cfg.EnableCopilot {
		return nil, fmt.Errorf("copilot is disabled in configuration")
	}
	
	// Create installer
	installer := NewInstaller(cfg)
	
	// Auto-install if configured
	if cfg.AutoInstall {
		if err := installer.EnsureInstalled(ctx); err != nil {
			logging.Warn("Failed to auto-install Copilot server", "error", err)
			if !cfg.FallbackToGopls {
				return nil, fmt.Errorf("failed to install Copilot server: %w", err)
			}
		}
	}
	
	// Determine server path
	serverPath, err := installer.GetServerPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get server path: %w", err)
	}
	
	// Prepare server args
	args := cfg.ServerArgs
	if args == nil {
		args = []string{"--stdio"}
	}
	
	// Create base LSP client
	lspClient, err := lsp.NewClient(ctx, serverPath, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create LSP client: %w", err)
	}
	
	// Create auth manager
	authManager := NewAuthManager(cfg)
	
	// Create migration manager
	migrator := NewMigrationManager(cfg)
	
	client := &CopilotClient{
		Client:            lspClient,
		config:            cfg,
		authManager:       authManager,
		installer:         installer,
		migrator:          migrator,
		chatEnabled:       cfg.ChatEnabled,
		completionEnabled: cfg.CompletionEnabled,
	}
	
	// Setup Copilot-specific handlers
	client.setupHandlers()
	
	return client, nil
}

// Initialize initializes the Copilot client with authentication and workspace setup
func (c *CopilotClient) Initialize(ctx context.Context, workspaceRoot string) error {
	// Authenticate first
	if err := c.authManager.Authenticate(ctx); err != nil {
		logging.Error("Failed to authenticate with GitHub Copilot", err)
		if !c.config.FallbackToGopls {
			return fmt.Errorf("authentication failed: %w", err)
		}
		logging.Warn("Authentication failed, but fallback is enabled")
	} else {
		c.mu.Lock()
		c.isAuthenticated = true
		c.mu.Unlock()
		logging.Info("Successfully authenticated with GitHub Copilot")
	}
	
	// Initialize the base LSP client
	_, err := c.Client.InitializeLSPClient(ctx, workspaceRoot)
	if err != nil {
		return fmt.Errorf("failed to initialize LSP client: %w", err)
	}
	
	// Wait for server to be ready
	if err := c.Client.WaitForServerReady(ctx); err != nil {
		logging.Error("Copilot server failed to become ready", err)
		if !c.config.FallbackToGopls {
			return fmt.Errorf("server not ready: %w", err)
		}
		logging.Warn("Server not ready, but fallback is enabled")
	} else {
		c.mu.Lock()
		c.serverReady = true
		c.mu.Unlock()
		logging.Info("Copilot server is ready")
	}
	
	return nil
}

// IsReady returns whether the Copilot client is ready for use
func (c *CopilotClient) IsReady() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isAuthenticated && c.serverReady
}

// GetStatus returns the current status of the Copilot client
func (c *CopilotClient) GetStatus() CopilotStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return CopilotStatus{
		Enabled:         c.config.EnableCopilot,
		Authenticated:   c.isAuthenticated,
		ServerReady:     c.serverReady,
		ChatEnabled:     c.chatEnabled,
		CompletionEnabled: c.completionEnabled,
		RequestCount:    c.requestCount,
		ErrorCount:      c.errorCount,
		AverageLatency:  c.averageLatency,
	}
}

// EnableChat enables chat functionality
func (c *CopilotClient) EnableChat() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.chatEnabled = true
}

// DisableChat disables chat functionality
func (c *CopilotClient) DisableChat() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.chatEnabled = false
}

// EnableCompletion enables completion functionality
func (c *CopilotClient) EnableCompletion() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.completionEnabled = true
}

// DisableCompletion disables completion functionality
func (c *CopilotClient) DisableCompletion() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.completionEnabled = false
}

// setupHandlers sets up Copilot-specific notification and request handlers
func (c *CopilotClient) setupHandlers() {
	// Register Copilot-specific notification handlers
	c.Client.RegisterNotificationHandler("$/copilot/statusChange", c.handleStatusChange)
	c.Client.RegisterNotificationHandler("$/copilot/progress", c.handleProgress)
	c.Client.RegisterNotificationHandler("$/copilot/error", c.handleError)
	
	// Register Copilot-specific request handlers
	c.Client.RegisterServerRequestHandler("copilot/auth", c.handleAuthRequest)
	c.Client.RegisterServerRequestHandler("copilot/chat", c.handleChatRequest)
}

// handleStatusChange handles Copilot status change notifications
func (c *CopilotClient) handleStatusChange(params json.RawMessage) {
	logging.Debug("Copilot status change", "params", string(params))
	// Handle status changes here
}

// handleProgress handles Copilot progress notifications
func (c *CopilotClient) handleProgress(params json.RawMessage) {
	logging.Debug("Copilot progress", "params", string(params))
	// Handle progress updates here
}

// handleError handles Copilot error notifications
func (c *CopilotClient) handleError(params json.RawMessage) {
	logging.Error("Copilot error notification", string(params))
	c.mu.Lock()
	c.errorCount++
	c.mu.Unlock()
}

// handleAuthRequest handles authentication requests from the server
func (c *CopilotClient) handleAuthRequest(params json.RawMessage) (interface{}, error) {
	logging.Debug("Copilot auth request", "params", string(params))
	return c.authManager.HandleAuthRequest(params)
}

// handleChatRequest handles chat requests from the server
func (c *CopilotClient) handleChatRequest(params json.RawMessage) (interface{}, error) {
	if !c.chatEnabled {
		return nil, fmt.Errorf("chat is disabled")
	}
	logging.Debug("Copilot chat request", "params", string(params))
	// Handle chat requests here
	return nil, nil
}

// CopilotStatus represents the current status of the Copilot client
type CopilotStatus struct {
	Enabled           bool    `json:"enabled"`
	Authenticated     bool    `json:"authenticated"`
	ServerReady       bool    `json:"server_ready"`
	ChatEnabled       bool    `json:"chat_enabled"`
	CompletionEnabled bool    `json:"completion_enabled"`
	RequestCount      int64   `json:"request_count"`
	ErrorCount        int64   `json:"error_count"`
	AverageLatency    float64 `json:"average_latency"`
}