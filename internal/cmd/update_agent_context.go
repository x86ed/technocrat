package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// AgentType represents the type of AI agent
type AgentType string

const (
	AgentClaude    AgentType = "claude"
	AgentGemini    AgentType = "gemini"
	AgentCopilot   AgentType = "copilot"
	AgentCursor    AgentType = "cursor"
	AgentQwen      AgentType = "qwen"
	AgentOpenCode  AgentType = "opencode"
	AgentCodex     AgentType = "codex"
	AgentWindsurf  AgentType = "windsurf"
	AgentKiloCode  AgentType = "kilocode"
	AgentAuggie    AgentType = "auggie"
	AgentRoo       AgentType = "roo"
	AgentCodeBuddy AgentType = "codebuddy"
	AgentQ         AgentType = "q"
)

// AgentFileConfig holds the configuration for each agent type
type AgentFileConfig struct {
	Path string
	Name string
}

// PlanData holds extracted information from plan.md
type PlanData struct {
	Language    string
	Framework   string
	Database    string
	ProjectType string
}

var (
	agentType string
)

// updateAgentContextCmd represents the update-agent-context command
var updateAgentContextCmd = &cobra.Command{
	Use:   "update-agent-context [agent-type]",
	Short: "Update agent context files with information from plan.md",
	Long: `Update agent context files with information from plan.md.

This command maintains AI agent context files by parsing feature specifications
and updating agent-specific configuration files with project information.

Supported agent types:
  claude     - Claude Code (CLAUDE.md)
  gemini     - Gemini CLI (GEMINI.md)
  copilot    - GitHub Copilot (.github/copilot-instructions.md)
  cursor     - Cursor IDE (.cursor/rules/tchncrt-rules.mdc)
  qwen       - Qwen Code (QWEN.md)
  opencode   - opencode (AGENTS.md)
  codex      - Codex CLI (AGENTS.md)
  windsurf   - Windsurf (.windsurf/rules/tchncrt-rules.md)
  kilocode   - Kilo Code (.kilocode/rules/tchncrt-rules.md)
  auggie     - Auggie CLI (.augment/rules/tchncrt-rules.md)
  roo        - Roo Code (.roo/rules/tchncrt-rules.md)
  codebuddy  - CodeBuddy (.codebuddy/rules/tchncrt-rules.md)
  q          - Amazon Q Developer CLI (AGENTS.md)

If no agent type is specified, all existing agent files will be updated.
If no agent files exist, a default Claude file will be created.`,
	ValidArgs: []string{"claude", "gemini", "copilot", "cursor", "qwen", "opencode", "codex", "windsurf", "kilocode", "auggie", "roo", "codebuddy", "q"},
	RunE:      runUpdateAgentContext,
}

func init() {
	rootCmd.AddCommand(updateAgentContextCmd)
}

// Logging helper functions for consistent output
func logInfo(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "INFO: "+format+"\n", args...)
}

func logSuccess(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "✓ "+format+"\n", args...)
}

func logWarning(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "WARNING: "+format+"\n", args...)
}

func logError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
}

