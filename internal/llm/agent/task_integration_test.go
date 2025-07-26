package agent

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kirmad/superopencode/internal/llm/tools"
)

func TestTaskToolParameterParsing(t *testing.T) {
	// Create task tool instance for testing
	_ = NewTaskTool(nil, nil, nil, nil, nil)
	
	// Test valid parameters
	validParams := TaskParams{
		Description:  "Test task",
		Prompt:       "This is a test prompt",
		SubagentType: "general-purpose",
	}
	
	paramBytes, err := json.Marshal(validParams)
	if err != nil {
		t.Fatalf("Failed to marshal valid parameters: %v", err)
	}
	
	// Verify the JSON structure matches expected format
	var parsed TaskParams
	if err := json.Unmarshal(paramBytes, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal parameters: %v", err)
	}
	
	if parsed.Description != validParams.Description {
		t.Errorf("Expected description %s, got %s", validParams.Description, parsed.Description)
	}
	if parsed.Prompt != validParams.Prompt {
		t.Errorf("Expected prompt %s, got %s", validParams.Prompt, parsed.Prompt)
	}
	if parsed.SubagentType != validParams.SubagentType {
		t.Errorf("Expected subagent_type %s, got %s", validParams.SubagentType, parsed.SubagentType)
	}
}

func TestTaskToolParameterValidation(t *testing.T) {
	taskTool := NewTaskTool(nil, nil, nil, nil, nil)
	
	testCases := []struct {
		name        string
		params      TaskParams
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid parameters",
			params: TaskParams{
				Description:  "Valid task",
				Prompt:       "Valid prompt",
				SubagentType: "general-purpose",
			},
			expectError: false,
		},
		{
			name: "missing description",
			params: TaskParams{
				Prompt:       "Valid prompt",
				SubagentType: "general-purpose",
			},
			expectError: true,
			errorMsg:    "description is required",
		},
		{
			name: "missing prompt",
			params: TaskParams{
				Description:  "Valid task",
				SubagentType: "general-purpose",
			},
			expectError: true,
			errorMsg:    "prompt is required",
		},
		{
			name: "missing subagent_type",
			params: TaskParams{
				Description: "Valid task",
				Prompt:      "Valid prompt",
			},
			expectError: true,
			errorMsg:    "subagent_type is required",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paramBytes, _ := json.Marshal(tc.params)
			
			toolCall := tools.ToolCall{
				ID:    "test-call-id",
				Name:  "Task",
				Input: string(paramBytes),
			}
			
			// Test parameter validation without context (should fail for missing context)
			ctx := context.Background()
			response, err := taskTool.Run(ctx, toolCall)
			
			if tc.expectError {
				// Should either return error response or actual error
				if err == nil && !response.IsError {
					t.Errorf("Expected error for test case %s, but got success", tc.name)
				}
				if response.IsError && response.Content != tc.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tc.errorMsg, response.Content)
				}
			}
		})
	}
}

func TestTaskToolInfo(t *testing.T) {
	taskTool := NewTaskTool(nil, nil, nil, nil, nil)
	info := taskTool.Info()
	
	// Validate tool info structure
	if info.Name != "Task" {
		t.Errorf("Expected tool name 'Task', got '%s'", info.Name)
	}
	
	if len(info.Description) == 0 {
		t.Error("Tool description should not be empty")
	}
	
	// Check required parameters
	expectedRequired := []string{"description", "prompt", "subagent_type"}
	if len(info.Required) != len(expectedRequired) {
		t.Errorf("Expected %d required parameters, got %d", len(expectedRequired), len(info.Required))
	}
	
	for i, expected := range expectedRequired {
		if i >= len(info.Required) || info.Required[i] != expected {
			t.Errorf("Expected required parameter '%s' at index %d, got '%s'", expected, i, info.Required[i])
		}
	}
	
	// Validate parameters structure
	params, ok := info.Parameters["type"]
	if !ok {
		t.Error("Expected 'type' in parameters")
	}
	
	if params != "object" {
		t.Errorf("Expected type 'object', got %v", params)
	}
	
	// Check properties exist
	properties, ok := info.Parameters["properties"].(map[string]any)
	if !ok {
		t.Error("Expected 'properties' to be a map")
	}
	
	for _, prop := range expectedRequired {
		if _, exists := properties[prop]; !exists {
			t.Errorf("Expected property '%s' to exist in parameters", prop)
		}
	}
}

func TestTaskToolWithMockContext(t *testing.T) {
	taskTool := NewTaskTool(nil, nil, nil, nil, nil)
	
	// Create context with session and message IDs
	ctx := context.WithValue(context.Background(), tools.SessionIDContextKey, "test-session-id")
	ctx = context.WithValue(ctx, tools.MessageIDContextKey, "test-message-id")
	
	validParams := TaskParams{
		Description:  "Test task",
		Prompt:       "Test prompt",
		SubagentType: "general-purpose",
	}
	
	paramBytes, _ := json.Marshal(validParams)
	toolCall := tools.ToolCall{
		ID:    "test-call-id",
		Name:  "Task",
		Input: string(paramBytes),
	}
	
	// This should fail due to nil services, but should pass parameter validation
	// We use defer/recover to catch the expected panic from nil pointer
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic due to nil services")
		}
	}()
	
	_, _ = taskTool.Run(ctx, toolCall)
}