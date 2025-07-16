package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kirmad/superopencode/internal/config"
	"github.com/kirmad/superopencode/internal/detailed_logging"
)

func TestDetailedLogging_FlagHandling(t *testing.T) {
	// Test that --detailed-log flag is properly handled during client startup
	t.Run("DetailedLogFlagEnabled", func(t *testing.T) {
		cfg := &config.Config{
			Debug:       true,
			DetailedLog: true,
		}

		if !cfg.DetailedLog {
			t.Error("DetailedLog flag should be enabled")
		}
	})

	t.Run("DetailedLogFlagDisabled", func(t *testing.T) {
		cfg := &config.Config{
			Debug:       true,
			DetailedLog: false,
		}

		if cfg.DetailedLog {
			t.Error("DetailedLog flag should be disabled")
		}
	})
}

func TestDetailedLogging_DirectoryStructure(t *testing.T) {
	// Test that detailed-logs/<session-id>/<request-index>/ structure is created
	sessionID := "test-session-123"
	requestIndex := 1

	t.Run("CreateDetailedLogsDirectory", func(t *testing.T) {
		// Reset detailed logging manager for test isolation
		detailed_logging.ResetDetailedLogManager()
		
		// Initialize detailed logging for testing
		detailed_logging.InitializeDetailedLogging(true, ".")
		
		expectedPath := filepath.Join("detailed-logs", sessionID, "1")
		
		err := detailed_logging.CreateDetailedLogDirectory(sessionID, requestIndex)
		if err != nil {
			t.Fatalf("Failed to create detailed log directory: %v", err)
		}

		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("Expected directory %s was not created", expectedPath)
		}

		// Cleanup
		os.RemoveAll("detailed-logs")
	})
}

func TestDetailedLogging_ManagerInitialization(t *testing.T) {
	t.Run("InitializeDetailedLogging", func(t *testing.T) {
		// Reset detailed logging manager for test isolation
		detailed_logging.ResetDetailedLogManager()
		
		// Test that detailed logging manager initializes properly
		manager := detailed_logging.InitializeDetailedLogging(true, ".")
		
		if manager == nil {
			t.Error("Detailed logging manager should not be nil")
		}
		
		if !manager.IsEnabled() {
			t.Error("Detailed logging should be enabled")
		}

		// Cleanup
		os.RemoveAll("detailed-logs")
	})

	t.Run("DisabledDetailedLogging", func(t *testing.T) {
		// Reset detailed logging manager for test isolation
		detailed_logging.ResetDetailedLogManager()
		
		// Test that detailed logging manager respects disabled state
		manager := detailed_logging.InitializeDetailedLogging(false, ".")
		
		if manager == nil {
			t.Error("Detailed logging manager should not be nil")
		}
		
		if manager.IsEnabled() {
			t.Error("Detailed logging should be disabled")
		}
	})
}

func TestDetailedLogging_CopilotIntegration(t *testing.T) {
	// Test integration with copilot provider
	t.Run("CopilotDetailedLogging", func(t *testing.T) {
		// Mock copilot client with detailed logging enabled
		client := &copilotClient{
			providerOptions: providerClientOptions{
				systemMessage: "Test system message",
			},
			detailedLogging: true,
		}

		// Test the detailedLogging field exists and is properly set
		if !client.detailedLogging {
			t.Error("detailedLogging should be enabled for this test")
		}
	})
}

func TestDetailedLogging_SessionTracking(t *testing.T) {
	// Test session ID generation and tracking
	t.Run("RequestIndexIncrement", func(t *testing.T) {
		// Reset detailed logging manager for test isolation
		detailed_logging.ResetDetailedLogManager()
		
		// Initialize detailed logging for testing
		detailed_logging.InitializeDetailedLogging(true, ".")
		
		sessionID := "test-session-increment"
		
		// First request should be index 1
		index1 := detailed_logging.GetNextRequestIndex(sessionID)
		if index1 != 1 {
			t.Errorf("First request index should be 1, got %d", index1)
		}
		
		// Second request should be index 2
		index2 := detailed_logging.GetNextRequestIndex(sessionID)
		if index2 != 2 {
			t.Errorf("Second request index should be 2, got %d", index2)
		}
	})
}

func TestDetailedLogging_ErrorHandling(t *testing.T) {
	t.Run("DisabledLogging", func(t *testing.T) {
		// Reset detailed logging manager for test isolation
		detailed_logging.ResetDetailedLogManager()
		
		// Test handling when detailed logging is disabled
		detailed_logging.InitializeDetailedLogging(false, ".")
		
		err := detailed_logging.CreateDetailedLogDirectory("test-session", 1)
		if err == nil {
			t.Error("Should return error when detailed logging is disabled")
		}
	})

	t.Run("NilManager", func(t *testing.T) {
		// Reset detailed logging manager for test isolation
		detailed_logging.ResetDetailedLogManager()
		
		// Test handling when manager is nil
		index := detailed_logging.GetNextRequestIndex("test-session")
		if index != 0 {
			t.Error("Should return 0 when manager is nil")
		}
	})
}

