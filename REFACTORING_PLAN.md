# Technocrat Refactoring Plan: MCP-First Architecture

## Executive Summary

This refactoring transforms Technocrat from an agent-specific slash command system to a universal MCP (Model Context Protocol) server architecture. The key changes:

1. **Embedded Templates**: Templates compiled into the binary using Go's `embed` package
2. **MCP-First Commands**: Slash commands replaced by MCP prompts that invoke tools
3. **Editor-Based Installation**: Init detects installed editors and configures MCP server for them
4. **Auto-Configuration**: MCP server installed and configured automatically for VS Code, Claude, Cursor, etc.
5. **Bundled Commands**: Commands bundled with the MCP server, not as separate files

---

## Current State Analysis

### Problems with Current Architecture

1. **Template Distribution**: Templates downloaded from GitHub releases as zip files
2. **Agent-Specific Setup**: init.go generates different command files for each AI agent
3. **Maintenance Burden**: Supporting multiple agent formats (Markdown, TOML, etc.)
4. **Limited Portability**: Slash commands don't work well in all environments (Amazon Q)
5. **Manual Configuration**: Users must manually configure MCP servers
6. **Coupling**: Init logic tightly coupled to agent ecosystem

### Current Components

```
init.go
├── Agent selection (copilot, claude, gemini, etc.)
├── Script type selection (bash/powershell)
├── Template download from GitHub
├── Template extraction to project
├── Agent-specific command file generation
└── Git initialization

server.go
├── Basic MCP server (HTTP)
├── Simple tool examples (echo, system_info)
├── Basic resource examples
└── Simple prompt examples (welcome)

templates/
├── spec-template.md
├── plan-template.md
├── tasks-template.md
├── checklist-template.md
├── agent-file-template.md
├── vscode-settings.json
└── commands/ (will become MCP prompts)
    ├── spec.md
    ├── plan.md
    ├── tasks.md
    ├── implement.md
    ├── constitution.md
    ├── checklist.md
    ├── clarify.md
    └── analyze.md
```

---

## Target Architecture

### New Component Structure

```
internal/
├── cmd/
│   ├── init.go (editor detection & MCP config)
│   ├── server.go (HTTP + stdio support)
│   ├── create_feature.go
│   ├── setup_plan.go
│   └── check_prerequisites.go
├── installer/
│   ├── editor_detect.go (NEW - detect installed editors)
│   ├── mcp_config.go (NEW - generate MCP configs)
│   ├── vscode.go (NEW - VS Code MCP setup)
│   ├── claude.go (NEW - Claude Desktop setup)
│   ├── cursor.go (NEW - Cursor setup)
│   └── amazonq.go (NEW - Amazon Q setup)
├── mcp/
│   ├── server.go (HTTP + stdio transports)
│   ├── handler.go (protocol routing)
│   ├── tools.go (NEW - command tools)
│   ├── prompts.go (NEW - workflow prompts)
│   └── resources.go (NEW - template resources)
└── templates/
    ├── embedded.go (NEW - embed directives)
    └── accessor.go (NEW - read functions)
```

### MCP Architecture

```
Editor (VS Code, Claude Desktop, Cursor, etc.)
    │
    │ Launches technocrat server (stdio or HTTP)
    │
    ↓
Technocrat MCP Server (embedded in binary)
    │
    ├── Prompts (embedded from templates/commands/)
    │   ├── tchncrt.spec
    │   ├── tchncrt.plan
    │   ├── tchncrt.tasks
    │   └── tchncrt.implement
    │
    ├── Tools (call internal/cmd functions)
    │   ├── create_spec
    │   ├── create_plan
    │   ├── create_tasks
    │   └── run_implementation
    │
    └── Resources (embedded templates)
        ├── template://spec-template.md
        ├── template://plan-template.md
        └── template://tasks-template.md

User invokes prompt in editor → Server executes tool → Returns result
```

---

## Phase 1: Embed Templates

### Goal

Eliminate external template dependencies by embedding templates into the binary.

### Implementation

#### 1.1 Create Embedded Filesystem

**File**: `internal/templates/embedded.go`

```go
package templates

import (
 "embed"
 "io/fs"
)

//go:embed all:../../templates
var embeddedFS embed.FS

// GetFS returns the embedded filesystem
func GetFS() fs.FS {
 fsys, _ := fs.Sub(embeddedFS, "templates")
 return fsys
}
```

#### 1.2 Create Template Accessors

**File**: `internal/templates/accessor.go`

```go
package templates

import (
 "fmt"
 "io/fs"
 "path/filepath"
)

// GetTemplate reads a root-level template file
func GetTemplate(name string) ([]byte, error) {
 fsys := GetFS()
 return fs.ReadFile(fsys, name)
}

// GetCommandTemplate reads a command template
func GetCommandTemplate(name string) ([]byte, error) {
 fsys := GetFS()
 path := filepath.Join("commands", name)
 return fs.ReadFile(fsys, path)
}

// ListCommands returns all available command templates
func ListCommands() ([]string, error) {
 fsys := GetFS()
 var commands []string
 
 entries, err := fs.ReadDir(fsys, "commands")
 if err != nil {
  return nil, err
 }
 
 for _, entry := range entries {
  if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
   commands = append(commands, entry.Name())
  }
 }
 
 return commands, nil
}

// GetVSCodeSettings returns the VS Code settings template
func GetVSCodeSettings() ([]byte, error) {
 return GetTemplate("vscode-settings.json")
}
```

#### 1.3 Update Build Process

- Update `build.go` to ensure templates are included in binary
- Test that embedded files are accessible in compiled binary
- Verify file sizes and compression

### Testing

```go
// internal/templates/embedded_test.go
func TestEmbeddedTemplates(t *testing.T) {
 // Verify all expected templates exist
 templates := []string{
  "spec-template.md",
  "plan-template.md",
  "tasks-template.md",
  "checklist-template.md",
  "vscode-settings.json",
 }
 
 for _, tmpl := range templates {
  _, err := GetTemplate(tmpl)
  assert.NoError(t, err)
 }
}

func TestCommandTemplates(t *testing.T) {
 commands, err := ListCommands()
 assert.NoError(t, err)
 assert.Contains(t, commands, "spec.md")
 assert.Contains(t, commands, "plan.md")
}
```

---

## Phase 2: MCP Command Tools

### Goal

Create MCP tools that execute Technocrat commands directly.

### Tool Design

Each tool maps to an internal command:

| Tool Name | Command Function | Input Schema | Output |
|-----------|-----------------|--------------|--------|
| `create_spec` | `create_feature.go::RunCreateFeature` | `{description: string}` | `{branch: string, spec_file: string}` |
| `create_plan` | `setup_plan.go::RunSetupPlan` | `{}` | `{feature_dir: string, plan_file: string}` |
| `create_tasks` | (new) | `{}` | `{tasks_file: string}` |
| `run_implementation` | `check_prerequisites.go` | `{require_tasks: bool}` | `{status: string, docs: []}` |
| `create_constitution` | (new) | `{}` | `{constitution_file: string}` |
| `create_checklist` | (new) | `{}` | `{checklist_files: []}` |
| `clarify_requirements` | (new) | `{question: string}` | `{clarifications: []}` |
| `analyze_artifacts` | (new) | `{}` | `{analysis: string}` |

### Implementation

#### 2.1 Create Tools Module

**File**: `internal/mcp/tools.go`

```go
package mcp

import (
 "fmt"
 "technocrat/internal/cmd"
)

// RegisterCommandTools registers all Technocrat command tools
func (h *Handler) RegisterCommandTools() {
 h.RegisterTool(createSpecTool())
 h.RegisterTool(createPlanTool())
 h.RegisterTool(createTasksTool())
 h.RegisterTool(runImplementationTool())
 h.RegisterTool(createConstitutionTool())
 h.RegisterTool(createChecklistTool())
 h.RegisterTool(clarifyRequirementsTool())
 h.RegisterTool(analyzeArtifactsTool())
}

func createSpecTool() Tool {
 return Tool{
  Name:        "create_spec",
  Description: "Create a feature specification from a natural language description",
  InputSchema: map[string]interface{}{
   "type": "object",
   "properties": map[string]interface{}{
    "description": map[string]interface{}{
     "type":        "string",
     "description": "Natural language feature description",
    },
   },
   "required": []string{"description"},
  },
  Handler: func(args map[string]interface{}) (interface{}, error) {
   description, ok := args["description"].(string)
   if !ok {
    return nil, fmt.Errorf("description must be a string")
   }
   
   // Call internal command
   result, err := cmd.CreateFeature(description)
   if err != nil {
    return nil, err
   }
   
   return map[string]interface{}{
    "branch":    result.Branch,
    "spec_file": result.SpecFile,
    "success":   true,
   }, nil
  },
 }
}

func createPlanTool() Tool {
 return Tool{
  Name:        "create_plan",
  Description: "Create an implementation plan from the feature specification",
  InputSchema: map[string]interface{}{
   "type":       "object",
   "properties": map[string]interface{}{},
  },
  Handler: func(args map[string]interface{}) (interface{}, error) {
   result, err := cmd.SetupPlan()
   if err != nil {
    return nil, err
   }
   
   return map[string]interface{}{
    "feature_dir": result.FeatureDir,
    "plan_file":   result.PlanFile,
    "success":     true,
   }, nil
  },
 }
}

// ... similar implementations for other tools
```

#### 2.2 Refactor Command Functions

Make command functions library-friendly (not just CLI):

**Example**: `internal/cmd/create_feature.go`

```go
package cmd

// CreateFeatureResult holds the result of creating a feature
type CreateFeatureResult struct {
 Branch   string
 SpecFile string
}

// CreateFeature creates a new feature specification
// This function is used by both CLI and MCP server
func CreateFeature(description string) (*CreateFeatureResult, error) {
 // Existing logic from runCreateFeature
 // Returns structured result instead of printing
 
 return &CreateFeatureResult{
  Branch:   branch,
  SpecFile: specFile,
 }, nil
}

// runCreateFeature remains as CLI wrapper
func runCreateFeature(cmd *cobra.Command, args []string) error {
 result, err := CreateFeature(args[0])
 if err != nil {
  return err
 }
 
 // CLI-specific output formatting
 fmt.Printf("Branch: %s\n", result.Branch)
 fmt.Printf("Spec: %s\n", result.SpecFile)
 return nil
}
```

