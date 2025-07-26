package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kirmad/superopencode/internal/config"
	"github.com/kirmad/superopencode/internal/llm/tools"
	"github.com/kirmad/superopencode/internal/logging"
	"github.com/kirmad/superopencode/internal/lsp"
	"github.com/kirmad/superopencode/internal/message"
	"github.com/kirmad/superopencode/internal/permission"
	"github.com/kirmad/superopencode/internal/session"
	"github.com/kirmad/superopencode/internal/history"
)

type taskTool struct {
	sessions    session.Service
	messages    message.Service
	permissions permission.Service
	history     history.Service
	lspClients  map[string]*lsp.Client
}

const (
	TaskToolName = "Task"
)

type TaskParams struct {
	Description  string `json:"description"`
	Prompt       string `json:"prompt"`
	SubagentType string `json:"subagent_type"`
}

// SubagentCapabilities defines the capabilities of each subagent type
type SubagentCapabilities struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Tools        string   `json:"tools"`
	Performance  string   `json:"performance"`
	OptimizedFor []string `json:"optimized_for"`
}

// TaskMetrics tracks performance metrics for task execution
type TaskMetrics struct {
	TaskID           string        `json:"task_id"`
	SubagentType     string        `json:"subagent_type"`
	StartTime        time.Time     `json:"start_time"`
	EndTime          time.Time     `json:"end_time"`
	Duration         time.Duration `json:"duration"`
	SessionCreationTime time.Duration `json:"session_creation_time"`
	ExecutionTime    time.Duration `json:"execution_time"`
	Success          bool          `json:"success"`
	ErrorType        string        `json:"error_type,omitempty"`
	RetryAttempts    int           `json:"retry_attempts"`
	CostIncurred     float64       `json:"cost_incurred"`
	TokensUsed       int           `json:"tokens_used"`
	ResultLength     int           `json:"result_length"`
}

// Valid subagent types with their capabilities
var validSubagentTypes = map[string]SubagentCapabilities{
	"general-purpose": {
		Name:         "general-purpose",
		Description:  "General-purpose agent for researching complex questions, searching for code, and executing multi-step tasks",
		Tools:        "*",
		Performance:  "balanced",
		OptimizedFor: []string{"research", "code-search", "multi-step-tasks"},
	},
	"research": {
		Name:         "research",
		Description:  "Research-specialized agent optimized for information gathering, analysis, and knowledge synthesis",
		Tools:        "read, grep, glob, sourcegraph, web-search",
		Performance:  "analysis-optimized",
		OptimizedFor: []string{"information-gathering", "analysis", "synthesis"},
	},
	"coding": {
		Name:         "coding",
		Description:  "Code-specialized agent optimized for software development, debugging, and code generation",
		Tools:        "read, write, edit, bash, grep, glob, lsp",
		Performance:  "code-optimized",
		OptimizedFor: []string{"development", "debugging", "code-generation"},
	},
	"analysis": {
		Name:         "analysis",
		Description:  "Analysis-specialized agent optimized for data analysis, pattern recognition, and insights",
		Tools:        "read, grep, glob, bash, data-tools",
		Performance:  "analytical",
		OptimizedFor: []string{"data-analysis", "pattern-recognition", "insights"},
	},
}

// Task execution errors
var (
	ErrInvalidSubagentType = errors.New("invalid subagent type")
	ErrTaskTimeout         = errors.New("task execution timeout")
	ErrTaskCancelled      = errors.New("task execution cancelled")
)

