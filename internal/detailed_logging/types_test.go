package detailed_logging

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewID(t *testing.T) {
	id1 := NewID()
	id2 := NewID()
	
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
}

func TestCalculateDuration(t *testing.T) {
	start := time.Now()
	
	// Test with nil end time
	duration := CalculateDuration(start, nil)
	assert.Equal(t, int64(0), duration)
	
	// Test with valid end time
	end := start.Add(500 * time.Millisecond)
	duration = CalculateDuration(start, &end)
	assert.Equal(t, int64(500), duration)
}

func TestSessionLog(t *testing.T) {
	session := &SessionLog{
		ID:        NewID(),
		StartTime: time.Now(),
		Metadata:  make(map[string]string),
		LLMCalls:  []LLMCallLog{},
		ToolCalls: []ToolCallLog{},
		HTTPCalls: []HTTPLog{},
	}
	
	assert.NotEmpty(t, session.ID)
	assert.Nil(t, session.EndTime)
	assert.Empty(t, session.LLMCalls)
	
	// Add metadata
	session.Metadata["test"] = "value"
	assert.Equal(t, "value", session.Metadata["test"])
}

func TestLLMCallLog(t *testing.T) {
	llmCall := &LLMCallLog{
		ID:        NewID(),
		SessionID: "test-session",
		Provider:  "openai",
		Model:     "gpt-4",
		StartTime: time.Now(),
		Request:   map[string]interface{}{"prompt": "test"},
	}
	
	assert.NotEmpty(t, llmCall.ID)
	assert.Equal(t, "test-session", llmCall.SessionID)
	assert.Equal(t, "openai", llmCall.Provider)
	assert.Equal(t, "gpt-4", llmCall.Model)
	assert.Nil(t, llmCall.EndTime)
	assert.Equal(t, int64(0), llmCall.DurationMs)
}

func TestToolCallLog(t *testing.T) {
	toolCall := &ToolCallLog{
		ID:        NewID(),
		SessionID: "test-session",
		Name:      "test_tool",
		StartTime: time.Now(),
		Input:     map[string]interface{}{"param": "value"},
		ChildIDs:  []string{},
	}
	
	assert.NotEmpty(t, toolCall.ID)
	assert.Equal(t, "test-session", toolCall.SessionID)
	assert.Equal(t, "test_tool", toolCall.Name)
	assert.Empty(t, toolCall.ChildIDs)
	assert.Empty(t, toolCall.ParentID)
}

func TestHTTPLog(t *testing.T) {
	httpLog := &HTTPLog{
		ID:        NewID(),
		SessionID: "test-session",
		Method:    "GET",
		URL:       "https://api.example.com/test",
		Headers:   map[string][]string{"Content-Type": {"application/json"}},
		StartTime: time.Now(),
	}
	
	assert.NotEmpty(t, httpLog.ID)
	assert.Equal(t, "test-session", httpLog.SessionID)
	assert.Equal(t, "GET", httpLog.Method)
	assert.Equal(t, "https://api.example.com/test", httpLog.URL)
	assert.Equal(t, []string{"application/json"}, httpLog.Headers["Content-Type"])
}

func TestTokenUsage(t *testing.T) {
	usage := &TokenUsage{
		Prompt:     100,
		Completion: 50,
		Total:      150,
	}
	
	assert.Equal(t, 100, usage.Prompt)
	assert.Equal(t, 50, usage.Completion)
	assert.Equal(t, 150, usage.Total)
}

func TestStreamEvent(t *testing.T) {
	event := StreamEvent{
		Type:      "content",
		Data:      map[string]interface{}{"text": "Hello"},
		Timestamp: time.Now(),
	}
	
	assert.Equal(t, "content", event.Type)
	assert.Equal(t, "Hello", event.Data["text"])
	assert.NotZero(t, event.Timestamp)
}

func TestStorageMetadata(t *testing.T) {
	now := time.Now()
	later := now.Add(1 * time.Hour)
	
	meta := StorageMetadata{
		ID:            NewID(),
		SessionID:     "test-session",
		StartTime:     now,
		EndTime:       &later,
		LLMCallCount:  5,
		ToolCallCount: 10,
		HTTPCallCount: 3,
		TotalTokens:   1500,
		TotalCost:     0.05,
		HasError:      false,
	}
	
	assert.NotEmpty(t, meta.ID)
	assert.Equal(t, "test-session", meta.SessionID)
	assert.Equal(t, 5, meta.LLMCallCount)
	assert.Equal(t, 10, meta.ToolCallCount)
	assert.Equal(t, 3, meta.HTTPCallCount)
	assert.Equal(t, 1500, meta.TotalTokens)
	assert.Equal(t, 0.05, meta.TotalCost)
	assert.False(t, meta.HasError)
	
	require.NotNil(t, meta.EndTime)
	assert.Equal(t, later, *meta.EndTime)
}