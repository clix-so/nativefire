package firebase

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// MockPlatform implements PlatformInterface for testing
type MockPlatform struct {
	name           string
	configFileName string
	configPath     string
}

func (m *MockPlatform) Name() string           { return m.name }
func (m *MockPlatform) ConfigFileName() string { return m.configFileName }
func (m *MockPlatform) ConfigPath() string     { return m.configPath }

func TestNewClient(t *testing.T) {
	client := NewClient(true)
	if client == nil {
		t.Error("NewClient returned nil")
		return
	}
	if !client.verbose {
		t.Error("Expected verbose to be true")
	}

	client = NewClient(false)
	if client.verbose {
		t.Error("Expected verbose to be false")
	}
}

func TestCheckFirebaseCLI(t *testing.T) {
	client := NewClient(false)

	// Test when firebase CLI is available (assuming it's installed in the test environment)
	err := client.checkFirebaseCLI()
	if err != nil {
		// If firebase CLI is not installed, skip this test
		t.Skipf("Firebase CLI not installed: %v", err)
	}
}

func TestCheckFirebaseCLIMissing(t *testing.T) {
	client := NewClient(false)

	// Temporarily modify PATH to simulate missing firebase CLI
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", originalPath)

	err := client.checkFirebaseCLI()
	if err == nil {
		t.Error("Expected error when Firebase CLI is missing")
	}

	expectedError := "firebase CLI not found"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedError, err)
	}
}

