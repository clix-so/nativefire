package platform

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/clix-so/nativefire/internal/firebase"
)

// Constants for repeated strings
const (
	googleServiceInfoPlistDesktop = "GoogleService-Info.plist"
	googleServicesJSONDesktop     = "google-services.json"
)

func (p *MacOSPlatform) Name() string {
	return "macOS"
}

func (p *MacOSPlatform) Type() Type {
	return MacOS
}

func (p *MacOSPlatform) Detect() bool {
	return fileExists("macos") ||
		(findFile(".", "*.xcodeproj") != "" && fileExists("Podfile")) ||
		findFile(".", "main.swift") != ""
}

func (p *MacOSPlatform) ConfigFileName() string {
	return googleServiceInfoPlistDesktop
}

func (p *MacOSPlatform) ConfigPath() string {
	if fileExists("macos") {
		return "macos"
	}
	return "."
}

func (p *MacOSPlatform) InstallConfig(config *firebase.Config) error {
	return p.installConfigHelper()
}

func (p *MacOSPlatform) AddInitializationCode(config *firebase.Config) error {
	return p.addInitializationHelper()
}

func (p *WindowsPlatform) Name() string {
	return "Windows"
}

func (p *WindowsPlatform) Type() Type {
	return Windows
}

func (p *WindowsPlatform) Detect() bool {
	return fileExists("windows") ||
		findFile(".", "*.vcxproj") != "" ||
		findFile(".", "*.sln") != "" ||
		fileExists("CMakeLists.txt")
}

func (p *WindowsPlatform) ConfigFileName() string {
	return googleServicesJSONDesktop
}

func (p *WindowsPlatform) ConfigPath() string {
	if fileExists("windows") {
		return "windows"
	}
	return "."
}

func (p *WindowsPlatform) InstallConfig(config *firebase.Config) error {
	return p.installConfigHelper()
}

func (p *WindowsPlatform) AddInitializationCode(config *firebase.Config) error {
	return p.addInitializationHelper()
}

func (p *LinuxPlatform) Name() string {
	return "Linux"
}

func (p *LinuxPlatform) Type() Type {
	return Linux
}

func (p *LinuxPlatform) Detect() bool {
	return fileExists("linux") ||
		fileExists("CMakeLists.txt") ||
		findFile(".", "Makefile") != ""
}

func (p *LinuxPlatform) ConfigFileName() string {
	return googleServicesJSONDesktop
}

func (p *LinuxPlatform) ConfigPath() string {
	if fileExists("linux") {
		return "linux"
	}
	return "."
}

func (p *LinuxPlatform) InstallConfig(config *firebase.Config) error {
	return p.installConfigHelper()
}

func (p *LinuxPlatform) AddInitializationCode(config *firebase.Config) error {
	return p.addInitializationHelper()
}

func (p *MacOSPlatform) installConfigHelper() error {
	configPath := p.ConfigPath()
	targetPath := filepath.Join(configPath, p.ConfigFileName())

	if err := os.MkdirAll(configPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", configPath, err)
	}

	sourceFile := filepath.Join(os.TempDir(), p.ConfigFileName())

	sourceData, err := os.ReadFile(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to read source config file: %w", err)
	}

	if err := os.WriteFile(targetPath, sourceData, 0644); err != nil {
		return fmt.Errorf("failed to write config file to %s: %w", targetPath, err)
	}

	fmt.Printf("✅ Configuration file installed at: %s\n", targetPath)
	return nil
}

func (p *MacOSPlatform) addInitializationHelper() error {
	fmt.Printf("⚠️  Please manually add Firebase initialization code to your %s application.\n", p.Name())
	fmt.Println("💡 Refer to Firebase documentation for platform-specific initialization steps.")
	return nil
}

func (p *WindowsPlatform) installConfigHelper() error {
	configPath := p.ConfigPath()
	targetPath := filepath.Join(configPath, p.ConfigFileName())

	if err := os.MkdirAll(configPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", configPath, err)
	}

	sourceFile := filepath.Join(os.TempDir(), p.ConfigFileName())

	sourceData, err := os.ReadFile(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to read source config file: %w", err)
	}

	if err := os.WriteFile(targetPath, sourceData, 0644); err != nil {
		return fmt.Errorf("failed to write config file to %s: %w", targetPath, err)
	}

	fmt.Printf("✅ Configuration file installed at: %s\n", targetPath)
	return nil
}

func (p *WindowsPlatform) addInitializationHelper() error {
	fmt.Printf("⚠️  Please manually add Firebase initialization code to your %s application.\n", p.Name())
	fmt.Println("💡 Refer to Firebase documentation for platform-specific initialization steps.")
	return nil
}

func (p *LinuxPlatform) installConfigHelper() error {
	configPath := p.ConfigPath()
	targetPath := filepath.Join(configPath, p.ConfigFileName())

	if err := os.MkdirAll(configPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", configPath, err)
	}

	sourceFile := filepath.Join(os.TempDir(), p.ConfigFileName())

	sourceData, err := os.ReadFile(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to read source config file: %w", err)
	}

	if err := os.WriteFile(targetPath, sourceData, 0644); err != nil {
		return fmt.Errorf("failed to write config file to %s: %w", targetPath, err)
	}

	fmt.Printf("✅ Configuration file installed at: %s\n", targetPath)
	return nil
}

func (p *LinuxPlatform) addInitializationHelper() error {
	fmt.Printf("⚠️  Please manually add Firebase initialization code to your %s application.\n", p.Name())
	fmt.Println("💡 Refer to Firebase documentation for platform-specific initialization steps.")
	return nil
}
