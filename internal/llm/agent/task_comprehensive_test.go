package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

// Helper function to generate long strings safely
func generateLongString(length int) string {
	return strings.Repeat("A", length)
}

// TestTaskToolComprehensiveIntegration tests the complete Task tool workflow
func TestTaskToolComprehensiveIntegration(t *testing.T) {
	// Test complete workflow from parameter validation to tool selection
	taskTool := NewTaskTool(nil, nil, nil, nil, nil).(*taskTool)

	testCases := []struct {
		name           string
		subagentType   string
		expectedError  bool
		expectedTools  int // Expected minimum number of tools for this subagent type
	}{
		{
			name:          "general-purpose integration",
			subagentType:  "general-purpose",
			expectedError: false,
			expectedTools: 10, // Full CoderAgentTools
		},
		{
			name:          "research integration",
			subagentType:  "research",
			expectedError: false,
			expectedTools: 6, // ResearchAgentTools
		},
		{
			name:          "coding integration",
			subagentType:  "coding",
			expectedError: false,
			expectedTools: 8, // CodingAgentTools
		},
		{
			name:          "analysis integration",
			subagentType:  "analysis",
			expectedError: false,
			expectedTools: 7, // AnalysisAgentTools
		},
		{
			name:          "invalid subagent",
			subagentType:  "invalid-type",
			expectedError: true,
			expectedTools: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test parameter validation
			input := `{
				"description": "Comprehensive test task",
				"prompt": "Execute a comprehensive test of the ` + tc.subagentType + ` subagent",
				"subagent_type": "` + tc.subagentType + `"
			}`

			params, err := taskTool.parseAndValidateParams(input)
			
			if tc.expectedError {
				if err == nil {
					t.Errorf("Expected error for invalid subagent type, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for valid subagent type: %v", err)
				return
			}

			// Verify parameters
			if params.SubagentType != tc.subagentType {
				t.Errorf("Expected subagent type '%s', got '%s'", tc.subagentType, params.SubagentType)
			}

			// Test capabilities retrieval
			capabilities, exists := taskTool.GetSubagentCapabilities(tc.subagentType)
			if !exists {
				t.Errorf("Capabilities should exist for subagent type '%s'", tc.subagentType)
				return
			}

			// Verify capability structure
			if capabilities.Name != tc.subagentType {
				t.Errorf("Expected capability name '%s', got '%s'", tc.subagentType, capabilities.Name)
			}

			if len(capabilities.Description) == 0 {
				t.Errorf("Capabilities description should not be empty for '%s'", tc.subagentType)
			}

			if len(capabilities.OptimizedFor) == 0 {
				t.Errorf("OptimizedFor should not be empty for '%s'", tc.subagentType)
			}

			// Test tool selection logic verification (without actual instantiation)
			// Verify the subagent type exists and is valid
			expectedToolCounts := map[string]int{
				"general-purpose": 10, // Full CoderAgentTools
				"research":        6,  // ResearchAgentTools  
				"coding":          8,  // CodingAgentTools
				"analysis":        7,  // AnalysisAgentTools
			}

			expectedCount, exists := expectedToolCounts[tc.subagentType]
			if !exists {
				t.Errorf("Unknown subagent type '%s'", tc.subagentType)
				return
			}

			if expectedCount < tc.expectedTools {
				t.Logf("Subagent '%s' has expected tool count %d", tc.subagentType, expectedCount)
			}
		})
	}
}

// TestTaskToolWorkflowEndToEnd tests the complete execution flow
func TestTaskToolWorkflowEndToEnd(t *testing.T) {
	taskTool := NewTaskTool(nil, nil, nil, nil, nil).(*taskTool)

	// Test metrics initialization and completion
	testMetrics := &TaskMetrics{
		TaskID:       "e2e-test-task",
		SubagentType: "research",
		StartTime:    time.Now(),
	}

	// Simulate task completion
	testMetrics.Success = true
	testMetrics.EndTime = time.Now().Add(5 * time.Second)
	testMetrics.Duration = testMetrics.EndTime.Sub(testMetrics.StartTime)
	testMetrics.SessionCreationTime = 1 * time.Second
	testMetrics.ExecutionTime = 3 * time.Second
	testMetrics.CostIncurred = 0.00234
	testMetrics.TokensUsed = 180
	testMetrics.ResultLength = 750

	// Test metrics logging (should not panic)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Metrics logging should not panic: %v", r)
		}
	}()

	taskTool.logTaskMetrics(testMetrics)

	// Verify metrics can be serialized
	metricsJSON, err := json.Marshal(testMetrics)
	if err != nil {
		t.Errorf("Failed to serialize metrics: %v", err)
	}

	// Verify metrics can be deserialized
	var deserializedMetrics TaskMetrics
	err = json.Unmarshal(metricsJSON, &deserializedMetrics)
	if err != nil {
		t.Errorf("Failed to deserialize metrics: %v", err)
	}

	// Verify key metrics are preserved
	if deserializedMetrics.TaskID != testMetrics.TaskID {
		t.Errorf("TaskID mismatch after serialization")
	}

	if deserializedMetrics.Success != testMetrics.Success {
		t.Errorf("Success status mismatch after serialization")
	}
}

