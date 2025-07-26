# SuperOpenCode Internal Architecture

## Core Modules

### Application Layer (`internal/app`)
- `app.go`: Main application service coordination
- `lsp.go`: Language Server Protocol integration
- Manages core services and their lifecycles

### Configuration (`internal/config`)
- `config.go`: Configuration management and loading
- `init.go`: Initialization routines
- Supports JSON config files in multiple locations:
  - `$HOME/.opencode.json`
  - `$XDG_CONFIG_HOME/opencode/.opencode.json`
  - `./.opencode.json` (local directory)

### Database Layer (`internal/db`)
- **Technology**: SQLite with goose migrations
- **Key Files**:
  - `connect.go`: Database connection management
  - `db.go`: Core database operations
  - `models.go`: Data models
  - SQL query files: `files.sql.go`, `messages.sql.go`, `sessions.sql.go`
- **Migrations**: Located in `migrations/` directory
- **Generated Code**: Uses sqlc for type-safe SQL

### LLM Integration (`internal/llm`)
- **Agent System** (`agent/`): AI agent coordination and tool management
- **Models** (`models/`): Support for multiple providers (Anthropic, OpenAI, Gemini, etc.)
- **Prompts** (`prompt/`): Prompt engineering and management
- **Providers** (`provider/`): Provider-specific implementations
- **Tools** (`tools/`): AI-accessible tools (bash, file operations, diagnostics, etc.)

### Terminal UI (`internal/tui`)
- **Framework**: Charm's Bubble Tea
- **Components**: Chat interface, logs viewer, dialogs
- **Layout**: Container-based layout system
- **Themes**: Multiple theme support (Catppuccin, Dracula, Gruvbox, etc.)
- **Features**: Chat, file picker, help system, session management

### Logging System (`internal/logging`)
- **Detailed Logging**: Advanced logging for LLM interactions
- **Components**:
  - `detailed.go`: Detailed logging implementation
  - `interceptor.go`: HTTP request/response interception
  - `provider.go`: Logging provider abstraction
  - `storage.go`: Log data storage
  - `tool_tracker.go`: Tool usage tracking

### Language Server Protocol (`internal/lsp`)
- **Client Implementation**: Full LSP client
- **Supported Features**: Diagnostics, file watching, notifications
- **Protocol**: Standard LSP implementation with Go-specific extensions
- **Integration**: Exposes diagnostics to AI assistant

## Key Dependencies
- **Bubble Tea**: TUI framework
- **Cobra**: CLI framework
- **SQLite**: Database (ncruces/go-sqlite3)
- **Goose**: Database migrations
- **Various AI SDKs**: anthropic-sdk-go, openai-go, etc.
- **MCP**: mark3labs/mcp-go for Model Context Protocol