# Technocrat Quick Reference

## Build Commands

```bash
# Build the project
go generate ./cmd/technocrat
# or
go run build.go -build

# Clean build artifacts
go run build.go -clean

# Run tests
go run build.go -test

# Format code
go run build.go -fmt

# Run vet
go run build.go -vet

# Install
sudo go run build.go -install

# Uninstall
sudo go run build.go -uninstall
```

## CLI Commands

```bash
# Show help
technocrat --help

# Show version
technocrat version

# Start server (default port 8080)
technocrat server

# Start server on custom port
technocrat server --port 9090

# Install system-wide
sudo technocrat install

# Install with systemd (Linux)
sudo technocrat install --systemd

# Uninstall
sudo technocrat uninstall

# Uninstall with systemd
sudo technocrat uninstall --systemd
```

## MCP API Endpoints

```bash
# Health check
curl http://localhost:8080/health

# Initialize
curl -X POST http://localhost:8080/mcp/v1/initialize

# List tools
curl http://localhost:8080/mcp/v1/tools/list

# Call tool
curl -X POST http://localhost:8080/mcp/v1/tools/call \
  -H "Content-Type: application/json" \
  -d '{"name": "echo", "arguments": {"message": "Hello"}}'

# List resources
curl http://localhost:8080/mcp/v1/resources/list

# Read resource
curl -X POST http://localhost:8080/mcp/v1/resources/read \
  -H "Content-Type: application/json" \
  -d '{"uri": "info://server"}'

# List prompts
curl http://localhost:8080/mcp/v1/prompts/list

# Get prompt
curl -X POST http://localhost:8080/mcp/v1/prompts/get \
  -H "Content-Type: application/json" \
  -d '{"name": "welcome", "arguments": {"name": "User"}}'
```

## Development Workflow

```bash
# 1. Make changes to code

# 2. Format
go run build.go -fmt

# 3. Check with vet
go run build.go -vet

# 4. Run tests
go run build.go -test

# 5. Build
go generate ./cmd/technocrat

# 6. Test locally
./bin/technocrat server

# 7. Test in another terminal
curl http://localhost:8080/health
```

## File Locations

### Configuration
- Linux: `/etc/technocrat/config.json`
- macOS: `~/.config/technocrat/config.json`
- Other: `~/.technocrat/config.json`

### Binary (when installed)
- System: `/usr/local/bin/technocrat`
- User: `~/.local/bin/technocrat`
- Go install: `$GOPATH/bin/technocrat`

### Systemd Service (Linux)
- Service file: `/etc/systemd/system/technocrat.service`
- Control: `sudo systemctl {start|stop|restart|status} technocrat`
- Logs: `sudo journalctl -u technocrat -f`

## Common Tasks

### Fresh Build
```bash
go run build.go -clean
go generate ./cmd/technocrat
```

### Full Check
```bash
go run build.go -fmt && \
go run build.go -vet && \
go run build.go -test && \
go run build.go -build
```

### Install for Development
```bash
go run build.go -install -install-dir=$HOME/.local/bin
export PATH="$HOME/.local/bin:$PATH"
```

### Update Dependencies
```bash
go run build.go -deps
```

## Using Just (Optional)

If you have Just installed:

```bash
just build      # Build
just test       # Test
just run        # Build and run
just install    # Install
just check      # Format, vet, and test
just clean      # Clean
```

## Troubleshooting

### Port already in use
```bash
lsof -i :8080
# or use different port
technocrat server --port 9090
```

### Permission denied during install
```bash
sudo go run build.go -install
```

### Can't find binary after install
```bash
# Check PATH
echo $PATH

# For ~/.local/bin
export PATH="$HOME/.local/bin:$PATH"

# For GOPATH
export PATH="$GOPATH/bin:$PATH"
```

### Build fails
```bash
# Clean and rebuild
go run build.go -clean
go run build.go -deps
go run build.go -build
```

## Resources

- [Getting Started Guide](GETTING_STARTED.md)
- [Full Documentation](README.md)
- [Migration Guide](MIGRATION.md)
