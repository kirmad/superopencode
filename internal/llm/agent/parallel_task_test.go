package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestParallelTaskParams tests parallel task parameter parsing and validation
func TestParallelTaskParams(t *testing.T) {
	taskTool := NewTaskTool(nil, nil, nil, nil, nil).(*taskTool)

	testCases := []struct {
		name          string
		input         string
		expectedError bool
		expectedTasks int
		expectedWorkers int
		expectedMode  string
	}{
		{
			name: "valid parallel tasks",
			input: `{
				"tasks": [
					{"description": "Task 1", "prompt": "Research prompt", "subagent_type": "research"},
					{"description": "Task 2", "prompt": "Coding prompt", "subagent_type": "coding"}
				],
				"max_workers": 3,
				"aggregate_mode": "json_array"
			}`,
			expectedError: false,
			expectedTasks: 2,
			expectedWorkers: 3,
			expectedMode: "json_array",
		},
		{
			name: "parallel tasks with defaults",
			input: `{
				"tasks": [
					{"description": "Task 1", "prompt": "Research prompt", "subagent_type": "research"},
					{"description": "Task 2", "prompt": "Coding prompt", "subagent_type": "coding"},
					{"description": "Task 3", "prompt": "Analysis prompt", "subagent_type": "analysis"}
				]
			}`,
			expectedError: false,
			expectedTasks: 3,
			expectedWorkers: 3, // min(3, 5)
			expectedMode: "concat",
		},
		{
			name: "single task wrapped as parallel",
			input: `{
				"description": "Single task",
				"prompt": "Single prompt",
				"subagent_type": "general-purpose"
			}`,
			expectedError: false,
			expectedTasks: 1,
			expectedWorkers: 1,
			expectedMode: "concat",
		},
		{
			name: "invalid subagent type in parallel tasks",
			input: `{
				"tasks": [
					{"description": "Task 1", "prompt": "Research prompt", "subagent_type": "invalid-type"}
				]
			}`,
			expectedError: true,
		},
		{
			name: "empty task description",
			input: `{
				"tasks": [
					{"description": "", "prompt": "Research prompt", "subagent_type": "research"}
				]
			}`,
			expectedError: true,
		},
		{
			name: "missing prompt",
			input: `{
				"tasks": [
					{"description": "Task 1", "subagent_type": "research"}
				]
			}`,
			expectedError: true,
		},
		{
			name: "large worker count capped",
			input: `{
				"tasks": [
					{"description": "Task 1", "prompt": "Research prompt", "subagent_type": "research"},
					{"description": "Task 2", "prompt": "Coding prompt", "subagent_type": "coding"}
				],
				"max_workers": 10
			}`,
			expectedError: false,
			expectedTasks: 2,
			expectedWorkers: 10, // Not capped in this implementation
			expectedMode: "concat",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params, err := taskTool.parseParallelTaskParams(tc.input)
			
			if tc.expectedError {
				if err == nil {
					t.Errorf("Expected error for test case '%s', but got none", tc.name)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for test case '%s': %v", tc.name, err)
				return
			}

			if len(params.Tasks) != tc.expectedTasks {
				t.Errorf("Expected %d tasks, got %d", tc.expectedTasks, len(params.Tasks))
			}

			if params.MaxWorkers != tc.expectedWorkers {
				t.Errorf("Expected %d workers, got %d", tc.expectedWorkers, params.MaxWorkers)
			}

			if params.AggregateMode != tc.expectedMode {
				t.Errorf("Expected aggregate mode '%s', got '%s'", tc.expectedMode, params.AggregateMode)
			}
		})
	}
}