func (t *taskTool) Info() tools.ToolInfo {
	// Build dynamic description with current subagent capabilities
	description := "Launch a new agent to handle complex, multi-step tasks autonomously.\n\nAvailable agent types and the tools they have access to:\n"
	
	for _, capabilities := range validSubagentTypes {
		description += fmt.Sprintf("- %s: %s (Tools: %s, Performance: %s)\n", 
			capabilities.Name, capabilities.Description, capabilities.Tools, capabilities.Performance)
	}
	
	description += "\nWhen using the Task tool, you must specify a subagent_type parameter to select which agent type to use.\n\nWhen to use the Task tool:\n- When you are instructed to execute custom slash commands. Use the Task tool with the slash command invocation as the entire prompt. The slash command can take arguments. For example: Task(description=\"Check the file\", prompt=\"/check-file path/to/file.py\")\n- For complex multi-step operations that benefit from specialized agent capabilities\n- When you need context-isolated execution for sensitive operations\n\nWhen NOT to use the Task tool:\n- If you want to read a specific file path, use the Read or Glob tool instead of the Task tool, to find the match more quickly\n- If you are searching for a specific class definition like \"class Foo\", use the Glob tool instead, to find the match more quickly\n- If you are searching for code within a specific file or set of 2-3 files, use the Read tool instead of the Task tool, to find the match more quickly\n- Other tasks that are not related to the agent descriptions above\n\n\nUsage notes:\n1. Launch multiple agents concurrently whenever possible, to maximize performance; to do that, use a single message with multiple tool uses\n2. When the agent is done, it will return a single message back to you. The result returned by the agent is not visible to the user. To show the user the result, you should send a text message back to the user with a concise summary of the result.\n3. Each agent invocation is stateless. You will not be able to send additional messages to the agent, nor will the agent be able to communicate with you outside of its final report. Therefore, your prompt should contain a highly detailed task description for the agent to perform autonomously and you should specify exactly what information the agent should return back to you in its final and only message to you.\n4. The agent's outputs should generally be trusted\n5. Clearly tell the agent whether you expect it to write code or just to do research (search, file reads, web fetches, etc.), since it is not aware of the user's intent\n6. If the agent description mentions that it should be used proactively, then you should try your best to use it without the user having to ask for it first. Use your judgement.\n7. Choose the appropriate subagent type based on your task: research for information gathering, coding for development work, analysis for data processing, or general-purpose for mixed operations."
	
	return tools.ToolInfo{
		Name:        TaskToolName,
		Description: description,
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"description": map[string]any{
					"type":        "string",
					"description": "A short (3-5 word) description of the task",
				},
				"prompt": map[string]any{
					"type":        "string",
					"description": "The task for the agent to perform",
				},
				"subagent_type": map[string]any{
					"type":        "string",
					"description": "The type of specialized agent to use for this task",
				},
			},
			"required": []string{"description", "prompt", "subagent_type"},
		},
		Required: []string{"description", "prompt", "subagent_type"},
	}
}

