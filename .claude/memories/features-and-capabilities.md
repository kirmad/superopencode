# SuperOpenCode Features and Capabilities

## Core Features

### AI Provider Support
- **OpenAI**: GPT-4.1 family, GPT-4.5 Preview, GPT-4o family, O1 family, O3 family, O4 Mini
- **Anthropic**: Claude 4 Sonnet/Opus, Claude 3.5 Sonnet, Claude 3.7 Sonnet, Claude 3 Haiku/Opus
- **GitHub Copilot**: Multiple models including Claude, GPT, Gemini variants
- **Google**: Gemini 2.5, Gemini 2.0 Flash variants
- **AWS Bedrock**: Claude 3.7 Sonnet
- **Groq**: Llama 4 variants, QWEN, Deepseek R1
- **Azure OpenAI**: GPT and O1 families
- **Google Cloud VertexAI**: Gemini variants

### Interactive Features

#### Terminal User Interface
- **Framework**: Bubble Tea for smooth terminal experience
- **Vim-like Editor**: Integrated text input with familiar keybindings
- **Session Management**: Save, load, and switch between conversation sessions
- **Real-time Updates**: Live display of AI responses and tool executions

#### Keyboard Shortcuts
- **Global**: `Ctrl+C` quit, `Ctrl+?` help, `Ctrl+L` logs, `Ctrl+A` switch session
- **Chat**: `Ctrl+N` new session, `Ctrl+X` cancel, `i` focus editor
- **Editor**: `Ctrl+S` send message, `Ctrl+E` external editor, `Esc` blur

### AI Assistant Tools

#### File and Code Operations
- **glob**: Find files by pattern
- **grep**: Search file contents with regex support
- **ls**: List directory contents with filtering
- **view**: Read file contents with pagination
- **write**: Create and modify files
- **edit**: Advanced file editing capabilities
- **patch**: Apply diff patches to files
- **diagnostics**: Get LSP diagnostics information

#### System Integration
- **bash**: Execute shell commands with timeout support
- **fetch**: HTTP requests to external APIs
- **sourcegraph**: Search public code repositories
- **agent**: Delegate sub-tasks to AI agent

### Advanced Features

#### Language Server Protocol (LSP)
- **Multi-language Support**: Go (gopls), TypeScript, and others
- **Code Intelligence**: Error checking, diagnostics, file watching
- **AI Integration**: AI can access diagnostics for code analysis

#### Model Context Protocol (MCP)
- **External Tools**: Standardized protocol for tool integration
- **Connection Types**: stdio and Server-Sent Events (SSE)
- **Security**: Permission system for tool access control

#### Custom Commands
- **User Commands**: Store in `~/.config/opencode/commands/`
- **Project Commands**: Store in `<PROJECT>/.opencode/commands/`
- **Named Arguments**: Placeholders like `$ISSUE_NUMBER`, `$AUTHOR_NAME`
- **Slash Commands**: Quick execution with `/command` syntax

### Session and Data Management

#### Auto Compact Feature
- **Context Management**: Automatically summarizes conversations near token limits
- **Smart Continuation**: Creates new sessions with summaries
- **Token Monitoring**: Tracks usage and prevents out-of-context errors

#### Persistent Storage
- **SQLite Database**: Stores conversations, sessions, and metadata
- **Detailed Logging**: Optional detailed HTTP request/response logging
- **Cross-session State**: Maintains context across application restarts

### Operating Modes

#### Interactive Mode (Default)
- Full TUI experience with chat interface
- Real-time AI interaction with tool execution
- Session management and history

#### Non-interactive Mode
- Single prompt execution: `opencode -p "your prompt"`
- Output formats: text (default) or JSON
- Quiet mode: `--quiet` flag hides spinner
- Auto-approved permissions for automation

### Configuration System

#### Environment Variables
- API keys for all supported providers
- Custom endpoints for self-hosted models
- Shell configuration overrides

#### JSON Configuration
- Flexible provider settings
- Agent-specific model configurations
- MCP server definitions
- LSP server configurations
- Shell and debugging options