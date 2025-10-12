package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"technocrat/internal/installer"
)

var (
	installDir string
	configPath string
	useSystemd bool
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the Technocrat MCP server",
	Long: `Install the Technocrat MCP server to the system.

This command builds and installs the server binary to the specified
directory (default: /usr/local/bin). On Linux systems, you can optionally
install a systemd service for automatic startup.`,
	RunE: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().StringVarP(&installDir, "dir", "d", "/usr/local/bin", "Installation directory")
	installCmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to configuration file")
	installCmd.Flags().BoolVar(&useSystemd, "systemd", false, "Install systemd service (Linux only)")
}

func runInstall(cmd *cobra.Command, args []string) error {
	log.Println("Installing Technocrat MCP Server...")

	inst := installer.New(installDir, configPath)

	if err := inst.Install(useSystemd); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	log.Println("✓ Successfully installed Technocrat MCP Server")
	log.Printf("✓ Server binary installed to: %s/technocrat-server", installDir)

	if useSystemd {
		log.Println("✓ Systemd service installed. Run: sudo systemctl start technocrat")
	}

	return nil
}
