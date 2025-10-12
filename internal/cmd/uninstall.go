package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"technocrat/internal/installer"
)

var (
	uninstallSystemd bool
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the Technocrat MCP server",
	Long: `Uninstall the Technocrat MCP server from the system.

This command removes the server binary and optionally removes
the systemd service if it was installed.`,
	RunE: runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)

	uninstallCmd.Flags().BoolVar(&uninstallSystemd, "systemd", false, "Remove systemd service (Linux only)")
	uninstallCmd.Flags().StringVarP(&installDir, "dir", "d", "/usr/local/bin", "Installation directory")
}

func runUninstall(cmd *cobra.Command, args []string) error {
	log.Println("Uninstalling Technocrat MCP Server...")

	inst := installer.New(installDir, "")

	if err := inst.Uninstall(uninstallSystemd); err != nil {
		return fmt.Errorf("uninstallation failed: %w", err)
	}

	log.Println("✓ Successfully uninstalled Technocrat MCP Server")

	return nil
}
