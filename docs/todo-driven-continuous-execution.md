# Todo-Driven Continuous Agent Execution Design

## Overview

This document describes the design for updating the SuperOpenCode client to enable continuous agent execution until all todos are complete. Currently, the agent stops execution when it finishes a response, even if todos remain incomplete. This enhancement transforms the agent from an interactive assistant into an autonomous task executor.

## Problem Statement

The current agent architecture stops execution when `FinishReason != ToolUse` (in `internal/llm/agent/agent.go:310`), returning control to the user even when pending or in-progress todos exist. This requires manual intervention to continue working through task lists, reducing productivity and automation potential.

## Solution Architecture

### Core Approach

Intercept the agent's stopping condition and inject todo-driven continuation prompts when incomplete tasks remain. This maintains the existing execution loop while adding autonomous task completion capability.

### Key Design Principles

1. **Minimal Invasiveness**: Single hook point in existing execution loop
2. **Configurable**: Can be enabled/disabled per agent or globally  
3. **Safe**: Multiple safety mechanisms prevent infinite loops and resource exhaustion
4. **Transparent**: User sees natural continuation with clear feedback
5. **Backwards Compatible**: Existing behavior unchanged when disabled

## Implementation Plan

### Phase 1: Core Implementation

#### 1. Configuration Layer Enhancement

**File: `/internal/config/config.go`**

```go
type AgentConfig struct {
    Model                models.ModelID `json:"model"`
    MaxTokens           int            `json:"max_tokens,omitempty"`
    ReasoningEffort     int            `json:"reasoning_effort,omitempty"`
    
    // NEW: Todo-driven execution settings
    TodoDrivenExecution  bool `json:"todo_driven_execution,omitempty"`
    MaxTodoContinuations int  `json:"max_todo_continuations,omitempty"` // default: 10
    TodoCheckIntervalMs  int  `json:"todo_check_interval_ms,omitempty"`
}

// Helper method
func (c *Config) IsTodoDrivenExecutionEnabled(agentName AgentName) bool {
    if agentConfig, ok := c.Agents[agentName]; ok {
        return agentConfig.TodoDrivenExecution
    }
    return false
}
```

#### 2. Todo Continuation Logic

**File: `/internal/llm/tools/todo_continuation.go` (NEW)**

