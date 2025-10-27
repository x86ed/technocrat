package cmd

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestFindRepoRootForFeature tests the findRepoRootForFeature function
func TestFindRepoRootForFeature(t *testing.T) {
	t.Run("with git repository", func(t *testing.T) {
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

		root, err := findRepoRootForFeature()
		if err != nil {
			t.Errorf("findRepoRootForFeature() error = %v", err)
		}

		// Resolve symlinks for comparison
		expectedPath, _ := filepath.EvalSymlinks(tmpDir)
		actualPath, _ := filepath.EvalSymlinks(root)
		if actualPath != expectedPath {
			t.Errorf("findRepoRootForFeature() = %v, want %v", actualPath, expectedPath)
		}
	})

	t.Run("with .tchncrt directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create .tchncrt directory
		tchncrtDir := filepath.Join(tmpDir, ".tchncrt")
		if err := os.MkdirAll(tchncrtDir, 0755); err != nil {
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

		root, err := findRepoRootForFeature()
		if err != nil {
			t.Errorf("findRepoRootForFeature() error = %v", err)
		}

		expectedPath, _ := filepath.EvalSymlinks(tmpDir)
		actualPath, _ := filepath.EvalSymlinks(root)
		if actualPath != expectedPath {
			t.Errorf("findRepoRootForFeature() = %v, want %v", actualPath, expectedPath)
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
		subDir := filepath.Join(tmpDir, "internal", "pkg")
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

		root, err := findRepoRootForFeature()
		if err != nil {
			t.Errorf("findRepoRootForFeature() error = %v", err)
		}

		expectedPath, _ := filepath.EvalSymlinks(tmpDir)
		actualPath, _ := filepath.EvalSymlinks(root)
		if actualPath != expectedPath {
			t.Errorf("findRepoRootForFeature() = %v, want %v", actualPath, expectedPath)
		}
	})

	t.Run("no repository markers found", func(t *testing.T) {
		tmpDir := t.TempDir()

		origDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		_, err = findRepoRootForFeature()
		if err == nil {
			t.Error("findRepoRootForFeature() should return error when no markers found")
		}
		if !strings.Contains(err.Error(), "could not find repository root") {
			t.Errorf("findRepoRootForFeature() error message = %v, want 'could not find repository root'", err)
		}
	})
}

// TestFindHighestFeatureNumber tests the findHighestFeatureNumber function
func TestFindHighestFeatureNumber(t *testing.T) {
	t.Run("empty specs directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		specsDir := filepath.Join(tmpDir, "specs")
		if err := os.MkdirAll(specsDir, 0755); err != nil {
			t.Fatal(err)
		}

		highest, err := findHighestFeatureNumber(specsDir)
		if err != nil {
			t.Errorf("findHighestFeatureNumber() error = %v", err)
		}
		if highest != 0 {
			t.Errorf("findHighestFeatureNumber() = %d, want 0", highest)
		}
	})

	t.Run("non-existent specs directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		specsDir := filepath.Join(tmpDir, "specs")

		highest, err := findHighestFeatureNumber(specsDir)
		if err != nil {
			t.Errorf("findHighestFeatureNumber() error = %v", err)
		}
		if highest != 0 {
			t.Errorf("findHighestFeatureNumber() = %d, want 0", highest)
		}
	})

	t.Run("with numbered feature directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		specsDir := filepath.Join(tmpDir, "specs")

		// Create feature directories with different numbers
		features := []string{"001-first-feature", "005-fifth-feature", "003-third-feature"}
		for _, feature := range features {
			featureDir := filepath.Join(specsDir, feature)
			if err := os.MkdirAll(featureDir, 0755); err != nil {
				t.Fatal(err)
			}
		}

		highest, err := findHighestFeatureNumber(specsDir)
		if err != nil {
			t.Errorf("findHighestFeatureNumber() error = %v", err)
		}
		if highest != 5 {
			t.Errorf("findHighestFeatureNumber() = %d, want 5", highest)
		}
	})

	t.Run("with mixed directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		specsDir := filepath.Join(tmpDir, "specs")

		// Create mix of numbered and non-numbered directories
		dirs := []string{
			"001-feature",
			"010-another-feature",
			"non-numbered-feature",
			"README.md", // file, not directory
		}
		for _, dir := range dirs[:3] {
			if err := os.MkdirAll(filepath.Join(specsDir, dir), 0755); err != nil {
				t.Fatal(err)
			}
		}
		// Create file
		if err := os.WriteFile(filepath.Join(specsDir, dirs[3]), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		highest, err := findHighestFeatureNumber(specsDir)
		if err != nil {
			t.Errorf("findHighestFeatureNumber() error = %v", err)
		}
		if highest != 10 {
			t.Errorf("findHighestFeatureNumber() = %d, want 10", highest)
		}
	})

	t.Run("with leading zeros", func(t *testing.T) {
		tmpDir := t.TempDir()
		specsDir := filepath.Join(tmpDir, "specs")

		// Create feature directories with leading zeros
		features := []string{"001-feature", "099-feature", "100-feature"}
		for _, feature := range features {
			if err := os.MkdirAll(filepath.Join(specsDir, feature), 0755); err != nil {
				t.Fatal(err)
			}
		}

		highest, err := findHighestFeatureNumber(specsDir)
		if err != nil {
			t.Errorf("findHighestFeatureNumber() error = %v", err)
		}
		if highest != 100 {
			t.Errorf("findHighestFeatureNumber() = %d, want 100", highest)
		}
	})
}