// TestTaskResultAggregation tests different result aggregation modes
func TestTaskResultAggregation(t *testing.T) {
	taskTool := NewTaskTool(nil, nil, nil, nil, nil).(*taskTool)

	// Create test results
	results := []TaskResult{
		{
			TaskIndex: 0,
			Result:    "Task 0 completed successfully",
			Error:     nil,
			Metrics: &TaskMetrics{
				TaskID: "task-0",
				Success: true,
				Duration: 2 * time.Second,
			},
		},
		{
			TaskIndex: 1,
			Result:    "",
			Error:     fmt.Errorf("task failed"),
			Metrics: &TaskMetrics{
				TaskID: "task-1",
				Success: false,
				Duration: 1 * time.Second,
			},
		},
		{
			TaskIndex: 2,
			Result:    "Task 2 result with detailed information that should be handled properly",
			Error:     nil,
			Metrics: &TaskMetrics{
				TaskID: "task-2",
				Success: true,
				Duration: 3 * time.Second,
			},
		},
	}

	testCases := []struct {
		name         string
		mode         string
		expectError  bool
		validateFunc func(string) bool
	}{
		{
			name: "concat mode",
			mode: "concat",
			expectError: false,
			validateFunc: func(result string) bool {
				return len(result) > 0 && 
					   strings.Contains(result, "Task 0 Result:") &&
					   strings.Contains(result, "Task 1 FAILED:") &&
					   strings.Contains(result, "Task 2 Result:")
			},
		},
		{
			name: "json_array mode",
			mode: "json_array",
			expectError: false,
			validateFunc: func(result string) bool {
				var jsonArray []map[string]interface{}
				err := json.Unmarshal([]byte(result), &jsonArray)
				return err == nil && len(jsonArray) == 3
			},
		},
		{
			name: "summary mode",
			mode: "summary",
			expectError: false,
			validateFunc: func(result string) bool {
				return len(result) > 0 && 
					   strings.Contains(result, "Total Tasks: 3") &&
					   strings.Contains(result, "Successful: 2") &&
					   strings.Contains(result, "Failed: 1")
			},
		},
		{
			name: "invalid mode",
			mode: "invalid_mode",
			expectError: true,
			validateFunc: func(result string) bool {
				return true // Not used for error cases
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := taskTool.aggregateTaskResults(results, tc.mode)
			
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for mode '%s', but got none", tc.mode)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for mode '%s': %v", tc.mode, err)
				return
			}

			if !tc.validateFunc(result) {
				t.Errorf("Result validation failed for mode '%s'. Result: %s", tc.mode, result)
			}
		})
	}
}

