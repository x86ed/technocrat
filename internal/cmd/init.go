package cmd

import (
	"archive/zip"
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"technocrat/internal/ui"

	"github.com/spf13/cobra"
)

// AgentConfig holds configuration for each AI agent
type AgentConfig struct {
	Name        string
	Folder      string
	InstallURL  string
	RequiresCLI bool
}

var agentConfigs = map[string]AgentConfig{
	"copilot": {
		Name:        "GitHub Copilot",
		Folder:      ".github/",
		InstallURL:  "",
		RequiresCLI: false,
	},
	"claude": {
		Name:        "Claude Code",
		Folder:      ".claude/",
		InstallURL:  "https://docs.anthropic.com/en/docs/claude-code/setup",
		RequiresCLI: true,
	},
	"gemini": {
		Name:        "Gemini CLI",
		Folder:      ".gemini/",
		InstallURL:  "https://github.com/google-gemini/gemini-cli",
		RequiresCLI: true,
	},
	"cursor-agent": {
		Name:        "Cursor",
		Folder:      ".cursor/",
		InstallURL:  "",
		RequiresCLI: false,
	},
	"qwen": {
		Name:        "Qwen Code",
		Folder:      ".qwen/",
		InstallURL:  "https://github.com/QwenLM/qwen-code",
		RequiresCLI: true,
	},
	"opencode": {
		Name:        "opencode",
		Folder:      ".opencode/",
		InstallURL:  "https://opencode.ai",
		RequiresCLI: true,
	},
	"codex": {
		Name:        "Codex CLI",
		Folder:      ".codex/",
		InstallURL:  "https://github.com/openai/codex",
		RequiresCLI: true,
	},
	"windsurf": {
		Name:        "Windsurf",
		Folder:      ".windsurf/",
		InstallURL:  "",
		RequiresCLI: false,
	},
	"kilocode": {
		Name:        "Kilo Code",
		Folder:      ".kilocode/",
		InstallURL:  "",
		RequiresCLI: false,
	},
	"auggie": {
		Name:        "Auggie CLI",
		Folder:      ".augment/",
		InstallURL:  "https://docs.augmentcode.com/cli/setup-auggie/install-auggie-cli",
		RequiresCLI: true,
	},
	"codebuddy": {
		Name:        "CodeBuddy",
		Folder:      ".codebuddy/",
		InstallURL:  "https://www.codebuddy.ai",
		RequiresCLI: true,
	},
	"roo": {
		Name:        "Roo Code",
		Folder:      ".roo/",
		InstallURL:  "",
		RequiresCLI: false,
	},
	"q": {
		Name:        "Amazon Q Developer CLI",
		Folder:      ".amazonq/",
		InstallURL:  "https://aws.amazon.com/developer/learning/q-developer-cli/",
		RequiresCLI: true,
	},
}

var (
	aiAssistant      string
	scriptType       string
	ignoreAgentTools bool
	noGit            bool
	here             bool
	force            bool
	githubToken      string
	skipTLS          bool
	debug            bool
)

const banner = `
 ███████████                   █████                                                        █████   
░█░░░███░░░█                  ░░███                                                        ░░███    
░   ░███  ░   ██████   ██████  ░███████   ████████    ██████   ██████  ████████   ██████   ███████  
    ░███     ███░░███ ███░░███ ░███░░███ ░░███░░███  ███░░███ ███░░███░░███░░███ ░░░░░███ ░░░███░   
    ░███    ░███████ ░███ ░░░  ░███ ░███  ░███ ░███ ░███ ░███░███ ░░░  ░███ ░░░   ███████   ░███    
    ░███    ░███░░░  ░███  ███ ░███ ░███  ░███ ░███ ░███ ░███░███  ███ ░███      ███░░███   ░███ ███
    █████   ░░██████ ░░██████  ████ █████ ████ █████░░██████ ░░██████  █████    ░░████████  ░░█████ 
   ░░░░░     ░░░░░░   ░░░░░░  ░░░░ ░░░░░ ░░░░ ░░░░░  ░░░░░░   ░░░░░░  ░░░░░      ░░░░░░░░    ░░░░░  
`

