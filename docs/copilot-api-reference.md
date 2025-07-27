# GitHub Copilot Language Server API Reference

## Overview

This document provides technical implementation details for integrating the GitHub Copilot language server with SuperOpenCode. It covers the Go API structures, LSP extensions, and implementation patterns needed for development.

## Architecture Overview

```
┌─────────────────────────────────────────┐
│             SuperOpenCode               │
├─────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────────┐  │
│  │ LSP Client  │  │ Copilot Manager │  │
│  └─────────────┘  └─────────────────┘  │
├─────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────────┐  │
│  │    gopls    │  │ Copilot Server  │  │
│  └─────────────┘  └─────────────────┘  │
└─────────────────────────────────────────┘
```

## Core Data Structures

### CopilotConfig

```go
// CopilotConfig holds all Copilot-related configuration
type CopilotConfig struct {
    // Core settings
    EnableCopilot     bool   `json:"enable_copilot" yaml:"enable_copilot"`
    ServerPath        string `json:"server_path" yaml:"server_path"`
    NodePath          string `json:"node_path" yaml:"node_path"`
    UseNativeBinary   bool   `json:"use_native_binary" yaml:"use_native_binary"`
    ReplaceGopls      bool   `json:"replace_gopls" yaml:"replace_gopls"`
    
    // Authentication
    AuthToken         string `json:"auth_token" yaml:"auth_token"`
    
    // Feature flags
    ChatEnabled       bool   `json:"chat_enabled" yaml:"chat_enabled"`
    CompletionEnabled bool   `json:"completion_enabled" yaml:"completion_enabled"`
    
    // Installation
    AutoInstall       bool     `json:"auto_install" yaml:"auto_install"`
    ServerArgs        []string `json:"server_args" yaml:"server_args"`
    Environment       map[string]string `json:"environment" yaml:"environment"`
    
    // Performance
    Timeout           int    `json:"timeout" yaml:"timeout"`
    RetryAttempts     int    `json:"retry_attempts" yaml:"retry_attempts"`
    FallbackToGopls   bool   `json:"fallback_to_gopls" yaml:"fallback_to_gopls"`
    
    // Logging and debugging
    LogLevel          string `json:"log_level" yaml:"log_level"`
    
    // Advanced settings
    Performance       *PerformanceConfig `json:"performance,omitempty" yaml:"performance,omitempty"`
    Security          *SecurityConfig    `json:"security,omitempty" yaml:"security,omitempty"`
    AgentConfig       *AgentConfig       `json:"agent_config,omitempty" yaml:"agent_config,omitempty"`
}

// PerformanceConfig controls performance-related settings
type PerformanceConfig struct {
    MaxCompletionTime    int  `json:"max_completion_time" yaml:"max_completion_time"`
    DebounceDelay        int  `json:"debounce_delay" yaml:"debounce_delay"`
    MaxParallelRequests  int  `json:"max_parallel_requests" yaml:"max_parallel_requests"`
    CacheEnabled         bool `json:"cache_enabled" yaml:"cache_enabled"`
    CacheSize            int  `json:"cache_size" yaml:"cache_size"`
}

// SecurityConfig controls security and privacy settings
type SecurityConfig struct {
    DisableTelemetry bool     `json:"disable_telemetry" yaml:"disable_telemetry"`
    PrivateMode      bool     `json:"private_mode" yaml:"private_mode"`
    AllowedDomains   []string `json:"allowed_domains" yaml:"allowed_domains"`
}

// AgentConfig controls Copilot agent features
type AgentConfig struct {
    CodingAgent        bool `json:"coding_agent" yaml:"coding_agent"`
    DebuggingAgent     bool `json:"debugging_agent" yaml:"debugging_agent"`
    DocumentationAgent bool `json:"documentation_agent" yaml:"documentation_agent"`
}
```

### CopilotClient

