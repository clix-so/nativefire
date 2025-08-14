package firebase

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/clix-so/nativefire/internal/ui"
)

// Constants for repeated strings
const (
	androidPlatform = "android"
	iosPlatform     = "ios"
	macosPlatform   = "macos"
	successStatus   = "success"
	activeState     = "ACTIVE"
)

type Config struct {
	ProjectID   string
	AppID       string
	Platform    PlatformInterface
	BundleID    string // For iOS/macOS apps
	PackageName string // For Android apps
	ConfigFile  string // Path to downloaded config file
}

type PlatformInterface interface {
	Name() string
	ConfigFileName() string
	ConfigPath() string
}

type Client struct {
	verbose bool
}

type App struct {
	AppID       string `json:"appId"`
	DisplayName string `json:"displayName"`
	ProjectID   string `json:"projectId"`
	Platform    string `json:"platform"`
	BundleID    string `json:"bundleId,omitempty"`
	PackageName string `json:"packageName,omitempty"`
	Namespace   string `json:"namespace,omitempty"` // Firebase uses 'namespace' for both bundle ID and package name
}

type Project struct {
	ProjectID     string         `json:"projectId"`
	ProjectNumber string         `json:"projectNumber"`
	DisplayName   string         `json:"displayName"`
	Name          string         `json:"name"`
	Resources     map[string]any `json:"resources"`
	State         string         `json:"state"`
	Etag          string         `json:"etag"`
}

type ProjectsListResponse struct {
	Status string    `json:"status"`
	Result []Project `json:"result"`
}

type AppsListResponse struct {
	Status string `json:"status"`
	Result []App  `json:"result"`
}

func NewClient(verbose bool) *Client {
	return &Client{
		verbose: verbose,
	}
}

func (c *Client) checkFirebaseCLI() error {
	_, err := exec.LookPath("firebase")
	if err != nil {
		return fmt.Errorf("firebase CLI not found. Please install it first: npm install -g firebase-tools")
	}
	return nil
}

func (c *Client) checkAuthentication() error {
	cmd := exec.Command("firebase", "projects:list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "not authenticated") {
			return fmt.Errorf("not authenticated with Firebase. Please run: firebase login")
		}
		return fmt.Errorf("failed to check authentication: %w", err)
	}
	return nil
}

func (c *Client) RegisterApp(config *Config) error {
	if err := c.checkFirebaseCLI(); err != nil {
		return err
	}

	if err := c.checkAuthentication(); err != nil {
		return err
	}

	if config.AppID != "" {
		if c.verbose {
			ui.InfoMsg(fmt.Sprintf("Using existing app ID: %s", config.AppID))
		}
		return nil
	}

	// Check if an existing app matches our bundle ID/package name
	existingApp, err := c.FindExistingApp(config)
	if err != nil {
		if c.verbose {
			ui.WarningMsg(fmt.Sprintf("Could not check for existing apps: %v", err))
		}
	} else if existingApp != nil {
		config.AppID = existingApp.AppID
		ui.SuccessMsg(fmt.Sprintf("Using existing %s app: %s (%s)",
			existingApp.Platform,
			existingApp.DisplayName,
			existingApp.AppID))
		if c.verbose {
			ui.InfoMsg("Skipping app creation - existing app found and configured")
		}
		return nil
	}

	platformFlag := c.getPlatformFlag(config.Platform.Name())
	appName := c.generateAppName(config.Platform.Name())

	// Build the command with platform-specific identifiers
	cmd := c.buildCreateAppCommand(platformFlag, appName, config)

	if c.verbose {
		fmt.Printf("%s %s\n", ui.Dim.Sprint("Running:"), ui.Code(c.formatCommand(cmd.Args)))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return c.handleAppCreationError(config, string(output))
	}

	appID := c.extractAppIDFromOutput(string(output))
	if appID == "" {
		return fmt.Errorf("failed to extract app ID from Firebase CLI output")
	}

	config.AppID = appID

	if c.verbose {
		ui.SuccessMsg(fmt.Sprintf("Created Firebase app with ID: %s", appID))
	}

	return nil
}

