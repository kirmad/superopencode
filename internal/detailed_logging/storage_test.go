package detailed_logging

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStorage(t *testing.T) {
	tempDir := t.TempDir()
	
	storage, err := NewStorage(tempDir)
	require.NoError(t, err)
	require.NotNil(t, storage)
	defer storage.Close()
	
	// Check that database file was created
	dbPath := filepath.Join(tempDir, "sessions.db")
	_, err = os.Stat(dbPath)
	assert.NoError(t, err)
}

func TestStorageSaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewStorage(tempDir)
	require.NoError(t, err)
	defer storage.Close()
	
	// Create a test session
	now := time.Now()
	endTime := now.Add(1 * time.Hour)
	session := &SessionLog{
		ID:        "test-session-123",
		StartTime: now,
		EndTime:   &endTime,
		Metadata: map[string]string{
			"test": "value",
			"key":  "data",
		},
		CommandArgs: []string{"opencode", "--debug"},
		LLMCalls: []LLMCallLog{
			{
				ID:        "llm-1",
				SessionID: "test-session-123",
				Provider:  "openai",
				Model:     "gpt-4",
				StartTime: now,
				EndTime:   &endTime,
				Request: map[string]interface{}{
					"prompt": "test prompt",
				},
				Response: map[string]interface{}{
					"text": "test response",
				},
				TokensUsed: &TokenUsage{
					Prompt:     100,
					Completion: 50,
					Total:      150,
				},
				Cost:       ptrFloat64(0.05),
				DurationMs: 1000,
			},
		},
		ToolCalls: []ToolCallLog{
			{
				ID:        "tool-1",
				SessionID: "test-session-123",
				Name:      "test_tool",
				StartTime: now,
				EndTime:   &endTime,
				Input: map[string]interface{}{
					"param": "value",
				},
				Output:     "tool output",
				DurationMs: 500,
			},
		},
		HTTPCalls: []HTTPLog{
			{
				ID:        "http-1",
				SessionID: "test-session-123",
				Method:    "GET",
				URL:       "https://api.example.com/test",
				Headers: map[string][]string{
					"Content-Type": {"application/json"},
				},
				StatusCode: 200,
				StartTime:  now,
				EndTime:    &endTime,
				DurationMs: 200,
			},
		},
	}
	
	// Save the session
	err = storage.SaveSession(session)
	require.NoError(t, err)
	
	// Load the session back
	loaded, err := storage.LoadSession("test-session-123")
	require.NoError(t, err)
	require.NotNil(t, loaded)
	
	// Verify loaded data
	assert.Equal(t, session.ID, loaded.ID)
	assert.Equal(t, session.StartTime.Unix(), loaded.StartTime.Unix())
	assert.Equal(t, session.EndTime.Unix(), loaded.EndTime.Unix())
	assert.Equal(t, session.Metadata, loaded.Metadata)
	assert.Equal(t, session.CommandArgs, loaded.CommandArgs)
	
	// Verify LLM calls
	require.Len(t, loaded.LLMCalls, 1)
	assert.Equal(t, "llm-1", loaded.LLMCalls[0].ID)
	assert.Equal(t, "openai", loaded.LLMCalls[0].Provider)
	assert.Equal(t, "gpt-4", loaded.LLMCalls[0].Model)
	assert.Equal(t, 0.05, *loaded.LLMCalls[0].Cost)
	
	// Verify tool calls
	require.Len(t, loaded.ToolCalls, 1)
	assert.Equal(t, "tool-1", loaded.ToolCalls[0].ID)
	assert.Equal(t, "test_tool", loaded.ToolCalls[0].Name)
	
	// Verify HTTP calls
	require.Len(t, loaded.HTTPCalls, 1)
	assert.Equal(t, "http-1", loaded.HTTPCalls[0].ID)
	assert.Equal(t, "GET", loaded.HTTPCalls[0].Method)
	assert.Equal(t, 200, loaded.HTTPCalls[0].StatusCode)
}