```go
// CopilotClient extends the base LSP client with Copilot-specific functionality
type CopilotClient struct {
    *lsp.Client
    
    config       *CopilotConfig
    authManager  *AuthManager
    installer    *Installer
    chatManager  *ChatManager
    agentManager *AgentManager
    
    // State management
    authenticated  atomic.Bool
    serverVersion  string
    capabilities   *CopilotCapabilities
    
    // Performance monitoring
    stats         *CopilotStats
    requestCache  *sync.Map
    
    mu            sync.RWMutex
}

// CopilotCapabilities represents server capabilities specific to Copilot
type CopilotCapabilities struct {
    CompletionProvider      bool                    `json:"completionProvider"`
    ChatProvider           bool                    `json:"chatProvider"`
    AgentProvider          bool                    `json:"agentProvider"`
    CustomMethods          []string                `json:"customMethods"`
    SupportedLanguages     []string                `json:"supportedLanguages"`
    MaxCompletionLength    int                     `json:"maxCompletionLength"`
    SupportsStreamingChat  bool                    `json:"supportsStreamingChat"`
}

// CopilotStats tracks usage and performance metrics
type CopilotStats struct {
    CompletionsRequested   int64         `json:"completions_requested"`
    CompletionsReceived    int64         `json:"completions_received"`
    ChatMessages          int64         `json:"chat_messages"`
    AverageResponseTime   time.Duration `json:"average_response_time"`
    ErrorCount            int64         `json:"error_count"`
    CacheHits             int64         `json:"cache_hits"`
    
    mu                    sync.RWMutex
}
```

## Core Interfaces

### CopilotManager Interface

```go
// CopilotManager defines the main interface for Copilot integration
type CopilotManager interface {
    // Lifecycle management
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Restart(ctx context.Context) error
    IsReady() bool
    
    // LSP operations
    Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error)
    Completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error)
    Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error)
    
    // Copilot-specific operations
    Chat(ctx context.Context, message string) (*ChatResponse, error)
    GetAgents(ctx context.Context) ([]Agent, error)
    InvokeAgent(ctx context.Context, agentID string, params AgentParams) (*AgentResponse, error)
    
    // Configuration and status
    UpdateConfig(config *CopilotConfig) error
    GetStatus() *CopilotStatus
    GetStats() *CopilotStats
}
```

### Authentication Interface

```go
// AuthManager handles GitHub authentication for Copilot
type AuthManager interface {
    // Authentication flow
    Authenticate(ctx context.Context) error
    IsAuthenticated() bool
    GetToken() (string, error)
    RefreshToken(ctx context.Context) error
    Logout() error
    
    // Token management
    StoreToken(token string) error
    ValidateToken(ctx context.Context, token string) error
    
    // User information
    GetUser(ctx context.Context) (*GitHubUser, error)
    GetCopilotAccess(ctx context.Context) (*CopilotAccess, error)
}

// GitHubUser represents authenticated user information
type GitHubUser struct {
    Login     string `json:"login"`
    Email     string `json:"email"`
    Name      string `json:"name"`
    AvatarURL string `json:"avatar_url"`
}

// CopilotAccess represents user's Copilot access information
type CopilotAccess struct {
    HasAccess      bool      `json:"has_access"`
    SubscriptionID string    `json:"subscription_id"`
    PlanType       string    `json:"plan_type"`
    ExpiresAt      time.Time `json:"expires_at"`
}
```

### Installation Interface

```go
// Installer handles Copilot language server installation
type Installer interface {
    // Installation operations
    Install(ctx context.Context) error
    Uninstall(ctx context.Context) error
    Update(ctx context.Context) error
    IsInstalled() bool
    
    // Detection and validation
    DetectServerPath() (string, error)
    ValidateInstallation() error
    GetInstalledVersion() (string, error)
    
    // Path management
    GetServerExecutable() (string, error)
    GetNodeExecutable() (string, error)
}

// InstallationStatus represents current installation state
type InstallationStatus struct {
    Installed        bool      `json:"installed"`
    Version          string    `json:"version"`
    ServerPath       string    `json:"server_path"`
    NodePath         string    `json:"node_path"`
    UsingNativeBinary bool     `json:"using_native_binary"`
    LastChecked      time.Time `json:"last_checked"`
}
```

## LSP Protocol Extensions

### Copilot-Specific Methods

