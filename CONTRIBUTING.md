# Contributing to Technocrat

Thank you for your interest in contributing to Technocrat! This guide will help you understand the project structure and how to make meaningful contributions.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Code Standards](#code-standards)
- [Adding New Features](#adding-new-features)
- [Testing](#testing)
- [Documentation](#documentation)
- [Submitting Changes](#submitting-changes)
- [Adding New AI Agents](#adding-new-ai-agents)

## Getting Started

### Prerequisites

- **Go 1.24+** - [Install Go](https://go.dev/doc/install)
- **Git** - Version control
- **Make** (optional) - Build automation
- A code editor (VS Code, GoLand, etc.)

### Quick Start

```bash
# Clone the repository
git clone https://github.com/x86ed/technocrat.git
cd technocrat

# Build the project
go generate ./cmd/technocrat

# Run tests
go test ./...

# Install locally
go install ./cmd/technocrat
```

See [Local Development Guide](docs/local-development.md) for detailed instructions.

## Development Setup

### Project Structure

```
technocrat/
├── cmd/
│   └── technocrat/          # Main CLI entry point
│       └── main.go
├── internal/
│   ├── cmd/                 # Cobra command implementations
│   │   ├── root.go         # Root command setup
│   │   ├── init.go         # Init command
│   │   ├── create_feature.go
│   │   ├── setup_plan.go
│   │   ├── update_agent_context.go
│   │   ├── check.go
│   │   ├── server.go       # MCP server command
│   │   └── version.go
│   ├── mcp/                # MCP protocol implementation
│   │   ├── server.go       # HTTP server
│   │   └── handler.go      # MCP endpoints (tools, resources, prompts)
│   ├── ui/                 # UI utilities (colors, panels, selectors)
│   ├── installer/          # Installation utilities
│   └── tchncrt/            # Core utilities (paths, etc.)
├── templates/              # Spec/plan/task templates
├── memory/
│   └── constitution.md     # Project principles
├── docs/                   # Documentation (DocFX)
└── build.go               # Build script
```

### Key Technologies

- **[Cobra](https://cobra.dev/)** - CLI framework
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - TUI components
- **Go 1.24+** - Pure Go implementation

## Code Standards

### Go Conventions

Follow standard Go conventions:

- **Format code**: `go fmt ./...`
- **Vet code**: `go vet ./...`
- **Use gofmt**: All code must be formatted with gofmt
- **Error handling**: Always check and handle errors
- **Naming**: Follow [Effective Go](https://go.dev/doc/effective_go) naming conventions

### File Organization

- **Commands**: Place in `internal/cmd/`
- **Business logic**: Place in appropriate `internal/` subdirectories
- **Tests**: Name test files `*_test.go` alongside source files
- **Public packages**: Avoid - use `internal/` to prevent external dependencies

### Documentation

- **Document all exported functions/types**: Use godoc-style comments
- **Keep comments up-to-date**: Comments must match the code
- **Explain "why" not "what"**: Code shows what, comments explain why

Example:

```go
// CreateFeature creates a new feature specification directory and git branch.
// It auto-increments the feature number based on existing specs in the specs/ directory.
// Returns the feature path and any error encountered.
func CreateFeature(description string) (string, error) {
    // Implementation...
}
```

## Adding New Features

### Adding a New CLI Command

Technocrat uses [Cobra](https://cobra.dev/) for CLI commands. To add a new command:

#### 1. Create Command File

Create `internal/cmd/your_command.go`:

```go
package cmd

import (
    "github.com/spf13/cobra"
)

var yourCmd = &cobra.Command{
    Use:   "your-command [args]",
    Short: "Brief description",
    Long: `Detailed description of what your command does.
    
Can span multiple lines and include examples.`,
    RunE: runYourCommand,
}

func init() {
    // Add flags
    yourCmd.Flags().StringP("option", "o", "", "Option description")
    
    // Register with root command
    rootCmd.AddCommand(yourCmd)
}

func runYourCommand(cmd *cobra.Command, args []string) error {
    // Command implementation
    return nil
}
```

#### 2. Add Tests

Create `internal/cmd/your_command_test.go`:

```go
package cmd

import (
    "testing"
)

func TestYourCommand(t *testing.T) {
    // Test implementation
}
```

#### 3. Update Documentation

- Add to [Command Reference](docs/commands-reference.md)
- Update README.md if it's a major command
- Update help text in root command if needed

### Adding MCP Tools/Resources/Prompts

The MCP server is implemented in `internal/mcp/handler.go`.

#### Adding a Tool

Tools are functions AI agents can call.

**1. Add to `ListTools()`:**

```go
{
    Name:        "your_tool",
    Description: "What your tool does",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "param": map[string]interface{}{
                "type":        "string",
                "description": "Parameter description",
            },
        },
        "required": []string{"param"},
    },
}
```

**2. Add to `CallTool()`:**

```go
case "your_tool":
    param, ok := req.Arguments["param"].(string)
    if !ok {
        return CallToolResult{}, fmt.Errorf("param is required")
    }
    
    // Tool logic here
    result := doSomething(param)
    
    return CallToolResult{
        Content: []Content{{Type: "text", Text: result}},
    }, nil
```

**3. Add tests:**

```go
func TestCallTool_YourTool(t *testing.T) {
    handler := NewHandler()
    result, err := handler.CallTool(CallToolRequest{
        Name: "your_tool",
        Arguments: map[string]interface{}{"param": "value"},
    })
    // Assertions...
}
```

#### Adding a Resource

Resources provide read-only access to project data.

**1. Add to `ListResources()`:**

```go
{
    URI:         "yourtype://identifier",
    Name:        "Resource Name",
    Description: "What this resource provides",
    MimeType:    "text/plain",
}
```

**2. Add to `ReadResource()`:**

```go
case "yourtype":
    content := loadYourResource(uri)
    return ReadResourceResult{
        Contents: []ResourceContents{{
            URI:      uri,
            MimeType: "text/plain",
            Text:     content,
        }},
    }, nil
```

#### Adding a Prompt

Prompts are reusable conversation starters.

**1. Add to `ListPrompts()`:**

```go
{
    Name:        "your_prompt",
    Description: "What this prompt helps with",
    Arguments: []PromptArgument{{
        Name:        "arg",
        Description: "Argument description",
        Required:    true,
    }},
}
```

**2. Add to `GetPrompt()`:**

```go
case "your_prompt":
    arg := req.Arguments["arg"]
    text := fmt.Sprintf("Prompt template with %s", arg)
    
    return GetPromptResult{
        Messages: []PromptMessage{{
            Role: "user",
            Content: MessageContent{
                Type: "text",
                Text: text,
            },
        }},
    }, nil
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test ./internal/cmd
go test ./internal/mcp

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Writing Tests

- **Test file naming**: `*_test.go`
- **Test function naming**: `TestFunctionName` or `TestFunctionName_Scenario`
- **Use table-driven tests** for multiple scenarios
- **Test both success and error paths**

Example:

```go
func TestCreateFeature(t *testing.T) {
    tests := []struct {
        name        string
        description string
        wantErr     bool
    }{
        {
            name:        "valid feature",
            description: "add user login",
            wantErr:     false,
        },
        {
            name:        "empty description",
            description: "",
            wantErr:     true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := CreateFeature(tt.description)
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateFeature() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Test Coverage

Aim for:

- **80%+ coverage** for new code
- **100% coverage** for critical paths (init, create-feature, etc.)
- **Error path coverage** - Test all error conditions

## Documentation

### Updating Documentation

When adding features:

1. **Update command help text** in Cobra command definition
2. **Update [Command Reference](docs/commands-reference.md)** with usage examples
3. **Update README.md** if it's a user-facing feature
4. **Add inline godoc comments** for exported functions/types

### Building Documentation Locally

```bash
cd docs
dotnet tool install -g docfx
docfx docfx.json --serve
```

Open `http://localhost:8080` to view.

## Submitting Changes

### Before Creating a Pull Request

1. **Run tests**:

   ```bash
   go test ./...
   ```

2. **Format code**:

   ```bash
   go fmt ./...
   ```

3. **Vet code**:

   ```bash
   go vet ./...
   ```

4. **Update documentation** as needed

5. **Write clear commit messages**:

   ```sh
   feat: add support for feature X
   fix: resolve issue with Y
   docs: update command reference
   test: add tests for Z
   ```

### Pull Request Process

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/your-feature`
3. **Make your changes**
4. **Commit with clear messages**
5. **Push to your fork**: `git push origin feature/your-feature`
6. **Open a Pull Request** on GitHub
7. **Respond to review feedback**

### PR Guidelines

- **One feature per PR** - Keep PRs focused
- **Include tests** - All new code should have tests
- **Update docs** - Document new features
- **Pass CI checks** - All tests must pass
- **Describe changes** - Explain what and why in PR description

## Adding New AI Agents

To add support for a new AI agent, see [AGENTS.md](AGENTS.md) for the complete guide. Summary:

1. **Add to `agentConfigs`** in `internal/cmd/init.go`
2. **Update CLI help text** in init command
3. **Update README.md** with new agent in supported agents table
4. **Update [Agent Integration Guide](docs/agent-integration.md)**
5. **Test init and update-agent-context** with new agent

See [AGENTS.md](AGENTS.md) for step-by-step instructions.

## Getting Help

- **Issues**: Open an [issue on GitHub](https://github.com/x86ed/technocrat/issues)
- **Discussions**: Use [GitHub Discussions](https://github.com/x86ed/technocrat/discussions)
- **Documentation**: Check [docs/](docs/) for guides

## Code of Conduct

Be respectful, constructive, and collaborative. We want this project to be welcoming to all contributors.

## License

By contributing to Technocrat, you agree that your contributions will be licensed under the project's [LICENSE](LICENSE).

---

Thank you for contributing to Technocrat! 🚀
