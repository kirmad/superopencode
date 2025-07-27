package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestTodoReadTool_Info(t *testing.T) {
	tool := NewTodoReadTool()
	info := tool.Info()

	if info.Name != TodoReadToolName {
		t.Errorf("Expected name '%s', got '%s'", TodoReadToolName, info.Name)
	}

	if len(info.Required) != 0 {
		t.Errorf("Expected no required parameters, got %v", info.Required)
	}

	params := info.Parameters
	if params["type"] != "object" {
		t.Errorf("Expected type 'object', got '%v'", params["type"])
	}
}

func TestTodoWriteTool_Info(t *testing.T) {
	tool := NewTodoWriteTool()
	info := tool.Info()

	if info.Name != TodoWriteToolName {
		t.Errorf("Expected name '%s', got '%s'", TodoWriteToolName, info.Name)
	}

	if len(info.Required) != 1 || info.Required[0] != "todos" {
		t.Errorf("Expected required parameter 'todos', got %v", info.Required)
	}
}

func TestTodoReadTool_EmptySession(t *testing.T) {
	tool := NewTodoReadTool()
	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")

	call := ToolCall{
		ID:    "test",
		Name:  TodoReadToolName,
		Input: "{}",
	}

	response, err := tool.Run(ctx, call)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.IsError {
		t.Errorf("Expected success, got error: %s", response.Content)
	}

	if !strings.Contains(response.Content, "No todos found for this session.") {
		t.Errorf("Expected empty message, got: %s", response.Content)
	}
}

func TestTodoWriteTool_ValidTodos(t *testing.T) {
	tool := NewTodoWriteTool()
	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")

	todos := []TodoItem{
		{
			ID:       "test-1",
			Content:  "Test task",
			Status:   "pending",
			Priority: "high",
		},
	}

	inputBytes, _ := json.Marshal(map[string]interface{}{
		"todos": todos,
	})

	call := ToolCall{
		ID:    "test",
		Name:  TodoWriteToolName,
		Input: string(inputBytes),
	}

	response, err := tool.Run(ctx, call)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.IsError {
		t.Errorf("Expected success, got error: %s", response.Content)
	}
}

func TestTodoWriteTool_MultipleInProgress(t *testing.T) {
	tool := NewTodoWriteTool()
	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")

	todos := []TodoItem{
		{
			ID:       "test-1",
			Content:  "Test task 1",
			Status:   "in_progress",
			Priority: "high",
		},
		{
			ID:       "test-2",
			Content:  "Test task 2",
			Status:   "in_progress",
			Priority: "medium",
		},
	}

	inputBytes, _ := json.Marshal(map[string]interface{}{
		"todos": todos,
	})

	call := ToolCall{
		ID:    "test",
		Name:  TodoWriteToolName,
		Input: string(inputBytes),
	}

	response, err := tool.Run(ctx, call)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !response.IsError {
		t.Error("Expected error for multiple in_progress tasks")
	}

	if response.Content != "Only one task can be in_progress at a time" {
		t.Errorf("Expected specific error message, got: %s", response.Content)
	}
}

