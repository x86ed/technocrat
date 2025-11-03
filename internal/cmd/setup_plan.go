package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"technocrat/internal/templates"

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
	HasGit      bool   `json:"HAS_GIT"`
}

// setupPlanCmd represents the setup-plan command
var setupPlanCmd = &cobra.Command{
	Use:   "setup-plan",
	Short: "Set up the implementation plan for the current feature",
	Long: `Set up the implementation plan for the current feature.

This command:
- Verifies you're on a proper feature branch (XXX-feature-name format)
- Ensures the feature directory exists
- Creates plan.md from embedded template
- Outputs the paths to feature files

The plan template is embedded in the technocrat binary.`,
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

	// Copy plan template from embedded filesystem
	if err := copyPlanTemplate(paths.ImplPlan); err != nil {
		return fmt.Errorf("failed to set up plan file: %w", err)
	}

	// Output results
	output := SetupPlanOutput{
		FeatureSpec: paths.FeatureSpec,
		ImplPlan:    paths.ImplPlan,
		SpecsDir:    paths.FeatureDir,
		Branch:      paths.CurrentBranch,
		HasGit:      paths.HasGit,
	}

	if setupPlanJSON {
		return outputSetupPlanJSON(output)
	}

	return outputSetupPlanText(output)
}

// copyPlanTemplate copies the plan template from embedded filesystem to the destination
func copyPlanTemplate(destPath string) error {
	// Get plan template from embedded filesystem
	planData, err := templates.GetTemplate("plan-template.md")
	if err != nil {
		return fmt.Errorf("failed to get plan template: %w", err)
	}

	// Write template to destination
	if err := os.WriteFile(destPath, planData, 0644); err != nil {
		return fmt.Errorf("failed to write plan file: %w", err)
	}

	// Only print message if not in JSON mode
	if !setupPlanJSON {
		fmt.Printf("Copied plan template to %s\n", destPath)
	}
	return nil
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
	fmt.Printf("HAS_GIT: %s\n", formatBool(output.HasGit))
	return nil
}
