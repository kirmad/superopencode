package prompt

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kirmad/superopencode/internal/llm/tools"
)

func TestGetTodoReminder_EmptyTodos(t *testing.T) {
	sessionID := "reminder-test-empty"
	
	// Test with empty todos - should return reminder
	reminder := getTodoReminder(sessionID)
	
	if reminder == "" {
		t.Error("Should return reminder when todos are empty")
	}
	
	if reminder != tools.GetTodoReminderMessage() {
		t.Error("Should return the correct reminder message")
	}
}

func TestGetTodoReminder_WithTodos(t *testing.T) {
	sessionID := "reminder-test-with-todos"
	
	// Add a todo first
	writeTool := tools.NewTodoWriteTool()
	ctx := context.WithValue(context.Background(), tools.SessionIDContextKey, sessionID)
	
	todos := []tools.TodoItem{
		{ID: "task1", Content: "Test task", Status: "pending", Priority: "high"},
	}
	
	inputBytes, _ := json.Marshal(map[string]interface{}{"todos": todos})
	writeCall := tools.ToolCall{ID: "write", Name: "TodoWrite", Input: string(inputBytes)}
	
	_, err := writeTool.Run(ctx, writeCall)
	if err != nil {
		t.Fatalf("Failed to write todos: %v", err)
	}
	
	// Test with existing todos - should NOT return reminder
	reminder := getTodoReminder(sessionID)
	
	if reminder != "" {
		t.Error("Should not return reminder when todos exist")
	}
}

func TestGetTodoReminder_EmptySessionID(t *testing.T) {
	// Test with empty session ID - should not return reminder
	reminder := getTodoReminder("")
	
	if reminder != "" {
		t.Error("Should not return reminder for empty session ID")
	}
}