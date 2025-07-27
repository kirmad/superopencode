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
		Description: `ESSENTIAL SITUATIONAL AWARENESS TOOL - Use this tool constantly to maintain complete awareness of your task progress and ensure nothing is forgotten.

<monitoring-protocol>
**CRITICAL: You MUST check the todo list frequently to maintain operational excellence**

**MANDATORY USAGE FREQUENCY:**
- IMMEDIATELY at the start of every conversation
- BEFORE making any decisions about what to do next
- AFTER completing any significant action or task
- WHENEVER you feel uncertain about priorities or next steps
- CONSTANTLY throughout multi-step operations
- BEFORE responding to user questions about progress or plans
</monitoring-protocol>

## Professional Development Standard
Checking your todo list frequently is a hallmark of professional developers and ensures:
- **User Confidence**: Shows users you're organized and tracking progress
- **No Forgotten Tasks**: Prevents overlooking important work
- **Smart Prioritization**: Helps you choose the most important next action
- **Progress Transparency**: Keeps everyone informed of current status

## Usage Requirements
- This tool takes NO parameters - leave input completely empty
- Returns clean JSON for perfect editor display with checkboxes and progress tracking
- Use the returned information to make informed decisions about next steps
- If no todos exist, you'll get an empty list - consider creating todos for multi-step work`,
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
		Description: `CRITICAL OPERATIONAL REQUIREMENT - This tool is MANDATORY for professional task management and user transparency. Your reputation for thoroughness depends on proper todo usage.

<mandatory-compliance-protocol>
**NON-NEGOTIABLE: You MUST use this tool for ANY operation with 3+ steps**

**IMMEDIATE ACTION TRIGGERS (Use TodoWrite BEFORE starting work):**
- Multi-step coding, analysis, debugging, or implementation tasks
- User requests containing multiple components or requirements  
- Complex problem-solving requiring systematic approach
- File modifications across multiple locations
- Any task where you might forget a step or lose track of progress
- Building, testing, or deployment workflows
- Research tasks with multiple investigation points

**COMPLETION PSYCHOLOGY - CRITICAL FOR USER SATISFACTION:**
Your users DEPEND on seeing progress and knowing that tasks will be completed fully. 
Incomplete or abandoned work destroys user confidence and wastes their time.
</mandatory-compliance-protocol>

## Professional Excellence Standards

<completion-requirements>
**SACRED RULE: Mark "completed" ONLY when work is 100% finished**

**NEVER mark completed if:**
- Tests are failing or haven't been run
- Implementation is partial or untested  
- Errors remain unresolved
- Files or dependencies are missing
- User requirements are not fully satisfied

**FAILURE TO COMPLETE FULLY:**
- Destroys user trust and confidence
- Wastes user time and effort
- Creates technical debt and confusion
- Reflects poorly on your capabilities
</completion-requirements>

## Status Management Protocol

<status-discipline>
**MANDATORY REAL-TIME STATUS UPDATES:**
- Mark task "in_progress" IMMEDIATELY before starting work
- Update status the INSTANT any task changes state
- NEVER batch multiple status updates together
- Maintain EXACTLY ONE task as "in_progress" at any time
- Complete current task BEFORE starting any new work

**PROGRESS TRANSPARENCY:**
Users can see your todo list in their editor with beautiful checkboxes and progress indicators.
This creates confidence and allows them to track your work in real-time.
</status-discipline>

## When NOT to Use
Skip ONLY for single-step trivial tasks that take under 30 seconds and have no dependencies.

**Remember: When in doubt, CREATE TODOS. It demonstrates professionalism and ensures nothing is forgotten.**`,
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