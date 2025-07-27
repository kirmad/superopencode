package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/kirmad/superopencode/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetContextFromPaths(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	_, err := config.Load(tmpDir, false)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	cfg := config.Get()
	cfg.WorkingDir = tmpDir
	cfg.ContextPaths = []string{
		"file.txt",
		"directory/",
	}
	testFiles := []string{
		"file.txt",
		"directory/file_a.txt",
		"directory/file_b.txt",
		"directory/file_c.txt",
	}

	createTestFiles(t, tmpDir, testFiles)

	context := getContextFromPaths()
	expectedContext := fmt.Sprintf("# From:%s/file.txt\nfile.txt: test content\n# From:%s/directory/file_a.txt\ndirectory/file_a.txt: test content\n# From:%s/directory/file_b.txt\ndirectory/file_b.txt: test content\n# From:%s/directory/file_c.txt\ndirectory/file_c.txt: test content", tmpDir, tmpDir, tmpDir, tmpDir)
	assert.Equal(t, expectedContext, context)
}

func createTestFiles(t *testing.T, tmpDir string, testFiles []string) {
	t.Helper()
	for _, path := range testFiles {
		fullPath := filepath.Join(tmpDir, path)
		if path[len(path)-1] == '/' {
			err := os.MkdirAll(fullPath, 0755)
			require.NoError(t, err)
		} else {
			dir := filepath.Dir(fullPath)
			err := os.MkdirAll(dir, 0755)
			require.NoError(t, err)
			err = os.WriteFile(fullPath, []byte(path+": test content"), 0644)
			require.NoError(t, err)
		}
	}
}

func TestProcessFile_WithFileExpansion(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create test files
	refFile := filepath.Join(tmpDir, "REF.md")
	err := os.WriteFile(refFile, []byte("Reference content"), 0644)
	require.NoError(t, err)
	
	mainFile := filepath.Join(tmpDir, "main.md")
	err = os.WriteFile(mainFile, []byte("Header\n@REF.md\nFooter"), 0644)
	require.NoError(t, err)
	
	result := processFile(mainFile)
	
	// Should contain original content unchanged
	assert.Contains(t, result, "# From:")
	assert.Contains(t, result, "Header")
	assert.Contains(t, result, "@REF.md") // Original @mention should remain
	assert.Contains(t, result, "Footer")
	
	// Should also contain appended file content
	assert.Contains(t, result, "Contents of")
	assert.Contains(t, result, "Reference content")
}

func TestExpandFileReferences_Nested(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create child file
	childFile := filepath.Join(tmpDir, "CHILD.md")
	err := os.WriteFile(childFile, []byte("Child content"), 0644)
	require.NoError(t, err)
	
	// Create parent file that references child
	parentFile := filepath.Join(tmpDir, "PARENT.md")
	err = os.WriteFile(parentFile, []byte("Parent: @CHILD.md"), 0644)
	require.NoError(t, err)
	
	// Create main file that references parent
	mainFile := filepath.Join(tmpDir, "main.md")
	content := "Main content\n@PARENT.md"
	
	result := expandFileReferences(content, mainFile)
	
	// Should contain original content
	assert.Contains(t, result, "Main content")
	assert.Contains(t, result, "@PARENT.md")
	
	// Should contain nested file contents
	assert.Contains(t, result, "Parent:")
	assert.Contains(t, result, "Child content")
}

func TestExpandFileReferences_CycleDetection(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create files that reference each other
	fileA := filepath.Join(tmpDir, "A.md")
	err := os.WriteFile(fileA, []byte("File A content @B.md"), 0644)
	require.NoError(t, err)
	
	fileB := filepath.Join(tmpDir, "B.md")
	err = os.WriteFile(fileB, []byte("File B content @A.md"), 0644)
	require.NoError(t, err)
	
	// Should not infinite loop
	result := expandFileReferences("@A.md", fileA)
	
	assert.Contains(t, result, "File A content")
	assert.Contains(t, result, "File B content")
}

func TestExpandFileReferences_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	
	mainFile := filepath.Join(tmpDir, "main.md")
	err := os.WriteFile(mainFile, []byte("Header\n@MISSING.md\nFooter"), 0644)
	require.NoError(t, err)
	
	result := expandFileReferences("Header\n@MISSING.md\nFooter", mainFile)
	
	// Should leave @MISSING.md unchanged since file doesn't exist
	assert.Contains(t, result, "@MISSING.md")
	assert.Contains(t, result, "Header")
	assert.Contains(t, result, "Footer")
}

func TestExpandFileReferences_NonMarkdownFile(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a non-markdown file
	txtFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(txtFile, []byte("@REF.md should not expand"), 0644)
	require.NoError(t, err)
	
	result := processFile(txtFile)
	
	// Should not process @file mentions in non-markdown files
	assert.Contains(t, result, "@REF.md should not expand")
}

