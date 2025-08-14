package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	t.Run("Version Output", func(t *testing.T) {
		// Reset command state
		rootCmd.SetArgs(nil)
		var stdout bytes.Buffer
		rootCmd.SetOut(&stdout)
		rootCmd.SetErr(&stdout)
		rootCmd.SetArgs([]string{"version"})

		err := rootCmd.Execute()
		if err != nil {
			t.Errorf("Version command failed: %v", err)
		}

		output := stdout.String()
		if !strings.Contains(output, "nativefire v") {
			t.Errorf("Expected version output to contain 'nativefire v', got: %s", output)
		}

		if !strings.Contains(output, Version) {
			t.Errorf("Expected version output to contain version '%s', got: %s", Version, output)
		}
	})

	t.Run("Version Variable", func(t *testing.T) {
		if Version == "" {
			t.Error("Version should not be empty")
		}

		// Version should follow semantic versioning pattern
		if !strings.Contains(Version, ".") {
			t.Errorf("Version '%s' should contain dots for semantic versioning", Version)
		}
	})
}
