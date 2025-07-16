package dialog

import (
	"testing"
)

func TestSlashCommandProcessor_IsSlashCommand(t *testing.T) {
	processor := NewSlashCommandProcessor([]Command{})
	
	tests := []struct {
		input    string
		expected bool
	}{
		{"/test", true},
		{"/test command", true},
		{"  /test  ", true},
		{"test", false},
		{"", false},
		{" ", false},
		{"/", false}, // Just slash without command
	}
	
	for _, test := range tests {
		result := processor.IsSlashCommand(test.input)
		if result != test.expected {
			t.Errorf("IsSlashCommand(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestValidateSlashCommand(t *testing.T) {
	tests := []struct {
		input     string
		shouldErr bool
	}{
		{"/test", false},
		{"/test argument", false},
		{"/user:test", false},
		{"/project:test", false},
		{"test", true},     // No slash
		{"/", true},        // Just slash
		{"/ ", true},       // Slash with just space
		{"/test arg", false}, // Valid with args
	}
	
	for _, test := range tests {
		err := ValidateSlashCommand(test.input)
		hasErr := err != nil
		if hasErr != test.shouldErr {
			t.Errorf("ValidateSlashCommand(%q) error = %v, expected error = %v", 
				test.input, err, test.shouldErr)
		}
	}
}

func TestSlashCommandProcessor_ProcessSlashCommand(t *testing.T) {
	commands := []Command{
		{
			ID:      "test",
			Title:   "Test Command",
			Content: "This is a test command\n\nExecute: $ARGUMENTS",
		},
		{
			ID:      "user:custom",
			Title:   "Custom User Command",
			Content: "User custom command content",
		},
		{
			ID:      "simple",
			Title:   "Simple Command",
			Content: "Simple command without arguments",
		},
	}
	
	processor := NewSlashCommandProcessor(commands)
	
	tests := []struct {
		input       string
		shouldErr   bool
		needsDialog bool
		commandID   string
	}{
		{"/test some arguments", false, true, "test"},    // Has named args
		{"/simple", false, false, "simple"},              // No named args
		{"/custom", false, false, "user:custom"},         // Should find user:custom
		{"/nonexistent", true, false, ""},                // Command not found
	}
	
	for _, test := range tests {
		result := processor.ProcessSlashCommand(test.input)
		
		hasErr := result.Error != nil
		if hasErr != test.shouldErr {
			t.Errorf("ProcessSlashCommand(%q) error = %v, expected error = %v", 
				test.input, result.Error, test.shouldErr)
			continue
		}
		
		if !hasErr {
			if result.NeedsArgDialog != test.needsDialog {
				t.Errorf("ProcessSlashCommand(%q) needsDialog = %v, expected %v", 
					test.input, result.NeedsArgDialog, test.needsDialog)
			}
			
			if result.Processed.Command.ID != test.commandID {
				t.Errorf("ProcessSlashCommand(%q) commandID = %v, expected %v", 
					test.input, result.Processed.Command.ID, test.commandID)
			}
		}
	}
}

func TestSlashCommandProcessor_GetAvailableCommands(t *testing.T) {
	commands := []Command{
		{ID: "test", Title: "Test"},
		{ID: "user:custom", Title: "Custom"},
		{ID: "project:deploy", Title: "Deploy"},
		{ID: "user:test", Title: "User Test"}, // Duplicate base name
	}
	
	processor := NewSlashCommandProcessor(commands)
	available := processor.GetAvailableCommands()
	
	// Should deduplicate and return base names
	expected := []string{"test", "custom", "deploy"}
	
	if len(available) != len(expected) {
		t.Errorf("GetAvailableCommands() returned %d commands, expected %d", 
			len(available), len(expected))
	}
	
	// Check that all expected commands are present
	availableMap := make(map[string]bool)
	for _, cmd := range available {
		availableMap[cmd] = true
	}
	
	for _, exp := range expected {
		if !availableMap[exp] {
			t.Errorf("GetAvailableCommands() missing expected command: %s", exp)
		}
	}
}

func TestContentConcatenation(t *testing.T) {
	commands := []Command{
		{
			ID:      "test",
			Title:   "Test Command",
			Content: "Base command content",
		},
	}
	
	processor := NewSlashCommandProcessor(commands)
	
	result := processor.ProcessSlashCommand("/test additional user text")
	if result.Error != nil {
		t.Fatalf("ProcessSlashCommand failed: %v", result.Error)
	}
	
	expected := "Base command content\n\nadditional user text"
	if result.Processed.Content != expected {
		t.Errorf("Content concatenation failed.\nExpected: %q\nGot: %q", 
			expected, result.Processed.Content)
	}
}

func TestCommandResolutionOrder(t *testing.T) {
	commands := []Command{
		{ID: "test", Title: "Base Test", Content: "base"},
		{ID: "user:test", Title: "User Test", Content: "user"},
		{ID: "project:test", Title: "Project Test", Content: "project"},
	}
	
	processor := NewSlashCommandProcessor(commands)
	
	// Should find base command first
	result := processor.ProcessSlashCommand("/test")
	if result.Error != nil {
		t.Fatalf("ProcessSlashCommand failed: %v", result.Error)
	}
	
	if result.Processed.Command.ID != "test" {
		t.Errorf("Expected to find 'test' command, got '%s'", result.Processed.Command.ID)
	}
}

// Tests for slash command suggestions functionality
func TestSlashCommandProcessor_GetSuggestions(t *testing.T) {
	commands := []Command{
		{ID: "build", Title: "Build Project", Description: "Build the current project"},
		{ID: "test", Title: "Run Tests", Description: "Execute test suite"},
		{ID: "deploy", Title: "Deploy App", Description: "Deploy to production"},
		{ID: "user:backup", Title: "Backup Files", Description: "Create backup"},
		{ID: "project:setup", Title: "Setup Project", Description: "Initialize project"},
	}
	
	processor := NewSlashCommandProcessor(commands)
	
	tests := []struct {
		input    string
		expected []string
		maxCount int
	}{
		// Empty input should return all commands
		{"", []string{"build", "test", "deploy", "backup", "setup"}, 10},
		// Partial matches
		{"/", []string{"build", "test", "deploy", "backup", "setup"}, 10},
		{"/b", []string{"build", "backup"}, 10},
		{"/te", []string{"test"}, 10},
		{"/dep", []string{"deploy"}, 10},
		// Case insensitive
		{"/B", []string{"build", "backup"}, 10},
		{"/TEST", []string{"test"}, 10},
		// No matches
		{"/xyz", []string{}, 10},
		// Limit results
		{"/", []string{"build", "test"}, 2},
		// Non-slash prefix
		{"b", []string{}, 10},
	}
	
	for _, test := range tests {
		suggestions := processor.GetSuggestions(test.input, test.maxCount)
		
		if len(suggestions) != len(test.expected) {
			t.Errorf("GetSuggestions(%q, %d) returned %d suggestions, expected %d", 
				test.input, test.maxCount, len(suggestions), len(test.expected))
			continue
		}
		
		for i, expected := range test.expected {
			if i >= len(suggestions) || suggestions[i].Command != expected {
				t.Errorf("GetSuggestions(%q, %d) suggestion[%d] = %v, expected command '%s'", 
					test.input, test.maxCount, i, suggestions[i], expected)
			}
		}
	}
}

func TestSlashCommandProcessor_GetSuggestionsWithDetails(t *testing.T) {
	commands := []Command{
		{ID: "build", Title: "Build Project", Description: "Build the current project"},
		{ID: "test", Title: "Run Tests", Description: "Execute test suite"},
	}
	
	processor := NewSlashCommandProcessor(commands)
	suggestions := processor.GetSuggestions("/b", 10)
	
	if len(suggestions) != 1 {
		t.Fatalf("Expected 1 suggestion, got %d", len(suggestions))
	}
	
	suggestion := suggestions[0]
	if suggestion.Command != "build" {
		t.Errorf("Expected command 'build', got '%s'", suggestion.Command)
	}
	if suggestion.Title != "Build Project" {
		t.Errorf("Expected title 'Build Project', got '%s'", suggestion.Title)
	}
	if suggestion.Description != "Build the current project" {
		t.Errorf("Expected description 'Build the current project', got '%s'", suggestion.Description)
	}
}

func TestSlashCommandProcessor_AutofillCommand(t *testing.T) {
	commands := []Command{
		{ID: "build", Title: "Build Project", Description: "Build the current project"},
		{ID: "backup", Title: "Backup Files", Description: "Create backup"},
		{ID: "test", Title: "Run Tests", Description: "Execute test suite"},
	}
	
	processor := NewSlashCommandProcessor(commands)
	
	tests := []struct {
		input    string
		expected string
	}{
		// Unique matches
		{"/bu", "/build"},  // Only build starts with "bu"
		{"/t", "/test"},    // Only test starts with "t"
		// Non-unique matches
		{"/b", "/b"},       // Both build and backup start with "b", no autofill
		// No matches
		{"/xyz", "/xyz"},   // No command starts with "xyz"
		// Already complete
		{"/build", "/build"}, // Already complete, no change
		// Non-slash input
		{"bu", "bu"},       // Not a slash command
	}
	
	for _, test := range tests {
		result := processor.AutofillCommand(test.input)
		if result != test.expected {
			t.Errorf("AutofillCommand(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestSlashCommandProcessor_GetCommonPrefix(t *testing.T) {
	commands := []Command{
		{ID: "build", Title: "Build Project", Description: "Build the current project"},
		{ID: "backup", Title: "Backup Files", Description: "Create backup"},
		{ID: "test", Title: "Run Tests", Description: "Execute test suite"},
	}
	
	processor := NewSlashCommandProcessor(commands)
	
	tests := []struct {
		input    string
		expected string
	}{
		// Common prefix cases
		{"/b", "/b"},       // "build" and "backup" have common prefix "b", but no improvement
		{"/bu", "/build"},  // Only "build" starts with "bu", so return full command
		{"/bui", "/build"}, // Only "build" starts with "bui", so return full command
		// No matches
		{"/xyz", "/xyz"},   // No command starts with "xyz"
		// Single character after slash
		{"/t", "/test"},    // Only "test" starts with "t"
		// Non-slash input
		{"b", "b"},         // Not a slash command
	}
	
	for _, test := range tests {
		result := processor.GetCommonPrefix(test.input)
		if result != test.expected {
			t.Errorf("GetCommonPrefix(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}