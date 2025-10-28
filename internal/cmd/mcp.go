package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"technocrat/internal/editor"
	"technocrat/internal/installer"

	"github.com/spf13/cobra"
)

// checkMcpCmd represents the check-mcp command
var checkMcpCmd = &cobra.Command{
	Use:   "check-mcp",
	Short: "Check MCP server configuration for detected editors",
	Long: `Check MCP server configuration for detected editors.

This command:
- Detects installed editors that support MCP
- Verifies MCP configuration exists and is valid
- Reports the status of each editor's MCP setup
- Provides troubleshooting information if issues are found`,
	RunE: runCheckMCP,
}

// configureMcpCmd represents the configure-mcp command
var configureMcpCmd = &cobra.Command{
	Use:   "configure-mcp",
	Short: "Configure MCP server for specific editors",
	Long: `Configure MCP server for specific editors.

This command allows manual configuration of the MCP server for specific editors.
Use this if automatic configuration during 'technocrat init' failed or if you
want to add MCP support to additional editors later.`,
	RunE: runConfigureMCP,
}

func init() {
	rootCmd.AddCommand(checkMcpCmd)
	rootCmd.AddCommand(configureMcpCmd)

	// Add flags for configure-mcp
	configureMcpCmd.Flags().StringSlice("editors", nil, "Specific editors to configure (e.g., --editors claude,vscode)")
	configureMcpCmd.Flags().Bool("all", false, "Configure all detected editors")
}

func runCheckMCP(cmd *cobra.Command, args []string) error {
	// Get current directory as project path
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Detect available editors
	fmt.Println("🔍 Detecting editors with MCP support...")
	editors := editor.DetectEditors()

	if len(editors) == 0 {
		fmt.Println("❌ No compatible editors found")
		fmt.Println("\nSupported editors:")
		fmt.Println("  • VS Code (GitHub Copilot)")
		fmt.Println("  • Claude Desktop")
		fmt.Println("  • Cursor")
		fmt.Println("  • Amazon Q Developer")
		fmt.Println("  • Windsurf")
		return nil
	}

	fmt.Printf("✅ Found %d compatible editor(s)\n\n", len(editors))

	// Check each editor's MCP configuration
	allConfigured := true
	for _, ed := range editors {
		fmt.Printf("📝 %s:\n", ed.Name)

		configPath := installer.GetMCPConfigPath(ed, projectPath)
		if configPath == "" {
			fmt.Printf("   ❌ No config path available\n")
			allConfigured = false
			continue
		}

		fmt.Printf("   📁 Config file: %s\n", configPath)

		// Check if file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			fmt.Printf("   ❌ Config file does not exist\n")
			allConfigured = false
			continue
		}

		// Check if technocrat is configured
		if installer.IsMCPConfigured(ed, projectPath) {
			fmt.Printf("   ✅ Technocrat MCP server is configured\n")

			// Show the actual configuration
			if data, err := os.ReadFile(configPath); err == nil {
				var config map[string]interface{}
				if err := json.Unmarshal(data, &config); err == nil {
					showMCPConfig(ed, config)
				}
			}
		} else {
			fmt.Printf("   ❌ Technocrat MCP server is NOT configured\n")
			allConfigured = false
		}
		fmt.Println()
	}

	if allConfigured {
		fmt.Println("🎉 All detected editors are properly configured!")
		fmt.Println("\nNext steps:")
		fmt.Println("  1. Restart your editor(s) if you haven't already")
		fmt.Println("  2. Try using a Technocrat MCP prompt like 'tchncrt.spec'")
	} else {
		fmt.Println("⚠️  Some editors need configuration")
		fmt.Println("\nTo fix this:")
		fmt.Println("  • Run: technocrat configure-mcp --all")
		fmt.Println("  • Or: technocrat init --here (in existing project)")
	}

	return nil
}

