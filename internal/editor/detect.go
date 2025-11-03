package editor

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Editor represents a detected editor installation
type Editor struct {
	Name      string
	Type      EditorType
	Installed bool
	ConfigDir string
	Version   string
	Transport TransportType
}

// EditorType defines the type of editor
type EditorType string

const (
	VSCode        EditorType = "vscode"
	ClaudeDesktop EditorType = "claude"
	Cursor        EditorType = "cursor"
	AmazonQ       EditorType = "amazonq"
	Windsurf      EditorType = "windsurf"
)

// TransportType defines how the MCP server communicates
type TransportType string

const (
	Stdio TransportType = "stdio" // JSON-RPC over stdin/stdout
	HTTP  TransportType = "http"  // REST API
)

// DetectEditors scans for installed editors
func DetectEditors() []Editor {
	editors := []Editor{
		detectVSCode(),
		detectClaudeDesktop(),
		detectCursor(),
		detectAmazonQ(),
		detectWindsurf(),
	}

	// Filter to only installed editors
	installed := []Editor{}
	for _, editor := range editors {
		if editor.Installed {
			installed = append(installed, editor)
		}
	}

	return installed
}

// detectVSCode detects VS Code installation
func detectVSCode() Editor {
	editor := Editor{
		Name:      "VS Code",
		Type:      VSCode,
		Installed: false,
		Transport: Stdio,
	}

	// Try to find VS Code executable or application
	var found bool
	var versionCmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		// First try the command line tool
		if cmd := exec.Command("which", "code"); cmd != nil {
			if output, err := cmd.Output(); err == nil && len(output) > 0 {
				found = true
				versionCmd = exec.Command("code", "--version")
			}
		}

		// If not found, check for the application bundle
		if !found {
			appPath := "/Applications/Visual Studio Code.app"
			if _, err := os.Stat(appPath); err == nil {
				found = true
				// Use the full path to the code binary inside the app bundle
				codePath := filepath.Join(appPath, "Contents", "Resources", "app", "bin", "code")
				versionCmd = exec.Command(codePath, "--version")
			}
		}
	case "windows":
		if cmd := exec.Command("where", "code"); cmd != nil {
			if output, err := cmd.Output(); err == nil && len(output) > 0 {
				found = true
				versionCmd = exec.Command("code", "--version")
			}
		}
	default: // linux
		if cmd := exec.Command("which", "code"); cmd != nil {
			if output, err := cmd.Output(); err == nil && len(output) > 0 {
				found = true
				versionCmd = exec.Command("code", "--version")
			}
		}
	}

	if found {
		editor.Installed = true

		// Get version if possible
		if versionCmd != nil {
			if versionOutput, err := versionCmd.Output(); err == nil {
				lines := strings.Split(string(versionOutput), "\n")
				if len(lines) > 0 {
					editor.Version = strings.TrimSpace(lines[0])
				}
			}
		}

		// Determine config directory
		editor.ConfigDir = getVSCodeConfigDir()
	}

	return editor
}

// detectClaudeDesktop detects Claude Desktop installation
func detectClaudeDesktop() Editor {
	editor := Editor{
		Name:      "Claude Desktop",
		Type:      ClaudeDesktop,
		Installed: false,
		Transport: Stdio,
	}

	configDir := getClaudeConfigDir()
	configPath := filepath.Join(configDir, "claude_desktop_config.json")

	// Check if config file exists
	if _, err := os.Stat(configPath); err == nil {
		editor.Installed = true
		editor.ConfigDir = configDir
	}

	return editor
}

// detectCursor detects Cursor editor installation
func detectCursor() Editor {
	editor := Editor{
		Name:      "Cursor",
		Type:      Cursor,
		Installed: false,
		Transport: Stdio,
	}

	// Try to find cursor executable
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("which", "cursor")
	case "windows":
		cmd = exec.Command("where", "cursor")
	default: // linux
		cmd = exec.Command("which", "cursor")
	}

	if output, err := cmd.Output(); err == nil && len(output) > 0 {
		editor.Installed = true
		editor.ConfigDir = getCursorConfigDir()
	}

	return editor
}

// detectAmazonQ detects Amazon Q CLI installation
func detectAmazonQ() Editor {
	editor := Editor{
		Name:      "Amazon Q",
		Type:      AmazonQ,
		Installed: false,
		Transport: HTTP, // Amazon Q uses HTTP transport
	}

	// Try to find q executable
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("which", "q")
	case "windows":
		cmd = exec.Command("where", "q")
	default: // linux
		cmd = exec.Command("which", "q")
	}

	if output, err := cmd.Output(); err == nil && len(output) > 0 {
		editor.Installed = true

		// Get version
		versionCmd := exec.Command("q", "--version")
		if versionOutput, err := versionCmd.Output(); err == nil {
			editor.Version = strings.TrimSpace(string(versionOutput))
		}

		editor.ConfigDir = getAmazonQConfigDir()
	}

	return editor
}

// detectWindsurf detects Windsurf IDE installation
func detectWindsurf() Editor {
	editor := Editor{
		Name:      "Windsurf",
		Type:      Windsurf,
		Installed: false,
		Transport: Stdio,
	}

	// Try to find windsurf executable
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("which", "windsurf")
	case "windows":
		cmd = exec.Command("where", "windsurf")
	default: // linux
		cmd = exec.Command("which", "windsurf")
	}

	if output, err := cmd.Output(); err == nil && len(output) > 0 {
		editor.Installed = true
		editor.ConfigDir = getWindsurfConfigDir()
	}

	return editor
}

// getVSCodeConfigDir returns the VS Code configuration directory
func getVSCodeConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Code", "User")
	case "windows":
		return filepath.Join(home, "AppData", "Roaming", "Code", "User")
	default: // linux
		return filepath.Join(home, ".config", "Code", "User")
	}
}

// getClaudeConfigDir returns the Claude Desktop configuration directory
func getClaudeConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Claude")
	case "windows":
		return filepath.Join(home, "AppData", "Roaming", "Claude")
	default: // linux
		return filepath.Join(home, ".config", "claude")
	}
}

// getCursorConfigDir returns the Cursor configuration directory
func getCursorConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, ".cursor")
	case "windows":
		return filepath.Join(home, ".cursor")
	default: // linux
		return filepath.Join(home, ".cursor")
	}
}

// getAmazonQConfigDir returns the Amazon Q configuration directory
func getAmazonQConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ".aws", "q")
}

// getWindsurfConfigDir returns the Windsurf configuration directory
func getWindsurfConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Windsurf")
	case "windows":
		return filepath.Join(home, "AppData", "Roaming", "Windsurf")
	default: // linux
		return filepath.Join(home, ".config", "windsurf")
	}
}

// GetEditor returns a specific editor by type
func GetEditor(editorType EditorType) *Editor {
	editors := DetectEditors()
	for _, editor := range editors {
		if editor.Type == editorType {
			return &editor
		}
	}
	return nil
}

// IsInstalled checks if a specific editor type is installed
func IsInstalled(editorType EditorType) bool {
	editor := GetEditor(editorType)
	return editor != nil && editor.Installed
}