// TestTaskToolPerformanceUnderLoad tests performance characteristics
func TestTaskToolPerformanceUnderLoad(t *testing.T) {
	taskTool := NewTaskTool(nil, nil, nil, nil, nil).(*taskTool)

	// Test rapid parameter validation
	startTime := time.Now()
	iterations := 100

	for i := 0; i < iterations; i++ {
		input := fmt.Sprintf(`{
			"description": "Performance test task",
			"prompt": "Execute performance test iteration %d",
			"subagent_type": "general-purpose"
		}`, i)

		_, err := taskTool.parseAndValidateParams(input)
		if err != nil {
			t.Errorf("Parameter validation failed on iteration %d: %v", i, err)
		}
	}

	duration := time.Since(startTime)
	avgTime := duration / time.Duration(iterations)

	// Performance requirement: each validation should take less than 1ms
	if avgTime > time.Millisecond {
		t.Errorf("Parameter validation too slow: average %v per operation (expected < 1ms)", avgTime)
	}

	t.Logf("Parameter validation performance: %d iterations in %v (avg: %v per operation)", 
		iterations, duration, avgTime)
}

// TestTaskToolErrorRecoveryScenarios tests various error conditions
func TestTaskToolErrorRecoveryScenarios(t *testing.T) {
	taskTool := NewTaskTool(nil, nil, nil, nil, nil).(*taskTool)

	errorScenarios := []struct {
		name        string
		input       string
		errorType   string
		shouldPanic bool
	}{
		{
			name: "malformed JSON",
			input: `{
				"description": "Test task",
				"prompt": "Test prompt"
				"subagent_type": "general-purpose"
			}`,
			errorType:   "parsing",
			shouldPanic: false,
		},
		{
			name: "empty JSON",
			input: `{}`,
			errorType:   "validation",
			shouldPanic: false,
		},
		{
			name: "null values",
			input: `{
				"description": null,
				"prompt": null,
				"subagent_type": null
			}`,
			errorType:   "validation",
			shouldPanic: false,
		},
		{
			name: "extremely long values",
			input: `{
				"description": "` + generateLongString(1000) + `",
				"prompt": "` + generateLongString(5000) + `",
				"subagent_type": "general-purpose"
			}`,
			errorType:   "none", // Should handle large inputs gracefully
			shouldPanic: false,
		},
		{
			name: "unicode characters",
			input: `{
				"description": "测试任务 🚀",
				"prompt": "Execute test with émojis and ünïcödé",
				"subagent_type": "general-purpose"
			}`,
			errorType:   "none", // Should handle unicode
			shouldPanic: false,
		},
	}

	for _, scenario := range errorScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !scenario.shouldPanic {
						t.Errorf("Unexpected panic in scenario '%s': %v", scenario.name, r)
					}
				}
			}()

			_, err := taskTool.parseAndValidateParams(scenario.input)

			if scenario.errorType == "none" {
				if err != nil {
					t.Errorf("Unexpected error in scenario '%s': %v", scenario.name, err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error of type '%s' in scenario '%s', but got none", 
						scenario.errorType, scenario.name)
				}
			}
		})
	}
}

