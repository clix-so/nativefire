package platform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/clix-so/nativefire/internal/firebase"
)

func TestAndroidPlatformInstallConfig(t *testing.T) {
	tests := []struct {
		name      string
		setupDirs []string
		expected  string
	}{
		{
			name:      "Standard Android project structure",
			setupDirs: []string{"app", "app/src", "app/src/main"},
			expected:  "app",
		},
		{
			name:      "Flutter Android project structure",
			setupDirs: []string{"android", "android/app", "android/app/src", "android/app/src/main"},
			expected:  "android/app",
		},
		{
			name:      "Default structure",
			setupDirs: []string{},
			expected:  "app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := setupTestEnvironment(t, tt.setupDirs, []string{})
			defer os.RemoveAll(tempDir)

			oldWd, _ := os.Getwd()
			_ = os.Chdir(tempDir)
			defer func() { _ = os.Chdir(oldWd) }()

			// Create a mock config file in temp directory
			configFile := filepath.Join(os.TempDir(), "google-services.json")
			configContent := `{"project_info":{"project_id":"test-project"}}`
			if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
				t.Fatalf("Failed to create mock config file: %v", err)
			}
			defer os.Remove(configFile)

			platform := &AndroidPlatform{}
			config := &firebase.Config{
				ProjectID: "test-project",
				AppID:     "test-app-id",
				Platform:  platform,
			}

			err := platform.InstallConfig(config)
			if err != nil {
				t.Errorf("InstallConfig failed: %v", err)
				return
			}

			// Check if config file was installed in the correct location
			expectedPath := filepath.Join(tt.expected, "google-services.json")
			if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
				t.Errorf("Config file not found at expected path: %s", expectedPath)
			}

			// Verify content
			content, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Errorf("Failed to read installed config file: %v", err)
				return
			}

			if string(content) != configContent {
				t.Errorf("Config file content mismatch. Expected: %s, Got: %s", configContent, string(content))
			}
		})
	}
}

func TestAndroidPlatformAddInitializationCode(t *testing.T) {
	tests := []struct {
		name                  string
		setupFiles            []string
		buildGradle           string
		expectedInBuildGradle string
	}{
		{
			name:       "Add google-services plugin to plugins block",
			setupFiles: []string{"app/build.gradle"},
			buildGradle: `plugins {
    id 'com.android.application'
}

dependencies {
    implementation 'androidx.core:core-ktx:1.7.0'
}`,
			expectedInBuildGradle: "id 'com.google.gms.google-services'",
		},
		{
			name:       "Add google-services plugin with apply plugin syntax",
			setupFiles: []string{"app/build.gradle"},
			buildGradle: `apply plugin: 'com.android.application'

dependencies {
    implementation 'androidx.core:core-ktx:1.7.0'
}`,
			expectedInBuildGradle: "apply plugin: 'com.google.gms.google-services'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := setupTestEnvironment(t, []string{"app"}, []string{})
			defer os.RemoveAll(tempDir)

			oldWd, _ := os.Getwd()
			_ = os.Chdir(tempDir)
			defer func() { _ = os.Chdir(oldWd) }()

			// Create build.gradle with specific content
			buildGradlePath := filepath.Join(tempDir, "app/build.gradle")
			if err := os.WriteFile(buildGradlePath, []byte(tt.buildGradle), 0644); err != nil {
				t.Fatalf("Failed to create build.gradle: %v", err)
			}

			platform := &AndroidPlatform{}
			config := &firebase.Config{
				ProjectID: "test-project",
				AppID:     "test-app-id",
				Platform:  platform,
			}

			err := platform.AddInitializationCode(config)
			if err != nil {
				t.Errorf("AddInitializationCode failed: %v", err)
				return
			}

			// Verify that google-services plugin was added
			content, err := os.ReadFile(buildGradlePath)
			if err != nil {
				t.Errorf("Failed to read build.gradle: %v", err)
				return
			}

			contentStr := string(content)
			if !contains(contentStr, tt.expectedInBuildGradle) {
				t.Errorf("Expected build.gradle to contain '%s', but it doesn't. Content: %s", tt.expectedInBuildGradle, contentStr)
			}
		})
	}
}

func TestAndroidPlatformFindBuildGradle(t *testing.T) {
	tests := []struct {
		name       string
		setupFiles []string
		expected   string
	}{
		{
			name:       "Standard Android project",
			setupFiles: []string{"app/build.gradle"},
			expected:   "app/build.gradle",
		},
		{
			name:       "Flutter Android project",
			setupFiles: []string{"android/app/build.gradle"},
			expected:   "android/app/build.gradle",
		},
		{
			name:       "Root build.gradle",
			setupFiles: []string{"build.gradle"},
			expected:   "build.gradle",
		},
		{
			name:       "No build.gradle",
			setupFiles: []string{},
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := setupTestEnvironment(t, []string{"app", "android", "android/app"}, tt.setupFiles)
			defer os.RemoveAll(tempDir)

			oldWd, _ := os.Getwd()
			_ = os.Chdir(tempDir)
			defer func() { _ = os.Chdir(oldWd) }()

			platform := &AndroidPlatform{}
			result := platform.findBuildGradle()

			if result != tt.expected {
				t.Errorf("Expected findBuildGradle to return '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
