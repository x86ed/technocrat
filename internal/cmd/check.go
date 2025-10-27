package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"technocrat/internal/ui"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check that required tools are installed",
	Long: `Check verifies that common development tools and AI assistants are installed.
	
This command checks for:
  - Git version control
  - AI assistant CLIs (claude, gemini, qwen, opencode, codex)
  - VS Code and variants (code, cursor, windsurf)`,
	RunE: runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	// Show banner
	showBanner()

	// Create tracker for tool checks
	tracker := ui.NewStepTracker("Tool Availability Check")

	// Define tools to check
	tools := []struct {
		key         string
		name        string
		command     string
		required    bool
		category    string
		specialPath string // For tools like claude that may be in special locations
	}{
		// Version control
		{"git", "Git", "git", true, "version-control", ""},

		// AI Assistants with CLI
		{"claude", "Claude CLI", "claude", false, "ai-assistant", filepath.Join(os.Getenv("HOME"), ".claude", "local", "claude")},
		{"gemini", "Gemini CLI", "gemini", false, "ai-assistant", ""},
		{"qwen", "Qwen Code", "qwen", false, "ai-assistant", ""},
		{"opencode", "OpenCode CLI", "opencode", false, "ai-assistant", ""},
		{"codex", "Codex CLI", "codex", false, "ai-assistant", ""},

		// Editors
		{"vscode", "VS Code", "code", false, "editor", ""},
		{"vscode-insiders", "VS Code Insiders", "code-insiders", false, "editor", ""},
		{"cursor", "Cursor", "cursor", false, "editor", ""},
		{"windsurf", "Windsurf", "windsurf", false, "editor", ""},
	}

	// Add all tools to tracker
	for _, tool := range tools {
		tracker.Add(tool.key, tool.name)
	}

	// Start live rendering
	if err := tracker.StartLive(); err != nil && debug {
		fmt.Fprintf(os.Stderr, "Debug: Could not start live tracker: %v\n", err)
	}

	// Show progress message for non-TTY mode
	if !ui.IsInteractive() {
		fmt.Fprintln(os.Stderr, "\nChecking tool availability...")
	}

	// Check each tool
	var missingRequired []string
	var missingOptional []string

	for _, tool := range tools {
		tracker.Start(tool.key, "Checking...")

		installed := false

		// Check special path first if defined
		if tool.specialPath != "" {
			if _, err := os.Stat(tool.specialPath); err == nil {
				installed = true
			}
		}

		// Check PATH if not found in special location
		if !installed {
			if _, err := exec.LookPath(tool.command); err == nil {
				installed = true
			}
		}

		if installed {
			// Try to get version
			version := getToolVersion(tool.command, tool.specialPath)
			if version != "" {
				tracker.Complete(tool.key, version)
				if !ui.IsInteractive() {
					fmt.Fprintf(os.Stderr, "  ✓ %s: %s\n", tool.name, version)
				}
			} else {
				tracker.Complete(tool.key, "Installed")
				if !ui.IsInteractive() {
					fmt.Fprintf(os.Stderr, "  ✓ %s: Installed\n", tool.name)
				}
			}
		} else {
			if tool.required {
				tracker.Error(tool.key, "Not found")
				missingRequired = append(missingRequired, tool.name)
				if !ui.IsInteractive() {
					fmt.Fprintf(os.Stderr, "  ✗ %s: Not found\n", tool.name)
				}
			} else {
				tracker.Skip(tool.key, "Not installed")
				missingOptional = append(missingOptional, tool.name)
				if !ui.IsInteractive() {
					fmt.Fprintf(os.Stderr, "  ○ %s: Not installed\n", tool.name)
				}
			}
		}
	}

	// Stop live rendering
	tracker.StopLive()

	// Show summary
	fmt.Fprintf(os.Stderr, "\n%s %s\n", ui.ColorGreen.Sprint(ui.SymbolCheckmark), tracker.Summary())

	// Show missing tools if any
	if len(missingRequired) > 0 {
		msg := "The following required tools are missing:\n\n"
		for _, tool := range missingRequired {
			msg += fmt.Sprintf("  • %s\n", tool)
		}
		msg += "\nPlease install these tools before continuing."
		ui.ShowError("Missing Required Tools", msg)
		return fmt.Errorf("missing required tools")
	}

	if len(missingOptional) > 0 {
		msg := "The following optional tools are not installed:\n\n"
		for _, tool := range missingOptional {
			msg += fmt.Sprintf("  • %s\n", tool)
		}
		msg += "\nThese are optional but enhance your development experience."
		ui.ShowInfo("Optional Tools Not Found", msg)
	}

	// Show installation tips
	if len(missingRequired) > 0 || len(missingOptional) > 0 {
		showInstallationTips(missingRequired, missingOptional)
	}

	if len(missingRequired) == 0 {
		ui.ShowSuccess("All Set!", "All required tools are installed and ready to use.")
	}

	return nil
}

func getToolVersion(command, specialPath string) string {
	var cmd *exec.Cmd

	// Use special path if provided
	if specialPath != "" {
		if _, err := os.Stat(specialPath); err == nil {
			cmd = exec.Command(specialPath, "--version")
		}
	}

	// Fallback to PATH
	if cmd == nil {
		cmd = exec.Command(command, "--version")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try -v flag
		if specialPath != "" {
			if _, err := os.Stat(specialPath); err == nil {
				cmd = exec.Command(specialPath, "-v")
			}
		} else {
			cmd = exec.Command(command, "-v")
		}
		output, err = cmd.CombinedOutput()
		if err != nil {
			// Try version subcommand
			if specialPath != "" {
				if _, err := os.Stat(specialPath); err == nil {
					cmd = exec.Command(specialPath, "version")
				}
			} else {
				cmd = exec.Command(command, "version")
			}
			output, err = cmd.CombinedOutput()
			if err != nil {
				return ""
			}
		}
	}

	// Parse version from output (take first line)
	lines := string(output)
	if len(lines) > 0 {
		// Limit to first 50 characters for display
		if len(lines) > 50 {
			return lines[:50] + "..."
		}
		// Find first newline
		for i, c := range lines {
			if c == '\n' {
				return lines[:i]
			}
		}
		return lines
	}

	return ""
}

func showInstallationTips(missingRequired, missingOptional []string) {
	allMissing := append(missingRequired, missingOptional...)
	sort.Strings(allMissing)

	tips := make(map[string]string)

	// Installation tips for each tool
	tips["Git"] = "https://git-scm.com/downloads"
	tips["Claude CLI"] = "https://docs.anthropic.com/en/docs/claude-code/setup"
	tips["Gemini CLI"] = "https://github.com/google-gemini/gemini-cli"
	tips["Qwen Code"] = "https://github.com/QwenLM/qwen-code"
	tips["OpenCode CLI"] = "https://opencode.ai"
	tips["Codex CLI"] = "https://github.com/openai/codex"
	tips["VS Code"] = "https://code.visualstudio.com/download"
	tips["Cursor"] = "https://cursor.sh"
	tips["Windsurf"] = "https://codeium.com/windsurf"

	msg := "Installation resources:\n\n"
	for _, tool := range allMissing {
		if url, ok := tips[tool]; ok {
			msg += fmt.Sprintf("  • %s\n    %s\n\n", tool, ui.ColorDim.Sprint(url))
		}
	}

	ui.ShowInfo("Installation Tips", msg)
}
