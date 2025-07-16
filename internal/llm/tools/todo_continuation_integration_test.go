package tools

import (
	"context"
	"os"
	"testing"
)

// TestTodoContinuationIntegration tests the full integration of todo-driven execution
func TestTodoContinuationIntegration(t *testing.T) {
	// Set up test environment
	sessionID := "integration-test-session"
	ctx := context.Background()
	
	// Clean state
	ResetContinuationCount(sessionID)
	delete(todoStorage.todos, sessionID)
	
	// Enable todo-driven execution
	os.Setenv("SUPEROPENCODE_AUTO_COMPLETE_TODOS", "true")
	defer os.Unsetenv("SUPEROPENCODE_AUTO_COMPLETE_TODOS")
	
	t.Run("FullWorkflow", func(t *testing.T) {
		// Step 1: Create a todo list with multiple tasks
		todos := []TodoItem{
			{ID: "task1", Content: "Implement feature X", Status: "pending", Priority: "high"},
			{ID: "task2", Content: "Write tests for feature X", Status: "pending", Priority: "medium"},
			{ID: "task3", Content: "Update documentation", Status: "pending", Priority: "low"},
		}
		
		todoStorage.mu.Lock()
		todoStorage.todos[sessionID] = todos
		todoStorage.mu.Unlock()
		
		// Step 2: Check that continuation is needed
		if !ShouldContinueForTodos(ctx, sessionID, "end_turn") {
			t.Error("Should continue when there are incomplete todos")
		}
		
		// Step 3: Generate continuation prompt text directly
		promptText := generateTodoContinuationPromptText(sessionID)
		
		if promptText == "" {
			t.Error("Prompt text should not be empty")
		}
		
		if !contains(promptText, "Implement feature X") {
			t.Error("Prompt should mention the high priority task")
		}
		
		if !contains(promptText, "3") {
			t.Error("Prompt should mention the number of incomplete tasks")
		}
		
		// Step 4: Track continuation count
		count := IncrementContinuationCount(sessionID)
		if count != 1 {
			t.Errorf("Expected count 1, got %d", count)
		}
		
		// Step 5: Simulate completing the first task
		todoStorage.mu.Lock()
		todoStorage.todos[sessionID][0].Status = "completed"
		todoStorage.mu.Unlock()
		
		// Should still continue with remaining tasks
		if !ShouldContinueForTodos(ctx, sessionID, "end_turn") {
			t.Error("Should continue when tasks remain")
		}
		
		// Step 6: Complete second task
		todoStorage.mu.Lock()
		todoStorage.todos[sessionID][1].Status = "completed"
		todoStorage.mu.Unlock()
		
		// Should still continue with last task
		if !ShouldContinueForTodos(ctx, sessionID, "end_turn") {
			t.Error("Should continue when last task remains")
		}
		
		// Step 7: Complete all tasks
		todoStorage.mu.Lock()
		todoStorage.todos[sessionID][2].Status = "completed"
		todoStorage.mu.Unlock()
		
		// Should stop continuing when all tasks are complete
		if ShouldContinueForTodos(ctx, sessionID, "end_turn") {
			t.Error("Should not continue when all tasks are completed")
		}
	})
	
	t.Run("MaxContinuationsExceeded", func(t *testing.T) {
		// Clean state
		ResetContinuationCount(sessionID)
		
		// Create incomplete todos
		todoStorage.mu.Lock()
		todoStorage.todos[sessionID] = []TodoItem{
			{ID: "task1", Content: "Task 1", Status: "pending", Priority: "high"},
		}
		todoStorage.mu.Unlock()
		
		// Exceed max continuations
		for i := 0; i < 15; i++ {
			IncrementContinuationCount(sessionID)
		}
		
		// Should not continue when max exceeded
		if ShouldContinueForTodos(ctx, sessionID, "end_turn") {
			t.Error("Should not continue when max continuations exceeded")
		}
	})
	
	t.Run("CancelledContext", func(t *testing.T) {
		// Create cancelled context
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		
		// Create incomplete todos
		todoStorage.mu.Lock()
		todoStorage.todos[sessionID] = []TodoItem{
			{ID: "task1", Content: "Task 1", Status: "pending", Priority: "high"},
		}
		todoStorage.mu.Unlock()
		
		ResetContinuationCount(sessionID)
		
		// Should not continue with cancelled context
		if ShouldContinueForTodos(cancelledCtx, sessionID, "end_turn") {
			t.Error("Should not continue with cancelled context")
		}
	})
	
	t.Run("EdgeCases", func(t *testing.T) {
		// Test empty session ID
		if ShouldContinueForTodos(ctx, "", "end_turn") {
			t.Error("Should not continue with empty session ID")
		}
		
		// Test invalid finish reason
		ResetContinuationCount(sessionID)
		todoStorage.mu.Lock()
		todoStorage.todos[sessionID] = []TodoItem{
			{ID: "task1", Content: "Task 1", Status: "pending", Priority: "high"},
		}
		todoStorage.mu.Unlock()
		
		if ShouldContinueForTodos(ctx, sessionID, "tool_use") {
			t.Error("Should not continue for non-end_turn finish reason")
		}
		
		// Test todos with empty content
		todoStorage.mu.Lock()
		todoStorage.todos[sessionID] = []TodoItem{
			{ID: "task1", Content: "", Status: "pending", Priority: "high"},
		}
		todoStorage.mu.Unlock()
		
		if ShouldContinueForTodos(ctx, sessionID, "end_turn") {
			t.Error("Should not continue for todos with empty content")
		}
	})
	
	t.Run("MessageSanitization", func(t *testing.T) {
		// Test very long todo content
		longContent := "This is a very long todo content that exceeds the normal length limit. " +
			"It contains a lot of text that should be truncated to prevent excessively long prompts " +
			"from being generated. This helps maintain reasonable token usage and ensures the system remains efficient."
			
		todoStorage.mu.Lock()
		todoStorage.todos[sessionID] = []TodoItem{
			{ID: "task1", Content: longContent, Status: "pending", Priority: "high"},
		}
		todoStorage.mu.Unlock()
		
		promptText := generateTodoContinuationPromptText(sessionID)
		
		// Should be truncated
		if contains(promptText, longContent) {
			t.Error("Long content should be truncated")
		}
		
		if !contains(promptText, "...") {
			t.Error("Truncated content should contain ellipsis")
		}
	})
}

