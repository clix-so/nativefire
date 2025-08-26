package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectPlatform(t *testing.T) {
	tests := []struct {
		name        string
		setupFiles  []string
		setupDirs   []string
		expected    Type
		shouldError bool
	}{
		{
			name:        "Android - build.gradle",
			setupFiles:  []string{"build.gradle"},
			expected:    Android,
			shouldError: false,
		},
		{
			name:        "Android - app/build.gradle",
			setupDirs:   []string{"app"},
			setupFiles:  []string{"app/build.gradle"},
			expected:    Android,
			shouldError: false,
		},
		{
			name:        "Android - settings.gradle",
			setupFiles:  []string{"settings.gradle"},
			expected:    Android,
			shouldError: false,
		},
		{
			name:        "iOS - xcodeproj",
			setupFiles:  []string{"MyApp.xcodeproj/project.pbxproj"},
			setupDirs:   []string{"MyApp.xcodeproj"},
			expected:    iOS,
			shouldError: false,
		},
		{
			name:        "iOS - Podfile",
			setupFiles:  []string{"Podfile"},
			expected:    iOS,
			shouldError: false,
		},
		{
			name:        "macOS - macos directory",
			setupDirs:   []string{"macos"},
			expected:    MacOS,
			shouldError: false,
		},
		{
			name:        "Windows - vcxproj",
			setupFiles:  []string{"MyApp.vcxproj"},
			expected:    Windows,
			shouldError: false,
		},
		{
			name:        "Windows - CMakeLists.txt",
			setupFiles:  []string{"CMakeLists.txt"},
			expected:    Windows,
			shouldError: false,
		},
		{
			name:        "Linux - Makefile",
			setupFiles:  []string{"Makefile"},
			expected:    Linux,
			shouldError: false,
		},
		{
			name:        "No platform detected",
			setupFiles:  []string{"random.txt"},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := setupTestEnvironment(t, tt.setupDirs, tt.setupFiles)
			defer os.RemoveAll(tempDir)

			oldWd, _ := os.Getwd()
			_ = os.Chdir(tempDir)
			defer func() { _ = os.Chdir(oldWd) }()

			platform, err := DetectPlatform()

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if platform.Type() != tt.expected {
				t.Errorf("Expected platform %d, got %d", tt.expected, platform.Type())
			}
		})
	}
}

func TestFromString(t *testing.T) {
	tests := []struct {
		input       string
		expected    Type
		shouldError bool
	}{
		{"android", Android, false},
		{"ios", iOS, false},
		{"macos", MacOS, false},
		{"windows", Windows, false},
		{"linux", Linux, false},
		{"Android", Android, false},
		{"iOS", iOS, false},
		{"macOS", MacOS, false},
		{"Windows", Windows, false},
		{"Linux", Linux, false},
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			platform, err := FromString(tt.input)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for input %s", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for input %s: %v", tt.input, err)
				return
			}

			if platform.Type() != tt.expected {
				t.Errorf("Expected platform %d for input %s, got %d", tt.expected, tt.input, platform.Type())
			}
		})
	}
}

func TestAndroidPlatform(t *testing.T) {
	platform := &AndroidPlatform{}

	t.Run("Name", func(t *testing.T) {
		if platform.Name() != "Android" {
			t.Errorf("Expected name 'Android', got '%s'", platform.Name())
		}
	})

	t.Run("Type", func(t *testing.T) {
		if platform.Type() != Android {
			t.Errorf("Expected type %d, got %d", Android, platform.Type())
		}
	})

	t.Run("ConfigFileName", func(t *testing.T) {
		if platform.ConfigFileName() != "google-services.json" {
			t.Errorf("Expected config file name 'google-services.json', got '%s'", platform.ConfigFileName())
		}
	})

	t.Run("Detect", func(t *testing.T) {
		tempDir := setupTestEnvironment(t, []string{"app"}, []string{"app/build.gradle"})
		defer os.RemoveAll(tempDir)

		oldWd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(oldWd) }()

		if !platform.Detect() {
			t.Error("Expected Android platform to be detected")
		}
	})

	t.Run("ConfigPath", func(t *testing.T) {
		tempDir := setupTestEnvironment(t, []string{"app", "app/src", "app/src/main"}, []string{})
		defer os.RemoveAll(tempDir)

		oldWd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(oldWd) }()

		expected := "app"
		if platform.ConfigPath() != expected {
			t.Errorf("Expected config path '%s', got '%s'", expected, platform.ConfigPath())
		}
	})
}

func TestIOSPlatform(t *testing.T) {
	platform := &IOSPlatform{}

	t.Run("Name", func(t *testing.T) {
		if platform.Name() != "iOS" {
			t.Errorf("Expected name 'iOS', got '%s'", platform.Name())
		}
	})

	t.Run("Type", func(t *testing.T) {
		if platform.Type() != iOS {
			t.Errorf("Expected type %d, got %d", iOS, platform.Type())
		}
	})

	t.Run("ConfigFileName", func(t *testing.T) {
		if platform.ConfigFileName() != "GoogleService-Info.plist" {
			t.Errorf("Expected config file name 'GoogleService-Info.plist', got '%s'", platform.ConfigFileName())
		}
	})
}

func TestMacOSPlatform(t *testing.T) {
	platform := &MacOSPlatform{}

	t.Run("Name", func(t *testing.T) {
		if platform.Name() != "macOS" {
			t.Errorf("Expected name 'macOS', got '%s'", platform.Name())
		}
	})

	t.Run("Type", func(t *testing.T) {
		if platform.Type() != MacOS {
			t.Errorf("Expected type %d, got %d", MacOS, platform.Type())
		}
	})
}

func TestWindowsPlatform(t *testing.T) {
	platform := &WindowsPlatform{}

	t.Run("Name", func(t *testing.T) {
		if platform.Name() != "Windows" {
			t.Errorf("Expected name 'Windows', got '%s'", platform.Name())
		}
	})

	t.Run("Type", func(t *testing.T) {
		if platform.Type() != Windows {
			t.Errorf("Expected type %d, got %d", Windows, platform.Type())
		}
	})
}

func TestLinuxPlatform(t *testing.T) {
	platform := &LinuxPlatform{}

	t.Run("Name", func(t *testing.T) {
		if platform.Name() != "Linux" {
			t.Errorf("Expected name 'Linux', got '%s'", platform.Name())
		}
	})

	t.Run("Type", func(t *testing.T) {
		if platform.Type() != Linux {
			t.Errorf("Expected type %d, got %d", Linux, platform.Type())
		}
	})
}

func setupTestEnvironment(t *testing.T, dirs []string, files []string) string {
	tempDir, err := os.MkdirTemp("", "nativefire_test")
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
