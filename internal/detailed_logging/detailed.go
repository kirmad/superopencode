package detailed_logging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SessionIDKey is the context key for session IDs
type SessionIDKey struct{}

// Message represents a message in the conversation
type Message interface {
	Role() string
	ContentString() string
	Parts() []ContentPart
}

// ContentPart represents a part of a message
type ContentPart interface {
	IsText() bool
	Text() string
	Type() string
}

// DetailedLogManager manages detailed logging for copilot requests/responses
type DetailedLogManager struct {
	enabled        bool
	baseDir        string
	sessionCounters map[string]int
	mu             sync.RWMutex
}

// RequestContext contains metadata for a specific request
type RequestContext struct {
	RequestID   string            `json:"request_id"`
	SessionID   string            `json:"session_id"`
	RequestIndex int              `json:"request_index"`
	Timestamp   time.Time         `json:"timestamp"`
	ClientInfo  ClientInfo        `json:"client_info"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ClientInfo contains information about the client making the request
type ClientInfo struct {
	Model       string `json:"model"`
	Provider    string `json:"provider"`
	TokenSource string `json:"token_source"`
	APIEndpoint string `json:"api_endpoint"`
	UserAgent   string `json:"user_agent,omitempty"`
}

// RequestData contains the complete request information
type RequestData struct {
	Context  RequestContext `json:"context"`
	Messages []Message      `json:"messages"`
	Options  map[string]any `json:"options,omitempty"`
}

// ResponseData contains the complete response information
type ResponseData struct {
	Context    RequestContext `json:"context"`
	Content    string         `json:"content"`
	TokenUsage TokenUsage     `json:"token_usage,omitempty"`
	Duration   time.Duration  `json:"duration"`
	Error      string         `json:"error,omitempty"`
}

// TokenUsage contains token usage statistics
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

var detailedLogManager *DetailedLogManager
var detailedLogOnce sync.Once
var detailedLogMutex sync.RWMutex

// InitializeDetailedLogging initializes the detailed logging system
func InitializeDetailedLogging(enabled bool, dataDirectory string) *DetailedLogManager {
	detailedLogOnce.Do(func() {
		detailedLogMutex.Lock()
		defer detailedLogMutex.Unlock()
		
		baseDir := filepath.Join(dataDirectory, "detailed-logs")
		if enabled {
			// Ensure base directory exists
			if err := os.MkdirAll(baseDir, 0755); err != nil {
				slog.Warn("failed to create detailed logs directory", "error", err)
				detailedLogManager = &DetailedLogManager{enabled: false}
				return
			}
		}

		detailedLogManager = &DetailedLogManager{
			enabled:         enabled,
			baseDir:         baseDir,
			sessionCounters: make(map[string]int),
		}

		if enabled {
			slog.Info("detailed logging initialized", "base_dir", baseDir)
		}
	})
	return GetDetailedLogManager()
}

// GetDetailedLogManager returns the global detailed log manager
func GetDetailedLogManager() *DetailedLogManager {
	detailedLogMutex.RLock()
	defer detailedLogMutex.RUnlock()
	return detailedLogManager
}

// ResetDetailedLogManager resets the detailed logging manager for testing
func ResetDetailedLogManager() {
	detailedLogMutex.Lock()
	defer detailedLogMutex.Unlock()
	detailedLogManager = nil
	detailedLogOnce = sync.Once{}
}

// IsEnabled returns true if detailed logging is enabled
func (d *DetailedLogManager) IsEnabled() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.enabled
}

// CreateRequestContext creates a new request context with session tracking
func (d *DetailedLogManager) CreateRequestContext(ctx context.Context, clientInfo ClientInfo, sessionKey interface{}) *RequestContext {
	if !d.IsEnabled() {
		return nil
	}

	sessionID := ""
	if sid, ok := ctx.Value(sessionKey).(string); ok {
		sessionID = sid
	}
	if sessionID == "" {
		sessionID = generateSessionID()
	}

	requestIndex := d.getNextRequestIndex(sessionID)

	return &RequestContext{
		RequestID:    generateRequestID(),
		SessionID:    sessionID,
		RequestIndex: requestIndex,
		Timestamp:    time.Now(),
		ClientInfo:   clientInfo,
		Metadata:     make(map[string]string),
	}
}

// LogRequest logs the detailed request information
func (d *DetailedLogManager) LogRequest(reqCtx *RequestContext, messages []Message, options map[string]any) error {
	if !d.IsEnabled() || reqCtx == nil {
		return nil
	}

	requestData := RequestData{
		Context:  *reqCtx,
		Messages: messages,
		Options:  options,
	}

	logDir := d.getLogDirectory(reqCtx.SessionID, reqCtx.RequestIndex)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory %s: %w", logDir, err)
	}

	// Write JSON request data
	jsonPath := filepath.Join(logDir, "request.json")
	if err := d.writeJSON(jsonPath, requestData); err != nil {
		return fmt.Errorf("failed to write request JSON: %w", err)
	}

	// Write Markdown request data
	mdPath := filepath.Join(logDir, "request.md")
	if err := d.writeRequestMarkdown(mdPath, requestData); err != nil {
		return fmt.Errorf("failed to write request markdown: %w", err)
	}

	slog.Debug("detailed request logged", 
		"session_id", reqCtx.SessionID, 
		"request_index", reqCtx.RequestIndex,
		"request_id", reqCtx.RequestID,
		"log_dir", logDir)

	return nil
}

// LogResponse logs the detailed response information
func (d *DetailedLogManager) LogResponse(reqCtx *RequestContext, content string, tokenUsage TokenUsage, duration time.Duration, err error) error {
	if !d.IsEnabled() || reqCtx == nil {
		return nil
	}

	errorStr := ""
	if err != nil {
		errorStr = err.Error()
	}

	responseData := ResponseData{
		Context:    *reqCtx,
		Content:    content,
		TokenUsage: tokenUsage,
		Duration:   duration,
		Error:      errorStr,
	}

	logDir := d.getLogDirectory(reqCtx.SessionID, reqCtx.RequestIndex)
	
	// Write JSON response data
	jsonPath := filepath.Join(logDir, "response.json")
	if err := d.writeJSON(jsonPath, responseData); err != nil {
		return fmt.Errorf("failed to write response JSON: %w", err)
	}

	// Write Markdown response data
	mdPath := filepath.Join(logDir, "response.md")
	if err := d.writeResponseMarkdown(mdPath, responseData); err != nil {
		return fmt.Errorf("failed to write response markdown: %w", err)
	}

	slog.Debug("detailed response logged", 
		"session_id", reqCtx.SessionID, 
		"request_index", reqCtx.RequestIndex,
		"request_id", reqCtx.RequestID,
		"duration", duration,
		"error", errorStr != "")

	return nil
}

// getNextRequestIndex returns the next request index for a session
func (d *DetailedLogManager) getNextRequestIndex(sessionID string) int {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.sessionCounters[sessionID]++
	return d.sessionCounters[sessionID]
}

// getLogDirectory returns the directory path for a specific session and request
func (d *DetailedLogManager) getLogDirectory(sessionID string, requestIndex int) string {
	return filepath.Join(d.baseDir, sessionID, fmt.Sprintf("%d", requestIndex))
}

// writeJSON writes data as JSON to the specified file
func (d *DetailedLogManager) writeJSON(filepath string, data any) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, jsonData, 0644)
}

// writeRequestMarkdown writes request data in markdown format
func (d *DetailedLogManager) writeRequestMarkdown(filepath string, data RequestData) error {
	content := fmt.Sprintf(`# Request Log

**Request ID:** %s
**Session ID:** %s
**Request Index:** %d
**Timestamp:** %s

## Client Information
- **Model:** %s
- **Provider:** %s
- **Token Source:** %s
- **API Endpoint:** %s

## Messages

`, data.Context.RequestID, data.Context.SessionID, data.Context.RequestIndex, 
		data.Context.Timestamp.Format(time.RFC3339),
		data.Context.ClientInfo.Model, data.Context.ClientInfo.Provider,
		data.Context.ClientInfo.TokenSource, data.Context.ClientInfo.APIEndpoint)

	for i, msg := range data.Messages {
		content += fmt.Sprintf("### Message %d (%s)\n\n", i+1, msg.Role())
		
		for _, part := range msg.Parts() {
			if part.IsText() {
				content += fmt.Sprintf("```\n%s\n```\n\n", part.Text())
			} else {
				content += fmt.Sprintf("**[%s]**\n\n", part.Type())
			}
		}
	}

	if len(data.Options) > 0 {
		content += "## Options\n\n```json\n"
		optionsJSON, _ := json.MarshalIndent(data.Options, "", "  ")
		content += string(optionsJSON)
		content += "\n```\n"
	}

	return os.WriteFile(filepath, []byte(content), 0644)
}