```go
package tools

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "github.com/kirmad/superopencode/internal/message"
)

// TodoContinuationTracker manages continuation attempts per session
type TodoContinuationTracker struct {
    mu            sync.RWMutex
    continuations map[string]int    // sessionID -> count
    lastCheck     map[string]int64  // sessionID -> timestamp
}

var continuationTracker = &TodoContinuationTracker{
    continuations: make(map[string]int),
    lastCheck:     make(map[string]int64),
}

// hasIncompleteTodos checks if session has pending or in-progress todos
func hasIncompleteTodos(sessionID string) bool {
    todoStorage.mu.RLock()
    defer todoStorage.mu.RUnlock()
    
    todos := todoStorage.todos[sessionID]
    for _, todo := range todos {
        if todo.Status == "pending" || todo.Status == "in_progress" {
            return true
        }
    }
    return false
}

// getNextPriorityTodo returns the highest priority incomplete todo
func getNextPriorityTodo(sessionID string) *TodoItem {
    todoStorage.mu.RLock()
    defer todoStorage.mu.RUnlock()
    
    todos := todoStorage.todos[sessionID]
    
    // Priority order: high -> medium -> low
    priorities := []string{"high", "medium", "low"}
    for _, priority := range priorities {
        for _, todo := range todos {
            if (todo.Status == "pending" || todo.Status == "in_progress") && 
               todo.Priority == priority {
                return &todo
            }
        }
    }
    return nil
}

// generateTodoContinuationPrompt creates a continuation message
func generateTodoContinuationPrompt(sessionID string) message.Message {
    incompleteTodos := getIncompleteTodos(sessionID)
    nextTodo := getNextPriorityTodo(sessionID)
    
    var promptText string
    if nextTodo != nil {
        promptText = fmt.Sprintf(
            "You have %d incomplete tasks remaining. Please continue working on the next high-priority task: '%s'. "+
            "Continue until all todos are completed.",
            len(incompleteTodos), nextTodo.Content)
    } else {
        promptText = "Please check your todo list and continue working on any remaining incomplete tasks."
    }
    
    return message.Message{
        Role:  message.User,
        Parts: []message.ContentPart{message.TextContent{Text: promptText}},
    }
}

// incrementContinuationCount tracks continuation attempts
func incrementContinuationCount(sessionID string) int {
    continuationTracker.mu.Lock()
    defer continuationTracker.mu.Unlock()
    
    continuationTracker.continuations[sessionID]++
    continuationTracker.lastCheck[sessionID] = time.Now().Unix()
    
    return continuationTracker.continuations[sessionID]
}

// exceededMaxContinuations checks if continuation limit reached
func exceededMaxContinuations(sessionID string) bool {
    continuationTracker.mu.RLock()
    defer continuationTracker.mu.RUnlock()
    
    count := continuationTracker.continuations[sessionID]
    return count >= getMaxContinuations() // default: 10
}

// resetContinuationCount resets counter when new user message arrives
func resetContinuationCount(sessionID string) {
    continuationTracker.mu.Lock()
    defer continuationTracker.mu.Unlock()
    
    delete(continuationTracker.continuations, sessionID)
    delete(continuationTracker.lastCheck, sessionID)
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

// shouldContinueForTodos determines if execution should continue
func shouldContinueForTodos(ctx context.Context, sessionID string, finishReason message.FinishReason) bool {
    // Only continue if:
    // 1. Todo-driven execution is enabled
    // 2. Agent finished normally (not cancelled/error)
    // 3. There are incomplete todos
    // 4. Haven't exceeded max continuations
    
    if !isTodoDrivenExecutionEnabled(sessionID) {
        return false
    }
    
    if finishReason != message.FinishReasonEndTurn {
        return false
    }
    
    return hasIncompleteTodos(sessionID) && !exceededMaxContinuations(sessionID)
}

// isTodoDrivenExecutionEnabled checks if feature is enabled
func isTodoDrivenExecutionEnabled(sessionID string) bool {
    // TODO: Get from config based on agent type
    // For now, check environment variable
    return os.Getenv("SUPEROPENCODE_AUTO_COMPLETE_TODOS") == "true"
}

// getMaxContinuations returns the maximum allowed continuations
func getMaxContinuations() int {
    // TODO: Get from config
    return 10
}
```

#### 3. Agent Execution Enhancement

**File: `/internal/llm/agent/agent.go` (MODIFY)**

Modify the `processGeneration()` method around line 310:

```go
// In processGeneration() method
func (a *agent) processGeneration(ctx context.Context, sessionID, content string, attachmentParts []message.ContentPart) AgentEvent {
    // ... existing code ...
    
    for {
        // Check for cancellation before each iteration
        select {
        case <-ctx.Done():
            return a.err(ctx.Err())
        default:
            // Continue processing
        }
        
        agentMessage, toolResults, err := a.streamAndHandleEvents(ctx, sessionID, msgHistory)
        if err != nil {
            // ... existing error handling ...
        }
        
        // ... existing logging ...
        
        if (agentMessage.FinishReason() == message.FinishReasonToolUse) && toolResults != nil {
            // We are not done, we need to respond with the tool response
            msgHistory = append(msgHistory, agentMessage, *toolResults)
            continue
        }
        
        // NEW: Check for todo-driven continuation before stopping
        if tools.ShouldContinueForTodos(ctx, sessionID, agentMessage.FinishReason()) {
            todoPrompt := tools.GenerateTodoContinuationPrompt(sessionID)
            tools.IncrementContinuationCount(sessionID)
            msgHistory = append(msgHistory, agentMessage, todoPrompt)
            continue
        }
        
        return AgentEvent{
            Type:    AgentEventTypeResponse,
            Message: agentMessage,
            Done:    true,
        }
    }
}

// Also modify createUserMessage to reset continuation counter
func (a *agent) createUserMessage(ctx context.Context, sessionID, content string, attachmentParts []message.ContentPart) (message.Message, error) {
    // Reset continuation counter when new user message arrives
    tools.ResetContinuationCount(sessionID)
    
    parts := []message.ContentPart{message.TextContent{Text: content}}
    parts = append(parts, attachmentParts...)
    return a.messages.Create(ctx, sessionID, message.CreateMessageParams{
        Role:  message.User,
        Parts: parts,
    })
}
```

### Phase 2: Safety Mechanisms

#### Token Usage Monitoring

