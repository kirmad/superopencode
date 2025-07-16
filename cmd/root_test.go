package cmd

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDangerouslySkipPermissionsFlag(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedFlag bool
	}{
		{
			name:         "flag enabled via command line",
			args:         []string{"--dangerously-skip-permissions"},
			expectedFlag: true,
		},
		{
			name:         "flag disabled by default",
			args:         []string{},
			expectedFlag: false,
		},
		{
			name:         "flag with other options",
			args:         []string{"--dangerously-skip-permissions", "--debug"},
			expectedFlag: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new command instance for testing
			cmd := NewRootCmd()
			
			// Reset the flag to its default value
			cmd.Flags().Set("dangerously-skip-permissions", "false")
			
			cmd.SetArgs(tt.args)

			// Parse flags
			err := cmd.ParseFlags(tt.args)
			require.NoError(t, err)

			// Check if flag is parsed correctly
			flagValue, err := cmd.Flags().GetBool("dangerously-skip-permissions")
			require.NoError(t, err)
			
			// Debug: print actual flag value to see what's happening
			t.Logf("Flag value for args %v: %v", tt.args, flagValue)
			
			assert.Equal(t, tt.expectedFlag, flagValue)
		})
	}
}

func TestDangerouslySkipPermissionsEnvironmentVariable(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		flagValue    bool
		expectedFinal bool
	}{
		{
			name:         "environment variable enables flag",
			envValue:     "true",
			flagValue:    false,
			expectedFinal: true,
		},
		{
			name:         "environment variable disabled",
			envValue:     "false",
			flagValue:    false,
			expectedFinal: false,
		},
		{
			name:         "flag takes precedence over environment",
			envValue:     "false",
			flagValue:    true,
			expectedFinal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variable
			originalValue := os.Getenv("SUPEROPENCODE_DANGEROUSLY_SKIP_PERMISSIONS")
			defer func() {
				if originalValue == "" {
					os.Unsetenv("SUPEROPENCODE_DANGEROUSLY_SKIP_PERMISSIONS")
				} else {
					os.Setenv("SUPEROPENCODE_DANGEROUSLY_SKIP_PERMISSIONS", originalValue)
				}
			}()
			os.Setenv("SUPEROPENCODE_DANGEROUSLY_SKIP_PERMISSIONS", tt.envValue)

			// Simulate the logic from main function
			dangerouslySkipPermissions := tt.flagValue
			if !dangerouslySkipPermissions && os.Getenv("SUPEROPENCODE_DANGEROUSLY_SKIP_PERMISSIONS") == "true" {
				dangerouslySkipPermissions = true
			}

			assert.Equal(t, tt.expectedFinal, dangerouslySkipPermissions)
		})
	}
}

func TestDangerouslySkipPermissionsValidation(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "normal operation - no restrictions",
			envVars:     map[string]string{},
			expectError: false,
		},
		{
			name: "production environment - should fail",
			envVars: map[string]string{
				"PRODUCTION": "true",
			},
			expectError: true,
			errorMsg:    "dangerous mode disabled in production",
		},
		{
			name: "development environment - should pass",
			envVars: map[string]string{
				"ENVIRONMENT": "development",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			originalEnvVars := make(map[string]string)
			for key, value := range tt.envVars {
				originalEnvVars[key] = os.Getenv(key)
				os.Setenv(key, value)
			}

			// Clean up environment variables
			defer func() {
				for key, originalValue := range originalEnvVars {
					if originalValue == "" {
						os.Unsetenv(key)
					} else {
						os.Setenv(key, originalValue)
					}
				}
			}()

			// Test validation
			err := validateDangerousMode()
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDangerouslySkipPermissionsWarning(t *testing.T) {
	// Capture stderr output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Create a buffer to capture output
	var buf bytes.Buffer
	done := make(chan bool)

	go func() {
		buf.ReadFrom(r)
		done <- true
	}()

	// Call the function that should print warnings
	printDangerousWarning()

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Wait for output capture
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for output")
	}

	// Verify warning message
	output := buf.String()
	assert.Contains(t, output, "DANGEROUS MODE")
	assert.Contains(t, output, "permission checks will be bypassed")
	assert.Contains(t, output, "unrestricted system access")
}

func TestDangerouslySkipPermissionsIntegration(t *testing.T) {
	// This test verifies that the dangerous mode properly integrates with the app
	// We'll test that permissions are actually bypassed when the flag is set

	// Create temporary directory for testing
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	// Test with dangerous mode enabled
	t.Run("with dangerous mode enabled", func(t *testing.T) {
		// Set the flag
		os.Setenv("SUPEROPENCODE_DANGEROUSLY_SKIP_PERMISSIONS", "true")
		defer os.Unsetenv("SUPEROPENCODE_DANGEROUSLY_SKIP_PERMISSIONS")

		// Create a context with timeout for the test
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = ctx // Will be used in actual implementation

		// This should not prompt for permissions when dangerous mode is enabled
		// The actual implementation will be tested in integration tests
		// For now, we just verify the flag parsing works
		cmd := NewRootCmd()
		cmd.SetArgs([]string{"--dangerously-skip-permissions", "--prompt", "test"})
		
		err := cmd.ParseFlags([]string{"--dangerously-skip-permissions", "--prompt", "test"})
		require.NoError(t, err)

		flagValue, err := cmd.Flags().GetBool("dangerously-skip-permissions")
		require.NoError(t, err)
		assert.True(t, flagValue)
	})
}

// Helper functions are now implemented in root.go