// writeResponseMarkdown writes response data in markdown format
func (d *DetailedLogManager) writeResponseMarkdown(filepath string, data ResponseData) error {
	content := fmt.Sprintf(`# Response Log

**Request ID:** %s
**Session ID:** %s
**Request Index:** %d
**Timestamp:** %s
**Duration:** %s

## Token Usage
- **Prompt Tokens:** %d
- **Completion Tokens:** %d
- **Total Tokens:** %d

`, data.Context.RequestID, data.Context.SessionID, data.Context.RequestIndex,
		data.Context.Timestamp.Format(time.RFC3339), data.Duration,
		data.TokenUsage.PromptTokens, data.TokenUsage.CompletionTokens, data.TokenUsage.TotalTokens)

	if data.Error != "" {
		content += fmt.Sprintf("## Error\n\n```\n%s\n```\n\n", data.Error)
	}

	content += "## Response Content\n\n"
	content += "```\n" + data.Content + "\n```\n"

	return os.WriteFile(filepath, []byte(content), 0644)
}

// Helper functions for generating IDs
func generateSessionID() string {
	return uuid.New().String()
}

func generateRequestID() string {
	return uuid.New().String()
}

// CreateDetailedLogDirectory creates the directory structure for detailed logging
func CreateDetailedLogDirectory(sessionID string, requestIndex int) error {
	manager := GetDetailedLogManager()
	if manager == nil || !manager.IsEnabled() {
		return fmt.Errorf("detailed logging is not enabled")
	}
	
	logDir := manager.getLogDirectory(sessionID, requestIndex)
	return os.MkdirAll(logDir, 0755)
}

// GetNextRequestIndex returns the next request index for a session
func GetNextRequestIndex(sessionID string) int {
	manager := GetDetailedLogManager()
	if manager == nil {
		return 0
	}
	return manager.getNextRequestIndex(sessionID)
}