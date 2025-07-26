package agent

import (
	"testing"

	"github.com/kirmad/superopencode/internal/llm/tools"
)

func TestTodoToolsExist(t *testing.T) {
	// Test that we can create the TODO tools
	todoRead := tools.NewTodoReadTool()
	if todoRead == nil {
		t.Error("NewTodoReadTool() returned nil")
	}

	todoWrite := tools.NewTodoWriteTool()
	if todoWrite == nil {
		t.Error("NewTodoWriteTool() returned nil")
	}

	// Test tool info
	readInfo := todoRead.Info()
	if readInfo.Name != "TodoRead" {
		t.Errorf("Expected TodoRead, got %s", readInfo.Name)
	}

	writeInfo := todoWrite.Info()
	if writeInfo.Name != "TodoWrite" {
		t.Errorf("Expected TodoWrite, got %s", writeInfo.Name)
	}
}

func TestTaskToolCreation(t *testing.T) {
	// Test that we can create the Task tool
	taskTool := NewTaskTool(nil, nil, nil, nil, nil)
	if taskTool == nil {
		t.Error("NewTaskTool() returned nil")
	}

	// Test tool info
	info := taskTool.Info()
	if info.Name != "Task" {
		t.Errorf("Expected Task, got %s", info.Name)
	}

	// Test required parameters
	expectedRequired := []string{"description", "prompt", "subagent_type"}
	for i, req := range expectedRequired {
		if i >= len(info.Required) || info.Required[i] != req {
			t.Errorf("Expected required parameter %s at index %d", req, i)
		}
	}
}