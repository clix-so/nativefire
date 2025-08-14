package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/clix-so/nativefire/internal/firebase"
	"github.com/clix-so/nativefire/internal/ui"
)

// Constants for repeated strings
const (
	googleServicesJSON = "google-services.json"
	appDir             = "app"
)

func (p *AndroidPlatform) Name() string {
	return "Android"
}

func (p *AndroidPlatform) Type() Type {
	return Android
}

func (p *AndroidPlatform) Detect() bool {
	return fileExists("build.gradle") ||
		fileExists("app/build.gradle") ||
		fileExists("android/build.gradle") ||
		fileExists("settings.gradle")
}

func (p *AndroidPlatform) ConfigFileName() string {
	return googleServicesJSON
}

func (p *AndroidPlatform) ConfigPath() string {
	if fileExists("app/src/main") {
		return appDir
	}
	if fileExists("android/app/src/main") {
		return "android/app"
	}
	return appDir
}

func (p *AndroidPlatform) InstallConfig(config *firebase.Config) error {
	configPath := p.ConfigPath()
	targetPath := filepath.Join(configPath, p.ConfigFileName())

	if err := os.MkdirAll(configPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", configPath, err)
	}

	// Use the unique temp file path from config instead of hardcoded temp location
	sourceFile := config.ConfigFile
	if sourceFile == "" {
		// Fallback to old behavior if ConfigFile is not set
		sourceFile = filepath.Join(os.TempDir(), p.ConfigFileName())
	}

	sourceData, err := os.ReadFile(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to read source config file: %w", err)
	}

	if err := os.WriteFile(targetPath, sourceData, 0644); err != nil {
		return fmt.Errorf("failed to write config file to %s: %w", targetPath, err)
	}

	// Clean up the temp file after successful installation
	if config.ConfigFile != "" {
		os.Remove(config.ConfigFile)
	}

	ui.SuccessMsg(fmt.Sprintf("Configuration file installed at: %s", targetPath))
	return nil
}

func (p *AndroidPlatform) AddInitializationCode(config *firebase.Config) error {
	buildGradlePath := p.findBuildGradle()
	if buildGradlePath == "" {
		return fmt.Errorf("could not find app-level build.gradle file")
	}

	content, err := os.ReadFile(buildGradlePath)
	if err != nil {
		return fmt.Errorf("failed to read build.gradle: %w", err)
	}

	contentStr := string(content)
	gradleModified := false

	if !strings.Contains(contentStr, "google-services") {
		if strings.Contains(contentStr, "plugins {") {
			contentStr = strings.Replace(contentStr,
				"plugins {",
				"plugins {\n    id 'com.google.gms.google-services'", 1)
		} else {
			contentStr = "apply plugin: 'com.google.gms.google-services'\n\n" + contentStr
		}

		if err := os.WriteFile(buildGradlePath, []byte(contentStr), 0644); err != nil {
			return fmt.Errorf("failed to update build.gradle: %w", err)
		}
		ui.SuccessMsg(fmt.Sprintf("Added Google Services plugin to: %s", buildGradlePath))
		gradleModified = true
	}

	projectBuildGradlePath := p.findProjectBuildGradle()
	if projectBuildGradlePath != "" {
		if err := p.addClasspathToBuildGradle(projectBuildGradlePath); err != nil {
			return err
		}
		gradleModified = true
	}

	if err := p.addFirebaseImportsToMainActivity(); err != nil {
		return err
	}

	// Run Gradle sync after modifications
	if gradleModified {
		return p.runGradleSync()
	}

	return nil
}

func (p *AndroidPlatform) findBuildGradle() string {
	candidates := []string{
		"app/build.gradle",
		"android/app/build.gradle",
		"build.gradle",
	}

	for _, candidate := range candidates {
		if fileExists(candidate) {
			return candidate
		}
	}
	return ""
}

func (p *AndroidPlatform) findProjectBuildGradle() string {
	candidates := []string{
		"build.gradle",
		"android/build.gradle",
	}

	for _, candidate := range candidates {
		if fileExists(candidate) {
			content, err := os.ReadFile(candidate)
			if err == nil && strings.Contains(string(content), "buildscript") {
				return candidate
			}
		}
	}
	return ""
}

