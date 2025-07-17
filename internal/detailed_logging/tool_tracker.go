package detailed_logging

import (
	"sync"
	"time"
)

// ToolTracker manages hierarchical tool call tracking
type ToolTracker struct {
	mu            sync.RWMutex
	activeCalls   map[string]*ToolCallLog
	callStack     []string
	sessionID     string
	logger        *DetailedLogger
	threadLocal   sync.Map // goroutine ID -> tool call ID
}

// NewToolTracker creates a new tool tracker
func NewToolTracker(sessionID string, logger *DetailedLogger) *ToolTracker {
	return &ToolTracker{
		activeCalls: make(map[string]*ToolCallLog),
		callStack:   []string{},
		sessionID:   sessionID,
		logger:      logger,
	}
}

// StartToolCall begins tracking a new tool call
func (tt *ToolTracker) StartToolCall(name string, input map[string]interface{}) string {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	// Create new tool call log
	toolLog := &ToolCallLog{
		ID:        NewID(),
		SessionID: tt.sessionID,
		Name:      name,
		StartTime: time.Now(),
		Input:     input,
		ChildIDs:  []string{},
	}

	// Set parent if there's an active call
	if len(tt.callStack) > 0 {
		parentID := tt.callStack[len(tt.callStack)-1]
		toolLog.ParentID = parentID
		
		// Add this as a child to parent
		if parent, ok := tt.activeCalls[parentID]; ok {
			parent.ChildIDs = append(parent.ChildIDs, toolLog.ID)
		}
	}

	// Set parent LLM call if available
	if llmCallID := tt.getCurrentLLMCall(); llmCallID != "" {
		toolLog.ParentLLMCall = llmCallID
	}

	// Store and push to stack
	tt.activeCalls[toolLog.ID] = toolLog
	tt.callStack = append(tt.callStack, toolLog.ID)

	// Store in thread-local for current goroutine
	tt.setCurrentToolCall(toolLog.ID)

	return toolLog.ID
}

// EndToolCall completes tracking of a tool call
func (tt *ToolTracker) EndToolCall(id string, output interface{}, err error) {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	toolLog, ok := tt.activeCalls[id]
	if !ok {
		return
	}

	// Complete the log entry
	endTime := time.Now()
	toolLog.EndTime = &endTime
	toolLog.DurationMs = CalculateDuration(toolLog.StartTime, toolLog.EndTime)
	toolLog.Output = output
	
	if err != nil {
		toolLog.Error = err.Error()
	}

	// Remove from stack
	for i := len(tt.callStack) - 1; i >= 0; i-- {
		if tt.callStack[i] == id {
			tt.callStack = append(tt.callStack[:i], tt.callStack[i+1:]...)
			break
		}
	}

	// Log the completed tool call
	if tt.logger != nil {
		tt.logger.LogToolCall(toolLog)
	}

	// Remove from active calls
	delete(tt.activeCalls, id)

	// Clear thread-local if this was the current call
	if currentID, _ := tt.getCurrentToolCallFromThread(); currentID == id {
		tt.clearCurrentToolCall()
	}
}

// GetCurrentToolCall returns the ID of the currently active tool call
func (tt *ToolTracker) GetCurrentToolCall() string {
	// First check thread-local storage
	if id, ok := tt.getCurrentToolCallFromThread(); ok {
		return id
	}

	// Fall back to stack
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if len(tt.callStack) > 0 {
		return tt.callStack[len(tt.callStack)-1]
	}
	return ""
}

// GetActiveCallStack returns the current call stack
func (tt *ToolTracker) GetActiveCallStack() []string {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	result := make([]string, len(tt.callStack))
	copy(result, tt.callStack)
	return result
}

// GetToolCall retrieves a tool call by ID
func (tt *ToolTracker) GetToolCall(id string) *ToolCallLog {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if call, ok := tt.activeCalls[id]; ok {
		// Return a copy to prevent modification
		copyCall := *call
		return &copyCall
	}
	return nil
}

// Thread-local storage helpers
func (tt *ToolTracker) setCurrentToolCall(id string) {
	goroutineID := getGoroutineID()
	tt.threadLocal.Store(goroutineID, id)
}

func (tt *ToolTracker) getCurrentToolCallFromThread() (string, bool) {
	goroutineID := getGoroutineID()
	if val, ok := tt.threadLocal.Load(goroutineID); ok {
		return val.(string), true
	}
	return "", false
}

func (tt *ToolTracker) clearCurrentToolCall() {
	goroutineID := getGoroutineID()
	tt.threadLocal.Delete(goroutineID)
}

func (tt *ToolTracker) getCurrentLLMCall() string {
	// This would be set by the logger when an LLM call is active
	// For now, return empty
	return ""
}

// getGoroutineID returns the current goroutine ID
// This is a simplified version - in production you'd use a more robust method
func getGoroutineID() string {
	// In a real implementation, you'd extract the goroutine ID
	// For now, we'll use a simple approach
	return "main"
}