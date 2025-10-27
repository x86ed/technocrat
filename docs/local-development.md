# Local Development Guide

This guide shows how to develop and test Technocrat locally.

## Prerequisites

- Go 1.24 or later
- Git
- Your favorite code editor (VS Code, GoLand, etc.)

## 1. Clone the Repository

```bash
git clone https://github.com/x86ed/technocrat.git
cd technocrat

# Create a feature branch
git checkout -b your-feature-branch
```

## 2. Build from Source

### Using go generate (Recommended)

```bash
# Generate and build in one step
go generate ./cmd/technocrat

# Binary will be in ./bin/technocrat
./bin/technocrat --help
```

### Using the build script

```bash
# Download dependencies
go run build.go -deps

# Run tests
go run build.go -test

# Build the binary
go run build.go -build

# Binary will be in ./bin/technocrat
./bin/technocrat version
```

### Direct go build

```bash
# Build with go directly
go build -o bin/technocrat ./cmd/technocrat

# Run it
./bin/technocrat --help
```

## 3. Quick Testing (No Build Required)

Run Technocrat directly without building:

```bash
# Run directly from source
go run cmd/technocrat/main.go --help
go run cmd/technocrat/main.go version
go run cmd/technocrat/main.go init test-project --ai claude

# This is the fastest way to test changes during development
```

## 4. Running Tests

### Run all tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test ./internal/cmd
go test ./internal/mcp
```

### Run tests with coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# View coverage in browser
go tool cover -html=coverage.out
```

### Run specific tests

```bash
# Run tests matching a pattern
go test -run TestCreateFeature ./internal/cmd
go test -run TestUpdateAgentContext ./internal/cmd
```

## 5. Code Quality

### Format and lint

```bash
# Format all Go files
go fmt ./...

# Run go vet
go vet ./...

# Or use the build script
go run build.go -fmt
go run build.go -vet
```

## 6. Install Locally

```bash
# Install to GOPATH/bin
go install ./cmd/technocrat

# Now run from anywhere
technocrat version
```

## 7. Development Workflow

### Typical cycle

```bash
# 1. Make changes
# 2. Run tests
go test ./internal/cmd -v

# 3. Quick test with go run
go run cmd/technocrat/main.go init test --ai copilot

# 4. Build and test binary
go generate ./cmd/technocrat
./bin/technocrat init another-test

# 5. Format and vet
go fmt ./...
go vet ./...
```

## 8. Debugging

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug the application
dlv debug ./cmd/technocrat -- init test-project --ai claude

# Debug a test
dlv test ./internal/cmd -- -test.run TestCreateFeature
```

## 9. Submitting Changes

### Before creating a PR

```bash
# 1. Run all tests
go test ./...

# 2. Format code
go fmt ./...

# 3. Run linter
go vet ./...

# 4. Update docs

# 5. Commit and push
git add .
git commit -m "feat: description"
git push origin your-feature-branch
```

## 10. Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Cobra Documentation](https://cobra.dev/)
- [Contributing Guide](../CONTRIBUTING.md)
- [Adding New AI Agents](../AGENTS.md)

## Troubleshooting

### "command not found: technocrat"

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### "cannot find package"

```bash
go mod download
```

### Build errors

```bash
go run build.go -clean
go run build.go -build
```