func TestFindFile_SearchDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create subdirectories
	commandsDir := filepath.Join(tmpDir, "commands")
	err := os.MkdirAll(commandsDir, 0755)
	require.NoError(t, err)
	
	opencodeDir := filepath.Join(tmpDir, ".opencode")
	err = os.MkdirAll(opencodeDir, 0755)
	require.NoError(t, err)
	
	// Create test file in commands directory
	testFile := filepath.Join(commandsDir, "TEST.md")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)
	
	// Test finding file from main directory
	basePath := filepath.Join(tmpDir, "main.md")
	result := findFile(basePath, "TEST.md")
	
	assert.Equal(t, testFile, result)
}

func TestExpandFileReferences_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create multiple referenced files
	file1 := filepath.Join(tmpDir, "FILE1.md")
	err := os.WriteFile(file1, []byte("Content 1"), 0644)
	require.NoError(t, err)
	
	file2 := filepath.Join(tmpDir, "FILE2.md")
	err = os.WriteFile(file2, []byte("Content 2"), 0644)
	require.NoError(t, err)
	
	// Create main file that references both
	mainContent := "Header\n@FILE1.md\nMiddle\n@FILE2.md\nFooter"
	result := expandFileReferences(mainContent, filepath.Join(tmpDir, "main.md"))
	
	// Should contain original content
	assert.Contains(t, result, "Header")
	assert.Contains(t, result, "@FILE1.md")
	assert.Contains(t, result, "Middle")
	assert.Contains(t, result, "@FILE2.md")
	assert.Contains(t, result, "Footer")
	
	// Should contain appended file contents
	assert.Contains(t, result, "Content 1")
	assert.Contains(t, result, "Content 2")
}

func TestExpandFileReferences_RelativePaths(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create subdirectory structure
	subDir := filepath.Join(tmpDir, "sm")
	err := os.MkdirAll(subDir, 0755)
	require.NoError(t, err)
	
	parentDir := filepath.Join(tmpDir, "parent")
	err = os.MkdirAll(parentDir, 0755)
	require.NoError(t, err)
	
	// Create files in different directories
	subFile := filepath.Join(subDir, "TEST.md")
	err = os.WriteFile(subFile, []byte("Subfolder content"), 0644)
	require.NoError(t, err)
	
	parentFile := filepath.Join(parentDir, "PARENT.md")
	err = os.WriteFile(parentFile, []byte("Parent dir content"), 0644)
	require.NoError(t, err)
	
	// Create main file in tmpDir that references files with relative paths
	mainFile := filepath.Join(tmpDir, "main.md")
	content := "Main content\n@sm/TEST.md\n@parent/PARENT.md"
	
	result := expandFileReferences(content, mainFile)
	
	// Should contain original content
	assert.Contains(t, result, "Main content")
	assert.Contains(t, result, "@sm/TEST.md")
	assert.Contains(t, result, "@parent/PARENT.md")
	
	// Should contain file contents from subdirectories
	assert.Contains(t, result, "Subfolder content")
	assert.Contains(t, result, "Parent dir content")
}

func TestExpandFileReferences_DotDotRelativePaths(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create directory structure: tmpDir/subdir/main.md references tmpDir/parent.md
	subDir := filepath.Join(tmpDir, "subdir")
	err := os.MkdirAll(subDir, 0755)
	require.NoError(t, err)
	
	// Create parent file in tmpDir
	parentFile := filepath.Join(tmpDir, "PARENT.md")
	err = os.WriteFile(parentFile, []byte("Parent directory content"), 0644)
	require.NoError(t, err)
	
	// Create main file in subdir that references parent with ../
	mainFile := filepath.Join(subDir, "main.md")
	content := "Subdir content\n@../PARENT.md"
	
	result := expandFileReferences(content, mainFile)
	
	// Should contain original content
	assert.Contains(t, result, "Subdir content")
	assert.Contains(t, result, "@../PARENT.md")
	
	// Should contain parent file content
	assert.Contains(t, result, "Parent directory content")
}

func TestFindFile_RelativePaths(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create subdirectory and file
	subDir := filepath.Join(tmpDir, "sm")
	err := os.MkdirAll(subDir, 0755)
	require.NoError(t, err)
	
	testFile := filepath.Join(subDir, "TEST.md")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)
	
	// Test finding file with relative path
	basePath := filepath.Join(tmpDir, "main.md")
	result := findFile(basePath, "sm/TEST.md")
	
	abs, _ := filepath.Abs(testFile)
	assert.Equal(t, abs, result)
}

func TestFindFile_DotDotPaths(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create parent file
	parentFile := filepath.Join(tmpDir, "PARENT.md")
	err := os.WriteFile(parentFile, []byte("parent content"), 0644)
	require.NoError(t, err)
	
	// Create subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	require.NoError(t, err)
	
	// Test finding parent file from subdirectory using ../
	basePath := filepath.Join(subDir, "main.md")
	result := findFile(basePath, "../PARENT.md")
	
	abs, _ := filepath.Abs(parentFile)
	assert.Equal(t, abs, result)
}
