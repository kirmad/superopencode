# SuperOpenCode Code Conventions and Style Guide

## Go Code Conventions

### Naming Conventions
- **Packages:** Lowercase, single word when possible (`session`, `message`, `config`)
- **Types:** PascalCase (`App`, `ProviderResponse`, `AgentEvent`)
- **Functions/Methods:** PascalCase for exported, camelCase for private
- **Constants:** PascalCase with logical grouping
- **Variables:** camelCase, descriptive names

### Code Organization
- **File Structure:** One primary type per file, named after the type
- **Package Structure:** Clear single responsibility per package
- **Interface Placement:** Interfaces defined in consumer packages when possible

### Error Handling
```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to execute %s: %w", action, err)
}

// Use custom error types for different scenarios
type ValidationError struct {
    Field string
    Value any
    Msg   string
}
```

### Logging Patterns
```go
// Structured logging with context
slog.Info("session created", "sessionID", sessionID, "userID", userID)
slog.Error("failed to connect", "provider", provider, "error", err)

// Use logging.RecoverPanic for goroutines
defer logging.RecoverPanic("worker", func() {
    logging.ErrorPersist("Worker terminated due to panic")
})
```

### Concurrency Patterns
```go
// Proper mutex usage
type SafeMap struct {
    mu   sync.RWMutex
    data map[string]interface{}
}

// Context-aware goroutines
go func() {
    defer wg.Done()
    select {
    case <-ctx.Done():
        return
    case result := <-work():
        // Process result
    }
}()
```

## Database Conventions

### SQL Style (SQLC)
- **Naming:** snake_case for table and column names
- **Queries:** Descriptive names with CRUD prefix
```sql
-- name: CreateSession :one
-- name: GetSessionByID :one
-- name: ListRecentSessions :many
-- name: UpdateSessionTitle :exec
```

### Migration Naming
- **Pattern:** `YYYYMMDDHHMMSS_descriptive_name.sql`
- **Content:** Both up and down migrations when possible

## Testing Conventions

### Test File Organization
- **Unit Tests:** `*_test.go` in same package
- **Integration Tests:** `*_integration_test.go` with build tags
- **Benchmarks:** `*_benchmark_test.go` with benchmark functions

### Test Naming
```go
func TestServiceCreateSession(t *testing.T) { }
func TestServiceCreateSession_WithInvalidInput_ReturnsError(t *testing.T) { }
func BenchmarkAgentProcessing(b *testing.B) { }
```

## Documentation Conventions

### Code Comments
- **Package Comments:** Comprehensive overview of package purpose
- **Type Comments:** What the type represents and its main responsibility
- **Function Comments:** What it does, important parameters, return values
- **No Obvious Comments:** Avoid comments that just restate the code

### API Documentation
- Use godoc-compatible comments for all exported items
- Include examples for complex functions
- Document thread safety characteristics

## Import Organization
```go
package main

import (
    // Standard library
    "context"
    "fmt"
    "log"
    
    // Third-party libraries
    "github.com/spf13/cobra"
    "github.com/charmbracelet/bubbletea"
    
    // Internal packages
    "github.com/kirmad/superopencode/internal/app"
    "github.com/kirmad/superopencode/internal/config"
)
```

## Configuration Patterns
- Use Viper for configuration management
- Support multiple configuration sources (env, file, flags)
- Provide sensible defaults
- Validate configuration at startup

## Security Patterns
- Always validate inputs, especially file paths
- Use permission system for sensitive operations
- Never log sensitive information (API keys, tokens)
- Implement secure defaults and fail safely