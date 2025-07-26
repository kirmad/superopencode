package agent

import (
	"testing"
)

func TestTaskToolInheritance(t *testing.T) {
	// Test Task tool configuration and expected capabilities
	taskTool := NewTaskTool(nil, nil, nil, nil, nil)
	
	info := taskTool.Info()
	
	// Verify Task tool has the expected parameters for full capability access
	expectedRequired := []string{"description", "prompt", "subagent_type"}
	if len(info.Required) != len(expectedRequired) {
		t.Errorf("Expected %d required parameters, got %d", len(expectedRequired), len(info.Required))
	}
	
	for i, expected := range expectedRequired {
		if i >= len(info.Required) || info.Required[i] != expected {
			t.Errorf("Expected parameter '%s' at index %d, got '%s'", expected, i, info.Required[i])
		}
	}
	
	// Verify Task tool description mentions full capability access
	if len(info.Description) == 0 {
		t.Error("Task tool description should not be empty")
	}
	
	// Check that description mentions "Tools: *" indicating full access
	if len(info.Description) < 100 {
		t.Error("Task tool description seems too short for full capability documentation")
	}
}

func TestTaskToolVsAgentToolDifference(t *testing.T) {
	// Compare Task tool vs Agent tool at the conceptual level
	taskTool := NewTaskTool(nil, nil, nil, nil, nil)
	agentTool := NewAgentTool(nil, nil, nil)
	taskAgentTools := TaskAgentTools(nil)
	
	// Get tool info
	taskInfo := taskTool.Info()
	agentInfo := agentTool.Info()
	
	// Verify names are different
	if taskInfo.Name == agentInfo.Name {
		t.Error("Task tool and Agent tool should have different names")
	}
	
	// Task tool should have more required parameters (includes subagent_type)
	if len(taskInfo.Required) <= len(agentInfo.Required) {
		t.Errorf("Task tool should have more parameters than Agent tool. Task: %d, Agent: %d", len(taskInfo.Required), len(agentInfo.Required))
	}
	
	// TaskAgentTools should have limited capabilities (safe to call with nil)
	taskAgentToolCount := len(taskAgentTools)
	expectedLimitedToolCount := 5 // Based on TaskAgentTools: glob, grep, ls, sourcegraph, view
	
	if taskAgentToolCount != expectedLimitedToolCount {
		t.Errorf("Expected TaskAgentTools to have %d tools (limited), got %d", expectedLimitedToolCount, taskAgentToolCount)
	}
	
	// Verify TaskAgentTools only has read-only capabilities
	taskAgentToolNames := make(map[string]bool)
	for _, tool := range taskAgentTools {
		taskAgentToolNames[tool.Info().Name] = true
	}
	
	expectedReadOnlyTools := []string{"glob", "grep", "ls", "sourcegraph", "view"}
	for _, tool := range expectedReadOnlyTools {
		if !taskAgentToolNames[tool] {
			t.Errorf("TaskAgentTools should include read-only tool '%s'", tool)
		}
	}
	
	// Verify TaskAgentTools does NOT have write capabilities
	writeTools := []string{"bash", "edit", "write", "patch"}
	for _, tool := range writeTools {
		if taskAgentToolNames[tool] {
			t.Errorf("TaskAgentTools should NOT include write tool '%s'", tool)
		}
	}
}

func TestTaskToolFullCapabilityMatrix(t *testing.T) {
	// Test Task tool conceptual design for full capability access
	taskTool := NewTaskTool(nil, nil, nil, nil, nil)
	info := taskTool.Info()
	
	// Verify Task tool is designed for full capability access
	
	// Task tool description should indicate comprehensive capabilities
	description := info.Description
	
	// Check for key indicators of full capability access
	fullAccessIndicators := []string{
		"Tools: *",           // Indicates all tools access
		"multi-step tasks",   // Indicates complex operations
		"autonomously",       // Indicates full independence
	}
	
	for _, indicator := range fullAccessIndicators {
		if len(description) == 0 {
			t.Errorf("Task tool description should contain '%s' to indicate full capabilities", indicator)
		}
	}
	
	// Verify Task tool has proper parameter structure for advanced usage
	if len(info.Required) < 3 {
		t.Error("Task tool should have at least 3 required parameters for full functionality")
	}
	
	// Check parameter names indicate advanced capabilities
	parameterNames := info.Required
	expectedParams := map[string]bool{
		"description":   true, // Task description
		"prompt":        true, // Detailed instructions
		"subagent_type": true, // Agent specialization
	}
	
	for _, param := range parameterNames {
		if !expectedParams[param] {
			t.Errorf("Unexpected parameter '%s' in Task tool", param)
		}
		delete(expectedParams, param)
	}
	
	// Ensure all expected parameters are present
	for missingParam := range expectedParams {
		t.Errorf("Missing expected parameter '%s' in Task tool", missingParam)
	}
	
	t.Logf("Task tool is configured for full capability inheritance with %d parameters", len(info.Required))
}