func (c *Client) DownloadConfig(config *Config) error {
	if config.AppID == "" {
		return fmt.Errorf("app ID is required to download configuration")
	}

	// Generate a unique temp file path without creating the file
	// Let Firebase CLI create the file to avoid permission conflicts
	configExt := filepath.Ext(config.Platform.ConfigFileName())
	configBase := strings.TrimSuffix(config.Platform.ConfigFileName(), configExt)

	// Create a temporary file just to get a unique name, then remove it
	tempFile, err := os.CreateTemp("", fmt.Sprintf("nativefire_%s_*%s", configBase, configExt))
	if err != nil {
		return fmt.Errorf("failed to generate temp filename: %w", err)
	}
	configFile := tempFile.Name()
	tempFile.Close()
	os.Remove(configFile) // Remove the file so Firebase CLI can create it fresh

	// Store the temp file path in config for platform implementations to use
	config.ConfigFile = configFile

	var cmd *exec.Cmd
	platformName := strings.ToLower(config.Platform.Name())

	switch platformName {
	case androidPlatform:
		cmd = exec.Command("firebase", "apps:sdkconfig", androidPlatform, config.AppID,
			"--project", config.ProjectID, "--out", configFile)
	case iosPlatform, macosPlatform:
		cmd = exec.Command("firebase", "apps:sdkconfig", iosPlatform, config.AppID,
			"--project", config.ProjectID, "--out", configFile)
	default:
		return fmt.Errorf("platform %s does not support automatic config download", platformName)
	}

	if c.verbose {
		fmt.Printf("%s %s\n", ui.Dim.Sprint("Running:"), ui.Code(c.formatCommand(cmd.Args)))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Clean up temp file if download fails
		os.Remove(configFile)
		return fmt.Errorf("failed to download config: %s", string(output))
	}

	if c.verbose {
		ui.SuccessMsg(fmt.Sprintf("Configuration downloaded to: %s", configFile))
	}

	return nil
}

func (c *Client) getPlatformFlag(platformName string) string {
	switch strings.ToLower(platformName) {
	case androidPlatform:
		return androidPlatform
	case iosPlatform, macosPlatform:
		return iosPlatform
	case "web":
		return "web"
	default:
		return androidPlatform
	}
}

func (c *Client) generateAppName(platformName string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("My %s App", platformName)
	}

	projectName := filepath.Base(cwd)
	// Use safer format without parentheses to avoid shell interpretation issues
	return fmt.Sprintf("%s %s", projectName, platformName)
}

func (c *Client) extractAppIDFromOutput(output string) string {
	// Try to match "App ID: [app-id]" format
	appIDRegex := regexp.MustCompile(`App ID:\s+([^\s\n]+)`)
	if matches := appIDRegex.FindStringSubmatch(output); len(matches) > 1 {
		return matches[1]
	}

	// Try to match JSON format: "appId": "app-id"
	jsonRegex := regexp.MustCompile(`"appId":\s*"([^"]+)"`)
	if matches := jsonRegex.FindStringSubmatch(output); len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func (c *Client) buildCreateAppCommand(platformFlag, appName string, config *Config) *exec.Cmd {
	args := []string{"apps:create", platformFlag, appName, "--project", config.ProjectID}

	// Add platform-specific identifiers
	switch strings.ToLower(config.Platform.Name()) {
	case androidPlatform:
		packageName := config.PackageName
		if packageName == "" {
			packageName = c.detectAndroidPackageName()
		}
		if packageName == "" {
			// Use a default package name based on project
			packageName = c.generateDefaultPackageName(config.ProjectID)
		}
		args = append(args, "--package-name", packageName)
		if c.verbose {
			ui.InfoMsg(fmt.Sprintf("Using Android package name: %s", packageName))
		}
	case iosPlatform, macosPlatform:
		bundleID := config.BundleID
		if bundleID == "" {
			bundleID = c.detectIOSBundleID()
		}
		if bundleID == "" {
			// Use a default bundle ID based on project
			bundleID = c.generateDefaultBundleID(config.ProjectID)
		}
		args = append(args, "--bundle-id", bundleID)
		if c.verbose {
			ui.InfoMsg(fmt.Sprintf("Using iOS bundle ID: %s", bundleID))
		}
	}

	return exec.Command("firebase", args...)
}

func (c *Client) detectAndroidPackageName() string {
	// Try to find package name in build.gradle files
	buildGradleFiles := []string{
		"app/build.gradle",
		"android/app/build.gradle",
		"build.gradle",
	}

	for _, file := range buildGradleFiles {
		if packageName := c.extractPackageNameFromBuildGradle(file); packageName != "" {
			return packageName
		}
	}

	// Try to find package name in AndroidManifest.xml
	manifestFiles := []string{
		"app/src/main/AndroidManifest.xml",
		"android/app/src/main/AndroidManifest.xml",
		"src/main/AndroidManifest.xml",
	}

	for _, file := range manifestFiles {
		if packageName := c.extractPackageNameFromManifest(file); packageName != "" {
			return packageName
		}
	}

	return ""
}

func (c *Client) detectIOSBundleID() string {
	// Try to find bundle ID in Info.plist files
	infoPlistFiles := []string{
		"ios/Runner/Info.plist",
		"Info.plist",
		"Runner/Info.plist",
	}

	for _, file := range infoPlistFiles {
		if bundleID := c.extractBundleIDFromInfoPlist(file); bundleID != "" {
			return bundleID
		}
	}

	// Try to find in project.pbxproj files
	pbxprojFiles, _ := filepath.Glob("*.xcodeproj/project.pbxproj")
	for _, file := range pbxprojFiles {
		if bundleID := c.extractBundleIDFromPbxproj(file); bundleID != "" {
			return bundleID
		}
	}

	// Try iOS subdirectory
	iosPbxprojFiles, _ := filepath.Glob("ios/*.xcodeproj/project.pbxproj")
	for _, file := range iosPbxprojFiles {
		if bundleID := c.extractBundleIDFromPbxproj(file); bundleID != "" {
			return bundleID
		}
	}

	return ""
}

func (c *Client) extractPackageNameFromBuildGradle(filename string) string {
	content, err := os.ReadFile(filename)
	if err != nil {
		return ""
	}

	// Look for applicationId in build.gradle
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "applicationId") {
			// Extract applicationId "com.example.app"
			if idx := strings.Index(line, "\""); idx != -1 {
				remaining := line[idx+1:]
				if idx2 := strings.Index(remaining, "\""); idx2 != -1 {
					return remaining[:idx2]
				}
			}
		}
	}
	return ""
}

