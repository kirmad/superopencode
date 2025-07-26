# --dangerously-skip-permissions Flag Design

## Overview

Design specification for adding a `--dangerously-skip-permissions` flag that bypasses permission validation checks for tool calls by leveraging the existing `AutoApproveSession()` functionality.

## Real Implementation Analysis

**Existing Permission System**:
- **Location**: `internal/permission/permission.go`
- **Auto-Bypass**: `AutoApproveSession(sessionID string)` already exists
- **Usage**: Currently used for non-interactive sessions in `internal/app/app.go`
- **Tools Integration**: All tools use `permissions.Request()` for validation

**Current Auto-Approval Logic**:
```go
func (s *permissionService) Request(opts CreatePermissionRequest) bool {
    if slices.Contains(s.autoApproveSessions, opts.SessionID) {
        return true  // Bypass all permission checks
    }
    // Normal permission flow...
}
```

## Requirements

- **Core Functionality**: Leverage existing `AutoApproveSession()` for tool permission bypass
- **Simple Implementation**: Add command-line flag that calls existing auto-approval
- **Explicit Intent**: Flag name clearly indicates dangerous operation
- **Existing Parity**: Use the same mechanism as non-interactive sessions

## Flag Interface

### Command Line Flag
```bash
--dangerously-skip-permissions
```

### Behavior Specification
When this flag is active, the session is automatically added to `autoApproveSessions`, bypassing ALL tool permission checks:

- **All Tools**: Write, Edit, Bash, and all other tools skip permission validation
- **Session-Based**: Applies to entire session, not per-tool-call
- **Existing Mechanism**: Uses same auto-approval as non-interactive sessions

### Usage Context
- Development environments where permission prompts interfere with workflows
- Administrative operations requiring elevated access
- Automated scripts where confirmation prompts aren't feasible

## Security Implications

### Security Risks
- **Data Loss**: All Write/Edit operations bypass validation → potential overwrite of critical files
- **System Impact**: Bash commands bypass safety filtering → potential system damage  
- **Unrestricted Access**: All file operations bypass validation → access to any files/directories
- **No User Confirmation**: All tools execute without permission prompts

### Safety Guards (Minimal)
- **Clear Naming**: "dangerously" prefix communicates risk
- **Explicit Declaration**: Must be explicitly specified, no default activation
- **Existing Safety**: Bash tool still blocks dangerous commands (curl, wget, etc.)
- **Warning Messages**: Display clear warnings when auto-approval is active

## Implementation Specification

### Core Design Pattern
Use existing `AutoApproveSession()` mechanism triggered by command-line flag.

### Implementation Steps

1. **Command-Line Flag Parsing**
   - Add `--dangerously-skip-permissions` flag to CLI parsing
   - Store flag state for session creation

2. **Session Auto-Approval**
   - Call `app.Permissions.AutoApproveSession(sessionID)` when flag is set
   - Display warning message when auto-approval is activated

3. **No Tool Modifications Required**
   - Existing permission system handles bypass automatically
   - All tools already use `permissions.Request()` which checks auto-approval

### Real Implementation

```go
// In app initialization (where session is created)
if dangerouslySkipPermissions {
    log.Warn("⚠️ DANGEROUS: --dangerously-skip-permissions active. All tool permissions bypassed.")
    app.Permissions.AutoApproveSession(sessionID)
}
```

**Key Files to Modify**:
- Command-line parsing (likely `cmd/root.go`)
- Session creation in app initialization
- Warning display system

### Warning System

```yaml
warning_display:
  message: "⚠️ DANGEROUS: --dangerously-skip-permissions active. All tool permissions bypassed for this session."
  timing: "Session creation when flag is detected"
  
logging:
  - "Log auto-approval activation with timestamp and session ID"
  - "Existing permission system already logs bypassed requests"
```

## Implementation Complexity

**Estimated Changes**: ~5-10 lines of code (much simpler than originally thought)

**Files to Modify**:
- `cmd/root.go` - Add command-line flag
- Session creation logic - Call `AutoApproveSession()` when flag set
- Warning display system

**No Changes Needed**:
- Tool implementations (already use existing permission system)
- Permission validation logic (already has auto-approval built-in)

## Testing Considerations

### Test Cases
1. **Flag Activation**: Verify flag triggers `AutoApproveSession()` call
2. **Warning Display**: Confirm warning shown at session creation
3. **Permission Bypass**: Test that all tool permission requests return `true`
4. **Session Scope**: Verify bypass applies to entire session, not just first tool
5. **Normal Mode**: Ensure normal permission checking when flag not used

### Security Testing
- Verify bypass only works when flag is explicitly provided
- Test that auto-approval is session-specific
- Confirm existing Bash command filtering still works
- Validate warning messages are clear and prominent

## Design Decisions

### Key Design Choices
- ✅ **Leverage Existing**: Use proven `AutoApproveSession()` mechanism
- ✅ **Clear Intent**: "dangerously" prefix communicates risk effectively  
- ✅ **Minimal Changes**: Only ~5-10 lines of new code required
- ✅ **Session-Based**: Consistent with existing non-interactive session behavior

### Trade-offs
- **Security vs Functionality**: Chose functionality using existing safety patterns
- **Complexity vs Simplicity**: Leveraged existing infrastructure for maximum simplicity
- **User Safety vs Developer Convenience**: Balanced through explicit naming and warnings

## Future Considerations

### Potential Enhancements
- **Tool-Specific Bypass**: Modify `AutoApproveSession()` to accept tool filters
- **Enhanced Audit Trail**: Additional logging for auto-approved sessions
- **Time-based Expiration**: Auto-disable auto-approval after timeout
- **Confirmation for Dangerous Commands**: Additional prompts for high-risk Bash commands

### Migration Path
Implementation leverages existing infrastructure, making enhancements straightforward without breaking core functionality.

## Conclusion

The `--dangerously-skip-permissions` flag provides the requested functionality with:
- **Existing Infrastructure**: Leverages proven `AutoApproveSession()` mechanism
- **Clear Intent**: Explicit naming indicates dangerous operations
- **Minimal Implementation**: Only ~5-10 lines of new code required
- **Consistent Behavior**: Same mechanism as non-interactive sessions

This design provides maximum effectiveness with minimal complexity by using existing permission bypass infrastructure.