func TestGenerateAppName(t *testing.T) {
	client := NewClient(false)

	tests := []struct {
		name         string
		platformName string
		workingDir   string
		expected     string
	}{
		{
			name:         "Normal project directory",
			platformName: "Android",
			workingDir:   "/tmp/my-awesome-project",
			expected:     "my-awesome-project Android",
		},
		{
			name:         "iOS platform",
			platformName: "iOS",
			workingDir:   "/Users/dev/ios-app",
			expected:     "ios-app iOS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir, err := os.MkdirTemp("", "nativefire_test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Create subdirectory to simulate project name
			projectDir := filepath.Join(tempDir, filepath.Base(tt.workingDir))
			if err := os.MkdirAll(projectDir, 0755); err != nil {
				t.Fatalf("Failed to create project dir: %v", err)
			}

			oldWd, _ := os.Getwd()
			_ = os.Chdir(projectDir)
			defer func() { _ = os.Chdir(oldWd) }()

			result := client.generateAppName(tt.platformName)
			if result != tt.expected {
				t.Errorf("Expected app name '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetPlatformFlag(t *testing.T) {
	client := NewClient(false)

	tests := []struct {
		platformName string
		expected     string
	}{
		{"Android", "android"},
		{"android", "android"},
		{"iOS", "ios"},
		{"ios", "ios"},
		{"macOS", "ios"},
		{"macos", "ios"},
		{"web", "web"},
		{"unknown", "android"}, // default fallback
	}

	for _, tt := range tests {
		t.Run(tt.platformName, func(t *testing.T) {
			result := client.getPlatformFlag(tt.platformName)
			if result != tt.expected {
				t.Errorf("Expected platform flag '%s' for platform '%s', got '%s'", tt.expected, tt.platformName, result)
			}
		})
	}
}

func TestExtractAppIDFromOutput(t *testing.T) {
	client := NewClient(false)

	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{
			name: "App ID format 1",
			output: `✔ Firebase project some-project is ready! Here's your app info:
App ID: 1:123456789:android:abcdef123456
Display Name: My App (Android)`,
			expected: "1:123456789:android:abcdef123456",
		},
		{
			name: "App ID format 2",
			output: `{
  "appId": "1:987654321:ios:fedcba654321",
  "displayName": "My iOS App"
}`,
			expected: "1:987654321:ios:fedcba654321",
		},
		{
			name:     "No App ID found",
			output:   "Some random output without app ID",
			expected: "",
		},
		{
			name: "App ID with quotes and comma",
			output: `{
  "appId": "1:111222333:android:xyz789",
  "platform": "android"
}`,
			expected: "1:111222333:android:xyz789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.extractAppIDFromOutput(tt.output)
			if result != tt.expected {
				t.Errorf("Expected app ID '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestRegisterApp(t *testing.T) {
	client := NewClient(true)

	// Test with existing app ID
	t.Run("Existing App ID", func(t *testing.T) {
		mockPlatform := &MockPlatform{
			name:           "Android",
			configFileName: "google-services.json",
			configPath:     "/tmp",
		}

		config := &Config{
			ProjectID: "test-project",
			AppID:     "existing-app-id",
			Platform:  mockPlatform,
		}

		err := client.RegisterApp(config)
		if err != nil {
			// If Firebase CLI is not available, skip
			if strings.Contains(err.Error(), "firebase CLI not found") {
				t.Skip("Firebase CLI not available")
			}
			if strings.Contains(err.Error(), "not authenticated") {
				t.Skip("Not authenticated with Firebase")
			}
			if strings.Contains(err.Error(), "exit status 1") {
				t.Skip("Firebase CLI authentication or permission issue")
			}
			if strings.Contains(err.Error(), "failed to check authentication") {
				t.Skip("Firebase authentication check failed")
			}
			t.Errorf("RegisterApp failed: %v", err)
		}

		if config.AppID != "existing-app-id" {
			t.Errorf("Expected app ID to remain 'existing-app-id', got '%s'", config.AppID)
		}
	})
}

func TestDownloadConfig(t *testing.T) {
	client := NewClient(false)

	t.Run("Missing App ID", func(t *testing.T) {
		mockPlatform := &MockPlatform{
			name:           "Android",
			configFileName: "google-services.json",
			configPath:     "/tmp",
		}

		config := &Config{
			ProjectID: "test-project",
			AppID:     "", // Missing app ID
			Platform:  mockPlatform,
		}

		err := client.DownloadConfig(config)
		if err == nil {
			t.Error("Expected error when app ID is missing")
		}

		expectedError := "app ID is required"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("Expected error to contain '%s', got: %v", expectedError, err)
		}
	})

	t.Run("Unsupported Platform", func(t *testing.T) {
		mockPlatform := &MockPlatform{
			name:           "UnsupportedPlatform",
			configFileName: "config.json",
			configPath:     "/tmp",
		}

		config := &Config{
			ProjectID: "test-project",
			AppID:     "test-app-id",
			Platform:  mockPlatform,
		}

		err := client.DownloadConfig(config)
		if err == nil {
			t.Error("Expected error for unsupported platform")
		}

		expectedError := "does not support automatic config download"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("Expected error to contain '%s', got: %v", expectedError, err)
		}
	})
}

// TestIntegrationFirebaseCLI tests Firebase CLI integration
// This test requires Firebase CLI to be installed and authenticated
func TestIntegrationFirebaseCLI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if firebase CLI is available
	if _, err := exec.LookPath("firebase"); err != nil {
		t.Skip("Firebase CLI not available for integration test")
	}

	client := NewClient(true)

	// Test authentication check
	t.Run("Authentication Check", func(t *testing.T) {
		err := client.checkAuthentication()
		if err != nil {
			if strings.Contains(err.Error(), "not authenticated") {
				t.Skip("Not authenticated with Firebase - run 'firebase login' to test authentication")
			}
			t.Errorf("Authentication check failed: %v", err)
		}
	})
}

// Benchmark tests
func BenchmarkGenerateAppName(b *testing.B) {
	client := NewClient(false)
	for i := 0; i < b.N; i++ {
		client.generateAppName("Android")
	}
}

func BenchmarkGetPlatformFlag(b *testing.B) {
	client := NewClient(false)
	for i := 0; i < b.N; i++ {
		client.getPlatformFlag("Android")
	}
}

// Test utility functions that don't require Firebase CLI
func TestFormatCommand(t *testing.T) {
	client := NewClient(false)

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "Simple command",
			args:     []string{"firebase", "projects:list"},
			expected: "firebase projects:list",
		},
		{
			name:     "Command with spaces",
			args:     []string{"firebase", "apps:create", "My App Name"},
			expected: "firebase apps:create \"My App Name\"",
		},
		{
			name:     "Command with special characters",
			args:     []string{"firebase", "apps:create", "app(test)"},
			expected: "firebase apps:create \"app(test)\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.formatCommand(tt.args)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGenerateDefaultPackageName(t *testing.T) {
	client := NewClient(false)

	tests := []struct {
		name      string
		projectID string
		expected  string
	}{
		{
			name:      "Simple project ID",
			projectID: "myproject",
			expected:  "com.firebase.myproject",
		},
		{
			name:      "Project ID with hyphens",
			projectID: "my-awesome-project",
			expected:  "com.my.awesome.project",
		},
		{
			name:      "Project ID with dots",
			projectID: "com.example.project",
			expected:  "com.com.example.project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.generateDefaultPackageName(tt.projectID)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGenerateDefaultBundleID(t *testing.T) {
	client := NewClient(false)

	tests := []struct {
		name      string
		projectID string
		expected  string
	}{
		{
			name:      "Simple project ID",
			projectID: "myproject",
			expected:  "com.firebase.myproject",
		},
		{
			name:      "Project ID with hyphens",
			projectID: "my-awesome-project",
			expected:  "com.my.awesome.project",
		},
		{
			name:      "Project ID with dots",
			projectID: "com.example.project",
			expected:  "com.com.example.project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.generateDefaultBundleID(tt.projectID)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestIsDuplicateAppError(t *testing.T) {
	client := NewClient(false)

	tests := []struct {
		name     string
		output   string
		expected bool
	}{
		{
			name:     "Duplicate app error",
			output:   "Error: App with bundle ID com.example.app already exists",
			expected: true,
		},
		{
			name:     "Package name exists error",
			output:   "Error: package name already exists for this project",
			expected: true,
		},
		{
			name:     "Bundle ID exists error",
			output:   "Error: bundle id already exists in this project",
			expected: true,
		},
		{
			name:     "Failed to create iOS app",
			output:   "Failed to create iOS app: duplicate bundle identifier",
			expected: true,
		},
		{
			name:     "No duplicate error",
			output:   "Successfully created Firebase app",
			expected: false,
		},
		{
			name:     "Different error",
			output:   "Error: Invalid project ID",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.isDuplicateAppError(tt.output)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFilterAppsByPlatform(t *testing.T) {
	client := NewClient(false)

	apps := []App{
		{AppID: "1:123:android:abc", Platform: "android"},
		{AppID: "1:123:ios:def", Platform: "ios"},
		{AppID: "1:123:android:ghi", Platform: "android"},
		{AppID: "1:123:web:jkl", Platform: "web"},
	}

	androidApps := client.filterAppsByPlatform(apps, "android")
	if len(androidApps) != 2 {
		t.Errorf("Expected 2 Android apps, got %d", len(androidApps))
	}

	iosApps := client.filterAppsByPlatform(apps, "ios")
	if len(iosApps) != 1 {
		t.Errorf("Expected 1 iOS app, got %d", len(iosApps))
	}

	webApps := client.filterAppsByPlatform(apps, "web")
	if len(webApps) != 1 {
		t.Errorf("Expected 1 web app, got %d", len(webApps))
	}
}

func BenchmarkExtractAppIDFromOutput(b *testing.B) {
	client := NewClient(false)
	output := `✔ Firebase project some-project is ready! Here's your app info:
App ID: 1:123456789:android:abcdef123456
Display Name: My App (Android)`

	for i := 0; i < b.N; i++ {
		client.extractAppIDFromOutput(output)
	}
}