const tagline = "Technocrat - Spec-Driven Development Toolkit"

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new Technocrat project from template",
	Long: `Initialize a new Technocrat project from the latest template.

This command will:
  1. Check that required tools are installed
  2. Let you choose your AI assistant (or use --ai flag)
  3. Download the appropriate template from GitHub
  4. Extract the template to a new project directory or current directory
  5. Initialize a fresh git repository (unless --no-git)
  6. Set up executable permissions for scripts

Examples:
  technocrat init my-project
  technocrat init my-project --ai claude
  technocrat init my-project --ai copilot --no-git
  technocrat init . --ai claude         # Initialize in current directory
  technocrat init --here --ai claude    # Alternative syntax
  technocrat init --here --force        # Skip confirmation for non-empty dir`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVar(&aiAssistant, "ai", "", "AI assistant to use (claude, gemini, copilot, cursor-agent, etc.)")
	initCmd.Flags().StringVar(&scriptType, "script", "", "Script type: sh or ps")
	initCmd.Flags().BoolVar(&ignoreAgentTools, "ignore-agent-tools", false, "Skip checks for AI agent tools")
	initCmd.Flags().BoolVar(&noGit, "no-git", false, "Skip git repository initialization")
	initCmd.Flags().BoolVar(&here, "here", false, "Initialize in current directory")
	initCmd.Flags().BoolVar(&force, "force", false, "Force overwrite when using --here")
	initCmd.Flags().StringVar(&githubToken, "github-token", "", "GitHub token for API requests")
	initCmd.Flags().BoolVar(&skipTLS, "skip-tls", false, "Skip SSL/TLS verification (not recommended)")
	initCmd.Flags().BoolVar(&debug, "debug", false, "Show verbose diagnostic output")
}

func showDebugEnvironment() {
	if !debug {
		return
	}

	cwd, _ := os.Getwd()
	envInfo := fmt.Sprintf("Go:       %s\nOS:       %s\nArch:     %s\nCWD:      %s",
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		cwd)

	ui.ShowInfo("Debug Environment", envInfo)
}

func showBanner() {
	// ANSI color codes
	colors := []string{
		"\033[94m", // bright_blue
		"\033[34m", // blue
		"\033[36m", // cyan
		"\033[96m", // bright_cyan
		"\033[37m", // white
		"\033[97m", // bright_white
	}
	reset := "\033[0m"
	yellow := "\033[93m"
	italic := "\033[3m"

	// Get terminal width for centering
	termWidth := 80 // default width
	if width, err := getTerminalWidth(); err == nil && width > 0 {
		termWidth = width
	}

	// Split banner into lines and colorize each line
	bannerLines := strings.Split(strings.TrimSpace(banner), "\n")
	for i, line := range bannerLines {
		color := colors[i%len(colors)]
		centeredLine := centerText(line, termWidth)
		fmt.Fprintf(os.Stderr, "%s%s%s\n", color, centeredLine, reset)
	}

	// Center and colorize tagline
	centeredTagline := centerText(tagline, termWidth)
	fmt.Fprintf(os.Stderr, "\n%s%s%s%s\n\n", italic, yellow, centeredTagline, reset)
}

func getTerminalWidth() (int, error) {
	// Try to get terminal width using stty
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	parts := strings.Fields(string(output))
	if len(parts) == 2 {
		width := 0
		fmt.Sscanf(parts[1], "%d", &width)
		return width, nil
	}

	return 0, fmt.Errorf("could not parse terminal size")
}

func centerText(text string, width int) string {
	textLen := len(text)
	if textLen >= width {
		return text
	}

	padding := (width - textLen) / 2
	return strings.Repeat(" ", padding) + text
}

