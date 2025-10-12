package installer

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Installer handles the installation of the Technocrat MCP server
type Installer struct {
	installDir string
	configPath string
}

// New creates a new installer instance
func New(installDir, configPath string) *Installer {
	return &Installer{
		installDir: installDir,
		configPath: configPath,
	}
}

// Install installs the MCP server
func (i *Installer) Install(systemd bool) error {
	// Build the server binary
	if err := i.buildServer(); err != nil {
		return fmt.Errorf("failed to build server: %w", err)
	}

	// Copy binary to installation directory
	if err := i.copyBinary(); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Create configuration directory
	if err := i.createConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Install systemd service if requested (Linux only)
	if systemd && runtime.GOOS == "linux" {
		if err := i.installSystemdService(); err != nil {
			return fmt.Errorf("failed to install systemd service: %w", err)
		}
	}

	return nil
}

// Uninstall removes the MCP server
func (i *Installer) Uninstall(systemd bool) error {
	// Stop and remove systemd service if it exists
	if systemd && runtime.GOOS == "linux" {
		if err := i.uninstallSystemdService(); err != nil {
			return fmt.Errorf("failed to uninstall systemd service: %w", err)
		}
	}

	// Remove binary
	binaryPath := filepath.Join(i.installDir, "technocrat-server")
	if err := os.Remove(binaryPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove binary: %w", err)
	}

	return nil
}

// buildServer builds the server binary
func (i *Installer) buildServer() error {
	cmd := exec.Command("go", "build", "-o", "technocrat-server", "./cmd/server")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// copyBinary copies the server binary to the installation directory
func (i *Installer) copyBinary() error {
	// Ensure installation directory exists
	if err := os.MkdirAll(i.installDir, 0755); err != nil {
		return err
	}

	// Open source file
	src, err := os.Open("technocrat-server")
	if err != nil {
		return err
	}
	defer src.Close()

	// Create destination file
	destPath := filepath.Join(i.installDir, "technocrat-server")
	dest, err := os.OpenFile(destPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer dest.Close()

	// Copy file
	if _, err := io.Copy(dest, src); err != nil {
		return err
	}

	// Remove temporary binary
	os.Remove("technocrat-server")

	return nil
}

// createConfigDir creates the configuration directory
func (i *Installer) createConfigDir() error {
	var configDir string

	switch runtime.GOOS {
	case "linux":
		configDir = "/etc/technocrat"
	case "darwin":
		homeDir, _ := os.UserHomeDir()
		configDir = filepath.Join(homeDir, ".config", "technocrat")
	default:
		homeDir, _ := os.UserHomeDir()
		configDir = filepath.Join(homeDir, ".technocrat")
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Create default config file if it doesn't exist
	configFile := filepath.Join(configDir, "config.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		defaultConfig := `{
  "port": 8080,
  "log_level": "info"
}
`
		if err := os.WriteFile(configFile, []byte(defaultConfig), 0644); err != nil {
			return err
		}
	}

	return nil
}

// installSystemdService installs the systemd service
func (i *Installer) installSystemdService() error {
	serviceContent := `[Unit]
Description=Technocrat MCP Server
After=network.target

[Service]
Type=simple
ExecStart=%s/technocrat-server
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
`
	serviceContent = fmt.Sprintf(serviceContent, i.installDir)

	servicePath := "/etc/systemd/system/technocrat.service"
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return err
	}

	// Reload systemd
	cmd := exec.Command("systemctl", "daemon-reload")
	if err := cmd.Run(); err != nil {
		return err
	}

	// Enable service
	cmd = exec.Command("systemctl", "enable", "technocrat")
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// uninstallSystemdService removes the systemd service
func (i *Installer) uninstallSystemdService() error {
	// Stop service
	cmd := exec.Command("systemctl", "stop", "technocrat")
	cmd.Run() // Ignore error if service doesn't exist

	// Disable service
	cmd = exec.Command("systemctl", "disable", "technocrat")
	cmd.Run() // Ignore error if service doesn't exist

	// Remove service file
	servicePath := "/etc/systemd/system/technocrat.service"
	os.Remove(servicePath)

	// Reload systemd
	cmd = exec.Command("systemctl", "daemon-reload")
	cmd.Run()

	return nil
}

// GetSystemInfo returns system information
func GetSystemInfo() map[string]string {
	info := make(map[string]string)

	info["os"] = runtime.GOOS
	info["arch"] = runtime.GOARCH
	info["go_version"] = runtime.Version()

	// Get home directory
	if home, err := os.UserHomeDir(); err == nil {
		info["home_dir"] = home
	}

	// Get current user
	if output, err := exec.Command("whoami").Output(); err == nil {
		info["user"] = strings.TrimSpace(string(output))
	}

	return info
}
