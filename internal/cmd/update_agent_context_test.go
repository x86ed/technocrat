package cmd

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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

func TestCreateNewAgentFile(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	repoRoot := tmpDir

	// Create .tchncrt/templates directory and template file
	templateDir := filepath.Join(repoRoot, ".tchncrt", "templates")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("Failed to create template directory: %v", err)
	}

	templateContent := `# [PROJECT NAME] Development Guidelines

Auto-generated from all feature plans. Last updated: [DATE]

## Active Technologies

[EXTRACTED FROM ALL PLAN.MD FILES]

## Project Structure

` + "```" + `
[ACTUAL STRUCTURE FROM PLANS]
` + "```" + `

## Commands

[ONLY COMMANDS FOR ACTIVE TECHNOLOGIES]

## Code Style

[LANGUAGE-SPECIFIC, ONLY FOR LANGUAGES IN USE]

## Recent Changes

[LAST 3 FEATURES AND WHAT THEY ADDED]
`

	templatePath := filepath.Join(templateDir, "agent-file-template.md")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	tests := []struct {
		name        string
		config      AgentFileConfig
		paths       *FeaturePaths
		planData    *PlanData
		wantErr     bool
		checkOutput func(*testing.T, string)
	}{
		{
			name: "create with full plan data",
			config: AgentFileConfig{
				Path: filepath.Join(tmpDir, "CLAUDE.md"),
				Name: "Claude Code",
			},
			paths: &FeaturePaths{
				RepoRoot:      repoRoot,
				CurrentBranch: "001-test-feature",
				HasGit:        true,
			},
			planData: &PlanData{
				Language:    "Go 1.21",
				Framework:   "Cobra",
				Database:    "PostgreSQL",
				ProjectType: "CLI Tool",
			},
			wantErr: false,
			checkOutput: func(t *testing.T, content string) {
				if !strings.Contains(content, "Go 1.21 + Cobra") {
					t.Error("Expected tech stack not found in output")
				}
				if !strings.Contains(content, "001-test-feature") {
					t.Error("Expected branch name not found")
				}
				if !strings.Contains(content, "src/") {
					t.Error("Expected project structure not found")
				}
			},
		},
		{
			name: "create with minimal plan data",
			config: AgentFileConfig{
				Path: filepath.Join(tmpDir, "GEMINI.md"),
				Name: "Gemini CLI",
			},
			paths: &FeaturePaths{
				RepoRoot:      repoRoot,
				CurrentBranch: "002-minimal",
				HasGit:        false,
			},
			planData: &PlanData{},
			wantErr:  false,
			checkOutput: func(t *testing.T, content string) {
				if !strings.Contains(content, "002-minimal") {
					t.Error("Expected branch name not found")
				}
			},
		},
		{
			name: "fail with missing template",
			config: AgentFileConfig{
				Path: filepath.Join(t.TempDir(), "FAIL.md"),
				Name: "Fail Test",
			},
			paths: &FeaturePaths{
				RepoRoot:      t.TempDir(), // Different root without template
				CurrentBranch: "003-fail",
				HasGit:        true,
			},
			planData: &PlanData{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := createNewAgentFile(tt.config, tt.paths, tt.planData)
			if (err != nil) != tt.wantErr {
				t.Errorf("createNewAgentFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkOutput != nil {
				content, err := os.ReadFile(tt.config.Path)
				if err != nil {
					t.Fatalf("Failed to read created file: %v", err)
				}
				tt.checkOutput(t, string(content))
			}
		})
	}
}

