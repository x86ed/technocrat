package cmd

import (
	"encoding/json"
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
	jsonOutput bool
)

// FeatureInfo represents the information about a newly created feature
type FeatureInfo struct {
	BranchName string `json:"BRANCH_NAME"`
	SpecFile   string `json:"SPEC_FILE"`
	FeatureNum string `json:"FEATURE_NUM"`
	FeatureDir string `json:"FEATURE_DIR,omitempty"`
	EnvVarSet  bool   `json:"ENV_VAR_SET,omitempty"`
}

// createFeatureCmd represents the create-feature command
var createFeatureCmd = &cobra.Command{
	Use:   "create-feature <feature_description>",
	Short: "Create a new feature branch and specification directory",
	Long: `Create a new feature branch and specification directory.

This command:
- Finds the repository root (with or without git)
- Determines the next feature number based on existing specs
- Creates a feature directory in specs/XXX-feature-name
- Creates a git branch (if git is available)
- Copies the spec template if it exists
- Sets the SET_FEATURE environment variable

Feature branches follow the naming convention: XXX-feature-name
where XXX is a zero-padded 3-digit number.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runCreateFeature,
}

func init() {
	rootCmd.AddCommand(createFeatureCmd)

	createFeatureCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output result in JSON format")
}

func runCreateFeature(cmd *cobra.Command, args []string) error {
	featureDescription := strings.Join(args, " ")

	// Find repository root
	repoRoot, err := findRepoRootForFeature()
	if err != nil {
		return fmt.Errorf("could not determine repository root: %w", err)
	}

	// Check if we have git
	hasGitRepo := hasGit()

	// Ensure specs directory exists
	specsDir := filepath.Join(repoRoot, "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		return fmt.Errorf("failed to create specs directory: %w", err)
	}

	// Find the highest existing feature number
	highestNum, err := findHighestFeatureNumber(specsDir)
	if err != nil {
		return fmt.Errorf("failed to find highest feature number: %w", err)
	}

	// Calculate next feature number
	nextNum := highestNum + 1
	featureNum := fmt.Sprintf("%03d", nextNum)

	// Create branch name from feature description
	branchName := createBranchName(featureDescription, featureNum)

	// Create feature directory
	featureDir := filepath.Join(specsDir, branchName)
	if err := os.MkdirAll(featureDir, 0755); err != nil {
		return fmt.Errorf("failed to create feature directory: %w", err)
	}

	// Create git branch if git is available
	if hasGitRepo {
		if err := createGitBranch(branchName); err != nil {
			return fmt.Errorf("failed to create git branch: %w", err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "[specify] Warning: Git repository not detected; skipped branch creation for %s\n", branchName)
	}

	// Copy template if it exists
	specFile := filepath.Join(featureDir, "spec.md")
	templatePath := filepath.Join(repoRoot, ".tchncrt", "templates", "spec-template.md")
	if err := copyTemplateIfExists(templatePath, specFile); err != nil {
		return fmt.Errorf("failed to copy template: %w", err)
	}

	// Set environment variable (note: this only affects this process and its children)
	os.Setenv("SPECIFY_FEATURE", branchName)

	// Output results
	info := FeatureInfo{
		BranchName: branchName,
		SpecFile:   specFile,
		FeatureNum: featureNum,
		FeatureDir: featureDir,
		EnvVarSet:  true,
	}

	if jsonOutput {
		return outputFeatureJSON(info)
	}

	return outputFeatureText(info)
}

// findRepoRootForFeature finds the repository root by searching for markers
func findRepoRootForFeature() (string, error) {
	// Try git first
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	// Fall back to searching for repository markers
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	dir := cwd
	for {
		// Check for .git or .tchncrt directory
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, ".tchncrt")); err == nil {
			return dir, nil
		}

		// Check for go.mod as a fallback
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return "", fmt.Errorf("could not find repository root (no .git, .tchncrt, or go.mod found)")
		}
		dir = parent
	}
}

// findHighestFeatureNumber finds the highest feature number in the specs directory
func findHighestFeatureNumber(specsDir string) (int, error) {
	highest := 0

	// Check if specs directory exists
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		return highest, nil
	}

	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read specs directory: %w", err)
	}

	// Regular expression to match feature numbers at the start of directory names
	re := regexp.MustCompile(`^(\d+)`)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		matches := re.FindStringSubmatch(entry.Name())
		if len(matches) > 1 {
			num, err := strconv.Atoi(matches[1])
			if err == nil && num > highest {
				highest = num
			}
		}
	}

	return highest, nil
}

// createBranchName creates a branch name from the feature description
func createBranchName(description, featureNum string) string {
	// Convert to lowercase
	name := strings.ToLower(description)

	// Replace non-alphanumeric characters with hyphens
	re := regexp.MustCompile(`[^a-z0-9]+`)
	name = re.ReplaceAllString(name, "-")

	// Remove leading and trailing hyphens
	name = strings.Trim(name, "-")

	// Take only the first 3 words
	words := strings.Split(name, "-")
	if len(words) > 3 {
		words = words[:3]
	}
	name = strings.Join(words, "-")

	// Combine with feature number
	return fmt.Sprintf("%s-%s", featureNum, name)
}

// createGitBranch creates and checks out a new git branch
func createGitBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create and checkout branch: %w", err)
	}

	return nil
}

// copyTemplateIfExists copies the template file if it exists, otherwise creates an empty file
func copyTemplateIfExists(templatePath, destPath string) error {
	// Check if template exists
	if _, err := os.Stat(templatePath); err == nil {
		// Template exists, copy it
		data, err := os.ReadFile(templatePath)
		if err != nil {
			return fmt.Errorf("failed to read template: %w", err)
		}

		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write spec file: %w", err)
		}
	} else {
		// Template doesn't exist, create empty file
		if err := os.WriteFile(destPath, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to create spec file: %w", err)
		}
	}

	return nil
}

// outputFeatureJSON outputs the feature info as JSON
func outputFeatureJSON(info FeatureInfo) error {
	encoder := json.NewEncoder(os.Stdout)
	if err := encoder.Encode(info); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}

// outputFeatureText outputs the feature info as plain text
func outputFeatureText(info FeatureInfo) error {
	fmt.Printf("BRANCH_NAME: %s\n", info.BranchName)
	fmt.Printf("SPEC_FILE: %s\n", info.SpecFile)
	fmt.Printf("FEATURE_NUM: %s\n", info.FeatureNum)
	fmt.Printf("SET_FEATURE environment variable set to: %s\n", info.BranchName)
	return nil
}
