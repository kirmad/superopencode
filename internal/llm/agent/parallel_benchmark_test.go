package agent

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"
)

// BenchmarkParallelTaskParamsParsing benchmarks parallel task parameter parsing
func BenchmarkParallelTaskParamsParsing(b *testing.B) {
	taskTool := NewTaskTool(nil, nil, nil, nil, nil).(*taskTool)
	input := `{
		"tasks": [
			{"description": "Task 1", "prompt": "Research prompt", "subagent_type": "research"},
			{"description": "Task 2", "prompt": "Coding prompt", "subagent_type": "coding"},
			{"description": "Task 3", "prompt": "Analysis prompt", "subagent_type": "analysis"}
		],
		"max_workers": 3,
		"aggregate_mode": "concat"
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := taskTool.parseParallelTaskParams(input)
		if err != nil {
			b.Fatalf("Parameter parsing failed: %v", err)
		}
	}
}

// BenchmarkTaskResultsConcatAggregation benchmarks concat aggregation mode
func BenchmarkTaskResultsConcatAggregation(b *testing.B) {
	taskTool := NewTaskTool(nil, nil, nil, nil, nil).(*taskTool)
	
	// Create benchmark results with varying sizes
	results := make([]TaskResult, 10)
	for i := 0; i < 10; i++ {
		results[i] = TaskResult{
			TaskIndex: i,
			Result:    "Task " + strconv.Itoa(i) + " completed with detailed result information that simulates real task output",
			Error:     nil,
			Metrics: &TaskMetrics{
				TaskID:   "benchmark-task-" + strconv.Itoa(i),
				Success:  true,
				Duration: time.Duration(i+1) * time.Second,
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := taskTool.aggregateTaskResults(results, "concat")
		if err != nil {
			b.Fatalf("Result aggregation failed: %v", err)
		}
	}
}

// BenchmarkTaskResultsJSONAggregation benchmarks JSON array aggregation mode
func BenchmarkTaskResultsJSONAggregation(b *testing.B) {
	taskTool := NewTaskTool(nil, nil, nil, nil, nil).(*taskTool)
	
	// Create benchmark results
	results := make([]TaskResult, 10)
	for i := 0; i < 10; i++ {
		results[i] = TaskResult{
			TaskIndex: i,
			Result:    "Task " + strconv.Itoa(i) + " completed successfully",
			Error:     nil,
			Metrics: &TaskMetrics{
				TaskID:   "benchmark-task-" + strconv.Itoa(i),
				Success:  true,
				Duration: time.Duration(i+1) * time.Second,
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := taskTool.aggregateTaskResults(results, "json_array")
		if err != nil {
			b.Fatalf("JSON aggregation failed: %v", err)
		}
	}
}

// BenchmarkParallelTaskMetricsSerialization benchmarks metrics serialization
func BenchmarkParallelTaskMetricsSerialization(b *testing.B) {
	metrics := &ParallelTaskMetrics{
		ParallelTaskID:    "benchmark-parallel-123",
		TotalTasks:        10,
		SuccessfulTasks:   8,
		FailedTasks:       2,
		MaxWorkers:        5,
		StartTime:         time.Now().Add(-30 * time.Second),
		EndTime:           time.Now(),
		TotalDuration:     30 * time.Second,
		AverageTaskTime:   3 * time.Second,
		MaxTaskTime:       8 * time.Second,
		MinTaskTime:       1 * time.Second,
		TotalCostIncurred: 0.15,
		TotalTokensUsed:   2500,
		AggregateMode:     "concat",
		TaskMetrics: make([]TaskMetrics, 10),
	}

	// Fill task metrics
	for i := 0; i < 10; i++ {
		metrics.TaskMetrics[i] = TaskMetrics{
			TaskID:       "benchmark-task-" + strconv.Itoa(i),
			SubagentType: "general-purpose",
			Success:      i%4 != 0, // Some failures
			Duration:     time.Duration(i+1) * time.Second,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(metrics)
		if err != nil {
			b.Fatalf("Metrics serialization failed: %v", err)
		}
	}
}

// BenchmarkTaskWorkStructureCreation benchmarks TaskWork creation
func BenchmarkTaskWorkStructureCreation(b *testing.B) {
	params := TaskParams{
		Description:  "Benchmark task",
		Prompt:       "Execute benchmark task with detailed prompt information",
		SubagentType: "general-purpose",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		work := TaskWork{
			Index:  i,
			Params: params,
		}
		// Use the work struct to avoid compiler optimization
		_ = work.Index
	}
}

// BenchmarkParallelTaskValidation benchmarks parameter validation for different task counts
func BenchmarkParallelTaskValidation(b *testing.B) {
	taskTool := NewTaskTool(nil, nil, nil, nil, nil).(*taskTool)
	
	// Test with different task counts
	testCases := []int{1, 5, 10, 20}
	
	for _, taskCount := range testCases {
		b.Run("tasks_"+strconv.Itoa(taskCount), func(b *testing.B) {
			// Create parallel params with specified task count
			params := ParallelTaskParams{
				Tasks:         make([]TaskParams, taskCount),
				MaxWorkers:    5,
				AggregateMode: "concat",
			}

			// Fill tasks with valid data
			for i := 0; i < taskCount; i++ {
				params.Tasks[i] = TaskParams{
					Description:  "Benchmark task " + strconv.Itoa(i),
					Prompt:       "Execute benchmark task " + strconv.Itoa(i),
					SubagentType: "general-purpose",
				}
			}

			input, _ := json.Marshal(params)
			inputStr := string(input)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := taskTool.parseParallelTaskParams(inputStr)
				if err != nil {
					b.Fatalf("Validation failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkMinFunction benchmarks the min utility function
func BenchmarkMinFunction(b *testing.B) {
	testCases := []struct {
		name string
		a, b int
	}{
		{"small_numbers", 3, 7},
		{"large_numbers", 1000000, 999999},
		{"equal_numbers", 42, 42},
		{"negative_numbers", -10, -5},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result := min(tc.a, tc.b)
				_ = result // Prevent compiler optimization
			}
		})
	}
}

// TestParallelTaskPerformanceComparison provides a performance comparison between different aggregation modes
func TestParallelTaskPerformanceComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance comparison in short mode")
	}

	taskTool := NewTaskTool(nil, nil, nil, nil, nil).(*taskTool)
	
	// Create test results with varying sizes
	taskCounts := []int{5, 10, 20, 50}
	aggregationModes := []string{"concat", "json_array", "summary"}
	
	for _, taskCount := range taskCounts {
		for _, mode := range aggregationModes {
			t.Run(fmt.Sprintf("tasks_%d_mode_%s", taskCount, mode), func(t *testing.T) {
				// Create results
				results := make([]TaskResult, taskCount)
				for i := 0; i < taskCount; i++ {
					results[i] = TaskResult{
						TaskIndex: i,
						Result:    fmt.Sprintf("Task %d completed with detailed result information", i),
						Error:     nil,
						Metrics: &TaskMetrics{
							TaskID:   fmt.Sprintf("perf-task-%d", i),
							Success:  true,
							Duration: time.Duration(i+1) * time.Second,
						},
					}
				}

				// Measure aggregation time
				start := time.Now()
				result, err := taskTool.aggregateTaskResults(results, mode)
				duration := time.Since(start)

				if err != nil {
					t.Errorf("Aggregation failed: %v", err)
					return
				}

				t.Logf("Tasks: %d, Mode: %s, Duration: %v, Result length: %d", 
					taskCount, mode, duration, len(result))

				// Performance assertions
				if duration > 100*time.Millisecond {
					t.Errorf("Aggregation took too long: %v (should be < 100ms)", duration)
				}
			})
		}
	}
}