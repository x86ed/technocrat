package cmd

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestGetRepoRoot tests the getRepoRoot function
func TestGetRepoRoot(t *testing.T) {
	t.Run("in git repository", func(t *testing.T) {
		// Create a temporary directory with git
		tmpDir := t.TempDir()

		// Initialize git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Skipf("Skipping test: git not available or failed: %v", err)
		}

		// Change to temp directory
		origDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		root, err := getRepoRoot()
		if err != nil {
			t.Errorf("getRepoRoot() error = %v", err)
		}

		// Compare absolute paths, resolving symlinks
		expectedPath, _ := filepath.EvalSymlinks(tmpDir)
		actualPath, _ := filepath.EvalSymlinks(root)
		if actualPath != expectedPath {
			t.Errorf("getRepoRoot() = %v, want %v", actualPath, expectedPath)
		}
	})

	t.Run("with go.mod file", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create go.mod
		goModPath := filepath.Join(tmpDir, "go.mod")
		if err := os.WriteFile(goModPath, []byte("module test\n"), 0644); err != nil {
			t.Fatal(err)
		}

		// Create subdirectory and change to it
		subDir := filepath.Join(tmpDir, "subdir")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatal(err)
		}

		origDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		if err := os.Chdir(subDir); err != nil {
			t.Fatal(err)
		}

		root, err := getRepoRoot()
		if err != nil {
			t.Errorf("getRepoRoot() error = %v", err)
		}

		// Should find the directory with go.mod, resolving symlinks
		expectedPath, _ := filepath.EvalSymlinks(tmpDir)
		actualPath, _ := filepath.EvalSymlinks(root)
		if actualPath != expectedPath {
			t.Errorf("getRepoRoot() = %v, want %v", actualPath, expectedPath)
		}
	})

	t.Run("fallback to current directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		origDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		root, err := getRepoRoot()
		if err != nil {
			t.Errorf("getRepoRoot() error = %v", err)
		}

		// Should return current directory when no git or go.mod found
		if root == "" {
			t.Error("getRepoRoot() should return a path")
		}
	})
}

// TestGetCurrentBranch tests the getCurrentBranch function
func TestGetCurrentBranch(t *testing.T) {
	t.Run("with setFeature parameter", func(t *testing.T) {
		branch, err := getCurrentBranch("001-test-feature")
		if err != nil {
			t.Errorf("getCurrentBranch() error = %v", err)
		}
		if branch != "001-test-feature" {
			t.Errorf("getCurrentBranch() = %v, want %v", branch, "001-test-feature")
		}
	})

	t.Run("with TCHNCRT_FEATURE environment variable", func(t *testing.T) {
		os.Setenv("TCHNCRT_FEATURE", "002-env-feature")
		defer os.Unsetenv("TCHNCRT_FEATURE")

		branch, err := getCurrentBranch("")
		if err != nil {
			t.Errorf("getCurrentBranch() error = %v", err)
		}
		if branch != "002-env-feature" {
			t.Errorf("getCurrentBranch() = %v, want %v", branch, "002-env-feature")
		}
	})

	t.Run("in git repository", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Initialize git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Skipf("Skipping test: git not available: %v", err)
		}

		// Configure git
		cmd = exec.Command("git", "config", "user.name", "Test User")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Skip("Skipping test: git config failed")
		}

		cmd = exec.Command("git", "config", "user.email", "test@example.com")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Skip("Skipping test: git config failed")
		}

		// Create initial commit
		testFile := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		cmd = exec.Command("git", "add", ".")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Skip("Skipping test: git add failed")
		}

		cmd = exec.Command("git", "commit", "-m", "Initial commit")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Skip("Skipping test: git commit failed")
		}

		// Create and checkout a feature branch
		cmd = exec.Command("git", "checkout", "-b", "003-git-feature")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Skip("Skipping test: git checkout failed")
		}

		origDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		branch, err := getCurrentBranch("")
		if err != nil {
			t.Errorf("getCurrentBranch() error = %v", err)
		}
		if branch != "003-git-feature" {
			t.Errorf("getCurrentBranch() = %v, want %v", branch, "003-git-feature")
		}
	})

	t.Run("fallback with specs directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create specs directory with numbered features
		specsDir := filepath.Join(tmpDir, "specs")
		if err := os.MkdirAll(filepath.Join(specsDir, "001-first-feature"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(specsDir, "003-third-feature"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(specsDir, "002-second-feature"), 0755); err != nil {
			t.Fatal(err)
		}

		origDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		branch, err := getCurrentBranch("")
		if err != nil {
			t.Errorf("getCurrentBranch() error = %v", err)
		}
		// Should return the highest numbered feature
		if branch != "003-third-feature" {
			t.Errorf("getCurrentBranch() = %v, want %v", branch, "003-third-feature")
		}
	})

	t.Run("fallback to main", func(t *testing.T) {
		tmpDir := t.TempDir()

		origDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		branch, err := getCurrentBranch("")
		if err != nil {
			t.Errorf("getCurrentBranch() error = %v", err)
		}
		if branch != "main" {
			t.Errorf("getCurrentBranch() = %v, want %v", branch, "main")
		}
	})
}