func TestUpdateExistingAgentFile(t *testing.T) {
	tests := []struct {
		name         string
		existingFile string
		config       AgentFileConfig
		paths        *FeaturePaths
		planData     *PlanData
		wantErr      bool
		checkOutput  func(*testing.T, string)
	}{
		{
			name: "update with new tech stack",
			existingFile: `# Project Guidelines

**Last updated**: 2024-01-01

## Active Technologies

- Python 3.11 (001-old-feature)

## Recent Changes

- 001-old-feature: Added Python 3.11
`,
			config: AgentFileConfig{
				Path: filepath.Join(t.TempDir(), "CLAUDE.md"),
				Name: "Claude Code",
			},
			paths: &FeaturePaths{
				CurrentBranch: "002-new-feature",
				HasGit:        true,
			},
			planData: &PlanData{
				Language:  "Go 1.21",
				Framework: "Cobra",
			},
			wantErr: false,
			checkOutput: func(t *testing.T, content string) {
				if !strings.Contains(content, "Go 1.21 + Cobra") {
					t.Error("Expected new tech stack not added")
				}
				if !strings.Contains(content, "Python 3.11") {
					t.Error("Expected old tech stack removed")
				}
				// The Recent Changes section should contain the new entry
				// Let's check that we have a changes section and the branch is mentioned
				hasChangesSection := strings.Contains(content, "## Recent Changes")
				hasBranchInChanges := strings.Contains(content, "002-new-feature")
				if !hasChangesSection {
					t.Error("Recent Changes section missing")
				}
				if !hasBranchInChanges {
					t.Error("New feature branch not mentioned in changes")
				}
				// Check date was updated
				if strings.Contains(content, "2024-01-01") {
					t.Error("Date should have been updated")
				}
			},
		},
		{
			name: "update preserves manual additions",
			existingFile: `# Project Guidelines

## Active Technologies

- Python 3.11 (001-feature)

## Recent Changes

- 001-feature: Added Python 3.11

<!-- MANUAL ADDITIONS START -->
Custom content here
<!-- MANUAL ADDITIONS END -->
`,
			config: AgentFileConfig{
				Path: filepath.Join(t.TempDir(), "CLAUDE2.md"),
				Name: "Claude Code",
			},
			paths: &FeaturePaths{
				CurrentBranch: "002-feature",
				HasGit:        true,
			},
			planData: &PlanData{
				Language: "Go 1.21",
			},
			wantErr: false,
			checkOutput: func(t *testing.T, content string) {
				if !strings.Contains(content, "Custom content here") {
					t.Error("Manual additions were not preserved")
				}
			},
		},
		{
			name: "limits recent changes to 3 entries",
			existingFile: `# Project Guidelines

## Active Technologies

- Python 3.11 (001-feature)

## Recent Changes

- 001-feature: Added Python 3.11
- 002-feature: Added Django
- 003-feature: Added PostgreSQL
`,
			config: AgentFileConfig{
				Path: filepath.Join(t.TempDir(), "CLAUDE3.md"),
				Name: "Claude Code",
			},
			paths: &FeaturePaths{
				CurrentBranch: "004-feature",
				HasGit:        true,
			},
			planData: &PlanData{
				Language: "Go 1.21",
			},
			wantErr: false,
			checkOutput: func(t *testing.T, content string) {
				lines := strings.Split(content, "\n")
				changeCount := 0
				inChanges := false
				for _, line := range lines {
					if strings.HasPrefix(line, "## Recent Changes") {
						inChanges = true
						continue
					}
					if inChanges && strings.HasPrefix(line, "## ") {
						break
					}
					if inChanges && strings.HasPrefix(line, "- ") {
						changeCount++
					}
				}
				if changeCount > 3 {
					t.Errorf("Expected max 3 recent changes, got %d", changeCount)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write existing file
			if err := os.WriteFile(tt.config.Path, []byte(tt.existingFile), 0644); err != nil {
				t.Fatalf("Failed to create existing file: %v", err)
			}

			err := updateExistingAgentFile(tt.config, tt.paths, tt.planData)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateExistingAgentFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkOutput != nil {
				content, err := os.ReadFile(tt.config.Path)
				if err != nil {
					t.Fatalf("Failed to read updated file: %v", err)
				}
				tt.checkOutput(t, string(content))
			}
		})
	}
}

func TestUpdateAgentFile(t *testing.T) {
	tmpDir := t.TempDir()
	repoRoot := tmpDir

	// Create template
	templateDir := filepath.Join(repoRoot, ".tchncrt", "templates")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("Failed to create template directory: %v", err)
	}

	templateContent := `# [PROJECT NAME] Development Guidelines

## Active Technologies

[EXTRACTED FROM ALL PLAN.MD FILES]

## Recent Changes

[LAST 3 FEATURES AND WHAT THEY ADDED]
`

	templatePath := filepath.Join(templateDir, "agent-file-template.md")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	t.Run("creates new file when not exists", func(t *testing.T) {
		config := AgentFileConfig{
			Path: filepath.Join(tmpDir, "NEW.md"),
			Name: "New Agent",
		}
		paths := &FeaturePaths{
			RepoRoot:      repoRoot,
			CurrentBranch: "001-test",
			HasGit:        true,
		}
		planData := &PlanData{
			Language: "Go 1.21",
		}

		err := updateAgentFile(config, paths, planData)
		if err != nil {
			t.Errorf("updateAgentFile() error = %v", err)
		}

		if _, err := os.Stat(config.Path); os.IsNotExist(err) {
			t.Error("File was not created")
		}
	})

	t.Run("updates existing file", func(t *testing.T) {
		existingPath := filepath.Join(tmpDir, "EXISTING.md")
		existingContent := `# Project

## Active Technologies

- Old Tech

## Recent Changes

- old: stuff
`
		if err := os.WriteFile(existingPath, []byte(existingContent), 0644); err != nil {
			t.Fatalf("Failed to create existing file: %v", err)
		}

		config := AgentFileConfig{
			Path: existingPath,
			Name: "Existing Agent",
		}
		paths := &FeaturePaths{
			RepoRoot:      repoRoot,
			CurrentBranch: "002-test",
			HasGit:        true,
		}
		planData := &PlanData{
			Language: "Python 3.11",
		}

		err := updateAgentFile(config, paths, planData)
		if err != nil {
			t.Errorf("updateAgentFile() error = %v", err)
		}

		content, _ := os.ReadFile(existingPath)
		if !strings.Contains(string(content), "Python 3.11") {
			t.Error("File was not updated with new data")
		}
	})
}

