# @File Mention Expansion System Design

## Overview

Simple design for supporting `@file` mentions in markdown files that get expanded with actual file contents during prompt processing. This feature enhances the existing context processing system with minimal code changes.

## Problem Statement

Users want to reference other files within markdown files using `@file` mentions (e.g., `@COMMANDS.md`, `@FLAGS.md`) and have these automatically expanded to include the actual file contents in the generated prompt.

### Example Use Case

**Input markdown file:**
```markdown
# SuperClaude Entry Point

@COMMANDS.md
@FLAGS.md
@PRINCIPLES.md

Additional context...
```

**Desired output:**
```markdown
# SuperClaude Entry Point

Contents of /Users/kirmadi/.claude/COMMANDS.md

(contents of COMMANDS.md here)


Contents of /Users/kirmadi/.claude/FLAGS.md

(contents of FLAGS.md here)


Contents of /Users/kirmadi/.claude/PRINCIPLES.md

(contents of PRINCIPLES.md here)

Additional context...
```

## Current System Analysis

The existing `processFile()` function in `internal/llm/prompt/prompt.go` simply reads files and prepends a header:

```go
func processFile(filePath string) string {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return ""
    }
    return "# From:" + filePath + "\n" + string(content)
}
```

**Enhancement**: Add @file expansion before returning content.

## Simple Design

### Core Features
- **Pattern**: `@([A-Z_][A-Z0-9_]*\.md)` (e.g., `@COMMANDS.md`, `@FLAGS.md`)
- **Search Order**: Same directory → `commands/` → `.opencode/` → parent directory
- **Output Format**: `Contents of /path/to/file.md\n\n[file contents]\n\n`
- **Recursive Support**: Handle nested @file mentions with cycle detection

## Simple Implementation

### Modified processFile Function

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

## Testing

Add tests to existing `internal/llm/prompt/prompt_test.go`:

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
```

## Benefits

- **Minimal Code**: ~40 lines added to existing file
- **Simple Integration**: Uses existing patterns and functions
- **No New Dependencies**: Uses standard library only
- **Graceful Fallback**: Leaves @file mentions unchanged if files not found
- **Cycle Safety**: Prevents infinite recursion
- **Easy Testing**: Straightforward test scenarios

## Implementation Notes

- **File Pattern**: Regex prevents directory traversal (no `..` allowed)
- **Depth Limit**: Max 10 levels prevents infinite recursion
- **Error Handling**: Silent fallback preserves @mentions if files missing
- **Performance**: Only processes `.md` files, minimal overhead

## Future Enhancements

If needed later:
- Configuration options for search paths
- Caching for frequently referenced files  
- Metrics and logging
- Additional file types support

This simple design delivers the core @file mention functionality with minimal code changes while maintaining the existing system's reliability and performance.