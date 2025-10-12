# Getting Started with Technocrat

This guide will help you get started with building, testing, and using Technocrat.

## Prerequisites

- Go 1.24 or later
- Git (for retrieving version information)

## Quick Start

### 1. Build the Project

Technocrat uses `go generate` and a custom Go build script instead of Make or shell scripts.

```bash
# Option 1: Using go generate (recommended)
go generate ./cmd/technocrat

# Option 2: Using the build script directly
go run build.go -build

# Option 3: Using go install (installs to $GOPATH/bin)
go install ./cmd/technocrat
```

### 2. Run the Server

```bash
# Run from the build directory
./bin/technocrat server

# Or if installed via go install
technocrat server

# Run with custom port
./bin/technocrat server --port 9090
```

### 3. Test the Server

In another terminal:

```bash
# Health check
curl http://localhost:8080/health

# Initialize MCP connection
curl -X POST http://localhost:8080/mcp/v1/initialize

# List available tools
curl http://localhost:8080/mcp/v1/tools/list

# Call the echo tool
curl -X POST http://localhost:8080/mcp/v1/tools/call \
  -H "Content-Type: application/json" \
  -d '{"name": "echo", "arguments": {"message": "Hello, MCP!"}}'
```

## Build Commands

All build operations can be done with the `build.go` script:

```bash
# Build the binary
go run build.go -build

# Run tests
go run build.go -test

# Format code
go run build.go -fmt

# Run go vet
go run build.go -vet

# Clean build artifacts
go run build.go -clean

# Download and tidy dependencies
go run build.go -deps

# Install to /usr/local/bin
sudo go run build.go -install

# Install to custom directory
go run build.go -install -install-dir=$HOME/.local/bin

# Uninstall
sudo go run build.go -uninstall
```

## Using go generate

The project is configured to use `go generate` for building:

```bash
# Generate and build from the main package
cd cmd/technocrat
go generate

# Or from the project root
go generate ./cmd/technocrat

# Clean and rebuild
go run ../../build.go -clean
go generate
```

## Installation Options

### Option 1: System-wide Installation (Linux/macOS)

```bash
# Build and install
sudo go run build.go -install

# Verify installation
which technocrat
technocrat version
```

### Option 2: User Installation

```bash
# Install to user directory
go run build.go -install -install-dir=$HOME/.local/bin

# Make sure ~/.local/bin is in your PATH
export PATH="$HOME/.local/bin:$PATH"

# Add to your shell profile for persistence
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc  # or ~/.zshrc
```

### Option 3: Using go install

```bash
# Install to $GOPATH/bin (or $GOBIN if set)
go install ./cmd/technocrat

# Make sure $GOPATH/bin is in your PATH
export PATH="$GOPATH/bin:$PATH"
```

### Option 4: Using the CLI installer

Once you have the binary, you can use it to install itself:

```bash
# Build first
go run build.go -build

# Then install
sudo ./bin/technocrat install

# With systemd service (Linux only)
sudo ./bin/technocrat install --systemd
```

## Development Workflow

### 1. Make Changes

Edit the source files as needed.

### 2. Format and Vet

```bash
go run build.go -fmt
go run build.go -vet
```

### 3. Run Tests

```bash
go run build.go -test

# Or use go test directly
go test ./...

# Run specific tests
go test -v ./internal/mcp -run TestCallTool

# Run with coverage
go test -cover ./...
```

### 4. Build and Test

```bash
go generate ./cmd/technocrat
./bin/technocrat server

# In another terminal
curl http://localhost:8080/health
```

## Project Structure

```
technocrat/
├── cmd/
│   └── technocrat/         # Main CLI entry point
│       └── main.go         # Contains //go:generate directive
├── internal/
│   ├── cmd/               # Cobra commands
│   │   ├── root.go        # Root command
│   │   ├── server.go      # Server command
│   │   ├── install.go     # Install command
│   │   ├── uninstall.go   # Uninstall command
│   │   └── version.go     # Version command
│   ├── mcp/               # MCP implementation
│   │   ├── server.go      # HTTP server
│   │   ├── handler.go     # Protocol handlers
│   │   └── handler_test.go # Tests
│   └── installer/         # Installation logic
│       └── installer.go
├── tools/
│   └── build.go           # Build tool tracking
├── build.go               # Main build script
├── justfile               # Just task runner (optional)
├── go.mod                 # Go module definition
├── config.example.json    # Example configuration
└── README.md              # Main documentation
```

## Adding New MCP Tools

To add a new tool to the MCP server:

1. Edit `internal/mcp/handler.go`
2. Add your tool in the `registerDefaultTools()` function:

```go
h.tools["my_tool"] = Tool{
    Name:        "my_tool",
    Description: "What my tool does",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "param1": map[string]interface{}{
                "type":        "string",
                "description": "Description of param1",
            },
        },
        "required": []string{"param1"},
    },
    Handler: func(args map[string]interface{}) (interface{}, error) {
        param1, ok := args["param1"].(string)
        if !ok {
            return nil, fmt.Errorf("param1 must be a string")
        }
        
        // Your tool logic here
        result := doSomething(param1)
        
        return map[string]interface{}{
            "result": result,
        }, nil
    },
}
```

3. Add tests in `internal/mcp/handler_test.go`
4. Rebuild and test:

```bash
go run build.go -test
go generate ./cmd/technocrat
./bin/technocrat server
```

## Troubleshooting

### Build fails with "directory not found"

Make sure you're running commands from the project root directory.

### Tests fail

```bash
# Clean and rebuild
go run build.go -clean
go run build.go -deps
go run build.go -test
```

### Server won't start

Check if port is already in use:

```bash
# Linux/macOS
lsof -i :8080

# Use a different port
./bin/technocrat server --port 9090
```

### go generate fails

Make sure you're in the project root or the cmd/technocrat directory:

```bash
# From project root
go generate ./cmd/technocrat

# From cmd/technocrat
cd cmd/technocrat
go generate
```

## Using Just (Optional)

If you prefer, you can use [Just](https://github.com/casey/just) as a command runner:

```bash
# Install just (if not already installed)
# macOS: brew install just
# Linux: cargo install just

# Then use it
just build
just test
just run
just install
```

See the `justfile` for available commands.

## Next Steps

- Read the [README.md](README.md) for complete documentation
- Explore the MCP protocol endpoints
- Add custom tools and resources
- Set up systemd service (Linux)
- Configure custom settings in config.json

## Support

For issues and questions, please open an issue on GitHub.
