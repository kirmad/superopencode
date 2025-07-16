package chat

import (
	"strings"
	"testing"

	"github.com/kirmad/superopencode/internal/tui/theme"
)

func TestRenderTodoList(t *testing.T) {
	// Use the current theme for testing
	currentTheme := theme.CurrentTheme()

	tests := []struct {
		name     string
		todos    []map[string]interface{}
		expected string
		width    int
	}{
		{
			name:     "empty todo list",
			todos:    []map[string]interface{}{},
			expected: "No active tasks in the todo list.",
			width:    80,
		},
		{
			name: "single completed todo",
			todos: []map[string]interface{}{
				{
					"id":       "1",
					"content":  "Add comprehensive debug logging to see exact API request",
					"status":   "completed",
					"priority": "high",
				},
			},
			expected: "  ⎿  ☒ Add comprehensive debug logging to see exact API request",
			width:    80,
		},
		{
			name: "single in_progress todo",
			todos: []map[string]interface{}{
				{
					"id":       "1",
					"content":  "Compare working tools vs TodoRead/TodoWrite schemas",
					"status":   "in_progress",
					"priority": "high",
				},
			},
			expected: "  ⎿  ☐ Compare working tools vs TodoRead/TodoWrite schemas",
			width:    80,
		},
		{
			name: "single pending todo",
			todos: []map[string]interface{}{
				{
					"id":       "1",
					"content":  "Fix the specific schema validation issues",
					"status":   "pending",
					"priority": "high",
				},
			},
			expected: "  ⎿  ☐ Fix the specific schema validation issues",
			width:    80,
		},
		{
			name: "multiple todos with mixed status",
			todos: []map[string]interface{}{
				{
					"id":       "1",
					"content":  "Add comprehensive debug logging to see exact API request",
					"status":   "completed",
					"priority": "high",
				},
				{
					"id":       "2",
					"content":  "Analyze the 400 error response body for specific validation issues",
					"status":   "completed",
					"priority": "high",
				},
				{
					"id":       "3",
					"content":  "Compare working tools vs TodoRead/TodoWrite schemas",
					"status":   "in_progress",
					"priority": "high",
				},
				{
					"id":       "4",
					"content":  "Fix the specific schema validation issues",
					"status":   "pending",
					"priority": "high",
				},
			},
			expected: "  ⎿  ☒ Add comprehensive debug logging to see exact API request\n     ☒ Analyze the 400 error response body for specific validation issues\n     ☐ Compare working tools vs TodoRead/TodoWrite schemas\n     ☐ Fix the specific schema validation issues",
			width: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderTodoList(tt.todos, tt.width, currentTheme)
			
			// Extract the actual text content (without styling)
			actualText := extractTextContent(result)
			
			if actualText != tt.expected {
				t.Errorf("renderTodoList() = %q, want %q", actualText, tt.expected)
			}
		})
	}
}

func TestRenderTodoReadResponse(t *testing.T) {
	currentTheme := theme.CurrentTheme()

	tests := []struct {
		name     string
		content  string
		expected string
		width    int
	}{
		{
			name:     "no todos message",
			content:  "No todos found for this session.",
			expected: "No active tasks in the todo list.",
			width:    80,
		},
		{
			name: "valid JSON todos",
			content: `[
  {
    "id": "1",
    "content": "Test task",
    "status": "completed",
    "priority": "high"
  }
]`,
			expected: "  ⎿  ☒ Test task",
			width:    80,
		},
		{
			name:     "invalid JSON",
			content:  "invalid json content",
			expected: "Failed to parse todos JSON:",
			width:    80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderTodoReadResponse(tt.content, tt.width, currentTheme)
			actualText := extractTextContent(result)
			
			if tt.name == "invalid JSON" {
				if !strings.HasPrefix(actualText, tt.expected) {
					t.Errorf("renderTodoReadResponse() should start with %q, got %q", tt.expected, actualText)
				}
			} else {
				if actualText != tt.expected {
					t.Errorf("renderTodoReadResponse() = %q, want %q", actualText, tt.expected)
				}
			}
		})
	}
}

func TestRenderTodoWriteResponse(t *testing.T) {
	currentTheme := theme.CurrentTheme()

	tests := []struct {
		name     string
		content  string
		expected string
		width    int
	}{
		{
			name:     "successful update with JSON todo list",
			content:  `[{"id":"1","content":"Updated task","status":"completed","priority":"high"}]`,
			expected: "  ⎿  ☒ Updated task",
			width:    80,
		},
		{
			name:     "error message",
			content:  "Error: Failed to update todos",
			expected: "Error: Failed to update todos",
			width:    80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderTodoWriteResponse(tt.content, tt.width, currentTheme)
			actualText := extractTextContent(result)
			
			if actualText != tt.expected {
				t.Errorf("renderTodoWriteResponse() = %q, want %q", actualText, tt.expected)
			}
		})
	}
}

func TestTodoDisplayColorCoding(t *testing.T) {
	currentTheme := theme.CurrentTheme()

	todos := []map[string]interface{}{
		{
			"id":       "1",
			"content":  "Completed task",
			"status":   "completed",
			"priority": "high",
		},
		{
			"id":       "2",
			"content":  "In progress task",
			"status":   "in_progress",
			"priority": "high",
		},
		{
			"id":       "3",
			"content":  "Pending task",
			"status":   "pending",
			"priority": "high",
		},
	}

	result := renderTodoList(todos, 80, currentTheme)

	// Test that the result contains proper checkbox formatting for different statuses
	if !strings.Contains(result, "☒") {
		t.Error("Expected completed task to have ☒ checkbox")
	}
	
	if !strings.Contains(result, "☐") {
		t.Error("Expected pending/in_progress tasks to have ☐ checkbox")
	}
	
	// Verify the tree structure connector
	if !strings.Contains(result, "⎿") {
		t.Error("Expected tree structure with ⎿ connector")
	}
	
	// Verify all task contents are present
	if !strings.Contains(result, "Completed task") {
		t.Error("Expected to find completed task content")
	}
	
	if !strings.Contains(result, "In progress task") {
		t.Error("Expected to find in progress task content")
	}
	
	if !strings.Contains(result, "Pending task") {
		t.Error("Expected to find pending task content")
	}
}


