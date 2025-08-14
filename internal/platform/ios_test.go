package platform

import (
	"os"
	"strings"
	"testing"
)

func TestIOSPlatformBasics(t *testing.T) {
	platform := &IOSPlatform{}

	// Test basic methods
	if platform.Name() != "iOS" {
		t.Errorf("Expected Name() to return 'iOS', got '%s'", platform.Name())
	}

	if platform.Type() != iOS {
		t.Errorf("Expected Type() to return iOS, got %v", platform.Type())
	}

	if platform.ConfigFileName() != googleServiceInfoPlist {
		t.Errorf("Expected ConfigFileName() to return '%s', got '%s'", googleServiceInfoPlist, platform.ConfigFileName())
	}
}

func TestIOSPlatformDetect(t *testing.T) {
	platform := &IOSPlatform{}

	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "ios_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tmpDir)

	// Test detection without iOS indicators
	if platform.Detect() {
		t.Error("Expected Detect() to return false with no iOS indicators")
	}

	// Test detection with ios directory
	_ = os.Mkdir("ios", 0755)
	if !platform.Detect() {
		t.Error("Expected Detect() to return true with ios directory")
	}
	os.RemoveAll("ios")

	// Test detection with Podfile
	f, _ := os.Create("Podfile")
	f.Close()
	if !platform.Detect() {
		t.Error("Expected Detect() to return true with Podfile")
	}
}

func TestIOSPlatformConfigPath(t *testing.T) {
	platform := &IOSPlatform{}

	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "ios_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tmpDir)

	// Test default config path
	if platform.ConfigPath() != "." {
		t.Errorf("Expected ConfigPath() to return '.', got '%s'", platform.ConfigPath())
	}

	// Test with ios directory
	_ = os.Mkdir("ios", 0755)
	if platform.ConfigPath() != "ios" {
		t.Errorf("Expected ConfigPath() to return 'ios', got '%s'", platform.ConfigPath())
	}
}

func TestIOSPlatformFindProjectName(t *testing.T) {
	platform := &IOSPlatform{}

	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "ios_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tmpDir)

	// Test with no xcodeproj
	if projectName := platform.findProjectName(); projectName != "" {
		t.Errorf("Expected empty project name, got '%s'", projectName)
	}

	// Test with xcodeproj
	_ = os.Mkdir("TestProject.xcodeproj", 0755)
	if projectName := platform.findProjectName(); projectName != "TestProject" {
		t.Errorf("Expected project name 'TestProject', got '%s'", projectName)
	}
}

func TestIOSPlatformFindPodfile(t *testing.T) {
	platform := &IOSPlatform{}

	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "ios_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tmpDir)

	// Test with no Podfile
	if podfile := platform.findPodfile(); podfile != "" {
		t.Errorf("Expected empty Podfile path, got '%s'", podfile)
	}

	// Test with root Podfile
	f, _ := os.Create("Podfile")
	f.Close()
	if podfile := platform.findPodfile(); podfile != "Podfile" {
		t.Errorf("Expected 'Podfile', got '%s'", podfile)
	}
	os.Remove("Podfile")

	// Test with ios/Podfile
	_ = os.Mkdir("ios", 0755)
	f, _ = os.Create("ios/Podfile")
	f.Close()
	if podfile := platform.findPodfile(); podfile != "ios/Podfile" {
		t.Errorf("Expected 'ios/Podfile', got '%s'", podfile)
	}
}

func TestIOSPlatformFindAppDelegate(t *testing.T) {
	platform := &IOSPlatform{}

	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "ios_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tmpDir)

	// Test with no AppDelegate
	if appDelegate := platform.findAppDelegate(); appDelegate != "" {
		t.Errorf("Expected empty AppDelegate path, got '%s'", appDelegate)
	}

	// Test with Swift AppDelegate
	_ = os.Mkdir("ios", 0755)
	f, _ := os.Create("ios/AppDelegate.swift")
	f.Close()
	if appDelegate := platform.findAppDelegate(); !strings.Contains(appDelegate, "AppDelegate.swift") {
		t.Errorf("Expected path containing 'AppDelegate.swift', got '%s'", appDelegate)
	}
	os.Remove("ios/AppDelegate.swift")

	// Test with Objective-C AppDelegate
	f, _ = os.Create("ios/AppDelegate.m")
	f.Close()
	if appDelegate := platform.findAppDelegate(); !strings.Contains(appDelegate, "AppDelegate.m") {
		t.Errorf("Expected path containing 'AppDelegate.m', got '%s'", appDelegate)
	}
}

func TestIOSPlatformIsSwiftProject(t *testing.T) {
	platform := &IOSPlatform{}

	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "ios_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tmpDir)

	// Test default (should be Swift)
	if !platform.isSwiftProject() {
		t.Error("Expected isSwiftProject() to return true by default")
	}

	// Test with Swift file
	f, _ := os.Create("ContentView.swift")
	f.Close()
	if !platform.isSwiftProject() {
		t.Error("Expected isSwiftProject() to return true with Swift file present")
	}
	os.Remove("ContentView.swift")

	// Test with Objective-C files
	f, _ = os.Create("ViewController.m")
	f.Close()
	f, _ = os.Create("ViewController.h")
	f.Close()
	if platform.isSwiftProject() {
		t.Error("Expected isSwiftProject() to return false with Objective-C files present")
	}
}

func TestIOSPlatformDetermineAppDelegatePath(t *testing.T) {
	platform := &IOSPlatform{}

	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "ios_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tmpDir)

	// Test default path
	if path := platform.determineAppDelegatePath(); path != "." {
		t.Errorf("Expected '.', got '%s'", path)
	}

	// Test with ios directory
	_ = os.Mkdir("ios", 0755)
	if path := platform.determineAppDelegatePath(); path != "ios" {
		t.Errorf("Expected 'ios', got '%s'", path)
	}

	// Test with project-specific directory
	_ = os.Mkdir("TestProject.xcodeproj", 0755)
	_ = os.Mkdir("ios/TestProject", 0755)
	if path := platform.determineAppDelegatePath(); path != "ios/TestProject" {
		t.Errorf("Expected 'ios/TestProject', got '%s'", path)
	}
}

func TestIOSPlatformHasSwiftPackages(t *testing.T) {
	platform := &IOSPlatform{}

	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "ios_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tmpDir)

	// Test with no Swift packages
	if platform.hasSwiftPackages() {
		t.Error("Expected hasSwiftPackages() to return false with no packages")
	}

	// Test with Package.swift
	f, _ := os.Create("Package.swift")
	f.Close()
	if !platform.hasSwiftPackages() {
		t.Error("Expected hasSwiftPackages() to return true with Package.swift")
	}
	os.Remove("Package.swift")

	// Test with .swiftpm directory
	_ = os.Mkdir(".swiftpm", 0755)
	if !platform.hasSwiftPackages() {
		t.Error("Expected hasSwiftPackages() to return true with .swiftpm directory")
	}
}
