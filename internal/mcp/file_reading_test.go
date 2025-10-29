package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFeatureFile(t *testing.T) {
	// Create temporary workspace structure
	tmpDir := t.TempDir()
	workspaceRoot := filepath.Join(tmpDir, "project")
	featureDir := filepath.Join(workspaceRoot, "specs", "user-auth")

	if err := os.MkdirAll(featureDir, 0755); err != nil {
		t.Fatalf("Failed to create feature dir: %v", err)
	}

	// Create test files
	specContent := `# User Authentication Specification

## Overview
This feature implements user authentication using JWT tokens.
`
	planContent := `# Implementation Plan

## Phase 1
- Create login endpoint
- Implement JWT generation
`

	specPath := filepath.Join(featureDir, "spec.md")
	planPath := filepath.Join(featureDir, "plan.md")

	if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
		t.Fatalf("Failed to write spec.md: %v", err)
	}
	if err := os.WriteFile(planPath, []byte(planContent), 0644); err != nil {
		t.Fatalf("Failed to write plan.md: %v", err)
	}

	tests := []struct {
		name          string
		workspaceRoot string
		featureName   string
		filename      string
		want          string
		wantEmpty     bool
	}{
		{
			name:          "Read existing spec.md",
			workspaceRoot: workspaceRoot,
			featureName:   "user-auth",
			filename:      "spec.md",
			want:          "User Authentication Specification",
			wantEmpty:     false,
		},
		{
			name:          "Read existing plan.md",
			workspaceRoot: workspaceRoot,
			featureName:   "user-auth",
			filename:      "plan.md",
			want:          "Implementation Plan",
			wantEmpty:     false,
		},
		{
			name:          "Read non-existent file",
			workspaceRoot: workspaceRoot,
			featureName:   "user-auth",
			filename:      "nonexistent.md",
			want:          "",
			wantEmpty:     true,
		},
		{
			name:          "Empty workspace root",
			workspaceRoot: "",
			featureName:   "user-auth",
			filename:      "spec.md",
			want:          "",
			wantEmpty:     true,
		},
		{
			name:          "Empty feature name",
			workspaceRoot: workspaceRoot,
			featureName:   "",
			filename:      "spec.md",
			want:          "",
			wantEmpty:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReadFeatureFile(tt.workspaceRoot, tt.featureName, tt.filename)

			if tt.wantEmpty {
				if got != "" {
					t.Errorf("ReadFeatureFile() = %q, want empty string", got)
				}
			} else {
				if !strings.Contains(got, tt.want) {
					t.Errorf("ReadFeatureFile() does not contain %q, got:\n%s", tt.want, got)
				}
			}
		})
	}
}