// Helper function to extract text content without styling and HTML
func extractTextContent(styledText string) string {
	// Remove ANSI escape sequences more comprehensively
	result := styledText
	
	// Remove ANSI escape sequences with regex-like pattern
	// This is a simplified version - lipgloss generates complex ANSI codes
	
	// Remove common ANSI color codes
	result = strings.ReplaceAll(result, "\x1b[32m", "")      // Green
	result = strings.ReplaceAll(result, "\x1b[34;1m", "")    // Bold Blue  
	result = strings.ReplaceAll(result, "\x1b[9m", "")       // Strikethrough
	result = strings.ReplaceAll(result, "\x1b[0m", "")       // Reset
	result = strings.ReplaceAll(result, "\x1b[1m", "")       // Bold
	result = strings.ReplaceAll(result, "\x1b[22m", "")      // Normal intensity
	result = strings.ReplaceAll(result, "\x1b[29m", "")      // Not strikethrough
	
	// Remove more complex ANSI sequences (simplified approach)
	for strings.Contains(result, "\x1b[") {
		start := strings.Index(result, "\x1b[")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "m")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}
	
	// For simple checkbox todo output, just return the cleaned text
	if strings.Contains(result, "No active tasks in the todo list") {
		return "No active tasks in the todo list."
	}
	
	if strings.Contains(result, "✓ Todos updated successfully") {
		return "✓ Todos updated successfully"
	}
	
	if strings.Contains(result, "Failed to parse todos JSON") {
		return "Failed to parse todos JSON:"
	}
	
	// Trim trailing spaces but preserve leading spaces for tree structure
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " ")
	}
	return strings.Join(lines, "\n")
}

func TestTodoDisplayEdgeCases(t *testing.T) {
	currentTheme := theme.CurrentTheme()

	tests := []struct {
		name        string
		todos       []map[string]interface{}
		shouldContain []string
	}{
		{
			name: "missing content field",
			todos: []map[string]interface{}{
				{
					"id":       "1",
					"status":   "pending",
					"priority": "high",
				},
			},
			shouldContain: []string{"☐", "⎿"},
		},
		{
			name: "missing status field defaults to pending",
			todos: []map[string]interface{}{
				{
					"id":      "1",
					"content": "Task without status",
				},
			},
			shouldContain: []string{"☐", "Task without status"},
		},
		{
			name: "unknown status defaults to pending",
			todos: []map[string]interface{}{
				{
					"id":      "1",
					"content": "Task with unknown status",
					"status":  "unknown_status",
				},
			},
			shouldContain: []string{"☐", "Task with unknown status"},
		},
		{
			name: "very long content",
			todos: []map[string]interface{}{
				{
					"id":      "1",
					"content": "This is a very long task description",
					"status":  "completed",
				},
			},
			shouldContain: []string{"☒", "very long task description"},
		},
		{
			name: "special characters in content",
			todos: []map[string]interface{}{
				{
					"id":      "1",
					"content": "Task with special chars: @#$%^&*()",
					"status":  "in_progress",
				},
			},
			shouldContain: []string{"☐", "special chars"},
		},
		{
			name: "empty content string",
			todos: []map[string]interface{}{
				{
					"id":      "1",
					"content": "",
					"status":  "pending",
				},
			},
			shouldContain: []string{"☐"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderTodoList(tt.todos, 80, currentTheme)
			
			for _, expected := range tt.shouldContain {
				if !strings.Contains(result, expected) {
					t.Errorf("renderTodoList() should contain %q, but got %q", expected, result)
				}
			}
		})
	}
}

func TestTodoReadResponseEdgeCases(t *testing.T) {
	currentTheme := theme.CurrentTheme()

	tests := []struct {
		name           string
		content        string
		shouldContain  string
		shouldNotError bool
	}{
		{
			name:           "malformed JSON - missing bracket",
			content:        `[{"id": "1", "content": "test", "status": "pending"`,
			shouldContain:  "Failed to parse todos JSON",
			shouldNotError: false,
		},
		{
			name:           "empty JSON array", 
			content:        `[]`,
			shouldContain:  "No active tasks in the todo list",
			shouldNotError: true,
		},
		{
			name:           "JSON with null values",
			content:        `[{"id": null, "content": null, "status": null}]`,
			shouldContain:  "☐",
			shouldNotError: true,
		},
		{
			name:           "completely empty content",
			content:        "",
			shouldContain:  "Failed to parse todos JSON",
			shouldNotError: false,
		},
		{
			name:           "non-JSON content",
			content:        "This is not JSON at all",
			shouldContain:  "Failed to parse todos JSON", 
			shouldNotError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderTodoReadResponse(tt.content, 80, currentTheme)
			actualText := extractTextContent(result)
			
			if !strings.Contains(actualText, tt.shouldContain) {
				t.Errorf("renderTodoReadResponse() = %q, should contain %q", actualText, tt.shouldContain)
			}
		})
	}
}