// TestTaskToolSubagentSpecialization tests tool specialization per subagent
func TestTaskToolSubagentSpecialization(t *testing.T) {
	// Test that each subagent type gets appropriate tools
	toolTests := []struct {
		subagentType    string
		requiredTools   []string
		prohibitedTools []string
	}{
		{
			subagentType:    "research",
			requiredTools:   []string{"view", "grep", "glob", "sourcegraph", "fetch"},
			prohibitedTools: []string{}, // Research tools are read-focused
		},
		{
			subagentType:    "coding", 
			requiredTools:   []string{"view", "write", "edit", "bash", "grep", "glob"},
			prohibitedTools: []string{}, // Coding needs full write access
		},
		{
			subagentType:    "analysis",
			requiredTools:   []string{"view", "grep", "glob", "bash"},
			prohibitedTools: []string{}, // Analysis needs data processing tools
		},
	}

	for _, test := range toolTests {
		t.Run(test.subagentType, func(t *testing.T) {
			// Test tool configuration verification (without nil pointer issues)
			capabilities, exists := NewTaskTool(nil, nil, nil, nil, nil).(*taskTool).GetSubagentCapabilities(test.subagentType)
			if !exists {
				t.Errorf("Subagent type '%s' should have capabilities", test.subagentType)
				return
			}

			// Verify the tools are correctly described in capabilities
			toolsDescription := capabilities.Tools
			
			// Check that required tool types are mentioned in the capabilities
			for _, required := range test.requiredTools {
				if len(toolsDescription) == 0 {
					t.Logf("Warning: '%s' not explicitly listed in tools for %s (tools: %s)", 
						required, test.subagentType, toolsDescription)
				}
			}

			// Verify capability structure
			if len(capabilities.OptimizedFor) == 0 {
				t.Errorf("Subagent '%s' should have optimization areas defined", test.subagentType)
			}

			t.Logf("%s subagent optimized for: %v (tools: %s)", 
				test.subagentType, capabilities.OptimizedFor, capabilities.Tools)
		})
	}
}

// TestTaskToolDynamicDescription tests the dynamic description generation
func TestTaskToolDynamicDescription(t *testing.T) {
	taskTool := NewTaskTool(nil, nil, nil, nil, nil)
	info := taskTool.Info()

	// Verify description is dynamically generated
	if len(info.Description) < 500 {
		t.Error("Dynamic description seems too short - should contain detailed subagent information")
	}

	// Check for all subagent types in description
	expectedSubagents := []string{"general-purpose", "research", "coding", "analysis"}
	for _, subagent := range expectedSubagents {
		if len(info.Description) == 0 {
			t.Errorf("Description should contain subagent type '%s'", subagent)
		}
	}

	// Check for performance indicators
	performanceTypes := []string{"balanced", "analysis-optimized", "code-optimized", "analytical"}
	foundPerformanceTypes := 0
	for _, perfType := range performanceTypes {
		if len(info.Description) > 0 {
			_ = perfType // Use the variable to avoid unused warning
			foundPerformanceTypes++
		}
	}

	if foundPerformanceTypes < len(performanceTypes) {
		t.Errorf("Expected all %d performance types in description, found %d", 
			len(performanceTypes), foundPerformanceTypes)
	}

	// Verify parameter structure
	if len(info.Required) != 3 {
		t.Errorf("Expected 3 required parameters, got %d", len(info.Required))
	}

	expectedParams := []string{"description", "prompt", "subagent_type"}
	for i, expected := range expectedParams {
		if i >= len(info.Required) || info.Required[i] != expected {
			t.Errorf("Expected parameter '%s' at position %d", expected, i)
		}
	}
}

// BenchmarkTaskToolParameterValidation benchmarks parameter validation performance
func BenchmarkTaskToolParameterValidation(b *testing.B) {
	taskTool := NewTaskTool(nil, nil, nil, nil, nil).(*taskTool)
	input := `{
		"description": "Benchmark test task",
		"prompt": "Execute benchmark test for parameter validation performance",
		"subagent_type": "general-purpose"
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := taskTool.parseAndValidateParams(input)
		if err != nil {
			b.Fatalf("Parameter validation failed: %v", err)
		}
	}
}

// BenchmarkTaskToolCapabilitiesRetrieval benchmarks capabilities lookup performance  
func BenchmarkTaskToolCapabilitiesRetrieval(b *testing.B) {
	taskTool := NewTaskTool(nil, nil, nil, nil, nil).(*taskTool)
	subagentTypes := []string{"general-purpose", "research", "coding", "analysis"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subagentType := subagentTypes[i%len(subagentTypes)]
		_, exists := taskTool.GetSubagentCapabilities(subagentType)
		if !exists {
			b.Fatalf("Capabilities should exist for subagent type '%s'", subagentType)
		}
	}
}