### Testing

```go
// internal/mcp/tools_test.go
func TestCreateSpecTool(t *testing.T) {
 h := NewHandler()
 h.RegisterCommandTools()
 
 args := map[string]interface{}{
  "description": "A user login feature with OAuth support",
 }
 
 result, err := h.CallTool("create_spec", args)
 assert.NoError(t, err)
 
 resultMap := result.(map[string]interface{})
 assert.NotEmpty(t, resultMap["branch"])
 assert.NotEmpty(t, resultMap["spec_file"])
 assert.True(t, resultMap["success"].(bool))
}
```

---

## Phase 3: MCP Prompts from Templates

### Goal

Convert command templates into MCP prompts that guide the AI through the workflow.

### Prompt Architecture

Each prompt:

1. Loads the command template from embedded filesystem
2. Parses the markdown to extract workflow instructions
3. Embeds the instructions as the prompt content
4. Includes a tool call recommendation
5. Provides argument templates

### Implementation

#### 3.1 Create Prompts Module

**File**: `internal/mcp/prompts.go`

```go
package mcp

import (
 "fmt"
 "strings"
 "technocrat/internal/templates"
)

// RegisterCommandPrompts registers all workflow prompts
func (h *Handler) RegisterCommandPrompts() {
 commands := []string{
  "spec", "plan", "tasks", "implement",
  "constitution", "checklist", "clarify", "analyze",
 }
 
 for _, cmd := range commands {
  if err := h.registerCommandPrompt(cmd); err != nil {
   // Log error but continue
   continue
  }
 }
}

func (h *Handler) registerCommandPrompt(commandName string) error {
 // Load template from embedded FS
 content, err := templates.GetCommandTemplate(commandName + ".md")
 if err != nil {
  return err
 }
 
 // Parse template to extract workflow
 workflow, toolCall := parseCommandTemplate(string(content))
 
 // Create prompt
 prompt := Prompt{
  Name:        "tchncrt." + commandName,
  Description: fmt.Sprintf("Execute the %s workflow", commandName),
  Arguments: []PromptArgument{
   {
    Name:        "user_input",
    Description: "Optional user input to guide the workflow",
    Required:    false,
   },
  },
  Handler: func(args map[string]interface{}) (interface{}, error) {
   userInput := ""
   if input, ok := args["user_input"].(string); ok {
    userInput = input
   }
   
   // Build prompt message with workflow instructions
   message := buildPromptMessage(workflow, toolCall, userInput)
   
   return map[string]interface{}{
    "messages": []map[string]string{
     {
      "role":    "user",
      "content": message,
     },
    },
   }, nil
  },
 }
 
 h.RegisterPrompt(prompt)
 return nil
}

func parseCommandTemplate(content string) (workflow string, toolCall string) {
 // Parse YAML frontmatter for command metadata
 // Extract workflow sections (## Outline, ## Phases, etc.)
 // Generate tool call syntax
 
 lines := strings.Split(content, "\n")
 inWorkflow := false
 var workflowLines []string
 
 for _, line := range lines {
  if strings.HasPrefix(line, "## Outline") {
   inWorkflow = true
   continue
  }
  if inWorkflow && strings.HasPrefix(line, "##") {
   break
  }
  if inWorkflow {
   workflowLines = append(workflowLines, line)
  }
 }
 
 workflow = strings.Join(workflowLines, "\n")
 
 // Determine tool call based on command
 // This maps command templates to tool names
 
 return workflow, toolCall
}

func buildPromptMessage(workflow, toolCall, userInput string) string {
 var sb strings.Builder
 
 sb.WriteString("# Technocrat Workflow\n\n")
 
 if userInput != "" {
  sb.WriteString("## User Input\n\n")
  sb.WriteString(userInput)
  sb.WriteString("\n\n")
 }
 
 sb.WriteString("## Instructions\n\n")
 sb.WriteString(workflow)
 sb.WriteString("\n\n")
 
 sb.WriteString("## Tool Call\n\n")
 sb.WriteString("To execute this workflow, call the following tool:\n\n")
 sb.WriteString("```\n")
 sb.WriteString(toolCall)
 sb.WriteString("\n```\n")
 
 return sb.String()
}
```

#### 3.2 Template-to-Tool Mapping

Create a mapping between templates and tools:

```go
var commandToolMap = map[string]string{
 "spec":         "create_spec",
 "plan":         "create_plan",
 "tasks":        "create_tasks",
 "implement":    "run_implementation",
 "constitution": "create_constitution",
 "checklist":    "create_checklist",
 "clarify":      "clarify_requirements",
 "analyze":      "analyze_artifacts",
}
```

### Testing

```go
// internal/mcp/prompts_test.go
func TestCommandPrompts(t *testing.T) {
 h := NewHandler()
 h.RegisterCommandPrompts()
 
 prompts := h.ListPrompts()
 
 expectedPrompts := []string{
  "tchncrt.spec",
  "tchncrt.plan",
  "tchncrt.tasks",
  "tchncrt.implement",
 }
 
 for _, expected := range expectedPrompts {
  found := false
  for _, p := range prompts {
   if p.Name == expected {
    found = true
    break
   }
  }
  assert.True(t, found, "Prompt %s not found", expected)
 }
}

func TestPromptExecution(t *testing.T) {
 h := NewHandler()
 h.RegisterCommandPrompts()
 
 args := map[string]interface{}{
  "user_input": "Create a login feature",
 }
 
 result, err := h.GetPrompt("tchncrt.spec", args)
 assert.NoError(t, err)
 
 resultMap := result.(map[string]interface{})
 messages := resultMap["messages"].([]map[string]string)
 assert.NotEmpty(t, messages)
 assert.Contains(t, messages[0]["content"], "User Input")
 assert.Contains(t, messages[0]["content"], "create_spec")
}
```

---

## Phase 4: Editor Detection and MCP Configuration

### Goal

Detect installed editors and automatically configure MCP server integration.

### Supported Editors

| Editor | Config File | Transport | Auto-Start |
|--------|------------|-----------|------------|
| **VS Code Copilot** | `.vscode/settings.json` | stdio | Yes |
| **Claude Desktop** | `~/Library/Application Support/Claude/claude_desktop_config.json` | stdio | Yes |
| **Cursor** | `~/.cursor/mcp_servers.json` | stdio | Yes |
| **Amazon Q CLI** | `~/.aws/q/mcp-config.json` | HTTP | No |
| **Windsurf** | `.windsurf/mcp_config.json` | stdio | Yes |

### Implementation

#### 4.1 Create Editor Detection Module

**File**: `internal/installer/editor_detect.go`

```go
package installer

import (
 "os"
 "os/exec"
 "path/filepath"
 "runtime"
)

// Editor represents a detected editor installation
type Editor struct {
 Name      string
 Type      EditorType
 Installed bool
 ConfigDir string
 Version   string
}

// EditorType defines the type of editor
type EditorType string

const (
 VSCode       EditorType = "vscode"
 ClaudeDesktop EditorType = "claude"
 Cursor       EditorType = "cursor"
 AmazonQ      EditorType = "amazonq"
 Windsurf     EditorType = "windsurf"
)

// DetectEditors scans for installed editors
func DetectEditors() ([]Editor, error) {
 editors := []Editor{
  detectVSCode(),
  detectClaudeDesktop(),
  detectCursor(),
  detectAmazonQ(),
  detectWindsurf(),
 }
 
 // Filter out non-installed editors
 installed := []Editor{}
 for _, editor := range editors {
  if editor.Installed {
   installed = append(installed, editor)
  }
 }
 
 return installed, nil
}

func detectVSCode() Editor {
 editor := Editor{
  Name:      "VS Code",
  Type:      VSCode,
  Installed: false,
 }
 
 // Try to find VS Code binary
 binNames := []string{"code", "code-insiders"}
 for _, bin := range binNames {
  if _, err := exec.LookPath(bin); err == nil {
   editor.Installed = true
   
   // Get version
   cmd := exec.Command(bin, "--version")
   if output, err := cmd.Output(); err == nil {
    editor.Version = string(output)
   }
   
   break
  }
 }
 
 // Determine config directory
 home, _ := os.UserHomeDir()
 switch runtime.GOOS {
 case "darwin":
  editor.ConfigDir = filepath.Join(home, "Library", "Application Support", "Code")
 case "linux":
  editor.ConfigDir = filepath.Join(home, ".config", "Code")
 case "windows":
  editor.ConfigDir = filepath.Join(os.Getenv("APPDATA"), "Code")
 }
 
 return editor
}

func detectClaudeDesktop() Editor {
 editor := Editor{
  Name:      "Claude Desktop",
  Type:      ClaudeDesktop,
  Installed: false,
 }
 
 home, _ := os.UserHomeDir()
 
 // Check for Claude Desktop config directory
 var configPath string
 switch runtime.GOOS {
 case "darwin":
  configPath = filepath.Join(home, "Library", "Application Support", "Claude")
 case "linux":
  configPath = filepath.Join(home, ".config", "Claude")
 case "windows":
  configPath = filepath.Join(os.Getenv("APPDATA"), "Claude")
 }
 
 if _, err := os.Stat(configPath); err == nil {
  editor.Installed = true
  editor.ConfigDir = configPath
 }
 
 return editor
}

func detectCursor() Editor {
 editor := Editor{
  Name:      "Cursor",
  Type:      Cursor,
  Installed: false,
 }
 
 // Try to find Cursor binary
 if _, err := exec.LookPath("cursor"); err == nil {
  editor.Installed = true
  
  // Get version
  cmd := exec.Command("cursor", "--version")
  if output, err := cmd.Output(); err == nil {
   editor.Version = string(output)
  }
 }
 
 // Determine config directory
 home, _ := os.UserHomeDir()
 editor.ConfigDir = filepath.Join(home, ".cursor")
 
 return editor
}

func detectAmazonQ() Editor {
 editor := Editor{
  Name:      "Amazon Q Developer CLI",
  Type:      AmazonQ,
  Installed: false,
 }
 
 // Check for Amazon Q CLI
 if _, err := exec.LookPath("q"); err == nil {
  editor.Installed = true
  
  // Get version
  cmd := exec.Command("q", "--version")
  if output, err := cmd.Output(); err == nil {
   editor.Version = string(output)
  }
 }
 
 // Config directory
 home, _ := os.UserHomeDir()
 editor.ConfigDir = filepath.Join(home, ".aws", "q")
 
 return editor
}