func runUpdateAgentContext(cmd *cobra.Command, args []string) error {
	// Get feature paths
	paths, err := getFeaturePaths("")
	if err != nil {
		return fmt.Errorf("failed to get feature paths: %w", err)
	}

	// Validate environment
	if paths.CurrentBranch == "" {
		return fmt.Errorf("unable to determine current feature")
	}

	logInfo("=== Updating agent context files for feature %s ===", paths.CurrentBranch)

	// Check if plan.md exists
	if _, err := os.Stat(paths.ImplPlan); os.IsNotExist(err) {
		return fmt.Errorf("no plan.md found at %s\n\nMake sure you're working on a feature with a corresponding spec directory.\nYou may need to run 'technocrat setup-plan' first", paths.ImplPlan)
	}

	// Parse plan data
	logInfo("Parsing plan data from %s", paths.ImplPlan)
	planData, err := parsePlanData(paths.ImplPlan)
	if err != nil {
		return fmt.Errorf("failed to parse plan data: %w", err)
	}

	// Log what we found
	if planData.Language != "" {
		logInfo("Found language: %s", planData.Language)
	} else {
		logWarning("No language information found in plan")
	}
	if planData.Framework != "" {
		logInfo("Found framework: %s", planData.Framework)
	}
	if planData.Database != "" && planData.Database != "N/A" {
		logInfo("Found database: %s", planData.Database)
	}
	if planData.ProjectType != "" {
		logInfo("Found project type: %s", planData.ProjectType)
	}

	// Determine which agent to update
	if len(args) > 0 {
		agentType = args[0]
		// Update specific agent
		if err := updateSpecificAgent(paths, planData, AgentType(agentType)); err != nil {
			return err
		}
	} else {
		// Update all existing agents
		if err := updateAllExistingAgents(paths, planData); err != nil {
			return err
		}
	}

	// Print summary
	printUpdateSummary(planData)

	return nil
}

// getAgentFileConfig returns the file path and name for a given agent type
func getAgentFileConfig(repoRoot string, agent AgentType) AgentFileConfig {
	configs := map[AgentType]AgentFileConfig{
		AgentClaude:    {Path: filepath.Join(repoRoot, "CLAUDE.md"), Name: "Claude Code"},
		AgentGemini:    {Path: filepath.Join(repoRoot, "GEMINI.md"), Name: "Gemini CLI"},
		AgentCopilot:   {Path: filepath.Join(repoRoot, ".github", "copilot-instructions.md"), Name: "GitHub Copilot"},
		AgentCursor:    {Path: filepath.Join(repoRoot, ".cursor", "rules", "tchncrt-rules.mdc"), Name: "Cursor IDE"},
		AgentQwen:      {Path: filepath.Join(repoRoot, "QWEN.md"), Name: "Qwen Code"},
		AgentOpenCode:  {Path: filepath.Join(repoRoot, "AGENTS.md"), Name: "opencode"},
		AgentCodex:     {Path: filepath.Join(repoRoot, "AGENTS.md"), Name: "Codex CLI"},
		AgentWindsurf:  {Path: filepath.Join(repoRoot, ".windsurf", "rules", "tchncrt-rules.md"), Name: "Windsurf"},
		AgentKiloCode:  {Path: filepath.Join(repoRoot, ".kilocode", "rules", "tchncrt-rules.md"), Name: "Kilo Code"},
		AgentAuggie:    {Path: filepath.Join(repoRoot, ".augment", "rules", "tchncrt-rules.md"), Name: "Auggie CLI"},
		AgentRoo:       {Path: filepath.Join(repoRoot, ".roo", "rules", "tchncrt-rules.md"), Name: "Roo Code"},
		AgentCodeBuddy: {Path: filepath.Join(repoRoot, ".codebuddy", "rules", "tchncrt-rules.md"), Name: "CodeBuddy"},
		AgentQ:         {Path: filepath.Join(repoRoot, "AGENTS.md"), Name: "Amazon Q Developer CLI"},
	}

	return configs[agent]
}

