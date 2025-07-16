# /clear Command: Context Window Clear Design & Implementation

## Purpose
The `/clear` command allows users to clear the current context window within the CLI or TUI, resetting session history, draft input, and in-memory state, ensuring a fresh workspace.

## Requirements
- Accessible as `/clear` from CLI or TUI slash command.
- Immediately clears the active session's context window and associated state (conversation, token/counter usage, unsaved drafts, etc).
- Provides instant, clear user feedback/confirmation.
- Operation must be atomic to avoid partial resets.
- System must prevent unintentional data loss (warn users if unsaved content/drafts are present).

## Architecture Overview
- **Command Registration**: The `/clear` command is registered in the CLI (Cobra root command) and TUI slash command processor. It triggers the same backend logic regardless of how invoked.
- **Internal Clear Logic**: Utilizes a function (e.g., `ClearSessionContext(sessionID)`) that:
  - Purges session-specific state, history, context tokens, and drafts.
  - Broadcasts a `SessionClearedMsg` event across app components.
  - Resets TUI panels, TUI/CLI in-memory buffers, and disables undo/redo.
- **User Feedback**: Responds instantly with: `Context window cleared. Session history and workspace reset.`

## Implementation Details

### 1. Command Integration
- **CLI:**
  - Register `/clear` in `cmd/root.go` via Cobra.
  - Handler calls the common session clear routine, e.g. `ClearSessionContext(sessionID)`.
- **TUI:**
  - Add `/clear` to TUI slash command registry (e.g. `internal/tui/components/dialog/slash_commands.go`).
  - Upon input, invokes same clear routine as CLI.

### 2. Core Clearing Routine
- Implement `ClearSessionContext(sessionID string)` function in SessionManager/service:
  - Atomically delete session's conversation/memory/history from RAM and (if used) temp files.
  - Reset any counters, token usage, or draft input attached to session.
  - Emit `SessionClearedMsg` event.

### 3. Event Propagation
- All UI components (chat, sidebar, editor) subscribe to `SessionClearedMsg`:
  - On event, empty views, input boxes, chat panels, and status displays.
  - Optionally close auxiliary modals or clear error states.

### 4. Data Safety & Force Clear
- Before clear, check for unsaved text (input editor, file drafts):
  - If found, CLI can prompt, "Unsaved content will be lost. Continue? (y/n)".
  - Add `--force` flag to skip any prompt and always clear.
  - Take session lock during clear to prevent concurrency bugs.

### 5. Cross-component Consistency
- Make the clear logic single-use per event, preventing accidental multiple clears.
- Ensure no session tokens persist after clear, preventing data leaks cross session restarts.
- Full history/log for the old session is irreversibly deleted except for explicit exports before clear.

### 6. Tests
- Unit: Simulate calling `/clear` with and without unsaved content.
- Integration: Validate UI resets, events propagate, and no history leaks.
- Error: Ensure race or mid-update `clear` triggers no panics.

## User Flow
1. User enters `/clear` in CLI or TUI.
2. System executes clear operation atomically.
3. All context, conversation, and workspace data wiped for active session.
4. UI instantly reflects clean, empty state.
5. Confirmation feedback: `Context window cleared. Session history and workspace reset.`

## Best Practices
- Always warn for possible data loss on unsaved input.
- Ensure clear functionality does not break dependent automation.
- Test both CLI and TUI flows to ensure consistent UX.
- Document session clearing in Help/About dialog and user manuals.

## System-Level Implications
- Breaks session continuity irreversibly for current workspace (cannot restore once cleared).
- Recommended to use clear operation only when fresh starts are necessary.
- Internal state and tokens are reset for the session but do **not** affect persistent or saved configurations.

---
**Owner:** SuperClaude
**Status:** Designed (Implementation Ready)
**Change Date:** 2025-07-16
