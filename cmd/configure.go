package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/clix-so/nativefire/internal/dependencies"
	"github.com/clix-so/nativefire/internal/firebase"
	"github.com/clix-so/nativefire/internal/platform"
	"github.com/clix-so/nativefire/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	projectID    string
	platformFlag string
	autoDetect   bool
	appID        string
	bundleID     string
	packageName  string
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "🚀 Configure Firebase for your native application",
	Long: ui.Primary.Sprint("🚀 Firebase Configuration Wizard\n\n") +
		"Automatically configure Firebase for your native application with smart detection and setup.\n\n" +
		ui.Bold.Sprint("Features:") + "\n" +
		"  • " + ui.Platform("📱") + " Multi-platform support (iOS, Android, macOS, Windows, Linux)\n" +
		"  • 🎯 Smart project and platform detection\n" +
		"  • 📦 Automatic config file placement\n" +
		"  • 🔧 Code injection for Firebase initialization\n" +
		"  • 🔍 Bundle ID and Package Name auto-detection\n\n" +
		ui.Bold.Sprint("Quick Examples:") + "\n" +
		"  " + ui.Code("nativefire configure") + "                               # Full auto mode (default)\n" +
		"  " + ui.Code("nativefire configure --project my-app") +
		"              # Auto-detect platform for specific project\n" +
		"  " + ui.Code("nativefire configure --project my-app --platform ios") + "  # Explicit platform\n\n" +
		ui.Bold.Sprint("Platform-Specific Options:") + "\n" +
		"  " + ui.Code("--bundle-id") + "     - iOS/macOS Bundle Identifier\n" +
		"  " + ui.Code("--package-name") + "  - Android Package Name\n" +
		"  " + ui.Code("--app-id") + "        - Use existing Firebase App ID\n\n" +
		ui.Dim.Sprint("Pro tip: Use") + " " + ui.Code("--verbose") + " " + ui.Dim.Sprint("to see detailed progress."),
	RunE: runConfigure,
}

func init() {
	rootCmd.AddCommand(configureCmd)

	configureCmd.Flags().StringVarP(&projectID, "project", "p", "", "Firebase project ID (will prompt if not provided)")
	configureCmd.Flags().StringVar(&platformFlag, "platform", "", "Target platform (android, ios, macos, windows, linux)")
	configureCmd.Flags().BoolVar(&autoDetect, "auto-detect", true,
		"Automatically detect the platform (enabled by default)")
	configureCmd.Flags().StringVar(&appID, "app-id", "", "Firebase app ID (optional, will generate if not provided)")
	configureCmd.Flags().StringVar(&bundleID, "bundle-id", "",
		"iOS Bundle ID (will auto-detect or generate if not provided)")
	configureCmd.Flags().StringVar(&packageName, "package-name", "",
		"Android Package Name (will auto-detect or generate if not provided)")

	// Make project optional - we'll prompt if not provided
}

func runConfigure(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")

	// Perform preflight dependency check
	platformForCheck := "all"
	if platformFlag != "" {
		platformForCheck = platformFlag
	}

	if err := dependencies.PreflightCheck(platformForCheck); err != nil {
		return fmt.Errorf("dependency check failed: %w", err)
	}

	firebaseClient := firebase.NewClient(verbose)

	// If project ID not provided, prompt user to select
	if projectID == "" {
		selectedProjectID, err := promptProjectSelection(firebaseClient, verbose)
		if err != nil {
			return err
		}
		projectID = selectedProjectID
	}

	// Validate project exists and user has access
	if verbose {
		ui.InfoMsg(fmt.Sprintf("Validating Firebase project: %s", projectID))
	}

	if err := firebaseClient.ValidateProject(projectID); err != nil {
		return fmt.Errorf("project validation failed: %w", err)
	}

	if verbose {
		ui.SuccessMsg("Project validation successful")
	}

	ui.ProjectHeader(projectID)

	var targetPlatform platform.Platform
	var err error

	switch {
	case platformFlag != "":
		targetPlatform, err = platform.FromString(platformFlag)
		if err != nil {
			return fmt.Errorf("invalid platform: %w", err)
		}
		fmt.Printf("%s %s\n\n", ui.Check.Sprint("🎯 Using platform:"), ui.Platform(targetPlatform.Name()))
	case autoDetect:
		ui.Step(1, "Auto-detecting platform...")
		targetPlatform, err = platform.DetectPlatform()
		if err != nil {
			return fmt.Errorf("failed to detect platform: %w", err)
		}
		fmt.Printf("   %s %s\n\n", ui.Check.Sprint("🎯 Detected platform:"), ui.Platform(targetPlatform.Name()))
	default:
		return fmt.Errorf("platform detection failed: auto-detect is disabled and no platform specified")
	}

	config := &firebase.Config{
		ProjectID:   projectID,
		AppID:       appID,
		Platform:    targetPlatform,
		BundleID:    bundleID,
		PackageName: packageName,
	}

	ui.Step(2, "Registering app with Firebase...")
	err = firebaseClient.RegisterApp(config)
	if err != nil {
		return fmt.Errorf("failed to register app with Firebase: %w", err)
	}

	ui.Step(3, "Downloading configuration file...")
	err = firebaseClient.DownloadConfig(config)
	if err != nil {
		return fmt.Errorf("failed to download configuration: %w", err)
	}

	ui.Step(4, "Installing configuration file...")
	err = targetPlatform.InstallConfig(config)
	if err != nil {
		return fmt.Errorf("failed to install configuration: %w", err)
	}

	// Additional platform-specific dependency check before initialization
	if verbose {
		ui.InfoMsg("Checking platform-specific dependencies...")
	}
	platformSpecificMissing := dependencies.CheckAllDependencies(strings.ToLower(targetPlatform.Name()))
	if len(platformSpecificMissing) > 0 {
		dependencies.ShowMissingDependencies(platformSpecificMissing)
	}

	ui.Step(5, "Adding Firebase initialization code...")
	err = targetPlatform.AddInitializationCode(config)
	if err != nil {
		return fmt.Errorf("failed to add initialization code: %w", err)
	}

	fmt.Printf("\n🎉 %s %s!\n",
		ui.Success.Sprint("Firebase configuration completed successfully for"),
		ui.Platform(targetPlatform.Name()))
	return nil
}

