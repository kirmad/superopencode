# Task Tool Implementation Documentation

## Overview

This document provides comprehensive implementation details for adding a **Task tool** to the superopencode CLI application. The Task tool enables context-isolated subagent execution, allowing complex operations to run in separate LLM sessions for better context management and parallel processing.

## Architecture Analysis

### Current System Structure

The codebase follows a modular architecture with clear separation of concerns:

```
internal/
├── llm/
│   ├── agent/           # Agent orchestration and lifecycle
│   ├── tools/           # Tool implementations and interfaces
│   ├── models/          # LLM model definitions
│   ├── provider/        # LLM provider integrations
│   └── prompt/          # Prompt templates and management
├── config/              # Configuration management
├── session/             # Session lifecycle management
└── message/             # Message handling and persistence
```

### Existing Agent System

The current system uses a sophisticated agent architecture:

1. **Agent Service** (`internal/llm/agent/agent.go`): Core orchestration
2. **Agent Tool** (`internal/llm/agent/agent-tool.go`): Task delegation implementation
3. **Tool Interface** (`internal/llm/tools/tools.go`): Standard tool contract

## Implementation Specification

### 1. Simplified Task Architecture

#### 1.1 Lightweight Context Isolation

The Task tool creates **isolated execution contexts** using the existing infrastructure with minimal changes:

1. **Reuse Existing Infrastructure**: Same sessions service, messages service, and agent system
2. **Fresh Context Window**: New session with independent message history
3. **Simple Implementation**: Leverage existing `agent.NewAgent()` with different sessionID
4. **Minimal Complexity**: No database schema changes or complex relationships

#### 1.2 Task Execution Flow

```go
// Simple task execution pattern
func (t *taskTool) Run(ctx context.Context, call ToolCall) (ToolResponse, error) {
    // 1. Create new session for isolated context
    taskSession, err := t.sessions.Create(ctx, session.CreateSessionParams{
        Title: fmt.Sprintf("Task: %s", params.Description),
    })
    
    // 2. Create agent with fresh context using existing infrastructure
    taskAgent, err := agent.NewAgent(
        config.AgentTask,    // Same configuration as main agent
        t.sessions,         // Shared infrastructure
        t.messages,         // Shared infrastructure
        t.tools,           // Inherit same tools
    )
    
    // 3. Execute in isolated context
    done, err := taskAgent.Run(ctx, taskSession.ID, params.Prompt)
    result := <-done
    
    // 4. Simple cost tracking
    t.updateParentSessionCost(ctx, parentSessionID, taskSession.Cost)
    
    return NewTextResponse(result.Message.Content().String()), nil
}
```

### 2. Core Components

#### 2.1 Task Tool Structure

Create a new **Task tool** that leverages lightweight subagent forking:

```go
// internal/llm/tools/task.go
package tools

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/kirmad/superopencode/internal/config"
    "github.com/kirmad/superopencode/internal/llm/agent"
    "github.com/kirmad/superopencode/internal/message"
    "github.com/kirmad/superopencode/internal/session"
)

type taskTool struct {
    sessions session.Service
    messages message.Service
}

const TaskToolName = "Task"

type TaskParams struct {
    Description string `json:"description"`
    Prompt      string `json:"prompt"`
}

func (t *taskTool) Info() ToolInfo {
    return ToolInfo{
        Name:        TaskToolName,
        Description: "Launch a new agent to handle complex, multi-step tasks autonomously with isolated context. The agent operates in a completely separate context window and returns the final response.",
        Parameters: map[string]any{
            "description": map[string]any{
                "type":        "string",
                "description": "A short (3-5 word) description of the task",
            },
            "prompt": map[string]any{
                "type":        "string", 
                "description": "The task for the agent to perform",
            },
        },
        Required: []string{"description", "prompt"},
    }
}

func (t *taskTool) Run(ctx context.Context, call ToolCall) (ToolResponse, error) {
    var params TaskParams
    if err := json.Unmarshal([]byte(call.Input), &params); err != nil {
        return NewTextErrorResponse(fmt.Sprintf("error parsing parameters: %s", err)), nil
    }
    
    if params.Description == "" || params.Prompt == "" {
        return NewTextErrorResponse("both description and prompt are required"), nil
    }

    parentSessionID, messageID := GetContextValues(ctx)
    if parentSessionID == "" || messageID == "" {
        return ToolResponse{}, fmt.Errorf("session_id and message_id are required")
    }

    // Create new session for isolated context (no schema changes needed)
    taskSession, err := t.sessions.Create(ctx, session.CreateSessionParams{
        Title: fmt.Sprintf("Task: %s", params.Description),
    })
    if err != nil {
        return ToolResponse{}, fmt.Errorf("error creating task session: %s", err)
    }

    // Create agent with fresh context using existing infrastructure
    taskAgent, err := agent.NewAgent(
        config.AgentTask,  // Same configuration as main agent
        t.sessions,       // Shared infrastructure  
        t.messages,       // Shared infrastructure
        t.tools,         // Inherit same tools (simplified)
    )
    if err != nil {
        return ToolResponse{}, fmt.Errorf("error creating task agent: %s", err)
    }

    // Execute task in isolated context
    done, err := taskAgent.Run(ctx, taskSession.ID, params.Prompt)
    if err != nil {
        return ToolResponse{}, fmt.Errorf("error executing task: %s", err)
    }

    // Wait for completion and get result
    result := <-done
    if result.Error != nil {
        return ToolResponse{}, fmt.Errorf("task execution failed: %s", result.Error)
    }

    // Validate response format
    response := result.Message
    if response.Role != message.Assistant {
        return NewTextErrorResponse("invalid task response"), nil
    }

    // Simple cost aggregation
    if err := t.updateParentSessionCost(ctx, parentSessionID, taskSession.Cost); err != nil {
        return ToolResponse{}, fmt.Errorf("error updating session costs: %s", err)
    }

    return NewTextResponse(response.Content().String()), nil
}

// Simple cost aggregation helper
func (t *taskTool) updateParentSessionCost(ctx context.Context, parentID string, taskCost float64) error {
    parentSession, err := t.sessions.Get(ctx, parentID)
    if err != nil {
        return err
    }

    parentSession.Cost += taskCost
    _, err = t.sessions.Save(ctx, parentSession)
    return err
}

func NewTaskTool(sessions session.Service, messages message.Service, tools []tools.BaseTool) BaseTool {
    return &taskTool{
        sessions: sessions,
        messages: messages,
        tools:    tools,  // Inherit tools from parent agent
    }
}
```

