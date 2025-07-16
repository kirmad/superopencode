package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func TestTodoReminder(t *testing.T) {
	sessionID := "reminder-test"
	
	// Initially no todos - should show reminder
	if !ShouldShowTodoReminder(sessionID) {
		t.Error("Should show reminder when no todos exist")
	}
	
	reminder := GetTodoReminderForSession(sessionID)
	if reminder == "" {
		t.Error("Should return reminder message when no todos exist")
	}
	
	expectedMessage := GetTodoReminderMessage()
	if reminder != expectedMessage {
		t.Error("Reminder message should match expected format")
	}
	
	// Add a todo - should not show reminder
	writeTool := NewTodoWriteTool()
	ctx := context.WithValue(context.Background(), SessionIDContextKey, sessionID)
	
	todos := []TodoItem{
		{ID: "task1", Content: "Test task", Status: "pending", Priority: "high"},
	}
	
	inputBytes, _ := json.Marshal(map[string]interface{}{"todos": todos})
	writeCall := ToolCall{ID: "write", Name: TodoWriteToolName, Input: string(inputBytes)}
	
	_, err := writeTool.Run(ctx, writeCall)
	if err != nil {
		t.Fatalf("Failed to write todos: %v", err)
	}
	
	// Now should not show reminder
	if ShouldShowTodoReminder(sessionID) {
		t.Error("Should not show reminder when todos exist")
	}
	
	reminder = GetTodoReminderForSession(sessionID)
	if reminder != "" {
		t.Error("Should not return reminder message when todos exist")
	}
}

func TestTodoReminderEmptySession(t *testing.T) {
	// Empty session ID should not show reminder
	if ShouldShowTodoReminder("") {
		t.Error("Should not show reminder for empty session ID")
	}
	
	reminder := GetTodoReminderForSession("")
	if reminder != "" {
		t.Error("Should not return reminder for empty session ID")
	}
}