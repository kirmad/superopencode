# SuperOpenCode Task Completion Guidelines

## What to Do When a Task is Completed

### 1. Code Quality Checks
Before marking any development task as complete, ensure the following:

```bash
# Format all Go code
go fmt ./...

# Run static analysis
go vet ./...

# Check for potential issues
golint ./... 2>/dev/null || echo "golint not installed"
```

### 2. Testing Requirements
All code changes must pass testing:

```bash
# Run all unit tests
go test ./...

# Run tests with race detection
go test -race ./...

# Run integration tests if applicable
go test -tags=integration ./internal/llm/provider

# Verify test coverage for new code
go test -cover ./internal/yourpackage
```

### 3. Build Verification
Ensure the application builds successfully:

```bash
# Standard build
go build -o opencode

# Test that the binary works
./opencode --help

# Verify no import cycles
go mod verify
```

### 4. Database Considerations
If database changes were made:

```bash
# Regenerate SQL code if .sql files changed
sqlc generate

# Verify migrations work (they run automatically on startup)
./opencode -d  # Check for migration errors in debug output
```

### 5. Documentation Updates
- Update relevant comments for exported functions/types
- Update DEVELOPER_DOCUMENTATION.md if architecture changes
- Update README.md if user-facing features change
- Add or update memory files if significant patterns change

### 6. Security Review
For security-sensitive changes:
- Ensure no sensitive data is logged
- Verify proper input validation
- Check permission system integration
- Review file path handling for traversal attacks

### 7. Performance Considerations
For performance-critical changes:
- Run relevant benchmarks
- Check for memory leaks in long-running operations
- Verify goroutine cleanup
- Test with realistic data sizes

### 8. Integration Testing
For features involving external systems:
- Test LSP integration with relevant language servers
- Test MCP integration with configured servers
- Test AI provider integration (if API keys available)
- Verify configuration loading from different sources

### 9. Error Handling Verification
Ensure robust error handling:
- All errors include sufficient context
- No panics in normal operation
- Graceful degradation when external services fail
- Proper cleanup in error paths

### 10. Commit Guidelines
When ready to commit:

```bash
# Stage changes
git add .

# Commit with descriptive message
git commit -m "feat: add new feature

- Detailed description of changes
- Any breaking changes noted
- References to issues if applicable"

# Ensure clean commit history
git log --oneline -5
```

## Checklist Template
For each completed task, verify:

- [ ] Code formatted with `go fmt`
- [ ] Static analysis passes with `go vet`
- [ ] All tests pass
- [ ] Application builds successfully
- [ ] No new linter warnings
- [ ] Documentation updated if needed
- [ ] Security considerations addressed
- [ ] Performance impact considered
- [ ] Error handling implemented
- [ ] Integration tests pass (if applicable)

## Special Considerations

### For LSP-Related Changes
- Test with multiple language servers
- Verify file watching works correctly
- Check diagnostic integration
- Test startup/shutdown behavior

### For LLM Provider Changes
- Test with rate limiting
- Verify streaming response handling
- Check token usage calculation
- Test error recovery

### For TUI Changes
- Test responsive design at different terminal sizes
- Verify keyboard shortcuts work
- Check theme compatibility
- Test with different terminal emulators

### For Database Changes
- Verify migration safety
- Check transaction boundaries
- Test with existing data
- Verify foreign key constraints

### For Tool System Changes
- Test permission system integration
- Verify parameter validation
- Check timeout handling
- Test error response formatting

## Definition of Done
A task is considered complete when:
1. All functional requirements are implemented
2. Code quality checks pass
3. Tests provide adequate coverage
4. Documentation is updated
5. No regressions are introduced
6. Security and performance are maintained
7. Integration with existing systems works
8. Error scenarios are handled gracefully

## Common Pitfalls to Avoid
- Don't skip testing edge cases
- Don't ignore compiler warnings
- Don't forget to handle context cancellation
- Don't leave TODO comments in production code
- Don't commit sensitive information (API keys, tokens)
- Don't break backward compatibility without documentation
- Don't introduce new dependencies without justification