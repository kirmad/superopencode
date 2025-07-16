package dialog

import (
	"fmt"
	"regexp"
	"strings"
)

// SlashCommandProcessor handles parsing and processing of slash commands
type SlashCommandProcessor struct {
	commands []Command
}

// ProcessedCommand represents a processed slash command ready for execution
type ProcessedCommand struct {
	Command         *Command
	Content         string
	Args            map[string]string
	HasNamedArgs    bool
	RemainingText   string
	OriginalCommand string
}

// SlashCommandResult represents the result of processing a slash command
type SlashCommandResult struct {
	Processed    *ProcessedCommand
	Error        error
	NeedsArgDialog bool
}

// Suggestion represents a slash command suggestion for autocomplete
type Suggestion struct {
	Command     string // The command name (without /)
	Title       string // Display title
	Description string // Command description
}

// namedArgPattern is used to find named arguments in command content
var slashNamedArgPattern = regexp.MustCompile(`\$([A-Z][A-Z0-9_]*)`)

// NewSlashCommandProcessor creates a new slash command processor
func NewSlashCommandProcessor(commands []Command) *SlashCommandProcessor {
	return &SlashCommandProcessor{
		commands: commands,
	}
}

// IsSlashCommand checks if the input text is a slash command
func (scp *SlashCommandProcessor) IsSlashCommand(input string) bool {
	trimmed := strings.TrimSpace(input)
	return strings.HasPrefix(trimmed, "/") && len(trimmed) > 1
}

// ProcessSlashCommand parses and processes a slash command input
func (scp *SlashCommandProcessor) ProcessSlashCommand(input string) *SlashCommandResult {
	if !scp.IsSlashCommand(input) {
		return &SlashCommandResult{
			Error: fmt.Errorf("not a slash command: %s", input),
		}
	}

	// Remove the leading slash and split into command and remaining text
	trimmed := strings.TrimSpace(input)[1:] // Remove leading "/"
	parts := strings.SplitN(trimmed, " ", 2)
	
	commandName := parts[0]
	remainingText := ""
	if len(parts) > 1 {
		remainingText = strings.TrimSpace(parts[1])
	}

	// Find the command
	command := scp.findCommand(commandName)
	if command == nil {
		return &SlashCommandResult{
			Error: fmt.Errorf("command not found: %s", commandName),
		}
	}

	// Use the stored command content
	commandContent := command.Content
	if commandContent == "" {
		return &SlashCommandResult{
			Error: fmt.Errorf("no content available for command: %s", commandName),
		}
	}

	// Check for named arguments in the command content
	matches := slashNamedArgPattern.FindAllStringSubmatch(commandContent, -1)
	hasNamedArgs := len(matches) > 0

	var combinedContent string
	if remainingText != "" {
		// Append user prompt to command content
		combinedContent = commandContent + "\n\n" + remainingText
	} else {
		// Use command content as-is
		combinedContent = commandContent
	}

	processed := &ProcessedCommand{
		Command:         command,
		Content:         combinedContent,
		Args:            make(map[string]string),
		HasNamedArgs:    hasNamedArgs,
		RemainingText:   remainingText,
		OriginalCommand: commandName,
	}

	// If the command has named arguments, we need to show the dialog
	if hasNamedArgs {
		// Extract unique argument names
		argNames := make([]string, 0)
		argMap := make(map[string]bool)

		for _, match := range matches {
			argName := match[1] // Group 1 is the name without $
			if !argMap[argName] {
				argMap[argName] = true
				argNames = append(argNames, argName)
			}
		}

		return &SlashCommandResult{
			Processed:      processed,
			NeedsArgDialog: true,
		}
	}

	return &SlashCommandResult{
		Processed:      processed,
		NeedsArgDialog: false,
	}
}

// findCommand searches for a command by name, supporting prefixed lookup
func (scp *SlashCommandProcessor) findCommand(name string) *Command {
	// Direct match first
	for _, cmd := range scp.commands {
		if scp.matchesCommandName(cmd.ID, name) {
			return &cmd
		}
	}

	// Try with user: prefix
	userCommand := "user:" + name
	for _, cmd := range scp.commands {
		if scp.matchesCommandName(cmd.ID, userCommand) {
			return &cmd
		}
	}

	// Try with project: prefix
	projectCommand := "project:" + name
	for _, cmd := range scp.commands {
		if scp.matchesCommandName(cmd.ID, projectCommand) {
			return &cmd
		}
	}

	return nil
}

// matchesCommandName checks if a command ID matches the given name
func (scp *SlashCommandProcessor) matchesCommandName(commandID, targetName string) bool {
	return commandID == targetName
}


// GetAvailableCommands returns a list of available command names for autocomplete
func (scp *SlashCommandProcessor) GetAvailableCommands() []string {
	var commands []string
	seen := make(map[string]bool)

	for _, cmd := range scp.commands {
		// Extract base command name
		name := cmd.ID
		if strings.HasPrefix(name, UserCommandPrefix) {
			name = strings.TrimPrefix(name, UserCommandPrefix)
		} else if strings.HasPrefix(name, ProjectCommandPrefix) {
			name = strings.TrimPrefix(name, ProjectCommandPrefix)
		}

		if !seen[name] {
			seen[name] = true
			commands = append(commands, name)
		}
	}

	return commands
}

