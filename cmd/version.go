package cmd

import (
	"github.com/spf13/cobra"
)

var Version = "1.0.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of nativefire",
	Long:  `Print the version number of nativefire`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("nativefire v%s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
