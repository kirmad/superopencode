package detailed_logging

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DetailedLogger is the main coordinator for detailed logging
type DetailedLogger struct {
	mu            sync.RWMutex
	saveMu        sync.Mutex  // Separate mutex for saving
	enabled       bool
	sessionID     string
	session       *SessionLog
	storage       *Storage
	toolTracker   *ToolTracker
	currentLLMCall string  // Track current LLM call for context
}

// NewDetailedLogger creates a new detailed logger instance
func NewDetailedLogger(enabled bool) (*DetailedLogger, error) {
	if !enabled {
		return &DetailedLogger{enabled: false}, nil
	}

	// Create data directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	dataDir := filepath.Join(homeDir, ".opencode", "detailed_logs")
	
	// Initialize storage
	storage, err := NewStorage(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Create session
	sessionID := NewID()
	session := &SessionLog{
		ID:        sessionID,
		StartTime: time.Now(),
		Metadata:  make(map[string]string),
		LLMCalls:  []LLMCallLog{},
		ToolCalls: []ToolCallLog{},
		HTTPCalls: []HTTPLog{},
	}

	// Create logger instance
	logger := &DetailedLogger{
		enabled:    enabled,
		sessionID:  sessionID,
		session:    session,
		storage:    storage,
	}

	// Initialize tool tracker
	logger.toolTracker = NewToolTracker(sessionID, logger)


	// Install HTTP interceptor
	InstallGlobalInterceptor(logger)

	return logger, nil
}

// IsEnabled returns whether detailed logging is enabled
func (dl *DetailedLogger) IsEnabled() bool {
	if dl == nil {
		return false
	}
	return dl.enabled
}

// SetCommandArgs sets the command arguments for the session
func (dl *DetailedLogger) SetCommandArgs(args []string) {
	if !dl.IsEnabled() {
		return
	}

	dl.mu.Lock()
	defer dl.mu.Unlock()
	dl.session.CommandArgs = args
}

// SetMetadata adds metadata to the session
func (dl *DetailedLogger) SetMetadata(key, value string) {
	if !dl.IsEnabled() {
		return
	}

	dl.mu.Lock()
	defer dl.mu.Unlock()
	dl.session.Metadata[key] = value
}

// LogLLMCall logs an LLM API call
func (dl *DetailedLogger) LogLLMCall(call *LLMCallLog) {
	if !dl.IsEnabled() {
		return
	}

	dl.mu.Lock()
	defer dl.mu.Unlock()
	
	dl.session.LLMCalls = append(dl.session.LLMCalls, *call)
	
	// Save session asynchronously
	go dl.saveSession()
}

// LogToolCall logs a tool invocation
func (dl *DetailedLogger) LogToolCall(call *ToolCallLog) {
	if !dl.IsEnabled() {
		return
	}

	dl.mu.Lock()
	defer dl.mu.Unlock()
	
	dl.session.ToolCalls = append(dl.session.ToolCalls, *call)
	
	// Save session asynchronously
	go dl.saveSession()
}

// LogHTTP logs an HTTP request/response
func (dl *DetailedLogger) LogHTTP(call *HTTPLog) {
	if !dl.IsEnabled() {
		return
	}

	dl.mu.Lock()
	defer dl.mu.Unlock()
	
	dl.session.HTTPCalls = append(dl.session.HTTPCalls, *call)
	
	// Save session asynchronously
	go dl.saveSession()
}

// StartToolCall begins tracking a tool call
func (dl *DetailedLogger) StartToolCall(name string, input map[string]interface{}) string {
	if !dl.IsEnabled() {
		return ""
	}

	return dl.toolTracker.StartToolCall(name, input)
}

// EndToolCall completes tracking of a tool call
func (dl *DetailedLogger) EndToolCall(id string, output interface{}, err error) {
	if !dl.IsEnabled() {
		return
	}

	dl.toolTracker.EndToolCall(id, output, err)
}

// GetCurrentToolCall returns the current tool call ID
func (dl *DetailedLogger) GetCurrentToolCall() string {
	if !dl.IsEnabled() {
		return ""
	}

	return dl.toolTracker.GetCurrentToolCall()
}

// SetCurrentLLMCall sets the current LLM call context
func (dl *DetailedLogger) SetCurrentLLMCall(callID string) {
	if !dl.IsEnabled() {
		return
	}

	dl.mu.Lock()
	defer dl.mu.Unlock()
	dl.currentLLMCall = callID
}

// GetCurrentLLMCall returns the current LLM call ID
func (dl *DetailedLogger) GetCurrentLLMCall() string {
	if !dl.IsEnabled() {
		return ""
	}

	dl.mu.RLock()
	defer dl.mu.RUnlock()
	return dl.currentLLMCall
}

// GetToolTracker returns the tool tracker instance
func (dl *DetailedLogger) GetToolTracker() *ToolTracker {
	if !dl.IsEnabled() {
		return nil
	}

	return dl.toolTracker
}

// EndSession marks the session as complete
func (dl *DetailedLogger) EndSession() {
	if !dl.IsEnabled() {
		return
	}

	dl.mu.Lock()
	endTime := time.Now()
	dl.session.EndTime = &endTime
	dl.mu.Unlock()

	// Final save
	dl.saveSession()

	// Clean up old sessions (retain for 30 days by default)
	go dl.storage.DeleteOldSessions(30)
}

// saveSession persists the current session to storage
func (dl *DetailedLogger) saveSession() {
	// Use save mutex to prevent concurrent saves
	dl.saveMu.Lock()
	defer dl.saveMu.Unlock()
	
	// Create a copy to avoid holding the lock during I/O
	dl.mu.RLock()
	sessionCopy := *dl.session
	dl.mu.RUnlock()
	
	if err := dl.storage.SaveSession(&sessionCopy); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save session: %v\n", err)
	}
}


// Close cleans up resources
func (dl *DetailedLogger) Close() error {
	if !dl.IsEnabled() {
		return nil
	}

	dl.EndSession()
	
	if dl.storage != nil {
		return dl.storage.Close()
	}
	
	return nil
}