#### Completion Request
```go
// CopilotCompletionParams extends standard completion parameters
type CopilotCompletionParams struct {
    protocol.CompletionParams
    
    // Copilot-specific fields
    MaxLength       int                    `json:"maxLength,omitempty"`
    Temperature     float64                `json:"temperature,omitempty"`
    Context         *CompletionContext     `json:"context,omitempty"`
    Stream          bool                   `json:"stream,omitempty"`
}

// CompletionContext provides additional context for AI completions
type CompletionContext struct {
    FileHistory     []FileHistoryItem      `json:"fileHistory,omitempty"`
    ProjectInfo     *ProjectInfo           `json:"projectInfo,omitempty"`
    RecentEdits     []EditInfo             `json:"recentEdits,omitempty"`
    UserPreferences *UserPreferences       `json:"userPreferences,omitempty"`
}

// CopilotCompletionItem extends standard completion items
type CopilotCompletionItem struct {
    protocol.CompletionItem
    
    // Copilot-specific fields
    Confidence      float64                `json:"confidence,omitempty"`
    Source          string                 `json:"source,omitempty"`
    Suggestion      *SuggestionInfo        `json:"suggestion,omitempty"`
}
```

#### Chat Request
```go
// ChatParams represents parameters for chat requests
type ChatParams struct {
    Message         string                 `json:"message"`
    ConversationID  string                 `json:"conversationId,omitempty"`
    Context         *ChatContext           `json:"context,omitempty"`
    Stream          bool                   `json:"stream,omitempty"`
}

// ChatContext provides context for chat interactions
type ChatContext struct {
    CurrentFile     *protocol.TextDocumentIdentifier `json:"currentFile,omitempty"`
    SelectedText    string                           `json:"selectedText,omitempty"`
    RecentFiles     []string                         `json:"recentFiles,omitempty"`
    WorkspaceInfo   *WorkspaceInfo                   `json:"workspaceInfo,omitempty"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
    ConversationID  string                 `json:"conversationId"`
    Message         string                 `json:"message"`
    MessageID       string                 `json:"messageId"`
    IsComplete      bool                   `json:"isComplete"`
    SuggestedActions []SuggestedAction     `json:"suggestedActions,omitempty"`
}

// SuggestedAction represents an action that can be taken based on chat
type SuggestedAction struct {
    Type            string                 `json:"type"`
    Title           string                 `json:"title"`
    Command         string                 `json:"command"`
    Arguments       []interface{}          `json:"arguments,omitempty"`
}
```

### Custom Notifications

```go
// Copilot-specific notification types
const (
    CopilotStatusNotification        = "copilot/statusNotification"
    CopilotCompletionNotification    = "copilot/completionNotification"
    CopilotChatNotification         = "copilot/chatNotification"
    CopilotAuthNotification         = "copilot/authNotification"
)

// CopilotStatusNotification represents server status updates
type CopilotStatusNotification struct {
    Status          string                 `json:"status"`
    Message         string                 `json:"message"`
    Timestamp       time.Time              `json:"timestamp"`
    ServerVersion   string                 `json:"serverVersion,omitempty"`
}

// CopilotAuthNotification represents authentication status changes
type CopilotAuthNotification struct {
    Authenticated   bool                   `json:"authenticated"`
    User            *GitHubUser            `json:"user,omitempty"`
    Error           string                 `json:"error,omitempty"`
}
```

## Implementation Guide

### 1. Basic Integration

```go
package copilot

import (
    "context"
    "fmt"
    
    "github.com/kirmad/superopencode/internal/lsp"
    "github.com/kirmad/superopencode/internal/lsp/protocol"
)

// NewCopilotClient creates a new Copilot client instance
func NewCopilotClient(ctx context.Context, config *CopilotConfig) (*CopilotClient, error) {
    // Validate configuration
    if err := validateConfig(config); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }
    
    // Create base LSP client
    serverPath, err := getServerExecutable(config)
    if err != nil {
        return nil, fmt.Errorf("failed to get server executable: %w", err)
    }
    
    args := buildServerArgs(config)
    baseClient, err := lsp.NewClient(ctx, serverPath, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to create LSP client: %w", err)
    }
    
    // Create Copilot-specific components
    client := &CopilotClient{
        Client:       baseClient,
        config:       config,
        authManager:  NewAuthManager(config),
        installer:    NewInstaller(config),
        chatManager:  NewChatManager(),
        agentManager: NewAgentManager(),
        stats:        &CopilotStats{},
        requestCache: &sync.Map{},
    }
    
    // Register Copilot-specific handlers
    client.registerHandlers()
    
    return client, nil
}

