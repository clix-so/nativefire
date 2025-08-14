package ui

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	// Primary colors for branding
	Primary   = color.New(color.FgHiBlue, color.Bold)
	Secondary = color.New(color.FgHiCyan)

	// Status colors
	Success = color.New(color.FgHiGreen, color.Bold)
	Warning = color.New(color.FgHiYellow, color.Bold)
	Error   = color.New(color.FgHiRed, color.Bold)
	Info    = color.New(color.FgHiBlue)

	// Text colors
	Bold   = color.New(color.Bold)
	Dim    = color.New(color.Faint)
	Italic = color.New(color.Italic)

	// Special colors
	Fire   = color.New(color.FgHiRed, color.Bold)
	Rocket = color.New(color.FgHiYellow, color.Bold)
	Check  = color.New(color.FgHiGreen, color.Bold)
)

// Header prints a styled header with fire emoji
func Header(text string) {
	fmt.Printf("\n🔥 %s\n", Primary.Sprint(text))
}

// Success prints a success message with checkmark
func SuccessMsg(text string) {
	fmt.Printf("✅ %s\n", Success.Sprint(text))
}

// Warning prints a warning message with warning emoji
func WarningMsg(text string) {
	fmt.Printf("⚠️  %s\n", Warning.Sprint(text))
}

// Error prints an error message with X emoji
func ErrorMsg(text string) {
	fmt.Printf("❌ %s\n", Error.Sprint(text))
}

// Info prints an info message with info emoji
func InfoMsg(text string) {
	fmt.Printf("💡 %s\n", Info.Sprint(text))
}

// Rocket prints a message with rocket emoji for build/deploy actions
func RocketMsg(text string) {
	fmt.Printf("🚀 %s\n", Rocket.Sprint(text))
}

// Step prints a numbered step
func Step(number int, text string) {
	fmt.Printf("%s %s\n",
		Primary.Sprintf("%d.", number),
		text)
}

// Bullet prints a bullet point
func Bullet(text string) {
	fmt.Printf("  • %s\n", text)
}

// Code prints text in a code-like format
func Code(text string) string {
	return Secondary.Sprintf("`%s`", text)
}

// ProjectHeader prints a styled project header
func ProjectHeader(projectName string) {
	fmt.Printf("\n🔥 %s %s\n",
		Primary.Sprint("Firebase Project:"),
		Bold.Sprint(projectName))
}

// Platform prints platform with appropriate emoji
func Platform(platform string) string {
	var emoji string
	switch platform {
	case "iOS":
		emoji = "📱"
	case "Android":
		emoji = "🤖"
	case "macOS":
		emoji = "🖥️"
	case "Windows":
		emoji = "🪟"
	case "Linux":
		emoji = "🐧"
	case "Web":
		emoji = "🌐"
	default:
		emoji = "📦"
	}
	return fmt.Sprintf("%s %s", emoji, Secondary.Sprint(platform))
}