// GetSuggestions returns slash command suggestions based on partial input
func (scp *SlashCommandProcessor) GetSuggestions(input string, maxCount int) []Suggestion {
	var suggestions []Suggestion
	
	trimmed := strings.TrimSpace(input)
	
	// Empty input or just "/" - return all commands
	if trimmed == "" || trimmed == "/" {
		return scp.getAllSuggestions(maxCount)
	}
	
	// If input doesn't start with /, return empty suggestions for now
	if !strings.HasPrefix(trimmed, "/") {
		return suggestions
	}
	
	// Extract the partial command name (remove leading / and whitespace)
	if len(trimmed) <= 1 {
		// Just "/" - return all commands
		return scp.getAllSuggestions(maxCount)
	}
	
	partial := strings.ToLower(trimmed[1:]) // Remove "/" and convert to lowercase for comparison
	seen := make(map[string]bool)
	
	for _, cmd := range scp.commands {
		// Extract base command name
		name := cmd.ID
		if strings.HasPrefix(name, UserCommandPrefix) {
			name = strings.TrimPrefix(name, UserCommandPrefix)
		} else if strings.HasPrefix(name, ProjectCommandPrefix) {
			name = strings.TrimPrefix(name, ProjectCommandPrefix)
		}
		
		// Skip if we've already added this command name
		if seen[name] {
			continue
		}
		
		// Check if the command name starts with the partial input (case insensitive)
		if strings.HasPrefix(strings.ToLower(name), partial) {
			suggestion := Suggestion{
				Command:     name,
				Title:       cmd.Title,
				Description: cmd.Description,
			}
			suggestions = append(suggestions, suggestion)
			seen[name] = true
			
			// Limit results
			if len(suggestions) >= maxCount {
				break
			}
		}
	}
	
	return suggestions
}

// getAllSuggestions returns all available commands as suggestions
func (scp *SlashCommandProcessor) getAllSuggestions(maxCount int) []Suggestion {
	var suggestions []Suggestion
	seen := make(map[string]bool)
	
	for _, cmd := range scp.commands {
		// Extract base command name
		name := cmd.ID
		if strings.HasPrefix(name, UserCommandPrefix) {
			name = strings.TrimPrefix(name, UserCommandPrefix)
		} else if strings.HasPrefix(name, ProjectCommandPrefix) {
			name = strings.TrimPrefix(name, ProjectCommandPrefix)
		}
		
		// Skip if we've already added this command name
		if seen[name] {
			continue
		}
		
		suggestion := Suggestion{
			Command:     name,
			Title:       cmd.Title,
			Description: cmd.Description,
		}
		suggestions = append(suggestions, suggestion)
		seen[name] = true
		
		// Limit results
		if len(suggestions) >= maxCount {
			break
		}
	}
	
	return suggestions
}

// AutofillCommand returns the completed command text if there's a unique match
func (scp *SlashCommandProcessor) AutofillCommand(input string) string {
	// Get more suggestions to check if there are multiple matches
	suggestions := scp.GetSuggestions(input, 10)
	if len(suggestions) == 1 {
		return "/" + suggestions[0].Command
	}
	return input // Return original if no matches or multiple matches
}

// GetCommonPrefix returns the longest common prefix among suggestions
func (scp *SlashCommandProcessor) GetCommonPrefix(input string) string {
	suggestions := scp.GetSuggestions(input, 100) // Get all suggestions
	if len(suggestions) == 0 {
		return input
	}
	
	if len(suggestions) == 1 {
		return "/" + suggestions[0].Command
	}
	
	// Find common prefix among all suggestions
	prefix := suggestions[0].Command
	for i := 1; i < len(suggestions); i++ {
		prefix = commonPrefix(prefix, suggestions[i].Command)
		if prefix == "" {
			break
		}
	}
	
	// Only return if the prefix is longer than current input
	currentInput := strings.TrimSpace(input)
	if len(currentInput) > 1 { // Remove "/"
		currentCommand := currentInput[1:]
		if len(prefix) > len(currentCommand) {
			return "/" + prefix
		}
	}
	
	return input // Return original if no improvement
}

// commonPrefix finds the common prefix of two strings
func commonPrefix(a, b string) string {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	
	for i := 0; i < minLen; i++ {
		if a[i] != b[i] {
			return a[:i]
		}
	}
	return a[:minLen]
}

// FormatSlashCommandError formats error messages for slash commands
func FormatSlashCommandError(err error, commandName string) string {
	if err == nil {
		return ""
	}

	baseMsg := err.Error()
	
	if strings.Contains(baseMsg, "command not found") {
		return fmt.Sprintf("Command '/%s' not found. Use Ctrl+K to see available commands or try with prefix: /user:%s or /project:%s", 
			commandName, commandName, commandName)
	}

	if strings.Contains(baseMsg, "no content available") {
		return fmt.Sprintf("Command '/%s' found but has no content. Check the command file.", commandName)
	}

	return fmt.Sprintf("Slash command error: %s", baseMsg)
}

// ValidateSlashCommand validates a slash command input for basic syntax
func ValidateSlashCommand(input string) error {
	trimmed := strings.TrimSpace(input)
	
	if !strings.HasPrefix(trimmed, "/") {
		return fmt.Errorf("slash commands must start with '/'")
	}
	
	if len(trimmed) == 1 {
		return fmt.Errorf("command name required after '/'")
	}
	
	// Extract command name
	parts := strings.SplitN(trimmed[1:], " ", 2)
	commandName := parts[0]
	
	if commandName == "" {
		return fmt.Errorf("command name cannot be empty")
	}
	
	// Check for invalid characters in command name
	if strings.ContainsAny(commandName, " \t\n\r") {
		return fmt.Errorf("command name cannot contain whitespace")
	}
	
	return nil
}