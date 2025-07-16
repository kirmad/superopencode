package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kirmad/superopencode/internal/llm/tools"
	"github.com/kirmad/superopencode/internal/message"
)

func TestAnthropicReminderIntegration_EmptyTodos(t *testing.T) {
	client := &anthropicClient{
		options: anthropicOptions{disableCache: true}, // Disable cache for test
	}
	
	sessionID := "reminder-integration-test"
	ctx := context.WithValue(context.Background(), tools.SessionIDContextKey, sessionID)
	
	// Create test messages
	messages := []message.Message{
		{
			Role: message.User,
			Parts: []message.ContentPart{message.TextContent{Text: "Test message"}},
		},
	}
	
	// Convert messages - should include reminder as last message
	anthropicMessages := client.convertMessages(ctx, messages)
	
	// Should have 2 messages: original + reminder
	if len(anthropicMessages) != 2 {
		t.Errorf("Expected 2 messages (original + reminder), got %d", len(anthropicMessages))
	}
	
	// Check that last message contains reminder
	lastMsg := anthropicMessages[len(anthropicMessages)-1]
	if lastMsg.Role != "user" {
		t.Error("Last message should be a user message")
	}
}

func TestAnthropicReminderIntegration_WithTodos(t *testing.T) {
	client := &anthropicClient{
		options: anthropicOptions{disableCache: true},
	}
	
	sessionID := "reminder-integration-test-with-todos"
	ctx := context.WithValue(context.Background(), tools.SessionIDContextKey, sessionID)
	
	// Add a todo first
	writeTool := tools.NewTodoWriteTool()
	todos := []tools.TodoItem{
		{ID: "task1", Content: "Test task", Status: "pending", Priority: "high"},
	}
	inputBytes, _ := json.Marshal(map[string]interface{}{"todos": todos})
	writeCall := tools.ToolCall{ID: "write", Name: "TodoWrite", Input: string(inputBytes)}
	
	_, err := writeTool.Run(ctx, writeCall)
	if err != nil {
		t.Fatalf("Failed to write todos: %v", err)
	}
	
	// Create test messages
	messages := []message.Message{
		{
			Role: message.User,
			Parts: []message.ContentPart{message.TextContent{Text: "Test message"}},
		},
	}
	
	// Convert messages - should NOT include reminder
	anthropicMessages := client.convertMessages(ctx, messages)
	
	// Should have only 1 message: original (no reminder)
	if len(anthropicMessages) != 1 {
		t.Errorf("Expected 1 message (no reminder when todos exist), got %d", len(anthropicMessages))
	}
}

func TestOpenAIReminderIntegration_EmptyTodos(t *testing.T) {
	client := &openaiClient{
		providerOptions: providerClientOptions{
			systemMessage: "Test system message",
		},
	}
	
	sessionID := "openai-reminder-test"
	ctx := context.WithValue(context.Background(), tools.SessionIDContextKey, sessionID)
	
	// Create test messages
	messages := []message.Message{
		{
			Role: message.User,
			Parts: []message.ContentPart{message.TextContent{Text: "Test message"}},
		},
	}
	
	// Convert messages - should include system + original + reminder
	openaiMessages := client.convertMessages(ctx, messages)
	
	// Should have 3 messages: system + original + reminder
	if len(openaiMessages) != 3 {
		t.Errorf("Expected 3 messages (system + original + reminder), got %d", len(openaiMessages))
	}
}

func TestEmptySessionID_NoReminder(t *testing.T) {
	client := &anthropicClient{
		options: anthropicOptions{disableCache: true},
	}
	
	// Empty context (no session ID)
	ctx := context.Background()
	
	messages := []message.Message{
		{
			Role: message.User,
			Parts: []message.ContentPart{message.TextContent{Text: "Test message"}},
		},
	}
	
	// Convert messages - should NOT include reminder
	anthropicMessages := client.convertMessages(ctx, messages)
	
	// Should have only 1 message: original (no reminder without session ID)
	if len(anthropicMessages) != 1 {
		t.Errorf("Expected 1 message (no reminder without session ID), got %d", len(anthropicMessages))
	}
}