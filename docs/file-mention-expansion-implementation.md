# @File Mention Expansion Implementation Guide

## Overview

Simple implementation guide for adding @file mention expansion to the existing `processFile()` function with minimal code changes.

## Implementation Checklist

- [ ] 1. Add expansion functions to `internal/llm/prompt/prompt.go`
- [ ] 2. Modify `processFile()` function
- [ ] 3. Add tests to existing test file
- [ ] 4. Test and validate

## File Changes

Only one file needs modification:
```
internal/llm/prompt/prompt.go (add ~40 lines)
```

Optional:
```
internal/llm/prompt/prompt_test.go (add tests)
```

## Implementation Steps

### Step 1: Add Functions to prompt.go

Add these functions to `internal/llm/prompt/prompt.go`:

```go
func expandFileReferences(content, basePath string) string {
    visited := make(map[string]bool)
    return expandWithCycleDetection(content, basePath, visited, 0)
}

func expandWithCycleDetection(content, basePath string, visited map[string]bool, depth int) string {
    if depth > 10 || visited[basePath] {
        return content
    }
    visited[basePath] = true
    
    pattern := regexp.MustCompile(`@([A-Z_][A-Z0-9_]*\.md)`)
    return pattern.ReplaceAllStringFunc(content, func(match string) string {
        fileName := strings.TrimPrefix(match, "@")
        if resolvedPath := findFile(basePath, fileName); resolvedPath != "" {
            if fileContent, err := os.ReadFile(resolvedPath); err == nil {
                expanded := expandWithCycleDetection(string(fileContent), resolvedPath, visited, depth+1)
                return fmt.Sprintf("Contents of %s\n\n%s\n\n", resolvedPath, expanded)
            }
        }
        return match // Leave unchanged if not found
    })
}

func findFile(basePath, fileName string) string {
    searchDirs := []string{
        filepath.Dir(basePath),
        filepath.Join(filepath.Dir(basePath), "commands"),
        filepath.Join(filepath.Dir(basePath), ".opencode"),
    }
    
    for _, dir := range searchDirs {
        path := filepath.Join(dir, fileName)
        if _, err := os.Stat(path); err == nil {
            abs, _ := filepath.Abs(path)
            return abs
        }
    }
    return ""
}
```

### Step 2: Modify processFile Function

Update the existing `processFile()` function in `internal/llm/prompt/prompt.go`:

```go
func processFile(filePath string) string {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return ""
    }
    
    processedContent := string(content)
    
    // Process @file mentions for markdown files
    if strings.HasSuffix(strings.ToLower(filePath), ".md") {
        processedContent = expandFileReferences(processedContent, filePath)
    }
    
    return "# From:" + filePath + "\n" + processedContent
}
```

### Step 3: Add Tests

Add tests to `internal/llm/prompt/prompt_test.go`:

```go
func TestProcessFile_WithFileExpansion(t *testing.T) {
    tmpDir, _ := os.MkdirTemp("", "test")
    defer os.RemoveAll(tmpDir)
    
    // Create test files
    refFile := filepath.Join(tmpDir, "REF.md")
    os.WriteFile(refFile, []byte("Reference content"), 0644)
    
    mainFile := filepath.Join(tmpDir, "main.md")
    os.WriteFile(mainFile, []byte("Header\n@REF.md\nFooter"), 0644)
    
    result := processFile(mainFile)
    
    assert.Contains(t, result, "# From:")
    assert.Contains(t, result, "Contents of")
    assert.Contains(t, result, "Reference content")
}

func TestExpandFileReferences_Nested(t *testing.T) {
    tmpDir, _ := os.MkdirTemp("", "test")
    defer os.RemoveAll(tmpDir)
    
    childFile := filepath.Join(tmpDir, "CHILD.md")
    os.WriteFile(childFile, []byte("Child content"), 0644)
    
    parentFile := filepath.Join(tmpDir, "PARENT.md")
    os.WriteFile(parentFile, []byte("Parent: @CHILD.md"), 0644)
    
    result := expandFileReferences("@PARENT.md", parentFile)
    
    assert.Contains(t, result, "Parent:")
    assert.Contains(t, result, "Child content")
}
```

### Step 4: Testing

Run tests to verify implementation:

```bash
# Run tests
go test ./internal/llm/prompt/

# Run with coverage
go test -cover ./internal/llm/prompt/
```

## Complete Implementation Summary

**Total Code Added**: ~40 lines to one existing file

**Files Modified**:
- `internal/llm/prompt/prompt.go` (add 3 functions + modify 1 function)

**Files Added**:
- None (optional: add tests to existing test file)

**Key Benefits**:
- Minimal code footprint
- Uses existing patterns
- Graceful error handling
- Zero configuration required
- Easy to test and debug

This simplified implementation delivers all the core functionality while maintaining the principle of keeping things simple and following existing patterns.