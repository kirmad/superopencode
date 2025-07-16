package tools

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/kirmad/superopencode/internal/config"
	"github.com/kirmad/superopencode/internal/message"
)

// TodoContinuationTracker manages continuation attempts per session
type TodoContinuationTracker struct {
	mu            sync.RWMutex
	continuations map[string]int   // sessionID -> count
	lastCheck     map[string]int64 // sessionID -> timestamp
}

var continuationTracker = &TodoContinuationTracker{
	continuations: make(map[string]int),
	lastCheck:     make(map[string]int64),
}

// hasIncompleteTodos checks if session has pending or in-progress todos
func hasIncompleteTodos(sessionID string) bool {
	if sessionID == "" {
		return false
	}

	todoStorage.mu.RLock()
	defer todoStorage.mu.RUnlock()

	todos := todoStorage.todos[sessionID]
	for _, todo := range todos {
		// Safety check: ensure todo has valid status
		if todo.Status == "pending" || todo.Status == "in_progress" {
			// Additional safety: ensure todo has content
			if todo.Content != "" {
				return true
			}
		}
	}
	return false
}

// getNextPriorityTodo returns the highest priority incomplete todo
func getNextPriorityTodo(sessionID string) *TodoItem {
	if sessionID == "" {
		return nil
	}

	todoStorage.mu.RLock()
	defer todoStorage.mu.RUnlock()

	todos := todoStorage.todos[sessionID]

	// Priority order: high -> medium -> low
	priorities := []string{"high", "medium", "low"}
	for _, priority := range priorities {
		for _, todo := range todos {
			if (todo.Status == "pending" || todo.Status == "in_progress") &&
				todo.Priority == priority && todo.Content != "" {
				// Return a copy to avoid data races
				todoCopy := todo
				return &todoCopy
			}
		}
	}
	return nil
}

// getIncompleteTodos returns all incomplete todos for session
func getIncompleteTodos(sessionID string) []TodoItem {
	todoStorage.mu.RLock()
	defer todoStorage.mu.RUnlock()

	var incomplete []TodoItem
	todos := todoStorage.todos[sessionID]
	for _, todo := range todos {
		if todo.Status == "pending" || todo.Status == "in_progress" {
			incomplete = append(incomplete, todo)
		}
	}
	return incomplete
}

// generateTodoContinuationPromptText creates a continuation prompt text
func generateTodoContinuationPromptText(sessionID string) string {
	if sessionID == "" {
		return "Please check your todo list and continue working on any remaining incomplete tasks."
	}

	incompleteTodos := getIncompleteTodos(sessionID)
	nextTodo := getNextPriorityTodo(sessionID)

	var promptText string
	if nextTodo != nil {
		// Sanitize todo content to prevent injection
		sanitizedContent := sanitizeTodoContent(nextTodo.Content)
		promptText = fmt.Sprintf(
			"You have %d incomplete tasks remaining. Please continue working on the next high-priority task: '%s'. "+
				"Continue until all todos are completed.",
			len(incompleteTodos), sanitizedContent)
	} else {
		promptText = "Please check your todo list and continue working on any remaining incomplete tasks."
	}

	return promptText
}

// sanitizeTodoContent removes any potentially harmful content from todo text
func sanitizeTodoContent(content string) string {
	// Basic sanitization - remove excessive line breaks and limit length
	if len(content) > 200 {
		content = content[:200] + "..."
	}
	return content
}

// CreateTodoContinuationMessage creates a message for todo continuation
func CreateTodoContinuationMessage(sessionID string) message.Message {
	promptText := generateTodoContinuationPromptText(sessionID)
	
	return message.Message{
		Role:  message.User,
		Parts: []message.ContentPart{message.TextContent{Text: promptText}},
	}
}

// IncrementContinuationCount tracks continuation attempts
func IncrementContinuationCount(sessionID string) int {
	if sessionID == "" {
		return 0
	}

	continuationTracker.mu.Lock()
	defer continuationTracker.mu.Unlock()

	// Safety check: prevent overflow
	if continuationTracker.continuations[sessionID] >= 1000 {
		return continuationTracker.continuations[sessionID]
	}

	continuationTracker.continuations[sessionID]++
	continuationTracker.lastCheck[sessionID] = time.Now().Unix()

	return continuationTracker.continuations[sessionID]
}

// exceededMaxContinuations checks if continuation limit reached
func exceededMaxContinuations(sessionID string) bool {
	continuationTracker.mu.RLock()
	defer continuationTracker.mu.RUnlock()

	count := continuationTracker.continuations[sessionID]
	return count >= getMaxContinuations()
}

// ResetContinuationCount resets counter when new user message arrives
func ResetContinuationCount(sessionID string) {
	continuationTracker.mu.Lock()
	defer continuationTracker.mu.Unlock()

	delete(continuationTracker.continuations, sessionID)
	delete(continuationTracker.lastCheck, sessionID)
}

// ShouldContinueForTodos determines if execution should continue
func ShouldContinueForTodos(ctx context.Context, sessionID string, finishReason string) bool {
	// Only continue if:
	// 1. Todo-driven execution is enabled
	// 2. Agent finished normally (not cancelled/error)
	// 3. There are incomplete todos
	// 4. Haven't exceeded max continuations
	// 5. Session ID is valid

	// Safety check: session ID must be provided
	if sessionID == "" {
		return false
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return false
	default:
		// Continue with checks
	}

	if !isTodoDrivenExecutionEnabled() {
		return false
	}

	if finishReason != "end_turn" {
		return false
	}

	return hasIncompleteTodos(sessionID) && !exceededMaxContinuations(sessionID)
}

// isTodoDrivenExecutionEnabled checks if feature is enabled
func isTodoDrivenExecutionEnabled() bool {
	cfg := config.Get()
	if cfg == nil {
		// In test environment, check environment variable
		return os.Getenv("SUPEROPENCODE_AUTO_COMPLETE_TODOS") == "true"
	}
	// Default to coder agent since it's most likely to use this feature
	return cfg.IsTodoDrivenExecutionEnabled(config.AgentCoder)
}

// getMaxContinuations returns the maximum allowed continuations
func getMaxContinuations() int {
	cfg := config.Get()
	if cfg == nil {
		// In test environment, check environment variable
		if envValue := os.Getenv("SUPEROPENCODE_MAX_TODO_CONTINUATIONS"); envValue != "" {
			if value := parseIntWithDefault(envValue, 10); value > 0 {
				return value
			}
		}
		return 10
	}
	// Default to coder agent since it's most likely to use this feature
	return cfg.GetMaxTodoContinuations(config.AgentCoder)
}

// parseIntWithDefault parses string to int with default fallback
func parseIntWithDefault(s string, defaultValue int) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return defaultValue
}