func (t *taskTool) Run(ctx context.Context, call tools.ToolCall) (tools.ToolResponse, error) {
	// Parse and validate parameters
	params, err := t.parseAndValidateParams(call.Input)
	if err != nil {
		return tools.NewTextErrorResponse(err.Error()), nil
	}

	// Get context values
	sessionID, messageID := tools.GetContextValues(ctx)
	if sessionID == "" || messageID == "" {
		return tools.ToolResponse{}, fmt.Errorf("session_id and message_id are required")
	}

	// Initialize performance metrics
	metrics := &TaskMetrics{
		TaskID:       call.ID,
		SubagentType: params.SubagentType,
		StartTime:    time.Now(),
	}

	// Log task initiation with subagent capabilities
	capabilities := validSubagentTypes[params.SubagentType]
	logging.InfoPersist(fmt.Sprintf("Task tool initiated: %s (session: %s, call: %s, subagent: %s, tools: %s)", 
		params.Description, sessionID, call.ID, capabilities.Name, capabilities.Tools))

	// Create task session with error recovery and timing
	sessionStartTime := time.Now()
	taskSession, err := t.createTaskSessionWithRetry(ctx, call.ID, sessionID, params.Description, 3)
	metrics.SessionCreationTime = time.Since(sessionStartTime)
	if err != nil {
		metrics.Success = false
		metrics.ErrorType = "session_creation_failed"
		metrics.EndTime = time.Now()
		metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
		t.logTaskMetrics(metrics)
		logging.ErrorPersist(fmt.Sprintf("Failed to create task session after retries: %v", err))
		return tools.NewTextErrorResponse(fmt.Sprintf("failed to create task session: %v", err)), err
	}

	// Create task agent with subagent-specific tools
	taskAgent, err := t.createTaskAgentWithSubagentType(params.SubagentType)
	if err != nil {
		metrics.Success = false
		metrics.ErrorType = "agent_creation_failed"
		metrics.EndTime = time.Now()
		metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
		t.logTaskMetrics(metrics)
		logging.ErrorPersist(fmt.Sprintf("Failed to create task agent: %v", err))
		return tools.NewTextErrorResponse(fmt.Sprintf("failed to create task agent: %v", err)), err
	}

	// Execute task with timeout and performance monitoring
	executionStartTime := time.Now()
	result, err := t.executeTaskWithTimeout(ctx, taskAgent, taskSession.ID, params, 30*time.Minute)
	metrics.ExecutionTime = time.Since(executionStartTime)
	
	if err != nil {
		metrics.Success = false
		metrics.EndTime = time.Now()
		metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
		
		// Handle different error types with metrics
		if errors.Is(err, ErrTaskTimeout) {
			metrics.ErrorType = "timeout"
			t.logTaskMetrics(metrics)
			logging.InfoPersist(fmt.Sprintf("Task timeout for session: %s", taskSession.ID))
			return tools.NewTextErrorResponse("task execution timed out"), nil
		}
		if errors.Is(err, ErrTaskCancelled) {
			metrics.ErrorType = "cancelled"
			t.logTaskMetrics(metrics)
			logging.InfoPersist(fmt.Sprintf("Task cancelled for session: %s", taskSession.ID))
			return tools.NewTextErrorResponse("task execution was cancelled"), nil
		}
		if errors.Is(err, context.Canceled) {
			metrics.ErrorType = "context_cancelled"
			t.logTaskMetrics(metrics)
			logging.InfoPersist(fmt.Sprintf("Task context cancelled for session: %s", taskSession.ID))
			return tools.NewTextErrorResponse("task execution was cancelled"), nil
		}
		
		metrics.ErrorType = "execution_failed"
		t.logTaskMetrics(metrics)
		logging.ErrorPersist(fmt.Sprintf("Task execution failed: %v", err))
		return tools.NewTextErrorResponse(fmt.Sprintf("task execution failed: %v", err)), err
	}

	// Complete metrics collection
	metrics.Success = true
	metrics.EndTime = time.Now()
	metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
	metrics.ResultLength = len(result)

	// Aggregate cost from child session to parent session with metrics
	if err := t.aggregateCostWithMetrics(ctx, sessionID, taskSession.ID, metrics); err != nil {
		logging.ErrorPersist(fmt.Sprintf("Failed to aggregate cost: %v", err))
		// Don't fail the task for cost aggregation issues
	}

	// Log performance metrics
	t.logTaskMetrics(metrics)
	
	// Log successful completion with performance summary
	logging.InfoPersist(fmt.Sprintf("Task completed successfully: %s (duration: %v, subagent: %s, cost: $%.6f)", 
		params.Description, metrics.Duration, params.SubagentType, metrics.CostIncurred))

	return tools.NewTextResponse(result), nil
}