func TestUpdateSpecificAgent(t *testing.T) {
	tmpDir := t.TempDir()
	repoRoot := tmpDir

	// Create template
	templateDir := filepath.Join(repoRoot, ".tchncrt", "templates")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("Failed to create template directory: %v", err)
	}

	templateContent := `# [PROJECT NAME]
## Active Technologies
[EXTRACTED FROM ALL PLAN.MD FILES]
`
	templatePath := filepath.Join(templateDir, "agent-file-template.md")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	paths := &FeaturePaths{
		RepoRoot:      repoRoot,
		CurrentBranch: "001-test",
		HasGit:        true,
	}
	planData := &PlanData{
		Language: "Go 1.21",
	}

	err := updateSpecificAgent(paths, planData, AgentClaude)
	if err != nil {
		t.Errorf("updateSpecificAgent() error = %v", err)
	}

	// Verify file was created
	claudePath := filepath.Join(repoRoot, "CLAUDE.md")
	if _, err := os.Stat(claudePath); os.IsNotExist(err) {
		t.Error("Claude file was not created")
	}
}

func TestUpdateAllExistingAgents(t *testing.T) {
	tmpDir := t.TempDir()
	repoRoot := tmpDir

	// Create template
	templateDir := filepath.Join(repoRoot, ".tchncrt", "templates")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("Failed to create template directory: %v", err)
	}

	templateContent := `# [PROJECT NAME]
## Active Technologies
[EXTRACTED FROM ALL PLAN.MD FILES]
`
	templatePath := filepath.Join(templateDir, "agent-file-template.md")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	t.Run("creates default Claude file when no agents exist", func(t *testing.T) {
		paths := &FeaturePaths{
			RepoRoot:      repoRoot,
			CurrentBranch: "001-test",
			HasGit:        true,
		}
		planData := &PlanData{
			Language: "Go 1.21",
		}

		err := updateAllExistingAgents(paths, planData)
		if err != nil {
			t.Errorf("updateAllExistingAgents() error = %v", err)
		}

		// Verify Claude file was created
		claudePath := filepath.Join(repoRoot, "CLAUDE.md")
		if _, err := os.Stat(claudePath); os.IsNotExist(err) {
			t.Error("Default Claude file was not created")
		}
	})

	t.Run("updates existing agent files", func(t *testing.T) {
		tmpDir2 := t.TempDir()
		repoRoot2 := tmpDir2

		// Create template in new dir
		templateDir2 := filepath.Join(repoRoot2, ".tchncrt", "templates")
		os.MkdirAll(templateDir2, 0755)
		templatePath2 := filepath.Join(templateDir2, "agent-file-template.md")
		os.WriteFile(templatePath2, []byte(templateContent), 0644)

		// Create existing agent files
		geminiPath := filepath.Join(repoRoot2, "GEMINI.md")
		geminiContent := `# Project
## Active Technologies
- Old
`
		if err := os.WriteFile(geminiPath, []byte(geminiContent), 0644); err != nil {
			t.Fatalf("Failed to create Gemini file: %v", err)
		}

		paths := &FeaturePaths{
			RepoRoot:      repoRoot2,
			CurrentBranch: "002-update",
			HasGit:        true,
		}
		planData := &PlanData{
			Language: "Python 3.11",
		}

		err := updateAllExistingAgents(paths, planData)
		if err != nil {
			t.Errorf("updateAllExistingAgents() error = %v", err)
		}

		// Verify Gemini file was updated
		content, _ := os.ReadFile(geminiPath)
		if !strings.Contains(string(content), "Python 3.11") {
			t.Error("Existing agent file was not updated")
		}
	})

	t.Run("handles shared paths correctly", func(t *testing.T) {
		tmpDir3 := t.TempDir()
		repoRoot3 := tmpDir3

		// Create template in new dir
		templateDir3 := filepath.Join(repoRoot3, ".tchncrt", "templates")
		os.MkdirAll(templateDir3, 0755)
		templatePath3 := filepath.Join(templateDir3, "agent-file-template.md")
		os.WriteFile(templatePath3, []byte(templateContent), 0644)

		// Create AGENTS.md (shared by opencode, codex, q)
		agentsPath := filepath.Join(repoRoot3, "AGENTS.md")
		agentsContent := `# Agents
## Active Technologies
- Shared
`
		if err := os.WriteFile(agentsPath, []byte(agentsContent), 0644); err != nil {
			t.Fatalf("Failed to create AGENTS.md: %v", err)
		}

		paths := &FeaturePaths{
			RepoRoot:      repoRoot3,
			CurrentBranch: "003-shared",
			HasGit:        true,
		}
		planData := &PlanData{
			Language: "Rust 1.70",
		}

		err := updateAllExistingAgents(paths, planData)
		if err != nil {
			t.Errorf("updateAllExistingAgents() error = %v", err)
		}

		// Verify AGENTS.md was updated only once
		content, _ := os.ReadFile(agentsPath)
		rustCount := strings.Count(string(content), "Rust 1.70")
		if rustCount > 1 {
			t.Errorf("Shared file was updated multiple times: %d occurrences", rustCount)
		}
	})
}