func detectWindsurf() Editor {
 editor := Editor{
  Name:      "Windsurf",
  Type:      Windsurf,
  Installed: false,
 }
 
 // Try to find Windsurf binary
 if _, err := exec.LookPath("windsurf"); err == nil {
  editor.Installed = true
 }
 
 // Config directory
 home, _ := os.UserHomeDir()
 switch runtime.GOOS {
 case "darwin":
  editor.ConfigDir = filepath.Join(home, "Library", "Application Support", "Windsurf")
 case "linux":
  editor.ConfigDir = filepath.Join(home, ".config", "Windsurf")
 case "windows":
  editor.ConfigDir = filepath.Join(os.Getenv("APPDATA"), "Windsurf")
 }
 
 return editor
}
```

#### 4.2 Create MCP Config Generator

**File**: `internal/installer/mcp_config.go`

```go
package installer

import (
 "encoding/json"
 "fmt"
 "os"
 "path/filepath"
)

// MCPServerConfig represents the MCP server configuration
type MCPServerConfig struct {
 Command string            `json:"command"`
 Args    []string          `json:"args"`
 Env     map[string]string `json:"env,omitempty"`
}

// InstallMCPConfig installs MCP server configuration for an editor
func InstallMCPConfig(editor Editor, projectPath string) error {
 switch editor.Type {
 case VSCode:
  return installVSCodeMCP(projectPath)
 case ClaudeDesktop:
  return installClaudeMCP(editor.ConfigDir)
 case Cursor:
  return installCursorMCP(editor.ConfigDir)
 case AmazonQ:
  return installAmazonQMCP(editor.ConfigDir)
 case Windsurf:
  return installWindsurfMCP(projectPath)
 default:
  return fmt.Errorf("unsupported editor: %s", editor.Name)
 }
}

func installVSCodeMCP(projectPath string) error {
 settingsPath := filepath.Join(projectPath, ".vscode", "settings.json")
 
 // Create .vscode directory if it doesn't exist
 if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
  return err
 }
 
 // Load existing settings or create new
 settings := make(map[string]interface{})
 if data, err := os.ReadFile(settingsPath); err == nil {
  json.Unmarshal(data, &settings)
 }
 
 // Get technocrat binary path
 technocratPath, _ := os.Executable()
 
 // Add MCP server configuration
 mcpServers := map[string]interface{}{
  "technocrat": map[string]interface{}{
   "command": technocratPath,
   "args":    []string{"server", "--stdio"},
   "env": map[string]string{
    "TECHNOCRAT_PROJECT": projectPath,
   },
  },
 }
 
 settings["mcp.servers"] = mcpServers
 
 // Write updated settings
 data, err := json.MarshalIndent(settings, "", "  ")
 if err != nil {
  return err
 }
 
 return os.WriteFile(settingsPath, data, 0644)
}

func installClaudeMCP(configDir string) error {
 configPath := filepath.Join(configDir, "claude_desktop_config.json")
 
 // Create config directory if it doesn't exist
 if err := os.MkdirAll(configDir, 0755); err != nil {
  return err
 }
 
 // Load existing config or create new
 config := make(map[string]interface{})
 if data, err := os.ReadFile(configPath); err == nil {
  json.Unmarshal(data, &config)
 }
 
 // Get technocrat binary path
 technocratPath, _ := os.Executable()
 
 // Add MCP server configuration
 mcpServers, ok := config["mcpServers"].(map[string]interface{})
 if !ok {
  mcpServers = make(map[string]interface{})
 }
 
 mcpServers["technocrat"] = map[string]interface{}{
  "command": technocratPath,
  "args":    []string{"server", "--stdio"},
 }
 
 config["mcpServers"] = mcpServers
 
 // Write updated config
 data, err := json.MarshalIndent(config, "", "  ")
 if err != nil {
  return err
 }
 
 return os.WriteFile(configPath, data, 0644)
}

func installCursorMCP(configDir string) error {
 configPath := filepath.Join(configDir, "mcp_servers.json")
 
 // Create config directory if it doesn't exist
 if err := os.MkdirAll(configDir, 0755); err != nil {
  return err
 }
 
 // Load existing config or create new
 config := make(map[string]interface{})
 if data, err := os.ReadFile(configPath); err == nil {
  json.Unmarshal(data, &config)
 }
 
 // Get technocrat binary path
 technocratPath, _ := os.Executable()
 
 // Add MCP server configuration
 config["technocrat"] = map[string]interface{}{
  "command": technocratPath,
  "args":    []string{"server", "--stdio"},
 }
 
 // Write updated config
 data, err := json.MarshalIndent(config, "", "  ")
 if err != nil {
  return err
 }
 
 return os.WriteFile(configPath, data, 0644)
}

func installAmazonQMCP(configDir string) error {
 configPath := filepath.Join(configDir, "mcp-config.json")
 
 // Create config directory if it doesn't exist
 if err := os.MkdirAll(configDir, 0755); err != nil {
  return err
 }
 
 // Get technocrat binary path
 technocratPath, _ := os.Executable()
 
 // Amazon Q uses HTTP, not stdio
 config := map[string]interface{}{
  "servers": map[string]interface{}{
   "technocrat": map[string]interface{}{
    "url": "http://localhost:8080",
    "description": "Technocrat Spec-Driven Development Server",
   },
  },
 }
 
 // Write config
 data, err := json.MarshalIndent(config, "", "  ")
 if err != nil {
  return err
 }
 
 if err := os.WriteFile(configPath, data, 0644); err != nil {
  return err
 }
 
 // For Amazon Q, also provide instructions to start server manually
 fmt.Println("\n⚠️  Amazon Q requires manual server start:")
 fmt.Printf("   Run: %s server --port 8080\n", technocratPath)
 
 return nil
}

func installWindsurfMCP(projectPath string) error {
 configPath := filepath.Join(projectPath, ".windsurf", "mcp_config.json")
 
 // Create config directory if it doesn't exist
 if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
  return err
 }
 
 // Get technocrat binary path
 technocratPath, _ := os.Executable()
 
 // Create config
 config := map[string]interface{}{
  "servers": map[string]interface{}{
   "technocrat": map[string]interface{}{
    "command": technocratPath,
    "args":    []string{"server", "--stdio"},
   },
  },
 }
 
 // Write config
 data, err := json.MarshalIndent(config, "", "  ")
 if err != nil {
  return err
 }
 
 return os.WriteFile(configPath, data, 0644)
}
```

### Testing

```go
// internal/installer/mcp_config_test.go
func TestDetectEditors(t *testing.T) {
 editors, err := DetectEditors()
 assert.NoError(t, err)
 
 // Should detect at least one editor in dev environment
 assert.NotEmpty(t, editors)
 
 for _, editor := range editors {
  assert.True(t, editor.Installed)
  assert.NotEmpty(t, editor.ConfigDir)
 }
}

func TestInstallVSCodeMCP(t *testing.T) {
 // Create temp project directory
 tmpDir := t.TempDir()
 
 // Install MCP config
 err := installVSCodeMCP(tmpDir)
 assert.NoError(t, err)
 
 // Verify settings.json was created
 settingsPath := filepath.Join(tmpDir, ".vscode", "settings.json")
 assert.FileExists(t, settingsPath)
 
 // Verify config structure
 data, _ := os.ReadFile(settingsPath)
 var settings map[string]interface{}
 json.Unmarshal(data, &settings)
 
 assert.Contains(t, settings, "mcp.servers")
 mcpServers := settings["mcp.servers"].(map[string]interface{})
 assert.Contains(t, mcpServers, "technocrat")
}
```

---

## Phase 5: Add stdio Transport to Server

### Goal

Support both HTTP and stdio transports for MCP protocol communication.

### Why Both Transports?

- **stdio**: Used by editors that launch MCP servers as child processes (VS Code, Claude, Cursor)
- **HTTP**: Used by standalone clients or editors that connect to running servers (Amazon Q)

### Implementation

**File**: `internal/mcp/server.go` (additions)

```go
package mcp

import (
 "bufio"
 "encoding/json"
 "fmt"
 "io"
 "log"
 "os"
)

// StartStdio starts the MCP server in stdio mode
func (s *Server) StartStdio() error {
 log.SetOutput(os.Stderr) // Log to stderr, keep stdout for JSON-RPC
 
 log.Println("Starting Technocrat MCP Server in stdio mode...")
 
 reader := bufio.NewReader(os.Stdin)
 writer := bufio.NewWriter(os.Stdout)
 
 for {
  // Read JSON-RPC message from stdin
  line, err := reader.ReadBytes('\n')
  if err != nil {
   if err == io.EOF {
    break
   }
   return fmt.Errorf("error reading from stdin: %w", err)
  }
  
  // Parse JSON-RPC request
  var request map[string]interface{}
  if err := json.Unmarshal(line, &request); err != nil {
   s.writeError(writer, nil, -32700, "Parse error")
   continue
  }
  
  // Handle request
  response := s.handleStdioRequest(request)
  
  // Write response to stdout
  responseBytes, _ := json.Marshal(response)
  writer.Write(responseBytes)
  writer.WriteByte('\n')
  writer.Flush()
 }
 
 return nil
}

func (s *Server) handleStdioRequest(request map[string]interface{}) map[string]interface{} {
 method, ok := request["method"].(string)
 if !ok {
  return s.errorResponse(request["id"], -32600, "Invalid Request")
 }
 
 // Route to appropriate handler
 switch method {
 case "initialize":
  return s.handleStdioInitialize(request)
 case "tools/list":
  return s.handleStdioToolsList(request)
 case "tools/call":
  return s.handleStdioToolsCall(request)
 case "prompts/list":
  return s.handleStdioPromptsList(request)
 case "prompts/get":
  return s.handleStdioPromptsGet(request)
 case "resources/list":
  return s.handleStdioResourcesList(request)
 case "resources/read":
  return s.handleStdioResourcesRead(request)
 default:
  return s.errorResponse(request["id"], -32601, "Method not found")
 }
}

