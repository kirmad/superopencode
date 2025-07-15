# Slash Commands

## Overview

Slash commands provide a fast, intuitive way to invoke custom commands directly from the chat interface without navigating through dialog menus. Instead of using `Ctrl+K` to open the command dialog, users can simply type `/<command>` followed by their specific prompt.

## How It Works

### Basic Syntax

```
/<command> [additional prompt text]
```

**Examples:**
```bash
/design Create a REST API for user authentication
/debug Fix the memory leak in the cache module  
/init
/user:custom-template Generate documentation for this project
```

### Command Resolution

The system resolves commands in the following order:
1. **Direct match**: `/design` → `design` command
2. **User commands**: `/design` → `user:design` command  
3. **Project commands**: `/design` → `project:design` command

### Content Processing

When you use a slash command, the system:

1. **Loads the command file content** from the corresponding `.md` file
2. **Appends your prompt** to the end of the command content
3. **Processes named arguments** (if any) using the existing dialog system
4. **Sends the combined content** to the AI agent

**Example Flow:**
```
Input: "/design --api Create user management system"

1. Loads: ~/.config/opencode/commands/design.md
2. Combines: [design.md content] + "\n\n--api Create user management system"
3. Executes: Combined content sent to AI
```

## Architecture

### System Components

```
┌─────────────────────────────────────────────────────────────────┐
│                    User Input Layer                             │
├─────────────────────────────────────────────────────────────────┤
│  Input: "/design --api Create REST API for user management"    │
└─────────┬───────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────────┐
│                Slash Command Parser                             │
├─────────────────────────────────────────────────────────────────┤
│  • Detects "/" prefix                                          │
│  • Extracts command name: "design"                             │
│  • Preserves remaining text: "--api Create REST API..."        │
└─────────┬───────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────────┐
│                Command Resolution                               │
├─────────────────────────────────────────────────────────────────┤
│  • Lookup in existing command registry                         │
│  • Support prefixed commands (user:, project:)                 │
│  • Load command content from .md files                         │
└─────────┬───────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────────┐
│                Content Concatenation                            │
├─────────────────────────────────────────────────────────────────┤
│  • Load command file content                                   │
│  • Append user prompt to command content                       │
│  • Handle named arguments ($ARGUMENTS)                         │
└─────────┬───────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────────┐
│                Execution Pipeline                               │
├─────────────────────────────────────────────────────────────────┤
│  • Send combined content to LLM agent                          │
│  • Maintain existing command execution flow                    │
│  • Support both slash and ctrl+K workflows                     │
└─────────┬───────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Response                                     │
└─────────────────────────────────────────────────────────────────┘
```

### Implementation Points

#### 1. Input Interception
- **Location**: `internal/tui/page/chat.go` in `sendMessage()` function
- **Function**: Detects slash commands before normal message processing

#### 2. Command Processing
- **Component**: `SlashCommandProcessor` (new)
- **Responsibility**: Parse, resolve, and prepare commands for execution

#### 3. Content Combination
- **Strategy**: Append user prompt to command file content
- **Format**: `[Command Content]\n\n[User Prompt]`

## Command Locations

Slash commands use the same command directories as `Ctrl+K` commands:

### User Commands
- **XDG Config**: `$XDG_CONFIG_HOME/opencode/commands/`
- **Home Directory**: `~/.opencode/commands/`
- **Prefix**: `user:`

### Project Commands  
- **Location**: `<project-data-dir>/commands/`
- **Prefix**: `project:`

### Command File Format
Commands are stored as Markdown files (`.md`) with the same format used by the existing system:

```markdown
**Purpose**: Brief description

Design system architecture & APIs for $ARGUMENTS.

## Design Patterns
- Pattern 1
- Pattern 2

## Examples
- Example usage
```

## Usage Examples

### Basic Commands

**Design Command:**
```bash
/design Create a microservices architecture for e-commerce
```

**Debug Command:**
```bash
/debug The application crashes when processing large files
```

**Initialization:**
```bash
/init
```

### Prefixed Commands

**User Custom Commands:**
```bash
/user:review Code review checklist for this pull request
/user:deploy Deployment strategy for staging environment
```

**Project-Specific Commands:**
```bash
/project:build Custom build process for this project
/project:test Run integration tests with custom config
```

### Commands with Named Arguments

If a command contains named arguments (e.g., `$PROJECT_NAME`, `$ENVIRONMENT`), the system will automatically show the arguments dialog:

