package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	setupPlanJSON bool
)

// SetupPlanOutput represents the output of the setup-plan command
type SetupPlanOutput struct {
	FeatureSpec string `json:"FEATURE_SPEC"`
	ImplPlan    string `json:"IMPL_PLAN"`
	SpecsDir    string `json:"SPECS_DIR"`
	Branch      string `json:"BRANCH"`
	HasGit      string `json:"HAS_GIT"`
}

// setupPlanCmd represents the setup-plan command
var setupPlanCmd = &cobra.Command{
	Use:   "setup-plan",
	Short: "Set up the implementation plan for the current feature",
	Long: `Set up the implementation plan for the current feature.

This command:
- Verifies you're on a proper feature branch (XXX-feature-name format)
- Ensures the feature directory exists
- Copies the plan template if available
- Outputs the paths to feature files

The plan template is located at .tchncrt/templates/plan-template.md
If the template doesn't exist, an empty plan.md file will be created.`,
	RunE: runSetupPlan,
}

func init() {
	rootCmd.AddCommand(setupPlanCmd)

	setupPlanCmd.Flags().BoolVar(&setupPlanJSON, "json", false, "Output results in JSON format")
}

func runSetupPlan(cmd *cobra.Command, args []string) error {
	// Get feature paths using the common function
	paths, err := getFeaturePaths("")
	if err != nil {
		return fmt.Errorf("failed to get feature paths: %w", err)
	}

	// Check if we're on a proper feature branch (only for git repos)
	if err := checkFeatureBranch(paths.CurrentBranch, paths.HasGit); err != nil {
		return err
	}

	// Ensure the feature directory exists
	if err := os.MkdirAll(paths.FeatureDir, 0755); err != nil {
		return fmt.Errorf("failed to create feature directory: %w", err)
	}

	// Copy plan template if it exists
	templatePath := filepath.Join(paths.RepoRoot, ".tchncrt", "templates", "plan-template.md")
	if err := copyPlanTemplate(templatePath, paths.ImplPlan); err != nil {
		return fmt.Errorf("failed to set up plan file: %w", err)
	}

	// Output results
	output := SetupPlanOutput{
		FeatureSpec: paths.FeatureSpec,
		ImplPlan:    paths.ImplPlan,
		SpecsDir:    paths.FeatureDir,
		Branch:      paths.CurrentBranch,
		HasGit:      formatBool(paths.HasGit),
	}

	if setupPlanJSON {
		return outputSetupPlanJSON(output)
	}

	return outputSetupPlanText(output)
}

// copyPlanTemplate copies the plan template to the implementation plan file
func copyPlanTemplate(templatePath, destPath string) error {
	// Check if template exists
	if _, err := os.Stat(templatePath); err == nil {
		// Template exists, copy it
		if err := copyFile(templatePath, destPath); err != nil {
			return fmt.Errorf("failed to copy template: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Copied plan template to %s\n", destPath)
	} else {
		// Template doesn't exist, create empty file
		fmt.Fprintf(os.Stderr, "Warning: Plan template not found at %s\n", templatePath)

		// Create a basic plan file if it doesn't exist
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			file, err := os.Create(destPath)
			if err != nil {
				return fmt.Errorf("failed to create plan file: %w", err)
			}
			file.Close()
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Sync to ensure all data is written
	return dstFile.Sync()
}

// formatBool converts a boolean to a string representation
func formatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// outputSetupPlanJSON outputs the result in JSON format
func outputSetupPlanJSON(output SetupPlanOutput) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}

// outputSetupPlanText outputs the result in text format
func outputSetupPlanText(output SetupPlanOutput) error {
	fmt.Printf("FEATURE_SPEC: %s\n", output.FeatureSpec)
	fmt.Printf("IMPL_PLAN: %s\n", output.ImplPlan)
	fmt.Printf("SPECS_DIR: %s\n", output.SpecsDir)
	fmt.Printf("BRANCH: %s\n", output.Branch)
	fmt.Printf("HAS_GIT: %s\n", output.HasGit)
	return nil
}
