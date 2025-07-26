package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
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

// ParallelTaskParams extends TaskParams for parallel execution
type ParallelTaskParams struct {
	Tasks         []TaskParams `json:"tasks"`
	MaxWorkers    int          `json:"max_workers,omitempty"`    // Optional: limit concurrent workers
	AggregateMode string       `json:"aggregate_mode,omitempty"` // "concat", "json_array", "summary"
	Timeout       string       `json:"timeout,omitempty"`        // Optional: overall timeout (e.g., "5m")
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

// ParallelTaskMetrics tracks metrics for parallel task execution
type ParallelTaskMetrics struct {
	ParallelTaskID     string        `json:"parallel_task_id"`
	TotalTasks         int           `json:"total_tasks"`
	SuccessfulTasks    int           `json:"successful_tasks"`
	FailedTasks        int           `json:"failed_tasks"`
	MaxWorkers         int           `json:"max_workers"`
	StartTime          time.Time     `json:"start_time"`
	EndTime            time.Time     `json:"end_time"`
	TotalDuration      time.Duration `json:"total_duration"`
	AverageTaskTime    time.Duration `json:"average_task_time"`
	MaxTaskTime        time.Duration `json:"max_task_time"`
	MinTaskTime        time.Duration `json:"min_task_time"`
	TotalCostIncurred  float64       `json:"total_cost_incurred"`
	TotalTokensUsed    int           `json:"total_tokens_used"`
	AggregateMode      string        `json:"aggregate_mode"`
	TaskMetrics        []TaskMetrics `json:"task_metrics"`
}

// TaskResult represents the result of a single task execution
type TaskResult struct {
	TaskIndex int
	Result    string
	Metrics   *TaskMetrics
	Error     error
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
	
	description += "\n## Single Task Usage\nWhen using the Task tool for a single task, you must specify a subagent_type parameter to select which agent type to use.\n\n## Parallel Task Usage\nFor multiple tasks that can be executed concurrently, provide a 'tasks' array instead of single task parameters:\n```json\n{\n  \"tasks\": [\n    {\"description\": \"Task 1\", \"prompt\": \"...\", \"subagent_type\": \"research\"},\n    {\"description\": \"Task 2\", \"prompt\": \"...\", \"subagent_type\": \"coding\"}\n  ],\n  \"max_workers\": 3,\n  \"aggregate_mode\": \"concat\",\n  \"timeout\": \"10m\"\n}\n```\n\nParallel execution options:\n- max_workers: Limit concurrent workers (default: min(tasks, 5))\n- aggregate_mode: \"concat\", \"json_array\", or \"summary\" (default: \"concat\")\n- timeout: Overall timeout (e.g., \"5m\", \"30s\") (default: \"30m\")\n\nWhen to use the Task tool:\n- When you are instructed to execute custom slash commands. Use the Task tool with the slash command invocation as the entire prompt. The slash command can take arguments. For example: Task(description=\"Check the file\", prompt=\"/check-file path/to/file.py\")\n- For complex multi-step operations that benefit from specialized agent capabilities\n- When you need context-isolated execution for sensitive operations\n- For batch operations that can benefit from parallel execution\n\nWhen NOT to use the Task tool:\n- If you want to read a specific file path, use the Read or Glob tool instead of the Task tool, to find the match more quickly\n- If you are searching for a specific class definition like \"class Foo\", use the Glob tool instead, to find the match more quickly\n- If you are searching for code within a specific file or set of 2-3 files, use the Read tool instead of the Task tool, to find the match more quickly\n- Other tasks that are not related to the agent descriptions above\n\n\nUsage notes:\n1. Launch multiple agents concurrently whenever possible, to maximize performance; to do that, use a single message with multiple tool uses or use parallel task execution\n2. When the agent is done, it will return a single message back to you. The result returned by the agent is not visible to the user. To show the user the result, you should send a text message back to the user with a concise summary of the result.\n3. Each agent invocation is stateless. You will not be able to send additional messages to the agent, nor will the agent be able to communicate with you outside of its final report. Therefore, your prompt should contain a highly detailed task description for the agent to perform autonomously and you should specify exactly what information the agent should return back to you in its final and only message to you.\n4. The agent's outputs should generally be trusted\n5. Clearly tell the agent whether you expect it to write code or just to do research (search, file reads, web fetches, etc.), since it is not aware of the user's intent\n6. If the agent description mentions that it should be used proactively, then you should try your best to use it without the user having to ask for it first. Use your judgement.\n7. Choose the appropriate subagent type based on your task: research for information gathering, coding for development work, analysis for data processing, or general-purpose for mixed operations.\n8. Use parallel execution for batch operations, bulk processing, or when multiple independent tasks can be performed simultaneously."
	
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
	// Try to parse as parallel task first, then fall back to single task
	parallelParams, err := t.parseParallelTaskParams(call.Input)
	if err == nil && len(parallelParams.Tasks) > 1 {
		return t.runParallelTasks(ctx, call, parallelParams)
	}
	
	// Parse and validate single task parameters
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

// parseParallelTaskParams parses and validates parallel task parameters
func (t *taskTool) parseParallelTaskParams(input string) (ParallelTaskParams, error) {
	var parallelParams ParallelTaskParams
	if err := json.Unmarshal([]byte(input), &parallelParams); err != nil {
		return parallelParams, fmt.Errorf("error parsing parallel parameters: %w", err)
	}

	// If no tasks field, try parsing as single task and wrap it
	if len(parallelParams.Tasks) == 0 {
		var singleTask TaskParams
		if err := json.Unmarshal([]byte(input), &singleTask); err != nil {
			return parallelParams, fmt.Errorf("error parsing as single task: %w", err)
		}
		parallelParams.Tasks = []TaskParams{singleTask}
	}

	// Validate each task
	for i, task := range parallelParams.Tasks {
		if strings.TrimSpace(task.Description) == "" {
			return parallelParams, fmt.Errorf("task %d: description is required", i)
		}
		if strings.TrimSpace(task.Prompt) == "" {
			return parallelParams, fmt.Errorf("task %d: prompt is required", i)
		}
		if strings.TrimSpace(task.SubagentType) == "" {
			return parallelParams, fmt.Errorf("task %d: subagent_type is required", i)
		}
		
		// Validate subagent type
		if _, exists := validSubagentTypes[task.SubagentType]; !exists {
			return parallelParams, fmt.Errorf("task %d: %w: %s (valid types: %v)", 
				i, ErrInvalidSubagentType, task.SubagentType, getValidSubagentTypes())
		}
	}

	// Set defaults
	if parallelParams.MaxWorkers <= 0 {
		parallelParams.MaxWorkers = min(len(parallelParams.Tasks), 5) // Default to 5 max workers
	}
	if parallelParams.AggregateMode == "" {
		parallelParams.AggregateMode = "concat" // Default aggregation mode
	}

	return parallelParams, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// runParallelTasks executes multiple tasks in parallel
func (t *taskTool) runParallelTasks(ctx context.Context, call tools.ToolCall, parallelParams ParallelTaskParams) (tools.ToolResponse, error) {
	// Get context values
	sessionID, messageID := tools.GetContextValues(ctx)
	if sessionID == "" || messageID == "" {
		return tools.ToolResponse{}, fmt.Errorf("session_id and message_id are required")
	}

	// Parse timeout if provided
	var totalTimeout time.Duration = 30 * time.Minute // Default timeout
	if parallelParams.Timeout != "" {
		if duration, err := time.ParseDuration(parallelParams.Timeout); err == nil {
			totalTimeout = duration
		}
	}

	// Initialize parallel metrics
	parallelMetrics := &ParallelTaskMetrics{
		ParallelTaskID: call.ID,
		TotalTasks:     len(parallelParams.Tasks),
		MaxWorkers:     parallelParams.MaxWorkers,
		StartTime:      time.Now(),
		AggregateMode:  parallelParams.AggregateMode,
		TaskMetrics:    make([]TaskMetrics, 0, len(parallelParams.Tasks)),
	}

	logging.InfoPersist(fmt.Sprintf("Parallel task execution started: %d tasks, %d max workers (session: %s, call: %s)", 
		len(parallelParams.Tasks), parallelParams.MaxWorkers, sessionID, call.ID))

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, totalTimeout)
	defer cancel()

	// Execute tasks in parallel
	results, err := t.executeTasksInParallel(timeoutCtx, sessionID, parallelParams, parallelMetrics)
	if err != nil {
		parallelMetrics.EndTime = time.Now()
		parallelMetrics.TotalDuration = parallelMetrics.EndTime.Sub(parallelMetrics.StartTime)
		t.logParallelTaskMetrics(parallelMetrics)
		logging.ErrorPersist(fmt.Sprintf("Parallel task execution failed: %v", err))
		return tools.NewTextErrorResponse(fmt.Sprintf("parallel task execution failed: %v", err)), err
	}

	// Complete metrics collection
	parallelMetrics.EndTime = time.Now()
	parallelMetrics.TotalDuration = parallelMetrics.EndTime.Sub(parallelMetrics.StartTime)
	
	// Calculate aggregate metrics
	if len(parallelMetrics.TaskMetrics) > 0 {
		var totalTaskTime time.Duration
		for _, taskMetric := range parallelMetrics.TaskMetrics {
			totalTaskTime += taskMetric.Duration
			if parallelMetrics.MaxTaskTime == 0 || taskMetric.Duration > parallelMetrics.MaxTaskTime {
				parallelMetrics.MaxTaskTime = taskMetric.Duration
			}
			if parallelMetrics.MinTaskTime == 0 || taskMetric.Duration < parallelMetrics.MinTaskTime {
				parallelMetrics.MinTaskTime = taskMetric.Duration
			}
		}
		parallelMetrics.AverageTaskTime = totalTaskTime / time.Duration(len(parallelMetrics.TaskMetrics))
	}

	// Count successful and failed tasks
	for _, result := range results {
		if result.Error == nil {
			parallelMetrics.SuccessfulTasks++
		} else {
			parallelMetrics.FailedTasks++
		}
	}

	// Aggregate results based on mode
	aggregatedResult, err := t.aggregateTaskResults(results, parallelParams.AggregateMode)
	if err != nil {
		t.logParallelTaskMetrics(parallelMetrics)
		logging.ErrorPersist(fmt.Sprintf("Failed to aggregate results: %v", err))
		return tools.NewTextErrorResponse(fmt.Sprintf("failed to aggregate results: %v", err)), err
	}

	// Log performance metrics
	t.logParallelTaskMetrics(parallelMetrics)
	
	// Log successful completion with performance summary
	logging.InfoPersist(fmt.Sprintf("Parallel tasks completed: %d/%d successful (duration: %v, avg task time: %v, total cost: $%.6f)", 
		parallelMetrics.SuccessfulTasks, parallelMetrics.TotalTasks, parallelMetrics.TotalDuration, 
		parallelMetrics.AverageTaskTime, parallelMetrics.TotalCostIncurred))

	return tools.NewTextResponse(aggregatedResult), nil
}

// executeTasksInParallel runs multiple tasks concurrently using a worker pool
func (t *taskTool) executeTasksInParallel(ctx context.Context, sessionID string, parallelParams ParallelTaskParams, parallelMetrics *ParallelTaskMetrics) ([]TaskResult, error) {
	// Create channels for work distribution
	taskChan := make(chan TaskWork, len(parallelParams.Tasks))
	resultChan := make(chan TaskResult, len(parallelParams.Tasks))
	
	// Create task work items
	for i, task := range parallelParams.Tasks {
		taskChan <- TaskWork{
			Index:  i,
			Params: task,
		}
	}
	close(taskChan)

	// Start worker goroutines
	var workerWg sync.WaitGroup
	for i := 0; i < parallelParams.MaxWorkers; i++ {
		workerWg.Add(1)
		go t.taskWorker(ctx, sessionID, taskChan, resultChan, &workerWg)
	}

	// Collect results in a separate goroutine
	results := make([]TaskResult, len(parallelParams.Tasks))
	var resultWg sync.WaitGroup
	resultWg.Add(1)
	go func() {
		defer resultWg.Done()
		for i := 0; i < len(parallelParams.Tasks); i++ {
			select {
			case result := <-resultChan:
				results[result.TaskIndex] = result
				if result.Metrics != nil {
					parallelMetrics.TaskMetrics = append(parallelMetrics.TaskMetrics, *result.Metrics)
					parallelMetrics.TotalCostIncurred += result.Metrics.CostIncurred
					parallelMetrics.TotalTokensUsed += result.Metrics.TokensUsed
				}
			case <-ctx.Done():
				// Handle cancellation
				for j := i; j < len(parallelParams.Tasks); j++ {
					results[j] = TaskResult{
						TaskIndex: j,
						Result:    "",
						Error:     fmt.Errorf("task cancelled due to timeout or cancellation"),
					}
				}
				return
			}
		}
	}()

	// Wait for all workers to complete
	workerWg.Wait()
	close(resultChan)
	
	// Wait for all results to be collected
	resultWg.Wait()

	return results, nil
}

// TaskWork represents a task to be processed by a worker
type TaskWork struct {
	Index  int
	Params TaskParams
}

// taskWorker processes tasks from the task channel
func (t *taskTool) taskWorker(ctx context.Context, parentSessionID string, taskChan <-chan TaskWork, resultChan chan<- TaskResult, wg *sync.WaitGroup) {
	defer wg.Done()
	
	for taskWork := range taskChan {
		select {
		case <-ctx.Done():
			// Send cancellation result
			resultChan <- TaskResult{
				TaskIndex: taskWork.Index,
				Result:    "",
				Error:     ctx.Err(),
			}
			return
		default:
			// Execute the task
			result := t.executeSingleTaskInParallel(ctx, parentSessionID, taskWork)
			resultChan <- result
		}
	}
}

// executeSingleTaskInParallel executes a single task as part of parallel execution
func (t *taskTool) executeSingleTaskInParallel(ctx context.Context, parentSessionID string, taskWork TaskWork) TaskResult {
	// Generate unique task ID
	taskID := fmt.Sprintf("%s-task-%d", parentSessionID, taskWork.Index)
	
	// Initialize task metrics
	metrics := &TaskMetrics{
		TaskID:       taskID,
		SubagentType: taskWork.Params.SubagentType,
		StartTime:    time.Now(),
	}

	// Create task session with error recovery
	sessionStartTime := time.Now()
	taskSession, err := t.createTaskSessionWithRetry(ctx, taskID, parentSessionID, taskWork.Params.Description, 3)
	metrics.SessionCreationTime = time.Since(sessionStartTime)
	if err != nil {
		metrics.Success = false
		metrics.ErrorType = "session_creation_failed"
		metrics.EndTime = time.Now()
		metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
		return TaskResult{
			TaskIndex: taskWork.Index,
			Result:    "",
			Metrics:   metrics,
			Error:     fmt.Errorf("failed to create task session: %w", err),
		}
	}

	// Create task agent with subagent-specific tools
	taskAgent, err := t.createTaskAgentWithSubagentType(taskWork.Params.SubagentType)
	if err != nil {
		metrics.Success = false
		metrics.ErrorType = "agent_creation_failed"
		metrics.EndTime = time.Now()
		metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
		return TaskResult{
			TaskIndex: taskWork.Index,
			Result:    "",
			Metrics:   metrics,
			Error:     fmt.Errorf("failed to create task agent: %w", err),
		}
	}

	// Execute task with timeout (use a shorter timeout for individual tasks in parallel)
	executionStartTime := time.Now()
	result, err := t.executeTaskWithTimeout(ctx, taskAgent, taskSession.ID, taskWork.Params, 10*time.Minute)
	metrics.ExecutionTime = time.Since(executionStartTime)
	
	if err != nil {
		metrics.Success = false
		metrics.EndTime = time.Now()
		metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
		
		// Handle different error types
		if errors.Is(err, ErrTaskTimeout) {
			metrics.ErrorType = "timeout"
		} else if errors.Is(err, ErrTaskCancelled) || errors.Is(err, context.Canceled) {
			metrics.ErrorType = "cancelled"
		} else {
			metrics.ErrorType = "execution_failed"
		}
		
		return TaskResult{
			TaskIndex: taskWork.Index,
			Result:    "",
			Metrics:   metrics,
			Error:     err,
		}
	}

	// Complete metrics collection
	metrics.Success = true
	metrics.EndTime = time.Now()
	metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
	metrics.ResultLength = len(result)

	// Aggregate cost from child session to parent session
	if err := t.aggregateCostWithMetrics(ctx, parentSessionID, taskSession.ID, metrics); err != nil {
		logging.ErrorPersist(fmt.Sprintf("Failed to aggregate cost for task %d: %v", taskWork.Index, err))
		// Don't fail the task for cost aggregation issues
	}

	return TaskResult{
		TaskIndex: taskWork.Index,
		Result:    result,
		Metrics:   metrics,
		Error:     nil,
	}
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

// aggregateTaskResults combines results from multiple tasks based on the specified mode
func (t *taskTool) aggregateTaskResults(results []TaskResult, mode string) (string, error) {
	switch mode {
	case "concat":
		var aggregated strings.Builder
		for i, result := range results {
			if result.Error != nil {
				aggregated.WriteString(fmt.Sprintf("Task %d FAILED: %v\n", i, result.Error))
			} else {
				aggregated.WriteString(fmt.Sprintf("Task %d Result:\n%s\n\n", i, result.Result))
			}
		}
		return aggregated.String(), nil
		
	case "json_array":
		var resultArray []map[string]interface{}
		for i, result := range results {
			resultMap := map[string]interface{}{
				"task_index": i,
				"success":    result.Error == nil,
			}
			if result.Error != nil {
				resultMap["error"] = result.Error.Error()
			} else {
				resultMap["result"] = result.Result
			}
			if result.Metrics != nil {
				resultMap["metrics"] = result.Metrics
			}
			resultArray = append(resultArray, resultMap)
		}
		
		jsonData, err := json.MarshalIndent(resultArray, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal results to JSON: %w", err)
		}
		return string(jsonData), nil
		
	case "summary":
		var summary strings.Builder
		successCount := 0
		for _, result := range results {
			if result.Error == nil {
				successCount++
			}
		}
		
		summary.WriteString(fmt.Sprintf("Parallel Task Execution Summary:\n"))
		summary.WriteString(fmt.Sprintf("Total Tasks: %d\n", len(results)))
		summary.WriteString(fmt.Sprintf("Successful: %d\n", successCount))
		summary.WriteString(fmt.Sprintf("Failed: %d\n", len(results)-successCount))
		summary.WriteString(fmt.Sprintf("\nDetailed Results:\n"))
		
		for i, result := range results {
			if result.Error != nil {
				summary.WriteString(fmt.Sprintf("- Task %d: FAILED (%v)\n", i, result.Error))
			} else {
				// Truncate very long results for summary
				resultText := result.Result
				if len(resultText) > 200 {
					resultText = resultText[:200] + "... (truncated)"
				}
				summary.WriteString(fmt.Sprintf("- Task %d: SUCCESS (%s)\n", i, resultText))
			}
		}
		return summary.String(), nil
		
	default:
		return "", fmt.Errorf("unsupported aggregation mode: %s", mode)
	}
}

// logParallelTaskMetrics logs performance metrics for parallel task execution
func (t *taskTool) logParallelTaskMetrics(metrics *ParallelTaskMetrics) {
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		logging.ErrorPersist(fmt.Sprintf("Failed to marshal parallel task metrics: %v", err))
		return
	}
	
	// Log structured metrics for analysis
	logging.InfoPersist(fmt.Sprintf("Parallel task performance metrics: %s", string(metricsJSON)))
	
	// Log human-readable summary
	logging.InfoPersist(fmt.Sprintf("Parallel Task %s: %d/%d successful | Duration: %v | Avg: %v | Workers: %d | Total Cost: $%.6f | Total Tokens: %d",
		metrics.ParallelTaskID, metrics.SuccessfulTasks, metrics.TotalTasks, metrics.TotalDuration, 
		metrics.AverageTaskTime, metrics.MaxWorkers, metrics.TotalCostIncurred, metrics.TotalTokensUsed))
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