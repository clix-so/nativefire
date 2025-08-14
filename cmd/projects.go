package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/clix-so/nativefire/internal/firebase"
	"github.com/clix-so/nativefire/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "🔥 Manage Firebase projects",
	Long: ui.Primary.Sprint("🔥 Firebase Project Management\n\n") +
		"Discover and manage your Firebase projects with ease.\n\n" +
		ui.Bold.Sprint("Available Commands:") + "\n" +
		"  • " + ui.Code("list") + "   - Show all your Firebase projects\n" +
		"  • " + ui.Code("select") + " - Pick a project interactively\n\n" +
		ui.Dim.Sprint("Pro tip: Use") + " " + ui.Code("--verbose") + " " + ui.Dim.Sprint("for detailed output."),
}

var projectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "📋 List all available Firebase projects",
	Long: ui.Info.Sprint("📋 Firebase Projects Overview\n\n") +
		"View all Firebase projects you have access to in a clean, organized format.\n\n" +
		ui.Bold.Sprint("Features:") + "\n" +
		"  • Clean, colorful project listing\n" +
		"  • Project IDs and display names\n" +
		"  • Quick copy-paste ready format\n" +
		"  • Usage examples included\n\n" +
		ui.Dim.Sprint("Example:") + " " + ui.Code("nativefire projects list"),
	RunE: runProjectsList,
}

var projectsSelectCmd = &cobra.Command{
	Use:   "select",
	Short: "🎯 Interactively select a Firebase project",
	Long: ui.Success.Sprint("🎯 Interactive Project Selection\n\n") +
		"Choose your Firebase project from a beautiful, interactive list.\n\n" +
		ui.Bold.Sprint("Features:") + "\n" +
		"  • Interactive project picker\n" +
		"  • Real-time project information\n" +
		"  • Automatic configuration guidance\n" +
		"  • Optional Firebase CLI integration\n\n" +
		ui.Bold.Sprint("Flags:") + "\n" +
		"  " + ui.Code("--use") + " - Set selected project as Firebase CLI default\n\n" +
		ui.Dim.Sprint("Example:") + " " + ui.Code("nativefire projects select --use"),
	RunE: runProjectsSelect,
}

func init() {
	rootCmd.AddCommand(projectsCmd)
	projectsCmd.AddCommand(projectsListCmd)
	projectsCmd.AddCommand(projectsSelectCmd)

	projectsSelectCmd.Flags().BoolVar(&autoUse, "use", false, "Automatically use the selected project for configuration")
}

var autoUse bool

func runProjectsList(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")
	firebaseClient := firebase.NewClient(verbose)

	if verbose {
		ui.InfoMsg("Fetching Firebase projects...")
	}

	projects, err := firebaseClient.ListProjects()
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	if len(projects) == 0 {
		ui.WarningMsg("No Firebase projects found")
		fmt.Printf("\n%s Create your first project at %s\n",
			ui.Fire.Sprint("🔗"),
			ui.Secondary.Sprint("https://console.firebase.google.com/"))
		return nil
	}

	// Beautiful header
	ui.Header("Your Firebase Projects")
	fmt.Printf("Found %s project(s)\n\n", ui.Success.Sprint(fmt.Sprintf("%d", len(projects))))

	// Display projects in a clean, modern format
	for i, project := range projects {
		// Project number with fire emoji
		fmt.Printf("%s %s\n",
			ui.Fire.Sprint(fmt.Sprintf("%d.", i+1)),
			ui.Bold.Sprint(project.DisplayName))

		// Project ID with subtle styling
		fmt.Printf("   %s %s\n",
			ui.Dim.Sprint("ID:"),
			ui.Secondary.Sprint(project.ProjectID))

		// Project number in verbose mode
		if verbose {
			fmt.Printf("   %s %s\n",
				ui.Dim.Sprint("Number:"),
				ui.Dim.Sprint(project.ProjectNumber))
		}

		fmt.Println() // Empty line between projects
	}

	// Usage instructions
	fmt.Printf("%s\n", ui.Bold.Sprint("Quick Start:"))
	fmt.Printf("  %s\n", ui.Code("nativefire configure --project <PROJECT_ID>"))
	fmt.Printf("  %s\n", ui.Code("nativefire projects select"))

	fmt.Printf("\n%s Copy any Project ID above and use it with the configure command.\n",
		ui.Info.Sprint("💡"))

	return nil
}

func runProjectsSelect(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")
	firebaseClient := firebase.NewClient(verbose)

	if verbose {
		ui.InfoMsg("Fetching Firebase projects...")
	}

	projects, err := firebaseClient.ListProjects()
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	if len(projects) == 0 {
		ui.WarningMsg("No Firebase projects found")
		fmt.Printf("\n%s Create your first project at %s\n",
			ui.Fire.Sprint("🔗"),
			ui.Secondary.Sprint("https://console.firebase.google.com/"))
		return nil
	}

	// Beautiful header for selection
	ui.Header("Select Your Firebase Project")
	fmt.Printf("Choose from %s available project(s):\n\n", ui.Success.Sprint(fmt.Sprintf("%d", len(projects))))

	// Display projects with beautiful formatting
	for i, project := range projects {
		fmt.Printf("  %s %s\n",
			ui.Primary.Sprint(fmt.Sprintf("[%d]", i+1)),
			ui.Bold.Sprint(project.DisplayName))
		fmt.Printf("      %s %s\n",
			ui.Dim.Sprint("ID:"),
			ui.Secondary.Sprint(project.ProjectID))
		if verbose {
			fmt.Printf("      %s %s\n",
				ui.Dim.Sprint("Number:"),
				ui.Dim.Sprint(project.ProjectNumber))
		}
		fmt.Println()
	}

	// Get user selection with styled prompt
	fmt.Printf("%s ", ui.Primary.Sprint(fmt.Sprintf("Select a project (1-%d):", len(projects))))
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	selection, err := strconv.Atoi(input)
	if err != nil {
		ui.ErrorMsg(fmt.Sprintf("Invalid selection: %s", input))
		return fmt.Errorf("invalid selection: %s", input)
	}

	if selection < 1 || selection > len(projects) {
		ui.ErrorMsg(fmt.Sprintf("Selection out of range: %d (valid: 1-%d)", selection, len(projects)))
		return fmt.Errorf("selection out of range: %d (valid range: 1-%d)", selection, len(projects))
	}

	selectedProject := projects[selection-1]

	// Success message with project info
	fmt.Printf("\n%s %s\n",
		ui.Check.Sprint("🎉 Project Selected:"),
		ui.Bold.Sprint(selectedProject.DisplayName))
	fmt.Printf("   %s %s\n\n",
		ui.Dim.Sprint("Project ID:"),
		ui.Success.Sprint(selectedProject.ProjectID))

	if autoUse {
		ui.InfoMsg("Setting as default project for Firebase CLI...")
		fmt.Printf("%s %s\n",
			ui.Dim.Sprint("Command:"),
			ui.Code(fmt.Sprintf("firebase use %s", selectedProject.ProjectID)))
	}

	// Next steps
	fmt.Printf("%s\n", ui.Bold.Sprint("Next Steps:"))
	fmt.Printf("  %s %s\n",
		ui.Rocket.Sprint("🚀"),
		ui.Code(fmt.Sprintf("nativefire configure --project %s", selectedProject.ProjectID)))

	fmt.Printf("\n%s Project ID ready to use: %s\n",
		ui.Info.Sprint("💡"),
		ui.Secondary.Sprint(selectedProject.ProjectID))

	return nil
}
