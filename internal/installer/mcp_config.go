package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"technocrat/internal/editor"
)

// MCPServerConfig represents the MCP server configuration
type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// getTechnocratPath returns the best path to the technocrat binary
func getTechnocratPath() (string, error) {
	// First try to find technocrat in PATH
	if path, err := exec.LookPath("technocrat"); err == nil {
		// Make sure it's an absolute path
		if absPath, err := filepath.Abs(path); err == nil {
			return absPath, nil
		}
		return path, nil
	}

	// Fall back to current executable path
	if path, err := os.Executable(); err == nil {
		// Resolve any symlinks
		if realPath, err := filepath.EvalSymlinks(path); err == nil {
			return realPath, nil
		}
		return path, nil
	}

	return "", fmt.Errorf("could not determine technocrat binary path")
}

// InstallMCPConfig installs MCP server configuration for an editor
func InstallMCPConfig(ed editor.Editor, projectPath string) error {
	var err error
	switch ed.Type {
	case editor.VSCode:
		err = installVSCodeMCP(projectPath)
	case editor.ClaudeDesktop:
		err = installClaudeMCP(ed.ConfigDir)
	case editor.Cursor:
		err = installCursorMCP(ed.ConfigDir)
	case editor.AmazonQ:
		err = installAmazonQMCP(ed.ConfigDir)
	case editor.Windsurf:
		err = installWindsurfMCP(projectPath)
	default:
		return fmt.Errorf("unsupported editor: %s", ed.Name)
	}
	
	if err != nil {
		return err
	}
	
	// Verify the configuration was written successfully
	if !IsMCPConfigured(ed, projectPath) {
		return fmt.Errorf("MCP configuration verification failed for %s", ed.Name)
	}
	
	return nil
}

// installVSCodeMCP configures MCP for VS Code
func installVSCodeMCP(projectPath string) error {
	settingsPath := filepath.Join(projectPath, ".vscode", "settings.json")

	// Create .vscode directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return fmt.Errorf("failed to create .vscode directory: %w", err)
	}

	// Load existing settings or create new
	settings := make(map[string]interface{})
	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			// If unmarshal fails, continue with empty settings
			settings = make(map[string]interface{})
		}
	}

	// Get technocrat binary path
	technocratPath, err := getTechnocratPath()
	if err != nil {
		return fmt.Errorf("failed to determine technocrat path: %w", err)
	}

	// Add MCP server configuration for GitHub Copilot
	mcpServers := map[string]interface{}{
		"technocrat": map[string]interface{}{
			"command": technocratPath,
			"args":    []string{"server"},
		},
	}

	settings["github.copilot.chat.codeGeneration.instructions"] = []map[string]string{
		{"text": "Use Technocrat MCP server for spec-driven development workflows"},
	}
	settings["github.copilot.chat.mcpServers"] = mcpServers

	// Write settings
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings: %w", err)
	}

	return nil
}

// installClaudeMCP configures MCP for Claude Desktop
func installClaudeMCP(configDir string) error {
	configPath := filepath.Join(configDir, "claude_desktop_config.json")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load existing config or create new
	config := make(map[string]interface{})
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			config = make(map[string]interface{})
		}
	}

	// Get technocrat binary path
	technocratPath, err := getTechnocratPath()
	if err != nil {
		return fmt.Errorf("failed to determine technocrat path: %w", err)
	}

	// Add MCP server configuration
	mcpServers := make(map[string]interface{})
	if existing, ok := config["mcpServers"].(map[string]interface{}); ok {
		mcpServers = existing
	}

	mcpServers["technocrat"] = map[string]interface{}{
		"command": technocratPath,
		"args":    []string{"server"},
	}

	config["mcpServers"] = mcpServers

	// Write config
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// installCursorMCP configures MCP for Cursor
func installCursorMCP(configDir string) error {
	configPath := filepath.Join(configDir, "mcp_servers.json")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load existing config or create new
	config := make(map[string]interface{})
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			config = make(map[string]interface{})
		}
	}

	// Get technocrat binary path
	technocratPath, err := getTechnocratPath()
	if err != nil {
		return fmt.Errorf("failed to determine technocrat path: %w", err)
	}

	// Add MCP server configuration
	config["technocrat"] = map[string]interface{}{
		"command": technocratPath,
		"args":    []string{"server"},
	}

	// Write config
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// installAmazonQMCP configures MCP for Amazon Q
func installAmazonQMCP(configDir string) error {
	configPath := filepath.Join(configDir, "mcp-config.json")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load existing config or create new
	config := make(map[string]interface{})
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			config = make(map[string]interface{})
		}
	}

	// Amazon Q uses HTTP transport
	config["technocrat"] = map[string]interface{}{
		"type": "http",
		"url":  "http://localhost:8080/mcp/v1",
	}

	// Write config
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// installWindsurfMCP configures MCP for Windsurf
func installWindsurfMCP(projectPath string) error {
	configPath := filepath.Join(projectPath, ".windsurf", "mcp_config.json")

	// Create .windsurf directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create .windsurf directory: %w", err)
	}

	// Get technocrat binary path
	technocratPath, err := getTechnocratPath()
	if err != nil {
		return fmt.Errorf("failed to determine technocrat path: %w", err)
	}

	// Create MCP config
	config := map[string]interface{}{
		"technocrat": map[string]interface{}{
			"command": technocratPath,
			"args":    []string{"server"},
		},
	}

	// Write config
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GetMCPConfigPath returns the MCP configuration file path for an editor
func GetMCPConfigPath(ed editor.Editor, projectPath string) string {
	switch ed.Type {
	case editor.VSCode:
		return filepath.Join(projectPath, ".vscode", "settings.json")
	case editor.ClaudeDesktop:
		return filepath.Join(ed.ConfigDir, "claude_desktop_config.json")
	case editor.Cursor:
		return filepath.Join(ed.ConfigDir, "mcp_servers.json")
	case editor.AmazonQ:
		return filepath.Join(ed.ConfigDir, "mcp-config.json")
	case editor.Windsurf:
		return filepath.Join(projectPath, ".windsurf", "mcp_config.json")
	default:
		return ""
	}
}

// IsMCPConfigured checks if MCP is already configured for an editor
func IsMCPConfigured(ed editor.Editor, projectPath string) bool {
	configPath := GetMCPConfigPath(ed, projectPath)
	if configPath == "" {
		return false
	}

	// Check if config file exists
	data, err := os.ReadFile(configPath)
	if err != nil {
		return false
	}

	// Check if config contains technocrat
	config := make(map[string]interface{})
	if err := json.Unmarshal(data, &config); err != nil {
		return false
	}

	// Look for technocrat in different config structures
	switch ed.Type {
	case editor.VSCode:
		if servers, ok := config["github.copilot.chat.mcpServers"].(map[string]interface{}); ok {
			_, exists := servers["technocrat"]
			return exists
		}
	case editor.ClaudeDesktop, editor.Cursor:
		if servers, ok := config["mcpServers"].(map[string]interface{}); ok {
			_, exists := servers["technocrat"]
			return exists
		}
		if _, exists := config["technocrat"]; exists {
			return true
		}
	case editor.AmazonQ, editor.Windsurf:
		_, exists := config["technocrat"]
		return exists
	}

	return false
}
