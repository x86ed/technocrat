package templates

import (
	"io/fs"
	"testing"
)

func TestEmbeddedFS(t *testing.T) {
	fsys := FS()

	// Test that we can read files without the "data" prefix
	expectedFiles := []string{
		"agent-file-template.md",
		"checklist-template.md",
		"plan-template.md",
		"spec-template.md",
		"tasks-template.md",
		"vscode-settings.json",
		"commands/analyze.md",
		"commands/checklist.md",
		"commands/clarify.md",
		"commands/constitution.md",
		"commands/implement.md",
		"commands/plan.md",
		"commands/spec.md",
		"commands/tasks.md",
	}

	for _, filename := range expectedFiles {
		t.Run(filename, func(t *testing.T) {
			data, err := fs.ReadFile(fsys, filename)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", filename, err)
			}
			if len(data) == 0 {
				t.Errorf("File %s is empty", filename)
			}
		})
	}
}

func TestFS_DirectoryStructure(t *testing.T) {
	fsys := FS()

	// Verify commands directory exists
	entries, err := fs.ReadDir(fsys, "commands")
	if err != nil {
		t.Fatalf("Failed to read commands directory: %v", err)
	}

	if len(entries) == 0 {
		t.Error("commands directory is empty")
	}

	// Count command files
	commandCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && len(entry.Name()) > 3 && entry.Name()[len(entry.Name())-3:] == ".md" {
			commandCount++
		}
	}

	expectedCommandCount := 8
	if commandCount != expectedCommandCount {
		t.Errorf("Expected %d command files, found %d", expectedCommandCount, commandCount)
	}
}

func TestGetTemplate(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{"spec template", "spec-template.md", false},
		{"plan template", "plan-template.md", false},
		{"tasks template", "tasks-template.md", false},
		{"checklist template", "checklist-template.md", false},
		{"agent file template", "agent-file-template.md", false},
		{"vscode settings", "vscode-settings.json", false},
		{"nonexistent file", "nonexistent.md", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := GetTemplate(tt.filename)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(data) == 0 {
					t.Error("Template data is empty")
				}
			}
		})
	}
}

func TestGetCommandTemplate(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{"spec command", "spec.md", false},
		{"plan command", "plan.md", false},
		{"tasks command", "tasks.md", false},
		{"checklist command", "checklist.md", false},
		{"analyze command", "analyze.md", false},
		{"clarify command", "clarify.md", false},
		{"constitution command", "constitution.md", false},
		{"implement command", "implement.md", false},
		{"nonexistent command", "nonexistent.md", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := GetCommandTemplate(tt.filename)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(data) == 0 {
					t.Error("Command template data is empty")
				}
			}
		})
	}
}

func TestGetVSCodeSettings(t *testing.T) {
	data, err := GetVSCodeSettings()
	if err != nil {
		t.Fatalf("Failed to get VSCode settings: %v", err)
	}
	if len(data) == 0 {
		t.Error("VSCode settings data is empty")
	}
}

func TestListCommands(t *testing.T) {
	commands, err := ListCommands()
	if err != nil {
		t.Fatalf("Failed to list commands: %v", err)
	}

	expectedCommands := []string{
		"analyze", "checklist", "clarify", "constitution",
		"implement", "plan", "spec", "tasks",
	}

	if len(commands) != len(expectedCommands) {
		t.Errorf("Expected %d commands, got %d", len(expectedCommands), len(commands))
	}

	// Create a map for easy lookup
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd] = true
	}

	// Check all expected commands are present
	for _, expected := range expectedCommands {
		if !commandMap[expected] {
			t.Errorf("Expected command %s not found in list", expected)
		}
	}
}

func TestListTemplates(t *testing.T) {
	templates, err := ListTemplates()
	if err != nil {
		t.Fatalf("Failed to list templates: %v", err)
	}

	expectedTemplates := []string{
		"agent-file-template.md",
		"checklist-template.md",
		"plan-template.md",
		"spec-template.md",
		"tasks-template.md",
		"vscode-settings.json",
	}

	if len(templates) != len(expectedTemplates) {
		t.Errorf("Expected %d templates, got %d", len(expectedTemplates), len(templates))
	}

	// Create a map for easy lookup
	templateMap := make(map[string]bool)
	for _, tmpl := range templates {
		templateMap[tmpl] = true
	}

	// Check all expected templates are present
	for _, expected := range expectedTemplates {
		if !templateMap[expected] {
			t.Errorf("Expected template %s not found in list", expected)
		}
	}
}
