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
	return `<system-reminder>
Your todo list is currently empty. DO NOT mention this explicitly to the user as they may already be aware. If you are working on tasks that would benefit from progress tracking, please use the TodoWrite tool to create a todo list. This helps maintain organization and ensures important tasks are completed. Continue with your current work if todo tracking is not applicable.
</system-reminder>`
}

// GetTodoReminderForSession returns the reminder message if todos are empty, empty string otherwise
func GetTodoReminderForSession(sessionID string) string {
	if ShouldShowTodoReminder(sessionID) {
		return GetTodoReminderMessage()
	}
	return ""
}