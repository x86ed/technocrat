package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectWorkspaceContext(t *testing.T) {
	// Create temporary directory structure for testing
	tmpDir := t.TempDir()
	
	// Create a mock project structure
	projectRoot := filepath.Join(tmpDir, "test-project")
	memoryDir := filepath.Join(projectRoot, "memory")
	specsDir := filepath.Join(projectRoot, "specs")
	featureDir := filepath.Join(specsDir, "user-authentication")
	
	if err := os.MkdirAll(memoryDir, 0755); err != nil {
		t.Fatalf("Failed to create memory dir: %v", err)
	}
	if err := os.MkdirAll(featureDir, 0755); err != nil {
		t.Fatalf("Failed to create feature dir: %v", err)
	}
	
	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	tests := []struct {
		name            string
		workingDir      string
		expectedRoot    string
		expectedProject string
		expectedFeature string
	}{
		{
			name:            "At project root",
			workingDir:      projectRoot,
			expectedRoot:    projectRoot,
			expectedProject: "test-project",
			expectedFeature: "",
		},
		{
			name:            "In feature directory",
			workingDir:      featureDir,
			expectedRoot:    projectRoot,
			expectedProject: "test-project",
			expectedFeature: "user-authentication",
		},
		{
			name:            "In specs directory",
			workingDir:      specsDir,
			expectedRoot:    projectRoot,
			expectedProject: "test-project",
			expectedFeature: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Change to test directory
			if err := os.Chdir(tt.workingDir); err != nil {
				t.Fatalf("Failed to change directory: %v", err)
			}
			
			ctx := DetectWorkspaceContext()
			
			// Resolve symlinks for macOS (/private/var vs /var)
			gotRoot, _ := filepath.EvalSymlinks(ctx.Root)
			wantRoot, _ := filepath.EvalSymlinks(tt.expectedRoot)
			
			if gotRoot != wantRoot {
				t.Errorf("Root = %q, want %q", gotRoot, wantRoot)
			}
			if ctx.ProjectName != tt.expectedProject {
				t.Errorf("ProjectName = %q, want %q", ctx.ProjectName, tt.expectedProject)
			}
			if ctx.FeatureName != tt.expectedFeature {
				t.Errorf("FeatureName = %q, want %q", ctx.FeatureName, tt.expectedFeature)
			}
		})
	}
}

func TestFindWorkspaceRoot(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "project")
	nestedDir := filepath.Join(projectRoot, "src", "internal", "pkg")
	
	// Create directory structure with memory/ marker
	memoryDir := filepath.Join(projectRoot, "memory")
	if err := os.MkdirAll(memoryDir, 0755); err != nil {
		t.Fatalf("Failed to create memory dir: %v", err)
	}
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested dir: %v", err)
	}
	
	tests := []struct {
		name     string
		startDir string
		want     string
	}{
		{
			name:     "From project root",
			startDir: projectRoot,
			want:     projectRoot,
		},
		{
			name:     "From deeply nested directory",
			startDir: nestedDir,
			want:     projectRoot,
		},
		{
			name:     "Outside project (no marker)",
			startDir: tmpDir,
			want:     "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findWorkspaceRoot(tt.startDir)
			if got != tt.want {
				t.Errorf("findWorkspaceRoot(%q) = %q, want %q", tt.startDir, got, tt.want)
			}
		})
	}
}

func TestExtractFeatureName(t *testing.T) {
	tests := []struct {
		name          string
		cwd           string
		workspaceRoot string
		want          string
	}{
		{
			name:          "In feature directory",
			cwd:           "/project/specs/user-auth",
			workspaceRoot: "/project",
			want:          "user-auth",
		},
		{
			name:          "In nested feature directory",
			cwd:           "/project/specs/user-auth/docs",
			workspaceRoot: "/project",
			want:          "user-auth",
		},
		{
			name:          "Not in specs",
			cwd:           "/project/src/main",
			workspaceRoot: "/project",
			want:          "",
		},
		{
			name:          "At specs root",
			cwd:           "/project/specs",
			workspaceRoot: "/project",
			want:          "",
		},
		{
			name:          "Empty workspace root",
			cwd:           "/project/specs/feature",
			workspaceRoot: "",
			want:          "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFeatureName(tt.cwd, tt.workspaceRoot)
			if got != tt.want {
				t.Errorf("extractFeatureName(%q, %q) = %q, want %q", 
					tt.cwd, tt.workspaceRoot, got, tt.want)
			}
		})
	}
}

func TestGetProjectNameFromConstitution(t *testing.T) {
	tmpDir := t.TempDir()
	
	tests := []struct {
		name              string
		constitutionContent string
		want              string
	}{
		{
			name: "Project name as first heading",
			constitutionContent: `# My Awesome Project

This is a great project.
`,
			want: "My Awesome Project",
		},
		{
			name: "Project name in dedicated section",
			constitutionContent: `# Constitution

## Project Name

TechnoSync Platform

## Description
...
`,
			want: "TechnoSync Platform",
		},
		{
			name: "No project name found",
			constitutionContent: `# About

This is some content.
`,
			want: "",
		},
		{
			name:              "Empty constitution",
			constitutionContent: "",
			want:              "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary workspace
			workspaceRoot := filepath.Join(tmpDir, tt.name)
			memoryDir := filepath.Join(workspaceRoot, "memory")
			if err := os.MkdirAll(memoryDir, 0755); err != nil {
				t.Fatalf("Failed to create memory dir: %v", err)
			}
			
			// Write constitution file
			constitutionPath := filepath.Join(memoryDir, "constitution.md")
			if err := os.WriteFile(constitutionPath, []byte(tt.constitutionContent), 0644); err != nil {
				t.Fatalf("Failed to write constitution: %v", err)
			}
			
			got := getProjectNameFromConstitution(workspaceRoot)
			if got != tt.want {
				t.Errorf("getProjectNameFromConstitution() = %q, want %q", got, tt.want)
			}
		})
	}
	
	t.Run("No constitution file", func(t *testing.T) {
		workspaceRoot := filepath.Join(tmpDir, "no-constitution")
		if err := os.MkdirAll(workspaceRoot, 0755); err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}
		
		got := getProjectNameFromConstitution(workspaceRoot)
		if got != "" {
			t.Errorf("getProjectNameFromConstitution() = %q, want empty string", got)
		}
	})
}
