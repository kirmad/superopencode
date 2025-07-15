# OpenCode Developer Documentation

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Core Modules](#core-modules)
3. [Configuration System](#configuration-system)
4. [Database Architecture](#database-architecture)
5. [LLM Provider Integration](#llm-provider-integration)
6. [Terminal User Interface (TUI)](#terminal-user-interface-tui)
7. [Tool System and Agent Integration](#tool-system-and-agent-integration)
8. [LSP and MCP Integrations](#lsp-and-mcp-integrations)
9. [Development Guidelines](#development-guidelines)
10. [Extension Patterns](#extension-patterns)

---

## Architecture Overview

OpenCode is a Go-based CLI application that brings AI assistance directly to your terminal. It provides a sophisticated Terminal User Interface (TUI) for interacting with various AI models to help with coding tasks, debugging, and development workflows.

### Core Components

```
┌─────────────────────────────────────────────────────────────┐
│                    OpenCode Architecture                     │
├─────────────────┬─────────────────┬─────────────────────────┤
│   Terminal UI   │      CLI        │      Background         │
│   (Bubble Tea)  │   (Non-TUI)     │      Services           │
├─────────────────┼─────────────────┼─────────────────────────┤
│ • Chat Interface│ • Single Prompt │ • LSP Clients          │
│ • File Browser  │ • JSON Output   │ • File Watchers        │
│ • Logs View     │ • Auto-approve  │ • MCP Servers          │
│ • Dialogs       │   permissions   │ • Database Pool        │
├─────────────────┴─────────────────┼─────────────────────────┤
│           Application Core        │                         │
├─────────────────┬─────────────────┼─────────────────────────┤
│   Agent System  │  Tool System    │     Data Layer          │
├─────────────────┼─────────────────┼─────────────────────────┤
│ • Coder Agent   │ • File Tools    │ • SQLite Database       │
│ • Task Agent    │ • Search Tools  │ • Session Management    │
│ • Summarizer    │ • Shell Tools   │ • Message Storage       │
│ • Title Agent   │ • LSP Tools     │ • File Versioning      │
│                 │ • MCP Tools     │                         │
├─────────────────┴─────────────────┼─────────────────────────┤
│         LLM Provider Layer        │                         │
├───────────────────────────────────┼─────────────────────────┤
│ • Anthropic    • OpenAI          │    External Services     │
│ • GitHub       • Google          ├─────────────────────────┤
│ • Groq         • Azure           │ • Language Servers      │
│ • OpenRouter   • AWS             │ • MCP Servers           │
│ • Local        • VertexAI        │ • External Editors      │
└───────────────────────────────────┴─────────────────────────┘
```

### Key Features

- **Interactive TUI**: Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for smooth terminal experience
- **Multiple AI Providers**: Supports 10+ LLM providers with unified interface
- **Session Management**: Persistent conversation storage with cost tracking
- **Tool Integration**: AI can execute commands, search files, and modify code
- **LSP Integration**: Real-time code diagnostics and language intelligence
- **MCP Support**: Extensible tool system through Model Control Protocol
- **File Versioning**: Track changes across development sessions
- **Permission System**: Secure tool execution with granular controls

### Entry Points

- **Interactive Mode**: `opencode` - Full TUI with all features
- **CLI Mode**: `opencode -p "prompt"` - Single prompt execution
- **Debug Mode**: `opencode -d` - Enable debug logging and diagnostics

---

## Core Modules

### Application Structure (`internal/app/`)

The application module serves as the central orchestrator, coordinating all services and managing the application lifecycle.

**Key Responsibilities:**
- Service initialization and dependency injection
- LSP client lifecycle management with automatic restarts
- Non-interactive mode execution for CLI usage
- Theme management and configuration
- Graceful shutdown with proper cleanup

**Service Dependencies:**
```go
type App struct {
    Sessions    session.Service
    Messages    message.Service
    History     history.Service
    Permissions permission.Service
    CoderAgent  agent.Service
    LSPClients  map[string]*lsp.Client
    Theme       theme.Theme
}
```

**LSP Client Management:**
- Background initialization to prevent blocking startup
- Workspace watchers for real-time file change detection
- Automatic restart on configuration changes
- Language-specific optimization strategies

### Configuration System (`internal/config/`)

Sophisticated multi-source configuration management with smart defaults and provider detection.

**Configuration Sources (Priority Order):**
1. Command-line flags
2. Environment variables (prefixed with `OPENCODE_`)
3. Local config file (`.opencode.json` in working directory)
4. Global config file (`~/.opencode.json` or XDG config directories)
5. Default values

**Key Features:**
- **Provider Auto-Detection**: Automatically configures models based on available API keys
- **Model Validation**: Validates model IDs and provider compatibility
- **Fallback Defaults**: Intelligent fallbacks when configurations are invalid
- **Runtime Updates**: Dynamic configuration updates (theme, models) with persistence

**Configuration Structure:**
```json
{
  "data": { "directory": ".opencode" },
  "providers": {
    "openai": { "apiKey": "sk-...", "disabled": false },
    "anthropic": { "apiKey": "sk-...", "disabled": false }
  },
  "agents": {
    "coder": { "model": "claude-4-sonnet", "maxTokens": 5000 },
    "task": { "model": "gpt-4.1-mini", "maxTokens": 3000 }
  },
  "lsp": {
    "gopls": { "command": "gopls", "args": ["serve"], "disabled": false }
  },
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path"],
      "type": "stdio"
    }
  },
  "shell": { "path": "/bin/bash", "args": ["-l"] },
  "autoCompact": true,
  "debug": false
}
```

### Database Architecture (`internal/db/`)

SQLite-based persistence layer with type-safe query generation and sophisticated versioning.

**Technology Stack:**
- **SQLite**: Embedded database with WAL mode for concurrency
- **SQLC**: Type-safe SQL code generation
- **Goose**: Database migrations
- **Foreign Keys**: Referential integrity with cascade deletes

**Core Schema:**
```sql
-- Sessions with hierarchical relationships
sessions (
    id TEXT PRIMARY KEY,
    parent_session_id TEXT,     -- Enables session trees
    title TEXT NOT NULL,
    message_count INTEGER DEFAULT 0,
    prompt_tokens INTEGER DEFAULT 0,
    completion_tokens INTEGER DEFAULT 0,
    cost REAL DEFAULT 0.0,
    summary_message_id TEXT,
    updated_at INTEGER NOT NULL,
    created_at INTEGER NOT NULL
)

-- Messages with polymorphic content
messages (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    role TEXT NOT NULL,         -- user/assistant/system
    parts TEXT DEFAULT '[]',    -- JSON array of content parts
    model TEXT,
    created_at INTEGER NOT NULL,
    finished_at INTEGER,        -- Completion timestamp
    FOREIGN KEY (session_id) REFERENCES sessions (id) ON DELETE CASCADE
)

-- File versioning system
files (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    path TEXT NOT NULL,
    content TEXT NOT NULL,
    version TEXT NOT NULL,      -- "initial", "v1", "v2", etc.
    created_at INTEGER NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions (id) ON DELETE CASCADE,
    UNIQUE(path, session_id, version)
)
```

**File Versioning Strategy:**
- Initial version: `"initial"` for first creation
- Subsequent versions: `"v1"`, `"v2"`, `"v3"`, etc.
- Complex queries for latest version retrieval per session
- Content deduplication and change tracking

**Performance Optimizations:**
```go
// SQLite configuration for optimal performance
pragmas := []string{
    "PRAGMA foreign_keys = ON;",
    "PRAGMA journal_mode = WAL;",     // Write-Ahead Logging
    "PRAGMA page_size = 4096;",
    "PRAGMA cache_size = -8000;",     // 8MB cache
    "PRAGMA synchronous = NORMAL;",
}
```

### Logging System (`internal/logging/`)

Structured logging with persistence and real-time display capabilities.

**Features:**
- **Multi-level logging**: Debug, Info, Warn, Error
- **Persistent logging**: Status bar display and file-based debug logs
- **Panic recovery**: Detailed stack traces with session context
- **Real-time streaming**: Pub/sub integration for live log display in TUI

**Integration Points:**
- Session-aware debug file organization
- Real-time log display in TUI logs page
- Structured logging with contextual attributes
- File-based logging when `OPENCODE_DEV_DEBUG=true`

---

## Configuration System

### Multi-Source Configuration Loading

The configuration system implements a sophisticated priority-based loading mechanism:

```go
func Load(workingDir string, debug bool) (*Config, error) {
    // 1. Configure Viper with paths and environment
    configureViper()
    
    // 2. Set defaults
    setDefaults(debug)
    
    // 3. Read global config
    viper.ReadInConfig()
    
    // 4. Merge local config
    mergeLocalConfig(workingDir)
    
    // 5. Set provider defaults based on environment
    setProviderDefaults()
    
    // 6. Validate and apply defaults
    Validate()
}
```

### Provider Detection and Defaults

The system automatically detects available LLM providers and sets intelligent defaults:

**Provider Priority Order:**
1. GitHub Copilot (if token available)
2. Anthropic Claude
3. OpenAI
4. Google Gemini
5. Groq
6. OpenRouter
7. AWS Bedrock
8. Azure OpenAI
9. Google Cloud VertexAI

**Authentication Detection:**
```go
// Environment variable detection
if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
    viper.SetDefault("providers.anthropic.apiKey", apiKey)
}

// GitHub Copilot token detection
if token, err := LoadGitHubToken(); err == nil && token != "" {
    // Searches standard locations:
    // - ~/.config/github-copilot/hosts.json
    // - ~/.config/github-copilot/apps.json
}

// AWS credentials detection
func hasAWSCredentials() bool {
    // Checks multiple AWS credential sources
    // Environment variables, profiles, instance profiles
}
```

### Agent Configuration

Agents are configured with model-specific parameters:

```go
type Agent struct {
    Model           models.ModelID `json:"model"`
    MaxTokens       int64          `json:"maxTokens"`
    ReasoningEffort string         `json:"reasoningEffort"` // For reasoning models
}
```

**Agent Types:**
- **Coder**: Main development agent with full tool access
- **Task**: Limited agent for specific analysis tasks
- **Summarizer**: Conversation summarization agent
- **Title**: Session title generation agent

### Runtime Configuration Updates

The system supports dynamic configuration updates:

```go
func UpdateAgentModel(agentName AgentName, modelID models.ModelID) error {
    // 1. Validate new model
    // 2. Update in-memory configuration
    // 3. Validate agent with new model
    // 4. Persist to configuration file
    // 5. Rollback on failure
}

func UpdateTheme(themeName string) error {
    // 1. Update in-memory theme
    // 2. Persist to configuration file
    // 3. Broadcast theme change event
}
```

---

## Database Architecture

### SQLC Integration

The database layer uses SQLC for type-safe SQL code generation:

```sql
-- name: CreateSession :one
INSERT INTO sessions (id, title, parent_session_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetSessionByID :one
SELECT * FROM sessions WHERE id = ? LIMIT 1;

-- name: ListLatestSessionFiles :many
SELECT f.*
FROM files f
INNER JOIN (
    SELECT path, MAX(created_at) as max_created_at
    FROM files
    WHERE session_id = ?
    GROUP BY path
) latest ON f.path = latest.path AND f.created_at = latest.max_created_at
ORDER BY f.path;
```

Generated Go code provides type safety:
```go
type Querier interface {
    CreateSession(ctx context.Context, arg CreateSessionParams) (Session, error)
    GetSessionByID(ctx context.Context, id string) (Session, error)
    ListLatestSessionFiles(ctx context.Context, sessionID string) ([]File, error)
}
```

### Transaction Management

Service layer handles transaction boundaries:

```go
func (s *Service) CreateFileWithHistory(ctx context.Context, sessionID, path, content string) (File, error) {
    return s.db.WithTx(ctx, func(tx *db.Queries) (File, error) {
        // 1. Get latest version
        version := s.getNextVersion(ctx, tx, sessionID, path)
        
        // 2. Create file with retry logic
        file, err := tx.CreateFile(ctx, CreateFileParams{
            ID:        uuid.New().String(),
            SessionID: sessionID,
            Path:      path,
            Content:   content,
            Version:   version,
            CreatedAt: time.Now().Unix(),
        })
        
        // 3. Handle version conflicts with retry
        if isVersionConflict(err) {
            return s.retryFileCreation(ctx, tx, params)
        }
        
        return file, err
    })
}
```

### Database Triggers

Automatic maintenance through SQL triggers:

```sql
-- Automatic timestamp updates
CREATE TRIGGER update_sessions_updated_at
AFTER UPDATE ON sessions
BEGIN
    UPDATE sessions SET updated_at = strftime('%s', 'now')
    WHERE id = new.id;
END;

-- Message count maintenance
CREATE TRIGGER update_session_message_count_on_insert
AFTER INSERT ON messages
BEGIN
    UPDATE sessions SET message_count = message_count + 1
    WHERE id = new.session_id;
END;
```

### Performance Optimizations

```go
// Strategic indexing for query performance
"CREATE INDEX IF NOT EXISTS idx_files_session_id ON files(session_id);",
"CREATE INDEX IF NOT EXISTS idx_files_path ON files(path);",
"CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id);",

// Connection pool configuration
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

---

## LLM Provider Integration

### Provider Architecture

The LLM integration uses a generic base provider pattern for type safety:

```go
type Provider interface {
    SendMessages(ctx context.Context, messages []message.Message, tools []tools.BaseTool) (*ProviderResponse, error)
    StreamResponse(ctx context.Context, messages []message.Message, tools []tools.BaseTool) <-chan ProviderEvent
    Model() models.Model
}

type baseProvider[C ProviderClient] struct {
    options providerClientOptions
    client  C
}
```

### Model Registry

Comprehensive model definitions with capabilities:

```go
type Model struct {
    ID                  ModelID       // Unique identifier
    Name                string        // Human-readable name
    Provider            ModelProvider // Provider hosting this model
    APIModel            string        // Provider-specific model ID
    CostPer1MIn         float64       // Input token cost per million
    CostPer1MOut        float64       // Output token cost per million
    ContextWindow       int64         // Maximum context window
    DefaultMaxTokens    int64         // Default output limit
    CanReason           bool          // Supports reasoning/thinking
    SupportsAttachments bool          // File/image support
}
```

**Provider Support Matrix:**

| Provider | Models | Features |
|----------|--------|----------|
| Anthropic | Claude 3.5-4 | Reasoning, Attachments, Large context |
| OpenAI | GPT-4.1, o1/o3 | Reasoning, Attachments, Function calling |
| GitHub Copilot | Multiple | Enterprise auth, Various models |
| Google Gemini | 2.0/2.5 | Large context, Multimodal |
| Groq | Llama, QWEN | Fast inference, Open source |
| Azure OpenAI | Enterprise GPT | Enterprise features, Compliance |
| AWS Bedrock | Claude via AWS | AWS integration, Security |
| OpenRouter | Multiple | Model marketplace, Unified API |

### Streaming Response Handling

Event-driven streaming architecture:

```go
type ProviderEvent struct {
    Type EventType
    Content  string           // Text content
    Thinking string           // Reasoning content (Claude, o1)
    Response *ProviderResponse
    ToolCall *message.ToolCall
    Error    error
}

// Event types for different streaming phases
const (
    EventContentStart EventType = "content_start"
    EventContentDelta EventType = "content_delta"
    EventThinkingDelta EventType = "thinking_delta"
    EventToolUseStart EventType = "tool_use_start"
    EventComplete     EventType = "complete"
    EventError        EventType = "error"
)
```

### Error Handling and Retry Logic

Sophisticated retry mechanism with exponential backoff:

```go
func shouldRetry(attempts int, err error) (bool, int64, error) {
    if attempts > maxRetries {
        return false, 0, fmt.Errorf("maximum retry attempts reached")
    }
    
    // Exponential backoff with jitter
    backoffMs := 2000 * (1 << (attempts - 1))
    jitterMs := int(float64(backoffMs) * 0.2)
    retryMs := backoffMs + jitterMs
    
    // Honor Retry-After headers for rate limiting
    if retryAfter := parseRetryAfterHeader(err); retryAfter > 0 {
        retryMs = retryAfter
    }
    
    return true, int64(retryMs), nil
}
```

### Authentication Patterns

**API Key-Based Authentication:**
```go
// Standard API key configuration
ANTHROPIC_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
GEMINI_API_KEY=...
```

**OAuth Token Exchange (GitHub Copilot):**
```go
func exchangeGitHubToken(githubToken string) (string, error) {
    // Exchange GitHub OAuth token for Copilot bearer token
    // Uses: https://api.github.com/copilot_internal/v2/token
    resp, err := http.Post(tokenURL, "application/json", bytes.NewReader(payload))
    // Parse response and return bearer token
}
```

**Cloud Provider Authentication:**
- **AWS**: Environment variables, instance profiles, credential files
- **Google Cloud**: Service account keys, Application Default Credentials
- **Azure**: Environment variables, managed identity

### Token Usage and Cost Tracking

```go
type TokenUsage struct {
    InputTokens         int64
    OutputTokens        int64
    CacheCreationTokens int64  // For prompt caching
    CacheReadTokens     int64  // For cache hits
}

func calculateCost(model Model, usage TokenUsage) float64 {
    cost := model.CostPer1MIn/1e6*float64(usage.InputTokens) +
            model.CostPer1MOut/1e6*float64(usage.OutputTokens) +
            model.CostPer1MInCached/1e6*float64(usage.CacheCreationTokens) +
            model.CostPer1MOutCached/1e6*float64(usage.CacheReadTokens)
    return cost
}
```

---

## Terminal User Interface (TUI)

### Bubble Tea Architecture

OpenCode implements a sophisticated TUI using the Bubble Tea framework with a component-based architecture:

```go
type appModel struct {
    // Core state
    pages      map[string]tea.Model
    currentPage string
    
    // Dialog system
    dialogs    []tea.Model
    overlay    tea.Model
    
    // Global services
    theme      theme.Theme
    broker     *pubsub.Broker[any]
    
    // Application state
    size       tea.WindowSizeMsg
    commands   map[string]func() tea.Cmd
}
```

### Layout Management System

**Container System:**
```go
type Container struct {
    content tea.Model
    options ContainerOptions
}

type ContainerOptions struct {
    Padding    Padding
    Border     BorderOptions
    Background lipgloss.AdaptiveColor
}
```

**Split Layout:**
```go
type Split struct {
    left         tea.Model
    right        tea.Model
    bottom       tea.Model
    ratios       SplitRatios
    bottomHeight int
}
```

**Overlay System:**
```go
type Overlay struct {
    background tea.Model
    foreground tea.Model
    shadow     bool
    position   Position
}
```

### Component Architecture

**Chat Components:**

```go
// Message List with caching
type MessageList struct {
    messages     []message.Message
    viewport     viewport.Model
    contentCache map[string]map[int]string  // Width-aware caching
    theme        theme.Theme
}

// Editor with external editor support
type Editor struct {
    textInput    textarea.Model
    attachments  []message.Attachment
    theme        theme.Theme
}

// Sidebar with real-time file tracking
type Sidebar struct {
    sessionInfo  SessionInfo
    modifiedFiles []FileChange
    lspStatus    LSPStatus
}
```

**Dialog System:**
```go
// Command palette
type CommandDialog struct {
    list     list.Model
    commands []Command
    filter   string
}

// Multi-argument input
type ArgumentsDialog struct {
    inputs   []textinput.Model
    current  int
    args     map[string]string
}
```

### Theme System

**Theme Interface:**
```go
type Theme interface {
    // Base colors
    Primary() lipgloss.AdaptiveColor
    Secondary() lipgloss.AdaptiveColor
    Accent() lipgloss.AdaptiveColor
    
    // Status colors  
    Error() lipgloss.AdaptiveColor
    Warning() lipgloss.AdaptiveColor
    Success() lipgloss.AdaptiveColor
    
    // Specialized colors for diffs, markdown, syntax highlighting
    DiffAdded() lipgloss.AdaptiveColor
    DiffRemoved() lipgloss.AdaptiveColor
    SyntaxKeyword() lipgloss.AdaptiveColor
    // ... extensive color palette
}
```

**Available Themes:**
- OpenCode (default)
- Catppuccin (Mocha, Macchiato, Frappé, Latte)
- Dracula
- Gruvbox (Dark, Light)
- Monokai Pro
- One Dark
- Tokyo Night
- Flexoki
- Tron

### Event Handling and State Management

**Event Flow:**
1. Input events captured by Bubble Tea
2. Main TUI router distributes to active components
3. Components process relevant events
4. State updates and command generation
5. Re-render cycle triggered

**Global Shortcuts:**
```go
var GlobalBindings = []key.Binding{
    key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
    key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl+l", "logs")),
    key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "sessions")),
    key.NewBinding(key.WithKeys("ctrl+k"), key.WithHelp("ctrl+k", "commands")),
    key.NewBinding(key.WithKeys("ctrl+o"), key.WithHelp("ctrl+o", "models")),
}
```

### Real-time Updates with PubSub

**Event Broadcasting:**
```go
type Broker[T any] struct {
    subs      map[chan Event[T]]struct{}
    mu        sync.RWMutex
    maxEvents int
}

// Event types
type CreatedEvent[T any] struct { Data T }
type UpdatedEvent[T any] struct { Data T }
type DeletedEvent[T any] struct { ID string }
```

**Component Integration:**
```go
func (m *MessageList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case pubsub.Event[message.Message]:
        switch msg.Type() {
        case pubsub.EventTypeCreated:
            m.messages = append(m.messages, msg.Data())
            return m, m.viewport.GotoBottom()
        }
    }
    return m, nil
}
```

### Performance Optimizations

**Content Caching:**
```go
// Width-aware message content caching
type MessageList struct {
    contentCache map[string]map[int]string  // messageID -> width -> content
}

func (m *MessageList) renderMessage(msg message.Message, width int) string {
    if cache, exists := m.contentCache[msg.ID]; exists {
        if content, cached := cache[width]; cached {
            return content
        }
    }
    
    // Render and cache
    content := m.renderMessageContent(msg, width)
    if m.contentCache[msg.ID] == nil {
        m.contentCache[msg.ID] = make(map[int]string)
    }
    m.contentCache[msg.ID][width] = content
    return content
}
```

**Responsive Design:**
```go
func (m *StatusBar) View() string {
    width, _ := m.GetSize()
    
    // Adaptive layout based on available width
    if width < 80 {
        return m.compactView()
    } else if width < 120 {
        return m.standardView()
    } else {
        return m.expandedView()
    }
}
```

---

## Tool System and Agent Integration

### Tool Interface and Architecture

All tools implement a standardized interface for consistent behavior:

```go
type BaseTool interface {
    Info() ToolInfo
    Run(ctx context.Context, params ToolCall) (ToolResponse, error)
}

type ToolInfo struct {
    Name        string         `json:"name"`
    Description string         `json:"description"`
    Parameters  map[string]any `json:"parameters"`  // JSON Schema
    Required    []string       `json:"required"`
}

type ToolResponse struct {
    Type     toolResponseType `json:"type"`     // "text" or "image"
    Content  string           `json:"content"`
    Metadata string           `json:"metadata,omitempty"`
    IsError  bool             `json:"is_error"`
}
```

### Built-in Tools

**File Management Tools:**

```go
// Edit Tool - Precise file modifications
type editTool struct {
    fs          history.Service
    permissions permission.Service
    lspClients  map[string]*lsp.Client
}

// Features:
// - String-based find/replace operations
// - Requires file to be read first (security)
// - LSP integration for immediate diagnostics
// - History tracking for all changes

// View Tool - File content reading
type viewTool struct {
    fs          history.Service
    permissions permission.Service
    lspClients  map[string]*lsp.Client
}

// Features:
// - Line number display with cat -n format
// - Offset/limit support for large files
// - Image detection and handling
// - LSP diagnostics integration
// - 250KB file size limit for safety
```

**Search and Discovery Tools:**

```go
// Grep Tool - Content searching
type grepTool struct {
    permissions permission.Service
}

// Features:
// - Prefers ripgrep for performance
// - Fallback to Go native regex
// - Include pattern filtering
// - Literal text search mode
// - Results sorted by modification time

// Glob Tool - File pattern matching
type globTool struct {
    permissions permission.Service
}

// Features:
// - Fast pattern matching with doublestar
// - Complex glob pattern support
// - Sorted results by modification time
```

**Command Execution:**

```go
// Bash Tool - Shell command execution
type bashTool struct {
    shell       *shell.Shell
    permissions permission.Service
}

// Features:
// - Persistent shell session across commands
// - Command filtering for security
// - Output truncation for large results
// - Timeout handling (1-10 minutes)
// - Safe command auto-approval
```

### Permission System

**Architecture:**
```go
type Service interface {
    Request(ctx context.Context, req PermissionRequest) bool
    Subscribe() <-chan PermissionRequest
    Respond(id string, granted bool)
}

type PermissionRequest struct {
    ID          string `json:"id"`
    SessionID   string `json:"session_id"`
    ToolName    string `json:"tool_name"`
    Description string `json:"description"`
    Action      string `json:"action"`
    Params      any    `json:"params"`
    Path        string `json:"path"`
}
```

**Permission Flow:**
1. Tool requests permission before execution
2. System checks for cached session permissions
3. If not found, publishes request to UI layer
4. Blocks until user grants/denies permission
5. Caches approved permissions for session duration

**Auto-Approval Rules:**
```go
var safeCommands = []string{
    "ls", "cat", "head", "tail", "grep", "find", "wc", "file",
    "pwd", "whoami", "id", "date", "env", "echo", "which",
}

func isSafeCommand(cmd string) bool {
    // Parse command and check against safe list
    // No network access, no system modification
    // Read-only operations generally approved
}
```

### Agent System

**Agent Types and Responsibilities:**

```go
// Coder Agent - Full development capabilities
type CoderAgent struct {
    model    models.Model
    tools    []tools.BaseTool  // All tools available
    sessions map[string]*AgentSession
}

// Available tools: bash, edit, view, write, glob, grep, ls, patch, 
//                  diagnostics, agent, sourcegraph, fetch

// Task Agent - Limited analysis capabilities  
type TaskAgent struct {
    model models.Model
    tools []tools.BaseTool  // Subset of tools
}

// Available tools: glob, grep, ls, view, sourcegraph
```

**Agent Execution Flow:**

```go
func (s *Service) Run(ctx context.Context, sessionID, content string, attachments []message.Attachment) (<-chan AgentEvent, error) {
    // 1. Create event channel for streaming
    events := make(chan AgentEvent)
    
    // 2. Build conversation context
    messages := s.buildConversation(sessionID, content, attachments)
    
    // 3. Start streaming from LLM provider
    providerEvents := s.provider.StreamResponse(ctx, messages, s.tools)
    
    // 4. Process events and handle tool calls
    go s.processProviderEvents(providerEvents, events, sessionID)
    
    return events, nil
}
```

**Tool Call Processing:**
```go
func (s *Service) executeToolCall(ctx context.Context, toolCall message.ToolCall, sessionID string) (message.ToolResult, error) {
    // 1. Find tool by name
    tool := s.findTool(toolCall.Name)
    
    // 2. Add session context for permissions
    ctx = context.WithValue(ctx, SessionIDKey, sessionID)
    ctx = context.WithValue(ctx, MessageIDKey, toolCall.ID)
    
    // 3. Execute tool with parameters
    response, err := tool.Run(ctx, tools.ToolCall{
        ID:   toolCall.ID,
        Name: toolCall.Name,
        Input: toolCall.Input,
    })
    
    // 4. Convert to message format
    return message.ToolResult{
        ID:      toolCall.ID,
        Content: response.Content,
        IsError: response.IsError,
    }, err
}
```

### MCP Tool Integration

**MCP Tool Wrapper:**
```go
type mcpTool struct {
    mcpName     string              // Server identifier
    tool        mcp.Tool           // MCP tool definition
    mcpConfig   config.MCPServer   // Server configuration
    permissions permission.Service  // Permission integration
}

func (m *mcpTool) Run(ctx context.Context, params tools.ToolCall) (tools.ToolResponse, error) {
    // 1. Request permission with MCP context
    permissionReq := permission.PermissionRequest{
        SessionID:   getSessionID(ctx),
        ToolName:    m.tool.Name,
        Description: fmt.Sprintf("MCP Tool: %s", m.tool.Description),
        Action:      "execute",
        Params:      params.Input,
    }
    
    if !m.permissions.Request(ctx, permissionReq) {
        return tools.NewTextErrorResponse("permission denied"), nil
    }
    
    // 2. Execute via MCP client
    result, err := m.client.CallTool(ctx, mcp.CallToolRequest{
        Name:      m.tool.Name,
        Arguments: params.Input,
    })
    
    // 3. Convert MCP response to tool response
    return tools.NewTextResponse(result.Content), nil
}
```

**MCP Server Discovery:**
```go
func GetMcpTools(ctx context.Context, permissions permission.Service) []tools.BaseTool {
    var mcpTools []tools.BaseTool
    
    for serverName, serverConfig := range config.Get().MCPServers {
        // Connect to MCP server
        client, err := connectMCPServer(serverConfig)
        if err != nil {
            continue
        }
        
        // List available tools
        tools, err := client.ListTools(ctx)
        if err != nil {
            continue
        }
        
        // Wrap each tool with permission system
        for _, tool := range tools {
            mcpTool := &mcpTool{
                mcpName:     serverName,
                tool:        tool,
                mcpConfig:   serverConfig,
                permissions: permissions,
            }
            mcpTools = append(mcpTools, mcpTool)
        }
    }
    
    return mcpTools
}
```

### Session and State Management

**File Operation Tracking:**
```go
type FileHistory struct {
    session   string
    readTimes map[string]time.Time   // Track file read times
    writes    map[string][]FileWrite // Track write operations
}

func (h *FileHistory) RecordRead(path string) {
    h.readTimes[path] = time.Now()
}

func (h *FileHistory) RecordWrite(path, content string) {
    h.writes[path] = append(h.writes[path], FileWrite{
        Content:   content,
        Timestamp: time.Now(),
        Version:   h.getNextVersion(path),
    })
}
```

**Shell Session Persistence:**
```go
type Shell struct {
    cmd     *exec.Cmd
    stdin   io.WriteCloser
    stdout  io.ReadCloser
    stderr  io.ReadCloser
    workDir string
    env     []string
}

func (s *Shell) Execute(command string) (string, error) {
    // Write command to persistent shell session
    // Maintain working directory and environment
    // Capture and return output
}
```

### Error Handling and Recovery

**Tool Error Categories:**
1. **Permission Denied**: User rejected operation
2. **Validation Errors**: Invalid parameters or preconditions
3. **Execution Errors**: Command failures or filesystem errors
4. **Timeout Errors**: Operations exceeding time limits
5. **Resource Errors**: File too large, disk full, etc.

**Recovery Strategies:**
```go
func (s *Service) handleToolError(err error, toolName string) AgentEvent {
    switch {
    case isPermissionError(err):
        return AgentEvent{
            Type:  EventError,
            Error: fmt.Errorf("permission denied for %s", toolName),
        }
    case isTimeoutError(err):
        return AgentEvent{
            Type:  EventWarning,
            Error: fmt.Errorf("timeout executing %s, try simpler operation", toolName),
        }
    case isValidationError(err):
        return AgentEvent{
            Type:  EventError,
            Error: fmt.Errorf("invalid parameters for %s: %v", toolName, err),
        }
    default:
        return AgentEvent{
            Type:  EventError,
            Error: fmt.Errorf("unexpected error in %s: %v", toolName, err),
        }
    }
}
```

---

## LSP and MCP Integrations

### Language Server Protocol (LSP) Integration

**Client Architecture:**
```go
type Client struct {
    Cmd            *exec.Cmd
    stdin          io.WriteCloser
    stdout         io.ReadCloser
    conn           *jsonrpc2.Conn
    
    // State management
    serverCapabilities protocol.ServerCapabilities
    openFiles         map[string]*protocol.VersionedTextDocumentIdentifier
    diagnostics       map[string][]protocol.Diagnostic
    
    // Configuration
    serverType        ServerType
    workspaceFolder   string
    initializationOptions any
}
```

**Server Type Detection and Configuration:**
```go
func (c *Client) detectServerType() ServerType {
    cmdPath := strings.ToLower(c.Cmd.Path)
    switch {
    case strings.Contains(cmdPath, "gopls"):
        return ServerTypeGo
    case strings.Contains(cmdPath, "typescript"):
        return ServerTypeTypeScript
    case strings.Contains(cmdPath, "rust-analyzer"):
        return ServerTypeRust
    case strings.Contains(cmdPath, "pylsp") || strings.Contains(cmdPath, "pyright"):
        return ServerTypePython
    case strings.Contains(cmdPath, "clangd"):
        return ServerTypeCPP
    case strings.Contains(cmdPath, "jdtls"):
        return ServerTypeJava
    default:
        return ServerTypeGeneric
    }
}
```

**Language Support Matrix:**

| Language | Server | Detection | Features |
|----------|--------|-----------|----------|
| Go | gopls | `*.go`, `go.mod` | Diagnostics, formatting, completion |
| TypeScript | typescript-language-server | `*.ts`, `*.tsx`, `tsconfig.json` | Full IDE support |
| JavaScript | typescript-language-server | `*.js`, `*.jsx`, `package.json` | Diagnostics, IntelliSense |
| Rust | rust-analyzer | `*.rs`, `Cargo.toml` | Comprehensive Rust support |
| Python | pylsp/pyright | `*.py`, `requirements.txt` | Linting, type checking |
| C/C++ | clangd | `*.c`, `*.cpp`, `*.h` | Compilation database support |
| Java | jdtls | `*.java`, `pom.xml` | Project-aware analysis |

**Workspace Watcher System:**
```go
type Watcher struct {
    client     *Client
    workingDir string
    watcher    *fsnotify.Watcher
    
    // File management
    openFiles     map[string]bool
    pendingFiles  []string
    excludePatterns []string
}

func (w *Watcher) Start() error {
    // 1. Initialize file system watcher
    w.watcher, _ = fsnotify.NewWatcher()
    
    // 2. Add workspace directory
    w.watcher.Add(w.workingDir)
    
    // 3. Preload high-priority files
    w.preloadFiles()
    
    // 4. Start event processing
    go w.processEvents()
    
    return nil
}
```

**Smart File Loading Strategy:**
```go
func (w *Watcher) preloadFiles() {
    serverType := w.client.detectServerType()
    
    switch serverType {
    case ServerTypeTypeScript:
        // TypeScript servers benefit from loading more context
        patterns := []string{
            "**/tsconfig.json", "**/package.json",
            "**/index.ts", "**/main.ts", "**/app.ts",
        }
        w.loadFilesByPatterns(patterns)
        
    case ServerTypeGo:
        // Go servers work well with minimal preloading
        patterns := []string{
            "**/go.mod", "**/go.sum", "**/main.go",
        }
        w.loadFilesByPatterns(patterns)
        
    case ServerTypeRust:
        // Rust analyzer needs Cargo.toml and lib.rs/main.rs
        patterns := []string{
            "**/Cargo.toml", "**/lib.rs", "**/main.rs",
        }
        w.loadFilesByPatterns(patterns)
    }
}
```

**Diagnostic Integration:**
```go
func (c *Client) handleDiagnostics(params protocol.PublishDiagnosticsParams) {
    uri := params.URI
    diagnostics := params.Diagnostics
    
    // Update diagnostic cache
    c.diagnostics[uri] = diagnostics
    
    // Broadcast diagnostic update for tools
    c.diagnosticBroker.Publish(DiagnosticEvent{
        URI:         uri,
        Diagnostics: diagnostics,
    })
}

// Diagnostics tool integration
func (d *diagnosticsTool) Run(ctx context.Context, call ToolCall) (ToolResponse, error) {
    var params DiagnosticsParams
    json.Unmarshal(call.Input, &params)
    
    if params.FilePath != "" {
        // Get diagnostics for specific file
        diags := d.lspClients.GetDiagnostics(params.FilePath)
        return formatDiagnostics(diags), nil
    } else {
        // Get all project diagnostics
        allDiags := d.lspClients.GetAllDiagnostics()
        return formatAllDiagnostics(allDiags), nil
    }
}
```

### Model Control Protocol (MCP) Integration

**MCP Server Types and Configuration:**

```go
type MCPServer struct {
    Command string            `json:"command"`  // For stdio servers
    Env     []string          `json:"env"`
    Args    []string          `json:"args"`
    Type    MCPType           `json:"type"`     // "stdio" or "sse"
    URL     string            `json:"url"`      // For SSE servers
    Headers map[string]string `json:"headers"`
}
```

**STDIO Server Connection:**
```go
func NewStdioMCPClient(command string, env []string, args ...string) (*Client, error) {
    // 1. Start MCP server process
    cmd := exec.Command(command, args...)
    cmd.Env = append(os.Environ(), env...)
    
    // 2. Set up stdio pipes
    stdin, _ := cmd.StdinPipe()
    stdout, _ := cmd.StdoutPipe()
    
    // 3. Start server process
    cmd.Start()
    
    // 4. Initialize MCP client with JSON-RPC over stdio
    client := &Client{
        transport: NewStdioTransport(stdin, stdout),
        process:   cmd,
    }
    
    // 5. Perform MCP handshake
    client.Initialize()
    
    return client, nil
}
```

**SSE Server Connection:**
```go
func NewSSEMCPClient(url string, options ...ClientOption) (*Client, error) {
    // 1. Create HTTP client with custom headers
    httpClient := &http.Client{
        Timeout: 30 * time.Second,
    }
    
    // 2. Connect to SSE endpoint
    req, _ := http.NewRequest("GET", url+"/sse", nil)
    for key, value := range headers {
        req.Header.Set(key, value)
    }
    
    // 3. Set up Server-Sent Events stream
    resp, err := httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    
    // 4. Initialize MCP client with SSE transport
    client := &Client{
        transport: NewSSETransport(resp.Body),
        httpClient: httpClient,
    }
    
    return client, nil
}
```

**Tool Discovery and Registration:**
```go
func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
    // 1. Send list_tools request to MCP server
    request := &ListToolsRequest{}
    response := &ListToolsResponse{}
    
    err := c.Call(ctx, "tools/list", request, response)
    if err != nil {
        return nil, err
    }
    
    // 2. Return discovered tools
    return response.Tools, nil
}

func (c *Client) CallTool(ctx context.Context, req CallToolRequest) (*CallToolResponse, error) {
    // 1. Validate tool request
    if req.Name == "" {
        return nil, fmt.Errorf("tool name is required")
    }
    
    // 2. Send tool call to MCP server
    response := &CallToolResponse{}
    err := c.Call(ctx, "tools/call", req, response)
    
    // 3. Return tool execution results
    return response, err
}
```

**MCP Tool Wrapper for Agent Integration:**
```go
func GetMcpTools(ctx context.Context, permissions permission.Service) []tools.BaseTool {
    var mcpTools []tools.BaseTool
    
    // Iterate through configured MCP servers
    for serverName, serverConfig := range config.Get().MCPServers {
        // Connect to MCP server
        client, err := connectMCPServer(serverConfig)
        if err != nil {
            slog.Warn("failed to connect to MCP server", "server", serverName, "error", err)
            continue
        }
        
        // Discover available tools
        tools, err := client.ListTools(ctx)
        if err != nil {
            slog.Warn("failed to list tools from MCP server", "server", serverName, "error", err)
            continue
        }
        
        // Wrap each tool with permission system
        for _, mcpTool := range tools {
            wrappedTool := &mcpToolWrapper{
                mcpName:     serverName,
                tool:        mcpTool,
                client:      client,
                permissions: permissions,
            }
            mcpTools = append(mcpTools, wrappedTool)
        }
    }
    
    return mcpTools
}
```

**Error Handling and Connection Management:**
```go
func (c *Client) Call(ctx context.Context, method string, params, result interface{}) error {
    // Add timeout to context
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // Retry logic for transient failures
    var lastErr error
    for attempt := 0; attempt < 3; attempt++ {
        err := c.transport.Call(ctx, method, params, result)
        if err == nil {
            return nil
        }
        
        lastErr = err
        
        // Check if error is retryable
        if !isRetryableError(err) {
            break
        }
        
        // Exponential backoff
        backoff := time.Duration(attempt) * time.Second
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(backoff):
            continue
        }
    }
    
    return fmt.Errorf("MCP call failed after retries: %w", lastErr)
}
```

**Performance Optimizations:**

1. **Connection Pooling**: Reuse MCP connections across tool calls
2. **Lazy Loading**: Connect to MCP servers only when tools are used
3. **Caching**: Cache tool discovery results to avoid repeated queries
4. **Timeout Management**: Prevent hanging operations with appropriate timeouts
5. **Error Recovery**: Graceful degradation when MCP servers are unavailable

**Configuration Examples:**

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/allowed/path"],
      "type": "stdio",
      "env": ["DEBUG=mcp:*"]
    },
    "brave-search": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-brave-search"],
      "type": "stdio",
      "env": ["BRAVE_API_KEY=your-api-key"]
    },
    "github": {
      "command": "npx", 
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "type": "stdio",
      "env": ["GITHUB_PERSONAL_ACCESS_TOKEN=your-token"]
    },
    "web-service": {
      "url": "https://api.example.com/mcp",
      "type": "sse",
      "headers": {
        "Authorization": "Bearer your-token",
        "Content-Type": "application/json"
      }
    }
  }
}
```

This comprehensive integration enables OpenCode to leverage both language server capabilities for code intelligence and MCP servers for extended functionality, providing a powerful and extensible development environment.

---

## Development Guidelines

### Setting Up Development Environment

**Prerequisites:**
- Go 1.24.0 or higher
- Node.js (for MCP server development)
- Git
- Your preferred language servers (gopls, typescript-language-server, etc.)

**Building from Source:**
```bash
# Clone the repository
git clone https://github.com/opencode-ai/opencode.git
cd opencode

# Build the application
go build -o opencode

# Run with debug logging
./opencode -d
```

**Development Configuration:**
```bash
# Enable development debugging
export OPENCODE_DEV_DEBUG=true

# Use local development config
cp .opencode.example.json .opencode.json
# Edit configuration for your development needs
```

### Code Organization

**Module Structure:**
```
internal/
├── app/           # Application coordination
├── config/        # Configuration management
├── db/            # Database operations
├── llm/           # LLM integration
│   ├── agent/     # Agent system
│   ├── models/    # Model definitions
│   ├── provider/  # Provider implementations
│   └── tools/     # Tool implementations
├── lsp/           # LSP integration
├── tui/           # Terminal UI
│   ├── components/
│   ├── layout/
│   ├── page/
│   └── theme/
├── logging/       # Logging system
├── message/       # Message handling
├── permission/    # Permission system
├── pubsub/        # Event system
└── session/       # Session management
```

### Testing Guidelines

**Unit Testing:**
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/config
```

**Integration Testing:**
```bash
# Test with real LLM providers (requires API keys)
go test -tags=integration ./internal/llm/provider

# Test LSP integration (requires language servers)
go test -tags=integration ./internal/lsp
```

### Database Development

**Working with Database Schema:**

```bash
# Generate queries from SQL
sqlc generate

# Create new migration
goose -dir internal/db/migrations create migration_name sql

# Apply migrations in development
go run main.go # Migrations run automatically on startup
```

**Adding New Database Operations:**

1. Add SQL query to `internal/db/sql/*.sql`
2. Run `sqlc generate` to generate Go code
3. Use generated methods in service layer

### Adding New LLM Providers

**Step-by-Step Guide:**

1. **Create Provider Client** (`internal/llm/provider/newprovider.go`):
```go
type newProviderClient struct {
    providerOptions providerClientOptions
    options         newProviderOptions
    client          NewProviderSDKClient
}

func newNewProviderClient(opts providerClientOptions) NewProviderClient {
    return &newProviderClient{
        providerOptions: opts,
        client:          initializeSDK(opts),
    }
}

func (n *newProviderClient) send(ctx context.Context, messages []message.Message, tools []tools.BaseTool) (*ProviderResponse, error) {
    // Convert messages and make API call
}

func (n *newProviderClient) stream(ctx context.Context, messages []message.Message, tools []tools.BaseTool) <-chan ProviderEvent {
    // Implement streaming response
}
```

2. **Add Model Definitions** (`internal/llm/models/newprovider.go`):
```go
const ProviderNewProvider ModelProvider = "newprovider"

const (
    NewProviderModel1 ModelID = "newprovider.model1"
    NewProviderModel2 ModelID = "newprovider.model2"
)

var NewProviderModels = map[ModelID]Model{
    NewProviderModel1: {
        ID:                  NewProviderModel1,
        Name:                "New Provider Model 1",
        Provider:            ProviderNewProvider,
        APIModel:            "model-1-api-name",
        CostPer1MIn:         1.0,
        CostPer1MOut:        3.0,
        ContextWindow:       128_000,
        DefaultMaxTokens:    4096,
        SupportsAttachments: true,
    },
}
```

3. **Register Provider** (`internal/llm/provider/provider.go`):
```go
func NewProvider(providerName models.ModelProvider, opts ...ProviderClientOption) (Provider, error) {
    switch providerName {
    // ... existing cases
    case models.ProviderNewProvider:
        return &baseProvider[NewProviderClient]{
            options: clientOptions,
            client:  newNewProviderClient(clientOptions),
        }, nil
    }
}
```

4. **Update Configuration** (`internal/config/config.go`):
```go
func setProviderDefaults() {
    if apiKey := os.Getenv("NEWPROVIDER_API_KEY"); apiKey != "" {
        viper.SetDefault("providers.newprovider.apiKey", apiKey)
    }
    
    // Add to provider priority order
    if key := viper.GetString("providers.newprovider.apiKey"); strings.TrimSpace(key) != "" {
        viper.SetDefault("agents.coder.model", models.NewProviderModel1)
        return
    }
}
```

### Adding New Tools

**Tool Implementation Pattern:**

```go
type myTool struct {
    // Dependencies (permissions, services, etc.)
    permissions permission.Service
    fs          history.Service
}

func NewMyTool(permissions permission.Service, fs history.Service) tools.BaseTool {
    return &myTool{
        permissions: permissions,
        fs:          fs,
    }
}

func (t *myTool) Info() tools.ToolInfo {
    return tools.ToolInfo{
        Name:        "my_tool",
        Description: "Description of what this tool does",
        Parameters: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "required_param": map[string]any{
                    "type":        "string",
                    "description": "Description of parameter",
                },
            },
            "required": []string{"required_param"},
        },
    }
}

func (t *myTool) Run(ctx context.Context, call tools.ToolCall) (tools.ToolResponse, error) {
    // 1. Parse and validate parameters
    var params MyToolParams
    if err := json.Unmarshal(call.Input, &params); err != nil {
        return tools.NewTextErrorResponse("invalid parameters"), nil
    }
    
    // 2. Get session context
    sessionID := getSessionID(ctx)
    
    // 3. Request permissions if needed
    if !t.permissions.Request(ctx, permission.PermissionRequest{
        SessionID:   sessionID,
        ToolName:    "my_tool",
        Description: "Description for user",
        Action:      "action_name",
        Params:      params,
    }) {
        return tools.NewTextErrorResponse("permission denied"), nil
    }
    
    // 4. Execute tool logic
    result, err := t.executeLogic(ctx, params)
    if err != nil {
        return tools.NewTextErrorResponse(err.Error()), nil
    }
    
    // 5. Return response
    return tools.NewTextResponse(result), nil
}
```

### TUI Component Development

**Component Pattern:**

```go
type myComponent struct {
    // State
    data   []MyData
    list   list.Model
    
    // Dependencies
    theme  theme.Theme
    broker *pubsub.Broker[MyEvent]
    
    // Layout
    size tea.WindowSizeMsg
}

func NewMyComponent(theme theme.Theme, broker *pubsub.Broker[MyEvent]) *myComponent {
    return &myComponent{
        theme:  theme,
        broker: broker,
        list:   list.New([]list.Item{}, myDelegate{}, 0, 0),
    }
}

func (c *myComponent) Init() tea.Cmd {
    return c.broker.Subscribe()
}

func (c *myComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        c.size = msg
        c.list.SetSize(msg.Width, msg.Height)
        
    case tea.KeyMsg:
        switch msg.String() {
        case "enter":
            return c, c.handleEnter()
        }
        
    case pubsub.Event[MyEvent]:
        return c, c.handleEvent(msg)
    }
    
    var cmd tea.Cmd
    c.list, cmd = c.list.Update(msg)
    return c, cmd
}

func (c *myComponent) View() string {
    return c.list.View()
}

// Implement layout interfaces if needed
func (c *myComponent) SetSize(width, height int) tea.Cmd {
    c.size = tea.WindowSizeMsg{Width: width, Height: height}
    return c.list.SetSize(width, height)
}
```

### Error Handling Best Practices

**Error Types:**
```go
// Define custom error types for different scenarios
type ValidationError struct {
    Field string
    Value any
    Msg   string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation error for %s: %s", e.Field, e.Msg)
}

type PermissionError struct {
    Tool      string
    Action    string
    SessionID string
}

func (e PermissionError) Error() string {
    return fmt.Sprintf("permission denied for %s action in session %s", e.Action, e.SessionID)
}
```

**Error Handling Pattern:**
```go
func (s *Service) ProcessRequest(ctx context.Context, req Request) error {
    // Validate input
    if err := req.Validate(); err != nil {
        return ValidationError{
            Field: "request",
            Value: req,
            Msg:   err.Error(),
        }
    }
    
    // Check permissions
    if !s.permissions.Check(ctx, req.Action) {
        return PermissionError{
            Tool:      req.Tool,
            Action:    req.Action,
            SessionID: req.SessionID,
        }
    }
    
    // Execute with proper error context
    if err := s.execute(ctx, req); err != nil {
        return fmt.Errorf("failed to execute %s: %w", req.Action, err)
    }
    
    return nil
}
```

---

## Extension Patterns

### Creating Custom MCP Servers

**Basic MCP Server Structure:**

```typescript
import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';

const server = new Server(
  {
    name: 'my-custom-server',
    version: '0.1.0',
  },
  {
    capabilities: {
      tools: {},
    },
  }
);

// Define tools
server.setRequestHandler('tools/list', async () => {
  return {
    tools: [
      {
        name: 'my_tool',
        description: 'Description of my custom tool',
        inputSchema: {
          type: 'object',
          properties: {
            input: { type: 'string', description: 'Input parameter' }
          },
          required: ['input']
        }
      }
    ]
  };
});

server.setRequestHandler('tools/call', async (request) => {
  const { name, arguments: args } = request.params;
  
  switch (name) {
    case 'my_tool':
      const result = await executeMyTool(args.input);
      return {
        content: [
          {
            type: 'text',
            text: result
          }
        ]
      };
    
    default:
      throw new Error(`Unknown tool: ${name}`);
  }
});

// Start server
const transport = new StdioServerTransport();
await server.connect(transport);
```

**Integration with OpenCode:**

```json
{
  "mcpServers": {
    "my-custom-server": {
      "command": "node",
      "args": ["path/to/my-server.js"],
      "type": "stdio",
      "env": ["API_KEY=your-key"]
    }
  }
}
```

### Adding Custom Themes

**Theme Implementation:**

```go
package theme

import "github.com/charmbracelet/lipgloss"

type MyTheme struct{}

func NewMyTheme() Theme {
    return &MyTheme{}
}

func (t *MyTheme) Primary() lipgloss.AdaptiveColor {
    return lipgloss.AdaptiveColor{
        Light: "#0066CC",
        Dark:  "#4A9EFF",
    }
}

func (t *MyTheme) Secondary() lipgloss.AdaptiveColor {
    return lipgloss.AdaptiveColor{
        Light: "#6B7280",
        Dark:  "#9CA3AF",
    }
}

// Implement all required Theme interface methods...

func init() {
    RegisterTheme("mytheme", NewMyTheme)
}
```

**Theme Registration:**

```go
// In theme/manager.go
func RegisterTheme(name string, constructor func() Theme) {
    themes[name] = constructor
}

func GetTheme(name string) (Theme, error) {
    if constructor, exists := themes[name]; exists {
        return constructor(), nil
    }
    return nil, fmt.Errorf("theme %s not found", name)
}
```

### Custom LSP Integration

**Adding New Language Server:**

```go
// In internal/lsp/language.go
const (
    ServerTypeMyLanguage ServerType = "mylanguage"
)

func detectLanguageFromFile(filename string) string {
    ext := strings.ToLower(filepath.Ext(filename))
    switch ext {
    case ".myl":
        return "mylanguage"
    // ... other cases
    }
}

// In internal/lsp/client.go
func (c *Client) getInitializationOptions() any {
    switch c.serverType {
    case ServerTypeMyLanguage:
        return map[string]any{
            "myLanguageSpecificOption": true,
        }
    // ... other cases
    }
}
```

**Configuration:**

```json
{
  "lsp": {
    "mylanguage": {
      "command": "mylang-lsp",
      "args": ["--stdio"],
      "disabled": false
    }
  }
}
```

### Extending Agent Capabilities

**Custom Agent Type:**

```go
type CustomAgent struct {
    model   models.Model
    tools   []tools.BaseTool
    config  CustomAgentConfig
}

type CustomAgentConfig struct {
    MaxIterations int
    SpecialMode   bool
}

func NewCustomAgent(model models.Model, tools []tools.BaseTool, config CustomAgentConfig) agent.Service {
    return &CustomAgent{
        model:  model,
        tools:  tools,
        config: config,
    }
}

func (a *CustomAgent) Run(ctx context.Context, sessionID, content string, attachments []message.Attachment) (<-chan agent.AgentEvent, error) {
    // Custom agent logic
    events := make(chan agent.AgentEvent)
    
    go func() {
        defer close(events)
        
        // Custom processing logic
        for i := 0; i < a.config.MaxIterations; i++ {
            // Process with special logic
            result := a.processIteration(ctx, content, i)
            
            events <- agent.AgentEvent{
                Type:    agent.EventProgress,
                Content: result,
            }
        }
        
        events <- agent.AgentEvent{
            Type: agent.EventComplete,
        }
    }()
    
    return events, nil
}
```

**Agent Registration:**

```go
// In internal/llm/agent/registry.go
type AgentRegistry struct {
    agents map[config.AgentName]func(models.Model, []tools.BaseTool) agent.Service
}

func (r *AgentRegistry) Register(name config.AgentName, constructor func(models.Model, []tools.BaseTool) agent.Service) {
    r.agents[name] = constructor
}

func init() {
    registry := GetRegistry()
    registry.Register("custom", func(model models.Model, tools []tools.BaseTool) agent.Service {
        return NewCustomAgent(model, tools, CustomAgentConfig{
            MaxIterations: 5,
            SpecialMode:   true,
        })
    })
}
```

This comprehensive developer documentation provides everything needed to understand, extend, and contribute to the OpenCode project. The modular architecture and clear extension patterns make it straightforward to add new capabilities while maintaining code quality and consistency.

---

## Contributing

### Code Style and Standards

- Follow Go conventions and use `gofmt`
- Write comprehensive tests for new features
- Update documentation for API changes
- Use structured logging with appropriate context
- Implement proper error handling with context

### Pull Request Process

1. Fork the repository and create a feature branch
2. Implement changes with tests and documentation
3. Ensure all tests pass and code is properly formatted
4. Submit pull request with clear description of changes
5. Address review feedback and maintain backwards compatibility

### Security Considerations

- Always validate user inputs and file paths
- Implement proper permission checks for sensitive operations
- Use secure defaults and fail safely
- Avoid logging sensitive information (API keys, tokens)
- Test security boundaries and edge cases

This documentation serves as a comprehensive guide for understanding and extending OpenCode's architecture, providing developers with the knowledge needed to contribute effectively to the project.