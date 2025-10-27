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

// TestFormatBool tests the formatBool helper function
func TestFormatBool(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected string
	}{
		{
			name:     "true value",
			input:    true,
			expected: "true",
		},
		{
			name:     "false value",
			input:    false,
			expected: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBool(tt.input)
			if result != tt.expected {
				t.Errorf("formatBool(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestCopyFile tests the copyFile function
func TestCopyFile(t *testing.T) {
	t.Run("successful copy", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create source file with content
		srcPath := filepath.Join(tmpDir, "source.txt")
		content := "test content\nmultiline\n"
		if err := os.WriteFile(srcPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		// Copy to destination
		dstPath := filepath.Join(tmpDir, "dest.txt")
		if err := copyFile(srcPath, dstPath); err != nil {
			t.Errorf("copyFile() error = %v", err)
		}

		// Verify destination file content
		dstContent, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatal(err)
		}

		if string(dstContent) != content {
			t.Errorf("copied content = %v, want %v", string(dstContent), content)
		}
	})

	t.Run("source file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		srcPath := filepath.Join(tmpDir, "nonexistent.txt")
		dstPath := filepath.Join(tmpDir, "dest.txt")

		err := copyFile(srcPath, dstPath)
		if err == nil {
			t.Error("copyFile() expected error for nonexistent source file")
		}
	})

	t.Run("destination directory does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create source file
		srcPath := filepath.Join(tmpDir, "source.txt")
		if err := os.WriteFile(srcPath, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}

		// Try to copy to nonexistent directory
		dstPath := filepath.Join(tmpDir, "nonexistent", "dest.txt")
		err := copyFile(srcPath, dstPath)
		if err == nil {
			t.Error("copyFile() expected error for nonexistent destination directory")
		}
	})
}

// TestCopyPlanTemplate tests the copyPlanTemplate function
func TestCopyPlanTemplate(t *testing.T) {
	t.Run("template exists and is copied", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create template file
		templatePath := filepath.Join(tmpDir, "plan-template.md")
		templateContent := "# Plan Template\n\nContent here"
		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatal(err)
		}

		destPath := filepath.Join(tmpDir, "plan.md")

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := copyPlanTemplate(templatePath, destPath)

		w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Errorf("copyPlanTemplate() error = %v", err)
		}

		// Verify destination file exists and has correct content
		destContent, err := os.ReadFile(destPath)
		if err != nil {
			t.Fatal(err)
		}

		if string(destContent) != templateContent {
			t.Errorf("copied content = %v, want %v", string(destContent), templateContent)
		}

		// Verify stdout message
		outputBytes, _ := io.ReadAll(r)
		output := string(outputBytes)
		if !strings.Contains(output, "Copied plan template to") {
			t.Error("copyPlanTemplate() should output success message to stdout")
		}
	})

	t.Run("template does not exist, creates empty file", func(t *testing.T) {
		tmpDir := t.TempDir()

		templatePath := filepath.Join(tmpDir, "nonexistent-template.md")
		destPath := filepath.Join(tmpDir, "plan.md")

		// Capture stderr
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		err := copyPlanTemplate(templatePath, destPath)

		w.Close()
		os.Stderr = oldStderr

		if err != nil {
			t.Errorf("copyPlanTemplate() error = %v", err)
		}

		// Verify destination file exists and is empty
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			t.Error("copyPlanTemplate() should create empty file when template doesn't exist")
		}

		destContent, _ := os.ReadFile(destPath)
		if len(destContent) != 0 {
			t.Error("copyPlanTemplate() should create empty file")
		}

		// Verify stderr message
		outputBytes, _ := io.ReadAll(r)
		output := string(outputBytes)
		if !strings.Contains(output, "Warning: Plan template not found") {
			t.Error("copyPlanTemplate() should output warning to stderr when template doesn't exist")
		}
	})

	t.Run("overwrites existing destination file", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create template file
		templatePath := filepath.Join(tmpDir, "plan-template.md")
		newContent := "# New Plan Template"
		if err := os.WriteFile(templatePath, []byte(newContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Create existing destination file with different content
		destPath := filepath.Join(tmpDir, "plan.md")
		oldContent := "# Old Plan Content"
		if err := os.WriteFile(destPath, []byte(oldContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := copyPlanTemplate(templatePath, destPath)

		w.Close()
		os.Stdout = oldStdout
		io.ReadAll(r) // Drain pipe

		if err != nil {
			t.Errorf("copyPlanTemplate() error = %v", err)
		}

		// Verify destination file has new content
		destContent, err := os.ReadFile(destPath)
		if err != nil {
			t.Fatal(err)
		}

		if string(destContent) != newContent {
			t.Errorf("copied content = %v, want %v", string(destContent), newContent)
		}
	})
}

// TestOutputSetupPlanJSON tests the outputSetupPlanJSON function
func TestOutputSetupPlanJSON(t *testing.T) {
	t.Run("valid JSON output", func(t *testing.T) {
		output := SetupPlanOutput{
			FeatureSpec: "/path/to/specs/001-test/spec.md",
			ImplPlan:    "/path/to/specs/001-test/plan.md",
			SpecsDir:    "/path/to/specs/001-test",
			Branch:      "001-test-feature",
			HasGit:      true,
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := outputSetupPlanJSON(output)

		w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Errorf("outputSetupPlanJSON() error = %v", err)
		}

		outputBytes, _ := io.ReadAll(r)
		outputStr := string(outputBytes)

		// Verify it's valid JSON
		var result SetupPlanOutput
		if err := json.Unmarshal(outputBytes, &result); err != nil {
			t.Errorf("outputSetupPlanJSON() produced invalid JSON: %v", err)
		}

		// Verify fields
		if result.FeatureSpec != output.FeatureSpec {
			t.Errorf("FeatureSpec = %v, want %v", result.FeatureSpec, output.FeatureSpec)
		}
		if result.ImplPlan != output.ImplPlan {
			t.Errorf("ImplPlan = %v, want %v", result.ImplPlan, output.ImplPlan)
		}
		if result.SpecsDir != output.SpecsDir {
			t.Errorf("SpecsDir = %v, want %v", result.SpecsDir, output.SpecsDir)
		}
		if result.Branch != output.Branch {
			t.Errorf("Branch = %v, want %v", result.Branch, output.Branch)
		}
		if result.HasGit != output.HasGit {
			t.Errorf("HasGit = %v, want %v", result.HasGit, output.HasGit)
		}

		// Verify JSON is compact (single line)
		if strings.Count(outputStr, "\n") > 1 {
			t.Error("outputSetupPlanJSON() should produce compact JSON (single line)")
		}
	})

	t.Run("HasGit as boolean in JSON", func(t *testing.T) {
		tests := []struct {
			name   string
			hasGit bool
		}{
			{"HasGit true", true},
			{"HasGit false", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				output := SetupPlanOutput{
					FeatureSpec: "/path/to/spec.md",
					ImplPlan:    "/path/to/plan.md",
					SpecsDir:    "/path/to/specs",
					Branch:      "001-test",
					HasGit:      tt.hasGit,
				}

				// Capture stdout
				oldStdout := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w

				outputSetupPlanJSON(output)

				w.Close()
				os.Stdout = oldStdout

				outputBytes, _ := io.ReadAll(r)
				outputStr := string(outputBytes)

				// Verify HasGit is a boolean, not a string
				expectedBool := "true"
				if !tt.hasGit {
					expectedBool = "false"
				}
				if strings.Contains(outputStr, `"HAS_GIT":"`+expectedBool+`"`) {
					t.Error("HasGit should be boolean in JSON, not string")
				}
				if !strings.Contains(outputStr, `"HAS_GIT":`+expectedBool) {
					t.Error("HasGit should be boolean in JSON")
				}
			})
		}
	})
}

// TestOutputSetupPlanText tests the outputSetupPlanText function
func TestOutputSetupPlanText(t *testing.T) {
	t.Run("valid text output", func(t *testing.T) {
		output := SetupPlanOutput{
			FeatureSpec: "/path/to/specs/001-test/spec.md",
			ImplPlan:    "/path/to/specs/001-test/plan.md",
			SpecsDir:    "/path/to/specs/001-test",
			Branch:      "001-test-feature",
			HasGit:      true,
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := outputSetupPlanText(output)

		w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Errorf("outputSetupPlanText() error = %v", err)
		}

		outputBytes, _ := io.ReadAll(r)
		outputStr := string(outputBytes)

		// Verify all expected fields are present
		expectedStrings := []string{
			"FEATURE_SPEC: /path/to/specs/001-test/spec.md",
			"IMPL_PLAN: /path/to/specs/001-test/plan.md",
			"SPECS_DIR: /path/to/specs/001-test",
			"BRANCH: 001-test-feature",
			"HAS_GIT: true",
		}

		for _, expected := range expectedStrings {
			if !strings.Contains(outputStr, expected) {
				t.Errorf("outputSetupPlanText() missing expected string: %v", expected)
			}
		}
	})

	t.Run("HasGit as string in text output", func(t *testing.T) {
		tests := []struct {
			name     string
			hasGit   bool
			expected string
		}{
			{"HasGit true", true, "HAS_GIT: true"},
			{"HasGit false", false, "HAS_GIT: false"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				output := SetupPlanOutput{
					FeatureSpec: "/path/to/spec.md",
					ImplPlan:    "/path/to/plan.md",
					SpecsDir:    "/path/to/specs",
					Branch:      "001-test",
					HasGit:      tt.hasGit,
				}

				// Capture stdout
				oldStdout := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w

				outputSetupPlanText(output)

				w.Close()
				os.Stdout = oldStdout

				outputBytes, _ := io.ReadAll(r)
				outputStr := string(outputBytes)

				if !strings.Contains(outputStr, tt.expected) {
					t.Errorf("outputSetupPlanText() should contain %v", tt.expected)
				}
			})
		}
	})
}

