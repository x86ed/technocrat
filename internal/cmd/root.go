package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "0.5.1"
	commit  = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "technocrat",
	Short: "Technocrat - MCP Server and Management Tool",
	Long: `Technocrat is a Model Context Protocol (MCP) server implementation
with built-in installation and management capabilities.

Use 'technocrat server' to run the MCP server.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

// SetVersionInfo sets the version information for the CLI
func SetVersionInfo(v, c string) {
	version = v
	commit = c
}

func init() {
	rootCmd.Version = fmt.Sprintf("%s (commit: %s)", version, commit)
	rootCmd.SetVersionTemplate(`{{printf "Technocrat %s\n" .Version}}`)
}