// registerHandlers sets up Copilot-specific LSP handlers
func (c *CopilotClient) registerHandlers() {
    c.RegisterNotificationHandler(CopilotStatusNotification, c.handleStatusNotification)
    c.RegisterNotificationHandler(CopilotAuthNotification, c.handleAuthNotification)
    c.RegisterServerRequestHandler("copilot/authRequest", c.handleAuthRequest)
}
```

### 2. Server Type Integration

```go
// Add to internal/lsp/client.go

const (
    ServerTypeUnknown ServerType = iota
    ServerTypeGo
    ServerTypeTypeScript
    ServerTypeRust
    ServerTypePython
    ServerTypeCopilot  // Add this
    ServerTypeGeneric
)

// Update detectServerType method
func (c *Client) detectServerType() ServerType {
    if c.Cmd == nil {
        return ServerTypeUnknown
    }

    cmdPath := strings.ToLower(c.Cmd.Path)

    switch {
    case strings.Contains(cmdPath, "copilot-language-server"):
        return ServerTypeCopilot
    case strings.Contains(cmdPath, "gopls"):
        return ServerTypeGo
    case strings.Contains(cmdPath, "typescript") || strings.Contains(cmdPath, "vtsls") || strings.Contains(cmdPath, "tsserver"):
        return ServerTypeTypeScript
    case strings.Contains(cmdPath, "rust-analyzer"):
        return ServerTypeRust
    case strings.Contains(cmdPath, "pyright") || strings.Contains(cmdPath, "pylsp") || strings.Contains(cmdPath, "python"):
        return ServerTypePython
    default:
        return ServerTypeGeneric
    }
}
```

### 3. Configuration Integration

```go
// Add to internal/config/config.go

type Config struct {
    // ... existing fields ...
    
    Copilot *CopilotConfig `json:"copilot,omitempty" yaml:"copilot,omitempty"`
}

// LoadConfig updates to handle Copilot configuration
func LoadConfig() (*Config, error) {
    // ... existing logic ...
    
    // Set Copilot defaults
    if config.Copilot == nil {
        config.Copilot = &CopilotConfig{}
    }
    setDefaults(config.Copilot)
    
    return config, nil
}

func setDefaults(config *CopilotConfig) {
    if config.ServerPath == "" {
        config.ServerPath = "auto"
    }
    if config.NodePath == "" {
        config.NodePath = "node"
    }
    if config.Timeout == 0 {
        config.Timeout = 30000
    }
    if config.RetryAttempts == 0 {
        config.RetryAttempts = 3
    }
    // ... other defaults ...
}
```

### 4. Authentication Implementation

```go
package copilot

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
)

type AuthManager struct {
    config    *CopilotConfig
    client    *http.Client
    tokenStore TokenStore
}

// Authenticate handles the authentication flow
func (am *AuthManager) Authenticate(ctx context.Context) error {
    // Check for existing token
    if token, err := am.GetToken(); err == nil && token != "" {
        if err := am.ValidateToken(ctx, token); err == nil {
            return nil
        }
    }
    
    // Start OAuth flow or use provided token
    if am.config.AuthToken != "" {
        return am.StoreToken(am.config.AuthToken)
    }
    
    return am.startOAuthFlow(ctx)
}

// startOAuthFlow initiates the GitHub OAuth flow
func (am *AuthManager) startOAuthFlow(ctx context.Context) error {
    // Implementation for OAuth flow
    // This would typically:
    // 1. Open browser to GitHub OAuth
    // 2. Start local server to receive callback
    // 3. Exchange code for token
    // 4. Store token securely
    
    // Simplified implementation for reference
    fmt.Println("Please authenticate with GitHub Copilot...")
    fmt.Println("Visit: https://github.com/login/oauth/authorize?client_id=YOUR_CLIENT_ID")
    
    // In real implementation, handle the full OAuth flow
    return nil
}

// ValidateToken checks if a token is valid for Copilot access
func (am *AuthManager) ValidateToken(ctx context.Context, token string) error {
    req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
    if err != nil {
        return err
    }
    
    req.Header.Set("Authorization", "token "+token)
    req.Header.Set("Accept", "application/vnd.github.v3+json")
    
    resp, err := am.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("token validation failed: %d", resp.StatusCode)
    }
    
    return nil
}
```

### 5. Installation Management

```go
package copilot

