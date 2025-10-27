# Technocrat

A Spec Driven Development Framework with MCP Server

Technocrat is a Model Context Protocol (MCP) server implementation that provides tools, resources, and prompts for spec-driven development workflows.

**📚 New to Technocrat? Start with the [Getting Started Guide](GETTING_STARTED.md)!**

## Features

- **MCP Server**: Full implementation of the Model Context Protocol with HTTP endpoints
- **Pure Go**: Written entirely in Go, using Cobra for CLI
- **Go Generate**: Modern build system using `go generate` instead of Make
- **Tools**: Extensible tool system for executing operations
- **Resources**: Access to server resources and information
- **Prompts**: Pre-built prompts for common development tasks
- **Easy Installation**: Built-in installation and management commands
- **Cross-platform**: Works on Linux, macOS, and Windows

## Quick Start

```bash
# Build using go generate (recommended)
go generate ./cmd/technocrat

# Or use the build script
go run build.go -build

# Run the server
./bin/technocrat server

# Install system-wide
sudo go run build.go -install
```

For detailed instructions, see [GETTING_STARTED.md](GETTING_STARTED.md).

## Installation

### Quick Install

```bash
# Clone the repository
git clone https://github.com/x86ed/technocrat.git
cd technocrat

# Build the CLI
make build

# Install to system (requires sudo on Unix-like systems)
sudo make install
```

### Using the CLI Installer

```bash
# Build first
make build

# Install using the built-in installer
sudo ./bin/technocrat install

# Install with systemd service (Linux only)
sudo ./bin/technocrat install --systemd

# Custom installation directory
./bin/technocrat install --dir ~/.local/bin
```

### Using Go Build Script

```bash
# Build the binary
go run build.go -build

# Install to /usr/local/bin
sudo go run build.go -install

# Install to custom directory
go run build.go -install -install-dir=~/.local/bin
```

### Using Go Generate

```bash
# Generate and build
go generate ./...

# Install using go install
go install ./cmd/technocrat
```

### Using Just (optional)

If you have [just](https://github.com/casey/just) installed:

```bash
# Build
just build

# Install
just install

# Run tests
just test
```

## Usage

### Command-Line Interface

```bash
# Show available commands
technocrat --help

# Show version
technocrat version

# Start the MCP server (default port 8080)
technocrat server

# Start with custom port
technocrat server --port 9090

# Check prerequisites for Spec-Driven Development
technocrat check-prerequisites

# Check prerequisites with JSON output
technocrat check-prerequisites --json

# Get feature paths only
technocrat check-prerequisites --paths-only

# Check implementation prerequisites (requires tasks.md)
technocrat check-prerequisites --require-tasks --include-tasks

# Install the binary to system
sudo technocrat install

# Install with systemd service (Linux)
sudo technocrat install --systemd

# Uninstall
sudo technocrat uninstall

# Uninstall including systemd service
sudo technocrat uninstall --systemd
```

### Systemd Service (Linux)

If you installed with systemd support:

```bash
# Start the service
sudo systemctl start technocrat

# Enable on boot
sudo systemctl enable technocrat

# Check status
sudo systemctl status technocrat

# View logs
sudo journalctl -u technocrat -f
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
│   └── technocrat/     # Unified CLI entry point
├── internal/
│   ├── cmd/           # Cobra command definitions
│   │   ├── root.go    # Root command
│   │   ├── server.go  # Server command
│   │   ├── install.go # Install command
│   │   ├── uninstall.go # Uninstall command
│   │   └── version.go # Version command
│   ├── mcp/           # MCP protocol implementation
│   │   ├── server.go  # HTTP server and endpoints
│   │   └── handler.go # Tools, resources, and prompts
│   └── installer/     # Installation logic
│       └── installer.go
├── Makefile           # Build automation
├── go.mod             # Go module definition
└── README.md          # This file
```

## Development

### Prerequisites

- Go 1.24 or later

### Building from Source

```bash
# Download dependencies
go run build.go -deps

# Format code
go run build.go -fmt

# Run go vet
go run build.go -vet

# Run tests
go run build.go -test

# Build binary
go run build.go -build

# Or use go generate
go generate ./...
```

### Build Script Options

The `build.go` script supports the following options:

```bash
go run build.go [options]

Options:
  -build              Build the binary
  -install            Build and install the binary
  -uninstall          Uninstall the binary
  -clean              Clean build artifacts
  -test               Run tests
  -fmt                Format code
  -vet                Run go vet
  -deps               Download and tidy dependencies
  -all                Build everything
  -install-dir string Installation directory (default "/usr/local/bin")
```

### Adding New Tools

To add a new tool, edit `internal/mcp/handler.go` and register it in the `registerDefaultTools` function:

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

Configuration files are stored in:
- **Linux**: `/etc/technocrat/config.json`
- **macOS**: `~/.config/technocrat/config.json`
- **Other**: `~/.technocrat/config.json`

Example configuration:
```json
{
  "port": 8080,
  "log_level": "info"
}
```

## Uninstallation

### Using the CLI

```bash
# Basic uninstall
sudo technocrat uninstall

# Remove systemd service too (Linux)
sudo technocrat uninstall --systemd
```

### Using the build script

```bash
sudo go run build.go -uninstall
```

### Manual Uninstallation

```bash
# Remove binary
sudo rm /usr/local/bin/technocrat

# Remove systemd service (if installed)
sudo systemctl stop technocrat
sudo systemctl disable technocrat
sudo rm /etc/systemd/system/technocrat.service
sudo systemctl daemon-reload

# Remove configuration (optional)
sudo rm -rf /etc/technocrat  # Linux
rm -rf ~/.config/technocrat   # macOS
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

See LICENSE file for details.

## Support

For issues and questions, please open an issue on GitHub