func (c *Client) extractPackageNameFromManifest(filename string) string {
	content, err := os.ReadFile(filename)
	if err != nil {
		return ""
	}

	// Look for package attribute in AndroidManifest.xml
	contentStr := string(content)
	if idx := strings.Index(contentStr, "package=\""); idx != -1 {
		start := idx + len("package=\"")
		if end := strings.Index(contentStr[start:], "\""); end != -1 {
			return contentStr[start : start+end]
		}
	}
	return ""
}

func (c *Client) extractBundleIDFromInfoPlist(filename string) string {
	content, err := os.ReadFile(filename)
	if err != nil {
		return ""
	}

	// Look for CFBundleIdentifier in Info.plist
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.Contains(line, "CFBundleIdentifier") && i+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[i+1])
			// Extract from <string>com.example.app</string>
			if strings.HasPrefix(nextLine, "<string>") && strings.HasSuffix(nextLine, "</string>") {
				bundleID := nextLine[8 : len(nextLine)-9] // Remove <string> and </string>
				if bundleID != "$(PRODUCT_BUNDLE_IDENTIFIER)" {
					return bundleID
				}
			}
		}
	}
	return ""
}

func (c *Client) extractBundleIDFromPbxproj(filename string) string {
	content, err := os.ReadFile(filename)
	if err != nil {
		return ""
	}

	// Look for PRODUCT_BUNDLE_IDENTIFIER in project.pbxproj
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "PRODUCT_BUNDLE_IDENTIFIER") {
			// Extract PRODUCT_BUNDLE_IDENTIFIER = com.example.app;
			if idx := strings.Index(line, "="); idx != -1 {
				remaining := strings.TrimSpace(line[idx+1:])
				remaining = strings.TrimSuffix(remaining, ";")
				remaining = strings.Trim(remaining, "\"")
				if remaining != "" && !strings.Contains(remaining, "$") {
					return remaining
				}
			}
		}
	}
	return ""
}

func (c *Client) generateDefaultPackageName(projectID string) string {
	// Generate a valid Android package name from project ID
	// Replace hyphens with dots and ensure it starts with a domain-like structure
	sanitized := strings.ReplaceAll(projectID, "-", ".")
	if !strings.Contains(sanitized, ".") {
		return fmt.Sprintf("com.firebase.%s", sanitized)
	}
	return fmt.Sprintf("com.%s", sanitized)
}

