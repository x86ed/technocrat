# Technocrat

A Spec-Driven Development Framework with MCP Server

Technocrat is a Model Context Protocol (MCP) server implementation that provides tools, resources, and prompts for spec-driven development workflows. It combines a powerful CLI for managing feature development with a fully-compliant MCP server for AI agent integration.

## Features

### Spec-Driven Development
- **Project Initialization**: Bootstrap new projects with templates and agent configurations
- **Feature Management**: Create and manage feature branches with structured specifications
- **Multi-Agent Support**: Works with 13+ AI agents (Claude, Copilot, Gemini, Cursor, Windsurf, and more)
- **Template System**: Pre-built templates for specs, plans, tasks, and agent context files
- **Agent Context Updates**: Automatically sync project information to AI agent configuration files

### MCP Server
- **Full MCP Protocol**: Complete implementation of Model Context Protocol with HTTP endpoints
- **Tools API**: Extensible tool system for executing operations
- **Resources API**: Access to server resources and information
- **Prompts API**: Pre-built prompts for common development tasks

### Developer Experience
- **Pure Go**: Written entirely in Go, using Cobra for CLI - single binary, no dependencies
- **Cross-platform**: Works on Linux, macOS, and Windows
- **Modern Build**: Uses `go generate` for streamlined builds
- **Comprehensive Testing**: Unit and integration tests with >95% coverage

## Quick Start

### Building from Source

```bash
# Clone the repository
git clone https://github.com/x86ed/technocrat.git
cd technocrat

# Build using go generate (recommended)
go generate ./cmd/technocrat

# Or use the build script
go run build.go -build

# Binary will be in ./bin/technocrat
./bin/technocrat --help
```

### Initialize a New Project

```bash
# Initialize a new project with AI agent support
./bin/technocrat init my-project --ai claude

# Or initialize in current directory
./bin/technocrat init . --ai copilot

# Supported agents: claude, copilot, gemini, cursor-agent, qwen, opencode, 
#                   codex, windsurf, kilocode, auggie, codebuddy, roo, q
```

### Create and Manage Features

```bash
# Create a new feature branch and spec directory
./bin/technocrat create-feature "add user authentication"

# Set up implementation plan
./bin/technocrat setup-plan

# Update AI agent context files with feature information
./bin/technocrat update-agent-context
```

### Optional: Run the MCP Server

```bash
# Start the MCP server (default port 8080)
./bin/technocrat server

# Or specify a custom port
./bin/technocrat server --port 9090
```

## Installation

### Option 1: Install from Source

```bash
# Clone and build
git clone https://github.com/x86ed/technocrat.git
cd technocrat
go run build.go -build

# Install to system (optional)
go install ./cmd/technocrat
```

### Option 2: Using Go Install

```bash
# Install directly from GitHub
go install github.com/x86ed/technocrat/cmd/technocrat@latest
```

### Option 3: Using Go Generate

```bash
# Clone repository
git clone https://github.com/x86ed/technocrat.git
cd technocrat

# Generate and build
go generate ./cmd/technocrat

# Binary will be in ./bin/technocrat
```

## Spec-Driven Development Workflow

Technocrat provides a complete workflow for managing features using Spec-Driven Development principles:

### 1. Initialize Your Project

```bash
# Create a new project with AI agent support
technocrat init my-project --ai claude

# This creates:
# - .tchncrt/ directory with templates
# - specs/ directory for feature specifications  
# - Agent-specific configuration (e.g., CLAUDE.md)
# - memory/ directory for project constitution
```

### 2. Check Prerequisites

```bash
# Verify required tools are installed
technocrat check

# This checks for:
# - Git
# - AI assistant CLIs (claude, gemini, qwen, etc.)
# - Code editors (VS Code, Cursor, Windsurf)
```

### 3. Create a New Feature

```bash
# Create a feature branch and spec directory
technocrat create-feature "add user authentication"

# Creates:
# - specs/001-add-user-authentication/ directory
# - Git branch: 001-add-user-authentication (if git available)
# - Copies spec template to new directory
# - Sets TCHNCRT_FEATURE environment variable

# Get feature info in JSON format
technocrat create-feature "add dashboard" --json
```

### 4. Set Up Implementation Plan

```bash
# Create plan.md from template
technocrat setup-plan

# Creates:
# - plan.md in current feature directory
# - Copies from .tchncrt/templates/plan-template.md

# Get paths in JSON format
technocrat setup-plan --json
```

### 5. Update Agent Context