func runInit(cmd *cobra.Command, args []string) error {
	var projectName string
	var projectPath string

	// Show banner
	showBanner()

	// Parse project name argument
	if len(args) > 0 {
		projectName = args[0]
		if projectName == "." {
			here = true
			projectName = ""
		}
	}

	// Validate arguments
	if here && projectName != "" {
		showDebugEnvironment()
		return fmt.Errorf("cannot specify both project name and --here flag")
	}

	if !here && projectName == "" {
		showDebugEnvironment()
		return fmt.Errorf("must specify either a project name, use '.' for current directory, or use --here flag")
	}

	// Determine project path
	if here {
		cwd, err := os.Getwd()
		if err != nil {
			showDebugEnvironment()
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		projectPath = cwd
		projectName = filepath.Base(cwd)

		// Check if current directory is not empty
		entries, err := os.ReadDir(projectPath)
		if err != nil {
			showDebugEnvironment()
			return fmt.Errorf("failed to read current directory: %w", err)
		}

		// Filter out hidden files and git directory for the count
		visibleCount := 0
		for _, entry := range entries {
			if !strings.HasPrefix(entry.Name(), ".") {
				visibleCount++
			}
		}

		if len(entries) > 0 {
			fmt.Fprintf(os.Stderr, "\n⚠ Warning: Current directory is not empty (%d items, %d visible)\n", len(entries), visibleCount)
			fmt.Fprintln(os.Stderr, "Template files will be merged with existing content and may overwrite existing files")

			if !force {
				fmt.Fprint(os.Stderr, "Do you want to continue? [y/N]: ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				response = strings.TrimSpace(strings.ToLower(response))
				if response != "y" && response != "yes" {
					fmt.Fprintln(os.Stderr, "\nOperation cancelled")
					return nil
				}
			}
		}
	} else {
		absPath, err := filepath.Abs(projectName)
		if err != nil {
			showDebugEnvironment()
			return fmt.Errorf("failed to resolve project path: %w", err)
		}
		projectPath = absPath

		if _, err := os.Stat(projectPath); err == nil {
			ui.ShowError("Directory Conflict",
				fmt.Sprintf("Directory '%s' already exists.\n\nPlease choose a different project name or use --here to initialize in current directory.", projectName))
			showDebugEnvironment()
			return fmt.Errorf("directory '%s' already exists", projectName)
		}
	}

	// Print setup information in a panel
	cwd, _ := os.Getwd()
	setupInfo := fmt.Sprintf("Project:      %s\nWorking Path: %s", projectName, cwd)
	if !here {
		setupInfo += fmt.Sprintf("\nTarget Path:  %s", projectPath)
	}
	ui.ShowInfo("Technocrat Project Setup", setupInfo)

	// Check for git
	shouldInitGit := false
	if !noGit {
		if checkToolInstalled("git") {
			shouldInitGit = true
		} else {
			fmt.Fprintln(os.Stderr, "⚠ Warning: Git not found - will skip repository initialization")
		}
	}

	// Select AI assistant
	selectedAI := aiAssistant
	if selectedAI == "" {
		var err error
		selectedAI, err = promptForAgent()
		if err != nil {
			showDebugEnvironment()
			return err
		}
	} else {
		// Validate provided AI assistant
		if _, ok := agentConfigs[selectedAI]; !ok {
			showDebugEnvironment()
			return fmt.Errorf("invalid AI assistant '%s'. Valid options: %s", selectedAI, getAgentList())
		}
	}

	// Check if agent CLI is required and installed
	if !ignoreAgentTools {
		config := agentConfigs[selectedAI]
		if config.RequiresCLI {
			if !checkToolInstalled(selectedAI) {
				msg := fmt.Sprintf("%s CLI not found\n\nInstall from: %s\n\nTip: Use --ignore-agent-tools to skip this check", config.Name, config.InstallURL)
				if config.InstallURL == "" {
					msg = fmt.Sprintf("%s CLI not found\n\nTip: Use --ignore-agent-tools to skip this check", config.Name)
				}
				ui.ShowError("Agent Tool Not Found", msg)
				showDebugEnvironment()
				return fmt.Errorf("required agent tool not found: %s", selectedAI)
			}
		}
	}

	// Select script type
	selectedScript := scriptType
	if selectedScript == "" {
		selectedScript = promptForScriptType()
	} else if selectedScript != "sh" && selectedScript != "ps" {
		showDebugEnvironment()
		return fmt.Errorf("invalid script type '%s'. Choose 'sh' or 'ps'", selectedScript)
	}

	fmt.Fprintf(os.Stderr, "\n✓ Selected AI assistant: %s\n", selectedAI)
	fmt.Fprintf(os.Stderr, "✓ Selected script type:  %s\n\n", selectedScript)

	// Create progress tracker
	tracker := ui.NewStepTracker("Installation Progress")
	tracker.Add("download", "Downloading template")
	tracker.Add("extract", "Extracting files")
	if runtime.GOOS != "windows" {
		tracker.Add("permissions", "Setting permissions")
	}
	if shouldInitGit && !isGitRepo(projectPath) {
		tracker.Add("git", "Initializing git repository")
	}

	// Start live tracker if in interactive mode
	if err := tracker.StartLive(); err != nil && debug {
		fmt.Fprintf(os.Stderr, "Debug: Could not start live tracker: %v\n", err)
	}

	// Show progress message for non-TTY mode
	if !ui.IsInteractive() {
		fmt.Fprintln(os.Stderr, "\n→ Downloading and extracting template...")
	}

	// Download and extract template
	if err := downloadAndExtractTemplate(projectPath, selectedAI, selectedScript, here, tracker); err != nil {
		tracker.StopLive()
		showDebugEnvironment()
		return fmt.Errorf("failed to download template: %w", err)
	}

	// Ensure scripts are executable (Unix-like systems)
	if runtime.GOOS != "windows" {
		if !ui.IsInteractive() {
			fmt.Fprintln(os.Stderr, "→ Setting script permissions...")
		}
		tracker.Start("permissions", "Making scripts executable...")
		if err := makeScriptsExecutable(projectPath); err != nil {
			tracker.Error("permissions", err.Error())
			if !ui.IsInteractive() {
				fmt.Fprintf(os.Stderr, "  ⚠ Warning: Failed to set script permissions: %v\n", err)
			}
		} else {
			tracker.Complete("permissions", "Scripts ready")
			if !ui.IsInteractive() {
				fmt.Fprintln(os.Stderr, "  ✓ Scripts are executable")
			}
		}
	}

	// Initialize git repository
	var gitError error
	if shouldInitGit && !isGitRepo(projectPath) {
		if !ui.IsInteractive() {
			fmt.Fprintln(os.Stderr, "→ Initializing git repository...")
		}
		tracker.Start("git", "Running git init...")
		if err := initGitRepo(projectPath); err != nil {
			tracker.Error("git", err.Error())
			if !ui.IsInteractive() {
				fmt.Fprintf(os.Stderr, "  ⚠ Warning: Failed to initialize git repository: %v\n", err)
			}
			gitError = err
		} else {
			tracker.Complete("git", "Repository initialized")
			if !ui.IsInteractive() {
				fmt.Fprintln(os.Stderr, "  ✓ Git repository initialized")
			}
		}
	}

	// Stop live tracker
	tracker.StopLive()

	// Print final summary
	fmt.Fprintf(os.Stderr, "\n%s %s\n", ui.ColorGreen.Sprint(ui.SymbolCheckmark), tracker.Summary())

	// Print success message
	fmt.Fprintln(os.Stderr, "\n═══════════════════════════════════════════════════════════")
	fmt.Fprintln(os.Stderr, "  ✓ Project initialized successfully!")
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")

	// Show git error details if any
	if gitError != nil {
		msg := fmt.Sprintf("%v\n\nYou can initialize git manually later with:\n  git init\n  git add .\n  git commit -m \"Initial commit\"", gitError)
		ui.ShowWarning("Git Initialization Failed", msg)
	}

	// Show agent folder security notice
	config := agentConfigs[selectedAI]
	msg := fmt.Sprintf("Some agents may store credentials in %s\n\nConsider adding %s to .gitignore to prevent credential leakage", config.Folder, config.Folder)
	ui.ShowWarning("Security Notice", msg)

	// Build Next Steps content
	var nextSteps strings.Builder
	stepNum := 1

	if !here {
		nextSteps.WriteString(fmt.Sprintf("%d. cd %s\n\n", stepNum, projectName))
		stepNum++
	}

	// Add Codex-specific environment variable setup
	if selectedAI == "codex" {
		codexPath := filepath.Join(projectPath, ".codex")
		var cmd string
		if runtime.GOOS == "windows" {
			cmd = fmt.Sprintf("setx CODEX_HOME \"%s\"", codexPath)
		} else {
			cmd = fmt.Sprintf("export CODEX_HOME='%s'", strings.ReplaceAll(codexPath, "'", "'\\''"))
		}
		nextSteps.WriteString(fmt.Sprintf("%d. Set Codex environment variable:\n   %s\n", stepNum, cmd))
		if runtime.GOOS != "windows" {
			nextSteps.WriteString(fmt.Sprintf("   %s\n", ui.ColorDim.Sprint("Note: Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.)")))
		}
		nextSteps.WriteString("\n")
		stepNum++
	}

	nextSteps.WriteString(fmt.Sprintf("%d. Start using slash commands with your AI agent:\n\n", stepNum))
	nextSteps.WriteString("   Core workflow commands:\n")
	nextSteps.WriteString("     • /tchncrt.constitution - Establish project principles\n")
	nextSteps.WriteString("     • /tchncrt.spec         - Create baseline specification\n")
	nextSteps.WriteString("     • /tchncrt.plan         - Create implementation plan\n")
	nextSteps.WriteString("     • /tchncrt.tasks        - Generate actionable tasks\n")
	nextSteps.WriteString("     • /tchncrt.implement    - Execute implementation\n")

	ui.ShowSuccess("Next Steps", nextSteps.String())

	// Build Enhancement Commands content
	var enhancements strings.Builder
	enhancements.WriteString("Use these commands to improve quality & confidence:\n\n")
	enhancements.WriteString(fmt.Sprintf("  • /tchncrt.clarify   - Ask structured questions\n    %s\n\n",
		ui.ColorDim.Sprint("Use before creating your plan")))
	enhancements.WriteString(fmt.Sprintf("  • /tchncrt.checklist - Quality validation checklists\n    %s\n\n",
		ui.ColorDim.Sprint("Use after creating your plan")))
	enhancements.WriteString(fmt.Sprintf("  • /tchncrt.analyze   - Cross-artifact consistency report\n    %s",
		ui.ColorDim.Sprint("Use after generating tasks")))

	ui.ShowInfo("Enhancement Commands", enhancements.String())

	return nil
}

func checkToolInstalled(tool string) bool {
	// Special handling for Claude CLI after `claude migrate-installer`
	// The migrate-installer command removes the original executable from PATH
	// and creates an alias at ~/.claude/local/claude instead
	if tool == "claude" {
		claudePath := filepath.Join(os.Getenv("HOME"), ".claude", "local", "claude")
		if _, err := os.Stat(claudePath); err == nil {
			return true
		}
	}

	_, err := exec.LookPath(tool)
	return err == nil
}

func promptForAgent() (string, error) {
	// Check if running in interactive mode (TTY)
	if ui.IsInteractive() {
		// Use interactive arrow-key selection
		options := make(map[string]string)
		for key, config := range agentConfigs {
			options[key] = config.Name
		}

		selected, err := ui.SelectWithArrows("Choose your AI assistant", options, "copilot")
		if err != nil {
			// Fall back to text input on error
			return promptForAgentText()
		}
		return selected, nil
	}

	// Fall back to text input for non-interactive mode
	return promptForAgentText()
}

func promptForAgentText() (string, error) {
	fmt.Fprintln(os.Stderr, "Available AI assistants:")
	fmt.Fprintln(os.Stderr, "─────────────────────────────────────────────")

	// Sort keys for consistent display
	agents := []string{}
	for key := range agentConfigs {
		agents = append(agents, key)
	}

	for _, key := range agents {
		config := agentConfigs[key]
		fmt.Fprintf(os.Stderr, "  %-15s - %s\n", key, config.Name)
	}
	fmt.Fprintln(os.Stderr, "─────────────────────────────────────────────")

	fmt.Fprint(os.Stderr, "\nEnter AI assistant name (default: copilot): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return "copilot", nil
	}

	if _, ok := agentConfigs[input]; !ok {
		return "", fmt.Errorf("invalid AI assistant: %s", input)
	}

	return input, nil
}

func getAgentList() string {
	agents := make([]string, 0, len(agentConfigs))
	for key := range agentConfigs {
		agents = append(agents, key)
	}
	return strings.Join(agents, ", ")
}

func getDefaultScriptType() string {
	if runtime.GOOS == "windows" {
		return "ps"
	}
	return "sh"
}

func promptForScriptType() string {
	defaultScript := getDefaultScriptType()

	// Check if running in interactive mode (TTY)
	if ui.IsInteractive() {
		// Use interactive arrow-key selection
		options := map[string]string{
			"sh": "POSIX Shell (bash/zsh)",
			"ps": "PowerShell",
		}

		selected, err := ui.SelectWithArrows("Choose script type", options, defaultScript)
		if err != nil {
			// Fall back to default on error
			fmt.Fprintf(os.Stderr, "Using default script type: %s\n", defaultScript)
			return defaultScript
		}
		return selected
	}

	// Fall back to default for non-interactive mode
	fmt.Fprintf(os.Stderr, "Using default script type: %s\n", defaultScript)
	return defaultScript
}

func downloadAndExtractTemplate(projectPath, aiAssistant, scriptType string, inCurrentDir bool, tracker *ui.StepTracker) error {
	// GitHub repository details
	repoOwner := "x86ed"
	repoName := "technocrat"

	// Start download step
	if tracker != nil {
		tracker.Start("download", "Fetching latest release...")
	}

	// Get GitHub token from flag or environment
	// Check CLI flag first, then GITHUB_TOKEN, then GH_TOKEN
	token := githubToken
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	// Fetch latest release
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Configure HTTP client based on skip-tls flag
	client := &http.Client{Timeout: 30 * time.Second}
	if skipTLS {
		fmt.Fprintln(os.Stderr, "  ⚠ Warning: Skipping TLS verification (not recommended)")
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	} else {
		// Use system certificate pool for better SSL handling
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		if debug {
			fmt.Fprintf(os.Stderr, "  Debug: Failed to fetch release info: %v\n", err)
			fmt.Fprintf(os.Stderr, "  Debug: Request URL: %s\n", apiURL)
		}
		return fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if debug {
			fmt.Fprintf(os.Stderr, "  Debug: GitHub API returned status %d\n", resp.StatusCode)
			fmt.Fprintf(os.Stderr, "  Debug: Response headers: %v\n", resp.Header)
			body, _ := io.ReadAll(resp.Body)
			fmt.Fprintf(os.Stderr, "  Debug: Response body (truncated): %s\n", string(body[:min(400, len(body))]))
		}
		return fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var releaseData struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int    `json:"size"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&releaseData); err != nil {
		return fmt.Errorf("failed to parse release data: %w", err)
	}

	// Find matching asset
	pattern := fmt.Sprintf("technocrat-template-%s-%s", aiAssistant, scriptType)
	var downloadURL string
	var assetName string
	var assetSize int

	for _, asset := range releaseData.Assets {
		if strings.Contains(asset.Name, pattern) && strings.HasSuffix(asset.Name, ".zip") {
			downloadURL = asset.BrowserDownloadURL
			assetName = asset.Name
			assetSize = asset.Size
			break
		}
	}

	if downloadURL == "" {
		if debug {
			fmt.Fprintf(os.Stderr, "  Debug: Available assets:\n")
			for _, asset := range releaseData.Assets {
				fmt.Fprintf(os.Stderr, "    - %s\n", asset.Name)
			}
			fmt.Fprintf(os.Stderr, "  Debug: Looking for pattern: %s\n", pattern)
		}
		return fmt.Errorf("no matching template found for %s with script type %s", aiAssistant, scriptType)
	}

	fmt.Fprintf(os.Stderr, "  ✓ Found template: %s (%s bytes)\n", assetName, formatBytes(assetSize))
	fmt.Fprintf(os.Stderr, "  ✓ Release: %s\n", releaseData.TagName)

	// Download template
	if tracker != nil {
		tracker.Start("download", fmt.Sprintf("Downloading %s...", assetName))
	}
	downloadReq, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	if token != "" {
		downloadReq.Header.Set("Authorization", "Bearer "+token)
	}

	downloadResp, err := client.Do(downloadReq)
	if err != nil {
		if debug {
			fmt.Fprintf(os.Stderr, "  Debug: Download failed: %v\n", err)
			fmt.Fprintf(os.Stderr, "  Debug: Download URL: %s\n", downloadURL)
		}
		return fmt.Errorf("failed to download template: %w", err)
	}
	defer downloadResp.Body.Close()

	if downloadResp.StatusCode != http.StatusOK {
		if debug {
			fmt.Fprintf(os.Stderr, "  Debug: Download status: %d\n", downloadResp.StatusCode)
			fmt.Fprintf(os.Stderr, "  Debug: Response headers: %v\n", downloadResp.Header)
		}
		return fmt.Errorf("download failed with status %d", downloadResp.StatusCode)
	}

	// Save to temporary file
	tempFile, err := os.CreateTemp("", "technocrat-template-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	// Copy with progress
	written, err := io.Copy(tempFile, downloadResp.Body)
	if err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to save template: %w", err)
	}
	tempFile.Close()

	if tracker != nil {
		tracker.Complete("download", formatBytes(int(written)))
	}
	fmt.Fprintf(os.Stderr, "  ✓ Downloaded: %s (%s)\n", assetName, formatBytes(int(written)))

	// Extract template
	if tracker != nil {
		tracker.Start("extract", "Unpacking archive...")
	}

	// Extract template with detailed tracking
	extractedCount, extractedSize, err := extractZipWithStats(tempPath, projectPath, inCurrentDir, tracker)
	if err != nil {
		if tracker != nil {
			tracker.Error("extract", err.Error())
		}
		return fmt.Errorf("failed to extract template: %w", err)
	}

	if tracker != nil {
		tracker.Complete("extract", fmt.Sprintf("%d files, %s", extractedCount, formatBytes(extractedSize)))
	}
	fmt.Fprintf(os.Stderr, "  ✓ Template extracted: %d files (%s)\n", extractedCount, formatBytes(extractedSize))

	return nil
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func extractZipWithStats(zipPath, destPath string, inCurrentDir bool, tracker *ui.StepTracker) (int, int, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return 0, 0, err
	}
	defer reader.Close()

	// Report archive contents
	if tracker != nil {
		tracker.Add("zip-list", "Archive contents")
		tracker.Complete("zip-list", fmt.Sprintf("%d items", len(reader.File)))
	}

	// If not extracting to current directory, create the destination
	if !inCurrentDir {
		if err := os.MkdirAll(destPath, 0755); err != nil {
			return 0, 0, err
		}
	}

	// Check if zip has a single root directory
	var rootDir string
	if len(reader.File) > 0 {
		firstName := reader.File[0].Name
		if idx := strings.Index(firstName, "/"); idx > 0 {
			potentialRoot := firstName[:idx]
			allInSameRoot := true
			for _, f := range reader.File {
				if !strings.HasPrefix(f.Name, potentialRoot+"/") {
					allInSameRoot = false
					break
				}
			}
			if allInSameRoot {
				rootDir = potentialRoot + "/"
			}
		}
	}

	extractedCount := 0
	extractedSize := 0

	for _, file := range reader.File {
		// Strip root directory if present
		targetName := file.Name
		if rootDir != "" && strings.HasPrefix(targetName, rootDir) {
			targetName = strings.TrimPrefix(targetName, rootDir)
		}

		if targetName == "" {
			continue
		}

		targetPath := filepath.Join(destPath, targetName)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, file.Mode()); err != nil {
				return extractedCount, extractedSize, err
			}
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return extractedCount, extractedSize, err
		}

		// Extract file
		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return extractedCount, extractedSize, err
		}

		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return extractedCount, extractedSize, err
		}

		written, err := io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()

		if err != nil {
			return extractedCount, extractedSize, err
		}

		extractedCount++
		extractedSize += int(written)
	}

	// Report extraction summary
	if tracker != nil {
		tracker.Add("extracted-summary", "Extraction summary")
		tracker.Complete("extracted-summary", fmt.Sprintf("%d files, %s", extractedCount, formatBytes(extractedSize)))
	}

	return extractedCount, extractedSize, nil
}

// Legacy function for backward compatibility
func extractZip(zipPath, destPath string, inCurrentDir bool) error {
	_, _, err := extractZipWithStats(zipPath, destPath, inCurrentDir, nil)
	return err
}

func makeScriptsExecutable(projectPath string) error {
	scriptsDir := filepath.Join(projectPath, ".tchncrt", "scripts")
	if _, err := os.Stat(scriptsDir); os.IsNotExist(err) {
		// No scripts directory, nothing to do
		return nil
	}

	updated := 0
	skipped := 0
	failed := 0
	err := filepath.Walk(scriptsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".sh") {
			// Check if file starts with shebang
			file, err := os.Open(path)
			if err != nil {
				failed++
				return nil // Skip on error
			}
			defer file.Close()

			header := make([]byte, 2)
			if n, _ := file.Read(header); n == 2 && string(header) == "#!" {
				// Get current mode
				currentMode := info.Mode()

				// Set execute bits based on read bits
				newMode := currentMode
				if currentMode&0400 != 0 {
					newMode |= 0100
				}
				if currentMode&0040 != 0 {
					newMode |= 0010
				}
				if currentMode&0004 != 0 {
					newMode |= 0001
				}

				// Only chmod if mode needs to change
				if newMode != currentMode {
					if err := os.Chmod(path, newMode); err != nil {
						failed++
						return nil // Skip on error
					}
					updated++
				} else {
					skipped++
				}
			}
		}
		return nil
	})

	if updated > 0 || failed > 0 {
		fmt.Fprintf(os.Stderr, "  ✓ Scripts: %d updated, %d skipped", updated, skipped)
		if failed > 0 {
			fmt.Fprintf(os.Stderr, ", %d failed", failed)
		}
		fmt.Fprintln(os.Stderr)
	}

	return err
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func isGitRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return true
	}

	// Also check using git command
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = path
	err := cmd.Run()
	return err == nil
}

func initGitRepo(path string) error {
	// Initialize repository
	cmd := exec.Command("git", "init")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git init failed: %w", err)
	}

	// Add all files
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	// Initial commit
	cmd = exec.Command("git", "commit", "-m", "Initial commit from Technocrat template")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	return nil
}