func TestProcessTemplateWithContext(t *testing.T) {
	// Create temporary workspace structure
	tmpDir := t.TempDir()
	workspaceRoot := filepath.Join(tmpDir, "project")
	featureDir := filepath.Join(workspaceRoot, "specs", "api-feature")

	if err := os.MkdirAll(featureDir, 0755); err != nil {
		t.Fatalf("Failed to create feature dir: %v", err)
	}

	// Create feature files
	specContent := "# API Feature Spec\n\nDetailed specification here."
	planContent := "# API Feature Plan\n\nImplementation steps here."
	tasksContent := "# API Feature Tasks\n\n- [ ] Task 1\n- [ ] Task 2"

	files := map[string]string{
		"spec.md":  specContent,
		"plan.md":  planContent,
		"tasks.md": tasksContent,
	}

	for filename, content := range files {
		filePath := filepath.Join(featureDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", filename, err)
		}
	}

	tests := []struct {
		name          string
		template      string
		data          TemplateData
		shouldContain []string
	}{
		{
			name: "Read spec file",
			template: `User input: {{.Arguments}}

Previous Spec:
{{readSpec}}`,
			data: TemplateData{
				Arguments:     "Add new endpoint",
				WorkspaceRoot: workspaceRoot,
				FeatureName:   "api-feature",
			},
			shouldContain: []string{
				"User input: Add new endpoint",
				"API Feature Spec",
				"Detailed specification here",
			},
		},
		{
			name: "Read plan file",
			template: `{{if readPlan}}
Existing Plan:
{{readPlan}}
{{else}}
No previous plan found.
{{end}}`,
			data: TemplateData{
				WorkspaceRoot: workspaceRoot,
				FeatureName:   "api-feature",
			},
			shouldContain: []string{
				"Existing Plan",
				"API Feature Plan",
				"Implementation steps here",
			},
		},
		{
			name: "Read tasks file",
			template: `Current Tasks:
{{readTasks}}`,
			data: TemplateData{
				WorkspaceRoot: workspaceRoot,
				FeatureName:   "api-feature",
			},
			shouldContain: []string{
				"Current Tasks",
				"API Feature Tasks",
				"Task 1",
				"Task 2",
			},
		},
		{
			name:     "Read custom file",
			template: `{{readFile "spec.md"}}`,
			data: TemplateData{
				WorkspaceRoot: workspaceRoot,
				FeatureName:   "api-feature",
			},
			shouldContain: []string{
				"API Feature Spec",
			},
		},
		{
			name:     "No feature context",
			template: `{{if readSpec}}Has spec{{else}}No spec{{end}}`,
			data: TemplateData{
				WorkspaceRoot: workspaceRoot,
				FeatureName:   "", // No feature
			},
			shouldContain: []string{
				"No spec",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProcessTemplateWithContext(tt.template, tt.data)
			if err != nil {
				t.Fatalf("ProcessTemplateWithContext failed: %v", err)
			}

			for _, expected := range tt.shouldContain {
				if !strings.Contains(result, expected) {
					t.Errorf("Result should contain %q but doesn't:\n%s", expected, result)
				}
			}
		})
	}
}

func TestTemplateFunctionIntegration(t *testing.T) {
	// Create temporary workspace
	tmpDir := t.TempDir()
	workspaceRoot := filepath.Join(tmpDir, "project")
	featureDir := filepath.Join(workspaceRoot, "specs", "test-feature")

	if err := os.MkdirAll(featureDir, 0755); err != nil {
		t.Fatalf("Failed to create feature dir: %v", err)
	}

	// Create spec file
	specContent := `# Test Feature

**Status**: In Progress
`
	specPath := filepath.Join(featureDir, "spec.md")
	if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
		t.Fatalf("Failed to write spec.md: %v", err)
	}

	// Template that combines multiple functions
	templateContent := `# {{upper .CommandName}} Command

Project: {{.ProjectName}}
Feature: {{.FeatureName}}

{{if .Arguments}}
User Guidance: {{trim .Arguments}}
{{end}}

{{if readSpec}}
## Existing Specification

{{readSpec}}

Use this spec as context for the current task.
{{else}}
No existing specification found.
{{end}}`

	data := TemplateData{
		CommandName:   "implement",
		ProjectName:   "MyProject",
		FeatureName:   "test-feature",
		Arguments:     "  Add unit tests  ",
		WorkspaceRoot: workspaceRoot,
	}

	result, err := ProcessTemplateWithContext(templateContent, data)
	if err != nil {
		t.Fatalf("ProcessTemplateWithContext failed: %v", err)
	}

	// Verify combined functionality
	checks := []string{
		"# IMPLEMENT Command",           // upper function
		"Project: MyProject",            // metadata
		"Feature: test-feature",         // metadata
		"User Guidance: Add unit tests", // trim function
		"## Existing Specification",     // conditional
		"# Test Feature",                // readSpec function
		"**Status**: In Progress",       // readSpec content
	}

	for _, expected := range checks {
		if !strings.Contains(result, expected) {
			t.Errorf("Result should contain %q but doesn't:\n%s", expected, result)
		}
	}
}