```bash
# Update all existing agent files with feature information
technocrat update-agent-context

# Or update a specific agent
technocrat update-agent-context claude
technocrat update-agent-context copilot

# Supported agents:
# claude, gemini, copilot, cursor, qwen, opencode, codex,
# windsurf, kilocode, auggie, roo, codebuddy, q
```

This workflow ensures your AI coding assistant has up-to-date context about your project structure, technology stack, and current feature development.

## Command-Line Interface

### Available Commands

```bash
technocrat --help              # Show all commands
technocrat version             # Show version information
```

#### Spec-Driven Development Commands

```bash
# Initialize a new project
technocrat init <project-name> [flags]
  --ai string           AI assistant (claude, copilot, gemini, cursor-agent, etc.)
  --ignore-agent-tools  Skip agent tool availability checks
  --no-templates        Skip template installation

# Check prerequisites
technocrat check              # Verify required tools are installed

# Create a new feature
technocrat create-feature <description> [flags]
  --json                Output in JSON format

# Set up implementation plan
technocrat setup-plan [flags]
  --json                Output in JSON format

# Update agent context files
technocrat update-agent-context [agent-type]
  # agent-type: claude, gemini, copilot, cursor, qwen, opencode, codex,
  #             windsurf, kilocode, auggie, roo, codebuddy, q
  # If no agent specified, updates all existing agent files
```

#### MCP Server Commands

```bash
# Start the MCP server
technocrat server [flags]
  -p, --port int        Port to listen on (default 8080)
```

### Usage Examples

```bash
# Initialize a project with Claude support
technocrat init my-app --ai claude

# Create a feature for adding authentication
technocrat create-feature "add user authentication system"

# Set up the implementation plan
technocrat setup-plan

# Update Claude's context file with feature info
technocrat update-agent-context claude

# Start MCP server on custom port
technocrat server --port 9000

# Check all tool installations
technocrat check
```

## MCP Protocol Endpoints

The server implements the following MCP endpoints:

- `POST /mcp/v1/initialize` - Initialize MCP connection
- `GET /mcp/v1/tools/list` - List available tools
- `POST /mcp/v1/tools/call` - Execute a tool
- `GET /mcp/v1/resources/list` - List available resources
- `POST /mcp/v1/resources/read` - Read a resource
- `GET /mcp/v1/prompts/list` - List available prompts
- `POST /mcp/v1/prompts/get` - Get a prompt
- `GET /health` - Health check endpoint

## Available Tools

### echo
Echoes back the input message.

**Parameters:**
- `message` (string, required): The message to echo

**Example:**
```bash
curl -X POST http://localhost:8080/mcp/v1/tools/call \
  -H "Content-Type: application/json" \
  -d '{"name": "echo", "arguments": {"message": "Hello, World!"}}'
```

### system_info
Returns basic system information about the server.

**Example:**
```bash
curl -X POST http://localhost:8080/mcp/v1/tools/call \
  -H "Content-Type: application/json" \
  -d '{"name": "system_info", "arguments": {}}'
```

## Project Structure

```
technocrat/
├── cmd/
│   └── technocrat/          # Main CLI entry point
├── internal/
│   ├── cmd/                 # Cobra command implementations
│   │   ├── root.go          # Root command
│   │   ├── check.go         # Tool checking
│   │   ├── init.go          # Project initialization
│   │   ├── create_feature.go # Feature creation
│   │   ├── setup_plan.go    # Plan setup
│   │   ├── update_agent_context.go # Agent context updates
│   │   ├── server.go        # MCP server command
│   │   ├── version.go       # Version command
│   │   └── common.go        # Shared utilities
│   ├── mcp/                 # MCP protocol implementation
│   │   ├── server.go        # HTTP server and endpoints
│   │   └── handler.go       # Tools, resources, and prompts
│   ├── ui/                  # UI components (panels, trackers)
│   ├── installer/           # Installation logic
│   └── tchncrt/             # Core utilities
├── templates/               # Project templates
│   ├── agent-file-template.md
│   ├── spec-template.md
│   ├── plan-template.md
│   ├── tasks-template.md
│   ├── checklist-template.md
│   └── commands/            # Agent command templates
│       ├── specify.md
│       ├── plan.md
│       ├── tasks.md
│       └── ...
├── memory/
│   └── constitution.md      # Project constitution template
├── build.go                 # Build script
├── go.mod                   # Go module definition
└── README.md                # This file
```

## Development

### Prerequisites

- Go 1.24 or later
- Git

### Building from Source