// TestRunSetupPlan tests the runSetupPlan command
func TestRunSetupPlan(t *testing.T) {
	// Prevent subtests from running in parallel since they modify global state (current directory)
	// This is necessary because os.Chdir affects the entire process
	t.Run("setup plan in git repo", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Initialize git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Skipf("Skipping test: git not available: %v", err)
		}

		// Configure git to avoid errors
		cmd = exec.Command("git", "config", "user.email", "test@example.com")
		cmd.Dir = tmpDir
		cmd.Run()

		cmd = exec.Command("git", "config", "user.name", "Test User")
		cmd.Dir = tmpDir
		cmd.Run()

		// Create an initial commit so we can create branches
		testFile := filepath.Join(tmpDir, "test.txt")
		os.WriteFile(testFile, []byte("test"), 0644)

		cmd = exec.Command("git", "add", ".")
		cmd.Dir = tmpDir
		cmd.Run()

		cmd = exec.Command("git", "commit", "-m", "initial")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Skipf("Skipping test: git commit failed: %v", err)
		}

		// Create a feature branch
		cmd = exec.Command("git", "checkout", "-b", "001-test-feature")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
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

		// Create .tchncrt directory structure
		tchncrtDir := filepath.Join(tmpDir, ".tchncrt", "templates")
		if err := os.MkdirAll(tchncrtDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create plan template
		templatePath := filepath.Join(tchncrtDir, "plan-template.md")
		templateContent := "# Implementation Plan\n\n## Overview"
		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Run command
		setupPlanJSON = false
		err = runSetupPlan(nil, []string{})
		if err != nil {
			t.Errorf("runSetupPlan() error = %v", err)
		}

		// Verify plan file was created
		planPath := filepath.Join(tmpDir, "specs", "001-test-feature", "plan.md")
		if _, err := os.Stat(planPath); os.IsNotExist(err) {
			t.Error("runSetupPlan() should create plan.md file")
		}

		// Verify content
		content, _ := os.ReadFile(planPath)
		if string(content) != templateContent {
			t.Errorf("plan.md content = %v, want %v", string(content), templateContent)
		}
	})

	t.Run("setup plan without git", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create specs directory with feature branch structure
		featureDir := filepath.Join(tmpDir, "specs", "001-test-feature")
		if err := os.MkdirAll(featureDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create .tchncrt marker
		tchncrtDir := filepath.Join(tmpDir, ".tchncrt", "templates")
		if err := os.MkdirAll(tchncrtDir, 0755); err != nil {
			t.Fatal(err)
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

		// Set TCHNCRT_FEATURE environment variable to ensure we use the correct branch
		oldEnv := os.Getenv("TCHNCRT_FEATURE")
		os.Setenv("TCHNCRT_FEATURE", "001-test-feature")
		defer os.Setenv("TCHNCRT_FEATURE", oldEnv)

		// Run command
		setupPlanJSON = false
		err = runSetupPlan(nil, []string{})
		if err != nil {
			t.Errorf("runSetupPlan() error = %v", err)
		}

		// Verify plan file was created
		planPath := filepath.Join(featureDir, "plan.md")
		if _, err := os.Stat(planPath); os.IsNotExist(err) {
			t.Error("runSetupPlan() should create plan.md file")
		}
	})

	t.Run("setup plan with JSON output", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Initialize git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Skipf("Skipping test: git not available: %v", err)
		}

		// Configure git
		cmd = exec.Command("git", "config", "user.email", "test@example.com")
		cmd.Dir = tmpDir
		cmd.Run()

		cmd = exec.Command("git", "config", "user.name", "Test User")
		cmd.Dir = tmpDir
		cmd.Run()

		// Create initial commit
		testFile := filepath.Join(tmpDir, "test.txt")
		os.WriteFile(testFile, []byte("test"), 0644)

		cmd = exec.Command("git", "add", ".")
		cmd.Dir = tmpDir
		cmd.Run()

		cmd = exec.Command("git", "commit", "-m", "initial")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Skipf("Skipping test: git commit failed: %v", err)
		}

		// Create a feature branch
		cmd = exec.Command("git", "checkout", "-b", "002-json-test")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
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

		// Create .tchncrt directory
		tchncrtDir := filepath.Join(tmpDir, ".tchncrt", "templates")
		if err := os.MkdirAll(tchncrtDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Set TCHNCRT_FEATURE environment variable
		oldEnv := os.Getenv("TCHNCRT_FEATURE")
		os.Setenv("TCHNCRT_FEATURE", "002-json-test")
		defer os.Setenv("TCHNCRT_FEATURE", oldEnv)

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Run command with JSON flag
		setupPlanJSON = true
		err = runSetupPlan(nil, []string{})

		w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Errorf("runSetupPlan() error = %v", err)
		}

		// Verify JSON output
		outputBytes, _ := io.ReadAll(r)
		var result SetupPlanOutput
		if err := json.Unmarshal(outputBytes, &result); err != nil {
			t.Errorf("runSetupPlan() produced invalid JSON: %v", err)
		}

		if result.Branch != "002-json-test" {
			t.Errorf("Branch = %v, want 002-json-test", result.Branch)
		}

		if !result.HasGit {
			t.Error("HasGit should be true in git repo")
		}
	})

	t.Run("invalid branch name", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Initialize git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Skipf("Skipping test: git not available: %v", err)
		}

		// Stay on main/master branch (invalid for setup-plan)
		// Change to temp directory
		origDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		// Create .tchncrt directory
		tchncrtDir := filepath.Join(tmpDir, ".tchncrt")
		if err := os.MkdirAll(tchncrtDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Set TCHNCRT_FEATURE to invalid branch name (main)
		oldEnv := os.Getenv("TCHNCRT_FEATURE")
		os.Setenv("TCHNCRT_FEATURE", "main")
		defer os.Setenv("TCHNCRT_FEATURE", oldEnv)

		// Run command - should fail
		setupPlanJSON = false
		err = runSetupPlan(nil, []string{})
		if err == nil {
			t.Error("runSetupPlan() should error on invalid branch name")
		} else if !strings.Contains(err.Error(), "Not on a feature branch") {
			t.Errorf("runSetupPlan() error should mention invalid branch, got: %v", err)
		}
	})

	t.Run("creates feature directory if missing", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Initialize git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Skipf("Skipping test: git not available: %v", err)
		}

		// Configure git
		cmd = exec.Command("git", "config", "user.email", "test@example.com")
		cmd.Dir = tmpDir
		cmd.Run()

		cmd = exec.Command("git", "config", "user.name", "Test User")
		cmd.Dir = tmpDir
		cmd.Run()

		// Create initial commit
		testFile := filepath.Join(tmpDir, "test.txt")
		os.WriteFile(testFile, []byte("test"), 0644)

		cmd = exec.Command("git", "add", ".")
		cmd.Dir = tmpDir
		cmd.Run()

		cmd = exec.Command("git", "commit", "-m", "initial")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Skipf("Skipping test: git commit failed: %v", err)
		}

		// Create a feature branch
		cmd = exec.Command("git", "checkout", "-b", "003-new-feature")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
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

		// Create .tchncrt directory
		tchncrtDir := filepath.Join(tmpDir, ".tchncrt", "templates")
		if err := os.MkdirAll(tchncrtDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Set TCHNCRT_FEATURE environment variable
		oldEnv := os.Getenv("TCHNCRT_FEATURE")
		os.Setenv("TCHNCRT_FEATURE", "003-new-feature")
		defer os.Setenv("TCHNCRT_FEATURE", oldEnv)

		// Verify specs directory doesn't exist yet
		specsDir := filepath.Join(tmpDir, "specs", "003-new-feature")
		if _, err := os.Stat(specsDir); !os.IsNotExist(err) {
			t.Fatal("specs directory should not exist before running command")
		}

		// Run command
		setupPlanJSON = false
		err = runSetupPlan(nil, []string{})
		if err != nil {
			t.Errorf("runSetupPlan() error = %v", err)
		}

		// Verify specs directory was created
		if _, err := os.Stat(specsDir); os.IsNotExist(err) {
			t.Error("runSetupPlan() should create feature directory")
		}
	})
}
