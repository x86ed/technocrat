package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePlanData(t *testing.T) {
	// Create a temporary plan file for testing
	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, "plan.md")

	content := `# Feature Plan

## Metadata
**Language/Version**: Go 1.21
**Primary Dependencies**: Cobra + Viper
**Storage**: PostgreSQL
**Project Type**: CLI Tool

## Description
Test feature
`

	if err := os.WriteFile(planPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test plan file: %v", err)
	}

	// Test parsing
	data, err := parsePlanData(planPath)
	if err != nil {
		t.Fatalf("parsePlanData failed: %v", err)
	}

	// Verify extracted data
	if data.Language != "Go 1.21" {
		t.Errorf("Expected Language 'Go 1.21', got '%s'", data.Language)
	}

	if data.Framework != "Cobra + Viper" {
		t.Errorf("Expected Framework 'Cobra + Viper', got '%s'", data.Framework)
	}

	if data.Database != "PostgreSQL" {
		t.Errorf("Expected Database 'PostgreSQL', got '%s'", data.Database)
	}

	if data.ProjectType != "CLI Tool" {
		t.Errorf("Expected ProjectType 'CLI Tool', got '%s'", data.ProjectType)
	}
}

func TestParsePlanDataWithNeedsClairification(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, "plan.md")

	content := `# Feature Plan

**Language/Version**: NEEDS CLARIFICATION
**Primary Dependencies**: N/A
**Storage**: NEEDS CLARIFICATION
**Project Type**: Web Application
`

	if err := os.WriteFile(planPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test plan file: %v", err)
	}

	data, err := parsePlanData(planPath)
	if err != nil {
		t.Fatalf("parsePlanData failed: %v", err)
	}

	// Should ignore NEEDS CLARIFICATION and N/A values
	if data.Language != "" {
		t.Errorf("Expected empty Language, got '%s'", data.Language)
	}

	if data.Framework != "" {
		t.Errorf("Expected empty Framework, got '%s'", data.Framework)
	}

	if data.Database != "" {
		t.Errorf("Expected empty Database, got '%s'", data.Database)
	}

	if data.ProjectType != "Web Application" {
		t.Errorf("Expected ProjectType 'Web Application', got '%s'", data.ProjectType)
	}
}

func TestFormatTechnologyStack(t *testing.T) {
	tests := []struct {
		name      string
		lang      string
		framework string
		expected  string
	}{
		{
			name:      "both present",
			lang:      "Go 1.21",
			framework: "Cobra",
			expected:  "Go 1.21 + Cobra",
		},
		{
			name:      "only language",
			lang:      "Python 3.11",
			framework: "",
			expected:  "Python 3.11",
		},
		{
			name:      "only framework",
			lang:      "",
			framework: "React",
			expected:  "React",
		},
		{
			name:      "both empty",
			lang:      "",
			framework: "",
			expected:  "",
		},
		{
			name:      "needs clarification ignored",
			lang:      "NEEDS CLARIFICATION",
			framework: "FastAPI",
			expected:  "FastAPI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTechnologyStack(tt.lang, tt.framework)
			if result != tt.expected {
				t.Errorf("formatTechnologyStack(%q, %q) = %q, want %q",
					tt.lang, tt.framework, result, tt.expected)
			}
		})
	}
}

func TestGetCommandsForLanguage(t *testing.T) {
	tests := []struct {
		lang     string
		expected string
	}{
		{"Python 3.11", "cd src && pytest && ruff check ."},
		{"Rust 1.70", "cargo test && cargo clippy"},
		{"JavaScript", "npm test && npm run lint"},
		{"TypeScript", "npm test && npm run lint"},
		{"Go 1.21", "go test ./... && go vet ./..."},
		{"Ruby", "# Add commands for Ruby"},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			result := getCommandsForLanguage(tt.lang)
			if result != tt.expected {
				t.Errorf("getCommandsForLanguage(%q) = %q, want %q",
					tt.lang, result, tt.expected)
			}
		})
	}
}

func TestGetProjectStructure(t *testing.T) {
	tests := []struct {
		projectType string
		expected    string
	}{
		{"Web Application", "backend/\nfrontend/\ntests/"},
		{"web service", "backend/\nfrontend/\ntests/"},
		{"CLI Tool", "src/\ntests/"},
		{"Library", "src/\ntests/"},
		{"", "src/\ntests/"},
	}

	for _, tt := range tests {
		t.Run(tt.projectType, func(t *testing.T) {
			result := getProjectStructure(tt.projectType)
			if result != tt.expected {
				t.Errorf("getProjectStructure(%q) = %q, want %q",
					tt.projectType, result, tt.expected)
			}
		})
	}
}

func TestGetAgentFileConfig(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		agent        AgentType
		expectedName string
		expectedPath string
	}{
		{AgentClaude, "Claude Code", "CLAUDE.md"},
		{AgentGemini, "Gemini CLI", "GEMINI.md"},
		{AgentCopilot, "GitHub Copilot", ".github/copilot-instructions.md"},
		{AgentCursor, "Cursor IDE", ".cursor/rules/tchncrt-rules.mdc"},
		{AgentWindsurf, "Windsurf", ".windsurf/rules/tchncrt-rules.md"},
	}

	for _, tt := range tests {
		t.Run(string(tt.agent), func(t *testing.T) {
			config := getAgentFileConfig(tmpDir, tt.agent)
			if config.Name != tt.expectedName {
				t.Errorf("Expected name %q, got %q", tt.expectedName, config.Name)
			}

			expectedFullPath := filepath.Join(tmpDir, tt.expectedPath)
			if config.Path != expectedFullPath {
				t.Errorf("Expected path %q, got %q", expectedFullPath, config.Path)
			}
		})
	}
}

func TestGetLanguageConventions(t *testing.T) {
	tests := []struct {
		lang     string
		expected string
	}{
		{"Go", "Go: Follow standard conventions"},
		{"Python", "Python: Follow standard conventions"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			result := getLanguageConventions(tt.lang)
			if result != tt.expected {
				t.Errorf("getLanguageConventions(%q) = %q, want %q",
					tt.lang, result, tt.expected)
			}
		})
	}
}
