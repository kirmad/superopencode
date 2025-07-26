# SuperOpenCode Code Structure and Architecture

## Project Structure

```
superopencode/
├── main.go                    # Application entry point
├── cmd/                       # CLI command definitions
│   ├── root.go               # Root cobra command
│   └── schema/               # Schema generation
├── internal/                  # Internal application code
│   ├── app/                  # Application coordination and lifecycle
│   ├── config/               # Configuration management
│   ├── db/                   # Database operations and migrations
│   ├── llm/                  # LLM provider integration
│   │   ├── agent/           # Agent system (coder, task, etc.)
│   │   ├── models/          # Model definitions and registry
│   │   ├── provider/        # Provider implementations
│   │   ├── prompt/          # Prompt templates
│   │   └── tools/           # Tool implementations
│   ├── tui/                  # Terminal User Interface
│   │   ├── components/      # Reusable UI components
│   │   ├── layout/          # Layout managers
│   │   ├── page/            # Page implementations
│   │   └── theme/           # Theme system
│   ├── lsp/                  # Language Server Protocol integration
│   ├── session/              # Session management
│   ├── message/              # Message handling and storage
│   ├── permission/           # Permission system
│   ├── history/              # File history and versioning
│   ├── logging/              # Logging infrastructure
│   ├── pubsub/              # Event system
│   ├── format/              # Output formatting
│   ├── version/             # Version information
│   ├── fileutil/            # File utilities
│   ├── diff/                # Diff utilities
│   ├── completions/         # Shell completions
│   └── detailed_logging/    # Detailed logging system
├── docs/                     # Documentation
├── scripts/                  # Build and utility scripts
├── .github/workflows/        # CI/CD workflows
└── logs/                     # Log files (gitignored)
```

## Core Architecture Patterns

### Service-Oriented Architecture
The application uses a service pattern where each major feature is implemented as a service with a well-defined interface:

```go
type Service interface {
    // Methods specific to each service
}
```

**Example Services:**
- `session.Service` - Session management
- `message.Service` - Message storage and retrieval
- `agent.Service` - AI agent coordination
- `permission.Service` - Permission management

### Dependency Injection
Services are injected into components that need them, typically through constructor functions:

```go
type App struct {
    Sessions    session.Service
    Messages    message.Service
    History     history.Service
    Permissions permission.Service
    CoderAgent  agent.Service
    // ...
}
```

### Event-Driven Architecture
Uses pub/sub pattern for real-time updates across components:

```go
type Broker[T any] struct {
    subs map[chan Event[T]]struct{}
    // ...
}
```

### Tool System
Standardized tool interface for AI capabilities:

```go
type BaseTool interface {
    Info() ToolInfo
    Run(ctx context.Context, params ToolCall) (ToolResponse, error)
}
```

## Key Design Principles
1. **Modular Design** - Clear separation of concerns
2. **Interface-Based** - Heavy use of interfaces for testability
3. **Context-Aware** - Proper context.Context usage throughout
4. **Error Handling** - Comprehensive error handling with context
5. **Type Safety** - Strong typing with minimal use of `any`
6. **Concurrency Safe** - Proper mutex usage and goroutine management