#### 1.3 Tool Inheritance

Task agents inherit the same tools as the parent agent for simplicity:

```go
// Task agents inherit tools from parent - no special tool management needed
// This eliminates complexity of managing tool subsets

type taskTool struct {
    sessions session.Service
    messages message.Service
    tools    []tools.BaseTool  // Inherited from parent agent
}
```

### 2. Simple Integration

#### 2.1 Agent Configuration

Task agent uses the same configuration as the main agent (simplified):

```go
// No special configuration needed - AgentTask uses existing config
// This eliminates configuration complexity
```

#### 2.2 Tool Registration

Register the Task tool with existing tools:

```go
// Add Task tool to existing tool registry
func CoderAgentTools(lspClients map[string]*lsp.Client, sessions session.Service, messages message.Service) []tools.BaseTool {
    tools := []tools.BaseTool{
        // ... existing tools
        tools.NewGrepTool(),
        tools.NewGlobTool(),
        // ... other tools
    }
    
    // Add Task tool with inherited tools
    tools = append(tools, tools.NewTaskTool(sessions, messages, tools))
    return tools
}
```

### 3. Simple Session Management

#### 3.1 Basic Session Creation

Use existing session infrastructure with no schema changes:

```go
// Use existing session creation - no database changes needed
taskSession, err := t.sessions.Create(ctx, session.CreateSessionParams{
    Title: fmt.Sprintf("Task: %s", params.Description),
    // No parent_id, tool_call_id, or type fields needed
})
```

#### 3.2 UI Visibility Through Logging

Task execution is visible through the existing logging and message system:

```go
// Task execution visibility through existing systems
func (t *taskTool) Run(ctx context.Context, call ToolCall) (ToolResponse, error) {
    // Log task start
    logging.InfoPersist(fmt.Sprintf("Starting task: %s", params.Description))
    
    // Execute task with existing agent
    result := <-done
    
    // Log task completion
    logging.InfoPersist(fmt.Sprintf("Task completed: %s", params.Description))
    
    return NewTextResponse(result.Message.Content().String()), nil
}
```

### 4. Context Isolation Mechanism

#### 4.1 Session-Based Isolation

Context isolation is achieved through simple session separation:

```go
// Context isolation through different sessionID
func (a *agent) Run(ctx context.Context, sessionID string, content string) {
    // Agent loads messages ONLY from the provided sessionID
    msgs, err := a.messages.List(ctx, sessionID)  // Fresh context
    
    // Empty message history = fresh context window
    if len(msgs) == 0 {
        // Completely isolated execution
    }
    
    // All processing happens in this session's context
    msgHistory := append(msgs, userMsg)  // Isolated message history
}
```

#### 4.2 Existing Infrastructure Reuse

The Task tool leverages existing agent creation without modifications:

