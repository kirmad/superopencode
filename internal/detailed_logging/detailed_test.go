package detailed_logging

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDetailedLogger(t *testing.T) {
	t.Run("disabled logger", func(t *testing.T) {
		logger, err := NewDetailedLogger(false)
		require.NoError(t, err)
		assert.NotNil(t, logger)
		assert.False(t, logger.IsEnabled())
	})
	
	t.Run("enabled logger", func(t *testing.T) {
		// Save original transport
		originalTransport := http.DefaultTransport
		defer func() {
			http.DefaultTransport = originalTransport
		}()
		
		logger, err := NewDetailedLogger(true)
		require.NoError(t, err)
		require.NotNil(t, logger)
		defer logger.Close()
		
		assert.True(t, logger.IsEnabled())
		assert.NotEmpty(t, logger.sessionID)
		assert.NotNil(t, logger.session)
		assert.NotNil(t, logger.storage)
		assert.NotNil(t, logger.toolTracker)
		
		// Verify HTTP interceptor was installed
		_, ok := http.DefaultTransport.(*HTTPInterceptor)
		assert.True(t, ok)
	})
}

func TestDetailedLoggerIsEnabled(t *testing.T) {
	// Test nil logger
	var logger *DetailedLogger
	assert.False(t, logger.IsEnabled())
	
	// Test disabled logger
	logger = &DetailedLogger{enabled: false}
	assert.False(t, logger.IsEnabled())
	
	// Test enabled logger
	logger = &DetailedLogger{enabled: true}
	assert.True(t, logger.IsEnabled())
}

func TestDetailedLoggerSetCommandArgs(t *testing.T) {
	logger := &DetailedLogger{
		enabled: true,
		session: &SessionLog{},
	}
	
	args := []string{"opencode", "--debug", "test"}
	logger.SetCommandArgs(args)
	
	assert.Equal(t, args, logger.session.CommandArgs)
}

func TestDetailedLoggerSetMetadata(t *testing.T) {
	logger := &DetailedLogger{
		enabled: true,
		session: &SessionLog{
			Metadata: make(map[string]string),
		},
	}
	
	logger.SetMetadata("key1", "value1")
	logger.SetMetadata("key2", "value2")
	
	assert.Equal(t, "value1", logger.session.Metadata["key1"])
	assert.Equal(t, "value2", logger.session.Metadata["key2"])
}

func TestDetailedLoggerLogLLMCall(t *testing.T) {
	// Create a temporary storage for testing
	tempDir := t.TempDir()
	storage, err := NewStorage(tempDir)
	require.NoError(t, err)
	defer storage.Close()
	
	logger := &DetailedLogger{
		enabled:   true,
		sessionID: "test-session",
		session: &SessionLog{
			ID:        "test-session",
			StartTime: time.Now(),
			Metadata:  make(map[string]string),
			LLMCalls:  []LLMCallLog{},
			ToolCalls: []ToolCallLog{},
			HTTPCalls: []HTTPLog{},
		},
		storage: storage,
	}
	
	llmCall := &LLMCallLog{
		ID:        "llm-1",
		SessionID: "test-session",
		Provider:  "openai",
		Model:     "gpt-4",
	}
	
	logger.LogLLMCall(llmCall)
	
	// Give async save time to run
	time.Sleep(10 * time.Millisecond)
	
	assert.Len(t, logger.session.LLMCalls, 1)
	assert.Equal(t, "llm-1", logger.session.LLMCalls[0].ID)
}

func TestDetailedLoggerLogToolCall(t *testing.T) {
	// Create a temporary storage for testing
	tempDir := t.TempDir()
	storage, err := NewStorage(tempDir)
	require.NoError(t, err)
	defer storage.Close()
	
	logger := &DetailedLogger{
		enabled:   true,
		sessionID: "test-session",
		session: &SessionLog{
			ID:        "test-session",
			StartTime: time.Now(),
			Metadata:  make(map[string]string),
			LLMCalls:  []LLMCallLog{},
			ToolCalls: []ToolCallLog{},
			HTTPCalls: []HTTPLog{},
		},
		storage: storage,
	}
	
	toolCall := &ToolCallLog{
		ID:        "tool-1",
		SessionID: "test-session",
		Name:      "test_tool",
	}
	
	logger.LogToolCall(toolCall)
	
	// Give async save time to run
	time.Sleep(10 * time.Millisecond)
	
	assert.Len(t, logger.session.ToolCalls, 1)
	assert.Equal(t, "tool-1", logger.session.ToolCalls[0].ID)
}

