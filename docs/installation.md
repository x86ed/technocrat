# Installation Guide

## Prerequisites

### Required

- **Go 1.24 or later** - [Download Go](https://golang.org/dl/)
- **Git** - [Download Git](https://git-scm.com/downloads)

### Optional (for Spec-Driven Development)

Choose one or more AI coding assistants to integrate with Technocrat:

#### IDE-Based (No CLI Required)

- **[GitHub Copilot](https://github.com/features/copilot)** - Built into VS Code and compatible editors
- **[Cursor](https://cursor.sh/)** - AI-first code editor
- **[Windsurf](https://codeium.com/windsurf)** - Codeium's AI editor
- **[Kilo Code](https://kilocode.com/)** - Collaborative AI coding
- **[Roo Code](https://roocode.com/)** - AI pair programming

#### CLI-Based (Requires Installation)

- **[Claude Code](https://docs.anthropic.com/en/docs/claude-code/setup)** - Anthropic's CLI tool
- **[Gemini CLI](https://github.com/google-gemini/gemini-cli)** - Google's AI CLI
- **[Qwen Code](https://github.com/QwenLM/qwen-code)** - Alibaba's coding assistant
- **[opencode](https://opencode.ai)** - Open source AI coding tool
- **[Codex CLI](https://github.com/openai/codex)** - OpenAI's CLI
- **[Auggie CLI](https://docs.augmentcode.com/cli/setup-auggie/install-auggie-cli)** - Augment's CLI tool
- **[CodeBuddy](https://www.codebuddy.ai)** - AI coding companion
- **[Amazon Q Developer CLI](https://aws.amazon.com/q/developer/)** - AWS's AI assistant

## Installation Methods

### Option 1: Install from Go (Recommended)

Install directly from GitHub using Go:

```bash
go install github.com/x86ed/technocrat/cmd/technocrat@latest
```

This installs `technocrat` to your `$GOPATH/bin` directory. Make sure it's in your PATH:

```bash
# Add to your .bashrc, .zshrc, or equivalent
export PATH="$PATH:$(go env GOPATH)/bin"

# Verify installation
technocrat version
```

### Option 2: Build from Source

Clone and build the repository:

```bash
# Clone the repository
git clone https://github.com/x86ed/technocrat.git
cd technocrat

# Build using go generate (recommended)
go generate ./cmd/technocrat

# Or use the build script
go run build.go -build

# Binary will be in ./bin/technocrat
./bin/technocrat version

# Optionally, install to GOPATH
go install ./cmd/technocrat
```

### Option 3: Download Pre-built Binary

Download the latest release for your platform from the [releases page](https://github.com/x86ed/technocrat/releases).

```bash
# Example for Linux/macOS
curl -L https://github.com/x86ed/technocrat/releases/latest/download/technocrat-{os}-{arch}.tar.gz | tar xz
sudo mv technocrat /usr/local/bin/
```

## Verification

Verify your installation:

```bash
# Check version
technocrat version

# Check available commands
technocrat --help

# Check tool availability (optional)
technocrat check
```

You should see output like:

```
Technocrat 0.3.0 (commit: abc123)
```

## Next Steps

- Initialize your first project: [Quick Start Guide](quickstart.md)
- Learn about available commands: [Command Reference](commands-reference.md)
- Configure your AI agent: [Agent Integration](agent-integration.md)