import (
    "context"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
)

type Installer struct {
    config *CopilotConfig
}

// Install installs the Copilot language server
func (i *Installer) Install(ctx context.Context) error {
    if i.IsInstalled() {
        return nil
    }
    
    // Install via npm
    cmd := exec.CommandContext(ctx, "npm", "install", "-g", "@github/copilot-language-server")
    
    // Set environment
    if i.config.Environment != nil {
        env := os.Environ()
        for k, v := range i.config.Environment {
            env = append(env, fmt.Sprintf("%s=%s", k, v))
        }
        cmd.Env = env
    }
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("npm install failed: %w\nOutput: %s", err, output)
    }
    
    return i.ValidateInstallation()
}

// DetectServerPath finds the Copilot language server executable
func (i *Installer) DetectServerPath() (string, error) {
    // Check explicit path first
    if i.config.ServerPath != "auto" && i.config.ServerPath != "global" {
        if _, err := os.Stat(i.config.ServerPath); err == nil {
            return i.config.ServerPath, nil
        }
    }
    
    // Search common installation paths
    searchPaths := []string{
        // Global npm installation
        filepath.Join(os.Getenv("HOME"), ".npm-global", "bin", "copilot-language-server"),
        "/usr/local/bin/copilot-language-server",
        
        // Local installation
        "./node_modules/.bin/copilot-language-server",
        "./node_modules/@github/copilot-language-server/dist/language-server.js",
        
        // Platform-specific paths
        i.getPlatformSpecificPath(),
    }
    
    for _, path := range searchPaths {
        if _, err := os.Stat(path); err == nil {
            return path, nil
        }
    }
    
    return "", fmt.Errorf("Copilot language server not found")
}

// getPlatformSpecificPath returns platform-specific binary path
func (i *Installer) getPlatformSpecificPath() string {
    var platform string
    switch runtime.GOOS {
    case "darwin":
        if runtime.GOARCH == "arm64" {
            platform = "darwin-arm64"
        } else {
            platform = "darwin-x64"
        }
    case "linux":
        platform = "linux-x64"
    case "windows":
        platform = "win32-x64"
    default:
        return ""
    }
    
    return filepath.Join("node_modules", "@github", "copilot-language-server", "native", platform, "copilot-language-server")
}
```

## Error Handling

### Error Types

```go
// CopilotError represents Copilot-specific errors
type CopilotError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Type    string `json:"type"`
    Data    interface{} `json:"data,omitempty"`
}

func (e *CopilotError) Error() string {
    return fmt.Sprintf("Copilot error %d (%s): %s", e.Code, e.Type, e.Message)
}

// Error codes
const (
    ErrAuthenticationFailed  = 1001
    ErrServerNotInstalled   = 1002
    ErrServerStartupFailed  = 1003
    ErrCompletionTimeout    = 1004
    ErrInvalidConfiguration = 1005
    ErrNetworkError         = 1006
    ErrPermissionDenied     = 1007
)

// Common error constructors
func NewAuthError(message string) *CopilotError {
    return &CopilotError{
        Code:    ErrAuthenticationFailed,
        Message: message,
        Type:    "authentication",
    }
}

func NewInstallationError(message string) *CopilotError {
    return &CopilotError{
        Code:    ErrServerNotInstalled,
        Message: message,
        Type:    "installation",
    }
}
```

### Error Recovery

```go
// ErrorRecovery handles automatic error recovery
type ErrorRecovery struct {
    client          *CopilotClient
    maxRetries      int
    backoffStrategy BackoffStrategy
}

// RecoverFromError attempts to recover from various error conditions
func (er *ErrorRecovery) RecoverFromError(ctx context.Context, err error) error {
    var copilotErr *CopilotError
    if !errors.As(err, &copilotErr) {
        return err // Not a Copilot error, can't recover
    }
    
    switch copilotErr.Code {
    case ErrAuthenticationFailed:
        return er.recoverFromAuthError(ctx)
    case ErrServerStartupFailed:
        return er.recoverFromServerError(ctx)
    case ErrCompletionTimeout:
        return er.recoverFromTimeoutError(ctx)
    default:
        return err
    }
}