func (c *Client) generateDefaultBundleID(projectID string) string {
	// Generate a valid iOS bundle ID from project ID
	// Replace hyphens with dots and ensure it starts with a domain-like structure
	sanitized := strings.ReplaceAll(projectID, "-", ".")
	if !strings.Contains(sanitized, ".") {
		return fmt.Sprintf("com.firebase.%s", sanitized)
	}
	return fmt.Sprintf("com.%s", sanitized)
}

func (c *Client) ListProjects() ([]Project, error) {
	if err := c.checkFirebaseCLI(); err != nil {
		return nil, err
	}

	if err := c.checkAuthentication(); err != nil {
		return nil, err
	}

	cmd := exec.Command("firebase", "projects:list", "--json")

	if c.verbose {
		fmt.Printf("%s %s\n", ui.Dim.Sprint("Running:"), ui.Code(c.formatCommand(cmd.Args)))
	}

	output, err := cmd.Output() // Only capture stdout, ignore stderr
	if err != nil {
		return nil, fmt.Errorf("failed to list Firebase projects: %w", err)
	}

	var response ProjectsListResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse projects response: %w", err)
	}

	if response.Status != successStatus {
		return nil, fmt.Errorf("Firebase CLI returned non-success status: %s", response.Status)
	}

	// Filter only active projects
	var activeProjects []Project
	for _, project := range response.Result {
		if project.State == activeState {
			activeProjects = append(activeProjects, project)
		}
	}

	return activeProjects, nil
}

func (c *Client) ValidateProject(projectID string) error {
	projects, err := c.ListProjects()
	if err != nil {
		return fmt.Errorf("failed to validate project: %w", err)
	}

	for _, project := range projects {
		if project.ProjectID == projectID {
			return nil
		}
	}

	return fmt.Errorf("project '%s' not found or you don't have access to it", projectID)
}

func (c *Client) ListApps(projectID string) ([]App, error) {
	if err := c.checkFirebaseCLI(); err != nil {
		return nil, err
	}

	if err := c.checkAuthentication(); err != nil {
		return nil, err
	}

	cmd := exec.Command("firebase", "apps:list", "--json", "--project", projectID)

	if c.verbose {
		fmt.Printf("%s %s\n", ui.Dim.Sprint("Running:"), ui.Code(c.formatCommand(cmd.Args)))
	}

	output, err := cmd.Output() // Only capture stdout, ignore stderr
	if err != nil {
		return nil, fmt.Errorf("failed to list Firebase apps: %w", err)
	}

	var response AppsListResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse apps response: %w", err)
	}

	if response.Status != successStatus {
		return nil, fmt.Errorf("Firebase CLI returned non-success status: %s", response.Status)
	}

	return response.Result, nil
}

func (c *Client) FindExistingApp(config *Config) (*App, error) {
	apps, err := c.ListApps(config.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list apps: %w", err)
	}

	platformName := strings.ToLower(config.Platform.Name())

	// Search through apps for matching platform and identifier
	for _, app := range apps {
		if strings.ToLower(app.Platform) != platformName {
			continue
		}

		if match := c.checkAppMatch(app, config, platformName); match {
			return &app, nil
		}
	}

	return nil, nil // No existing app found
}

// checkAppMatch checks if an app matches the expected identifier for a platform
func (c *Client) checkAppMatch(app App, config *Config, platformName string) bool {
	if platformName == iosPlatform || platformName == macosPlatform {
		return c.checkIOSAppMatch(app, config)
	}

	if platformName == androidPlatform {
		return c.checkAndroidAppMatch(app, config)
	}

	return false
}

// checkIOSAppMatch checks if an iOS/macOS app matches the expected bundle ID
func (c *Client) checkIOSAppMatch(app App, config *Config) bool {
	expectedBundleID := config.BundleID
	if expectedBundleID == "" {
		expectedBundleID = c.detectIOSBundleID()
	}
	if expectedBundleID == "" {
		expectedBundleID = c.generateDefaultBundleID(config.ProjectID)
	}

	// Check both bundleId field and namespace field
	bundleIDToCheck := app.BundleID
	if bundleIDToCheck == "" {
		bundleIDToCheck = app.Namespace
	}

	if bundleIDToCheck == expectedBundleID {
		if c.verbose {
			ui.InfoMsg(fmt.Sprintf("Found existing %s app with Bundle ID: %s", app.Platform, bundleIDToCheck))
		}
		return true
	}

	return false
}

