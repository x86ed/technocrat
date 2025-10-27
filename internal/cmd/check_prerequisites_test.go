package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"technocrat/internal/tchncrt"

	"github.com/spf13/cobra"
)

// Helper function to create a simple test directory structure for runCheckPrerequisites
func createSimpleTestStructure(tmpDir string) error {
	// Initialize a real git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		return err
	}

	// Configure git user
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		return err
	}

	// Create an initial commit (needed before we can create branches)
	readmeFile := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readmeFile, []byte("# Test\n"), 0644); err != nil {
		return err
	}

	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		return err
	}

	// Create and checkout a feature branch
	cmd = exec.Command("git", "checkout", "-b", "feature/test-feature")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		return err
	}

	// Create .tchncrt feature structure
	featureDir := filepath.Join(tmpDir, ".tchncrt", "features", "test-feature")
	if err := os.MkdirAll(featureDir, 0755); err != nil {
		return err
	}

	// Create required directories
	dirs := []string{
		filepath.Join(tmpDir, "docs"),
		filepath.Join(tmpDir, "prompts"),
		filepath.Join(tmpDir, "specs"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Create plan.md in the feature directory (not in docs/)
	planContent := `# Implementation Plan

## Overview
Test feature

## Technology Stack
- Go
`
	if err := os.WriteFile(filepath.Join(featureDir, "plan.md"), []byte(planContent), 0644); err != nil {
		return err
	}

	// Also create docs/plan.md for compatibility
	return os.WriteFile(filepath.Join(tmpDir, "docs", "plan.md"), []byte(planContent), 0644)
}

// Helper function to create a temporary feature structure
func createTestFeatureStructure(t *testing.T, includeFiles map[string]bool) (*tchncrt.FeaturePaths, func()) {
	t.Helper()

	// Create temporary directory
	tmpDir := t.TempDir()

	// Create feature directory structure
	featureDir := filepath.Join(tmpDir, ".tchncrt", "features", "test-feature")
	if err := os.MkdirAll(featureDir, 0755); err != nil {
		t.Fatalf("Failed to create feature directory: %v", err)
	}

	// Create contracts subdirectory if needed
	contractsDir := filepath.Join(featureDir, "contracts")

	// Helper to create file
	createFile := func(path, content string) {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}

	// Create files based on includeFiles map
	if includeFiles["spec.md"] {
		createFile(filepath.Join(featureDir, "spec.md"), "# Feature Spec\n")
	}
	if includeFiles["plan.md"] {
		createFile(filepath.Join(featureDir, "plan.md"), "# Implementation Plan\n")
	}
	if includeFiles["tasks.md"] {
		createFile(filepath.Join(featureDir, "tasks.md"), "# Tasks\n")
	}
	if includeFiles["research.md"] {
		createFile(filepath.Join(featureDir, "research.md"), "# Research\n")
	}
	if includeFiles["data-model.md"] {
		createFile(filepath.Join(featureDir, "data-model.md"), "# Data Model\n")
	}
	if includeFiles["quickstart.md"] {
		createFile(filepath.Join(featureDir, "quickstart.md"), "# Quickstart\n")
	}
	if includeFiles["contracts"] {
		if err := os.MkdirAll(contractsDir, 0755); err != nil {
			t.Fatalf("Failed to create contracts directory: %v", err)
		}
		createFile(filepath.Join(contractsDir, "api.yaml"), "openapi: 3.0.0\n")
	}

	// Build FeaturePaths
	paths := &tchncrt.FeaturePaths{
		RepoRoot:      tmpDir,
		CurrentBranch: "test-feature",
		HasGit:        false,
		FeatureDir:    featureDir,
		FeatureSpec:   filepath.Join(featureDir, "spec.md"),
		ImplPlan:      filepath.Join(featureDir, "plan.md"),
		Tasks:         filepath.Join(featureDir, "tasks.md"),
		Research:      filepath.Join(featureDir, "research.md"),
		DataModel:     filepath.Join(featureDir, "data-model.md"),
		ContractsDir:  contractsDir,
		Quickstart:    filepath.Join(featureDir, "quickstart.md"),
	}

	cleanup := func() {
		// t.TempDir() handles cleanup automatically
	}

	return paths, cleanup
}

func TestValidatePrerequisites(t *testing.T) {
	tests := []struct {
		name        string
		files       map[string]bool
		requireTask bool
		wantErr     bool
		errContains string
	}{
		{
			name: "all required files exist",
			files: map[string]bool{
				"plan.md": true,
			},
			requireTask: false,
			wantErr:     false,
		},
		{
			name: "with tasks.md when required",
			files: map[string]bool{
				"plan.md":  true,
				"tasks.md": true,
			},
			requireTask: true,
			wantErr:     false,
		},
		{
			name:        "missing feature directory",
			files:       map[string]bool{},
			requireTask: false,
			wantErr:     true,
			errContains: "feature directory not found",
		},
		{
			name:        "missing plan.md",
			files:       map[string]bool{},
			requireTask: false,
			wantErr:     true,
			errContains: "plan.md not found",
		},
		{
			name: "missing tasks.md when required",
			files: map[string]bool{
				"plan.md": true,
			},
			requireTask: true,
			wantErr:     true,
			errContains: "tasks.md not found",
		},
		{
			name: "tasks.md not required and missing",
			files: map[string]bool{
				"plan.md": true,
			},
			requireTask: false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global requireTasks flag
			originalRequireTasks := requireTasks
			requireTasks = tt.requireTask
			defer func() { requireTasks = originalRequireTasks }()

			paths, cleanup := createTestFeatureStructure(t, tt.files)
			defer cleanup()

			// For the "missing feature directory" test, use a non-existent path
			if tt.name == "missing feature directory" {
				paths.FeatureDir = filepath.Join(paths.RepoRoot, "nonexistent")
			}

			err := validatePrerequisites(paths)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validatePrerequisites() expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("validatePrerequisites() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("validatePrerequisites() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGetAvailableDocs(t *testing.T) {
	tests := []struct {
		name         string
		files        map[string]bool
		includeTasks bool
		want         []string
	}{
		{
			name:         "no optional docs",
			files:        map[string]bool{"plan.md": true},
			includeTasks: false,
			want:         []string{},
		},
		{
			name: "all optional docs present",
			files: map[string]bool{
				"plan.md":       true,
				"research.md":   true,
				"data-model.md": true,
				"contracts":     true,
				"quickstart.md": true,
			},
			includeTasks: false,
			want:         []string{"research.md", "data-model.md", "contracts/", "quickstart.md"},
		},
		{
			name: "some optional docs present",
			files: map[string]bool{
				"plan.md":     true,
				"research.md": true,
				"contracts":   true,
			},
			includeTasks: false,
			want:         []string{"research.md", "contracts/"},
		},
		{
			name: "tasks.md included when flag set",
			files: map[string]bool{
				"plan.md":  true,
				"tasks.md": true,
			},
			includeTasks: true,
			want:         []string{"tasks.md"},
		},
		{
			name: "tasks.md excluded when flag not set",
			files: map[string]bool{
				"plan.md":  true,
				"tasks.md": true,
			},
			includeTasks: false,
			want:         []string{},
		},
		{
			name: "tasks.md flag set but file missing",
			files: map[string]bool{
				"plan.md": true,
			},
			includeTasks: true,
			want:         []string{},
		},
		{
			name: "all docs including tasks",
			files: map[string]bool{
				"plan.md":       true,
				"tasks.md":      true,
				"research.md":   true,
				"data-model.md": true,
				"contracts":     true,
				"quickstart.md": true,
			},
			includeTasks: true,
			want:         []string{"research.md", "data-model.md", "contracts/", "quickstart.md", "tasks.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global includeTasks flag
			originalIncludeTasks := includeTasks
			includeTasks = tt.includeTasks
			defer func() { includeTasks = originalIncludeTasks }()

			paths, cleanup := createTestFeatureStructure(t, tt.files)
			defer cleanup()

			got := getAvailableDocs(paths)

			if len(got) != len(tt.want) {
				t.Errorf("getAvailableDocs() returned %d docs, want %d\ngot:  %v\nwant: %v",
					len(got), len(tt.want), got, tt.want)
				return
			}

			// Convert to maps for easier comparison (order doesn't matter)
			gotMap := make(map[string]bool)
			for _, doc := range got {
				gotMap[doc] = true
			}

			for _, doc := range tt.want {
				if !gotMap[doc] {
					t.Errorf("getAvailableDocs() missing expected doc: %s", doc)
				}
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a test directory
	testDir := filepath.Join(tmpDir, "testdir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "existing file",
			path: testFile,
			want: true,
		},
		{
			name: "non-existent file",
			path: filepath.Join(tmpDir, "nonexistent.txt"),
			want: false,
		},
		{
			name: "directory instead of file",
			path: testDir,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fileExists(tt.path)
			if got != tt.want {
				t.Errorf("fileExists(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestDirHasFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an empty directory
	emptyDir := filepath.Join(tmpDir, "empty")
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}

	// Create a directory with files
	filledDir := filepath.Join(tmpDir, "filled")
	if err := os.MkdirAll(filledDir, 0755); err != nil {
		t.Fatalf("Failed to create filled directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(filledDir, "file.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create file in filled directory: %v", err)
	}

	// Create a file (not a directory)
	testFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "directory with files",
			path: filledDir,
			want: true,
		},
		{
			name: "empty directory",
			path: emptyDir,
			want: false,
		},
		{
			name: "non-existent directory",
			path: filepath.Join(tmpDir, "nonexistent"),
			want: false,
		},
		{
			name: "file instead of directory",
			path: testFile,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dirHasFiles(tt.path)
			if got != tt.want {
				t.Errorf("dirHasFiles(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestOutputPathsJSON(t *testing.T) {
	paths := &tchncrt.FeaturePaths{
		RepoRoot:      "/test/repo",
		CurrentBranch: "test-branch",
		FeatureDir:    "/test/repo/.tchncrt/features/test-feature",
		FeatureSpec:   "/test/repo/.tchncrt/features/test-feature/spec.md",
		ImplPlan:      "/test/repo/.tchncrt/features/test-feature/plan.md",
		Tasks:         "/test/repo/.tchncrt/features/test-feature/tasks.md",
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputPathsJSON(paths)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputPathsJSON() unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Parse JSON output
	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify expected fields
	expected := map[string]string{
		"REPO_ROOT":    "/test/repo",
		"BRANCH":       "test-branch",
		"FEATURE_DIR":  "/test/repo/.tchncrt/features/test-feature",
		"FEATURE_SPEC": "/test/repo/.tchncrt/features/test-feature/spec.md",
		"IMPL_PLAN":    "/test/repo/.tchncrt/features/test-feature/plan.md",
		"TASKS":        "/test/repo/.tchncrt/features/test-feature/tasks.md",
	}

	for key, want := range expected {
		got, ok := result[key]
		if !ok {
			t.Errorf("outputPathsJSON() missing key: %s", key)
			continue
		}
		if got != want {
			t.Errorf("outputPathsJSON() %s = %q, want %q", key, got, want)
		}
	}
}

func TestOutputPathsText(t *testing.T) {
	paths := &tchncrt.FeaturePaths{
		RepoRoot:      "/test/repo",
		CurrentBranch: "test-branch",
		FeatureDir:    "/test/repo/.tchncrt/features/test-feature",
		FeatureSpec:   "/test/repo/.tchncrt/features/test-feature/spec.md",
		ImplPlan:      "/test/repo/.tchncrt/features/test-feature/plan.md",
		Tasks:         "/test/repo/.tchncrt/features/test-feature/tasks.md",
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputPathsText(paths)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputPathsText() unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify expected lines
	expectedLines := []string{
		"REPO_ROOT: /test/repo",
		"BRANCH: test-branch",
		"FEATURE_DIR: /test/repo/.tchncrt/features/test-feature",
		"FEATURE_SPEC: /test/repo/.tchncrt/features/test-feature/spec.md",
		"IMPL_PLAN: /test/repo/.tchncrt/features/test-feature/plan.md",
		"TASKS: /test/repo/.tchncrt/features/test-feature/tasks.md",
	}

	for _, expected := range expectedLines {
		if !strings.Contains(output, expected) {
			t.Errorf("outputPathsText() output missing expected line: %s\nGot:\n%s", expected, output)
		}
	}
}

func TestOutputJSON(t *testing.T) {
	paths := &tchncrt.FeaturePaths{
		FeatureDir: "/test/repo/.tchncrt/features/test-feature",
	}

	docs := []string{"research.md", "data-model.md", "contracts/"}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputJSON(paths, docs)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputJSON() unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify FEATURE_DIR
	featureDir, ok := result["FEATURE_DIR"].(string)
	if !ok {
		t.Errorf("outputJSON() FEATURE_DIR is not a string")
	} else if featureDir != "/test/repo/.tchncrt/features/test-feature" {
		t.Errorf("outputJSON() FEATURE_DIR = %q, want %q",
			featureDir, "/test/repo/.tchncrt/features/test-feature")
	}

	// Verify AVAILABLE_DOCS
	availableDocs, ok := result["AVAILABLE_DOCS"].([]interface{})
	if !ok {
		t.Fatalf("outputJSON() AVAILABLE_DOCS is not an array")
	}

	if len(availableDocs) != len(docs) {
		t.Errorf("outputJSON() AVAILABLE_DOCS has %d items, want %d",
			len(availableDocs), len(docs))
	}

	// Convert to string slice for comparison
	gotDocs := make([]string, len(availableDocs))
	for i, doc := range availableDocs {
		gotDocs[i] = doc.(string)
	}

	for i, want := range docs {
		if gotDocs[i] != want {
			t.Errorf("outputJSON() AVAILABLE_DOCS[%d] = %q, want %q", i, gotDocs[i], want)
		}
	}
}

func TestOutputText(t *testing.T) {
	tests := []struct {
		name         string
		files        map[string]bool
		includeTasks bool
		wantLines    []string
	}{
		{
			name: "no optional docs",
			files: map[string]bool{
				"plan.md": true,
			},
			includeTasks: false,
			wantLines: []string{
				"FEATURE_DIR:",
				"AVAILABLE_DOCS:",
				"✗ research.md",
				"✗ data-model.md",
				"✗ contracts/",
				"✗ quickstart.md",
			},
		},
		{
			name: "all optional docs present",
			files: map[string]bool{
				"plan.md":       true,
				"research.md":   true,
				"data-model.md": true,
				"contracts":     true,
				"quickstart.md": true,
			},
			includeTasks: false,
			wantLines: []string{
				"FEATURE_DIR:",
				"AVAILABLE_DOCS:",
				"✓ research.md",
				"✓ data-model.md",
				"✓ contracts/",
				"✓ quickstart.md",
			},
		},
		{
			name: "tasks.md included",
			files: map[string]bool{
				"plan.md":  true,
				"tasks.md": true,
			},
			includeTasks: true,
			wantLines: []string{
				"FEATURE_DIR:",
				"AVAILABLE_DOCS:",
				"✓ tasks.md",
			},
		},
		{
			name: "mixed present and missing",
			files: map[string]bool{
				"plan.md":     true,
				"research.md": true,
				"contracts":   true,
			},
			includeTasks: false,
			wantLines: []string{
				"FEATURE_DIR:",
				"AVAILABLE_DOCS:",
				"✓ research.md",
				"✗ data-model.md",
				"✓ contracts/",
				"✗ quickstart.md",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global includeTasks flag
			originalIncludeTasks := includeTasks
			includeTasks = tt.includeTasks
			defer func() { includeTasks = originalIncludeTasks }()

			paths, cleanup := createTestFeatureStructure(t, tt.files)
			defer cleanup()

			docs := getAvailableDocs(paths)

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := outputText(paths, docs)

			w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Fatalf("outputText() unexpected error: %v", err)
			}

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// Verify expected lines
			for _, expected := range tt.wantLines {
				if !strings.Contains(output, expected) {
					t.Errorf("outputText() output missing expected line: %s\nGot:\n%s",
						expected, output)
				}
			}
		})
	}
}

func TestCheckFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	existingFile := filepath.Join(tmpDir, "exists.txt")
	if err := os.WriteFile(existingFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		filename string
		want     string
	}{
		{
			name:     "existing file",
			path:     existingFile,
			filename: "exists.txt",
			want:     "✓ exists.txt",
		},
		{
			name:     "non-existent file",
			path:     filepath.Join(tmpDir, "missing.txt"),
			filename: "missing.txt",
			want:     "✗ missing.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			checkFile(tt.path, tt.filename)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := strings.TrimSpace(buf.String())

			if !strings.Contains(output, tt.want) {
				t.Errorf("checkFile() output = %q, want to contain %q", output, tt.want)
			}
		})
	}
}

func TestCheckDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an empty directory
	emptyDir := filepath.Join(tmpDir, "empty")
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}

	// Create a directory with files
	filledDir := filepath.Join(tmpDir, "filled")
	if err := os.MkdirAll(filledDir, 0755); err != nil {
		t.Fatalf("Failed to create filled directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(filledDir, "file.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create file in filled directory: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		dirname string
		want    string
	}{
		{
			name:    "directory with files",
			path:    filledDir,
			dirname: "filled/",
			want:    "✓ filled/",
		},
		{
			name:    "empty directory",
			path:    emptyDir,
			dirname: "empty/",
			want:    "✗ empty/",
		},
		{
			name:    "non-existent directory",
			path:    filepath.Join(tmpDir, "missing"),
			dirname: "missing/",
			want:    "✗ missing/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			checkDir(tt.path, tt.dirname)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := strings.TrimSpace(buf.String())

			if !strings.Contains(output, tt.want) {
				t.Errorf("checkDir() output = %q, want to contain %q", output, tt.want)
			}
		})
	}
}

func TestIntegration_ValidateAndOutput(t *testing.T) {
	// This test simulates the full flow of validation and output
	files := map[string]bool{
		"plan.md":       true,
		"tasks.md":      true,
		"research.md":   true,
		"data-model.md": true,
		"contracts":     true,
	}

	paths, cleanup := createTestFeatureStructure(t, files)
	defer cleanup()

	// Test validation passes
	if err := validatePrerequisites(paths); err != nil {
		t.Errorf("validatePrerequisites() unexpected error: %v", err)
	}

	// Test getting available docs
	originalIncludeTasks := includeTasks
	includeTasks = true
	defer func() { includeTasks = originalIncludeTasks }()

	docs := getAvailableDocs(paths)
	if len(docs) == 0 {
		t.Error("getAvailableDocs() returned no documents")
	}

	// Test JSON output doesn't error
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputJSON(paths, docs)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("outputJSON() unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	// Verify valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Errorf("outputJSON() produced invalid JSON: %v", err)
	}
}

// TestRunCheckPrerequisites tests the main command entry point
func TestRunCheckPrerequisites(t *testing.T) {
	// Save original flags and restore them after test
	originalPathsOnly := pathsOnly
	originalJsonMode := jsonMode
	defer func() {
		pathsOnly = originalPathsOnly
		jsonMode = originalJsonMode
	}()

	tests := []struct {
		name        string
		setupFunc   func() string // Returns tmpDir path
		pathsOnly   bool
		jsonMode    bool
		wantErr     bool
		errContains string
	}{
		{
			name: "paths only text mode",
			setupFunc: func() string {
				tmpDir := t.TempDir()
				createSimpleTestStructure(tmpDir)
				return tmpDir
			},
			pathsOnly: true,
			jsonMode:  false,
			wantErr:   false,
		},
		{
			name: "paths only json mode",
			setupFunc: func() string {
				tmpDir := t.TempDir()
				createSimpleTestStructure(tmpDir)
				return tmpDir
			},
			pathsOnly: true,
			jsonMode:  true,
			wantErr:   false,
		},
		{
			name: "full validation text mode",
			setupFunc: func() string {
				tmpDir := t.TempDir()
				createSimpleTestStructure(tmpDir)
				return tmpDir
			},
			pathsOnly: false,
			jsonMode:  false,
			wantErr:   false,
		},
		{
			name: "full validation json mode",
			setupFunc: func() string {
				tmpDir := t.TempDir()
				createSimpleTestStructure(tmpDir)
				return tmpDir
			},
			pathsOnly: false,
			jsonMode:  true,
			wantErr:   false,
		},
		{
			name: "missing feature directory",
			setupFunc: func() string {
				tmpDir := t.TempDir()
				// Initialize git properly
				exec.Command("git", "-C", tmpDir, "init").Run()
				exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()
				exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
				os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("test"), 0644)
				exec.Command("git", "-C", tmpDir, "add", ".").Run()
				exec.Command("git", "-C", tmpDir, "commit", "-m", "init").Run()
				exec.Command("git", "-C", tmpDir, "checkout", "-b", "feature/test").Run()
				// Don't create .tchncrt/features directory at all
				return tmpDir
			},
			pathsOnly:   false,
			jsonMode:    false,
			wantErr:     true,
			errContains: "feature directory not found",
		},
		{
			name: "missing plan.md",
			setupFunc: func() string {
				tmpDir := t.TempDir()
				// Initialize git properly
				exec.Command("git", "-C", tmpDir, "init").Run()
				exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()
				exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
				os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("test"), 0644)
				exec.Command("git", "-C", tmpDir, "add", ".").Run()
				exec.Command("git", "-C", tmpDir, "commit", "-m", "init").Run()
				exec.Command("git", "-C", tmpDir, "checkout", "-b", "feature/test").Run()
				// Create .tchncrt structure but not plan.md
				featureDir := filepath.Join(tmpDir, ".tchncrt", "features", "test")
				os.MkdirAll(featureDir, 0755)
				// Don't create plan.md
				return tmpDir
			},
			pathsOnly:   false,
			jsonMode:    false,
			wantErr:     true,
			errContains: "plan.md not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			tmpDir := tt.setupFunc()

			// Change to test directory
			originalDir, _ := os.Getwd()
			os.Chdir(tmpDir)
			defer os.Chdir(originalDir)

			// Set flags
			pathsOnly = tt.pathsOnly
			jsonMode = tt.jsonMode

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create a mock command (flags don't affect the function logic)
			cmd := &cobra.Command{}
			err := runCheckPrerequisites(cmd, []string{})

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read output
			var buf bytes.Buffer
			buf.ReadFrom(r)

			// Check error expectation
			if tt.wantErr {
				if err == nil {
					t.Errorf("runCheckPrerequisites() expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("runCheckPrerequisites() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("runCheckPrerequisites() unexpected error: %v", err)
				}

				// Verify output format
				output := buf.String()
				if tt.jsonMode {
					// Should be valid JSON
					var result map[string]interface{}
					if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
						t.Errorf("runCheckPrerequisites() produced invalid JSON: %v\nOutput: %s", err, output)
					}
				} else {
					// Text mode should have some output
					if len(output) == 0 {
						t.Error("runCheckPrerequisites() produced no output in text mode")
					}
				}
			}
		})
	}
}

// TestRunCheckPrerequisites_EdgeCases tests edge cases and error conditions
func TestRunCheckPrerequisites_EdgeCases(t *testing.T) {
	// Save original flags
	originalPathsOnly := pathsOnly
	originalJsonMode := jsonMode
	defer func() {
		pathsOnly = originalPathsOnly
		jsonMode = originalJsonMode
	}()

	tests := []struct {
		name        string
		setupFunc   func() (string, error)
		pathsOnly   bool
		jsonMode    bool
		wantErr     bool
		errContains string
	}{
		{
			name: "valid feature directory",
			setupFunc: func() (string, error) {
				tmpDir := t.TempDir()
				err := createSimpleTestStructure(tmpDir)
				return tmpDir, err
			},
			pathsOnly: false,
			jsonMode:  true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			originalDir, _ := os.Getwd()
			defer os.Chdir(originalDir)

			tmpDir, setupErr := tt.setupFunc()
			if setupErr != nil && !tt.wantErr {
				t.Fatalf("Setup failed: %v", setupErr)
			}

			if tmpDir != "" {
				os.Chdir(tmpDir)
			}

			pathsOnly = tt.pathsOnly
			jsonMode = tt.jsonMode

			// Capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			cmd := &cobra.Command{}
			err := runCheckPrerequisites(cmd, []string{})

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)

			if tt.wantErr {
				if err == nil {
					t.Errorf("runCheckPrerequisites() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("runCheckPrerequisites() unexpected error: %v", err)
				}
			}
		})
	}
}
