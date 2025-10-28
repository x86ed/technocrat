package templates

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

// data embeds all template files from the data/ subdirectory
//
//go:embed all:data
var data embed.FS

// FS returns the embedded templates filesystem
func FS() fs.FS {
	// Return sub-filesystem rooted at "data" to hide the path prefix
	fsys, err := fs.Sub(data, "data")
	if err != nil {
		panic("failed to create sub-filesystem: " + err.Error())
	}
	return fsys
}

// GetTemplate reads a template file by name from the root of the templates directory
// Example: GetTemplate("spec-template.md")
func GetTemplate(name string) ([]byte, error) {
	fsys := FS()
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("failed to read template %s: %w", name, err)
	}
	return data, nil
}

// GetCommandTemplate reads a command template file from the commands/ subdirectory
// Example: GetCommandTemplate("spec.md")
func GetCommandTemplate(name string) ([]byte, error) {
	fsys := FS()
	commandPath := path.Join("commands", name)
	data, err := fs.ReadFile(fsys, commandPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read command template %s: %w", name, err)
	}
	return data, nil
}

// GetVSCodeSettings reads the vscode-settings.json template
func GetVSCodeSettings() ([]byte, error) {
	return GetTemplate("vscode-settings.json")
}

// ListCommands returns a list of all command template filenames (without the .md extension)
// Example: ["analyze", "checklist", "clarify", ...]
func ListCommands() ([]string, error) {
	fsys := FS()
	entries, err := fs.ReadDir(fsys, "commands")
	if err != nil {
		return nil, fmt.Errorf("failed to read commands directory: %w", err)
	}

	var commands []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			// Remove .md extension
			name := strings.TrimSuffix(entry.Name(), ".md")
			commands = append(commands, name)
		}
	}
	return commands, nil
}

// ListTemplates returns a list of all template filenames in the root directory
func ListTemplates() ([]string, error) {
	fsys := FS()
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	var templates []string
	for _, entry := range entries {
		if !entry.IsDir() {
			templates = append(templates, entry.Name())
		}
	}
	return templates, nil
}