// TestCreateBranchName tests the createBranchName function
func TestCreateBranchName(t *testing.T) {
	tests := []struct {
		name        string
		description string
		featureNum  string
		expected    string
	}{
		{
			name:        "simple description",
			description: "Add User Authentication",
			featureNum:  "001",
			expected:    "001-add-user-authentication",
		},
		{
			name:        "more than 3 words",
			description: "Add User Authentication And Authorization System",
			featureNum:  "002",
			expected:    "002-add-user-authentication",
		},
		{
			name:        "special characters",
			description: "Fix Bug #123 in API",
			featureNum:  "003",
			expected:    "003-fix-bug-123",
		},
		{
			name:        "multiple spaces",
			description: "Add   Multiple    Spaces",
			featureNum:  "004",
			expected:    "004-add-multiple-spaces",
		},
		{
			name:        "punctuation",
			description: "Implement RESTful API endpoints!",
			featureNum:  "005",
			expected:    "005-implement-restful-api",
		},
		{
			name:        "leading and trailing spaces",
			description: "  Trim Spaces  ",
			featureNum:  "006",
			expected:    "006-trim-spaces",
		},
		{
			name:        "mixed case",
			description: "MixedCaseFeature",
			featureNum:  "007",
			expected:    "007-mixedcasefeature",
		},
		{
			name:        "single word",
			description: "Refactor",
			featureNum:  "008",
			expected:    "008-refactor",
		},
		{
			name:        "underscores and dashes",
			description: "fix_bug-in-system",
			featureNum:  "009",
			expected:    "009-fix-bug-in",
		},
		{
			name:        "numbers in description",
			description: "Update API v2.0",
			featureNum:  "010",
			expected:    "010-update-api-v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createBranchName(tt.description, tt.featureNum)
			if result != tt.expected {
				t.Errorf("createBranchName(%q, %q) = %q, want %q",
					tt.description, tt.featureNum, result, tt.expected)
			}
		})
	}
}

