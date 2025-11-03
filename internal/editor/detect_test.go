package editor

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDetectEditors(t *testing.T) {
	editors := DetectEditors()

	// Should return a slice (may be empty)
	if editors == nil {
		t.Error("DetectEditors returned nil")
	}

	// All returned editors should be installed
	for _, editor := range editors {
		if !editor.Installed {
			t.Errorf("Editor %s returned but not marked as installed", editor.Name)
		}
	}
}

func TestEditorTypes(t *testing.T) {
	tests := []struct {
		name         string
		editorType   EditorType
		expectedName string
	}{
		{"VSCode type", VSCode, "vscode"},
		{"Claude type", ClaudeDesktop, "claude"},
		{"Cursor type", Cursor, "cursor"},
		{"AmazonQ type", AmazonQ, "amazonq"},
		{"Windsurf type", Windsurf, "windsurf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.editorType) != tt.expectedName {
				t.Errorf("Expected %s, got %s", tt.expectedName, string(tt.editorType))
			}
		})
	}
}

func TestTransportTypes(t *testing.T) {
	tests := []struct {
		name           string
		transportType  TransportType
		expectedString string
	}{
		{"Stdio transport", Stdio, "stdio"},
		{"HTTP transport", HTTP, "http"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.transportType) != tt.expectedString {
				t.Errorf("Expected %s, got %s", tt.expectedString, string(tt.transportType))
			}
		})
	}
}

func TestDetectVSCode(t *testing.T) {
	editor := detectVSCode()

	if editor.Name != "VS Code" {
		t.Errorf("Expected name 'VS Code', got '%s'", editor.Name)
	}

	if editor.Type != VSCode {
		t.Errorf("Expected type VSCode, got %s", editor.Type)
	}

	if editor.Transport != Stdio {
		t.Errorf("Expected transport Stdio, got %s", editor.Transport)
	}

	// If installed, should have config dir
	if editor.Installed && editor.ConfigDir == "" {
		t.Error("VS Code marked as installed but ConfigDir is empty")
	}

	// Verify executable exists if marked as installed
	if editor.Installed {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = exec.Command("where", "code")
		default:
			cmd = exec.Command("which", "code")
		}
		if _, err := cmd.Output(); err != nil {
			t.Error("VS Code marked as installed but 'code' command not found")
		}
	}
}

func TestDetectClaudeDesktop(t *testing.T) {
	editor := detectClaudeDesktop()

	if editor.Name != "Claude Desktop" {
		t.Errorf("Expected name 'Claude Desktop', got '%s'", editor.Name)
	}

	if editor.Type != ClaudeDesktop {
		t.Errorf("Expected type ClaudeDesktop, got %s", editor.Type)
	}

	if editor.Transport != Stdio {
		t.Errorf("Expected transport Stdio, got %s", editor.Transport)
	}

	// If installed, config file should exist
	if editor.Installed {
		configPath := filepath.Join(editor.ConfigDir, "claude_desktop_config.json")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Claude marked as installed but config file doesn't exist")
		}
	}
}

func TestDetectCursor(t *testing.T) {
	editor := detectCursor()

	if editor.Name != "Cursor" {
		t.Errorf("Expected name 'Cursor', got '%s'", editor.Name)
	}

	if editor.Type != Cursor {
		t.Errorf("Expected type Cursor, got %s", editor.Type)
	}

	if editor.Transport != Stdio {
		t.Errorf("Expected transport Stdio, got %s", editor.Transport)
	}
}

func TestDetectAmazonQ(t *testing.T) {
	editor := detectAmazonQ()

	if editor.Name != "Amazon Q" {
		t.Errorf("Expected name 'Amazon Q', got '%s'", editor.Name)
	}

	if editor.Type != AmazonQ {
		t.Errorf("Expected type AmazonQ, got %s", editor.Type)
	}

	if editor.Transport != HTTP {
		t.Errorf("Expected transport HTTP, got %s", editor.Transport)
	}
}

func TestDetectWindsurf(t *testing.T) {
	editor := detectWindsurf()

	if editor.Name != "Windsurf" {
		t.Errorf("Expected name 'Windsurf', got '%s'", editor.Name)
	}

	if editor.Type != Windsurf {
		t.Errorf("Expected type Windsurf, got %s", editor.Type)
	}

	if editor.Transport != Stdio {
		t.Errorf("Expected transport Stdio, got %s", editor.Transport)
	}
}