// checkAndroidAppMatch checks if an Android app matches the expected package name
func (c *Client) checkAndroidAppMatch(app App, config *Config) bool {
	expectedPackageName := config.PackageName
	if expectedPackageName == "" {
		expectedPackageName = c.detectAndroidPackageName()
	}
	if expectedPackageName == "" {
		expectedPackageName = c.generateDefaultPackageName(config.ProjectID)
	}

	// Check both packageName field and namespace field
	packageNameToCheck := app.PackageName
	if packageNameToCheck == "" {
		packageNameToCheck = app.Namespace
	}

	if packageNameToCheck == expectedPackageName {
		if c.verbose {
			ui.InfoMsg(fmt.Sprintf("Found existing %s app with Package Name: %s", app.Platform, packageNameToCheck))
		}
		return true
	}

	return false
}

// formatCommand properly quotes command arguments that contain spaces or special characters
func (c *Client) formatCommand(args []string) string {
	quotedArgs := make([]string, len(args))
	for i, arg := range args {
		if strings.ContainsAny(arg, " \t\n()[]{}") {
			quotedArgs[i] = fmt.Sprintf(`"%s"`, arg)
		} else {
			quotedArgs[i] = arg
		}
	}
	return strings.Join(quotedArgs, " ")
}

// isDuplicateAppError checks if the Firebase CLI error indicates a duplicate app
func (c *Client) isDuplicateAppError(output string) bool {
	// Common indicators of duplicate app errors
	errorIndicators := []string{
		"already exists",
		"duplicate",
		"bundle id already exists",
		"package name already exists",
		"Bundle ID for iOS app cannot be empty", // Sometimes this indicates bundle ID conflict
		"Failed to create iOS app",
		"Failed to create Android app",
	}

	outputLower := strings.ToLower(output)
	for _, indicator := range errorIndicators {
		if strings.Contains(outputLower, strings.ToLower(indicator)) {
			return true
		}
	}

	return false
}

// findExistingAppByIdentifier performs a more thorough search for existing apps
// resolveExpectedIdentifier gets the expected identifier for the platform
func (c *Client) resolveExpectedIdentifier(config *Config, platformName string) string {
	if platformName == iosPlatform || platformName == macosPlatform {
		expectedIdentifier := config.BundleID
		if expectedIdentifier == "" {
			expectedIdentifier = c.detectIOSBundleID()
		}
		if expectedIdentifier == "" {
			expectedIdentifier = c.generateDefaultBundleID(config.ProjectID)
		}
		return expectedIdentifier
	}

	if platformName == androidPlatform {
		expectedIdentifier := config.PackageName
		if expectedIdentifier == "" {
			expectedIdentifier = c.detectAndroidPackageName()
		}
		if expectedIdentifier == "" {
			expectedIdentifier = c.generateDefaultPackageName(config.ProjectID)
		}
		return expectedIdentifier
	}

	return ""
}

// filterAppsByPlatform filters apps by platform name
func (c *Client) filterAppsByPlatform(apps []App, platformName string) []App {
	var platformApps []App
	for _, app := range apps {
		if strings.ToLower(app.Platform) == platformName {
			platformApps = append(platformApps, app)
		}
	}
	return platformApps
}

// logPlatformApps logs platform apps for debugging
func (c *Client) logPlatformApps(platformApps []App, platformName string) {
	ui.InfoMsg(fmt.Sprintf("Found %d %s apps in project:", len(platformApps), platformName))
	for i, app := range platformApps {
		ui.InfoMsg(fmt.Sprintf("  %d. %s (namespace: %s, bundleId: %s, packageName: %s)",
			i+1, app.DisplayName, app.Namespace, app.BundleID, app.PackageName))
	}
	ui.InfoMsg("Searching for exact matches only...")
}

// findMatchingApp searches for an app matching the expected identifier
func (c *Client) findMatchingApp(platformApps []App, expectedIdentifier, platformName string) *App {
	for _, app := range platformApps {
		// Check the namespace field (which contains bundle ID or package name)
		if app.Namespace == expectedIdentifier {
			if c.verbose {
				ui.InfoMsg(fmt.Sprintf("Found exact match by namespace: %s", app.Namespace))
			}
			return &app
		}

		// Also check the specific fields as fallback
		if platformName == iosPlatform || platformName == macosPlatform {
			if app.BundleID == expectedIdentifier {
				if c.verbose {
					ui.InfoMsg(fmt.Sprintf("Found exact match by bundle ID: %s", app.BundleID))
				}
				return &app
			}
		} else if platformName == androidPlatform {
			if app.PackageName == expectedIdentifier {
				if c.verbose {
					ui.InfoMsg(fmt.Sprintf("Found exact match by package name: %s", app.PackageName))
				}
				return &app
			}
		}
	}
	return nil
}

