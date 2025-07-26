# SuperOpenCode Suggested Commands

## Development Commands

### Building and Running
```bash
# Build the application
go build -o opencode

# Build with version information (production)
go build -ldflags "-s -w -X github.com/kirmad/superopencode/internal/version.Version=v1.0.0" -o opencode

# Run application directly
go run main.go

# Run with debug logging
go run main.go -d

# Run with specific working directory
go run main.go -c /path/to/project

# Non-interactive mode
go run main.go -p "your prompt here"
```

### Testing Commands
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/config
go test ./internal/llm/agent

# Run integration tests (requires API keys)
go test -tags=integration ./internal/llm/provider

# Run benchmarks
go test -bench=. ./internal/llm/agent

# Run tests with race detection
go test -race ./...
```

### Code Quality and Formatting
```bash
# Format all Go code
go fmt ./...

# Run Go vet (static analysis)
go vet ./...

# Run golint (if installed)
golint ./...

# Check for security issues (if gosec installed)
gosec ./...

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify
```

### Database Operations
```bash
# Generate SQL queries with SQLC
sqlc generate

# Check migration status (if using standalone goose)
goose -dir internal/db/migrations status

# Note: Migrations run automatically on application startup
```

### Build and Release
```bash
# Snapshot build (development)
./scripts/snapshot

# Full release build
./scripts/release

# Build with GoReleaser
goreleaser build --clean --snapshot
```

### Development Utilities
```bash
# Check for hidden characters in files
./scripts/check_hidden_chars.sh

# Install language servers (if needed)
go install golang.org/x/tools/gopls@latest
npm install -g typescript-language-server

# Run application with debug environment
OPENCODE_DEV_DEBUG=true ./opencode -d
```

## macOS-Specific Commands

### System Commands (Darwin)
```bash
# List files (BSD version)
ls -la

# Find files (BSD find)
find . -name "*.go" -type f

# Search in files (if ripgrep installed)
rg "pattern" --type go

# Open files with default application
open README.md

# Copy to clipboard
pbcopy < file.txt

# Paste from clipboard
pbpaste > file.txt
```

### Package Management
```bash
# Install via Homebrew
brew install opencode-ai/tap/opencode

# Update Homebrew packages
brew update && brew upgrade

# Install development dependencies
brew install go node git ripgrep
```

## Git Workflow Commands
```bash
# Standard git operations
git status
git add .
git commit -m "feat: add new feature"
git push origin feature-branch

# Create feature branch
git checkout -b feature/new-feature

# Update from main
git fetch origin
git rebase origin/main
```

## Configuration Commands
```bash
# View current configuration
cat ~/.opencode.json

# Edit configuration
$EDITOR ~/.opencode.json

# Test configuration
./opencode -d  # Check debug output for config issues
```

## Debugging Commands
```bash
# Enable debug logging
export OPENCODE_DEV_DEBUG=true

# Run with debug flags
./opencode -d

# Check logs directory
ls -la logs/

# Follow log file
tail -f logs/debug.log
```

## LSP Integration Commands
```bash
# Install common language servers
go install golang.org/x/tools/gopls@latest
npm install -g typescript-language-server
pip install python-lsp-server

# Check if language servers are available
which gopls
which typescript-language-server
```

## MCP Development Commands
```bash
# Install MCP servers
npx -y @modelcontextprotocol/server-filesystem
npx -y @modelcontextprotocol/server-brave-search

# Test MCP server
node path/to/mcp-server.js
```

## Performance and Monitoring
```bash
# Profile CPU usage
go test -cpuprofile=cpu.prof -bench=.

# Profile memory usage
go test -memprofile=mem.prof -bench=.

# Check binary size
ls -lh opencode

# Check dependencies
go mod graph
```

These commands should cover the essential development workflow for the SuperOpenCode project.