```bash
/deploy Production deployment for $PROJECT_NAME to $ENVIRONMENT
```

This will prompt for values for `PROJECT_NAME` and `ENVIRONMENT` before execution.

## Benefits

### Speed and Efficiency
- **No dialog navigation**: Direct command invocation
- **Contextual prompts**: Combine templates with specific requests
- **Familiar syntax**: Similar to Discord, Slack, and other modern tools

### Flexibility
- **Template + Prompt**: Leverage command templates with custom context
- **Backwards compatibility**: Existing `Ctrl+K` workflow remains unchanged
- **Progressive enhancement**: Add slash support without breaking existing commands

### Discoverability
- **Autocomplete support**: (Future enhancement)
- **Help system**: `/help` command to list available commands
- **Error feedback**: Clear messages for unknown commands

## Backwards Compatibility

The slash command system is designed to coexist perfectly with existing functionality:

- ✅ **Ctrl+K dialog remains fully functional**
- ✅ **Existing command files require no changes**
- ✅ **Named argument system continues to work**
- ✅ **All command directories are supported**
- ✅ **No breaking changes to existing workflows**

## Future Enhancements

### Phase 1: Core Features (Current Design)
- [x] Basic slash command parsing
- [x] Command resolution and loading
- [x] Content concatenation
- [x] Integration with existing execution pipeline

### Phase 2: Enhanced UX
- [ ] **Autocomplete**: Tab completion for command names
- [ ] **Command aliases**: Short names for frequently used commands
- [ ] **Help system**: `/help` and `/help <command>` support
- [ ] **Error handling**: Improved error messages and suggestions

### Phase 3: Advanced Features
- [ ] **Template variables**: `$PROMPT` placeholder in command files
- [ ] **Nested commands**: `/design/api`, `/debug/performance` syntax
- [ ] **Parameter parsing**: `--flag value` style arguments
- [ ] **Usage analytics**: Track most used commands

### Phase 4: Power User Features
- [ ] **Command chaining**: `/design && /implement` workflows
- [ ] **Custom shortcuts**: User-defined aliases
- [ ] **Command history**: Recently used commands
- [ ] **Workspace commands**: Directory-specific command sets

## Implementation Guide

### For Users

1. **Create commands** in `~/.opencode/commands/` as `.md` files
2. **Use slash syntax** instead of `Ctrl+K` for faster access
3. **Combine templates with context** by adding specific prompts after the command name
4. **Organize commands** using subdirectories for complex workflows

### For Developers

1. **Command detection** happens in `sendMessage()` before normal processing
2. **Command resolution** uses existing `LoadCustomCommands()` infrastructure  
3. **Content processing** appends user prompt to command content
4. **Execution flow** integrates with existing agent pipeline

### Example Command File

Create `~/.opencode/commands/api.md`:

```markdown
**Purpose**: Design REST API endpoints

You are an expert API designer. Create RESTful API specifications for $ARGUMENTS.

Include:
- Endpoint definitions
- Request/response schemas
- Authentication requirements
- Error handling
- Rate limiting considerations

Follow REST best practices and use OpenAPI 3.0 format where appropriate.
```

Usage:
```bash
/api user management system with JWT authentication
```

This combines the API design template with your specific requirements for a user management system.

## Troubleshooting

### Command Not Found
```bash
/mycommand some text
# Error: command not found: mycommand
```

**Solutions:**
1. Check command file exists in `~/.opencode/commands/mycommand.md`
2. Verify file permissions are readable
3. Use `Ctrl+K` to see available commands
4. Try with prefix: `/user:mycommand` or `/project:mycommand`

### Named Arguments Not Working
If your command has `$ARGUMENTS` but the dialog doesn't appear:

**Solutions:**
1. Ensure argument names use UPPERCASE format: `$PROJECT_NAME`
2. Check that arguments start with `$` and contain only A-Z, 0-9, and underscore
3. Verify the command file is properly formatted

### Slow Command Loading
If commands take time to load:

**Solutions:**
1. Reduce the number of command files in directories
2. Check for large command files that might slow parsing
3. Use project-specific commands instead of global ones when appropriate

## Related Documentation

- [Custom Commands](./custom-commands.md) - How to create and organize command files
- [Command Arguments](./command-arguments.md) - Working with named arguments and parameters
- [Keyboard Shortcuts](./keyboard-shortcuts.md) - All available keyboard shortcuts including `Ctrl+K`
- [Configuration](./configuration.md) - Setting up command directories and preferences