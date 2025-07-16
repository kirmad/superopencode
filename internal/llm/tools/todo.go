package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

const (
	TodoReadToolName  = "TodoRead"
	TodoWriteToolName = "TodoWrite"
)

// TodoItem represents a single todo item
type TodoItem struct {
	ID       string `json:"id"`
	Content  string `json:"content"`
	Status   string `json:"status"`   // pending, in_progress, completed
	Priority string `json:"priority"` // high, medium, low
}

// TodoStorage manages the in-memory todo list for the session
type TodoStorage struct {
	mu    sync.RWMutex
	todos map[string][]TodoItem // sessionID -> todos
}

var todoStorage = &TodoStorage{
	todos: make(map[string][]TodoItem),
}

// GetTodoCount returns the number of todos for a given session
func GetTodoCount(sessionID string) int {
	if sessionID == "" {
		return 0
	}
	
	todoStorage.mu.RLock()
	defer todoStorage.mu.RUnlock()
	
	todos := todoStorage.todos[sessionID]
	return len(todos)
}

// TodoReadTool implements the TodoRead functionality
type TodoReadTool struct{}

func NewTodoReadTool() *TodoReadTool {
	return &TodoReadTool{}
}

func (t *TodoReadTool) Info() ToolInfo {
	return ToolInfo{
		Name: TodoReadToolName,
		Description: `Use this tool to read the current to-do list for the session. This tool should be used proactively and frequently to ensure that you are aware of
the status of the current task list. You should make use of this tool as often as possible, especially in the following situations:
- At the beginning of conversations to see what's pending
- Before starting new tasks to prioritize work
- When the user asks about previous tasks or plans
- Whenever you're uncertain about what to do next
- After completing tasks to update your understanding of remaining work
- After every few messages to ensure you're on track

Usage:
- This tool takes in no parameters. So leave the input blank or empty. DO NOT include a dummy object, placeholder string or a key like "input" or "empty". LEAVE IT BLANK.
- Returns a list of todo items with their status, priority, and content
- Use this information to track progress and plan next steps
- If no todos exist yet, an empty list will be returned`,
		Parameters: map[string]any{
			"type":                 "object",
			"properties":           map[string]any{},
			"additionalProperties": true,
			"description":          "No input required - returns current todo list state",
		},
		Required: []string{},
	}
}

func (t *TodoReadTool) Run(ctx context.Context, params ToolCall) (ToolResponse, error) {
	sessionID, _ := GetContextValues(ctx)
	if sessionID == "" {
		return NewTextErrorResponse("No session ID found"), nil
	}

	todoStorage.mu.RLock()
	todos := todoStorage.todos[sessionID]
	todoStorage.mu.RUnlock()

	if len(todos) == 0 {
		return NewTextResponse("No todos found for this session."), nil
	}

	result, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return NewTextErrorResponse("Failed to serialize todos"), nil
	}

	return NewTextResponse(string(result)), nil
}

// TodoWriteTool implements the TodoWrite functionality
type TodoWriteTool struct{}

func NewTodoWriteTool() *TodoWriteTool {
	return &TodoWriteTool{}
}

func (t *TodoWriteTool) Info() ToolInfo {
	return ToolInfo{
		Name: TodoWriteToolName,
		Description: `Use this tool to create and manage a structured task list for your current coding session. This helps you track progress, organize complex tasks, and demonstrate thoroughness to the user.
It also helps the user understand the progress of the task and overall progress of their requests.

## When to Use This Tool
Use this tool proactively in these scenarios:

1. Complex multi-step tasks - When a task requires 3 or more distinct steps or actions
2. Non-trivial and complex tasks - Tasks that require careful planning or multiple operations
3. User explicitly requests todo list - When the user directly asks you to use the todo list
4. User provides multiple tasks - When users provide a list of things to be done (numbered or comma-separated)
5. After receiving new instructions - Immediately capture user requirements as todos
6. When you start working on a task - Mark it as in_progress BEFORE beginning work. Ideally you should only have one todo as in_progress at a time
7. After completing a task - Mark it as completed and add any new follow-up tasks discovered during implementation

## When NOT to Use This Tool

Skip using this tool when:
1. There is only a single, straightforward task
2. The task is trivial and tracking it provides no organizational benefit
3. The task can be completed in less than 3 trivial steps
4. The task is purely conversational or informational

NOTE that you should not use this tool if there is only one trivial task to do. In this case you are better off just doing the task directly.

## Task States and Management

1. **Task States**: Use these states to track progress:
   - pending: Task not yet started
   - in_progress: Currently working on (limit to ONE task at a time)
   - completed: Task finished successfully

2. **Task Management**:
   - Update task status in real-time as you work
   - Mark tasks complete IMMEDIATELY after finishing (don't batch completions)
   - Only have ONE task in_progress at any time
   - Complete current tasks before starting new ones
   - Remove tasks that are no longer relevant from the list entirely

3. **Task Completion Requirements**:
   - ONLY mark a task as completed when you have FULLY accomplished it
   - If you encounter errors, blockers, or cannot finish, keep the task as in_progress
   - When blocked, create a new task describing what needs to be resolved
   - Never mark a task as completed if:
     - Tests are failing
     - Implementation is partial
     - You encountered unresolved errors
     - You couldn't find necessary files or dependencies

4. **Task Breakdown**:
   - Create specific, actionable items
   - Break complex tasks into smaller, manageable steps
   - Use clear, descriptive task names

When in doubt, use this tool. Being proactive with task management demonstrates attentiveness and ensures you complete all requirements successfully.`,
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"todos": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"content": map[string]any{
								"type":      "string",
								"minLength": 1,
							},
							"status": map[string]any{
								"type": "string",
								"enum": []string{"pending", "in_progress", "completed"},
							},
							"priority": map[string]any{
								"type": "string",
								"enum": []string{"high", "medium", "low"},
							},
							"id": map[string]any{
								"type": "string",
							},
						},
						"required":             []string{"content", "status", "priority", "id"},
						"additionalProperties": false,
					},
					"description": "The updated todo list",
				},
			},
			"required":             []string{"todos"},
			"additionalProperties": false,
		},
		Required: []string{"todos"},
	}
}

func (t *TodoWriteTool) Run(ctx context.Context, params ToolCall) (ToolResponse, error) {
	sessionID, _ := GetContextValues(ctx)
	if sessionID == "" {
		return NewTextErrorResponse("No session ID found"), nil
	}

	var input struct {
		Todos []TodoItem `json:"todos"`
	}

	if err := json.Unmarshal([]byte(params.Input), &input); err != nil {
		return NewTextErrorResponse(fmt.Sprintf("Invalid input format: %v", err)), nil
	}

	// Validate todos
	inProgressCount := 0
	for _, todo := range input.Todos {
		if todo.Status == "in_progress" {
			inProgressCount++
		}
		if todo.Content == "" {
			return NewTextErrorResponse("Todo content cannot be empty"), nil
		}
		if todo.ID == "" {
			return NewTextErrorResponse("Todo ID cannot be empty"), nil
		}
	}

	if inProgressCount > 1 {
		return NewTextErrorResponse("Only one task can be in_progress at a time"), nil
	}

	// Store todos
	todoStorage.mu.Lock()
	todoStorage.todos[sessionID] = input.Todos
	todoStorage.mu.Unlock()

	// Return both success message and the updated todo list as JSON
	result, err := json.MarshalIndent(input.Todos, "", "  ")
	if err != nil {
		return NewTextErrorResponse("Failed to serialize updated todos"), nil
	}

	// Return JSON so the UI can render it as checkboxes
	return NewTextResponse(string(result)), nil
}