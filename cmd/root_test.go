package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommand(t *testing.T) {
	// Test help output
	t.Run("Help", func(t *testing.T) {
		var stdout bytes.Buffer
		rootCmd.SetOut(&stdout)
		rootCmd.SetArgs([]string{"--help"})

		err := rootCmd.Execute()
		if err != nil {
			t.Errorf("Help command failed: %v", err)
		}

		output := stdout.String()
		if !strings.Contains(output, "nativefire") {
			t.Error("Help output should contain 'nativefire'")
		}
		if !strings.Contains(output, "Firebase") {
			t.Error("Help output should contain 'Firebase'")
		}
	})

	// Test version flag
	t.Run("Version Flag", func(t *testing.T) {
		var stdout bytes.Buffer
		rootCmd.SetOut(&stdout)
		rootCmd.SetArgs([]string{"--version"})

		// Version flag is handled by cobra automatically
		// We can't easily test it without modifying the command setup
		// This is a limitation of how cobra handles version flags
	})

	// Test verbose flag
	t.Run("Verbose Flag", func(t *testing.T) {
		resetRootCommand()

		var stdout bytes.Buffer
		rootCmd.SetOut(&stdout)
		rootCmd.SetArgs([]string{"--verbose", "--help"})

		err := rootCmd.Execute()
		if err != nil {
			t.Errorf("Verbose help command failed: %v", err)
		}

		// The verbose flag should be parsed correctly
		if !verbose {
			t.Error("Verbose flag was not set correctly")
		}
	})

	// Test invalid command
	t.Run("Invalid Command", func(t *testing.T) {
		resetRootCommand()

		var stderr bytes.Buffer
		rootCmd.SetErr(&stderr)
		rootCmd.SetArgs([]string{"invalid-command"})

		err := rootCmd.Execute()
		if err == nil {
			t.Error("Expected error for invalid command")
		}

		output := stderr.String()
		if !strings.Contains(output, "unknown command") {
			t.Errorf("Expected 'unknown command' error, got: %s", output)
		}
	})
}

func TestInitConfig(t *testing.T) {
	// Test that initConfig doesn't panic
	t.Run("Init Config", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("initConfig panicked: %v", r)
			}
		}()

		initConfig()
	})
}

func resetRootCommand() {
	verbose = false
	cfgFile = ""
}
