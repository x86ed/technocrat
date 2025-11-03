package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	jsonMode     bool
	requireTasks bool
	includeTasks bool
	pathsOnly    bool
)

var checkPrerequisitesCmd = &cobra.Command{
	Use:   "check-prerequisites",
	Short: "Check prerequisites for Spec-Driven Development workflow",
	Long: `Consolidated prerequisite checking for Spec-Driven Development workflow.
This command validates that required directories and files exist for the current
feature branch and outputs available documentation.`,
	Example: `  # Check task prerequisites (plan.md required)
  technocrat check-prerequisites --json
  
  # Check implementation prerequisites (plan.md + tasks.md required)
  technocrat check-prerequisites --json --require-tasks --include-tasks
  
  # Get feature paths only (no validation)
  technocrat check-prerequisites --paths-only`,
	RunE: runCheckPrerequisites,
}

func init() {
	rootCmd.AddCommand(checkPrerequisitesCmd)

	checkPrerequisitesCmd.Flags().BoolVar(&jsonMode, "json", false, "Output in JSON format")
	checkPrerequisitesCmd.Flags().BoolVar(&requireTasks, "require-tasks", false, "Require tasks.md to exist (for implementation phase)")
	checkPrerequisitesCmd.Flags().BoolVar(&includeTasks, "include-tasks", false, "Include tasks.md in AVAILABLE_DOCS list")
	checkPrerequisitesCmd.Flags().BoolVar(&pathsOnly, "paths-only", false, "Only output path variables (no validation)")
}

func runCheckPrerequisites(cmd *cobra.Command, args []string) error {
	// Get feature paths using common functionality
	paths, err := getFeaturePaths("")
	if err != nil {
		return fmt.Errorf("failed to get feature paths: %w", err)
	}

	// Validate feature branch
	if err := checkFeatureBranch(paths.CurrentBranch, paths.HasGit); err != nil {
		return err
	}

	// If paths-only mode, output paths and exit
	if pathsOnly {
		if jsonMode {
			return outputPathsJSON(paths)
		}
		return outputPathsText(paths)
	}

	// Validate required directories and files
	if err := validatePrerequisites(paths); err != nil {
		return err
	}

	// Build list of available documents
	docs := getAvailableDocs(paths)

	// Output results
	if jsonMode {
		return outputJSON(paths, docs)
	}
	return outputText(paths, docs)
}

func validatePrerequisites(paths *FeaturePaths) error {
	// Check feature directory exists
	if _, err := os.Stat(paths.FeatureDir); os.IsNotExist(err) {
		return fmt.Errorf("feature directory not found: %s\nRun /tchncrt.spec first to create the feature structure", paths.FeatureDir)
	}

	// Check plan.md exists
	if _, err := os.Stat(paths.ImplPlan); os.IsNotExist(err) {
		return fmt.Errorf("plan.md not found in %s\nRun /tchncrt.plan first to create the implementation plan", paths.FeatureDir)
	}

	// Check tasks.md if required
	if requireTasks {
		if _, err := os.Stat(paths.Tasks); os.IsNotExist(err) {
			return fmt.Errorf("tasks.md not found in %s\nRun /tchncrt.tasks first to create the task list", paths.FeatureDir)
		}
	}

	return nil
}

func getAvailableDocs(paths *FeaturePaths) []string {
	var docs []string

	// Check optional documents
	if fileExists(paths.Research) {
		docs = append(docs, "research.md")
	}

	if fileExists(paths.DataModel) {
		docs = append(docs, "data-model.md")
	}

	// Check contracts directory
	if dirHasFiles(paths.ContractsDir) {
		docs = append(docs, "contracts/")
	}

	if fileExists(paths.Quickstart) {
		docs = append(docs, "quickstart.md")
	}

	// Include tasks.md if requested and it exists
	if includeTasks && fileExists(paths.Tasks) {
		docs = append(docs, "tasks.md")
	}

	return docs
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirHasFiles(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	return len(entries) > 0
}

func outputPathsJSON(paths *FeaturePaths) error {
	output := map[string]string{
		"REPO_ROOT":    paths.RepoRoot,
		"BRANCH":       paths.CurrentBranch,
		"FEATURE_DIR":  paths.FeatureDir,
		"FEATURE_SPEC": paths.FeatureSpec,
		"IMPL_PLAN":    paths.ImplPlan,
		"TASKS":        paths.Tasks,
	}

	// Output compact JSON (no indentation) to match shell scripts
	encoder := json.NewEncoder(os.Stdout)
	return encoder.Encode(output)
}

func outputPathsText(paths *FeaturePaths) error {
	fmt.Printf("REPO_ROOT: %s\n", paths.RepoRoot)
	fmt.Printf("BRANCH: %s\n", paths.CurrentBranch)
	fmt.Printf("FEATURE_DIR: %s\n", paths.FeatureDir)
	fmt.Printf("FEATURE_SPEC: %s\n", paths.FeatureSpec)
	fmt.Printf("IMPL_PLAN: %s\n", paths.ImplPlan)
	fmt.Printf("TASKS: %s\n", paths.Tasks)
	return nil
}

func outputJSON(paths *FeaturePaths, docs []string) error {
	output := map[string]interface{}{
		"FEATURE_DIR":    paths.FeatureDir,
		"AVAILABLE_DOCS": docs,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputText(paths *FeaturePaths, _ []string) error {
	fmt.Printf("FEATURE_DIR:%s\n", paths.FeatureDir)
	fmt.Println("AVAILABLE_DOCS:")

	// Show status of each potential document
	checkFile(paths.Research, "research.md")
	checkFile(paths.DataModel, "data-model.md")
	checkDir(paths.ContractsDir, "contracts/")
	checkFile(paths.Quickstart, "quickstart.md")

	if includeTasks {
		checkFile(paths.Tasks, "tasks.md")
	}

	return nil
}

func checkFile(path, name string) {
	if fileExists(path) {
		fmt.Printf("  ✓ %s\n", name)
	} else {
		fmt.Printf("  ✗ %s\n", name)
	}
}

func checkDir(path, name string) {
	if dirHasFiles(path) {
		fmt.Printf("  ✓ %s\n", name)
	} else {
		fmt.Printf("  ✗ %s\n", name)
	}
}