func runConfigureMCP(cmd *cobra.Command, args []string) error {
	// Get flags
	editorNames, _ := cmd.Flags().GetStringSlice("editors")
	configureAll, _ := cmd.Flags().GetBool("all")

	// Get current directory as project path
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Detect available editors
	allEditors := editor.DetectEditors()
	if len(allEditors) == 0 {
		return fmt.Errorf("no compatible editors found")
	}

	// Determine which editors to configure
	var editorsToConfig []editor.Editor

	if configureAll {
		editorsToConfig = allEditors
	} else if len(editorNames) > 0 {
		// Filter by specified names
		editorMap := make(map[string]editor.Editor)
		for _, ed := range allEditors {
			editorMap[ed.Name] = ed
			// Also map common short names
			switch ed.Type {
			case editor.VSCode:
				editorMap["vscode"] = ed
				editorMap["code"] = ed
			case editor.ClaudeDesktop:
				editorMap["claude"] = ed
			case editor.Cursor:
				editorMap["cursor"] = ed
			case editor.AmazonQ:
				editorMap["amazonq"] = ed
				editorMap["q"] = ed
			case editor.Windsurf:
				editorMap["windsurf"] = ed
			}
		}

		for _, name := range editorNames {
			if ed, exists := editorMap[name]; exists {
				editorsToConfig = append(editorsToConfig, ed)
			} else {
				fmt.Printf("⚠️  Unknown editor: %s\n", name)
			}
		}
	} else {
		// Interactive selection
		fmt.Println("📝 Select editors to configure:")
		for i, ed := range allEditors {
			status := "❌ Not configured"
			if installer.IsMCPConfigured(ed, projectPath) {
				status = "✅ Already configured"
			}
			fmt.Printf("   %d. %s (%s)\n", i+1, ed.Name, status)
		}
		fmt.Print("\nEnter numbers (comma-separated) or 'all': ")

		var input string
		fmt.Scanln(&input)

		if input == "all" {
			editorsToConfig = allEditors
		} else {
			// Parse selection - simplified for now
			editorsToConfig = allEditors // Default to all if parsing fails
		}
	}

	if len(editorsToConfig) == 0 {
		return fmt.Errorf("no editors selected for configuration")
	}

	// Configure each selected editor
	fmt.Printf("\n🔧 Configuring MCP for %d editor(s)...\n\n", len(editorsToConfig))

	successCount := 0
	for _, ed := range editorsToConfig {
		fmt.Printf("📝 Configuring %s...\n", ed.Name)

		if err := installer.InstallMCPConfig(ed, projectPath); err != nil {
			fmt.Printf("   ❌ Failed: %v\n", err)
		} else {
			fmt.Printf("   ✅ Successfully configured\n")
			successCount++
		}

		configPath := installer.GetMCPConfigPath(ed, projectPath)
		if configPath != "" {
			fmt.Printf("   📁 Config saved to: %s\n", configPath)
		}
		fmt.Println()
	}

	if successCount > 0 {
		fmt.Printf("🎉 Successfully configured %d/%d editor(s)\n", successCount, len(editorsToConfig))
		fmt.Println("\nNext steps:")
		fmt.Println("  1. Restart your editor(s) to load the new configuration")
		fmt.Println("  2. Try using a Technocrat MCP prompt like 'tchncrt.spec'")
		fmt.Println("  3. Run 'technocrat check-mcp' to verify the configuration")
	} else {
		fmt.Println("❌ No editors were successfully configured")
		fmt.Println("\nTroubleshooting:")
		fmt.Println("  • Check that you have write permissions to the config directories")
		fmt.Println("  • Make sure the editors are properly installed")
		fmt.Println("  • Try running with --debug flag for more information")
	}

	return nil
}

func showMCPConfig(ed editor.Editor, config map[string]interface{}) {
	switch ed.Type {
	case editor.VSCode:
		if servers, ok := config["github.copilot.chat.mcpServers"].(map[string]interface{}); ok {
			if technocrat, ok := servers["technocrat"].(map[string]interface{}); ok {
				if command, ok := technocrat["command"].(string); ok {
					fmt.Printf("   🔧 Command: %s\n", command)
				}
			}
		}
	case editor.ClaudeDesktop:
		if servers, ok := config["mcpServers"].(map[string]interface{}); ok {
			if technocrat, ok := servers["technocrat"].(map[string]interface{}); ok {
				if command, ok := technocrat["command"].(string); ok {
					fmt.Printf("   🔧 Command: %s\n", command)
				}
			}
		}
	case editor.Cursor:
		if technocrat, ok := config["technocrat"].(map[string]interface{}); ok {
			if command, ok := technocrat["command"].(string); ok {
				fmt.Printf("   🔧 Command: %s\n", command)
			}
		}
	case editor.AmazonQ:
		if technocrat, ok := config["technocrat"].(map[string]interface{}); ok {
			if url, ok := technocrat["url"].(string); ok {
				fmt.Printf("   🔧 URL: %s\n", url)
			}
		}
	case editor.Windsurf:
		if technocrat, ok := config["technocrat"].(map[string]interface{}); ok {
			if command, ok := technocrat["command"].(string); ok {
				fmt.Printf("   🔧 Command: %s\n", command)
			}
		}
	}
}
