package cmd

import (
	"fmt"
	"os"

	"github.com/clix-so/nativefire/internal/ui"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var verbose bool

// Version information
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// SetVersion sets the version information
func SetVersion(v, c, d string) {
	version, commit, date = v, c, d
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
}

var rootCmd = &cobra.Command{
	Use:   "nativefire",
	Short: "🔥 Simplify Firebase setup in native development environments",
	Long: color.New(color.FgHiBlue).Sprint(`
🔥 NativeFire - Firebase Setup Made Easy

NativeFire automatically detects your native development environment and sets up 
Firebase configuration for multiple platforms. Think of it as flutterfire for native apps!

`) + color.New(color.FgHiGreen).Sprint("Supported Platforms:") + `
  📱 iOS       🤖 Android     🖥️  macOS
  🪟 Windows   🐧 Linux       🌐 Web

` + color.New(color.FgHiYellow).Sprint("Quick Start:") + `
  ` + ui.Code("nativefire configure") + `                       # Let nativefire do everything (auto-detect enabled)
  ` + ui.Code("nativefire projects list") + `                   # See your Firebase projects  
  ` + ui.Code("nativefire projects select") + `                 # Pick a project interactively

` + color.New(color.FgHiCyan).Sprint("Advanced Usage:") + `
  ` + ui.Code("nativefire configure --project my-app --platform ios") + `
  ` + ui.Code("nativefire configure --bundle-id com.example.app") + `
  ` + ui.Code("nativefire configure --package-name com.example.app") + `

Need help? Use ` + ui.Code("nativefire [command] --help") + ` for detailed information.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.nativefire.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".nativefire")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
