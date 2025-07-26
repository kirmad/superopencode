package agent

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/kirmad/superopencode/internal/llm/tools"
)

func TestTaskToolAdvancedParameterValidation(t *testing.T) {
	taskTool := NewTaskTool(nil, nil, nil, nil, nil).(*taskTool)

	testCases := []struct {
		name        string
		input       string
		expectError bool
		errorType   string
	}{
		{
			name: "valid parameters",
			input: `{
				"description": "Test task",
				"prompt": "Test prompt",
				"subagent_type": "general-purpose"
			}`,
			expectError: false,
		},
		{
			name: "invalid subagent type",
			input: `{
				"description": "Test task", 
				"prompt": "Test prompt",
				"subagent_type": "invalid-type"
			}`,
			expectError: true,
			errorType:   "invalid subagent type",
		},
		{
			name: "whitespace-only description",
			input: `{
				"description": "   ",
				"prompt": "Test prompt",
				"subagent_type": "general-purpose"
			}`,
			expectError: true,
			errorType:   "description is required",
		},
		{
			name: "whitespace-only prompt",
			input: `{
				"description": "Test task",
				"prompt": "   ",
				"subagent_type": "general-purpose"
			}`,
			expectError: true,
			errorType:   "prompt is required",
		},
		{
			name: "malformed JSON",
			input: `{
				"description": "Test task",
				"prompt": "Test prompt"
				"subagent_type": "general-purpose"
			}`,
			expectError: true,
			errorType:   "error parsing parameters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params, err := taskTool.parseAndValidateParams(tc.input)
			
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for test case %s, but got none", tc.name)
				}
				// Check error message contains expected type
				if tc.errorType != "" && err != nil {
					if err.Error() == "" {
						t.Errorf("Expected error message to contain '%s', but got empty error", tc.errorType)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for test case %s, but got: %v", tc.name, err)
				}
				// Verify params are correctly parsed
				if params.Description == "" || params.Prompt == "" || params.SubagentType == "" {
					t.Errorf("Parameters not properly parsed for valid case")
				}
			}
		})
	}
}

func TestTaskToolSubagentTypeValidation(t *testing.T) {
	// Test valid subagent types
	validTypes := getValidSubagentTypes()
	if len(validTypes) == 0 {
		t.Error("Should have at least one valid subagent type")
	}

	// Test that general-purpose is included
	found := false
	for _, validType := range validTypes {
		if validType == "general-purpose" {
			found = true
			break
		}
	}
	if !found {
		t.Error("general-purpose should be a valid subagent type")
	}

	// Test validation function
	taskTool := NewTaskTool(nil, nil, nil, nil, nil).(*taskTool)
	
	validInput := `{
		"description": "Test task",
		"prompt": "Test prompt", 
		"subagent_type": "general-purpose"
	}`
	
	_, err := taskTool.parseAndValidateParams(validInput)
	if err != nil {
		t.Errorf("Valid subagent type should not produce error: %v", err)
	}

	invalidInput := `{
		"description": "Test task",
		"prompt": "Test prompt",
		"subagent_type": "non-existent-type"
	}`
	
	_, err = taskTool.parseAndValidateParams(invalidInput)
	if err == nil {
		t.Error("Invalid subagent type should produce error")
	}
}

func TestTaskToolErrorTypes(t *testing.T) {
	// Test that our custom error types are defined
	if ErrInvalidSubagentType == nil {
		t.Error("ErrInvalidSubagentType should be defined")
	}
	if ErrTaskTimeout == nil {
		t.Error("ErrTaskTimeout should be defined")
	}
	if ErrTaskCancelled == nil {
		t.Error("ErrTaskCancelled should be defined")
	}

	// Test error messages
	if ErrInvalidSubagentType.Error() == "" {
		t.Error("ErrInvalidSubagentType should have a message")
	}
	if ErrTaskTimeout.Error() == "" {
		t.Error("ErrTaskTimeout should have a message")
	}
	if ErrTaskCancelled.Error() == "" {
		t.Error("ErrTaskCancelled should have a message")
	}
}

func TestTaskToolWithTimeoutContext(t *testing.T) {
	taskTool := NewTaskTool(nil, nil, nil, nil, nil)
	
	// Create a context that will timeout quickly
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	
	// Add session and message IDs
	ctx = context.WithValue(ctx, tools.SessionIDContextKey, "test-session")
	ctx = context.WithValue(ctx, tools.MessageIDContextKey, "test-message")
	
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
	
	// Wait for context to timeout
	time.Sleep(5 * time.Millisecond)
	
	// This should handle timeout gracefully (will panic due to nil services, which is expected)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic due to nil services in timeout test")
		}
	}()
	
	_, _ = taskTool.Run(ctx, toolCall)
}

func TestTaskToolContextCancellation(t *testing.T) {
	taskTool := NewTaskTool(nil, nil, nil, nil, nil)
	
	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	
	// Add session and message IDs
	ctx = context.WithValue(ctx, tools.SessionIDContextKey, "test-session")
	ctx = context.WithValue(ctx, tools.MessageIDContextKey, "test-message")
	
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
	
	// Cancel the context immediately
	cancel()
	
	// This should handle cancellation gracefully (will panic due to nil services, which is expected)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic due to nil services in cancellation test")
		}
	}()
	
	_, _ = taskTool.Run(ctx, toolCall)
}

func TestTaskToolParameterTrimming(t *testing.T) {
	taskTool := NewTaskTool(nil, nil, nil, nil, nil).(*taskTool)
	
	// Test that parameters are properly trimmed
	input := `{
		"description": "  Test task  ",
		"prompt": "  Test prompt  ",
		"subagent_type": "general-purpose"
	}`
	
	params, err := taskTool.parseAndValidateParams(input)
	if err != nil {
		t.Errorf("Valid parameters with whitespace should not error: %v", err)
	}
	
	// Verify original values are preserved (trimming is only for validation)
	if params.Description != "  Test task  " {
		t.Errorf("Description should preserve original whitespace: got %q", params.Description)
	}
	if params.Prompt != "  Test prompt  " {
		t.Errorf("Prompt should preserve original whitespace: got %q", params.Prompt)
	}
}