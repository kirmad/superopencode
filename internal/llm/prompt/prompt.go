package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/kirmad/superopencode/internal/config"
	"github.com/kirmad/superopencode/internal/llm/models"
	"github.com/kirmad/superopencode/internal/logging"
)

func GetAgentPrompt(agentName config.AgentName, provider models.ModelProvider) string {
	basePrompt := ""
	switch agentName {
	case config.AgentCoder:
		basePrompt = CoderPrompt(provider)
	case config.AgentTitle:
		basePrompt = TitlePrompt(provider)
	case config.AgentTask:
		basePrompt = TaskPrompt(provider)
	case config.AgentSummarizer:
		basePrompt = SummarizerPrompt(provider)
	default:
		basePrompt = "You are a helpful assistant"
	}

	if agentName == config.AgentCoder || agentName == config.AgentTask {
		// Add context from project-specific instruction files if they exist
		contextContent := getContextFromPaths()
		logging.Debug("Context content", "Context", contextContent)
		if contextContent != "" {
			return fmt.Sprintf("%s\n\n# Project-Specific Context\n Make sure to follow the instructions in the context below\n%s", basePrompt, contextContent)
		}
	}
	return basePrompt
}

var (
	onceContext    sync.Once
	contextContent string
)

func getContextFromPaths() string {
	onceContext.Do(func() {
		var (
			cfg          = config.Get()
			workDir      = cfg.WorkingDir
			contextPaths = cfg.ContextPaths
		)

		contextContent = processContextPaths(workDir, contextPaths)
	})

	return contextContent
}

func processContextPaths(workDir string, paths []string) string {
	var (
		wg       sync.WaitGroup
		resultCh = make(chan string)
	)

	// Track processed files to avoid duplicates
	processedFiles := make(map[string]bool)
	var processedMutex sync.Mutex

	for _, path := range paths {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()

			if strings.HasSuffix(p, "/") {
				filepath.WalkDir(filepath.Join(workDir, p), func(path string, d os.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if !d.IsDir() {
						// Check if we've already processed this file (case-insensitive)
						processedMutex.Lock()
						lowerPath := strings.ToLower(path)
						if !processedFiles[lowerPath] {
							processedFiles[lowerPath] = true
							processedMutex.Unlock()

							if result := processFile(path); result != "" {
								resultCh <- result
							}
						} else {
							processedMutex.Unlock()
						}
					}
					return nil
				})
			} else {
				fullPath := filepath.Join(workDir, p)

				// Check if we've already processed this file (case-insensitive)
				processedMutex.Lock()
				lowerPath := strings.ToLower(fullPath)
				if !processedFiles[lowerPath] {
					processedFiles[lowerPath] = true
					processedMutex.Unlock()

					result := processFile(fullPath)
					if result != "" {
						resultCh <- result
					}
				} else {
					processedMutex.Unlock()
				}
			}
		}(path)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	results := make([]string, 0)
	for result := range resultCh {
		results = append(results, result)
	}

	return strings.Join(results, "\n")
}

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

// ExpandFileReferences processes @file mentions in content and expands them with file contents
// This is a public function that can be used by other packages like slash command processing
func ExpandFileReferences(content string) string {
	// Use current working directory as base path for slash commands
	cfg := config.Get()
	// Create a dummy file path in the working directory for findFile compatibility
	basePath := filepath.Join(cfg.WorkingDir, "dummy.md")
	return expandFileReferences(content, basePath)
}

// ExpandFileReferencesWithBasePath processes @file mentions using a specific base file path
// This allows expansion relative to a specific command file location
func ExpandFileReferencesWithBasePath(content, basePath string) string {
	return expandFileReferences(content, basePath)
}

func expandFileReferences(content, basePath string) string {
	visited := make(map[string]bool)
	
	// Find all @file mentions and collect their content to append
	var additionalContent strings.Builder
	collectFileReferences(content, basePath, visited, 0, &additionalContent)
	
	// Return original content with additional file contents appended
	if additionalContent.Len() > 0 {
		return content + "\n\n" + additionalContent.String()
	}
	return content
}

func collectFileReferences(content, basePath string, visited map[string]bool, depth int, result *strings.Builder) {
	if depth > 10 {
		return
	}
	
	// Updated pattern to support relative paths and subfolders
	pattern := regexp.MustCompile(`@([a-zA-Z0-9_./\\-]+\.md)`)
	matches := pattern.FindAllStringSubmatch(content, -1)
	
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		
		filePath := match[1]
		if resolvedPath := findFile(basePath, filePath); resolvedPath != "" {
			// Use resolved absolute path for cycle detection
			if visited[resolvedPath] {
				continue // Skip if already visited to prevent cycles
			}
			visited[resolvedPath] = true
			
			if fileContent, err := os.ReadFile(resolvedPath); err == nil {
				result.WriteString(fmt.Sprintf("Contents of %s\n\n%s\n\n", resolvedPath, string(fileContent)))
				
				// Recursively collect references from this file
				collectFileReferences(string(fileContent), resolvedPath, visited, depth+1, result)
			}
		}
	}
}

func findFile(basePath, filePath string) string {
	// If it's already an absolute path, check if it exists
	if filepath.IsAbs(filePath) {
		if _, err := os.Stat(filePath); err == nil {
			return filePath
		}
		return ""
	}
	
	// Get the directory of the base file
	baseDir := filepath.Dir(basePath)
	
	// If it's a relative path starting with ./ or ../, resolve it relative to baseDir
	if strings.HasPrefix(filePath, "./") || strings.HasPrefix(filePath, "../") {
		resolvedPath := filepath.Join(baseDir, filePath)
		if abs, err := filepath.Abs(resolvedPath); err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs
			}
		}
		return ""
	}
	
	// Search in multiple directories for the file
	searchDirs := []string{
		baseDir,                                          // Same directory as base file
		filepath.Join(baseDir, "commands"),               // commands subdirectory
		filepath.Join(baseDir, ".opencode"),              // .opencode subdirectory
		filepath.Dir(baseDir),                            // Parent directory
	}
	
	for _, dir := range searchDirs {
		fullPath := filepath.Join(dir, filePath)
		if abs, err := filepath.Abs(fullPath); err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs
			}
		}
	}
	
	return ""
}