// TestParallelTaskMetrics tests parallel task metrics structure
func TestParallelTaskMetrics(t *testing.T) {
	metrics := &ParallelTaskMetrics{
		ParallelTaskID:    "parallel-test-123",
		TotalTasks:        3,
		SuccessfulTasks:   2,
		FailedTasks:       1,
		MaxWorkers:        3,
		StartTime:         time.Now().Add(-10 * time.Second),
		EndTime:           time.Now(),
		TotalDuration:     10 * time.Second,
		AverageTaskTime:   3 * time.Second,
		MaxTaskTime:       5 * time.Second,
		MinTaskTime:       1 * time.Second,
		TotalCostIncurred: 0.025,
		TotalTokensUsed:   500,
		AggregateMode:     "concat",
		TaskMetrics: []TaskMetrics{
			{TaskID: "task-0", Success: true, Duration: 3 * time.Second},
			{TaskID: "task-1", Success: false, Duration: 1 * time.Second},
			{TaskID: "task-2", Success: true, Duration: 5 * time.Second},
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		t.Errorf("Failed to marshal ParallelTaskMetrics: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaledMetrics ParallelTaskMetrics
	err = json.Unmarshal(jsonData, &unmarshaledMetrics)
	if err != nil {
		t.Errorf("Failed to unmarshal ParallelTaskMetrics: %v", err)
	}

	// Verify key fields
	if unmarshaledMetrics.ParallelTaskID != metrics.ParallelTaskID {
		t.Errorf("Expected ParallelTaskID '%s', got '%s'", metrics.ParallelTaskID, unmarshaledMetrics.ParallelTaskID)
	}

	if unmarshaledMetrics.TotalTasks != metrics.TotalTasks {
		t.Errorf("Expected TotalTasks %d, got %d", metrics.TotalTasks, unmarshaledMetrics.TotalTasks)
	}

	if unmarshaledMetrics.SuccessfulTasks != metrics.SuccessfulTasks {
		t.Errorf("Expected SuccessfulTasks %d, got %d", metrics.SuccessfulTasks, unmarshaledMetrics.SuccessfulTasks)
	}

	if len(unmarshaledMetrics.TaskMetrics) != len(metrics.TaskMetrics) {
		t.Errorf("Expected %d task metrics, got %d", len(metrics.TaskMetrics), len(unmarshaledMetrics.TaskMetrics))
	}
}

// TestParallelTaskLogging tests parallel task metrics logging
func TestParallelTaskLogging(t *testing.T) {
	taskTool := NewTaskTool(nil, nil, nil, nil, nil).(*taskTool)

	metrics := &ParallelTaskMetrics{
		ParallelTaskID:    "test-parallel-123",
		TotalTasks:        5,
		SuccessfulTasks:   4,
		FailedTasks:       1,
		MaxWorkers:        3,
		StartTime:         time.Now().Add(-15 * time.Second),
		EndTime:           time.Now(),
		TotalDuration:     15 * time.Second,
		AverageTaskTime:   3 * time.Second,
		TotalCostIncurred: 0.05,
		TotalTokensUsed:   1000,
		AggregateMode:     "json_array",
	}

	// Test that logParallelTaskMetrics doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("logParallelTaskMetrics should not panic: %v", r)
		}
	}()

	taskTool.logParallelTaskMetrics(metrics)
}

// TestMinFunction tests the min utility function
func TestMinFunction(t *testing.T) {
	testCases := []struct {
		a, b, expected int
	}{
		{1, 2, 1},
		{5, 3, 3},
		{10, 10, 10},
		{0, -1, -1},
		{-5, -3, -5},
	}

	for _, tc := range testCases {
		result := min(tc.a, tc.b)
		if result != tc.expected {
			t.Errorf("min(%d, %d) = %d, expected %d", tc.a, tc.b, result, tc.expected)
		}
	}
}

// TestParallelTaskWorkStructure tests the TaskWork structure
func TestParallelTaskWorkStructure(t *testing.T) {
	work := TaskWork{
		Index: 42,
		Params: TaskParams{
			Description:  "Test work item",
			Prompt:       "Execute test work",
			SubagentType: "research",
		},
	}

	// Verify structure fields
	if work.Index != 42 {
		t.Errorf("Expected Index 42, got %d", work.Index)
	}

	if work.Params.Description != "Test work item" {
		t.Errorf("Expected Description 'Test work item', got '%s'", work.Params.Description)
	}

	if work.Params.SubagentType != "research" {
		t.Errorf("Expected SubagentType 'research', got '%s'", work.Params.SubagentType)
	}
}

// TestParallelTaskResultStructure tests the TaskResult structure
func TestParallelTaskResultStructure(t *testing.T) {
	result := TaskResult{
		TaskIndex: 1,
		Result:    "Task completed successfully",
		Metrics: &TaskMetrics{
			TaskID:   "test-task-1",
			Success:  true,
			Duration: 5 * time.Second,
		},
		Error: nil,
	}

	// Verify structure fields
	if result.TaskIndex != 1 {
		t.Errorf("Expected TaskIndex 1, got %d", result.TaskIndex)
	}

	if result.Result != "Task completed successfully" {
		t.Errorf("Expected Result 'Task completed successfully', got '%s'", result.Result)
	}

	if result.Metrics == nil {
		t.Error("Expected non-nil Metrics")
	} else {
		if result.Metrics.TaskID != "test-task-1" {
			t.Errorf("Expected TaskID 'test-task-1', got '%s'", result.Metrics.TaskID)
		}
		if !result.Metrics.Success {
			t.Error("Expected Success to be true")
		}
	}

	if result.Error != nil {
		t.Errorf("Expected nil Error, got %v", result.Error)
	}
}

// TestParallelTaskConfiguration tests various configuration combinations
func TestParallelTaskConfiguration(t *testing.T) {
	testCases := []struct {
		name            string
		maxWorkers      int
		numTasks        int
		expectedWorkers int
		timeout         string
		aggregateMode   string
	}{
		{
			name:            "default configuration",
			maxWorkers:      0,
			numTasks:        3,
			expectedWorkers: 3, // min(3, 5)
			timeout:         "",
			aggregateMode:   "",
		},
		{
			name:            "limited workers",
			maxWorkers:      2,
			numTasks:        5,
			expectedWorkers: 2,
			timeout:         "5m",
			aggregateMode:   "summary",
		},
		{
			name:            "many workers",
			maxWorkers:      10,
			numTasks:        3,
			expectedWorkers: 10,
			timeout:         "30s",
			aggregateMode:   "json_array",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create parallel params
			params := ParallelTaskParams{
				Tasks:         make([]TaskParams, tc.numTasks),
				MaxWorkers:    tc.maxWorkers,
				AggregateMode: tc.aggregateMode,
				Timeout:       tc.timeout,
			}

			// Fill tasks with valid data
			for i := 0; i < tc.numTasks; i++ {
				params.Tasks[i] = TaskParams{
					Description:  fmt.Sprintf("Task %d", i),
					Prompt:       fmt.Sprintf("Execute task %d", i),
					SubagentType: "general-purpose",
				}
			}

			// Validate using the parsing function
			taskTool := NewTaskTool(nil, nil, nil, nil, nil).(*taskTool)
			input, _ := json.Marshal(params)
			parsedParams, err := taskTool.parseParallelTaskParams(string(input))
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if parsedParams.MaxWorkers != tc.expectedWorkers {
				t.Errorf("Expected %d workers, got %d", tc.expectedWorkers, parsedParams.MaxWorkers)
			}

			expectedMode := tc.aggregateMode
			if expectedMode == "" {
				expectedMode = "concat"
			}
			if parsedParams.AggregateMode != expectedMode {
				t.Errorf("Expected aggregate mode '%s', got '%s'", expectedMode, parsedParams.AggregateMode)
			}
		})
	}
}