func promptProjectSelection(firebaseClient *firebase.Client, verbose bool) (string, error) {
	// In test environments, return an error instead of prompting for input
	if isTestEnvironment() {
		return "", fmt.Errorf("project ID is required in test environment")
	}

	if verbose {
		ui.InfoMsg("No project specified. Fetching available Firebase projects...")
	}

	projects, err := firebaseClient.ListProjects()
	if err != nil {
		return "", fmt.Errorf("failed to fetch projects for selection: %w", err)
	}

	if len(projects) == 0 {
		ui.WarningMsg("No Firebase projects found")
		fmt.Printf("%s Create your first project at %s\n",
			ui.Fire.Sprint("🔗"),
			ui.Secondary.Sprint("https://console.firebase.google.com/"))
		return "", fmt.Errorf("no Firebase projects available")
	}

	// If only one project, use it automatically
	if len(projects) == 1 {
		fmt.Printf("%s %s (%s)\n",
			ui.Info.Sprint("💡 Found 1 Firebase project:"),
			ui.Bold.Sprint(projects[0].DisplayName),
			ui.Secondary.Sprint(projects[0].ProjectID))
		fmt.Printf("%s %s\n\n",
			ui.Success.Sprint("🎯 Auto-selecting project:"),
			ui.Secondary.Sprint(projects[0].ProjectID))
		return projects[0].ProjectID, nil
	}

	// Multiple projects - show selection menu
	ui.Header("Select Firebase Project")
	fmt.Printf("Found %s project(s). Choose one:\n\n", ui.Success.Sprint(fmt.Sprintf("%d", len(projects))))

	for i, project := range projects {
		fmt.Printf("  %s %s\n",
			ui.Primary.Sprint(fmt.Sprintf("[%d]", i+1)),
			ui.Bold.Sprint(project.DisplayName))
		fmt.Printf("      %s %s\n",
			ui.Dim.Sprint("ID:"),
			ui.Secondary.Sprint(project.ProjectID))
		fmt.Println()
	}

	// Get user selection
	fmt.Printf("%s ", ui.Primary.Sprint(fmt.Sprintf("Select a project (1-%d):", len(projects))))
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	selection, err := strconv.Atoi(input)
	if err != nil {
		ui.ErrorMsg(fmt.Sprintf("Invalid selection '%s'. Please enter a number between 1 and %d", input, len(projects)))
		return "", fmt.Errorf("invalid selection '%s'", input)
	}

	if selection < 1 || selection > len(projects) {
		ui.ErrorMsg(fmt.Sprintf("Selection %d is out of range. Please enter a number between 1 and %d",
			selection, len(projects)))
		return "", fmt.Errorf("selection %d is out of range", selection)
	}

	selectedProject := projects[selection-1]
	fmt.Printf("\n%s %s (%s)\n\n",
		ui.Success.Sprint("✅ Selected project:"),
		ui.Bold.Sprint(selectedProject.DisplayName),
		ui.Secondary.Sprint(selectedProject.ProjectID))

	return selectedProject.ProjectID, nil
}

// isTestEnvironment checks if we're running in a test environment
func isTestEnvironment() bool {
	// Check if we're running under go test
	for _, arg := range os.Args {
		if strings.Contains(arg, "test") || strings.HasSuffix(arg, ".test") {
			return true
		}
	}

	// Check for test-specific environment variables
	if os.Getenv("GO_TEST") == "1" || os.Getenv("TESTING") == "1" {
		return true
	}

	return false
}