// TestHasGit tests the hasGit function
func TestHasGit(t *testing.T) {
	t.Run("in git repository", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Skipf("Skipping test: git not available: %v", err)
		}

		origDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		if !hasGit() {
			t.Error("hasGit() = false, want true")
		}
	})

	t.Run("not in git repository", func(t *testing.T) {
		tmpDir := t.TempDir()

		origDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		if hasGit() {
			t.Error("hasGit() = true, want false")
		}
	})
}

// TestCheckFeatureBranch tests the checkFeatureBranch function
func TestCheckFeatureBranch(t *testing.T) {
	tests := []struct {
		name      string
		branch    string
		hasGit    bool
		wantError bool
	}{
		{
			name:      "valid feature branch",
			branch:    "001-valid-feature",
			hasGit:    true,
			wantError: false,
		},
		{
			name:      "valid feature branch with more digits",
			branch:    "999-another-feature",
			hasGit:    true,
			wantError: false,
		},
		{
			name:      "invalid branch - no number",
			branch:    "main",
			hasGit:    true,
			wantError: true,
		},
		{
			name:      "invalid branch - wrong format",
			branch:    "feature-001",
			hasGit:    true,
			wantError: true,
		},
		{
			name:      "non-git repo - should not error",
			branch:    "main",
			hasGit:    false,
			wantError: false,
		},
		{
			name:      "invalid branch in non-git repo",
			branch:    "invalid",
			hasGit:    false,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkFeatureBranch(tt.branch, tt.hasGit)
			if (err != nil) != tt.wantError {
				t.Errorf("checkFeatureBranch() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestGetFeatureDir tests the getFeatureDir function
func TestGetFeatureDir(t *testing.T) {
	tests := []struct {
		name     string
		repoRoot string
		branch   string
		expected string
	}{
		{
			name:     "standard feature branch",
			repoRoot: "/path/to/repo",
			branch:   "001-feature",
			expected: "/path/to/repo/specs/001-feature",
		},
		{
			name:     "main branch",
			repoRoot: "/home/user/project",
			branch:   "main",
			expected: "/home/user/project/specs/main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFeatureDir(tt.repoRoot, tt.branch)
			if result != tt.expected {
				t.Errorf("getFeatureDir() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetFeaturePaths tests the getFeaturePaths function
func TestGetFeaturePaths(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a go.mod to establish repo root
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	paths, err := getFeaturePaths("001-test-feature")
	if err != nil {
		t.Fatalf("getFeaturePaths() error = %v", err)
	}

	// Verify all paths are populated
	if paths.RepoRoot == "" {
		t.Error("RepoRoot should not be empty")
	}
	if paths.CurrentBranch != "001-test-feature" {
		t.Errorf("CurrentBranch = %v, want %v", paths.CurrentBranch, "001-test-feature")
	}
	if paths.FeatureDir == "" {
		t.Error("FeatureDir should not be empty")
	}
	if paths.FeatureSpec == "" {
		t.Error("FeatureSpec should not be empty")
	}
	if paths.ImplPlan == "" {
		t.Error("ImplPlan should not be empty")
	}
	if paths.Tasks == "" {
		t.Error("Tasks should not be empty")
	}
	if paths.Research == "" {
		t.Error("Research should not be empty")
	}
	if paths.DataModel == "" {
		t.Error("DataModel should not be empty")
	}
	if paths.Quickstart == "" {
		t.Error("Quickstart should not be empty")
	}
	if paths.ContractsDir == "" {
		t.Error("ContractsDir should not be empty")
	}

	// Verify path structure
	expectedFeatureDir := filepath.Join(paths.RepoRoot, "specs", "001-test-feature")
	if paths.FeatureDir != expectedFeatureDir {
		t.Errorf("FeatureDir = %v, want %v", paths.FeatureDir, expectedFeatureDir)
	}

	expectedSpec := filepath.Join(expectedFeatureDir, "spec.md")
	if paths.FeatureSpec != expectedSpec {
		t.Errorf("FeatureSpec = %v, want %v", paths.FeatureSpec, expectedSpec)
	}

	expectedPlan := filepath.Join(expectedFeatureDir, "plan.md")
	if paths.ImplPlan != expectedPlan {
		t.Errorf("ImplPlan = %v, want %v", paths.ImplPlan, expectedPlan)
	}

	expectedTasks := filepath.Join(expectedFeatureDir, "tasks.md")
	if paths.Tasks != expectedTasks {
		t.Errorf("Tasks = %v, want %v", paths.Tasks, expectedTasks)
	}

	expectedResearch := filepath.Join(expectedFeatureDir, "research.md")
	if paths.Research != expectedResearch {
		t.Errorf("Research = %v, want %v", paths.Research, expectedResearch)
	}

	expectedDataModel := filepath.Join(expectedFeatureDir, "data-model.md")
	if paths.DataModel != expectedDataModel {
		t.Errorf("DataModel = %v, want %v", paths.DataModel, expectedDataModel)
	}

	expectedQuickstart := filepath.Join(expectedFeatureDir, "quickstart.md")
	if paths.Quickstart != expectedQuickstart {
		t.Errorf("Quickstart = %v, want %v", paths.Quickstart, expectedQuickstart)
	}

	expectedContractsDir := filepath.Join(expectedFeatureDir, "contracts")
	if paths.ContractsDir != expectedContractsDir {
		t.Errorf("ContractsDir = %v, want %v", paths.ContractsDir, expectedContractsDir)
	}
}

// TestPrintAllPaths tests the printAllPaths function
func TestPrintAllPaths(t *testing.T) {
	paths := &FeaturePaths{
		RepoRoot:      "/test/repo",
		CurrentBranch: "001-test",
		HasGit:        true,
		FeatureDir:    "/test/repo/specs/001-test",
		FeatureSpec:   "/test/repo/specs/001-test/spec.md",
		ImplPlan:      "/test/repo/specs/001-test/plan.md",
		Tasks:         "/test/repo/specs/001-test/tasks.md",
		Research:      "/test/repo/specs/001-test/research.md",
		DataModel:     "/test/repo/specs/001-test/data-model.md",
		Quickstart:    "/test/repo/specs/001-test/quickstart.md",
		ContractsDir:  "/test/repo/specs/001-test/contracts",
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printAllPaths(paths)

	w.Close()
	os.Stdout = oldStdout

	outputBytes, _ := io.ReadAll(r)
	output := string(outputBytes)

	// Verify output contains expected values
	expectedStrings := []string{
		"REPO_ROOT='/test/repo'",
		"CURRENT_BRANCH='001-test'",
		"HAS_GIT='true'",
		"FEATURE_DIR='/test/repo/specs/001-test'",
		"FEATURE_SPEC='/test/repo/specs/001-test/spec.md'",
		"IMPL_PLAN='/test/repo/specs/001-test/plan.md'",
		"TASKS='/test/repo/specs/001-test/tasks.md'",
		"RESEARCH='/test/repo/specs/001-test/research.md'",
		"DATA_MODEL='/test/repo/specs/001-test/data-model.md'",
		"QUICKSTART='/test/repo/specs/001-test/quickstart.md'",
		"CONTRACTS_DIR='/test/repo/specs/001-test/contracts'",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("printAllPaths() output missing: %v", expected)
		}
	}
}

// TestCheckFeatureFiles tests the checkFeatureFiles function
func TestCheckFeatureFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create feature directory structure
	featureDir := filepath.Join(tmpDir, "specs", "001-test")
	if err := os.MkdirAll(featureDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create some files
	if err := os.WriteFile(filepath.Join(featureDir, "spec.md"), []byte("# Spec"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(featureDir, "plan.md"), []byte("# Plan"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create contracts directory with a file
	contractsDir := filepath.Join(featureDir, "contracts")
	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(contractsDir, "api.md"), []byte("# API"), 0644); err != nil {
		t.Fatal(err)
	}

	paths := &FeaturePaths{
		RepoRoot:      tmpDir,
		CurrentBranch: "001-test",
		HasGit:        false,
		FeatureDir:    featureDir,
		FeatureSpec:   filepath.Join(featureDir, "spec.md"),
		ImplPlan:      filepath.Join(featureDir, "plan.md"),
		Tasks:         filepath.Join(featureDir, "tasks.md"),
		Research:      filepath.Join(featureDir, "research.md"),
		DataModel:     filepath.Join(featureDir, "data-model.md"),
		Quickstart:    filepath.Join(featureDir, "quickstart.md"),
		ContractsDir:  contractsDir,
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := checkFeatureFiles(paths)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("checkFeatureFiles() error = %v", err)
	}

	outputBytes, _ := io.ReadAll(r)
	output := string(outputBytes)

	// Verify output contains checkmarks for existing files
	if !strings.Contains(output, "✓ Feature Specification (spec.md)") {
		t.Error("checkFeatureFiles() should show spec.md as existing")
	}
	if !strings.Contains(output, "✓ Implementation Plan (plan.md)") {
		t.Error("checkFeatureFiles() should show plan.md as existing")
	}
	if !strings.Contains(output, "✓ Contracts Directory") {
		t.Error("checkFeatureFiles() should show contracts directory as existing")
	}

	// Verify output contains X marks for missing files
	if !strings.Contains(output, "✗ Tasks (tasks.md)") {
		t.Error("checkFeatureFiles() should show tasks.md as missing")
	}
	if !strings.Contains(output, "✗ Research (research.md)") {
		t.Error("checkFeatureFiles() should show research.md as missing")
	}
}

// TestCheckFeatureFile tests the checkFeatureFile function
func TestCheckFeatureFile(t *testing.T) {
	tmpDir := t.TempDir()

	existingFile := filepath.Join(tmpDir, "existing.md")
	if err := os.WriteFile(existingFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	missingFile := filepath.Join(tmpDir, "missing.md")

	tests := []struct {
		name        string
		path        string
		description string
		shouldExist bool
	}{
		{
			name:        "existing file",
			path:        existingFile,
			description: "Existing File",
			shouldExist: true,
		},
		{
			name:        "missing file",
			path:        missingFile,
			description: "Missing File",
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			checkFeatureFile(tt.path, tt.description)

			w.Close()
			os.Stdout = oldStdout

			outputBytes, _ := io.ReadAll(r)
			output := string(outputBytes)

			if tt.shouldExist {
				if !strings.Contains(output, "✓") {
					t.Errorf("checkFeatureFile() should show checkmark for existing file")
				}
			} else {
				if !strings.Contains(output, "✗") {
					t.Errorf("checkFeatureFile() should show X mark for missing file")
				}
			}

			if !strings.Contains(output, tt.description) {
				t.Errorf("checkFeatureFile() should contain description: %v", tt.description)
			}
		})
	}
}

// TestCheckFeatureDir tests the checkFeatureDir function
func TestCheckFeatureDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a non-empty directory
	nonEmptyDir := filepath.Join(tmpDir, "nonempty")
	if err := os.MkdirAll(nonEmptyDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nonEmptyDir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create an empty directory
	emptyDir := filepath.Join(tmpDir, "empty")
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatal(err)
	}

	missingDir := filepath.Join(tmpDir, "missing")

	tests := []struct {
		name        string
		path        string
		description string
		shouldExist bool
	}{
		{
			name:        "non-empty directory",
			path:        nonEmptyDir,
			description: "Non-empty Directory",
			shouldExist: true,
		},
		{
			name:        "empty directory",
			path:        emptyDir,
			description: "Empty Directory",
			shouldExist: false,
		},
		{
			name:        "missing directory",
			path:        missingDir,
			description: "Missing Directory",
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			checkFeatureDir(tt.path, tt.description)

			w.Close()
			os.Stdout = oldStdout

			outputBytes, _ := io.ReadAll(r)
			output := string(outputBytes)

			if tt.shouldExist {
				if !strings.Contains(output, "✓") {
					t.Errorf("checkFeatureDir() should show checkmark for valid directory")
				}
			} else {
				if !strings.Contains(output, "✗") {
					t.Errorf("checkFeatureDir() should show X mark for invalid/missing directory")
				}
			}

			if !strings.Contains(output, tt.description) {
				t.Errorf("checkFeatureDir() should contain description: %v", tt.description)
			}
		})
	}
}

// TestRunCommon tests the runCommon command execution
func TestRunCommon(t *testing.T) {
	t.Run("validate branch - valid", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		// Create go.mod to establish repo root
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n"), 0644); err != nil {
			t.Fatal(err)
		}

		// Reset flags
		validateBranch = true
		setFeature = "001-valid-feature"
		defer func() {
			validateBranch = false
			setFeature = ""
		}()

		err = runCommon(nil, []string{})
		if err != nil {
			t.Errorf("runCommon() with valid branch should not error: %v", err)
		}
	})

	t.Run("validate branch - invalid", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Initialize git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Skipf("Skipping test: git not available: %v", err)
		}

		origDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		// Reset flags
		validateBranch = true
		setFeature = "invalid-branch"
		defer func() {
			validateBranch = false
			setFeature = ""
		}()

		err = runCommon(nil, []string{})
		if err == nil {
			t.Error("runCommon() with invalid branch should error")
		}
	})

	t.Run("show all paths", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create go.mod
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n"), 0644); err != nil {
			t.Fatal(err)
		}

		origDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		// Reset flags
		showAll = true
		setFeature = "001-test-feature"
		defer func() {
			showAll = false
			setFeature = ""
		}()

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err = runCommon(nil, []string{})

		w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Errorf("runCommon() error = %v", err)
		}

		outputBytes, _ := io.ReadAll(r)
		output := string(outputBytes)

		// Verify output contains expected keys
		expectedKeys := []string{
			"REPO_ROOT=",
			"CURRENT_BRANCH=",
			"HAS_GIT=",
			"FEATURE_DIR=",
			"FEATURE_SPEC=",
		}

		for _, key := range expectedKeys {
			if !strings.Contains(output, key) {
				t.Errorf("runCommon() output missing key: %v", key)
			}
		}
	})

	t.Run("check files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create go.mod
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n"), 0644); err != nil {
			t.Fatal(err)
		}

		origDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		// Reset flags
		checkFiles = true
		setFeature = "001-test-feature"
		defer func() {
			checkFiles = false
			setFeature = ""
		}()

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err = runCommon(nil, []string{})

		w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Errorf("runCommon() error = %v", err)
		}

		outputBytes, _ := io.ReadAll(r)
		output := string(outputBytes)

		if !strings.Contains(output, "Feature Files Status:") {
			t.Error("runCommon() with --check should show feature files status")
		}
	})
}