// handleAppCreationError handles errors during app creation
func (c *Client) handleAppCreationError(config *Config, outputStr string) error {
	// Check if this is a duplicate app error
	if c.isDuplicateAppError(outputStr) {
		if c.verbose {
			ui.WarningMsg("App creation failed, searching for existing app...")
		}

		// Try to find the existing app again with a more thorough search
		existingApp, findErr := c.findExistingAppByIdentifier(config)
		if findErr == nil && existingApp != nil {
			config.AppID = existingApp.AppID
			ui.SuccessMsg(fmt.Sprintf("Found and using existing %s app: %s (%s)",
				existingApp.Platform,
				existingApp.DisplayName,
				existingApp.AppID))
			return nil
		}

		// If we still can't find an existing app, provide helpful guidance
		if c.verbose {
			ui.WarningMsg("Could not find existing app to use as fallback")
			c.suggestManualCreation(config)
		}
	}

	return fmt.Errorf("failed to create Firebase app: %s", outputStr)
}

func (c *Client) findExistingAppByIdentifier(config *Config) (*App, error) {
	apps, err := c.ListApps(config.ProjectID)
	if err != nil {
		return nil, err
	}

	platformName := strings.ToLower(config.Platform.Name())

	// Get the expected identifier (bundle ID or package name)
	expectedIdentifier := c.resolveExpectedIdentifier(config, platformName)

	if c.verbose {
		ui.InfoMsg(fmt.Sprintf("Searching for existing %s app with identifier: %s", platformName, expectedIdentifier))
		ui.InfoMsg(fmt.Sprintf("Found %d total apps in project", len(apps)))
	}

	// First, list all apps of the target platform for debugging
	platformApps := c.filterAppsByPlatform(apps, platformName)

	if c.verbose {
		c.logPlatformApps(platformApps, platformName)
	}

	// Search through all apps for matching platform and identifier
	matchedApp := c.findMatchingApp(platformApps, expectedIdentifier, platformName)
	if matchedApp != nil {
		return matchedApp, nil
	}

	if c.verbose {
		ui.WarningMsg(fmt.Sprintf("No existing %s app found with identifier: %s", platformName, expectedIdentifier))
	}

	return nil, nil
}

// suggestManualCreation provides helpful guidance when automatic app creation fails
func (c *Client) suggestManualCreation(config *Config) {
	platformName := strings.ToLower(config.Platform.Name())

	ui.InfoMsg("Manual creation options:")

	if platformName == iosPlatform || platformName == macosPlatform {
		expectedBundleID := config.BundleID
		if expectedBundleID == "" {
			expectedBundleID = c.detectIOSBundleID()
		}
		if expectedBundleID == "" {
			expectedBundleID = c.generateDefaultBundleID(config.ProjectID)
		}

		fmt.Printf("  1. Create an app manually in Firebase Console with Bundle ID: %s\n",
			ui.Secondary.Sprint(expectedBundleID))
		fmt.Printf("  2. Or run: %s\n",
			ui.Code(fmt.Sprintf("firebase apps:create ios \"My App\" --project %s --bundle-id %s",
				config.ProjectID, expectedBundleID)))
	} else if platformName == androidPlatform {
		expectedPackageName := config.PackageName
		if expectedPackageName == "" {
			expectedPackageName = c.detectAndroidPackageName()
		}
		if expectedPackageName == "" {
			expectedPackageName = c.generateDefaultPackageName(config.ProjectID)
		}

		fmt.Printf("  1. Create an app manually in Firebase Console with Package Name: %s\n",
			ui.Secondary.Sprint(expectedPackageName))
		fmt.Printf("  2. Or run: %s\n",
			ui.Code(fmt.Sprintf("firebase apps:create android \"My App\" --project %s --package-name %s",
				config.ProjectID, expectedPackageName)))
	}

	fmt.Printf("  3. Then use the app ID with: %s\n",
		ui.Code(fmt.Sprintf("nativefire configure --project %s --app-id YOUR_APP_ID", config.ProjectID)))
}