```go
// Add to todo_continuation.go
func checkTokenUsage(sessionID string, currentTokens int) bool {
    // Monitor approaching model token limits
    // Return false if nearing limit to prevent exhaustion
    maxTokens := getModelMaxTokens() // From config
    return currentTokens < (maxTokens * 0.9) // 90% threshold
}
```

#### Context Size Management

```go
// Add context trimming when conversation gets too long
func trimContextIfNeeded(msgHistory []message.Message) []message.Message {
    const maxHistoryLength = 50 // Configurable
    if len(msgHistory) > maxHistoryLength {
        // Keep recent messages and preserve todo context
        return msgHistory[len(msgHistory)-maxHistoryLength:]
    }
    return msgHistory
}
```

#### Error Recovery

```go
// Add error handling for failed todos
func handleTodoFailure(sessionID string, todoID string, error string) {
    // Mark todo as failed after multiple attempts
    // Skip or retry based on error type
    // Maintain session integrity
}
```

### Phase 3: Enhanced User Experience

#### CLI Integration

```bash
# Command line flags
superopencode --auto-complete-todos
superopencode --max-todo-continuations=20

# Environment variables
export SUPEROPENCODE_AUTO_COMPLETE_TODOS=true
export SUPEROPENCODE_MAX_TODO_CONTINUATIONS=15
```

#### Configuration File

```json
{
  "agents": {
    "coder": {
      "model": "claude-3-5-sonnet-20241022",
      "todo_driven_execution": true,
      "max_todo_continuations": 10,
      "todo_check_interval_ms": 1000
    },
    "task": {
      "model": "claude-3-5-sonnet-20241022", 
      "todo_driven_execution": false
    }
  }
}
```

## Safety Mechanisms

### 1. Infinite Loop Prevention
- **Max Continuations**: Default limit of 10 per session
- **Continuation Counter**: Tracks attempts per session
- **Reset Logic**: Counter resets on new user input

### 2. Resource Management
- **Token Monitoring**: Stop before approaching model limits
- **Context Trimming**: Manage growing conversation history
- **Memory Usage**: Monitor todo storage growth

### 3. Error Handling
- **Graceful Degradation**: Continue with other todos if one fails
- **Retry Logic**: Attempt failed todos once before skipping
- **Clear Messaging**: Inform user when limits are reached

### 4. User Control
- **Cancellation**: User can interrupt at any time
- **Override**: Manual control always available
- **Transparency**: Clear indication of auto-execution mode

## Edge Cases

### 1. Malformed Todos
- Skip todos with empty content
- Handle missing required fields gracefully
- Validate todo structure before processing

### 2. Conflicting Tasks
- Detect when todos conflict with each other
- Prioritize based on todo priority levels
- Provide clear error messages for conflicts

### 3. Permission Blocks
- Handle when agent needs user approval
- Pause auto-execution for permission requests
- Resume after user provides permission

### 4. Tool Failures
- Continue with other todos if specific tools fail
- Retry tool execution once before failing
- Log failures for debugging

## Testing Strategy

### Unit Tests
- Todo continuation decision logic
- Configuration parsing and validation
- Safety mechanism triggers
- Prompt generation for different todo states

### Integration Tests
- End-to-end agent execution with todo completion
- Multi-step task workflows
- Error handling and recovery
- Permission system integration

### Edge Case Testing
- Infinite loop prevention
- Token limit scenarios
- Malformed todo handling
- Concurrent session management

### Performance Testing
- Memory usage with large todo lists
- Execution time with many continuations
- Context size growth over iterations

## Success Metrics

1. **Completion Rate**: All pending todos transition to completed status
2. **Safety**: No infinite loops in normal usage
3. **Compatibility**: Existing functionality remains unaffected
4. **User Experience**: Clear feedback throughout process
5. **Efficiency**: Reasonable token usage and execution time

## Migration Path

### Phase 1: Optional Feature
- Deploy as disabled-by-default feature
- Allow opt-in via configuration or environment variable
- Gather user feedback and usage patterns

### Phase 2: Enhanced Configuration
- Add CLI flags and advanced settings
- Implement progress reporting and monitoring
- Add performance optimizations

### Phase 3: Production Ready
- Enable by default for coder agent
- Add comprehensive documentation
- Provide troubleshooting guides

## Conclusion

This design provides a robust, configurable solution for autonomous todo completion while maintaining all existing capabilities. The single hook point approach minimizes implementation complexity while the comprehensive safety mechanisms ensure reliable operation. The feature can be gradually rolled out with minimal risk to existing functionality.