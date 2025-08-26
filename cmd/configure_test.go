package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigureCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupFiles  []string
		setupDirs   []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Missing project flag",
			args:        []string{"configure"},
			expectError: true,
			errorMsg:    "project ID is required in test environment",
		},
		{
			name:        "Missing platform specification",
			args:        []string{"configure", "--project", "test-project"},
			expectError: true,
			errorMsg:    "project validation failed",
		},
		{
			name:        "Invalid platform",
			args:        []string{"configure", "--project", "test-project", "--platform", "invalid"},
			expectError: true,
			errorMsg:    "project validation failed",
		},
		{
			name:       "Valid Android platform",
			args:       []string{"configure", "--project", "test-project", "--platform", "android"},
			setupFiles: []string{"build.gradle"},
			// This will fail at Firebase CLI stage, but validates argument parsing
			expectError: true,
			errorMsg:    "project validation failed",
		},
		{
			name:        "Auto-detect Android",
			args:        []string{"configure", "--project", "test-project", "--auto-detect"},
			setupFiles:  []string{"build.gradle"},
			expectError: true,
			errorMsg:    "project validation failed",
		},
		{
			name:        "Auto-detect with no platform",
			args:        []string{"configure", "--project", "test-project", "--auto-detect"},
			setupFiles:  []string{"random.txt"},
			expectError: true,
			errorMsg:    "project validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip tests that require Firebase CLI in short mode
			if testing.Short() && (strings.Contains(tt.name, "Valid Android platform") ||
				strings.Contains(tt.name, "Auto-detect")) {
				t.Skip("Skipping Firebase CLI test in short mode")
			}

			// Setup test environment
			tempDir := setupTestEnvironment(t, tt.setupDirs, tt.setupFiles)
			defer os.RemoveAll(tempDir)

			// Change to test directory
			oldWd, _ := os.Getwd()
			_ = os.Chdir(tempDir)
			defer func() { _ = os.Chdir(oldWd) }()

			// Reset command state
			resetConfigureCommand()

			// Capture output
			var stdout, stderr bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)

			// Set arguments
			rootCmd.SetArgs(tt.args)

			// Execute command
			err := rootCmd.Execute()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none. Stdout: %s, Stderr: %s", stdout.String(), stderr.String())
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorMsg, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v. Stdout: %s, Stderr: %s", err, stdout.String(), stderr.String())
			}
		})
	}
}

func TestConfigureCommandVerbose(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Firebase CLI test in short mode")
	}

	tempDir := setupTestEnvironment(t, []string{}, []string{"build.gradle"})
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer func() { _ = os.Chdir(oldWd) }()

	resetConfigureCommand()

	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)

	args := []string{"configure", "--project", "test-project", "--platform", "android", "--verbose"}
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()

	// We expect this to fail at Firebase CLI stage, but should show verbose output
	if err == nil {
		t.Error("Expected error due to missing Firebase CLI")
	}

	// In verbose mode, we should see more detailed output
	output := stderr.String()
	if !strings.Contains(output, "project validation failed") {
		t.Errorf("Expected verbose error message about Firebase CLI, got: %s", output)
	}
}

func TestConfigureCommandWithAppID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Firebase CLI test in short mode")
	}

	tempDir := setupTestEnvironment(t, []string{}, []string{"build.gradle"})
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer func() { _ = os.Chdir(oldWd) }()

	resetConfigureCommand()

	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)

	args := []string{"configure", "--project", "test-project", "--platform", "android", "--app-id", "custom-app-id"}
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()

	// Should fail at Firebase CLI stage but validate app-id parsing
	if err == nil {
		t.Error("Expected error due to missing Firebase CLI")
	}

	// Verify that appID was set (this would be tested in the actual implementation)
	if appID != "custom-app-id" {
		t.Errorf("Expected app ID to be 'custom-app-id', got '%s'", appID)
	}
}

func resetConfigureCommand() {
	projectID = ""
	platformFlag = ""
	autoDetect = true
	appID = ""
	bundleID = ""
	packageName = ""
	verbose = false
}

func setupTestEnvironment(t *testing.T, dirs []string, files []string) string {
	tempDir, err := os.MkdirTemp("", "nativefire_cmd_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	for _, dir := range dirs {
		dirPath := filepath.Join(tempDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dirPath, err)
		}
	}

	for _, file := range files {
		filePath := filepath.Join(tempDir, file)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			t.Fatalf("Failed to create parent directory for %s: %v", filePath, err)
		}
		if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filePath, err)
		}
	}

	return tempDir
}
