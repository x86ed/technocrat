package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// Flags for common command
	showRepoRoot     bool
	showBranch       bool
	showHasGit       bool
	showFeatureDir   bool
	showFeatureSpec  bool
	showImplPlan     bool
	showTasks        bool
	showResearch     bool
	showDataModel    bool
	showQuickstart   bool
	showContractsDir bool
	showAll          bool
	checkFiles       bool
	validateBranch   bool
	setFeature       string
)

// FeaturePaths represents all the paths related to a feature
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
	Quickstart    string
	ContractsDir  string
}

// commonCmd represents the common command
var commonCmd = &cobra.Command{
	Use:   "common",
	Short: "Common functions for project and feature management",
	Long: `Provides common functionality for managing feature branches and project paths.

This command can:
- Get repository root and current branch information
- Determine feature directory paths based on branch names
- Validate feature branch naming conventions
- Check for the existence of feature files

Feature branches should follow the naming convention: 001-feature-name`,
	RunE: runCommon,
}

func init() {
	rootCmd.AddCommand(commonCmd)

	commonCmd.Flags().BoolVar(&showRepoRoot, "repo-root", false, "Show repository root path")
	commonCmd.Flags().BoolVar(&showBranch, "branch", false, "Show current branch")
	commonCmd.Flags().BoolVar(&showHasGit, "has-git", false, "Check if git repository exists")
	commonCmd.Flags().BoolVar(&showFeatureDir, "feature-dir", false, "Show feature directory path")
	commonCmd.Flags().BoolVar(&showFeatureSpec, "feature-spec", false, "Show feature spec file path")
	commonCmd.Flags().BoolVar(&showImplPlan, "impl-plan", false, "Show implementation plan file path")
	commonCmd.Flags().BoolVar(&showTasks, "tasks", false, "Show tasks file path")
	commonCmd.Flags().BoolVar(&showResearch, "research", false, "Show research file path")
	commonCmd.Flags().BoolVar(&showDataModel, "data-model", false, "Show data model file path")
	commonCmd.Flags().BoolVar(&showQuickstart, "quickstart", false, "Show quickstart file path")
	commonCmd.Flags().BoolVar(&showContractsDir, "contracts-dir", false, "Show contracts directory path")
	commonCmd.Flags().BoolVar(&showAll, "all", false, "Show all paths (default if no flags specified)")
	commonCmd.Flags().BoolVar(&checkFiles, "check", false, "Check existence of feature files and directories")
	commonCmd.Flags().BoolVar(&validateBranch, "validate-branch", false, "Validate feature branch naming convention")
	commonCmd.Flags().StringVar(&setFeature, "set-feature", "", "Override branch detection with specific feature name")
}

func runCommon(cmd *cobra.Command, args []string) error {
	// Get feature paths
	paths, err := getFeaturePaths(setFeature)
	if err != nil {
		return err
	}

	// If validate branch is requested
	if validateBranch {
		if err := checkFeatureBranch(paths.CurrentBranch, paths.HasGit); err != nil {
			return err
		}
		fmt.Println("✓ Branch naming is valid")
		return nil
	}

	// If check files is requested
	if checkFiles {
		return checkFeatureFiles(paths)
	}

	// Determine what to show
	showAnything := showRepoRoot || showBranch || showHasGit || showFeatureDir ||
		showFeatureSpec || showImplPlan || showTasks || showResearch ||
		showDataModel || showQuickstart || showContractsDir

	if showAll || !showAnything {
		printAllPaths(paths)
		return nil
	}

	// Show specific requested values
	if showRepoRoot {
		fmt.Println(paths.RepoRoot)
	}
	if showBranch {
		fmt.Println(paths.CurrentBranch)
	}
	if showHasGit {
		fmt.Println(paths.HasGit)
	}
	if showFeatureDir {
		fmt.Println(paths.FeatureDir)
	}
	if showFeatureSpec {
		fmt.Println(paths.FeatureSpec)
	}
	if showImplPlan {
		fmt.Println(paths.ImplPlan)
	}
	if showTasks {
		fmt.Println(paths.Tasks)
	}
	if showResearch {
		fmt.Println(paths.Research)
	}
	if showDataModel {
		fmt.Println(paths.DataModel)
	}
	if showQuickstart {
		fmt.Println(paths.Quickstart)
	}
	if showContractsDir {
		fmt.Println(paths.ContractsDir)
	}

	return nil
}

// getRepoRoot returns the repository root directory
func getRepoRoot() (string, error) {
	// Try git first
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	// Fall back to working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Try to navigate up to find the project root
	// Look for common indicators like go.mod, .git, etc.
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root, return cwd
			return cwd, nil
		}
		dir = parent
	}
}