func TestRunUpdateAgentContextIntegration(t *testing.T) {
	// Setup git repo
	tmpDir := t.TempDir()
	repoRoot := tmpDir

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = repoRoot
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoRoot
	cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoRoot
	cmd.Run()

	// Create initial commit
	readmePath := filepath.Join(repoRoot, "README.md")
	os.WriteFile(readmePath, []byte("# Test"), 0644)
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repoRoot
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = repoRoot
	cmd.Run()

	// Create feature branch
	cmd = exec.Command("git", "checkout", "-b", "001-test-feature")
	cmd.Dir = repoRoot
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create branch: %v", err)
	}

	// Create template
	templateDir := filepath.Join(repoRoot, ".tchncrt", "templates")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("Failed to create template dir: %v", err)
	}

	templateContent := `# [PROJECT NAME] Development Guidelines

Auto-generated from all feature plans. Last updated: [DATE]

## Active Technologies

[EXTRACTED FROM ALL PLAN.MD FILES]

## Project Structure

` + "```" + `
[ACTUAL STRUCTURE FROM PLANS]
` + "```" + `

## Commands

[ONLY COMMANDS FOR ACTIVE TECHNOLOGIES]

## Code Style

[LANGUAGE-SPECIFIC, ONLY FOR LANGUAGES IN USE]

## Recent Changes

[LAST 3 FEATURES AND WHAT THEY ADDED]
`
	templatePath := filepath.Join(templateDir, "agent-file-template.md")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Create specs directory and plan.md
	specsDir := filepath.Join(repoRoot, "specs", "001-test-feature")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("Failed to create specs dir: %v", err)
	}

	planContent := `# Implementation Plan

## Metadata
**Language/Version**: Go 1.21
**Primary Dependencies**: Cobra + Viper
**Storage**: PostgreSQL
**Project Type**: CLI Tool

## Overview
Test implementation
`
	planPath := filepath.Join(specsDir, "plan.md")
	if err := os.WriteFile(planPath, []byte(planContent), 0644); err != nil {
		t.Fatalf("Failed to create plan.md: %v", err)
	}

	t.Run("full command execution with no agent type", func(t *testing.T) {
		// Change to repo root
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(repoRoot)

		// Set environment variable
		t.Setenv("TCHNCRT_FEATURE", "001-test-feature")

		// Execute command directly (not through cobra's Execute which calls root)
		cmd := &cobra.Command{
			Use:  "update-agent-context",
			RunE: runUpdateAgentContext,
		}
		cmd.SetArgs([]string{})

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}

		// Verify Claude file was created
		claudePath := filepath.Join(repoRoot, "CLAUDE.md")
		if _, err := os.Stat(claudePath); os.IsNotExist(err) {
			t.Error("Claude file was not created")
		} else {
			content, _ := os.ReadFile(claudePath)
			contentStr := string(content)
			if !strings.Contains(contentStr, "Go 1.21 + Cobra + Viper") {
				t.Error("Tech stack not found in Claude file")
			}
			if !strings.Contains(contentStr, "001-test-feature") {
				t.Error("Branch name not found in Claude file")
			}
		}
	})

	t.Run("full command execution with specific agent type", func(t *testing.T) {
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(repoRoot)

		t.Setenv("TCHNCRT_FEATURE", "001-test-feature")

		// Execute command for Gemini
		cmd := &cobra.Command{
			Use:   "update-agent-context",
			RunE:  runUpdateAgentContext,
		}
		cmd.SetArgs([]string{"gemini"})

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}

		// Verify Gemini file was created
		geminiPath := filepath.Join(repoRoot, "GEMINI.md")
		if _, err := os.Stat(geminiPath); os.IsNotExist(err) {
			t.Error("Gemini file was not created")
		} else {
			content, _ := os.ReadFile(geminiPath)
			if !strings.Contains(string(content), "Go 1.21 + Cobra + Viper") {
				t.Error("Tech stack not found in Gemini file")
			}
		}
	})

	t.Run("error when plan.md missing", func(t *testing.T) {
		tmpDir2 := t.TempDir()
		
		// Initialize git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir2
		cmd.Run()
		
		cmd = exec.Command("git", "config", "user.email", "test@example.com")
		cmd.Dir = tmpDir2
		cmd.Run()
		cmd = exec.Command("git", "config", "user.name", "Test User")
		cmd.Dir = tmpDir2
		cmd.Run()

		// Create initial commit
		readme := filepath.Join(tmpDir2, "README.md")
		os.WriteFile(readme, []byte("# Test"), 0644)
		cmd = exec.Command("git", "add", ".")
		cmd.Dir = tmpDir2
		cmd.Run()
		cmd = exec.Command("git", "commit", "-m", "initial")
		cmd.Dir = tmpDir2
		cmd.Run()

		cmd = exec.Command("git", "checkout", "-b", "002-no-plan")
		cmd.Dir = tmpDir2
		cmd.Run()

		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tmpDir2)

		t.Setenv("TCHNCRT_FEATURE", "002-no-plan")

		// Execute command
		testCmd := &cobra.Command{
			Use:  "update-agent-context",
			RunE: runUpdateAgentContext,
		}
		testCmd.SetArgs([]string{})

		err := testCmd.Execute()
		if err == nil {
			t.Error("Expected error when plan.md is missing")
		}
		if !strings.Contains(err.Error(), "no plan.md found") {
			t.Errorf("Expected 'no plan.md found' error, got: %v", err)
		}
	})

	t.Run("error when template missing", func(t *testing.T) {
		tmpDir3 := t.TempDir()
		
		// Initialize git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir3
		cmd.Run()
		
		cmd = exec.Command("git", "config", "user.email", "test@example.com")
		cmd.Dir = tmpDir3
		cmd.Run()
		cmd = exec.Command("git", "config", "user.name", "Test User")
		cmd.Dir = tmpDir3
		cmd.Run()

		// Create initial commit
		readme := filepath.Join(tmpDir3, "README.md")
		os.WriteFile(readme, []byte("# Test"), 0644)
		cmd = exec.Command("git", "add", ".")
		cmd.Dir = tmpDir3
		cmd.Run()
		cmd = exec.Command("git", "commit", "-m", "initial")
		cmd.Dir = tmpDir3
		cmd.Run()

		cmd = exec.Command("git", "checkout", "-b", "003-no-template")
		cmd.Dir = tmpDir3
		cmd.Run()

		// Create specs and plan without template
		specsDir := filepath.Join(tmpDir3, "specs", "003-no-template")
		os.MkdirAll(specsDir, 0755)
		planPath := filepath.Join(specsDir, "plan.md")
		os.WriteFile(planPath, []byte(planContent), 0644)

		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tmpDir3)

		t.Setenv("TCHNCRT_FEATURE", "003-no-template")

		testCmd := &cobra.Command{
			Use:  "update-agent-context",
			RunE: runUpdateAgentContext,
		}
		testCmd.SetArgs([]string{})

		err := testCmd.Execute()
		if err == nil {
			t.Error("Expected error when template is missing")
		}
		if !strings.Contains(err.Error(), "template not found") {
			t.Errorf("Expected 'template not found' error, got: %v", err)
		}
	})
}