```go
// No changes needed to existing agent system
taskAgent, err := agent.NewAgent(
    config.AgentTask,  // Use existing agent configuration
    t.sessions,       // Use existing session service
    t.messages,       // Use existing message service  
    t.tools,         // Use existing tools
)

// Fresh context through different sessionID
done, err := taskAgent.Run(ctx, taskSession.ID, params.Prompt)
//                            ^^^^^^^^^^^^^ 
//                            Fresh session = fresh context
```

### 5. Simple Execution Flow

#### 5.1 End-to-End Process Flow

Simplified flow from request to completion:

```go
// 1. USER REQUEST IN MAIN SESSION
// User: "Use Task tool to analyze performance issues"
// Main session: "session-123"

// 2. TASK TOOL EXECUTION
func (t *taskTool) Run(ctx context.Context, call ToolCall) (ToolResponse, error) {
    parentSessionID := "session-123"  // From context
    
    // 3. CREATE NEW SESSION (FRESH CONTEXT)
    taskSession, err := t.sessions.Create(ctx, session.CreateSessionParams{
        Title: "Task: Performance Analysis",
    })
    // Result: taskSession.ID = "session-789" (completely isolated)
    
    // 4. CREATE AGENT WITH FRESH CONTEXT
    taskAgent, err := agent.NewAgent(config.AgentTask, t.sessions, t.messages, t.tools)
    
    // 5. EXECUTE IN ISOLATION
    done, err := taskAgent.Run(ctx, taskSession.ID, params.Prompt)
    //                              ^^^^^^^^^^^^^^
    //                              Fresh session = no parent history
    
    // 6. RETURN RESULT
    result := <-done
    t.updateParentSessionCost(ctx, parentSessionID, taskSession.Cost)
    return NewTextResponse(result.Message.Content().String()), nil
}
```

#### 5.2 Simple Message Flow

```
PARENT SESSION                    TASK SESSION (isolated)
┌─────────────────────────┐      ┌─────────────────────────┐
│ User: "Use Task tool"   │      │ User: "Analyze perf"    │
│                         │      │                         │
│ Assistant: [Running]    │      │ Assistant: [Analyzing]  │
│                         │      │   - Fresh context       │
│ Assistant: [Complete]   │←─────┤   - No parent history   │
│   Result: "Found..."    │      │   - Independent exec    │
└─────────────────────────┘      └─────────────────────────┘
```

### 6. Error Handling & Timeouts

#### 6.1 Timeout Management

Implement timeout handling for task execution:

```go
// Add timeout context to task execution
func (t *taskTool) Run(ctx context.Context, call ToolCall) (ToolResponse, error) {
    // Create timeout context (30 seconds default)
    timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // Use timeoutCtx for agent execution
    done, err := taskAgent.Run(timeoutCtx, taskSession.ID, params.Prompt)
    if err != nil {
        return ToolResponse{}, fmt.Errorf("error executing task: %s", err)
    }
    
    select {
    case result := <-done:
        // Handle result
        return t.processResult(result)
    case <-timeoutCtx.Done():
        return NewTextErrorResponse("task execution timeout"), nil
    }
}
```

#### 6.2 Error Recovery

Implement robust error handling:

```go
func (t *taskTool) processResult(result agent.AgentEvent) (ToolResponse, error) {
    if result.Error != nil {
        // Classify error type
        switch {
        case errors.Is(result.Error, context.DeadlineExceeded):
            return NewTextErrorResponse("task timeout"), nil
        case errors.Is(result.Error, context.Canceled):
            return NewTextErrorResponse("task cancelled"), nil
        default:
            return NewTextErrorResponse(fmt.Sprintf("task failed: %s", result.Error)), nil
        }
    }
    
    // Process successful result
    return NewTextResponse(result.Message.Content().String()), nil
}
```

### 7. Integration Patterns

#### 7.1 Tool Integration

Add Task tool to the standard tool registration:

```go
// In main tool setup (likely in cmd/root.go or similar)
tools := []tools.BaseTool{
    tools.NewBashTool(),
    tools.NewEditTool(),
    tools.NewGrepTool(),
    // ... other tools
    tools.NewTaskTool(sessionService, messageService), // Add here
}
```

#### 7.2 JSON Schema Integration

Ensure the Task tool is properly exposed via JSON schema:

```go
// Update schema generation to include Task tool
func generateToolSchema(tool tools.BaseTool) map[string]any {
    info := tool.Info()
    return map[string]any{
        "name":        info.Name,
        "description": info.Description,
        "input_schema": map[string]any{
            "type":       "object",
            "properties": info.Parameters,
            "required":   info.Required,
            "additionalProperties": false,
            "$schema": "http://json-schema.org/draft-07/schema#",
        },
    }
}
```

### 8. Testing Strategy

#### 8.1 Unit Tests

