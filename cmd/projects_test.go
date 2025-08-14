package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestProjectsListCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectSkip     string
		expectContains []string
	}{
		{
			name:           "Projects list help",
			args:           []string{"projects", "list", "--help"},
			expectError:    false,
			expectContains: []string{"Firebase Projects Overview", "nativefire projects list"},
		},
		{
			name:        "Projects list integration",
			args:        []string{"projects", "list"},
			expectError: false,
			expectSkip:  "Integration test requiring Firebase CLI",
		},
		{
			name:           "Projects list verbose",
			args:           []string{"projects", "list", "--verbose"},
			expectError:    false,
			expectSkip:     "Integration test requiring Firebase CLI",
			expectContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectSkip != "" && testing.Short() {
				t.Skip(tt.expectSkip)
			}

			// Reset command state
			resetProjectsCommand()

			var stdout, stderr bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none. Stdout: %s, Stderr: %s", stdout.String(), stderr.String())
					return
				}
			} else {
				if err != nil {
					// Check if it's a Firebase CLI related error that we should skip
					errorMsg := err.Error()
					if strings.Contains(errorMsg, "firebase CLI not found") ||
						strings.Contains(errorMsg, "not authenticated") ||
						strings.Contains(errorMsg, "failed to list projects") {
						t.Skipf("Firebase CLI not available or not authenticated: %v", err)
						return
					}
					t.Errorf("Unexpected error: %v. Stdout: %s, Stderr: %s", err, stdout.String(), stderr.String())
					return
				}

				output := stdout.String()
				for _, expected := range tt.expectContains {
					if !strings.Contains(output, expected) {
						t.Errorf("Expected output to contain '%s', but it doesn't. Output: %s", expected, output)
					}
				}
			}
		})
	}
}

func TestProjectsSelectCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectSkip     string
		expectContains []string
	}{
		{
			name:           "Projects select help",
			args:           []string{"projects", "select", "--help"},
			expectError:    false,
			expectContains: []string{"Interactive Project Selection", "nativefire projects select"},
		},
		{
			name:        "Projects select integration",
			args:        []string{"projects", "select"},
			expectError: false,
			expectSkip:  "Integration test requiring Firebase CLI and user input",
		},
		{
			name:        "Projects select with use flag",
			args:        []string{"projects", "select", "--use"},
			expectError: false,
			expectSkip:  "Integration test requiring Firebase CLI and user input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectSkip != "" && testing.Short() {
				t.Skip(tt.expectSkip)
			}

			// Reset command state
			resetProjectsCommand()
			autoUse = false // Reset the flag

			var stdout, stderr bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none. Stdout: %s, Stderr: %s", stdout.String(), stderr.String())
					return
				}
			} else {
				if err != nil {
					// Check if it's a Firebase CLI related error that we should skip
					errorMsg := err.Error()
					if strings.Contains(errorMsg, "firebase CLI not found") ||
						strings.Contains(errorMsg, "not authenticated") ||
						strings.Contains(errorMsg, "failed to list projects") ||
						strings.Contains(errorMsg, "failed to read input") {
						t.Skipf("Firebase CLI not available, not authenticated, or input error: %v", err)
						return
					}
					t.Errorf("Unexpected error: %v. Stdout: %s, Stderr: %s", err, stdout.String(), stderr.String())
					return
				}

				output := stdout.String()
				for _, expected := range tt.expectContains {
					if !strings.Contains(output, expected) {
						t.Errorf("Expected output to contain '%s', but it doesn't. Output: %s", expected, output)
					}
				}
			}
		})
	}
}

func TestProjectsCommand(t *testing.T) {
	t.Run("Projects help", func(t *testing.T) {
		resetProjectsCommand()

		var stdout bytes.Buffer
		rootCmd.SetOut(&stdout)
		rootCmd.SetArgs([]string{"projects", "--help"})

		err := rootCmd.Execute()
		if err != nil {
			t.Errorf("Projects help command failed: %v", err)
			return
		}

		output := stdout.String()
		expectedContent := []string{
			"Firebase projects",
			"list",
			"select",
		}

		for _, expected := range expectedContent {
			if !strings.Contains(output, expected) {
				t.Errorf("Expected help output to contain '%s', but it doesn't. Output: %s", expected, output)
			}
		}
	})

	t.Run("Projects without subcommand", func(t *testing.T) {
		resetProjectsCommand()

		var stdout bytes.Buffer
		rootCmd.SetOut(&stdout)
		rootCmd.SetArgs([]string{"projects"})

		err := rootCmd.Execute()
		if err != nil {
			t.Errorf("Projects command without subcommand failed: %v", err)
			return
		}

		// Should show help when no subcommand is provided
		output := stdout.String()
		if !strings.Contains(output, "Firebase projects") {
			t.Errorf("Expected help output when no subcommand provided. Output: %s", output)
		}
	})
}

func TestConfigureCommandWithProjectSelection(t *testing.T) {
	// Test that configure command now works without --project flag
	tests := []struct {
		name        string
		args        []string
		expectError bool
		expectSkip  string
	}{
		{
			name:        "Configure without project flag should prompt",
			args:        []string{"configure"},
			expectError: true,
			expectSkip:  "Integration test requiring Firebase CLI and user input",
		},
		{
			name:        "Configure with project validation",
			args:        []string{"configure", "--project", "invalid-project-id"},
			expectError: true,
			expectSkip:  "Integration test requiring Firebase CLI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectSkip != "" && testing.Short() {
				t.Skip(tt.expectSkip)
			}

			// Reset command state
			resetProjectsCommand()

			var stdout, stderr bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none. Stdout: %s, Stderr: %s", stdout.String(), stderr.String())
					return
				}
			} else {
				if err != nil {
					// Check if it's a Firebase CLI related error that we should skip
					errorMsg := err.Error()
					if strings.Contains(errorMsg, "firebase CLI not found") ||
						strings.Contains(errorMsg, "not authenticated") ||
						strings.Contains(errorMsg, "failed to fetch projects") ||
						strings.Contains(errorMsg, "failed to read input") {
						t.Skipf("Firebase CLI not available, not authenticated, or input error: %v", err)
						return
					}
					t.Errorf("Unexpected error: %v. Stdout: %s, Stderr: %s", err, stdout.String(), stderr.String())
				}
			}
		})
	}
}

func resetProjectsCommand() {
	verbose = false
	cfgFile = ""
	projectID = ""
	platformFlag = ""
	autoDetect = true
	appID = ""
	bundleID = ""
	packageName = ""
	autoUse = false
}