func TestPrintUpdateSummary(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	planData := &PlanData{
		Language:  "Go 1.21",
		Framework: "Cobra",
		Database:  "PostgreSQL",
	}

	printUpdateSummary(planData)

	w.Close()
	os.Stderr = oldStderr

	var buf strings.Builder
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Summary of changes") {
		t.Error("Summary header not found in output")
	}
	if !strings.Contains(output, "Go 1.21") {
		t.Error("Language not found in summary")
	}
	if !strings.Contains(output, "Cobra") {
		t.Error("Framework not found in summary")
	}
	if !strings.Contains(output, "PostgreSQL") {
		t.Error("Database not found in summary")
	}
	if !strings.Contains(output, "Usage: technocrat update-agent-context") {
		t.Error("Usage message not found in summary")
	}
}

func TestLogFunctions(t *testing.T) {
	tests := []struct {
		name     string
		logFunc  func(string, ...interface{})
		message  string
		expected string
	}{
		{"logInfo", logInfo, "test message", "INFO: test message"},
		{"logSuccess", logSuccess, "success", "✓ success"},
		{"logWarning", logWarning, "warning", "WARNING: warning"},
		{"logError", logError, "error", "ERROR: error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			tt.logFunc(tt.message)

			w.Close()
			os.Stderr = oldStderr

			var buf strings.Builder
			io.Copy(&buf, r)
			output := buf.String()

			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, output)
			}
		})
	}
}