func (s *Server) errorResponse(id interface{}, code int, message string) map[string]interface{} {
 return map[string]interface{}{
  "jsonrpc": "2.0",
  "id":      id,
  "error": map[string]interface{}{
   "code":    code,
   "message": message,
  },
 }
}

func (s *Server) successResponse(id interface{}, result interface{}) map[string]interface{} {
 return map[string]interface{}{
  "jsonrpc": "2.0",
  "id":      id,
  "result":  result,
 }
}

// Implement stdio-specific handlers...
func (s *Server) handleStdioInitialize(request map[string]interface{}) map[string]interface{} {
 result := map[string]interface{}{
  "protocolVersion": "2024-11-05",
  "serverInfo": map[string]string{
   "name":    "technocrat",
   "version": "2.0.0",
  },
  "capabilities": map[string]interface{}{
   "tools":     map[string]interface{}{"count": len(s.handler.tools)},
   "resources": map[string]interface{}{"count": len(s.handler.resources)},
   "prompts":   map[string]interface{}{"count": len(s.handler.prompts)},
  },
 }
 
 return s.successResponse(request["id"], result)
}

// ... similar implementations for other handlers
```

**File**: `internal/cmd/server.go` (update)

```go
var (
 serverPort int
 stdio      bool
 verbose    bool
)

func init() {
 rootCmd.AddCommand(serverCmd)

 serverCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "Port to listen on (HTTP mode)")
 serverCmd.Flags().BoolVar(&stdio, "stdio", false, "Use stdio transport instead of HTTP")
 serverCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
}

func runServer(cmd *cobra.Command, args []string) error {
 if verbose {
  log.SetFlags(log.LstdFlags | log.Lshortfile)
 }
 
 server := mcp.NewServer(serverPort)
 
 // Register all capabilities
 log.Printf("Registering command tools...")
 server.RegisterCommandTools()
 
 log.Printf("Registering workflow prompts...")
 server.RegisterCommandPrompts()
 
 log.Printf("Registering template resources...")
 server.RegisterTemplateResources()
 
 // Start server in appropriate mode
 if stdio {
  log.Println("Starting in stdio mode...")
  return server.StartStdio()
 } else {
  log.Printf("Starting HTTP server on port %d...", serverPort)
  return server.Start()
 }
}
```

---

## Phase 6: Update init.go with Editor Configuration

### Goal

Expose embedded templates as MCP resources for client access.

### Implementation

#### 4.1 Create Resources Module

**File**: `internal/mcp/resources.go`

```go
package mcp

import (
 "technocrat/internal/templates"
)

// RegisterTemplateResources registers all embedded templates as resources
func (h *Handler) RegisterTemplateResources() {
 // Root templates
 rootTemplates := []string{
  "spec-template.md",
  "plan-template.md",
  "tasks-template.md",
  "checklist-template.md",
  "agent-file-template.md",
 }
 
 for _, name := range rootTemplates {
  h.RegisterResource(Resource{
   URI:         "template://" + name,
   Name:        name,
   Description: "Template for " + name,
   MimeType:    "text/markdown",
  })
 }
 
 // Command templates
 commands, _ := templates.ListCommands()
 for _, cmd := range commands {
  h.RegisterResource(Resource{
   URI:         "template://commands/" + cmd,
   Name:        "Command: " + cmd,
   Description: "Workflow template for " + cmd,
   MimeType:    "text/markdown",
  })
 }
 
 // VS Code settings
 h.RegisterResource(Resource{
  URI:         "template://vscode-settings.json",
  Name:        "VS Code Settings",
  Description: "Recommended VS Code settings for Technocrat projects",
  MimeType:    "application/json",
 })
}

// Override ReadResource to serve embedded templates
func (h *Handler) ReadTemplateResource(uri string) (interface{}, error) {
 // Extract path from URI (remove "template://" prefix)
 path := strings.TrimPrefix(uri, "template://")
 
 var content []byte
 var err error
 
 if strings.HasPrefix(path, "commands/") {
  cmdName := strings.TrimPrefix(path, "commands/")
  content, err = templates.GetCommandTemplate(cmdName)
 } else {
  content, err = templates.GetTemplate(path)
 }
 
 if err != nil {
  return nil, err
 }
 
 return map[string]interface{}{
  "uri":      uri,
  "mimeType": guessMimeType(path),
  "text":     string(content),
 }, nil
}

func guessMimeType(path string) string {
 if strings.HasSuffix(path, ".md") {
  return "text/markdown"
 }
 if strings.HasSuffix(path, ".json") {
  return "application/json"
 }
 return "text/plain"
}
```

#### 4.2 Update Handler

**File**: `internal/mcp/handler.go`

```go
// Modify ReadResource to delegate to template reader
func (h *Handler) ReadResource(uri string) (interface{}, error) {
 if strings.HasPrefix(uri, "template://") {
  return h.ReadTemplateResource(uri)
 }
 
 // Existing logic for other resources
 resource, exists := h.resources[uri]
 if !exists {
  return nil, fmt.Errorf("resource not found: %s", uri)
 }
 
 // ... existing code
}
```

---

## Phase 6: Update init.go with Editor Configuration

### Goal
Transform init.go to detect editors and install MCP server configurations.

### New init.go Flow

```
1. Parse arguments (project name, --here flag)
2. Detect installed editors
3. Prompt user to select editor(s)
4. Create project directory structure
5. Install MCP server config for selected editor(s)
6. Copy VS Code settings (if applicable)
7. Initialize git repository
8. Show success message with next steps
```

### Implementation

**File**: `internal/cmd/init.go` (complete rewrite)

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	
	"technocrat/internal/installer"
	"technocrat/internal/templates"
	"technocrat/internal/ui"
	
	"github.com/spf13/cobra"
)

var (
	here  bool
	force bool
	noGit bool
	aiEditor string
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new Technocrat project",
	Long: `Initialize a new Technocrat project with MCP server integration.

This command will:
  1. Detect installed editors (VS Code, Claude, Cursor, etc.)
  2. Let you choose your editor (or use --editor flag)
  3. Install MCP server configuration for your editor
  4. Create project directory structure
  5. Initialize a git repository (unless --no-git)

The MCP server provides workflow prompts and tools that work
directly in your editor without needing separate command files.

Examples:
  technocrat init my-project
  technocrat init my-project --editor vscode
  technocrat init . --editor claude         # Initialize in current directory
  technocrat init --here --editor cursor    # Alternative syntax
  technocrat init --here --force            # Skip confirmation for non-empty dir`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	
	initCmd.Flags().StringVar(&aiEditor, "editor", "", "Editor to configure (vscode, claude, cursor, amazonq, windsurf)")
	initCmd.Flags().BoolVar(&here, "here", false, "Initialize in current directory")
	initCmd.Flags().BoolVar(&force, "force", false, "Force overwrite when using --here")
	initCmd.Flags().BoolVar(&noGit, "no-git", false, "Skip git repository initialization")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Show banner
	showBanner()
	
	// Determine project path
	projectPath, projectName, err := getProjectPath(args)
	if err != nil {
		return err
	}
	
	// Print setup information
	cwd, _ := os.Getwd()
	setupInfo := fmt.Sprintf("Project:      %s\nWorking Path: %s", projectName, cwd)
	if !here {
		setupInfo += fmt.Sprintf("\nTarget Path:  %s", projectPath)
	}
	ui.ShowInfo("Technocrat Project Setup", setupInfo)
	
	// Detect installed editors
	editors, err := installer.DetectEditors()
	if err != nil {
		return fmt.Errorf("failed to detect editors: %w", err)
	}
	
	if len(editors) == 0 {
		ui.ShowWarning("No Editors Detected", 
			"No supported editors were found.\n"+
			"MCP server can still be started manually.\n"+
			"See: https://github.com/x86ed/technocrat/docs/mcp-server.md")
	}
	
	// Select editor
	var selectedEditor *installer.Editor
	if aiEditor != "" {
		// Validate provided editor
		selectedEditor = findEditorByType(editors, aiEditor)
		if selectedEditor == nil {
			return fmt.Errorf("editor %s not found or not installed", aiEditor)
		}
	} else if len(editors) > 0 {
		// Prompt user to select
		selectedEditor, err = promptForEditor(editors)
		if err != nil {
			return err
		}
	}
	
	if selectedEditor != nil {
		fmt.Fprintf(os.Stderr, "\n✓ Selected editor: %s\n\n", selectedEditor.Name)
	}
	
	// Create progress tracker
	tracker := ui.NewStepTracker("Installation Progress")
	tracker.Add("structure", "Creating project structure")
	if selectedEditor != nil {
		tracker.Add("mcp", "Installing MCP server configuration")
	}
	if !noGit && !isGitRepo(projectPath) {
		tracker.Add("git", "Initializing git repository")
	}
	
	// Start live tracker
	if ui.IsInteractive() {
		tracker.StartLive()
	}
	
	// Create project structure
	tracker.Start("structure")
	if err := createProjectStructure(projectPath); err != nil {
		tracker.Fail("structure", err.Error())
		tracker.StopLive()
		return err
	}
	tracker.Complete("structure")
	
	// Install MCP server configuration
	if selectedEditor != nil {
		tracker.Start("mcp")
		if err := installer.InstallMCPConfig(*selectedEditor, projectPath); err != nil {
			tracker.Fail("mcp", err.Error())
			tracker.StopLive()
			return err
		}
		tracker.Complete("mcp")
	}
	
	// Initialize git repository
	if !noGit && !isGitRepo(projectPath) {
		tracker.Start("git")
		if err := initGitRepo(projectPath); err != nil {
			tracker.Fail("git", err.Error())
			// Non-fatal, continue
		} else {
			tracker.Complete("git")
		}
	}
	
	// Stop live tracker
	tracker.StopLive()
	
	// Print final summary
	fmt.Fprintf(os.Stderr, "\n%s %s\n", ui.ColorGreen.Sprint(ui.SymbolCheckmark), tracker.Summary())
	
	// Show success message
	showSuccessMessage(projectName, selectedEditor)
	
	return nil
}