func (er *ErrorRecovery) recoverFromAuthError(ctx context.Context) error {
    // Attempt to re-authenticate
    return er.client.authManager.Authenticate(ctx)
}

func (er *ErrorRecovery) recoverFromServerError(ctx context.Context) error {
    // Restart the server
    if err := er.client.Stop(ctx); err != nil {
        return err
    }
    return er.client.Start(ctx)
}
```

## Testing Framework

### Unit Test Structure

```go
package copilot

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockAuthManager for testing
type MockAuthManager struct {
    mock.Mock
}

func (m *MockAuthManager) Authenticate(ctx context.Context) error {
    args := m.Called(ctx)
    return args.Error(0)
}

func (m *MockAuthManager) IsAuthenticated() bool {
    args := m.Called()
    return args.Bool(0)
}

// Test basic client creation
func TestNewCopilotClient(t *testing.T) {
    config := &CopilotConfig{
        EnableCopilot: true,
        ServerPath:    "/mock/path",
    }
    
    ctx := context.Background()
    client, err := NewCopilotClient(ctx, config)
    
    assert.NoError(t, err)
    assert.NotNil(t, client)
    assert.Equal(t, config, client.config)
}

// Test authentication flow
func TestAuthentication(t *testing.T) {
    mockAuth := &MockAuthManager{}
    mockAuth.On("Authenticate", mock.Anything).Return(nil)
    mockAuth.On("IsAuthenticated").Return(true)
    
    client := &CopilotClient{
        authManager: mockAuth,
    }
    
    err := client.authManager.Authenticate(context.Background())
    assert.NoError(t, err)
    
    authenticated := client.authManager.IsAuthenticated()
    assert.True(t, authenticated)
    
    mockAuth.AssertExpectations(t)
}

// Integration test example
func TestCopilotIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    config := &CopilotConfig{
        EnableCopilot:   true,
        ServerPath:      "auto",
        AutoInstall:     true,
        Timeout:         30000,
        FallbackToGopls: true,
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()
    
    client, err := NewCopilotClient(ctx, config)
    assert.NoError(t, err)
    
    err = client.Start(ctx)
    assert.NoError(t, err)
    defer client.Stop(ctx)
    
    // Test basic functionality
    status := client.GetStatus()
    assert.NotNil(t, status)
    assert.True(t, client.IsReady())
}
```

### Benchmark Tests

```go
// Benchmark completion performance
func BenchmarkCopilotCompletion(b *testing.B) {
    client := setupTestClient(b)
    defer client.Stop(context.Background())
    
    params := &protocol.CompletionParams{
        TextDocument: protocol.TextDocumentIdentifier{
            URI: "file:///test.go",
        },
        Position: protocol.Position{
            Line:      10,
            Character: 5,
        },
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        _, err := client.Completion(ctx, params)
        cancel()
        
        if err != nil {
            b.Fatalf("Completion failed: %v", err)
        }
    }
}
```

## Configuration Validation

```go
// ValidateConfig ensures configuration is valid
func ValidateConfig(config *CopilotConfig) error {
    var errors []string
    
    if config.Timeout <= 0 {
        errors = append(errors, "timeout must be positive")
    }
    
    if config.RetryAttempts < 1 || config.RetryAttempts > 10 {
        errors = append(errors, "retry_attempts must be between 1 and 10")
    }
    
    if config.Performance != nil {
        if config.Performance.MaxCompletionTime <= 0 {
            errors = append(errors, "max_completion_time must be positive")
        }
        
        if config.Performance.MaxParallelRequests < 1 || config.Performance.MaxParallelRequests > 20 {
            errors = append(errors, "max_parallel_requests must be between 1 and 20")
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, ", "))
    }
    
    return nil
}
```

## Migration Utilities

```go
// MigrationManager handles migration from gopls to Copilot
type MigrationManager struct {
    currentConfig *Config
    backupConfig  *Config
    migrationLog  []MigrationStep
    
    copilotClient *CopilotClient
    goplsClient   *lsp.Client
}

