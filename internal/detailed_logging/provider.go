package detailed_logging

import (
	"context"
	"time"

	"github.com/kirmad/superopencode/internal/llm/models"
	"github.com/kirmad/superopencode/internal/llm/provider"
	"github.com/kirmad/superopencode/internal/llm/tools"
	"github.com/kirmad/superopencode/internal/message"
)

// LoggingProvider wraps an LLM provider to add logging
type LoggingProvider struct {
	wrapped  provider.Provider
	logger   *DetailedLogger
	provider string
}

// NewLoggingProvider creates a new logging provider wrapper
func NewLoggingProvider(provider provider.Provider, providerName string, logger *DetailedLogger) provider.Provider {
	return &LoggingProvider{
		wrapped:  provider,
		logger:   logger,
		provider: providerName,
	}
}


func (lp *LoggingProvider) calculateCost(model string, usage *TokenUsage) float64 {
	// Basic cost calculation - you'd want to expand this with actual pricing
	costPerMillion := map[string]struct{ input, output float64 }{
		"gpt-4":         {30.0, 60.0},
		"gpt-4-turbo":   {10.0, 30.0},
		"gpt-3.5-turbo": {0.5, 1.5},
		"claude-3-opus": {15.0, 75.0},
		"claude-3-sonnet": {3.0, 15.0},
	}

	costs, ok := costPerMillion[model]
	if !ok {
		return 0
	}

	inputCost := float64(usage.Prompt) / 1_000_000 * costs.input
	outputCost := float64(usage.Completion) / 1_000_000 * costs.output

	return inputCost + outputCost
}

// Model returns the underlying model
func (lp *LoggingProvider) Model() models.Model {
	return lp.wrapped.Model()
}

// SendMessages implements the Provider interface
func (lp *LoggingProvider) SendMessages(ctx context.Context, messages []message.Message, tools []tools.BaseTool) (*provider.ProviderResponse, error) {
	if lp.logger == nil || !lp.logger.IsEnabled() {
		return lp.wrapped.SendMessages(ctx, messages, tools)
	}

	// Create LLM call log
	llmLog := &LLMCallLog{
		ID:        NewID(),
		SessionID: lp.logger.sessionID,
		Provider:  lp.provider,
		Model:     string(lp.wrapped.Model().ID),
		StartTime: time.Now(),
		Request:   lp.messagesToMap(messages),
	}

	// Set parent tool call if in context
	if toolID := lp.logger.GetCurrentToolCall(); toolID != "" {
		llmLog.ParentToolCall = toolID
	}

	// Make the actual call
	resp, err := lp.wrapped.SendMessages(ctx, messages, tools)

	// Complete the log entry
	endTime := time.Now()
	llmLog.EndTime = &endTime
	llmLog.DurationMs = CalculateDuration(llmLog.StartTime, llmLog.EndTime)

	if err != nil {
		llmLog.Error = err.Error()
		lp.logger.LogLLMCall(llmLog)
		return resp, err
	}

	// Log response
	if resp != nil {
		llmLog.Response = lp.providerResponseToMap(resp)
		llmLog.TokensUsed = &TokenUsage{
			Prompt:     int(resp.Usage.InputTokens),
			Completion: int(resp.Usage.OutputTokens),
			Total:      int(resp.Usage.InputTokens + resp.Usage.OutputTokens),
		}

		// Calculate cost if possible
		if cost := lp.calculateCost(string(lp.wrapped.Model().ID), llmLog.TokensUsed); cost > 0 {
			llmLog.Cost = &cost
		}
	}

	lp.logger.LogLLMCall(llmLog)
	return resp, nil
}

