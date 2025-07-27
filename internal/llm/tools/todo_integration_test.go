package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// TestTodoWorkflow tests a complete TODO workflow scenario
func TestTodoWorkflow(t *testing.T) {
	readTool := NewTodoReadTool()
	writeTool := NewTodoWriteTool()
	ctx := context.WithValue(context.Background(), SessionIDContextKey, "workflow-test")

	// 1. Initially no todos
	readCall := ToolCall{ID: "read1", Name: TodoReadToolName, Input: "{}"}
	response, err := readTool.Run(ctx, readCall)
	if err != nil {
		t.Fatalf("Initial read failed: %v", err)
	}
	if !strings.Contains(response.Content, "No todos found for this session.") {
		t.Errorf("Expected empty todos message, got: %s", response.Content)
	}

	// 2. Create initial todo list
	todos := []TodoItem{
		{ID: "task1", Content: "Implement feature A", Status: "pending", Priority: "high"},
		{ID: "task2", Content: "Write tests for feature A", Status: "pending", Priority: "medium"},
		{ID: "task3", Content: "Update documentation", Status: "pending", Priority: "low"},
	}

	inputBytes, _ := json.Marshal(map[string]interface{}{"todos": todos})
	writeCall := ToolCall{ID: "write1", Name: TodoWriteToolName, Input: string(inputBytes)}
	
	response, err = writeTool.Run(ctx, writeCall)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if response.IsError {
		t.Errorf("Expected success, got error: %s", response.Content)
	}

	// 3. Read todos back
	response, err = readTool.Run(ctx, readCall)
	if err != nil {
		t.Fatalf("Read after write failed: %v", err)
	}

	var returnedTodos []TodoItem
	// Extract JSON part before any validation notes
	content := response.Content
	if validationIndex := strings.Index(content, "\n<validation-"); validationIndex != -1 {
		content = content[:validationIndex]
	}
	err = json.Unmarshal([]byte(content), &returnedTodos)
	if err != nil {
		t.Fatalf("Failed to parse todos: %v", err)
	}

	if len(returnedTodos) != 3 {
		t.Errorf("Expected 3 todos, got %d", len(returnedTodos))
	}

	// 4. Start working on first task
	todos[0].Status = "in_progress"
	inputBytes, _ = json.Marshal(map[string]interface{}{"todos": todos})
	writeCall.Input = string(inputBytes)
	
	response, err = writeTool.Run(ctx, writeCall)
	if err != nil {
		t.Fatalf("Update to in_progress failed: %v", err)
	}

	// 5. Complete first task, start second
	todos[0].Status = "completed"
	todos[1].Status = "in_progress"
	inputBytes, _ = json.Marshal(map[string]interface{}{"todos": todos})
	writeCall.Input = string(inputBytes)
	
	response, err = writeTool.Run(ctx, writeCall)
	if err != nil {
		t.Fatalf("Progress update failed: %v", err)
	}

	// 6. Verify final state
	response, err = readTool.Run(ctx, readCall)
	if err != nil {
		t.Fatalf("Final read failed: %v", err)
	}

	// Extract JSON part before any validation notes
	content = response.Content
	if validationIndex := strings.Index(content, "\n<validation-"); validationIndex != -1 {
		content = content[:validationIndex]
	}
	err = json.Unmarshal([]byte(content), &returnedTodos)
	if err != nil {
		t.Fatalf("Failed to parse final todos: %v", err)
	}

	// Verify states
	if returnedTodos[0].Status != "completed" {
		t.Errorf("Task 1 should be completed, got %s", returnedTodos[0].Status)
	}
	if returnedTodos[1].Status != "in_progress" {
		t.Errorf("Task 2 should be in_progress, got %s", returnedTodos[1].Status)
	}
	if returnedTodos[2].Status != "pending" {
		t.Errorf("Task 3 should be pending, got %s", returnedTodos[2].Status)
	}
}

// TestSessionIsolation verifies todos are isolated per session
func TestSessionIsolation(t *testing.T) {
	readTool := NewTodoReadTool()
	writeTool := NewTodoWriteTool()

	ctx1 := context.WithValue(context.Background(), SessionIDContextKey, "session-1")
	ctx2 := context.WithValue(context.Background(), SessionIDContextKey, "session-2")

	// Create todos in session 1
	todos1 := []TodoItem{
		{ID: "s1-task1", Content: "Session 1 task", Status: "pending", Priority: "high"},
	}
	inputBytes, _ := json.Marshal(map[string]interface{}{"todos": todos1})
	writeCall := ToolCall{ID: "write", Name: TodoWriteToolName, Input: string(inputBytes)}
	
	_, err := writeTool.Run(ctx1, writeCall)
	if err != nil {
		t.Fatalf("Session 1 write failed: %v", err)
	}

	// Create different todos in session 2
	todos2 := []TodoItem{
		{ID: "s2-task1", Content: "Session 2 task", Status: "pending", Priority: "medium"},
	}
	inputBytes, _ = json.Marshal(map[string]interface{}{"todos": todos2})
	writeCall.Input = string(inputBytes)
	
	_, err = writeTool.Run(ctx2, writeCall)
	if err != nil {
		t.Fatalf("Session 2 write failed: %v", err)
	}

	// Read from session 1
	readCall := ToolCall{ID: "read", Name: TodoReadToolName, Input: "{}"}
	response, err := readTool.Run(ctx1, readCall)
	if err != nil {
		t.Fatalf("Session 1 read failed: %v", err)
	}

	var session1Todos []TodoItem
	err = json.Unmarshal([]byte(response.Content), &session1Todos)
	if err != nil {
		t.Fatalf("Failed to parse session 1 todos: %v", err)
	}

	if len(session1Todos) != 1 || session1Todos[0].Content != "Session 1 task" {
		t.Error("Session 1 should only see its own todos")
	}

	// Read from session 2
	response, err = readTool.Run(ctx2, readCall)
	if err != nil {
		t.Fatalf("Session 2 read failed: %v", err)
	}

	var session2Todos []TodoItem
	err = json.Unmarshal([]byte(response.Content), &session2Todos)
	if err != nil {
		t.Fatalf("Failed to parse session 2 todos: %v", err)
	}

	if len(session2Todos) != 1 || session2Todos[0].Content != "Session 2 task" {
		t.Error("Session 2 should only see its own todos")
	}
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	writeTool := NewTodoWriteTool()
	ctx := context.WithValue(context.Background(), SessionIDContextKey, "edge-test")

	// Test empty content
	todos := []TodoItem{
		{ID: "task1", Content: "", Status: "pending", Priority: "high"},
	}
	inputBytes, _ := json.Marshal(map[string]interface{}{"todos": todos})
	writeCall := ToolCall{ID: "write", Name: TodoWriteToolName, Input: string(inputBytes)}
	
	response, err := writeTool.Run(ctx, writeCall)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !response.IsError {
		t.Error("Expected error for empty content")
	}

	// Test empty ID
	todos[0].Content = "Valid content"
	todos[0].ID = ""
	inputBytes, _ = json.Marshal(map[string]interface{}{"todos": todos})
	writeCall.Input = string(inputBytes)
	
	response, err = writeTool.Run(ctx, writeCall)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !response.IsError {
		t.Error("Expected error for empty ID")
	}

	// Test invalid JSON
	writeCall.Input = "invalid json"
	response, err = writeTool.Run(ctx, writeCall)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !response.IsError {
		t.Error("Expected error for invalid JSON")
	}
}