package detailed_logging

import (
	"time"

	"github.com/google/uuid"
)

// SessionLog represents a complete session with all its data
type SessionLog struct {
	ID          string            `json:"id"`
	StartTime   time.Time         `json:"start_time"`
	EndTime     *time.Time        `json:"end_time,omitempty"`
	Metadata    map[string]string `json:"metadata"`
	LLMCalls    []LLMCallLog      `json:"llm_calls"`
	ToolCalls   []ToolCallLog     `json:"tool_calls"`
	HTTPCalls   []HTTPLog         `json:"http_calls"`
	CommandArgs []string          `json:"command_args"`
	UserID      string            `json:"user_id,omitempty"`
}

// LLMCallLog represents a single LLM API call
type LLMCallLog struct {
	ID             string                 `json:"id"`
	SessionID      string                 `json:"session_id"`
	Provider       string                 `json:"provider"`
	Model          string                 `json:"model"`
	StartTime      time.Time              `json:"start_time"`
	EndTime        *time.Time             `json:"end_time,omitempty"`
	Request        map[string]interface{} `json:"request"`
	Response       map[string]interface{} `json:"response,omitempty"`
	StreamEvents   []StreamEvent          `json:"stream_events,omitempty"`
	Error          string                 `json:"error,omitempty"`
	TokensUsed     *TokenUsage            `json:"tokens_used,omitempty"`
	Cost           *float64               `json:"cost,omitempty"`
	DurationMs     int64                  `json:"duration_ms"`
	ParentToolCall string                 `json:"parent_tool_call,omitempty"`
}

// StreamEvent represents a single streaming event
type StreamEvent struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// TokenUsage tracks token consumption
type TokenUsage struct {
	Prompt     int `json:"prompt"`
	Completion int `json:"completion"`
	Total      int `json:"total"`
}

// ToolCallLog represents a tool invocation
type ToolCallLog struct {
	ID           string                 `json:"id"`
	SessionID    string                 `json:"session_id"`
	Name         string                 `json:"name"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      *time.Time             `json:"end_time,omitempty"`
	Input        map[string]interface{} `json:"input"`
	Output       interface{}            `json:"output,omitempty"`
	Error        string                 `json:"error,omitempty"`
	DurationMs   int64                  `json:"duration_ms"`
	ParentID     string                 `json:"parent_id,omitempty"`
	ChildIDs     []string               `json:"child_ids,omitempty"`
	ParentLLMCall string                `json:"parent_llm_call,omitempty"`
}

// HTTPLog represents an HTTP request/response
type HTTPLog struct {
	ID           string                 `json:"id"`
	SessionID    string                 `json:"session_id"`
	Method       string                 `json:"method"`
	URL          string                 `json:"url"`
	Headers      map[string][]string    `json:"headers"`
	Body         interface{}            `json:"body,omitempty"`
	StatusCode   int                    `json:"status_code,omitempty"`
	ResponseBody interface{}            `json:"response_body,omitempty"`
	ResponseHeaders map[string][]string  `json:"response_headers,omitempty"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      *time.Time             `json:"end_time,omitempty"`
	DurationMs   int64                  `json:"duration_ms"`
	Error        string                 `json:"error,omitempty"`
	ParentToolCall string               `json:"parent_tool_call,omitempty"`
}

// StorageMetadata is stored in SQLite for quick queries
type StorageMetadata struct {
	ID            string    `db:"id"`
	SessionID     string    `db:"session_id"`
	StartTime     time.Time `db:"start_time"`
	EndTime       *time.Time `db:"end_time"`
	LLMCallCount  int       `db:"llm_call_count"`
	ToolCallCount int       `db:"tool_call_count"`
	HTTPCallCount int       `db:"http_call_count"`
	TotalTokens   int       `db:"total_tokens"`
	TotalCost     float64   `db:"total_cost"`
	HasError      bool      `db:"has_error"`
}

// Helper function to generate IDs
func NewID() string {
	return uuid.New().String()
}

// Helper function to calculate duration
func CalculateDuration(start time.Time, end *time.Time) int64 {
	if end == nil {
		return 0
	}
	return end.Sub(start).Milliseconds()
}