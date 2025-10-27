# AI Agent Integration Guide

Technocrat integrates with 13 different AI coding assistants, allowing you to use your preferred tool while following Spec-Driven Development practices. This guide explains how agent integration works and how to customize it for your needs.

## Overview

When you initialize a project with `technocrat init`, the system:

1. **Creates agent-specific directories** - Sets up folders like `.claude/`, `.github/`, `.cursor/`, etc.
2. **Populates context files** - Copies templates with project information
3. **Configures agent rules** - Sets up AI assistant rules/prompts for your workflow

The `update-agent-context` command keeps these files synchronized with your feature specifications as you develop.

## Supported AI Agents

Technocrat supports both **CLI-based** agents (require command-line tools) and **IDE-based** agents (work within editors).

### CLI-Based Agents

These require a separate CLI tool to be installed:

| Agent | Directory | Context File | CLI Tool | Installation |
|-------|-----------|--------------|----------|--------------|
| **Claude Code** | `.claude/` | `CLAUDE.md` | `claude` | [Install Guide](https://docs.anthropic.com/en/docs/claude-code/setup) |
| **Gemini CLI** | `.gemini/` | `GEMINI.md` | `gemini` | [Install Guide](https://github.com/google-gemini/gemini-cli) |
| **Qwen Code** | `.qwen/` | `QWEN.md` | `qwen` | [Install Guide](https://github.com/QwenLM/qwen-code) |
| **opencode** | `.opencode/` | `OPENCODE.md` | `opencode` | [Install Guide](https://opencode.ai) |
| **Codex CLI** | `.codex/` | `CODEX.md` | `codex` | [Install Guide](https://github.com/openai/codex) |
| **Auggie CLI** | `.augment/` | `AUGGIE.md` | `auggie` | [Install Guide](https://docs.augmentcode.com/cli/setup-auggie/install-auggie-cli) |
| **CodeBuddy** | `.codebuddy/` | `CODEBUDDY.md` | `codebuddy` | [Install Guide](https://www.codebuddy.ai) |
| **Amazon Q Developer CLI** | `.amazonq/` | `Q.md` | `q` | [Install Guide](https://aws.amazon.com/developer/learning/q-developer-cli/) |

### IDE-Based Agents

These work within integrated development environments (no separate CLI needed):

| Agent | Directory | Context File | IDE/Editor |
|-------|-----------|--------------|------------|
| **GitHub Copilot** | `.github/` | `.github/copilot-instructions.md` | VS Code, JetBrains, etc. |
| **Cursor** | `.cursor/` | `.cursor/rules/tchncrt-rules.mdc` | Cursor IDE |
| **Windsurf** | `.windsurf/` | `.windsurf/rules/tchncrt-rules.md` | Windsurf IDE |
| **Kilo Code** | `.kilocode/` | `.kilocode/rules/tchncrt-rules.md` | Kilo Code IDE |
| **Roo Code** | `.roo/` | `.roo/rules/tchncrt-rules.md` | Roo Code IDE |

## Choosing an Agent During Init

### Interactive Selection

```bash
technocrat init my-app
# You'll be prompted to choose from available agents
```

### Command-Line Flag

```bash
# Initialize with Claude Code
technocrat init my-app --ai claude

# Initialize with GitHub Copilot
technocrat init my-app --ai copilot

# Initialize with Cursor
technocrat init my-app --ai cursor-agent
```

### Skip Agent Tool Check

If you want to set up an agent but don't have the CLI tool installed yet:

```bash
technocrat init my-app --ai gemini --ignore-agent-tools
```

## What Gets Created

### For Claude Code (`.claude/`)

```
.claude/
└── CLAUDE.md          # Main context file with project info
```

### For GitHub Copilot (`.github/`)

```
.github/
└── copilot-instructions.md    # Copilot context and instructions
```

### For Cursor (`.cursor/`)

```
.cursor/
└── rules/
    └── tchncrt-rules.mdc      # Cursor rules in MDC format
```

### For Windsurf (`.windsurf/`)

```
.windsurf/
└── rules/
    └── tchncrt-rules.md       # Windsurf workflow rules
```

### For CLI Agents (Gemini, Qwen, etc.)

```
.<agent-name>/
└── <AGENT_NAME>.md            # Agent-specific context file
```

## Agent Context Files

Agent context files provide the AI with:

- **Project Overview** - What the project does
- **Technology Stack** - Languages, frameworks, databases
- **Active Feature** - Current feature being developed
- **Workflow Instructions** - How to use Technocrat commands
- **File Structure** - Project organization
- **Development Guidelines** - Coding standards and practices

### Example: Claude Code Context

```markdown
# Claude Code Context for my-app

## Project Overview

This project follows Spec-Driven Development using Technocrat.

## Technology Stack

- **Language**: Python 3.11
- **Framework**: FastAPI, SQLAlchemy
- **Database**: PostgreSQL 15
- **Testing**: pytest

## Active Feature

Currently working on: 001-add-user-authentication

Spec: specs/001-add-user-authentication/spec.md
Plan: specs/001-add-user-authentication/plan.md

## Workflow Commands

- `technocrat create-feature "description"` - Create new feature
- `technocrat setup-plan` - Generate implementation plan
- `technocrat update-agent-context` - Update this file

## Recent Changes

- 2025-10-27: Added user authentication feature
- 2025-10-26: Initial project setup
```

## Updating Agent Context

The `update-agent-context` command synchronizes agent context files with your current feature specifications.

### Update All Existing Agent Files

```bash
technocrat update-agent-context
```

This will:
1. Detect which agent files exist in your project
2. Parse the current feature's `plan.md`
3. Extract technology stack information
4. Update all agent context files

### Update Specific Agent

```bash
# Update only Claude context
technocrat update-agent-context claude

# Update only Copilot instructions
technocrat update-agent-context copilot
```

### When to Update

Run `update-agent-context`:
- After creating a new feature with `create-feature`
- After modifying your `plan.md` file
- When switching to a different feature branch
- When technology stack changes

### What Gets Updated

From your `plan.md`, the command extracts:

**Language/Version:**
```yaml
language_version: Python 3.11
```
→ Updates "Language: Python 3.11"

**Framework:**
```yaml
primary_dependencies:
  framework: FastAPI, SQLAlchemy
```
→ Updates "Framework: FastAPI, SQLAlchemy"

**Database:**
```yaml
storage:
  database: PostgreSQL 15
```
→ Updates "Database: PostgreSQL 15"

**Project Type:**
```yaml
project_type: backend_api
```
→ Updates "Project Type: backend_api"

## Customizing Agent Files

### Manual Editing

You can manually edit agent context files to add:
- Project-specific conventions
- Team guidelines
- Custom workflows
- External resources

Example additions to `CLAUDE.md`:

```markdown
## Team Guidelines

- Always write tests first (TDD)
- Use type hints for all Python functions
- Follow PEP 8 style guide

## External Resources

- API Documentation: https://api.example.com/docs
- Design System: https://design.example.com
```

### Template Customization

The initial content comes from `templates/agent-file-template.md`. To customize:

1. After `init`, modify `.tchncrt/templates/agent-file-template.md`
2. Future features will use your customized template

```bash
# Edit the template
vim .tchncrt/templates/agent-file-template.md

# New features will use this template
technocrat create-feature "new feature"
```

## Multi-Agent Workflows

You can use multiple agents in the same project:

### Setup Multiple Agents

```bash
# Initialize with Claude
technocrat init my-app --ai claude

# Manually create files for other agents
mkdir -p .github
cp .claude/CLAUDE.md .github/copilot-instructions.md

mkdir -p .cursor/rules
cp .claude/CLAUDE.md .cursor/rules/tchncrt-rules.mdc
```

### Update All Agents

```bash
# Updates all existing agent files
technocrat update-agent-context
```

The command automatically detects and updates:
- `CLAUDE.md` (if exists)
- `.github/copilot-instructions.md` (if exists)
- `.cursor/rules/tchncrt-rules.mdc` (if exists)
- Any other agent context files

## Agent-Specific Features

### Claude Code

- **Slash Commands**: Use `/technocrat-specify`, `/technocrat-plan`, etc.
- **Context Files**: Automatically reads `CLAUDE.md`
- **Project Awareness**: Understands Technocrat workflow

### GitHub Copilot

- **Instructions File**: `.github/copilot-instructions.md` provides context
- **IDE Integration**: Works in VS Code, JetBrains
- **Inline Suggestions**: Context-aware code completions

### Cursor

- **Rules System**: `.cursor/rules/` directory for custom rules
- **Composer**: Multi-file editing with context
- **MDC Format**: Uses Markdown with special syntax

### Windsurf

- **Workflows**: `.windsurf/rules/` for workflow definitions
- **Cascade**: Multi-step task execution
- **Real-time Collab**: Works with team workflows

## Troubleshooting

### Agent Context Not Updating

**Problem**: `update-agent-context` says "no plan.md found"

**Solution**:
```bash
# Ensure you're on a feature branch
git branch

# Create plan if missing
technocrat setup-plan
```

### Wrong Technology Stack

**Problem**: Agent context shows outdated tech stack

**Solution**:
1. Update `specs/XXX-feature/plan.md` with correct info
2. Run `technocrat update-agent-context`

### Agent CLI Not Found

**Problem**: `init` fails with "CLI tool not found"

**Solution**:
```bash
# Install the agent's CLI tool, or skip the check
technocrat init my-app --ai gemini --ignore-agent-tools
```

### Multiple Agent Files Out of Sync

**Problem**: Using multiple agents, context files differ

**Solution**:
```bash
# Update all existing agent files at once
technocrat update-agent-context

# This updates all agent context files in your project
```

## Adding New Agent Support

To add support for a new AI agent, see [AGENTS.md](../AGENTS.md) in the repository root. The process involves:

1. Adding the agent to `agentConfigs` map in `internal/cmd/init.go`
2. Updating command help text
3. Adding agent to documentation
4. Testing init and update-agent-context commands

## Best Practices

### Keep Context Updated

- Run `update-agent-context` after significant plan changes
- Update manually for team-specific guidelines
- Review context files periodically for accuracy

### Use Feature-Specific Context

- Each feature branch should have updated context
- Include relevant architecture decisions
- Document dependencies and integrations

### Version Control

- **DO** commit `.github/copilot-instructions.md`
- **DO** commit `.cursor/rules/`, `.windsurf/rules/`
- **CONSIDER** committing `CLAUDE.md`, `GEMINI.md` (team-specific)
- **DON'T** commit API keys or secrets in context files

### Team Collaboration

- Standardize on one primary agent for consistency
- Document which agent(s) the team uses in README
- Share agent customizations via git

## See Also

- [Installation Guide](installation.md) - Install agent CLI tools
- [Command Reference](commands-reference.md) - `update-agent-context` command
- [Quick Start Guide](quickstart.md) - Using agents with Technocrat
- [AGENTS.md](../AGENTS.md) - Adding new agent support