func TestGetVSCodeConfigDir(t *testing.T) {
	configDir := getVSCodeConfigDir()

	if configDir == "" {
		t.Error("Config dir should not be empty")
	}

	// Should contain expected path components
	switch runtime.GOOS {
	case "darwin":
		if !contains(configDir, "Library") || !contains(configDir, "Application Support") {
			t.Errorf("macOS config dir doesn't match expected pattern: %s", configDir)
		}
	case "windows":
		if !contains(configDir, "AppData") || !contains(configDir, "Roaming") {
			t.Errorf("Windows config dir doesn't match expected pattern: %s", configDir)
		}
	default: // linux
		if !contains(configDir, ".config") {
			t.Errorf("Linux config dir doesn't match expected pattern: %s", configDir)
		}
	}
}

func TestGetClaudeConfigDir(t *testing.T) {
	configDir := getClaudeConfigDir()

	if configDir == "" {
		t.Error("Config dir should not be empty")
	}

	// Should contain "Claude"
	if !contains(configDir, "Claude") && !contains(configDir, "claude") {
		t.Errorf("Config dir should contain 'Claude': %s", configDir)
	}
}

func TestGetCursorConfigDir(t *testing.T) {
	configDir := getCursorConfigDir()

	if configDir == "" {
		t.Error("Config dir should not be empty")
	}

	// Should contain ".cursor"
	if !contains(configDir, ".cursor") {
		t.Errorf("Config dir should contain '.cursor': %s", configDir)
	}
}

func TestGetAmazonQConfigDir(t *testing.T) {
	configDir := getAmazonQConfigDir()

	if configDir == "" {
		t.Error("Config dir should not be empty")
	}

	// Should contain ".aws/q"
	if !contains(configDir, ".aws") {
		t.Errorf("Config dir should contain '.aws': %s", configDir)
	}
}

func TestGetWindsurfConfigDir(t *testing.T) {
	configDir := getWindsurfConfigDir()

	if configDir == "" {
		t.Error("Config dir should not be empty")
	}

	// Should contain "Windsurf" or "windsurf"
	if !contains(configDir, "Windsurf") && !contains(configDir, "windsurf") {
		t.Errorf("Config dir should contain 'Windsurf': %s", configDir)
	}
}

func TestGetEditor(t *testing.T) {
	// Test getting a specific editor
	editor := GetEditor(VSCode)

	// May be nil if not installed
	if editor != nil {
		if editor.Type != VSCode {
			t.Errorf("Expected VSCode type, got %s", editor.Type)
		}

		if !editor.Installed {
			t.Error("Returned editor should be installed")
		}
	}
}

func TestIsInstalled(t *testing.T) {
	// Test each editor type
	editorTypes := []EditorType{VSCode, ClaudeDesktop, Cursor, AmazonQ, Windsurf}

	for _, editorType := range editorTypes {
		result := IsInstalled(editorType)

		// Result should be boolean
		if result {
			// If installed, GetEditor should return non-nil
			editor := GetEditor(editorType)
			if editor == nil {
				t.Errorf("IsInstalled returned true but GetEditor returned nil for %s", editorType)
			}
		}
	}
}

func TestEditorStructure(t *testing.T) {
	editor := Editor{
		Name:      "Test Editor",
		Type:      VSCode,
		Installed: true,
		ConfigDir: "/test/config",
		Version:   "1.0.0",
		Transport: Stdio,
	}

	if editor.Name != "Test Editor" {
		t.Errorf("Expected Name 'Test Editor', got '%s'", editor.Name)
	}

	if editor.Type != VSCode {
		t.Errorf("Expected Type VSCode, got %s", editor.Type)
	}

	if !editor.Installed {
		t.Error("Expected Installed to be true")
	}

	if editor.ConfigDir != "/test/config" {
		t.Errorf("Expected ConfigDir '/test/config', got '%s'", editor.ConfigDir)
	}

	if editor.Version != "1.0.0" {
		t.Errorf("Expected Version '1.0.0', got '%s'", editor.Version)
	}

	if editor.Transport != Stdio {
		t.Errorf("Expected Transport Stdio, got %s", editor.Transport)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && containsSubstring(s, substr)
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