// getCurrentBranch returns the current git branch or feature name
func getCurrentBranch(setFeature string) (string, error) {
	// First check if feature is specified
	if setFeature != "" {
		return setFeature, nil
	}

	// Check environment variable (TCHNCRT_FEATURE for cross-platform compatibility)
	if envFeature := os.Getenv("TCHNCRT_FEATURE"); envFeature != "" {
		return envFeature, nil
	}

	// Try git
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	// For non-git repos, try to find the latest feature directory
	repoRoot, err := getRepoRoot()
	if err != nil {
		return "main", nil
	}

	specsDir := filepath.Join(repoRoot, "specs")
	if info, err := os.Stat(specsDir); err == nil && info.IsDir() {
		latestFeature := ""
		highest := 0

		entries, err := os.ReadDir(specsDir)
		if err == nil {
			re := regexp.MustCompile(`^(\d{3})-`)
			for _, entry := range entries {
				if entry.IsDir() {
					matches := re.FindStringSubmatch(entry.Name())
					if len(matches) > 1 {
						num, err := strconv.Atoi(matches[1])
						if err == nil && num > highest {
							highest = num
							latestFeature = entry.Name()
						}
					}
				}
			}
		}

		if latestFeature != "" {
			return latestFeature, nil
		}
	}

	return "main", nil
}

// hasGit checks if we're in a git repository
func hasGit() bool {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	err := cmd.Run()
	return err == nil
}

// checkFeatureBranch validates the feature branch naming convention
func checkFeatureBranch(branch string, hasGitRepo bool) error {
	// For non-git repos, just warn
	if !hasGitRepo {
		fmt.Fprintf(os.Stderr, "[tchncrt] Warning: Git repository not detected; skipped branch validation\n")
		return nil
	}

	// Check if branch follows naming convention
	re := regexp.MustCompile(`^\d{3}-`)
	if !re.MatchString(branch) {
		return fmt.Errorf("ERROR: Not on a feature branch. Current branch: %s\nFeature branches should be named like: 001-feature-name", branch)
	}

	return nil
}

// getFeatureDir returns the feature directory path
func getFeatureDir(repoRoot, branch string) string {
	return filepath.Join(repoRoot, "specs", branch)
}

// getFeaturePaths returns all feature-related paths
func getFeaturePaths(setFeature string) (*FeaturePaths, error) {
	repoRoot, err := getRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to get repo root: %w", err)
	}

	currentBranch, err := getCurrentBranch(setFeature)
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	hasGitRepo := hasGit()
	featureDir := getFeatureDir(repoRoot, currentBranch)

	return &FeaturePaths{
		RepoRoot:      repoRoot,
		CurrentBranch: currentBranch,
		HasGit:        hasGitRepo,
		FeatureDir:    featureDir,
		FeatureSpec:   filepath.Join(featureDir, "spec.md"),
		ImplPlan:      filepath.Join(featureDir, "plan.md"),
		Tasks:         filepath.Join(featureDir, "tasks.md"),
		Research:      filepath.Join(featureDir, "research.md"),
		DataModel:     filepath.Join(featureDir, "data-model.md"),
		Quickstart:    filepath.Join(featureDir, "quickstart.md"),
		ContractsDir:  filepath.Join(featureDir, "contracts"),
	}, nil
}

// printAllPaths prints all paths in the format similar to the original script
func printAllPaths(paths *FeaturePaths) {
	fmt.Printf("REPO_ROOT='%s'\n", paths.RepoRoot)
	fmt.Printf("CURRENT_BRANCH='%s'\n", paths.CurrentBranch)
	fmt.Printf("HAS_GIT='%t'\n", paths.HasGit)
	fmt.Printf("FEATURE_DIR='%s'\n", paths.FeatureDir)
	fmt.Printf("FEATURE_SPEC='%s'\n", paths.FeatureSpec)
	fmt.Printf("IMPL_PLAN='%s'\n", paths.ImplPlan)
	fmt.Printf("TASKS='%s'\n", paths.Tasks)
	fmt.Printf("RESEARCH='%s'\n", paths.Research)
	fmt.Printf("DATA_MODEL='%s'\n", paths.DataModel)
	fmt.Printf("QUICKSTART='%s'\n", paths.Quickstart)
	fmt.Printf("CONTRACTS_DIR='%s'\n", paths.ContractsDir)
}

// checkFeatureFile checks if a file exists and prints status
func checkFeatureFile(path, description string) {
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("  ✓ %s\n", description)
	} else {
		fmt.Printf("  ✗ %s\n", description)
	}
}

// checkFeatureDir checks if a directory exists and is non-empty
func checkFeatureDir(path, description string) {
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		entries, err := os.ReadDir(path)
		if err == nil && len(entries) > 0 {
			fmt.Printf("  ✓ %s\n", description)
			return
		}
	}
	fmt.Printf("  ✗ %s\n", description)
}

// checkFeatureFiles checks the existence of all feature files
func checkFeatureFiles(paths *FeaturePaths) error {
	fmt.Println("Feature Files Status:")
	checkFeatureFile(paths.FeatureSpec, "Feature Specification (spec.md)")
	checkFeatureFile(paths.ImplPlan, "Implementation Plan (plan.md)")
	checkFeatureFile(paths.Tasks, "Tasks (tasks.md)")
	checkFeatureFile(paths.Research, "Research (research.md)")
	checkFeatureFile(paths.DataModel, "Data Model (data-model.md)")
	checkFeatureFile(paths.Quickstart, "Quickstart Guide (quickstart.md)")
	checkFeatureDir(paths.ContractsDir, "Contracts Directory")
	return nil
}