func TestParsePlanDataWithIncompleteData(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, "plan.md")

	content := `# Feature Plan

**Language/Version**: Python 3.11
`

	if err := os.WriteFile(planPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test plan file: %v", err)
	}

	data, err := parsePlanData(planPath)
	if err != nil {
		t.Fatalf("parsePlanData failed: %v", err)
	}

	if data.Language != "Python 3.11" {
		t.Errorf("Expected Language 'Python 3.11', got '%s'", data.Language)
	}

	// These should be empty
	if data.Framework != "" {
		t.Errorf("Expected empty Framework, got '%s'", data.Framework)
	}
	if data.Database != "" {
		t.Errorf("Expected empty Database, got '%s'", data.Database)
	}
	if data.ProjectType != "" {
		t.Errorf("Expected empty ProjectType, got '%s'", data.ProjectType)
	}
}

func TestParsePlanDataErrors(t *testing.T) {
	t.Run("non-existent file", func(t *testing.T) {
		_, err := parsePlanData("/nonexistent/path/plan.md")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("unreadable file", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("Skipping test when running as root")
		}

		tmpDir := t.TempDir()
		planPath := filepath.Join(tmpDir, "plan.md")
		os.WriteFile(planPath, []byte("test"), 0000)
		defer os.Chmod(planPath, 0644)

		_, err := parsePlanData(planPath)
		if err == nil {
			t.Error("Expected error for unreadable file")
		}
	})
}