// TestCreateGitBranch tests the createGitBranch function
func TestCreateGitBranch(t *testing.T) {
	t.Run("create branch successfully", func(t *testing.T) {
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

		origDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		// Create branch
		branchName := "001-test-feature"
		err = createGitBranch(branchName)
		if err != nil {
			t.Errorf("createGitBranch() error = %v", err)
		}

		// Verify branch was created and checked out
		cmd = exec.Command("git", "branch", "--show-current")
		cmd.Dir = tmpDir
		output, err := cmd.Output()
		if err != nil {
			t.Fatal(err)
		}

		currentBranch := strings.TrimSpace(string(output))
		if currentBranch != branchName {
			t.Errorf("Current branch = %v, want %v", currentBranch, branchName)
		}
	})

	t.Run("branch already exists", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Initialize git repo with initial commit
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Skipf("Skipping test: git not available: %v", err)
		}

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

		testFile := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		cmd = exec.Command("git", "add", ".")
		cmd.Dir = tmpDir
		cmd.Run()

		cmd = exec.Command("git", "commit", "-m", "Initial commit")
		cmd.Dir = tmpDir
		cmd.Run()

		// Create branch first time
		cmd = exec.Command("git", "checkout", "-b", "001-existing-branch")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Skip("Skipping test: first branch creation failed")
		}

		// Go back to main
		cmd = exec.Command("git", "checkout", "main")
		cmd.Dir = tmpDir
		cmd.Run()

		origDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		// Try to create same branch again
		err = createGitBranch("001-existing-branch")
		if err == nil {
			t.Error("createGitBranch() should return error for existing branch")
		}
	})
}

// TestCopyTemplateIfExists tests the copyTemplateIfExists function
func TestCopyTemplateIfExists(t *testing.T) {
	t.Run("template exists", func(t *testing.T) {
		tmpDir := t.TempDir()

		templatePath := filepath.Join(tmpDir, "template.md")
		templateContent := "# Template\nThis is a template"
		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatal(err)
		}

		destPath := filepath.Join(tmpDir, "destination.md")

		err := copyTemplateIfExists(templatePath, destPath)
		if err != nil {
			t.Errorf("copyTemplateIfExists() error = %v", err)
		}

		// Verify file was copied
		content, err := os.ReadFile(destPath)
		if err != nil {
			t.Fatal(err)
		}

		if string(content) != templateContent {
			t.Errorf("Copied content = %q, want %q", string(content), templateContent)
		}
	})

	t.Run("template does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		templatePath := filepath.Join(tmpDir, "nonexistent-template.md")
		destPath := filepath.Join(tmpDir, "destination.md")

		err := copyTemplateIfExists(templatePath, destPath)
		if err != nil {
			t.Errorf("copyTemplateIfExists() error = %v", err)
		}

		// Verify empty file was created
		content, err := os.ReadFile(destPath)
		if err != nil {
			t.Fatal(err)
		}

		if len(content) != 0 {
			t.Errorf("Empty file should be created, got content: %q", string(content))
		}
	})

	t.Run("invalid destination path", func(t *testing.T) {
		tmpDir := t.TempDir()

		templatePath := filepath.Join(tmpDir, "template.md")
		destPath := filepath.Join(tmpDir, "nonexistent", "subdir", "destination.md")

		err := copyTemplateIfExists(templatePath, destPath)
		if err == nil {
			t.Error("copyTemplateIfExists() should return error for invalid destination path")
		}
	})
}

// TestOutputFeatureJSON tests the outputFeatureJSON function
func TestOutputFeatureJSON(t *testing.T) {
	t.Run("valid output", func(t *testing.T) {
		info := FeatureInfo{
			BranchName: "001-test-feature",
			SpecFile:   "/path/to/specs/001-test-feature/spec.md",
			FeatureNum: "001",
			HasGit:     true,
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := outputFeatureJSON(info)

		w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Errorf("outputFeatureJSON() error = %v", err)
		}

		outputBytes, _ := io.ReadAll(r)
		output := string(outputBytes)

		// Parse JSON to verify it's valid
		var parsedInfo FeatureInfo
		if err := json.Unmarshal(outputBytes, &parsedInfo); err != nil {
			t.Errorf("outputFeatureJSON() produced invalid JSON: %v", err)
		}

		// Verify fields
		if parsedInfo.BranchName != info.BranchName {
			t.Errorf("BranchName = %v, want %v", parsedInfo.BranchName, info.BranchName)
		}
		if parsedInfo.SpecFile != info.SpecFile {
			t.Errorf("SpecFile = %v, want %v", parsedInfo.SpecFile, info.SpecFile)
		}
		if parsedInfo.FeatureNum != info.FeatureNum {
			t.Errorf("FeatureNum = %v, want %v", parsedInfo.FeatureNum, info.FeatureNum)
		}

		// Verify output contains expected fields
		if !strings.Contains(output, "BRANCH_NAME") {
			t.Error("JSON output should contain BRANCH_NAME field")
		}
		if !strings.Contains(output, "SPEC_FILE") {
			t.Error("JSON output should contain SPEC_FILE field")
		}
		if !strings.Contains(output, "FEATURE_NUM") {
			t.Error("JSON output should contain FEATURE_NUM field")
		}
		if !strings.Contains(output, "HAS_GIT") {
			t.Error("JSON output should contain HAS_GIT field")
		}
	})
}

