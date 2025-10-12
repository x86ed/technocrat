package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print version and build information for Technocrat.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Technocrat MCP Server\n")
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Commit:  %s\n", commit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