Create comprehensive unit tests:

```go
// internal/llm/tools/task_test.go
package tools

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestTaskTool_BasicExecution(t *testing.T) {
    // Mock dependencies
    mockSessions := &MockSessionService{}
    mockMessages := &MockMessageService{}
    
    // Setup test data
    taskTool := NewTaskTool(mockSessions, mockMessages)
    
    // Test basic execution
    ctx := context.Background()
    call := ToolCall{
        ID:    "test-call-id",
        Name:  TaskToolName,
        Input: `{"description": "test task", "prompt": "analyze this code"}`,
    }
    
    response, err := taskTool.Run(ctx, call)
    
    assert.NoError(t, err)
    assert.False(t, response.IsError)
    assert.NotEmpty(t, response.Content)
}

func TestTaskTool_TimeoutHandling(t *testing.T) {
    // Test timeout scenarios
    // ...
}

func TestTaskTool_ErrorHandling(t *testing.T) {
    // Test various error conditions
    // ...
}
```

#### 8.2 Integration Tests

Create integration tests with real agent execution:

```go
// internal/llm/tools/task_integration_test.go
func TestTaskTool_RealExecution(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Setup real dependencies
    // Test with actual LLM calls
    // ...
}
```

### 9. Performance Considerations

#### 9.1 Resource Management

Implement resource limits and monitoring:

```go
type TaskResourceLimits struct {
    MaxConcurrentTasks int
    MaxTokensPerTask   int64
    MaxExecutionTime   time.Duration
}

var defaultLimits = TaskResourceLimits{
    MaxConcurrentTasks: 3,
    MaxTokensPerTask:   4096,
    MaxExecutionTime:   30 * time.Second,
}
```

#### 9.2 Caching Strategy

Implement result caching for repetitive tasks:

```go
type TaskCache struct {
    cache map[string]ToolResponse
    mutex sync.RWMutex
    ttl   time.Duration
}

func (c *TaskCache) Get(key string) (ToolResponse, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    response, exists := c.cache[key]
    return response, exists
}
```

### 10. Security Considerations

#### 10.1 Input Validation

Implement comprehensive input validation:

```go
func (t *taskTool) validateInput(params TaskParams) error {
    if len(params.Description) > 200 {
        return fmt.Errorf("description too long (max 200 characters)")
    }
    if len(params.Prompt) > 10000 {
        return fmt.Errorf("prompt too long (max 10000 characters)")
    }
    if strings.TrimSpace(params.Description) == "" {
        return fmt.Errorf("description cannot be empty")
    }
    if strings.TrimSpace(params.Prompt) == "" {
        return fmt.Errorf("prompt cannot be empty")
    }
    return nil
}
```

#### 10.2 Isolation Guarantees

Ensure proper context isolation:

```go
// Ensure task agents cannot access parent session data
func createIsolatedTaskAgent() *agent.Agent {
    return &agent.Agent{
        tools:     TaskAgentTools(), // Limited toolset
        context:   newIsolatedContext(), // Fresh context
        resources: limitedResources(), // Resource constraints
    }
}
```

## Usage Examples

### Basic Task Execution

```json
{
    "name": "Task",
    "input": {
        "description": "Code analysis",
        "prompt": "Analyze the main.go file and identify potential performance issues"
    }
}
```

### Complex Research Task

```json
{
    "name": "Task",
    "input": {
        "description": "API research",
        "prompt": "Research the latest best practices for REST API authentication and provide 3 specific recommendations with code examples"
    }
}
```

## Implementation Path

### Single Phase Implementation
1. Create Task tool with simple context isolation
2. Add to tool registry with inherited tools
3. Basic error handling and timeouts
4. Simple cost aggregation

**Benefits**: Complete functionality in single deployment, no complex migration needed.

## Deployment Considerations

### No Configuration Changes Needed

The simplified implementation requires no configuration updates:

- Uses existing agent configuration
- No database schema changes
- No new dependencies
- Backward compatible by design

### Simple Deployment

1. Add Task tool implementation file
2. Register tool in existing tool registry
3. Deploy - no migrations or config changes needed

## Monitoring & Observability

### Metrics Collection

Track key metrics for task execution:

- Task execution time
- Success/failure rates
- Resource usage per task
- Concurrent task counts

### Logging Integration

Integrate with existing logging system:

```go
func (t *taskTool) logTaskExecution(params TaskParams, duration time.Duration, success bool) {
    logging.InfoPersist(fmt.Sprintf(
        "Task executed: description=%s, duration=%v, success=%t",
        params.Description,
        duration,
        success,
    ))
}
```

This comprehensive implementation provides a robust, secure, and performant Task tool that enables context-isolated subagent execution while maintaining consistency with the existing codebase architecture and patterns.