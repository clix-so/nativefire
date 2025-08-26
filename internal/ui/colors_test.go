package ui

import (
	"testing"
)

func TestUIFunctions(t *testing.T) {
	// Test that UI functions don't panic and return strings
	tests := []struct {
		name string
		fn   func()
	}{
		{"Header", func() { Header("Test Header") }},
		{"SuccessMsg", func() { SuccessMsg("Success message") }},
		{"WarningMsg", func() { WarningMsg("Warning message") }},
		{"ErrorMsg", func() { ErrorMsg("Error message") }},
		{"InfoMsg", func() { InfoMsg("Info message") }},
		{"RocketMsg", func() { RocketMsg("Rocket message") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			tt.fn()
		})
	}
}

func TestUIFormatters(t *testing.T) {
	// Test Code function
	result := Code("test code")
	if result == "" {
		t.Error("Expected Code() to return formatted string")
	}
	if result == "test code" {
		t.Error("Expected Code() to format the input")
	}

	// Test Platform function
	result = Platform("iOS")
	if result == "" {
		t.Error("Expected Platform() to return formatted string")
	}
	if result == "iOS" {
		t.Error("Expected Platform() to format the input")
	}
}

func TestUIOutputFunctions(t *testing.T) {
	// Test functions that output to console (should not panic)
	Step(1, "Test step")
	Bullet("Test bullet point")
	ProjectHeader("Test Project")
}