func getProjectPath(args []string) (string, string, error) {
	if here {
		cwd, err := os.Getwd()
		if err != nil {
			return "", "", err
		}
		
		// Check if directory is empty (unless --force)
		if !force {
			entries, err := os.ReadDir(cwd)
			if err != nil {
				return "", "", err
			}
			
			// Filter out hidden files
			visibleEntries := 0
			for _, entry := range entries {
				if !strings.HasPrefix(entry.Name(), ".") {
					visibleEntries++
				}
			}
			
			if visibleEntries > 0 {
				return "", "", fmt.Errorf("directory is not empty (use --force to override)")
			}
		}
		
		projectName := filepath.Base(cwd)
		return cwd, projectName, nil
	}
	
	if len(args) == 0 {
		return "", "", fmt.Errorf("project name required (or use --here)")
	}
	
	projectName := args[0]
	projectPath := filepath.Join(".", projectName)
	
	// Check if directory exists
	if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
		return "", "", fmt.Errorf("directory %s already exists", projectPath)
	}
	
	return projectPath, projectName, nil
}

func createProjectStructure(projectPath string) error {
	dirs := []string{
		".technocrat",
		"memory",
		".vscode",
	}
	
	for _, dir := range dirs {
		fullPath := filepath.Join(projectPath, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	// Copy VS Code settings from embedded templates
	content, err := templates.GetVSCodeSettings()
	if err != nil {
		return fmt.Errorf("failed to load VS Code settings: %w", err)
	}
	
	destPath := filepath.Join(projectPath, ".vscode", "settings.json")
	if err := os.WriteFile(destPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write VS Code settings: %w", err)
	}
	
	return nil
}

func findEditorByType(editors []installer.Editor, editorType string) *installer.Editor {
	for _, editor := range editors {
		if string(editor.Type) == editorType {
			return &editor
		}
	}
	return nil
}

func promptForEditor(editors []installer.Editor) (*installer.Editor, error) {
	if !ui.IsInteractive() {
		// Non-interactive mode, skip editor selection
		fmt.Fprintln(os.Stderr, "Non-interactive mode: skipping editor selection")
		return nil, nil
	}
	
	// Build options for selector
	options := make([]string, len(editors)+1)
	for i, editor := range editors {
		options[i] = fmt.Sprintf("%s (%s)", editor.Name, editor.Type)
	}
	options[len(editors)] = "Skip (configure manually later)"
	
	fmt.Fprintln(os.Stderr, "\nDetected editors:")
	for i, opt := range options {
		fmt.Fprintf(os.Stderr, "  %d. %s\n", i+1, opt)
	}
	
	// Use selector
	selected, err := ui.SelectOption("Select editor to configure", options)
	if err != nil {
		return nil, err
	}
	
	// If "Skip" was selected
	if selected == len(editors) {
		return nil, nil
	}
	
	return &editors[selected], nil
}

func showSuccessMessage(projectName string, editor *installer.Editor) {
	fmt.Fprintln(os.Stderr, "\n═══════════════════════════════════════════════════════════")
	fmt.Fprintln(os.Stderr, "  ✓ Project initialized successfully!")
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
	
	// Build next steps
	var nextSteps strings.Builder
	stepNum := 1
	
	if !here {
		nextSteps.WriteString(fmt.Sprintf("%d. Navigate to your project:\n", stepNum))
		nextSteps.WriteString(fmt.Sprintf("   cd %s\n\n", projectName))
		stepNum++
	}
	
	if editor != nil {
		nextSteps.WriteString(fmt.Sprintf("%d. Open your project in %s\n\n", stepNum, editor.Name))
		stepNum++
		
		nextSteps.WriteString(fmt.Sprintf("%d. The MCP server is configured and ready to use!\n\n", stepNum))
		stepNum++
		
		nextSteps.WriteString(fmt.Sprintf("%d. Use workflow prompts in your editor:\n\n", stepNum))
	} else {
		nextSteps.WriteString(fmt.Sprintf("%d. Start the Technocrat MCP server:\n", stepNum))
		nextSteps.WriteString("   technocrat server\n\n")
		stepNum++
		
		nextSteps.WriteString(fmt.Sprintf("%d. Configure your editor to connect to the server\n", stepNum))
		nextSteps.WriteString("   See: https://github.com/x86ed/technocrat/docs/mcp-server.md\n\n")
		stepNum++
		
		nextSteps.WriteString(fmt.Sprintf("%d. Use workflow prompts:\n\n", stepNum))
	}
	
	// Core workflow commands
	nextSteps.WriteString("   Core workflow:\n")
	nextSteps.WriteString("     • tchncrt.constitution - Establish project principles\n")
	nextSteps.WriteString("     • tchncrt.spec         - Create feature specification\n")
	nextSteps.WriteString("     • tchncrt.plan         - Create implementation plan\n")
	nextSteps.WriteString("     • tchncrt.tasks        - Generate actionable tasks\n")
	nextSteps.WriteString("     • tchncrt.implement    - Execute implementation\n\n")
	
	nextSteps.WriteString("   Enhancement commands:\n")
	nextSteps.WriteString("     • tchncrt.clarify   - Ask structured questions\n")
	nextSteps.WriteString("     • tchncrt.checklist - Quality validation checklists\n")
	nextSteps.WriteString("     • tchncrt.analyze   - Cross-artifact consistency check\n")
	
	ui.ShowSuccess("Next Steps", nextSteps.String())
	
	// Editor-specific tips
	if editor != nil {
		var tips string
		switch editor.Type {
		case installer.VSCode:
			tips = "In VS Code:\n" +
				"• Open Command Palette (Cmd/Ctrl+Shift+P)\n" +
				"• Type 'MCP' to see available prompts\n" +
				"• Or use Copilot Chat with prompt names"
		case installer.ClaudeDesktop:
			tips = "In Claude Desktop:\n" +
				"• Look for MCP prompts in the sidebar\n" +
				"• Or type prompt names directly in chat"
		case installer.Cursor:
			tips = "In Cursor:\n" +
				"• Open Composer (Cmd/Ctrl+K)\n" +
				"• MCP prompts available in command menu\n" +
				"• Or type '@tchncrt' to access prompts"
		case installer.AmazonQ:
			tips = "For Amazon Q:\n" +
				"• Start the server: technocrat server --port 8080\n" +
				"• Prompts will appear in Q Developer interface"
		case installer.Windsurf:
			tips = "In Windsurf:\n" +
				"• Open Flow (Cmd/Ctrl+L)\n" +
				"• MCP prompts available in workflow menu"
		}
		
		if tips != "" {
			ui.ShowInfo(fmt.Sprintf("Using %s", editor.Name), tips)
		}
	}
}
```

### What Gets Removed from init.go

- ❌ `agentConfigs` map (entire thing)
- ❌ Template downloading from GitHub
- ❌ Zip file extraction
- ❌ `downloadAndExtractTemplate` function
- ❌ `makeScriptsExecutable` function
- ❌ Script type selection (bash/powershell)
- ❌ Agent-specific command file generation
- ❌ CLI tool checking for agents

### What Gets Added to init.go

- ✅ Editor detection via `installer.DetectEditors()`
- ✅ Editor selection prompt
- ✅ MCP config installation via `installer.InstallMCPConfig()`
- ✅ Editor-specific success messages
- ✅ Simplified project structure creation
- ✅ Direct use of embedded templates

---

## Phase 7: Update Architecture Documentation

### Updated Workflow Diagram

**Before (v1.x):**
```
User runs: technocrat init my-project --ai claude
  ↓
Downloads templates from GitHub release
  ↓
Generates .claude/commands/*.md files
  ↓
User types: /tchncrt.spec "Add login feature"
  ↓
Claude reads command file from disk
  ↓
Executes workflow manually
```

**After (v2.0):**
```
User runs: technocrat init my-project
  ↓
Detects installed editors
  ↓
User selects: VS Code
  ↓
Writes MCP config to .vscode/settings.json
  ↓
User opens project in VS Code
  ↓
VS Code auto-starts: technocrat server --stdio
  ↓
MCP server loads embedded prompts/tools
  ↓
User invokes: tchncrt.spec prompt
  ↓
Prompt guides AI through workflow
  ↓
AI calls: create_spec tool
  ↓
Tool executes and returns results
```

### Key Improvements

1. **No Template Downloads**: Everything embedded in binary
2. **Auto-Configuration**: MCP server configured during init
3. **Auto-Start**: Editors launch server automatically (stdio mode)
4. **Unified Interface**: Same prompts work in all editors
5. **Live Execution**: Tools execute immediately, no manual steps

---

## Phase 8: Template Resources

### Goal

Strip init.go down to essential project setup only.

### New init.go Structure

```go
package cmd

import (
 "fmt"
 "os"
 "path/filepath"
 "technocrat/internal/templates"
 "technocrat/internal/ui"
 
 "github.com/spf13/cobra"
)

var (
 here  bool
 force bool
 noGit bool
)

var initCmd = &cobra.Command{
 Use:   "init [project-name]",
 Short: "Initialize a new Technocrat project",
 Long: `Initialize a new Technocrat project structure.

This command creates a minimal project structure with:
  - .technocrat/ directory for specs and plans
  - .vscode/settings.json with recommended settings
  - memory/ directory for project constitution
  - Git repository (unless --no-git)

To use Technocrat commands, start the MCP server:
  technocrat server

Then connect your AI assistant (Amazon Q, Claude, etc.) to the server.`,
 RunE: runInit,
}

func init() {
 rootCmd.AddCommand(initCmd)
 
 initCmd.Flags().BoolVar(&here, "here", false, "Initialize in current directory")
 initCmd.Flags().BoolVar(&force, "force", false, "Force overwrite when using --here")
 initCmd.Flags().BoolVar(&noGit, "no-git", false, "Skip git repository initialization")
}

func runInit(cmd *cobra.Command, args []string) error {
 // Determine project path
 projectPath, err := getProjectPath(args)
 if err != nil {
  return err
 }
 
 // Create directory structure
 if err := createProjectStructure(projectPath); err != nil {
  return err
 }
 
 // Copy VS Code settings from embedded templates
 if err := copyVSCodeSettings(projectPath); err != nil {
  return err
 }
 
 // Initialize git repository
 if !noGit {
  if err := initGitRepo(projectPath); err != nil {
   // Non-fatal, just warn
   fmt.Fprintf(os.Stderr, "Warning: Git initialization failed: %v\n", err)
  }
 }
 
 // Show success message
 showSuccessMessage(projectPath)
 
 return nil
}

func getProjectPath(args []string) (string, error) {
 if here {
  cwd, err := os.Getwd()
  if err != nil {
   return "", err
  }
  
  // Check if directory is empty (unless --force)
  if !force {
   if err := checkDirectoryEmpty(cwd); err != nil {
    return "", err
   }
  }
  
  return cwd, nil
 }
 
 if len(args) == 0 {
  return "", fmt.Errorf("project name required (or use --here)")
 }
 
 projectPath := filepath.Join(".", args[0])
 
 // Check if directory exists
 if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
  return "", fmt.Errorf("directory %s already exists", projectPath)
 }
 
 return projectPath, nil
}

func createProjectStructure(projectPath string) error {
 dirs := []string{
  ".technocrat",
  "memory",
  ".vscode",
 }
 
 for _, dir := range dirs {
  fullPath := filepath.Join(projectPath, dir)
  if err := os.MkdirAll(fullPath, 0755); err != nil {
   return fmt.Errorf("failed to create directory %s: %w", dir, err)
  }
 }
 
 return nil
}

func copyVSCodeSettings(projectPath string) error {
 content, err := templates.GetVSCodeSettings()
 if err != nil {
  return fmt.Errorf("failed to load VS Code settings: %w", err)
 }
 
 destPath := filepath.Join(projectPath, ".vscode", "settings.json")
 if err := os.WriteFile(destPath, content, 0644); err != nil {
  return fmt.Errorf("failed to write VS Code settings: %w", err)
 }
 
 return nil
}

func showSuccessMessage(projectPath string) {
 fmt.Fprintln(os.Stderr, "\n═══════════════════════════════════════════════════════════")
 fmt.Fprintln(os.Stderr, "  ✓ Project initialized successfully!")
 fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
 
 ui.ShowSuccess("Next Steps", `
1. Start the Technocrat MCP server:
   technocrat server

2. Configure your AI assistant to connect to the MCP server:
   - Amazon Q Developer: Add server to config
   - Claude Desktop: Add to claude_desktop_config.json
   - VS Code Copilot: Use MCP extension

3. Use prompts with your AI assistant:
   - tchncrt.constitution - Establish project principles
   - tchncrt.spec         - Create feature specification
   - tchncrt.plan         - Create implementation plan
   - tchncrt.tasks        - Generate actionable tasks
   - tchncrt.implement    - Execute implementation

For detailed setup instructions, see:
https://github.com/x86ed/technocrat/docs/mcp-server.md
`)
}
```

### What Gets Removed

- ❌ `agentConfigs` map
- ❌ Agent selection prompts
- ❌ Script type selection
- ❌ Template downloading from GitHub
- ❌ Zip file extraction
- ❌ Agent-specific command file generation
- ❌ `makeScriptsExecutable` for templates
- ❌ CLI tool checking for agents

### What Stays

- ✅ Project directory creation
- ✅ Git initialization
- ✅ VS Code settings copy
- ✅ Directory structure validation
- ✅ `--here` and `--force` flags

---

## Phase 6: Update server.go

### Goal

Register all new tools, prompts, and resources in the MCP server.

### Implementation

**File**: `internal/cmd/server.go`

```go
package cmd

import (
 "fmt"
 "log"

 "github.com/spf13/cobra"
 "technocrat/internal/mcp"
)

var (
 serverPort int
 verbose    bool
)

var serverCmd = &cobra.Command{
 Use:   "server",
 Short: "Start the Technocrat MCP server",
 Long: `Start the Technocrat Model Context Protocol (MCP) server.

The server provides:
  - Tools: Execute Technocrat commands (create_spec, create_plan, etc.)
  - Prompts: Workflow guidance (tchncrt.spec, tchncrt.plan, etc.)
  - Resources: Access to embedded templates

AI assistants can connect to this server to use Technocrat's
Spec-Driven Development workflow through the MCP protocol.

Supported clients:
  - Amazon Q Developer CLI
  - Claude Desktop
  - VS Code with MCP extension
  - Any MCP-compatible client

Examples:
  technocrat server
  technocrat server --port 8080
  technocrat server --port 8080 --verbose`,
 RunE: runServer,
}

func init() {
 rootCmd.AddCommand(serverCmd)

 serverCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "Port to listen on")
 serverCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
}

func runServer(cmd *cobra.Command, args []string) error {
 if verbose {
  log.SetFlags(log.LstdFlags | log.Lshortfile)
 }
 
 log.Printf("Initializing Technocrat MCP Server...")
 
 server := mcp.NewServer(serverPort)
 
 // Register all command tools
 log.Printf("Registering command tools...")
 server.RegisterCommandTools()
 
 // Register all workflow prompts
 log.Printf("Registering workflow prompts...")
 server.RegisterCommandPrompts()
 
 // Register template resources
 log.Printf("Registering template resources...")
 server.RegisterTemplateResources()
 
 log.Printf("Starting server on port %d...", serverPort)
 log.Printf("MCP endpoints available at:")
 log.Printf("  - Tools:     http://localhost:%d/mcp/v1/tools/list", serverPort)
 log.Printf("  - Prompts:   http://localhost:%d/mcp/v1/prompts/list", serverPort)
 log.Printf("  - Resources: http://localhost:%d/mcp/v1/resources/list", serverPort)
 log.Printf("\nServer ready. Connect your AI assistant now.")
 
 if err := server.Start(); err != nil {
  return fmt.Errorf("failed to start server: %w", err)
 }

 return nil
}
```

**File**: `internal/mcp/server.go` (modifications)

```go
// Add method to NewServer initialization
func NewServer(port int) *Server {
 handler := NewHandler()

 return &Server{
  port:    port,
  handler: handler,
 }
}

// Add registration methods to Server
func (s *Server) RegisterCommandTools() {
 s.handler.RegisterCommandTools()
}

func (s *Server) RegisterCommandPrompts() {
 s.handler.RegisterCommandPrompts()
}

func (s *Server) RegisterTemplateResources() {
 s.handler.RegisterTemplateResources()
}

// Update handleInitialize to reflect actual capabilities
func (s *Server) handleInitialize(w http.ResponseWriter, r *http.Request) {
 if r.Method != http.MethodPost {
  http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
  return
 }

 toolCount := len(s.handler.tools)
 promptCount := len(s.handler.prompts)
 resourceCount := len(s.handler.resources)

 response := map[string]interface{}{
  "protocolVersion": "2024-11-05",
  "serverInfo": map[string]string{
   "name":    "technocrat",
   "version": "2.0.0", // Bump version for new architecture
  },
  "capabilities": map[string]interface{}{
   "tools": map[string]interface{}{
    "count": toolCount,
   },
   "resources": map[string]interface{}{
    "count": resourceCount,
   },
   "prompts": map[string]interface{}{
    "count": promptCount,
   },
  },
 }

 s.respondJSON(w, http.StatusOK, response)
}
```

---

## Phase 7: Documentation Updates

### Goal

Update all documentation to reflect the new MCP-first architecture.

### Files to Update

#### 7.1 README.md

```markdown
# Technocrat

Spec-Driven Development Toolkit with MCP Integration

## Quick Start

### 1. Install Technocrat

```bash
# Download latest release
curl -L https://github.com/x86ed/technocrat/releases/latest/download/technocrat-$(uname -s)-$(uname -m) -o technocrat
chmod +x technocrat
sudo mv technocrat /usr/local/bin/
```

### 2. Initialize Your Project

```bash
technocrat init my-project
# Detects: VS Code, Claude Desktop, Cursor
# Select: VS Code
# ✓ MCP server configured for VS Code
```

### 3. Open in Your Editor

```bash
cd my-project
code .  # Opens VS Code
# MCP server starts automatically
```

### 4. Use Workflow Prompts

In VS Code (or your editor):
- Open Copilot Chat or Command Palette
- Invoke `tchncrt.spec` prompt
- Describe your feature
- Watch the AI follow the workflow!

The prompts guide the AI through:
1. Creating specifications
2. Generating implementation plans
3. Breaking down tasks
4. Executing implementation

## Architecture

Technocrat uses the Model Context Protocol (MCP) to provide:

1. **Tools**: Execute Technocrat commands programmatically
2. **Prompts**: Workflow guidance with embedded instructions (from templates/commands/)
3. **Resources**: Access to templates and documentation

The MCP server is **embedded in the binary** and auto-configured during `init`.

### How It Works

```
Your Editor (VS Code, Claude, Cursor, etc.)
    │
    │ Reads MCP config (.vscode/settings.json, etc.)
    │ Auto-starts: technocrat server --stdio
    │
    ↓
Technocrat MCP Server
    │
    ├── Prompts (embedded workflows)
    │   • tchncrt.spec → guides feature specification
    │   • tchncrt.plan → guides implementation planning
    │   • tchncrt.tasks → generates task breakdown
    │   • tchncrt.implement → executes implementation
    │
    ├── Tools (command execution)
    │   • create_spec(description) → creates feature branch + spec
    │   • create_plan() → generates implementation plan
    │   • create_tasks() → breaks plan into tasks
    │   • run_implementation() → validates and executes
    │
    └── Resources (embedded templates)
        • template://spec-template.md
        • template://plan-template.md
        • template://tasks-template.md
```

### Editor Support

| Editor | Auto-Start | Config File | Status |
|--------|-----------|-------------|--------|
| **VS Code** | ✅ Yes (stdio) | `.vscode/settings.json` | Full Support |
| **Claude Desktop** | ✅ Yes (stdio) | `~/Library/.../claude_desktop_config.json` | Full Support |
| **Cursor** | ✅ Yes (stdio) | `~/.cursor/mcp_servers.json` | Full Support |
| **Windsurf** | ✅ Yes (stdio) | `.windsurf/mcp_config.json` | Full Support |
| **Amazon Q CLI** | ⚠️ Manual (HTTP) | `~/.aws/q/mcp-config.json` | Full Support |

All editors use the **same MCP server**, just different transport modes:
- **stdio**: Server runs as child process (most editors)
- **HTTP**: Server runs independently on port 8080 (Amazon Q)

## MCP Server Features

### Available Tools

| Tool | Description |
|------|-------------|
| `create_spec` | Create feature specification from description |
| `create_plan` | Generate implementation plan |
| `create_tasks` | Generate actionable task list |
| `run_implementation` | Execute implementation workflow |
| `create_constitution` | Establish project principles |
| `create_checklist` | Generate quality checklists |
| `clarify_requirements` | Ask structured questions |
| `analyze_artifacts` | Cross-artifact consistency check |

### Available Prompts

| Prompt | Description |
|--------|-------------|
| `tchncrt.spec` | Feature specification workflow |
| `tchncrt.plan` | Implementation planning workflow |
| `tchncrt.tasks` | Task generation workflow |
| `tchncrt.implement` | Implementation execution workflow |
| `tchncrt.constitution` | Constitution creation workflow |
| `tchncrt.checklist` | Checklist generation workflow |
| `tchncrt.clarify` | Requirements clarification workflow |
| `tchncrt.analyze` | Artifact analysis workflow |

### Available Resources

All templates are embedded in the binary and available via MCP:

- `template://spec-template.md`
- `template://plan-template.md`
- `template://tasks-template.md`
- `template://commands/spec.md`
- `template://commands/plan.md`
- And more...

## Supported AI Assistants

Any MCP-compatible editor/client can use Technocrat. Setup is automatic during `init`:

- ✅ **VS Code Copilot** - Full support, auto-configured
- ✅ **Claude Desktop** - Full support, auto-configured
- ✅ **Cursor** - Full support, auto-configured
- ✅ **Windsurf** - Full support, auto-configured
- ✅ **Amazon Q Developer CLI** - Full support, manual start
- ✅ **Custom MCP clients** - Full support via HTTP

**Note**: Technocrat `init` detects which editors you have installed and configures them automatically. No manual configuration needed!

## Documentation

- [Installation Guide](docs/installation.md)
- [Quick Start](docs/quickstart.md)
## Migration from v1.x

If you're upgrading from Technocrat v1.x (slash command based):

### What Changed

| v1.x | v2.0 |
|------|------|
| Slash commands in files | MCP prompts (embedded) |
| Agent-specific command files | Universal MCP server |
| Manual template download | Templates embedded in binary |
| Per-agent setup | Auto-configured during init |
| Static command files | Dynamic tool execution |

### Migration Steps

1. **Install v2.0**: Download new binary
2. **Run init in project**: `technocrat init --here --force`
3. **Select your editor**: Init will detect and configure
4. **Delete old command files**: Remove `.claude/`, `.gemini/`, etc.
5. **Use new prompts**: Same names, but via MCP (tchncrt.spec, etc.)

### Prompt Mapping

The prompt names are the same, just invoked differently:

| v1.x Command | v2.0 Prompt |
|--------------|-------------|
| `/tchncrt.spec` | `tchncrt.spec` |
| `/tchncrt.plan` | `tchncrt.plan` |
| `/tchncrt.tasks` | `tchncrt.tasks` |
| `/tchncrt.implement` | `tchncrt.implement` |

**Benefit**: Prompts now work in **all** supported editors, not just one!

See [Migration Guide](docs/migration-v1-to-v2.md) for detailed instructions.
1. The slash commands are now MCP prompts
2. No need for agent-specific command files
3. Start the MCP server instead of using static commands
4. See [Migration Guide](docs/migration-v1-to-v2.md)

## License

MIT License - see [LICENSE](LICENSE) for details.

```

#### 7.2 docs/mcp-server.md (NEW)

Create comprehensive MCP server documentation covering:
- Server setup and configuration
- Connecting different AI assistants
- Available tools, prompts, and resources
- Example workflows
- Troubleshooting

#### 7.3 docs/migration-v1-to-v2.md (NEW)

Create migration guide covering:
- What changed between v1 and v2
- Mapping of slash commands to MCP prompts
- How to migrate existing projects
- Benefits of the new architecture

#### 7.4 Update Existing Docs

- `docs/installation.md`: Update for new server-based workflow
- `docs/quickstart.md`: Focus on MCP server startup
- `docs/commands-reference.md`: Document MCP tools and prompts
- `docs/agent-integration.md`: Archive or remove (no longer needed)

---

## Phase 8: Build and Release Updates

### Goal
Update build scripts and release process for the new architecture.

### Changes

#### 8.1 Remove Release Package Scripts

**File**: `.github/workflows/scripts/create-release-packages.sh`

This entire script can be removed or simplified significantly since we're no longer creating agent-specific template packages.

#### 8.2 Update Release Workflow

**File**: `.github/workflows/scripts/create-github-release.sh`

```bash
#!/bin/bash
set -e

VERSION=$1

if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version>"
  exit 1
fi

# Build binaries for all platforms
platforms=(
  "darwin/amd64"
  "darwin/arm64"
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
)

mkdir -p .genreleases

for platform in "${platforms[@]}"; do
  os="${platform%/*}"
  arch="${platform#*/}"
  
  output="technocrat"
  if [ "$os" = "windows" ]; then
    output="technocrat.exe"
  fi
  
  echo "Building $os/$arch..."
  GOOS=$os GOARCH=$arch go build -o ".genreleases/$output" cmd/technocrat/main.go
  
  # Create archive
  if [ "$os" = "windows" ]; then
    zip ".genreleases/technocrat-$os-$arch-$VERSION.zip" ".genreleases/$output"
  else
    tar czf ".genreleases/technocrat-$os-$arch-$VERSION.tar.gz" -C .genreleases "$output"
  fi
  
  rm ".genreleases/$output"
done

# Create GitHub release
gh release create "$VERSION" \
  --title "Technocrat $VERSION" \
  --notes "See CHANGELOG.md for details" \
  .genreleases/*
```

#### 8.3 Update Build Script

**File**: `build.go`

Ensure embedded files are included:

```go
//go:build ignore

package main

import (
 "fmt"
 "os"
 "os/exec"
 "path/filepath"
 "runtime"
)

func main() {
 // Ensure templates directory exists (for embedding)
 if _, err := os.Stat("templates"); os.IsNotExist(err) {
  fmt.Println("Error: templates directory not found")
  fmt.Println("Templates are required for embedding into the binary")
  os.Exit(1)
 }
 
 // Build with embed support
 cmd := exec.Command("go", "build", "-o", "bin/technocrat", "./cmd/technocrat")
 cmd.Stdout = os.Stdout
 cmd.Stderr = os.Stderr
 
 if err := cmd.Run(); err != nil {
  fmt.Printf("Build failed: %v\n", err)
  os.Exit(1)
 }
 
 fmt.Println("Build successful!")
 fmt.Printf("Binary: %s\n", filepath.Join("bin", "technocrat"))
 
 // Verify embedded files
 fmt.Println("\nVerifying embedded templates...")
 verifyCmd := exec.Command("./bin/technocrat", "server", "--help")
 if err := verifyCmd.Run(); err != nil {
  fmt.Printf("Warning: Binary verification failed: %v\n", err)
 }
}
```

---

## Phase 9: Testing Strategy (Pragmatic Approach)

### Goal

Test the critical paths and core functionality while being pragmatic about test coverage.

### Testing Philosophy

**Focus on:**
- 🎯 Core functionality that must never break
- 🎯 Complex logic with edge cases
- 🎯 Code that will change frequently
- 🎯 Public APIs and interfaces

**Light testing or manual validation for:**
- ⚠️ Editor-specific integration (too brittle, hard to automate)
- ⚠️ UI/interaction flows (manual testing more practical)
- ⚠️ One-off setup code that rarely changes

### Test Categories

#### 9.1 Unit Tests (Critical - Must Have)

**Embedded Templates:**
- ✅ All expected templates exist and are readable
- ✅ GetTemplate() returns valid content
- ✅ GetCommandTemplate() works for all commands
- ✅ Error handling for missing templates

**MCP Tools:**
- ✅ Tool registration succeeds
- ✅ create_spec tool with valid input returns expected structure
- ✅ create_plan tool validates prerequisites
- ✅ Error handling for invalid arguments
- ✅ JSON schema validation

**Editor Detection:**
- ✅ detectVSCode() with mock filesystem
- ✅ detectClaude() finds config directory
- ✅ DetectEditors() returns non-empty list in dev environment
- ❌ Skip: Testing all OS variations (trust the logic, manual test)

**MCP Config Generation:**
- ✅ installVSCodeMCP() generates valid JSON
- ✅ Config includes correct command path
- ✅ Config merge doesn't break existing settings
- ❌ Skip: Testing actual file writes (use temp dirs, but don't test every editor)

#### 9.2 Integration Tests (Important - Should Have)

**Server Startup:**
- ✅ HTTP server starts and responds to /health
- ✅ stdio server accepts JSON-RPC messages
- ✅ Server registers all tools/prompts/resources
- ❌ Skip: Full MCP protocol compliance testing (trust the implementation)

**Tool Execution:**
- ✅ HTTP POST to /mcp/v1/tools/call with create_spec
- ✅ Tool returns valid result structure
- ⚠️ Manual: Test one full workflow (spec → plan → tasks)

#### 9.3 E2E Tests (Nice to Have - Manual OK)

**Project Initialization:**
- ⚠️ Manual: Run `technocrat init test-project` on macOS
- ⚠️ Manual: Verify .vscode/settings.json created
- ⚠️ Manual: Verify git repo initialized
- ❌ Skip: Automated E2E (requires too much setup)

**Editor Integration:**
- ⚠️ Manual: Open project in VS Code
- ⚠️ Manual: Verify MCP server starts
- ⚠️ Manual: Invoke tchncrt.spec prompt
- ⚠️ Manual: Verify tool executes
- ❌ Skip: Testing all editors (pick 1-2 for validation)

**Full Workflow:**
- ⚠️ Manual: Complete workflow from constitution → implement
- ⚠️ Manual: Verify all artifacts generated correctly
- ❌ Skip: Automated full workflow testing

### Minimum Viable Test Suite

To ship v2.0, we need:

```
✅ internal/templates/embedded_test.go      (10 tests)
✅ internal/templates/accessor_test.go       (5 tests)
✅ internal/mcp/tools_test.go                (15 tests - focus on create_spec, create_plan)
✅ internal/installer/editor_detect_test.go  (8 tests - mock filesystem)
✅ internal/installer/mcp_config_test.go     (10 tests - JSON validation)
✅ internal/mcp/server_test.go               (5 tests - basic startup)
⚠️ Manual testing checklist                  (documented below)
```

**Total: ~53 automated tests + manual validation**

### Manual Testing Checklist

Before release, manually validate:

```
□ macOS: technocrat init → select VS Code → open project → use prompt
□ Linux: technocrat init → select Claude → verify config written
□ Windows: technocrat init → verify .vscode/settings.json
□ HTTP mode: technocrat server → curl /health
□ stdio mode: echo '{"method":"initialize"}' | technocrat server --stdio
□ Full workflow: constitution → spec → plan → tasks → implement
```

### Test Files Structure

```
internal/
├── templates/
│   ├── embedded_test.go          ✅ Critical
│   └── accessor_test.go           ✅ Critical
├── mcp/
│   ├── tools_test.go              ✅ Critical
│   ├── server_test.go             ✅ Important
│   └── prompts_test.go            ⚠️ Nice to have
├── installer/
│   ├── editor_detect_test.go      ✅ Critical
│   └── mcp_config_test.go         ✅ Critical
└── cmd/
    └── init_test.go               ⚠️ Nice to have (complex to mock)
```

### When to Skip Tests

**Skip automated tests when:**
- Requires real editor installation
- Requires actual network calls
- Requires real filesystem writes (except temp dirs)
- Test is more brittle than the code it tests
- Manual testing is faster and more reliable

**Use manual tests when:**
- Testing user interaction flows
- Testing cross-platform behavior
- Testing editor-specific behavior
- Validating full workflows

---

## Phase 10: Rollout and Communication

### Goal

Successfully transition users to the new architecture.

### Communication Plan

#### 10.1 Version 2.0.0 Release

- **Breaking Changes**: Document clearly
- **Migration Guide**: Provide step-by-step instructions
- **Changelog**: Comprehensive list of changes
- **Announcement**: Blog post or forum announcement

#### 10.2 Documentation

- Update all docs before release
- Create video tutorial for new workflow
- Provide example configurations for different clients
- FAQ section for common questions

#### 10.3 Support

- Monitor issues closely after release
- Provide migration support
- Collect feedback for improvements
- Consider deprecation period for v1.x

---

## Implementation Checklist

### Phase 1: Embed Templates (Foundation)
- [ ] Create `internal/templates/embedded.go` with `//go:embed`
- [ ] Create `internal/templates/accessor.go` with read functions
- [ ] Write unit tests for embedded filesystem
- [ ] Update build process to include templates
- [ ] Verify templates in compiled binary

### Phase 2: MCP Command Tools
- [ ] Create `internal/mcp/tools.go`
- [ ] Refactor `internal/cmd/create_feature.go` for library use
- [ ] Refactor `internal/cmd/setup_plan.go` for library use
- [ ] Refactor `internal/cmd/check_prerequisites.go` for library use
- [ ] Implement all 8 command tools
- [ ] Write tool tests with mocks
- [ ] Document tool schemas

### Phase 3: MCP Prompts from Templates
- [ ] Create `internal/mcp/prompts.go`
- [ ] Implement template parser (extract workflow sections)
- [ ] Create prompt-to-tool mapping
- [ ] Register all workflow prompts
- [ ] Write prompt generation tests
- [ ] Verify prompt-to-tool invocation flow

### Phase 4: Editor Detection & MCP Config
- [ ] Create `internal/installer/editor_detect.go`
- [ ] Implement editor detection for VS Code, Claude, Cursor, Amazon Q, Windsurf
- [ ] Create `internal/installer/mcp_config.go`
- [ ] Implement config generators for each editor
- [ ] Write installer tests with mock filesystem
- [ ] Verify generated configs are valid JSON

### Phase 5: stdio Transport Support
- [ ] Add `StartStdio()` to `internal/mcp/server.go`
- [ ] Implement JSON-RPC over stdio
- [ ] Add stdio-specific request handlers
- [ ] Add `--stdio` flag to server command
- [ ] Test stdio communication
- [ ] Verify compatibility with editor launch

### Phase 6: Refactor init.go
- [ ] Add editor detection to init flow
- [ ] Add editor selection prompt
- [ ] Add MCP config installation step
- [ ] Remove `agentConfigs` map
- [ ] Remove template download functions
- [ ] Remove script generation logic
- [ ] Update success messages with editor-specific tips
- [ ] Test init with all supported editors

### Phase 7: Template Resources
- [ ] Create `internal/mcp/resources.go`
- [ ] Register all templates as resources
- [ ] Implement resource URI scheme (template://)
- [ ] Update resource reading to serve embedded files
- [ ] Write resource tests
- [ ] Verify client can access templates

### Phase 8: Update server.go
- [ ] Add tool registration to server startup
- [ ] Add prompt registration to server startup
- [ ] Add resource registration to server startup
- [ ] Update initialization response with capability counts
- [ ] Add verbose logging flag
- [ ] Test both HTTP and stdio modes
- [ ] Verify all endpoints work correctly

### Phase 9: Documentation Updates
- [ ] Update README.md with new workflow
- [ ] Create `docs/mcp-server.md` (comprehensive MCP guide)
- [ ] Create `docs/migration-v1-to-v2.md`
- [ ] Update `docs/installation.md`
- [ ] Update `docs/quickstart.md`
- [ ] Update `docs/commands-reference.md` (tools & prompts)
- [ ] Archive or remove `docs/agent-integration.md`
- [ ] Add editor-specific screenshots/examples

### Phase 10: Build & Release Updates
- [ ] Remove `.github/workflows/scripts/create-release-packages.sh`
- [ ] Simplify `.github/workflows/scripts/create-github-release.sh`
- [ ] Update `build.go` to verify embedded files
- [ ] Test multi-platform builds
- [ ] Verify binary size is reasonable
- [ ] Create release workflow for v2.0.0

### Phase 11: Testing (Prioritized)

**Critical Path (Must Have):**
- [ ] Unit tests for embedded templates (verify all files accessible)
- [ ] Unit tests for MCP tools (core create_spec, create_plan, create_tasks)
- [ ] Unit tests for editor detection (mock filesystem)
- [ ] Unit tests for MCP config generation (verify JSON validity)
- [ ] Integration test: server startup in both modes (HTTP + stdio)
- [ ] Manual test: init → open VS Code/Claude → invoke one prompt

**Nice to Have:**
- [ ] Unit tests for MCP prompts (template parsing)
- [ ] E2E automated test (requires headless editor)
- [ ] Performance benchmarks (server startup, tool execution)
- [ ] Test with all supported editors

**Skip/Defer:**
- [ ] ❌ Comprehensive E2E automation (too brittle, manual testing sufficient)
- [ ] ❌ Full editor integration tests (requires installed editors)
- [ ] ❌ Exhaustive tool testing (focus on happy path + error cases)

### Phase 12: Code Cleanup
- [ ] Remove `agentConfigs` from codebase
- [ ] Remove old template download logic
- [ ] Remove script generation code
- [ ] Remove agent-specific update scripts
- [ ] Clean up or archive AGENTS.md
- [ ] Remove unused imports
- [ ] Run linters and fix issues
- [ ] Update version strings to 2.0.0

### Phase 13: Pre-Release Validation
- [ ] Test init on macOS, Linux, Windows
- [ ] Verify MCP config for all supported editors
- [ ] Test server startup in both modes
- [ ] Verify all prompts work correctly
- [ ] Verify all tools execute successfully
- [ ] Check binary sizes across platforms
- [ ] Review all documentation for accuracy
- [ ] Create changelog for v2.0.0

### Phase 14: Release & Communication
- [ ] Tag v2.0.0 release
- [ ] Create GitHub release with binaries
- [ ] Write release announcement
- [ ] Update website/docs
- [ ] Post to community forums
- [ ] Monitor feedback and issues
- [ ] Prepare hotfix plan if needed

### Phase 15: Post-Release Support
- [ ] Monitor GitHub issues
- [ ] Provide migration support
- [ ] Collect user feedback
- [ ] Plan v2.1 improvements
- [ ] Consider v1.x deprecation timeline
- [ ] Update roadmap
- [ ] Create release workflow

### Phase 9: Testing

- [ ] Write all unit tests
- [ ] Write integration tests
- [ ] Write E2E tests
- [ ] Test with real MCP clients
- [ ] Performance testing

### Phase 10: Rollout

- [ ] Create changelog
- [ ] Write migration guide
- [ ] Prepare announcement
- [ ] Tag v2.0.0 release
- [ ] Monitor feedback

---

## Success Metrics

### Technical Metrics

- ✅ All templates embedded in binary
- ✅ Binary size remains reasonable (<50MB)
- ✅ Server starts in <2 seconds
- ✅ Tool execution time <1 second
- ✅ 100% test coverage for new modules

### User Metrics

- ✅ Simplified setup (3 steps instead of 5)
- ✅ Works with all MCP clients
- ✅ No agent-specific configuration needed
- ✅ Clear migration path from v1.x
- ✅ Positive user feedback

### Maintenance Metrics

- ✅ 50% reduction in codebase complexity
- ✅ No more agent-specific code
- ✅ Easier to add new features
- ✅ Simpler release process
- ✅ Better test coverage

---

## Risk Mitigation

### Risk 1: Breaking Changes

**Mitigation**:

- Comprehensive migration guide
- Consider parallel v1.x maintenance for 6 months
- Clear communication about changes

### Risk 2: MCP Client Compatibility

**Mitigation**:

- Test with multiple MCP clients
- Strict MCP protocol compliance
- Fallback documentation for manual workflows

### Risk 3: Performance Regression

**Mitigation**:

- Benchmark before/after
- Profile embedded filesystem access
- Optimize if needed

### Risk 4: User Confusion

**Mitigation**:

- Clear documentation
- Video tutorials
- FAQ section
- Active support during transition

---

## Conclusion

This refactoring transforms Technocrat from an agent-specific tool into a universal MCP server that works with any compatible AI assistant. The key benefits:

1. **Simplicity**: No agent-specific setup
2. **Portability**: Works anywhere MCP is supported
3. **Maintainability**: Single codebase for all clients
4. **Extensibility**: Easy to add new tools and prompts
5. **Performance**: Embedded templates, no downloads

The implementation is broken into manageable phases with clear success criteria and testing at each stage.