// MigrationStep represents a single migration operation
type MigrationStep struct {
    Timestamp   time.Time `json:"timestamp"`
    Action      string    `json:"action"`
    Description string    `json:"description"`
    Success     bool      `json:"success"`
    Error       string    `json:"error,omitempty"`
}

// MigrateToCopilot performs the migration process
func (mm *MigrationManager) MigrateToCopilot(ctx context.Context) error {
    // Step 1: Backup current configuration
    if err := mm.backupConfiguration(); err != nil {
        return fmt.Errorf("failed to backup configuration: %w", err)
    }
    mm.logStep("backup", "Configuration backed up", true, "")
    
    // Step 2: Install Copilot language server
    installer := NewInstaller(mm.currentConfig.Copilot)
    if err := installer.Install(ctx); err != nil {
        mm.logStep("install", "Install Copilot language server", false, err.Error())
        return fmt.Errorf("failed to install Copilot: %w", err)
    }
    mm.logStep("install", "Copilot language server installed", true, "")
    
    // Step 3: Test Copilot functionality
    if err := mm.testCopilotFunctionality(ctx); err != nil {
        mm.logStep("test", "Test Copilot functionality", false, err.Error())
        return fmt.Errorf("Copilot functionality test failed: %w", err)
    }
    mm.logStep("test", "Copilot functionality verified", true, "")
    
    // Step 4: Update configuration
    if err := mm.updateConfiguration(); err != nil {
        mm.logStep("config", "Update configuration", false, err.Error())
        return fmt.Errorf("failed to update configuration: %w", err)
    }
    mm.logStep("config", "Configuration updated", true, "")
    
    return nil
}

// Rollback reverts migration changes
func (mm *MigrationManager) Rollback(ctx context.Context) error {
    if mm.backupConfig == nil {
        return fmt.Errorf("no backup configuration available")
    }
    
    // Restore original configuration
    mm.currentConfig = mm.backupConfig
    
    // Save restored configuration
    return mm.saveConfiguration()
}

func (mm *MigrationManager) logStep(action, description string, success bool, errorMsg string) {
    step := MigrationStep{
        Timestamp:   time.Now(),
        Action:      action,
        Description: description,
        Success:     success,
        Error:       errorMsg,
    }
    mm.migrationLog = append(mm.migrationLog, step)
}
```

## Performance Monitoring

```go
// PerformanceMonitor tracks Copilot performance metrics
type PerformanceMonitor struct {
    stats        *CopilotStats
    requestTimes []time.Duration
    mu           sync.RWMutex
}

// RecordRequest records a request and its completion time
func (pm *PerformanceMonitor) RecordRequest(requestType string, duration time.Duration, success bool) {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    
    pm.requestTimes = append(pm.requestTimes, duration)
    
    // Update stats
    if success {
        switch requestType {
        case "completion":
            pm.stats.CompletionsReceived++
        case "chat":
            pm.stats.ChatMessages++
        }
    } else {
        pm.stats.ErrorCount++
    }
    
    // Update average response time
    pm.updateAverageResponseTime()
}

func (pm *PerformanceMonitor) updateAverageResponseTime() {
    if len(pm.requestTimes) == 0 {
        return
    }
    
    var total time.Duration
    for _, t := range pm.requestTimes {
        total += t
    }
    
    pm.stats.AverageResponseTime = total / time.Duration(len(pm.requestTimes))
}

// GetMetrics returns current performance metrics
func (pm *PerformanceMonitor) GetMetrics() map[string]interface{} {
    pm.mu.RLock()
    defer pm.mu.RUnlock()
    
    return map[string]interface{}{
        "total_requests":       pm.stats.CompletionsRequested + pm.stats.ChatMessages,
        "successful_requests":  pm.stats.CompletionsReceived + pm.stats.ChatMessages,
        "error_rate":          float64(pm.stats.ErrorCount) / float64(pm.stats.CompletionsRequested + pm.stats.ChatMessages),
        "average_response_time": pm.stats.AverageResponseTime.Milliseconds(),
        "cache_hit_rate":      float64(pm.stats.CacheHits) / float64(pm.stats.CompletionsRequested),
    }
}
```

This API reference provides the complete technical foundation needed to implement GitHub Copilot language server integration in SuperOpenCode. The structures, interfaces, and examples give developers everything needed to build a robust, performant integration that maintains compatibility with the existing architecture.