// TestOutputFeatureText tests the outputFeatureText function
func TestOutputFeatureText(t *testing.T) {
	t.Run("valid output", func(t *testing.T) {
		info := FeatureInfo{
			BranchName: "001-test-feature",
			SpecFile:   "/path/to/specs/001-test-feature/spec.md",
			FeatureNum: "001",
			HasGit:     true,
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := outputFeatureText(info)

		w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Errorf("outputFeatureText() error = %v", err)
		}

		outputBytes, _ := io.ReadAll(r)
		output := string(outputBytes)

		// Verify output contains expected information
		expectedStrings := []string{
			"BRANCH_NAME: 001-test-feature",
			"SPEC_FILE: /path/to/specs/001-test-feature/spec.md",
			"FEATURE_NUM: 001",
			"HAS_GIT: true",
			"TCHNCRT_FEATURE environment variable set to: 001-test-feature",
		}

		for _, expected := range expectedStrings {
			if !strings.Contains(output, expected) {
				t.Errorf("outputFeatureText() output missing: %v", expected)
			}
		}
	})
}

// TestRunCreateFeature tests the runCreateFeature command
func TestRunCreateFeature(t *testing.T) {
	t.Run("create feature in git repo", func(t *testing.T) {
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

		origDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		// Reset flags
		jsonOutput = false
		defer func() { jsonOutput = false }()

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err = runCreateFeature(nil, []string{"Test", "Feature"})

		w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Errorf("runCreateFeature() error = %v", err)
		}

		outputBytes, _ := io.ReadAll(r)
		output := string(outputBytes)

		// Verify specs directory was created
		specsDir := filepath.Join(tmpDir, "specs")
		if _, err := os.Stat(specsDir); os.IsNotExist(err) {
			t.Error("Specs directory should be created")
		}

		// Verify feature directory was created
		featureDir := filepath.Join(specsDir, "001-test-feature")
		if _, err := os.Stat(featureDir); os.IsNotExist(err) {
			t.Error("Feature directory should be created")
		}

		// Verify spec file was created
		specFile := filepath.Join(featureDir, "spec.md")
		if _, err := os.Stat(specFile); os.IsNotExist(err) {
			t.Error("Spec file should be created")
		}

		// Verify output
		if !strings.Contains(output, "001-test-feature") {
			t.Error("Output should contain branch name")
		}

		// Verify git branch was created
		cmd = exec.Command("git", "branch", "--list", "001-test-feature")
		cmd.Dir = tmpDir
		branchOutput, err := cmd.Output()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(branchOutput), "001-test-feature") {
			t.Error("Git branch should be created")
		}
	})

	t.Run("create feature without git", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create go.mod to establish repo root
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

		// Reset flags
		jsonOutput = false
		defer func() { jsonOutput = false }()

		// Capture stdout and stderr
		oldStdout := os.Stdout
		oldStderr := os.Stderr
		rOut, wOut, _ := os.Pipe()
		rErr, wErr, _ := os.Pipe()
		os.Stdout = wOut
		os.Stderr = wErr

		err = runCreateFeature(nil, []string{"No", "Git", "Feature"})

		wOut.Close()
		wErr.Close()
		os.Stdout = oldStdout
		os.Stderr = oldStderr

		if err != nil {
			t.Errorf("runCreateFeature() error = %v", err)
		}

		outputBytes, _ := io.ReadAll(rOut)
		output := string(outputBytes)

		stderrBytes, _ := io.ReadAll(rErr)
		stderr := string(stderrBytes)

		// Verify specs directory was created
		specsDir := filepath.Join(tmpDir, "specs")
		if _, err := os.Stat(specsDir); os.IsNotExist(err) {
			t.Error("Specs directory should be created")
		}

		// Verify feature directory was created
		featureDir := filepath.Join(specsDir, "001-no-git-feature")
		if _, err := os.Stat(featureDir); os.IsNotExist(err) {
			t.Error("Feature directory should be created")
		}

		// Verify warning about no git
		if !strings.Contains(stderr, "Warning") || !strings.Contains(stderr, "skipped branch creation") {
			t.Error("Should warn about skipped branch creation when git is not available")
		}

		// Verify output
		if !strings.Contains(output, "001-no-git-feature") {
			t.Error("Output should contain branch name")
		}
	})

	t.Run("create feature with JSON output", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create go.mod
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

		// Set JSON output flag
		jsonOutput = true
		defer func() { jsonOutput = false }()

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err = runCreateFeature(nil, []string{"JSON", "Output"})

		w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Errorf("runCreateFeature() error = %v", err)
		}

		outputBytes, _ := io.ReadAll(r)

		// Verify JSON is valid
		var info FeatureInfo
		if err := json.Unmarshal(outputBytes, &info); err != nil {
			t.Errorf("runCreateFeature() with --json produced invalid JSON: %v", err)
		}

		// Verify JSON fields
		if info.BranchName != "001-json-output" {
			t.Errorf("BranchName = %v, want %v", info.BranchName, "001-json-output")
		}
		if info.FeatureNum != "001" {
			t.Errorf("FeatureNum = %v, want %v", info.FeatureNum, "001")
		}
	})

	t.Run("create multiple features increments number", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create go.mod
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

		// Reset flags
		jsonOutput = false

		// Create first feature
		oldStdout := os.Stdout
		r1, w1, _ := os.Pipe()
		os.Stdout = w1

		err = runCreateFeature(nil, []string{"First", "Feature"})

		w1.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Errorf("First runCreateFeature() error = %v", err)
		}

		io.ReadAll(r1) // Discard output

		// Create second feature
		r2, w2, _ := os.Pipe()
		os.Stdout = w2

		err = runCreateFeature(nil, []string{"Second", "Feature"})

		w2.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Errorf("Second runCreateFeature() error = %v", err)
		}

		io.ReadAll(r2) // Discard output

		// Verify both feature directories exist
		specsDir := filepath.Join(tmpDir, "specs")
		feature1Dir := filepath.Join(specsDir, "001-first-feature")
		feature2Dir := filepath.Join(specsDir, "002-second-feature")

		if _, err := os.Stat(feature1Dir); os.IsNotExist(err) {
			t.Error("First feature directory should be created")
		}
		if _, err := os.Stat(feature2Dir); os.IsNotExist(err) {
			t.Error("Second feature directory should be created with incremented number")
		}
	})

	t.Run("create feature with template", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create go.mod
		goModPath := filepath.Join(tmpDir, "go.mod")
		if err := os.WriteFile(goModPath, []byte("module test\n"), 0644); err != nil {
			t.Fatal(err)
		}

		// Create template
		templateDir := filepath.Join(tmpDir, ".tchncrt", "templates")
		if err := os.MkdirAll(templateDir, 0755); err != nil {
			t.Fatal(err)
		}

		templateContent := "# Feature Specification\n\nThis is a template"
		templatePath := filepath.Join(templateDir, "spec-template.md")
		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
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
		jsonOutput = false
		defer func() { jsonOutput = false }()

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err = runCreateFeature(nil, []string{"Template", "Test"})

		w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Errorf("runCreateFeature() error = %v", err)
		}

		io.ReadAll(r) // Discard output

		// Verify spec file contains template content
		specFile := filepath.Join(tmpDir, "specs", "001-template-test", "spec.md")
		content, err := os.ReadFile(specFile)
		if err != nil {
			t.Fatal(err)
		}

		if string(content) != templateContent {
			t.Errorf("Spec file content = %q, want %q", string(content), templateContent)
		}
	})
}