func (p *AndroidPlatform) addClasspathToBuildGradle(buildGradlePath string) error {
	content, err := os.ReadFile(buildGradlePath)
	if err != nil {
		return fmt.Errorf("failed to read project build.gradle: %w", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "google-services") {
		if strings.Contains(contentStr, "dependencies {") {
			insertPoint := strings.Index(contentStr, "dependencies {") + len("dependencies {")
			newContent := contentStr[:insertPoint] +
				"\n        classpath 'com.google.gms:google-services:4.3.15'" +
				contentStr[insertPoint:]

			if err := os.WriteFile(buildGradlePath, []byte(newContent), 0644); err != nil {
				return fmt.Errorf("failed to update project build.gradle: %w", err)
			}
			ui.SuccessMsg(fmt.Sprintf("Added Google Services classpath to: %s", buildGradlePath))
		}
	}
	return nil
}

func (p *AndroidPlatform) addFirebaseImportsToMainActivity() error {
	mainActivityPath := findFile(".", "MainActivity.java")
	if mainActivityPath == "" {
		mainActivityPath = findFile(".", "MainActivity.kt")
	}

	if mainActivityPath == "" {
		ui.WarningMsg("MainActivity not found. Please manually add Firebase initialization code")
		return nil
	}

	content, err := os.ReadFile(mainActivityPath)
	if err != nil {
		return fmt.Errorf("failed to read MainActivity: %w", err)
	}

	contentStr := string(content)

	if strings.Contains(mainActivityPath, ".java") {
		if !strings.Contains(contentStr, "FirebaseApp.initializeApp") {
			// Add Firebase import after existing imports (safer approach)
			if !strings.Contains(contentStr, "import com.google.firebase.FirebaseApp;") {
				if strings.Contains(contentStr, "import android.os.Bundle;") {
					contentStr = strings.Replace(contentStr,
						"import android.os.Bundle;",
						"import android.os.Bundle;\nimport com.google.firebase.FirebaseApp;", 1)
				} else if strings.Contains(contentStr, "import androidx.appcompat.app.AppCompatActivity;") {
					contentStr = strings.Replace(contentStr,
						"import androidx.appcompat.app.AppCompatActivity;",
						"import androidx.appcompat.app.AppCompatActivity;\nimport com.google.firebase.FirebaseApp;", 1)
				} else {
					// Fallback: add after package declaration
					contentStr = strings.Replace(contentStr,
						"package",
						"import com.google.firebase.FirebaseApp;\n\npackage", 1)
				}
			}

			if strings.Contains(contentStr, "onCreate") {
				contentStr = strings.Replace(contentStr,
					"super.onCreate(savedInstanceState);",
					"super.onCreate(savedInstanceState);\n        FirebaseApp.initializeApp(this);", 1)
			}
		}
	} else if strings.Contains(mainActivityPath, ".kt") {
		if !strings.Contains(contentStr, "FirebaseApp.initializeApp") {
			// Add Firebase import after existing imports
			if !strings.Contains(contentStr, "import com.google.firebase.FirebaseApp") {
				if strings.Contains(contentStr, "import android.os.Bundle") {
					contentStr = strings.Replace(contentStr,
						"import android.os.Bundle",
						"import android.os.Bundle\nimport com.google.firebase.FirebaseApp", 1)
				} else if strings.Contains(contentStr, "import androidx.appcompat.app.AppCompatActivity") {
					contentStr = strings.Replace(contentStr,
						"import androidx.appcompat.app.AppCompatActivity",
						"import androidx.appcompat.app.AppCompatActivity\nimport com.google.firebase.FirebaseApp", 1)
				} else {
					// Fallback: add after package declaration
					contentStr = strings.Replace(contentStr,
						"package",
						"import com.google.firebase.FirebaseApp\n\npackage", 1)
				}
			}

			if strings.Contains(contentStr, "onCreate") {
				contentStr = strings.Replace(contentStr,
					"super.onCreate(savedInstanceState)",
					"super.onCreate(savedInstanceState)\n        FirebaseApp.initializeApp(this)", 1)
			}
		}
	}

	if err := os.WriteFile(mainActivityPath, []byte(contentStr), 0644); err != nil {
		return fmt.Errorf("failed to update MainActivity: %w", err)
	}

	ui.SuccessMsg(fmt.Sprintf("Added Firebase initialization to: %s", mainActivityPath))
	return nil
}

func (p *AndroidPlatform) runGradleSync() error {
	ui.InfoMsg("Running Gradle sync...")

	// Check for Gradle wrapper first
	if fileExists("gradlew") {
		return p.runGradlew()
	}

	// Check for system Gradle
	if p.hasSystemGradle() {
		return p.runSystemGradle()
	}

	ui.WarningMsg("Gradle not found. Please sync your project manually")
	ui.InfoMsg("In Android Studio: File > Sync Project with Gradle Files")
	return nil
}

func (p *AndroidPlatform) runGradlew() error {
	ui.InfoMsg("Using Gradle Wrapper...")

	// Run gradlew sync or build to trigger dependency resolution
	if err := p.runCommand("./gradlew", []string{"--refresh-dependencies"}, "Syncing Gradle dependencies"); err != nil {
		ui.WarningMsg("Gradle sync failed. Please sync manually in Android Studio")
		ui.InfoMsg("In Android Studio: File > Sync Project with Gradle Files")
		return nil
	}

	ui.SuccessMsg("Gradle dependencies synced successfully!")
	return nil
}

func (p *AndroidPlatform) hasSystemGradle() bool {
	cmd := exec.Command("which", "gradle")
	return cmd.Run() == nil
}

func (p *AndroidPlatform) runSystemGradle() error {
	ui.InfoMsg("Using system Gradle...")

	if err := p.runCommand("gradle", []string{"--refresh-dependencies"}, "Syncing Gradle dependencies"); err != nil {
		ui.WarningMsg("Gradle sync failed. Please sync manually in Android Studio")
		ui.InfoMsg("In Android Studio: File > Sync Project with Gradle Files")
		return nil
	}

	ui.SuccessMsg("Gradle dependencies synced successfully!")
	return nil
}

func (p *AndroidPlatform) runCommand(command string, args []string, description string) error {
	ui.InfoMsg(fmt.Sprintf("Running: %s %s", command, strings.Join(args, " ")))

	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		ui.WarningMsg(fmt.Sprintf("Command failed: %s", string(output)))
		return err
	}

	if len(output) > 0 {
		ui.InfoMsg(string(output))
	}

	return nil
}