// parseAndValidateParams parses and validates task parameters
func (t *taskTool) parseAndValidateParams(input string) (TaskParams, error) {
	var params TaskParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return params, fmt.Errorf("error parsing parameters: %w", err)
	}

	// Validate required parameters
	if strings.TrimSpace(params.Description) == "" {
		return params, fmt.Errorf("description is required")
	}
	if strings.TrimSpace(params.Prompt) == "" {
		return params, fmt.Errorf("prompt is required")
	}
	if strings.TrimSpace(params.SubagentType) == "" {
		return params, fmt.Errorf("subagent_type is required")
	}

	// Validate subagent type
	if _, exists := validSubagentTypes[params.SubagentType]; !exists {
		return params, fmt.Errorf("%w: %s (valid types: %v)", ErrInvalidSubagentType, params.SubagentType, getValidSubagentTypes())
	}

	return params, nil
}

// createTaskSessionWithRetry creates a task session with retry logic
func (t *taskTool) createTaskSessionWithRetry(ctx context.Context, toolCallID, parentSessionID, description string, maxRetries int) (session.Session, error) {
	var lastErr error
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		taskSession, err := t.sessions.CreateTaskSession(ctx, toolCallID, parentSessionID, fmt.Sprintf("Task: %s", description))
		if err == nil {
			return taskSession, nil
		}
		
		lastErr = err
		if attempt < maxRetries {
			// Exponential backoff
			backoff := time.Duration(attempt*attempt) * 100 * time.Millisecond
			logging.InfoPersist(fmt.Sprintf("Task session creation attempt %d failed, retrying in %v: %v", attempt, backoff, err))
			
			select {
			case <-ctx.Done():
				return session.Session{}, ctx.Err()
			case <-time.After(backoff):
				continue
			}
		}
	}
	
	return session.Session{}, fmt.Errorf("failed to create task session after %d attempts: %w", maxRetries, lastErr)
}

// createTaskAgentWithSubagentType creates a task agent with subagent-specific tools
func (t *taskTool) createTaskAgentWithSubagentType(subagentType string) (Service, error) {
	capabilities := validSubagentTypes[subagentType]
	
	// Select tools based on subagent type
	var agentTools []tools.BaseTool
	switch subagentType {
	case "general-purpose":
		// Full tool access for general-purpose agents
		agentTools = CoderAgentTools(t.permissions, t.sessions, t.messages, t.history, t.lspClients)
	case "research":
		// Research-optimized tools: read, grep, glob, sourcegraph, web-search
		agentTools = ResearchAgentTools(t.permissions, t.sessions, t.messages, t.history, t.lspClients)
	case "coding":
		// Code-optimized tools: read, write, edit, bash, grep, glob, lsp
		agentTools = CodingAgentTools(t.permissions, t.sessions, t.messages, t.history, t.lspClients)
	case "analysis":
		// Analysis-optimized tools: read, grep, glob, bash, data-tools
		agentTools = AnalysisAgentTools(t.permissions, t.sessions, t.messages, t.history, t.lspClients)
	default:
		// Fallback to general-purpose
		agentTools = CoderAgentTools(t.permissions, t.sessions, t.messages, t.history, t.lspClients)
	}
	
	logging.InfoPersist(fmt.Sprintf("Creating %s subagent with %d tools (optimized for: %v)", 
		capabilities.Name, len(agentTools), capabilities.OptimizedFor))
	
	return NewAgent(
		config.AgentTask,
		t.sessions,
		t.messages,
		agentTools,
	)
}

// executeTaskWithTimeout executes the task with timeout and cancellation support
func (t *taskTool) executeTaskWithTimeout(ctx context.Context, agent Service, sessionID string, params TaskParams, timeout time.Duration) (string, error) {
	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute task
	done, err := agent.Run(timeoutCtx, sessionID, params.Prompt)
	if err != nil {
		return "", fmt.Errorf("failed to start task: %w", err)
	}

	// Wait for completion with timeout handling
	select {
	case result := <-done:
		if result.Error != nil {
			// Check for specific error types
			if errors.Is(result.Error, ErrRequestCancelled) {
				return "", ErrTaskCancelled
			}
			if errors.Is(result.Error, context.Canceled) {
				return "", ErrTaskCancelled
			}
			if errors.Is(result.Error, context.DeadlineExceeded) {
				return "", ErrTaskTimeout
			}
			return "", fmt.Errorf("task execution error: %w", result.Error)
		}
		return result.Message.Content().String(), nil
		
	case <-timeoutCtx.Done():
		// Cancel the agent if timeout occurs
		agent.Cancel(sessionID)
		if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
			return "", ErrTaskTimeout
		}
		return "", ErrTaskCancelled
	}
}

