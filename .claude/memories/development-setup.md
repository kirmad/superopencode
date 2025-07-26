# SuperOpenCode Development Setup

## Prerequisites
- **Go**: Version 1.24.0 or higher
- **Node.js**: Version 18+ (for log viewer)
- **Git**: For version control
- **SQLite**: Database support (included in dependencies)

## Environment Setup

### Go Environment
- **GOPATH**: `/Users/kirmadi/go` (typical setup)
- **gopls**: Go language server installed at `/Users/kirmadi/go/bin/gopls`
- **PATH**: Should include `$GOPATH/bin` for tool access

### Required Tools
- **gopls**: `go install golang.org/x/tools/gopls@latest`
- **goose**: Database migration tool (included in dependencies)
- **sqlc**: Type-safe SQL code generation (configured via sqlc.yaml)

## Building from Source

### Main Application
```bash
# Clone repository
git clone https://github.com/kirmad/superopencode.git
cd superopencode

# Install dependencies
go mod download

# Build application
go build -o opencode

# Run application
./opencode
```

### Log Viewer Setup
```bash
# Navigate to log viewer
cd log-viewer

# Install dependencies
npm install

# Set up environment
cp .env.example .env.local
# Edit .env.local with appropriate paths

# Start development server
npm run dev
```

## Configuration

### Main Application Config
Location options (in priority order):
1. `$HOME/.opencode.json`
2. `$XDG_CONFIG_HOME/opencode/.opencode.json`
3. `./.opencode.json` (project local)

### Key Configuration Sections
- **data**: Database and storage configuration
- **providers**: AI provider API keys and settings
- **agents**: Model configurations for different agent types
- **shell**: Shell command configuration
- **mcpServers**: Model Context Protocol server definitions
- **lsp**: Language server configurations
- **debug/debugLSP**: Debugging options

### Environment Variables
Essential for API access:
- `ANTHROPIC_API_KEY`: For Claude models
- `OPENAI_API_KEY`: For OpenAI models
- `GEMINI_API_KEY`: For Google Gemini models
- `GITHUB_TOKEN`: For GitHub Copilot models
- Additional provider-specific keys as needed

## Database Setup

### SQLite Database
- **Location**: Configurable via `data.directory` in config
- **Migrations**: Automatically run on startup
- **Schema**: Defined in `internal/db/migrations/`
- **Generated Code**: SQL queries generated via sqlc

### Migration Management
- Uses **goose** for database migrations
- Migrations located in `internal/db/migrations/`
- Automatic migration on application startup

## Development Workflow

### Command Line Flags
- `--debug` / `-d`: Enable debug logging
- `--cwd` / `-c`: Set working directory
- `--prompt` / `-p`: Non-interactive mode
- `--output-format` / `-f`: Output format (text/json)
- `--quiet` / `-q`: Suppress spinner in non-interactive mode
- `--detailed-logs`: Enable detailed LLM interaction logging

### Testing and Quality
- **Testing Framework**: Uses testify for Go tests
- **Code Generation**: SQLC for type-safe SQL
- **Linting**: Standard Go tooling (go fmt, go vet)

### Log Viewer Development
- **Hot Reload**: Next.js development server with hot reload
- **TypeScript**: Full type safety with strict mode
- **Tailwind CSS**: Utility-first styling with customizable themes
- **API Routes**: Next.js API routes for data access

## Project Structure
```
superopencode/
├── cmd/                    # CLI commands and entry points
├── internal/               # Internal application code
│   ├── app/               # Core application services
│   ├── config/            # Configuration management
│   ├── db/                # Database layer
│   ├── llm/               # LLM integration
│   ├── tui/               # Terminal UI
│   ├── logging/           # Logging infrastructure
│   └── lsp/               # Language server protocol
├── log-viewer/            # Next.js log visualization app
├── docs/                  # Documentation
├── scripts/               # Build and utility scripts
├── main.go                # Application entry point
├── go.mod                 # Go module dependencies
└── sqlc.yaml              # SQLC configuration
```

## Known Issues

### LSP Integration
- **gopls PATH Issue**: May require manual PATH configuration for some environments
- **Language Server Setup**: Requires appropriate language servers installed for each supported language

### Development Environment
- **Early Development**: Project is in early development stage
- **API Changes**: Internal APIs may change frequently
- **Production Readiness**: Not yet ready for production use

## Build Scripts
- `scripts/release`: Production release build
- `scripts/snapshot`: Development snapshot build
- `scripts/check_hidden_chars.sh`: Code quality checks