func TestStorageListSessions(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewStorage(tempDir)
	require.NoError(t, err)
	defer storage.Close()
	
	// Create multiple test sessions
	now := time.Now()
	for i := 0; i < 3; i++ {
		endTime := now.Add(time.Duration(i+1) * time.Hour)
		session := &SessionLog{
			ID:        fmt.Sprintf("session-%d", i),
			StartTime: now.Add(time.Duration(i) * time.Hour),
			EndTime:   &endTime,
			Metadata:  map[string]string{},
			LLMCalls:  []LLMCallLog{},
			ToolCalls: []ToolCallLog{},
			HTTPCalls: []HTTPLog{},
		}
		
		// Add error to second session
		if i == 1 {
			session.LLMCalls = append(session.LLMCalls, LLMCallLog{
				ID:        "llm-error",
				SessionID: session.ID,
				Error:     "test error",
			})
		}
		
		err = storage.SaveSession(session)
		require.NoError(t, err)
	}
	
	// Test listing all sessions
	sessions, err := storage.ListSessions(SessionFilters{})
	require.NoError(t, err)
	assert.Len(t, sessions, 3)
	
	// Test listing with limit
	sessions, err = storage.ListSessions(SessionFilters{Limit: 2})
	require.NoError(t, err)
	assert.Len(t, sessions, 2)
	
	// Test filtering by error
	hasError := true
	sessions, err = storage.ListSessions(SessionFilters{HasError: &hasError})
	require.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Equal(t, "session-1", sessions[0].ID)
	
	// Test filtering by time range
	startTime := now.Add(30 * time.Minute)
	sessions, err = storage.ListSessions(SessionFilters{StartTime: &startTime})
	require.NoError(t, err)
	assert.Len(t, sessions, 2) // sessions 1 and 2
}

func TestStorageDeleteOldSessions(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewStorage(tempDir)
	require.NoError(t, err)
	defer storage.Close()
	
	// Create old and new sessions
	oldTime := time.Now().AddDate(0, 0, -40) // 40 days ago
	newTime := time.Now()
	
	oldSession := &SessionLog{
		ID:        "old-session",
		StartTime: oldTime,
		EndTime:   &oldTime,
		Metadata:  map[string]string{},
		LLMCalls:  []LLMCallLog{},
		ToolCalls: []ToolCallLog{},
		HTTPCalls: []HTTPLog{},
	}
	
	newSession := &SessionLog{
		ID:        "new-session",
		StartTime: newTime,
		EndTime:   &newTime,
		Metadata:  map[string]string{},
		LLMCalls:  []LLMCallLog{},
		ToolCalls: []ToolCallLog{},
		HTTPCalls: []HTTPLog{},
	}
	
	err = storage.SaveSession(oldSession)
	require.NoError(t, err)
	err = storage.SaveSession(newSession)
	require.NoError(t, err)
	
	// Delete sessions older than 30 days
	err = storage.DeleteOldSessions(30)
	require.NoError(t, err)
	
	// Verify old session is deleted
	_, err = storage.LoadSession("old-session")
	assert.Error(t, err)
	
	// Verify new session still exists
	loaded, err := storage.LoadSession("new-session")
	require.NoError(t, err)
	assert.Equal(t, "new-session", loaded.ID)
	
	// Verify JSON file is also deleted
	oldPath := filepath.Join(tempDir, "old-session.json")
	_, err = os.Stat(oldPath)
	assert.True(t, os.IsNotExist(err))
}

func TestCalculateMetadata(t *testing.T) {
	storage := &Storage{}
	
	now := time.Now()
	endTime := now.Add(1 * time.Hour)
	session := &SessionLog{
		ID:        "test-session",
		StartTime: now,
		EndTime:   &endTime,
		LLMCalls: []LLMCallLog{
			{
				TokensUsed: &TokenUsage{Total: 100},
				Cost:       ptrFloat64(0.01),
			},
			{
				TokensUsed: &TokenUsage{Total: 200},
				Cost:       ptrFloat64(0.02),
				Error:      "test error",
			},
		},
		ToolCalls: []ToolCallLog{
			{Error: ""},
			{Error: "tool error"},
		},
		HTTPCalls: []HTTPLog{
			{Error: ""},
			{Error: ""},
		},
	}
	
	meta := storage.calculateMetadata(session)
	
	assert.Equal(t, "test-session", meta.ID)
	assert.Equal(t, 2, meta.LLMCallCount)
	assert.Equal(t, 2, meta.ToolCallCount)
	assert.Equal(t, 2, meta.HTTPCallCount)
	assert.Equal(t, 300, meta.TotalTokens)
	assert.Equal(t, 0.03, meta.TotalCost)
	assert.True(t, meta.HasError)
}

// Helper function to create float64 pointer
func ptrFloat64(f float64) *float64 {
	return &f
}