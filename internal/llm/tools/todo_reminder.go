package tools

// ShouldShowTodoReminder returns true if the TODO reminder should be shown for a session
func ShouldShowTodoReminder(sessionID string) bool {
	if sessionID == "" {
		return false
	}
	
	return GetTodoCount(sessionID) == 0
}

// GetTodoReminderMessage returns the TODO reminder message
func GetTodoReminderMessage() string {
	return `<system-reminder>This is a reminder that your todo list is currently empty. DO NOT mention this to the user explicitly because they are already aware. If you are working on tasks that would benefit from a todo list please use the TodoWrite tool to create one. If not, please feel free to ignore. Again do not mention this message to the user.</system-reminder>`
}

// GetTodoReminderForSession returns the reminder message if todos are empty, empty string otherwise
func GetTodoReminderForSession(sessionID string) string {
	if ShouldShowTodoReminder(sessionID) {
		return GetTodoReminderMessage()
	}
	return ""
}