// aggregateCostWithMetrics aggregates cost from child session to parent session with metrics
func (t *taskTool) aggregateCostWithMetrics(ctx context.Context, parentSessionID, childSessionID string, metrics *TaskMetrics) error {
	// Get child session cost
	childSession, err := t.sessions.Get(ctx, childSessionID)
	if err != nil {
		return fmt.Errorf("failed to get child session: %w", err)
	}

	if childSession.Cost == 0 {
		return nil // No cost to aggregate
	}

	// Get parent session
	parentSession, err := t.sessions.Get(ctx, parentSessionID)
	if err != nil {
		return fmt.Errorf("failed to get parent session: %w", err)
	}

	// Add child cost to parent
	parentSession.Cost += childSession.Cost
	parentSession.PromptTokens += childSession.PromptTokens
	parentSession.CompletionTokens += childSession.CompletionTokens

	// Save updated parent session
	_, err = t.sessions.Save(ctx, parentSession)
	if err != nil {
		return fmt.Errorf("failed to save parent session: %w", err)
	}

	// Update metrics with cost information
	metrics.CostIncurred = childSession.Cost
	metrics.TokensUsed = int(childSession.PromptTokens + childSession.CompletionTokens)
	
	logging.InfoPersist(fmt.Sprintf("Cost aggregated from task session %s to parent %s: $%.6f (tokens: %d)", 
		childSessionID, parentSessionID, childSession.Cost, metrics.TokensUsed))
	return nil
}

// getValidSubagentTypes returns a list of valid subagent types
func getValidSubagentTypes() []string {
	var types []string
	for t := range validSubagentTypes {
		types = append(types, t)
	}
	return types
}

// logTaskMetrics logs performance metrics for task execution
func (t *taskTool) logTaskMetrics(metrics *TaskMetrics) {
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		logging.ErrorPersist(fmt.Sprintf("Failed to marshal task metrics: %v", err))
		return
	}
	
	// Log structured metrics for analysis
	logging.InfoPersist(fmt.Sprintf("Task performance metrics: %s", string(metricsJSON)))
	
	// Log human-readable summary
	status := "SUCCESS"
	if !metrics.Success {
		status = fmt.Sprintf("FAILED (%s)", metrics.ErrorType)
	}
	
	logging.InfoPersist(fmt.Sprintf("Task %s [%s]: %s | Duration: %v | Session: %v | Execution: %v | Cost: $%.6f | Tokens: %d",
		metrics.TaskID, metrics.SubagentType, status, metrics.Duration, 
		metrics.SessionCreationTime, metrics.ExecutionTime, metrics.CostIncurred, metrics.TokensUsed))
}

// GetSubagentCapabilities returns capabilities for a given subagent type
func (t *taskTool) GetSubagentCapabilities(subagentType string) (SubagentCapabilities, bool) {
	capabilities, exists := validSubagentTypes[subagentType]
	return capabilities, exists
}

// GetAllSubagentCapabilities returns all available subagent capabilities
func (t *taskTool) GetAllSubagentCapabilities() map[string]SubagentCapabilities {
	return validSubagentTypes
}

func NewTaskTool(sessions session.Service, messages message.Service, permissions permission.Service, history history.Service, lspClients map[string]*lsp.Client) tools.BaseTool {
	return &taskTool{
		sessions:    sessions,
		messages:    messages,
		permissions: permissions,
		history:     history,
		lspClients:  lspClients,
	}
}