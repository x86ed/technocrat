package tchncrt

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FeaturePaths contains all the paths related to a feature branch
type FeaturePaths struct {
	RepoRoot      string
	CurrentBranch string
	HasGit        bool
	FeatureDir    string
	FeatureSpec   string
	ImplPlan      string
	Tasks         string
	Research      string
	DataModel     string
	ContractsDir  string
	Quickstart    string
}

// GetFeaturePaths retrieves all feature-related paths
func GetFeaturePaths() (*FeaturePaths, error) {
	paths := &FeaturePaths{}

	// Get repository root
	repoRoot, err := getRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository root: %w", err)
	}
	paths.RepoRoot = repoRoot

	// Check if we're in a git repository
	paths.HasGit = isGitRepo()

	// Get current branch
	if paths.HasGit {
		branch, err := getCurrentBranch()
		if err != nil {
			return nil, fmt.Errorf("failed to get current branch: %w", err)
		}
		paths.CurrentBranch = branch
	} else {
		paths.CurrentBranch = "main"
	}

	// Build feature paths
	featureName := getFeatureName(paths.CurrentBranch)
	paths.FeatureDir = filepath.Join(repoRoot, ".tchncrt", "features", featureName)
	paths.FeatureSpec = filepath.Join(paths.FeatureDir, "spec.md")
	paths.ImplPlan = filepath.Join(paths.FeatureDir, "plan.md")
	paths.Tasks = filepath.Join(paths.FeatureDir, "tasks.md")
	paths.Research = filepath.Join(paths.FeatureDir, "research.md")
	paths.DataModel = filepath.Join(paths.FeatureDir, "data-model.md")
	paths.ContractsDir = filepath.Join(paths.FeatureDir, "contracts")
	paths.Quickstart = filepath.Join(paths.FeatureDir, "quickstart.md")

	return paths, nil
}

// CheckFeatureBranch validates that we're on a proper feature branch
func CheckFeatureBranch(branch string, hasGit bool) error {
	if !hasGit {
		return fmt.Errorf("not in a git repository")
	}

	if branch == "main" || branch == "master" {
		return fmt.Errorf("cannot run on main/master branch. Please create a feature branch first")
	}

	if branch == "" {
		return fmt.Errorf("could not determine current branch")
	}

	return nil
}

// getRepoRoot finds the repository root directory
func getRepoRoot() (string, error) {
	// Try git root first
	if isGitRepo() {
		cmd := exec.Command("git", "rev-parse", "--show-toplevel")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output)), nil
		}
	}

	// Fall back to current directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return cwd, nil
}

// isGitRepo checks if we're in a git repository
func isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

// getCurrentBranch gets the current git branch name
func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	branch := strings.TrimSpace(string(output))
	if branch == "" {
		return "", fmt.Errorf("empty branch name")
	}

	return branch, nil
}

// getFeatureName extracts feature name from branch name
// Handles common patterns like feature/name, feat/name, or just name
func getFeatureName(branch string) string {
	// Remove common prefixes
	prefixes := []string{"feature/", "feat/", "bugfix/", "fix/", "hotfix/"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(branch, prefix) {
			return strings.TrimPrefix(branch, prefix)
		}
	}

	return branch
}
