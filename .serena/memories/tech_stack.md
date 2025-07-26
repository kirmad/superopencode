# SuperOpenCode Tech Stack and Dependencies

## Core Technology Stack

### Language and Runtime
- **Go 1.24.0+** - Primary programming language
- **SQLite** - Embedded database for persistence
- **JSON-RPC 2.0** - Protocol for LSP and MCP communication

### Key Frameworks and Libraries

#### User Interface
- **Bubble Tea** (`github.com/charmbracelet/bubbletea`) - TUI framework
- **Lipgloss** (`github.com/charmbracelet/lipgloss`) - Terminal styling
- **Glamour** (`github.com/charmbracelet/glamour`) - Markdown rendering
- **Bubbles** (`github.com/charmbracelet/bubbles`) - TUI components

#### AI/LLM Integration
- **Anthropic SDK** (`github.com/anthropics/anthropic-sdk-go`) - Claude integration
- **OpenAI SDK** (`github.com/openai/openai-go`) - GPT models
- **Google GenAI** (`google.golang.org/genai`) - Gemini models
- **Azure SDK** (`github.com/Azure/azure-sdk-for-go`) - Azure OpenAI
- **AWS SDK v2** - Bedrock integration

#### Database and Storage
- **SQLite** (`github.com/ncruces/go-sqlite3`) - Primary database
- **SQLC** - Type-safe SQL code generation
- **Goose** (`github.com/pressly/goose/v3`) - Database migrations

#### CLI and Configuration
- **Cobra** (`github.com/spf13/cobra`) - CLI framework
- **Viper** (`github.com/spf13/viper`) - Configuration management

#### Development Tools
- **GoReleaser** - Build and release automation
- **GitHub Actions** - CI/CD pipeline

#### File Operations and Search
- **Doublestar** (`github.com/bmatcuk/doublestar/v4`) - Glob pattern matching
- **Go-diff** (`github.com/sergi/go-diff`) - Diff generation
- **FSNotify** (`github.com/fsnotify/fsnotify`) - File system watching

#### Protocol Support
- **MCP Go SDK** (`github.com/mark3labs/mcp-go`) - Model Context Protocol
- **Language Server Protocol** - LSP client implementation

## Development Dependencies
- **Testing:** `github.com/stretchr/testify` - Test assertions and mocking
- **Styling:** Catppuccin color scheme support
- **Build Tools:** Custom scripts using Bash and GoReleaser

## External Integrations
- **Language Servers:** gopls (Go), typescript-language-server, rust-analyzer, etc.
- **MCP Servers:** Filesystem, Brave Search, GitHub integration servers
- **AI Providers:** Multiple providers with unified interface
- **Shell Integration:** Persistent shell sessions with command execution