// TestConfigurationIntegration tests configuration-based todo continuation
func TestConfigurationIntegration(t *testing.T) {
	sessionID := "config-test-session"
	ctx := context.Background()
	
	// Clean state
	ResetContinuationCount(sessionID)
	delete(todoStorage.todos, sessionID)
	
	t.Run("EnvironmentVariableOverride", func(t *testing.T) {
		// Environment variable should override config
		os.Setenv("SUPEROPENCODE_AUTO_COMPLETE_TODOS", "true")
		defer os.Unsetenv("SUPEROPENCODE_AUTO_COMPLETE_TODOS")
		
		// Create todos
		todoStorage.mu.Lock()
		todoStorage.todos[sessionID] = []TodoItem{
			{ID: "task1", Content: "Task 1", Status: "pending", Priority: "high"},
		}
		todoStorage.mu.Unlock()
		
		if !ShouldContinueForTodos(ctx, sessionID, "end_turn") {
			t.Error("Environment variable should enable todo continuation")
		}
		
		// Disable via environment
		os.Setenv("SUPEROPENCODE_AUTO_COMPLETE_TODOS", "false")
		defer os.Unsetenv("SUPEROPENCODE_AUTO_COMPLETE_TODOS")
		
		if ShouldContinueForTodos(ctx, sessionID, "end_turn") {
			t.Error("Environment variable should disable todo continuation")
		}
	})
	
	t.Run("MaxContinuationsFromEnvironment", func(t *testing.T) {
		// Set custom max via environment
		os.Setenv("SUPEROPENCODE_MAX_TODO_CONTINUATIONS", "3")
		defer os.Unsetenv("SUPEROPENCODE_MAX_TODO_CONTINUATIONS")
		
		ResetContinuationCount(sessionID)
		
		// Should use environment value
		maxContinuations := getMaxContinuations()
		if maxContinuations != 3 {
			t.Errorf("Expected max continuations 3 from environment, got %d", maxContinuations)
		}
		
		// Test exceeding the custom limit
		for i := 0; i < 5; i++ {
			IncrementContinuationCount(sessionID)
		}
		
		if !exceededMaxContinuations(sessionID) {
			t.Error("Should exceed custom max continuations limit")
		}
	})
}