```bash
# Clone repository
git clone https://github.com/x86ed/technocrat.git
cd technocrat

# Download dependencies
go run build.go -deps

# Run tests
go run build.go -test
# Or: go test ./...

# Format code
go run build.go -fmt

# Run linting
go run build.go -vet

# Build binary
go run build.go -build

# Or use go generate
go generate ./cmd/technocrat
```

### Build Script Options

The `build.go` script supports the following options:

```bash
go run build.go [options]

Options:
  -build              Build the binary
  -clean              Clean build artifacts
  -test               Run tests
  -fmt                Format code
  -vet                Run go vet
  -deps               Download and tidy dependencies
  -all                Build everything
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run tests for specific package
go test ./internal/cmd
go test ./internal/mcp

# Run specific test
go test -run TestCreateFeature ./internal/cmd
```

### Development Workflow

```bash
# Quick testing without installing
go run cmd/technocrat/main.go --help
go run cmd/technocrat/main.go init test-project --ai claude

# Build and test locally
go generate ./cmd/technocrat
./bin/technocrat --help

# Install to GOPATH for system-wide use
go install ./cmd/technocrat
```

### Adding New Commands

To add a new Cobra command, create a file in `internal/cmd/`:

```go
// internal/cmd/mycommand.go
package cmd

import (
    "github.com/spf13/cobra"
)

var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Description of my command",
    Long:  `Detailed description of what mycommand does.`,
    RunE:  runMyCommand,
}

func init() {
    rootCmd.AddCommand(myCmd)
    // Add flags here
}

func runMyCommand(cmd *cobra.Command, args []string) error {
    // Implementation
    return nil
}
```

### Adding New MCP Tools

To add a new MCP tool, edit `internal/mcp/handler.go` and register it in the `registerDefaultTools` function:

```go
h.tools["my_tool"] = Tool{
    Name:        "my_tool",
    Description: "Description of my tool",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "param1": map[string]interface{}{
                "type":        "string",
                "description": "First parameter",
            },
        },
        "required": []string{"param1"},
    },
    Handler: func(args map[string]interface{}) (interface{}, error) {
        // Tool implementation
        return result, nil
    },
}
```

## Configuration

Technocrat uses a simple file-based configuration system. Configuration files can be stored in:

- **Linux**: `/etc/technocrat/config.json` or `~/.config/technocrat/config.json`
- **macOS**: `~/.config/technocrat/config.json`
- **Windows**: `%APPDATA%\technocrat\config.json`

Example configuration:

```json
{
  "port": 8080,
  "log_level": "info",
  "templates_dir": ".tchncrt/templates"
}
```

Most configuration is managed through the CLI commands and doesn't require manual editing.

## Supported AI Agents

Technocrat supports integration with 13+ AI coding assistants:

| Agent | Key | Folder | CLI Required |
|-------|-----|--------|--------------|
| **Claude Code** | `claude` | `.claude/` | Yes |
| **GitHub Copilot** | `copilot` | `.github/` | No (IDE-based) |
| **Gemini CLI** | `gemini` | `.gemini/` | Yes |
| **Cursor** | `cursor-agent` | `.cursor/` | No (IDE-based) |
| **Qwen Code** | `qwen` | `.qwen/` | Yes |
| **opencode** | `opencode` | `.opencode/` | Yes |
| **Codex CLI** | `codex` | `.codex/` | Yes |
| **Windsurf** | `windsurf` | `.windsurf/` | No (IDE-based) |
| **Kilo Code** | `kilocode` | `.kilocode/` | No (IDE-based) |
| **Auggie CLI** | `auggie` | `.augment/` | Yes |
| **Roo Code** | `roo` | `.roo/` | No (IDE-based) |
| **CodeBuddy** | `codebuddy` | `.codebuddy/` | Yes |
| **Amazon Q Developer** | `q` | `.amazonq/` | Yes |

Use the `--ai` flag with `technocrat init` to specify your preferred agent.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for:

- How to build and test
- Code standards and conventions
- How to add new commands
- How to add MCP tools/resources
- How to add support for new AI agents (see [AGENTS.md](AGENTS.md))
- Submitting pull requests

## Related Documentation

- **[AGENTS.md](AGENTS.md)** - Guide for adding new AI agent support
- **[docs/](docs/)** - Full documentation with DocFX
- **[templates/](templates/)** - Project and agent templates

## License

See [LICENSE](LICENSE) file for details.

## Support

For issues, questions, and feature requests, please open an issue on [GitHub](https://github.com/x86ed/technocrat/issues).
