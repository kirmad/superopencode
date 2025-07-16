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