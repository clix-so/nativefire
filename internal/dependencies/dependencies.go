package dependencies

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/clix-so/nativefire/internal/ui"
)

// Dependency represents an external CLI tool dependency
type Dependency struct {
	Name        string
	Command     string
	Required    bool
	Platform    string // "all", "ios", "android", "windows", "macos", "linux"
	InstallURL  string
	InstallCmd  string
	Description string
}

// Dependencies defines all external CLI dependencies
var Dependencies = []Dependency{
	{
		Name:        "Firebase CLI",
		Command:     "firebase",
		Required:    true,
		Platform:    "all",
		InstallURL:  "https://firebase.google.com/docs/cli#install_the_firebase_cli",
		InstallCmd:  "npm install -g firebase-tools",
		Description: "Required for Firebase project and app management",
	},
	{
		Name:        "CocoaPods",
		Command:     "pod",
		Required:    false,
		Platform:    "ios",
		InstallURL:  "https://cocoapods.org/",
		InstallCmd:  "sudo gem install cocoapods",
		Description: "iOS dependency manager (alternative to Swift Package Manager)",
	},
	{
		Name:        "Gradle",
		Command:     "gradle",
		Required:    false,
		Platform:    "android",
		InstallURL:  "https://gradle.org/install/",
		InstallCmd:  "Use Android Studio or install from https://gradle.org/install/",
		Description: "Android build system (gradlew wrapper preferred)",
	},
}

// CheckDependency checks if a specific CLI tool is available
func CheckDependency(command string) error {
	_, err := exec.LookPath(command)
	return err
}

// CheckAllDependencies checks all required dependencies for the current platform
func CheckAllDependencies(platform string) []Dependency {
	var missing []Dependency

	for _, dep := range Dependencies {
		// Skip if dependency is not for this platform
		if dep.Platform != "all" && dep.Platform != platform {
			continue
		}

		if err := CheckDependency(dep.Command); err != nil {
			missing = append(missing, dep)
		}
	}

	return missing
}

// CheckRequiredDependencies checks only required dependencies
func CheckRequiredDependencies(platform string) error {
	missing := CheckAllDependencies(platform)

	var requiredMissing []Dependency
	for _, dep := range missing {
		if dep.Required {
			requiredMissing = append(requiredMissing, dep)
		}
	}

	if len(requiredMissing) > 0 {
		return &MissingDependencyError{Dependencies: requiredMissing}
	}

	return nil
}

// ShowMissingDependencies displays user-friendly information about missing dependencies
func ShowMissingDependencies(missing []Dependency) {
	if len(missing) == 0 {
		return
	}

	ui.AnimatedError("Missing dependencies detected")
	fmt.Println()

	for _, dep := range missing {
		if dep.Required {
			ui.ErrorMsg(fmt.Sprintf("❌ %s (%s) - REQUIRED", dep.Name, dep.Command))
		} else {
			ui.WarningMsg(fmt.Sprintf("⚠️  %s (%s) - OPTIONAL", dep.Name, dep.Command))
		}

		ui.InfoMsg(fmt.Sprintf("   %s", dep.Description))

		// Show installation instructions
		ui.InfoMsg("   📋 Installation:")
		if dep.InstallCmd != "" {
			ui.InfoMsg(fmt.Sprintf("     %s", dep.InstallCmd))
		}
		if dep.InstallURL != "" {
			ui.InfoMsg(fmt.Sprintf("     More info: %s", dep.InstallURL))
		}
		fmt.Println()
	}
}

// MissingDependencyError represents an error when required dependencies are missing
type MissingDependencyError struct {
	Dependencies []Dependency
}

func (e *MissingDependencyError) Error() string {
	if len(e.Dependencies) == 1 {
		return fmt.Sprintf("required dependency missing: %s", e.Dependencies[0].Name)
	}
	return fmt.Sprintf("required dependencies missing: %d tools", len(e.Dependencies))
}

// GetPlatformFromOS returns the platform string based on the current OS
func GetPlatformFromOS() string {
	switch runtime.GOOS {
	case "darwin":
		return "ios" // Default to iOS for macOS (can also be "macos")
	case "linux":
		return "android" // Default to Android for Linux
	case "windows":
		return "android" // Default to Android for Windows
	default:
		return "all"
	}
}

// PreflightCheck performs a comprehensive dependency check before operations
func PreflightCheck(platform string) error {
	ui.InfoMsg("🔍 Checking dependencies...")

	missing := CheckAllDependencies(platform)

	// Show all missing dependencies (required and optional)
	if len(missing) > 0 {
		ShowMissingDependencies(missing)
	}

	// Check if any required dependencies are missing
	var requiredMissing []Dependency
	for _, dep := range missing {
		if dep.Required {
			requiredMissing = append(requiredMissing, dep)
		}
	}

	if len(requiredMissing) > 0 {
		ui.AnimatedError("Cannot proceed without required dependencies")
		return &MissingDependencyError{Dependencies: requiredMissing}
	}

	// Show optional missing dependencies as warnings
	var optionalMissing []Dependency
	for _, dep := range missing {
		if !dep.Required {
			optionalMissing = append(optionalMissing, dep)
		}
	}

	if len(optionalMissing) > 0 {
		ui.WarningMsg(fmt.Sprintf("Some optional dependencies are missing (%d), but you can continue", len(optionalMissing)))
	} else {
		ui.AnimatedSuccess("All dependencies are available")
	}

	return nil
}