func TestDetailedLoggerLogHTTP(t *testing.T) {
	// Create a temporary storage for testing
	tempDir := t.TempDir()
	storage, err := NewStorage(tempDir)
	require.NoError(t, err)
	defer storage.Close()
	
	logger := &DetailedLogger{
		enabled:   true,
		sessionID: "test-session",
		session: &SessionLog{
			ID:        "test-session",
			StartTime: time.Now(),
			Metadata:  make(map[string]string),
			LLMCalls:  []LLMCallLog{},
			ToolCalls: []ToolCallLog{},
			HTTPCalls: []HTTPLog{},
		},
		storage: storage,
	}
	
	httpCall := &HTTPLog{
		ID:        "http-1",
		SessionID: "test-session",
		Method:    "GET",
		URL:       "https://example.com",
	}
	
	logger.LogHTTP(httpCall)
	
	// Give async save time to run
	time.Sleep(10 * time.Millisecond)
	
	assert.Len(t, logger.session.HTTPCalls, 1)
	assert.Equal(t, "http-1", logger.session.HTTPCalls[0].ID)
}

func TestDetailedLoggerToolTracking(t *testing.T) {
	logger := &DetailedLogger{
		enabled:     true,
		toolTracker: NewToolTracker("test-session", nil),
	}
	
	// Start a tool call
	toolID := logger.StartToolCall("test_tool", map[string]interface{}{"param": "value"})
	assert.NotEmpty(t, toolID)
	
	// Get current tool call
	currentID := logger.GetCurrentToolCall()
	assert.Equal(t, toolID, currentID)
	
	// End tool call
	logger.EndToolCall(toolID, "output", nil)
	
	// Verify no current tool call
	assert.Empty(t, logger.GetCurrentToolCall())
}

func TestDetailedLoggerLLMCallContext(t *testing.T) {
	logger := &DetailedLogger{
		enabled: true,
	}
	
	// Set and get LLM call
	logger.SetCurrentLLMCall("llm-123")
	assert.Equal(t, "llm-123", logger.GetCurrentLLMCall())
	
	// Test with disabled logger
	disabledLogger := &DetailedLogger{enabled: false}
	disabledLogger.SetCurrentLLMCall("llm-456")
	assert.Empty(t, disabledLogger.GetCurrentLLMCall())
}

func TestDetailedLoggerEndSession(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewStorage(tempDir)
	require.NoError(t, err)
	
	logger := &DetailedLogger{
		enabled:   true,
		sessionID: "test-session",
		session: &SessionLog{
			ID:        "test-session",
			StartTime: time.Now(),
			Metadata:  make(map[string]string),
			LLMCalls:  []LLMCallLog{},
			ToolCalls: []ToolCallLog{},
			HTTPCalls: []HTTPLog{},
		},
		storage: storage,
	}
	
	// End session
	logger.EndSession()
	
	// Verify end time was set
	assert.NotNil(t, logger.session.EndTime)
	
	// Give async operations time to complete
	time.Sleep(50 * time.Millisecond)
	
	// Verify session was saved
	sessionPath := filepath.Join(tempDir, "test-session.json")
	_, err = os.Stat(sessionPath)
	assert.NoError(t, err)
}


func TestDetailedLoggerClose(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewStorage(tempDir)
	require.NoError(t, err)
	
	logger := &DetailedLogger{
		enabled:   true,
		sessionID: "test-session",
		session: &SessionLog{
			ID:        "test-session",
			StartTime: time.Now(),
			Metadata:  make(map[string]string),
			LLMCalls:  []LLMCallLog{},
			ToolCalls: []ToolCallLog{},
			HTTPCalls: []HTTPLog{},
		},
		storage: storage,
	}
	
	// Close logger
	err = logger.Close()
	assert.NoError(t, err)
	
	// Verify session end time was set
	assert.NotNil(t, logger.session.EndTime)
	
	// Test closing disabled logger
	disabledLogger := &DetailedLogger{enabled: false}
	err = disabledLogger.Close()
	assert.NoError(t, err)
}

func TestDetailedLoggerDisabledOperations(t *testing.T) {
	logger := &DetailedLogger{enabled: false}
	
	// All operations should be no-ops when disabled
	logger.SetCommandArgs([]string{"test"})
	logger.SetMetadata("key", "value")
	logger.LogLLMCall(&LLMCallLog{})
	logger.LogToolCall(&ToolCallLog{})
	logger.LogHTTP(&HTTPLog{})
	
	toolID := logger.StartToolCall("tool", nil)
	assert.Empty(t, toolID)
	
	logger.EndToolCall("any-id", nil, nil)
	assert.Empty(t, logger.GetCurrentToolCall())
	
	logger.SetCurrentLLMCall("llm-id")
	assert.Empty(t, logger.GetCurrentLLMCall())
	
	assert.Nil(t, logger.GetToolTracker())
	
	logger.EndSession()
}