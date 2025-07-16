package tools

import (
	"context"
	"os"
	"testing"
)

func TestTodoContinuationLogic(t *testing.T) {
	// Clean up any existing continuation state before tests
	continuationTracker = &TodoContinuationTracker{
		continuations: make(map[string]int),
		lastCheck:     make(map[string]int64),
	}

	t.Run("HasIncompleteTodos", func(t *testing.T) {
		sessionID := "test-session"
		
		// Initially no todos
		if hasIncompleteTodos(sessionID) {
			t.Error("Should have no incomplete todos initially")
		}
		
		// Add some todos with mixed statuses
		todoStorage.todos[sessionID] = []TodoItem{
			{ID: "1", Content: "Task 1", Status: "completed", Priority: "high"},
			{ID: "2", Content: "Task 2", Status: "pending", Priority: "medium"},
			{ID: "3", Content: "Task 3", Status: "in_progress", Priority: "low"},
		}
		
		if !hasIncompleteTodos(sessionID) {
			t.Error("Should have incomplete todos")
		}
		
		// All completed
		todoStorage.todos[sessionID] = []TodoItem{
			{ID: "1", Content: "Task 1", Status: "completed", Priority: "high"},
			{ID: "2", Content: "Task 2", Status: "completed", Priority: "medium"},
		}
		
		if hasIncompleteTodos(sessionID) {
			t.Error("Should have no incomplete todos when all completed")
		}
	})

	t.Run("GetNextPriorityTodo", func(t *testing.T) {
		sessionID := "test-session"
		
		// Setup todos with different priorities
		todoStorage.todos[sessionID] = []TodoItem{
			{ID: "1", Content: "Low priority", Status: "pending", Priority: "low"},
			{ID: "2", Content: "High priority", Status: "pending", Priority: "high"},
			{ID: "3", Content: "Medium priority", Status: "pending", Priority: "medium"},
			{ID: "4", Content: "Completed high", Status: "completed", Priority: "high"},
		}
		
		nextTodo := getNextPriorityTodo(sessionID)
		if nextTodo == nil {
			t.Fatal("Should find next priority todo")
		}
		
		if nextTodo.Priority != "high" || nextTodo.Content != "High priority" {
			t.Errorf("Should prioritize high priority incomplete todo, got: %s - %s", nextTodo.Priority, nextTodo.Content)
		}
		
		// Mark high priority as completed
		todoStorage.todos[sessionID][1].Status = "completed"
		
		nextTodo = getNextPriorityTodo(sessionID)
		if nextTodo.Priority != "medium" {
			t.Errorf("Should move to medium priority, got: %s", nextTodo.Priority)
		}
		
		// All completed
		todoStorage.todos[sessionID] = []TodoItem{
			{ID: "1", Content: "Task", Status: "completed", Priority: "high"},
		}
		
		nextTodo = getNextPriorityTodo(sessionID)
		if nextTodo != nil {
			t.Error("Should return nil when no incomplete todos")
		}
	})

	t.Run("ContinuationCountTracking", func(t *testing.T) {
		sessionID := "test-session"
		
		// Initially no continuations
		if exceededMaxContinuations(sessionID) {
			t.Error("Should not exceed max continuations initially")
		}
		
		// Increment a few times
		for i := 1; i <= 5; i++ {
			count := IncrementContinuationCount(sessionID)
			if count != i {
				t.Errorf("Expected count %d, got %d", i, count)
			}
		}
		
		// Should not exceed max yet (assuming default is 10)
		if exceededMaxContinuations(sessionID) {
			t.Error("Should not exceed max continuations yet")
		}
		
		// Reset counter
		ResetContinuationCount(sessionID)
		
		count := IncrementContinuationCount(sessionID)
		if count != 1 {
			t.Errorf("Should reset to 1, got %d", count)
		}
	})

	t.Run("ShouldContinueForTodos", func(t *testing.T) {
		sessionID := "test-session"
		ctx := context.Background()
		
		// Clean state
		ResetContinuationCount(sessionID)
		delete(todoStorage.todos, sessionID)
		
		// Set environment variable to enable feature
		os.Setenv("SUPEROPENCODE_AUTO_COMPLETE_TODOS", "true")
		defer os.Unsetenv("SUPEROPENCODE_AUTO_COMPLETE_TODOS")
		
		// Should not continue if no todos
		if ShouldContinueForTodos(ctx, sessionID, "end_turn") {
			t.Error("Should not continue when no todos")
		}
		
		// Add incomplete todos
		todoStorage.todos[sessionID] = []TodoItem{
			{ID: "1", Content: "Task", Status: "pending", Priority: "high"},
		}
		
		// Should continue for EndTurn with incomplete todos
		if !ShouldContinueForTodos(ctx, sessionID, "end_turn") {
			t.Error("Should continue for EndTurn with incomplete todos")
		}
		
		// Should not continue for other finish reasons
		if ShouldContinueForTodos(ctx, sessionID, "tool_use") {
			t.Error("Should not continue for ToolUse finish reason")
		}
		
		if ShouldContinueForTodos(ctx, sessionID, "canceled") {
			t.Error("Should not continue for Canceled finish reason")
		}
		
		// Test max continuations exceeded
		for i := 0; i < 15; i++ { // Exceed default max of 10
			IncrementContinuationCount(sessionID)
		}
		
		if ShouldContinueForTodos(ctx, sessionID, "end_turn") {
			t.Error("Should not continue when max continuations exceeded")
		}
	})

	t.Run("GenerateTodoContinuationPrompt", func(t *testing.T) {
		sessionID := "test-session"
		
		// Setup todos
		todoStorage.todos[sessionID] = []TodoItem{
			{ID: "1", Content: "High priority task", Status: "pending", Priority: "high"},
			{ID: "2", Content: "Medium priority task", Status: "pending", Priority: "medium"},
		}
		
		promptText := generateTodoContinuationPromptText(sessionID)
		
		if promptText == "" {
			t.Error("Prompt text should not be empty")
		}
		
		// Should mention the high priority task
		if !contains(promptText, "High priority task") {
			t.Error("Prompt should mention the next high priority task")
		}
		
		// Should mention the count
		if !contains(promptText, "2") {
			t.Error("Prompt should mention the number of incomplete tasks")
		}
	})

	t.Run("FeatureDisabled", func(t *testing.T) {
		sessionID := "test-session"
		ctx := context.Background()
		
		// Feature is disabled by default (no env var or config set)
		os.Unsetenv("SUPEROPENCODE_AUTO_COMPLETE_TODOS")
		
		// Add incomplete todos
		todoStorage.todos[sessionID] = []TodoItem{
			{ID: "1", Content: "Task", Status: "pending", Priority: "high"},
		}
		
		// Should not continue when feature is disabled
		if ShouldContinueForTodos(ctx, sessionID, "end_turn") {
			t.Error("Should not continue when feature is disabled")
		}
	})
}

func TestMaxContinuationsLimit(t *testing.T) {
	sessionID := "test-session"
	
	// Clean state
	ResetContinuationCount(sessionID)
	
	// Test default max continuations
	maxContinuations := getMaxContinuations()
	if maxContinuations != 10 {
		t.Errorf("Expected default max continuations to be 10, got %d", maxContinuations)
	}
	
	// Increment to just under limit
	for i := 0; i < maxContinuations-1; i++ {
		IncrementContinuationCount(sessionID)
	}
	
	if exceededMaxContinuations(sessionID) {
		t.Error("Should not exceed max when just under limit")
	}
	
	// One more should exceed
	IncrementContinuationCount(sessionID)
	
	if !exceededMaxContinuations(sessionID) {
		t.Error("Should exceed max after reaching limit")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && 
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}()
}