// parsePlanData extracts information from plan.md
func parsePlanData(planPath string) (*PlanData, error) {
	file, err := os.Open(planPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := &PlanData{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "**Language/Version**: ") {
			val := strings.TrimPrefix(line, "**Language/Version**: ")
			val = strings.TrimSpace(val)
			if val != "NEEDS CLARIFICATION" && val != "N/A" {
				data.Language = val
			}
		} else if strings.HasPrefix(line, "**Primary Dependencies**: ") {
			val := strings.TrimPrefix(line, "**Primary Dependencies**: ")
			val = strings.TrimSpace(val)
			if val != "NEEDS CLARIFICATION" && val != "N/A" {
				data.Framework = val
			}
		} else if strings.HasPrefix(line, "**Storage**: ") {
			val := strings.TrimPrefix(line, "**Storage**: ")
			val = strings.TrimSpace(val)
			if val != "NEEDS CLARIFICATION" && val != "N/A" {
				data.Database = val
			}
		} else if strings.HasPrefix(line, "**Project Type**: ") {
			val := strings.TrimPrefix(line, "**Project Type**: ")
			val = strings.TrimSpace(val)
			if val != "NEEDS CLARIFICATION" && val != "N/A" {
				data.ProjectType = val
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return data, nil
}

// formatTechnologyStack formats the technology stack string
func formatTechnologyStack(lang, framework string) string {
	var parts []string

	if lang != "" && lang != "NEEDS CLARIFICATION" {
		parts = append(parts, lang)
	}
	if framework != "" && framework != "NEEDS CLARIFICATION" && framework != "N/A" {
		parts = append(parts, framework)
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, " + ")
}

// updateSpecificAgent updates a single agent file
func updateSpecificAgent(paths *FeaturePaths, planData *PlanData, agent AgentType) error {
	config := getAgentFileConfig(paths.RepoRoot, agent)
	return updateAgentFile(config, paths, planData)
}

// updateAllExistingAgents updates all existing agent files
func updateAllExistingAgents(paths *FeaturePaths, planData *PlanData) error {
	agents := []AgentType{
		AgentClaude, AgentGemini, AgentCopilot, AgentCursor, AgentQwen,
		AgentOpenCode, AgentCodex, AgentWindsurf, AgentKiloCode, AgentAuggie,
		AgentRoo, AgentCodeBuddy, AgentQ,
	}

	foundAgent := false
	seenPaths := make(map[string]bool) // Track paths we've already updated

	for _, agent := range agents {
		config := getAgentFileConfig(paths.RepoRoot, agent)

		// Skip if we've already updated this path (e.g., AGENTS.md shared by multiple agents)
		if seenPaths[config.Path] {
			continue
		}

		if _, err := os.Stat(config.Path); err == nil {
			if err := updateAgentFile(config, paths, planData); err != nil {
				return fmt.Errorf("failed to update %s: %w", config.Name, err)
			}
			seenPaths[config.Path] = true
			foundAgent = true
		}
	}

	// If no agent files exist, create a default Claude file
	if !foundAgent {
		logInfo("No existing agent files found, creating default Claude file...")
		config := getAgentFileConfig(paths.RepoRoot, AgentClaude)
		if err := updateAgentFile(config, paths, planData); err != nil {
			return fmt.Errorf("failed to create default Claude file: %w", err)
		}
	}

	return nil
}

// updateAgentFile updates or creates an agent file
func updateAgentFile(config AgentFileConfig, paths *FeaturePaths, planData *PlanData) error {
	logInfo("Updating %s context file: %s", config.Name, config.Path)

	// Create directory if it doesn't exist
	dir := filepath.Dir(config.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(config.Path); os.IsNotExist(err) {
		// Create new file from template
		return createNewAgentFile(config, paths, planData)
	}

	// Update existing file
	return updateExistingAgentFile(config, paths, planData)
}

// createNewAgentFile creates a new agent file from template
func createNewAgentFile(config AgentFileConfig, paths *FeaturePaths, planData *PlanData) error {
	templatePath := filepath.Join(paths.RepoRoot, ".tchncrt", "templates", "agent-file-template.md")

	// Check if template exists
	info, err := os.Stat(templatePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("template not found at %s", templatePath)
	}
	if err != nil {
		return fmt.Errorf("failed to stat template file: %w", err)
	}

	// Check if template is readable
	if info.Mode().Perm()&0400 == 0 {
		return fmt.Errorf("template file is not readable: %s", templatePath)
	}

	// Read template
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	// Warn if not in a git repository
	if !paths.HasGit {
		logWarning("Git repository not detected; skipped branch validation")
	}

	// Replace placeholders
	projectName := filepath.Base(paths.RepoRoot)
	currentDate := time.Now().Format("2006-01-02")
	techStack := formatTechnologyStack(planData.Language, planData.Framework)

	text := string(content)
	text = strings.ReplaceAll(text, "[PROJECT NAME]", projectName)
	text = strings.ReplaceAll(text, "[DATE]", currentDate)

	// Build technology stack entry
	techEntry := ""
	if techStack != "" {
		techEntry = fmt.Sprintf("- %s (%s)", techStack, paths.CurrentBranch)
	} else {
		techEntry = fmt.Sprintf("- (%s)", paths.CurrentBranch)
	}
	text = strings.ReplaceAll(text, "[EXTRACTED FROM ALL PLAN.MD FILES]", techEntry)

	// Project structure
	projectStructure := getProjectStructure(planData.ProjectType)
	text = strings.ReplaceAll(text, "[ACTUAL STRUCTURE FROM PLANS]", projectStructure)

	// Commands
	commands := getCommandsForLanguage(planData.Language)
	text = strings.ReplaceAll(text, "[ONLY COMMANDS FOR ACTIVE TECHNOLOGIES]", commands)

	// Language conventions
	conventions := getLanguageConventions(planData.Language)
	text = strings.ReplaceAll(text, "[LANGUAGE-SPECIFIC, ONLY FOR LANGUAGES IN USE]", conventions)

	// Recent changes
	recentChange := ""
	if techStack != "" {
		recentChange = fmt.Sprintf("- %s: Added %s", paths.CurrentBranch, techStack)
	} else {
		recentChange = fmt.Sprintf("- %s: Added", paths.CurrentBranch)
	}
	text = strings.ReplaceAll(text, "[LAST 3 FEATURES AND WHAT THEY ADDED]", recentChange)

	// Write to temporary file first for atomic update
	tmpFile, err := os.CreateTemp(filepath.Dir(config.Path), ".agent-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // Clean up temp file if we fail

	if _, err := tmpFile.WriteString(text); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomically move temp file to target
	if err := os.Rename(tmpPath, config.Path); err != nil {
		return fmt.Errorf("failed to move temp file to target: %w", err)
	}

	logSuccess("Created new %s context file", config.Name)
	return nil
}

// updateExistingAgentFile updates an existing agent file
func updateExistingAgentFile(config AgentFileConfig, paths *FeaturePaths, planData *PlanData) error {
	// Read existing file
	content, err := os.ReadFile(config.Path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	currentDate := time.Now().Format("2006-01-02")
	techStack := formatTechnologyStack(planData.Language, planData.Framework)

	// Prepare new entries
	var newTechEntries []string
	if techStack != "" && !strings.Contains(string(content), techStack) {
		newTechEntries = append(newTechEntries, fmt.Sprintf("- %s (%s)", techStack, paths.CurrentBranch))
	}
	if planData.Database != "" && planData.Database != "N/A" && planData.Database != "NEEDS CLARIFICATION" && !strings.Contains(string(content), planData.Database) {
		newTechEntries = append(newTechEntries, fmt.Sprintf("- %s (%s)", planData.Database, paths.CurrentBranch))
	}

	// Prepare new change entry
	newChangeEntry := ""
	if techStack != "" {
		newChangeEntry = fmt.Sprintf("- %s: Added %s", paths.CurrentBranch, techStack)
	} else if planData.Database != "" && planData.Database != "N/A" && planData.Database != "NEEDS CLARIFICATION" {
		newChangeEntry = fmt.Sprintf("- %s: Added %s", paths.CurrentBranch, planData.Database)
	}

	// Process file
	var result []string
	inTechSection := false
	inChangesSection := false
	techEntriesAdded := false
	existingChangesCount := 0

	dateRegex := regexp.MustCompile(`\*\*Last updated\*\*:.*(\d{4}-\d{2}-\d{2})`)

	for _, line := range lines {
		// Handle Active Technologies section
		if strings.HasPrefix(line, "## Active Technologies") {
			result = append(result, line)
			inTechSection = true
			continue
		} else if inTechSection && strings.HasPrefix(line, "## ") {
			// Add new tech entries before closing the section
			if !techEntriesAdded && len(newTechEntries) > 0 {
				result = append(result, newTechEntries...)
				techEntriesAdded = true
			}
			result = append(result, line)
			inTechSection = false
			continue
		} else if inTechSection && line == "" {
			// Add new tech entries before empty line
			if !techEntriesAdded && len(newTechEntries) > 0 {
				result = append(result, newTechEntries...)
				techEntriesAdded = true
			}
			result = append(result, line)
			continue
		}

		// Handle Recent Changes section
		if strings.HasPrefix(line, "## Recent Changes") {
			result = append(result, line)
			// Add new change entry right after the heading
			if newChangeEntry != "" {
				result = append(result, newChangeEntry)
			}
			inChangesSection = true
			continue
		} else if inChangesSection && strings.HasPrefix(line, "## ") {
			result = append(result, line)
			inChangesSection = false
			continue
		} else if inChangesSection && strings.HasPrefix(line, "- ") {
			// Keep only first 2 existing changes
			if existingChangesCount < 2 {
				result = append(result, line)
				existingChangesCount++
			}
			continue
		}

		// Update timestamp
		if dateRegex.MatchString(line) {
			line = dateRegex.ReplaceAllString(line, fmt.Sprintf("**Last updated**: %s", currentDate))
		}

		result = append(result, line)
	}

	// Post-loop check: if we're still in the Active Technologies section
	if inTechSection && !techEntriesAdded && len(newTechEntries) > 0 {
		result = append(result, newTechEntries...)
	}

	// Write to temporary file first for atomic update
	tmpFile, err := os.CreateTemp(filepath.Dir(config.Path), ".agent-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // Clean up temp file if we fail

	updatedContent := strings.Join(result, "\n")
	if _, err := tmpFile.WriteString(updatedContent); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomically move temp file to target
	if err := os.Rename(tmpPath, config.Path); err != nil {
		return fmt.Errorf("failed to move temp file to target: %w", err)
	}

	logSuccess("Updated existing %s context file", config.Name)
	return nil
}

// getProjectStructure returns project structure based on project type
func getProjectStructure(projectType string) string {
	if strings.Contains(strings.ToLower(projectType), "web") {
		return "backend/\nfrontend/\ntests/"
	}
	return "src/\ntests/"
}

// getCommandsForLanguage returns build/test commands for a language
func getCommandsForLanguage(lang string) string {
	lower := strings.ToLower(lang)
	if strings.Contains(lower, "python") {
		return "cd src && pytest && ruff check ."
	} else if strings.Contains(lower, "rust") {
		return "cargo test && cargo clippy"
	} else if strings.Contains(lower, "javascript") || strings.Contains(lower, "typescript") {
		return "npm test && npm run lint"
	} else if strings.Contains(lower, "go") {
		return "go test ./... && go vet ./..."
	}
	return fmt.Sprintf("# Add commands for %s", lang)
}

// getLanguageConventions returns language-specific conventions
func getLanguageConventions(lang string) string {
	if lang == "" {
		return ""
	}
	return fmt.Sprintf("%s: Follow standard conventions", lang)
}

// printUpdateSummary prints a summary of changes
func printUpdateSummary(planData *PlanData) {
	fmt.Fprintln(os.Stderr, "")
	logInfo("Summary of changes:")

	if planData.Language != "" {
		fmt.Fprintf(os.Stderr, "  - Added language: %s\n", planData.Language)
	}
	if planData.Framework != "" {
		fmt.Fprintf(os.Stderr, "  - Added framework: %s\n", planData.Framework)
	}
	if planData.Database != "" && planData.Database != "N/A" {
		fmt.Fprintf(os.Stderr, "  - Added database: %s\n", planData.Database)
	}

	fmt.Fprintln(os.Stderr, "")
	logInfo("Usage: technocrat update-agent-context [claude|gemini|copilot|cursor|qwen|opencode|codex|windsurf|kilocode|auggie|roo|codebuddy|q]")
}