// StreamResponse implements the Provider interface
func (lp *LoggingProvider) StreamResponse(ctx context.Context, messages []message.Message, tools []tools.BaseTool) <-chan provider.ProviderEvent {
	if lp.logger == nil || !lp.logger.IsEnabled() {
		return lp.wrapped.StreamResponse(ctx, messages, tools)
	}

	// Create LLM call log
	llmLog := &LLMCallLog{
		ID:           NewID(),
		SessionID:    lp.logger.sessionID,
		Provider:     lp.provider,
		Model:        string(lp.wrapped.Model().ID),
		StartTime:    time.Now(),
		Request:      lp.messagesToMap(messages),
		StreamEvents: []StreamEvent{},
	}

	// Set parent tool call if in context
	if toolID := lp.logger.GetCurrentToolCall(); toolID != "" {
		llmLog.ParentToolCall = toolID
	}

	// Get the original stream
	originalStream := lp.wrapped.StreamResponse(ctx, messages, tools)

	// Create a new channel for our wrapped stream
	wrappedStream := make(chan provider.ProviderEvent)

	// Start a goroutine to process events
	go func() {
		defer close(wrappedStream)
		defer func() {
			endTime := time.Now()
			llmLog.EndTime = &endTime
			llmLog.DurationMs = CalculateDuration(llmLog.StartTime, llmLog.EndTime)
			lp.logger.LogLLMCall(llmLog)
		}()

		for event := range originalStream {
			// Log the event
			streamEvent := StreamEvent{
				Type:      string(event.Type),
				Timestamp: time.Now(),
				Data:      lp.eventToMap(event),
			}
			llmLog.StreamEvents = append(llmLog.StreamEvents, streamEvent)

			// Forward the event
			wrappedStream <- event

			// Capture final response if available
			if event.Type == provider.EventComplete && event.Response != nil {
				llmLog.Response = lp.providerResponseToMap(event.Response)
				llmLog.TokensUsed = &TokenUsage{
					Prompt:     int(event.Response.Usage.InputTokens),
					Completion: int(event.Response.Usage.OutputTokens),
					Total:      int(event.Response.Usage.InputTokens + event.Response.Usage.OutputTokens),
				}

				// Calculate cost if possible
				if cost := lp.calculateCost(string(lp.wrapped.Model().ID), llmLog.TokensUsed); cost > 0 {
					llmLog.Cost = &cost
				}
			}
		}
	}()

	return wrappedStream
}

// Helper methods

func (lp *LoggingProvider) messagesToMap(messages []message.Message) map[string]interface{} {
	// Convert messages to a simple map format
	result := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		msgMap := map[string]interface{}{
			"role": msg.Role,
			"id":   msg.ID,
		}
		
		// Add content parts
		parts := make([]interface{}, 0)
		for _, part := range msg.Parts {
			switch p := part.(type) {
			case message.TextContent:
				parts = append(parts, map[string]interface{}{
					"type": "text",
					"text": p.Text,
				})
			case message.ToolCall:
				parts = append(parts, map[string]interface{}{
					"type":      "tool_call",
					"id":        p.ID,
					"name":      p.Name,
					"input": p.Input,
				})
			case message.ToolResult:
				parts = append(parts, map[string]interface{}{
					"type":        "tool_result",
					"tool_call_id": p.ToolCallID,
					"content":     p.Content,
					"is_error":    p.IsError,
				})
			}
		}
		msgMap["parts"] = parts
		result[i] = msgMap
	}
	
	return map[string]interface{}{"messages": result}
}

func (lp *LoggingProvider) providerResponseToMap(resp *provider.ProviderResponse) map[string]interface{} {
	result := map[string]interface{}{
		"content":       resp.Content,
		"finish_reason": resp.FinishReason,
		"usage": map[string]interface{}{
			"input_tokens":          resp.Usage.InputTokens,
			"output_tokens":         resp.Usage.OutputTokens,
			"cache_creation_tokens": resp.Usage.CacheCreationTokens,
			"cache_read_tokens":     resp.Usage.CacheReadTokens,
		},
	}
	
	if len(resp.ToolCalls) > 0 {
		toolCalls := make([]map[string]interface{}, len(resp.ToolCalls))
		for i, tc := range resp.ToolCalls {
			toolCalls[i] = map[string]interface{}{
				"id":        tc.ID,
				"name":      tc.Name,
				"input": tc.Input,
			}
		}
		result["tool_calls"] = toolCalls
	}
	
	return result
}

func (lp *LoggingProvider) eventToMap(event provider.ProviderEvent) map[string]interface{} {
	result := map[string]interface{}{
		"type": string(event.Type),
	}
	
	if event.Content != "" {
		result["content"] = event.Content
	}
	
	if event.Thinking != "" {
		result["thinking"] = event.Thinking
	}
	
	if event.Error != nil {
		result["error"] = event.Error.Error()
	}
	
	if event.ToolCall != nil {
		result["tool_call"] = map[string]interface{}{
			"id":        event.ToolCall.ID,
			"name":      event.ToolCall.Name,
			"input": event.ToolCall.Input,
		}
	}
	
	if event.Response != nil {
		result["response"] = lp.providerResponseToMap(event.Response)
	}
	
	return result
}