# Command Reference

Complete reference for all Technocrat CLI commands.

## Global Flags

```bash
--help, -h    Show help information for any command
--version     Show version information
```

## Commands Overview

| Command | Description |
|---------|-------------|
| `init` | Initialize a new project with templates and agent config |
| `create-feature` | Create a new feature branch and spec directory |
| `setup-plan` | Set up implementation plan for current feature |
| `update-agent-context` | Update AI agent context files with feature info |
| `check` | Check that required development tools are installed |
| `server` | Start the MCP protocol server |
| `version` | Display version information |

---

## init

Initialize a new project with Spec-Driven Development templates and AI agent configuration.

### Usage

```bash
technocrat init <project-name> [flags]
technocrat init . [flags]
```

### Arguments

- `project-name` - Name of the project to initialize (or `.` for current directory)

### Flags

```bash
--ai string              AI assistant to use (required)
                         Options: claude, copilot, gemini, cursor-agent, qwen,
                                  opencode, codex, windsurf, kilocode, auggie,
                                  roo, codebuddy, q
--ignore-agent-tools     Skip checking for agent CLI tool availability
--no-templates           Skip installing templates
```

### Examples

```bash
# Initialize new project with Claude
technocrat init my-app --ai claude

# Initialize in current directory with Copilot
technocrat init . --ai copilot

# Initialize without checking for agent tools
technocrat init test-project --ai gemini --ignore-agent-tools

# Initialize without templates
technocrat init minimal --ai claude --no-templates
```

### What It Creates

```
project-name/
├── .tchncrt/
│   └── templates/              # Spec, plan, task templates
├── specs/                      # Feature specifications directory
├── memory/
│   └── constitution.md         # Project principles
└── AGENT_FILE.md              # Agent-specific context file
```

Agent files created based on `--ai` flag:
- `claude` → `CLAUDE.md`
- `copilot` → `.github/copilot-instructions.md`
- `gemini` → `GEMINI.md`
- `cursor-agent` → `.cursor/rules/tchncrt-rules.mdc`
- `windsurf` → `.windsurf/rules/tchncrt-rules.md`
- etc.

### Output

```
✓ Created project directory: my-app
✓ Created .tchncrt directory structure
✓ Installed templates
✓ Created specs directory
✓ Created memory directory with constitution
✓ Created CLAUDE.md agent file
✓ Project initialized successfully

Next steps:
  cd my-app
  technocrat create-feature "your feature description"
```

---

## create-feature

Create a new feature branch and specification directory.

### Usage

```bash
technocrat create-feature <description> [flags]
```

### Arguments

- `description` - Brief description of the feature (multiple words allowed)

### Flags

```bash
--json    Output result in JSON format
```

### Examples

```bash
# Create a feature
technocrat create-feature "add user authentication"

# Create with JSON output
technocrat create-feature "add dashboard" --json
```

### What It Creates

- Feature directory: `specs/XXX-feature-name/`
- Git branch: `XXX-feature-name` (if git available)
- `spec.md` file from template
- Sets `TCHNCRT_FEATURE` environment variable

Feature numbers are auto-incremented (001, 002, 003, etc.)

### Output

**Text format:**
```
[tchncrt] Created feature directory: specs/001-add-user-authentication
[tchncrt] Created branch: 001-add-user-authentication
[tchncrt] Copied spec template to: specs/001-add-user-authentication/spec.md
[tchncrt] Feature created: 001-add-user-authentication
```

**JSON format (`--json`):**
```json
{
  "BRANCH_NAME": "001-add-user-authentication",
  "SPEC_FILE": "specs/001-add-user-authentication/spec.md",
  "FEATURE_NUM": "001",
  "HAS_GIT": true
}
```

---

## setup-plan

Set up the implementation plan for the current feature.

### Usage

```bash
technocrat setup-plan [flags]
```

### Flags

```bash
--json    Output result in JSON format
```

### Prerequisites

- Must be on a feature branch (format: `XXX-feature-name`)
- Feature directory must exist in `specs/`

### Examples

```bash
# Set up plan for current feature
technocrat setup-plan

# Get output in JSON
technocrat setup-plan --json
```

### What It Creates

- `plan.md` in current feature directory
- Copied from `.tchncrt/templates/plan-template.md`

### Output

**Text format:**
```
Copied plan template to specs/001-add-user-authentication/plan.md

Feature: 001-add-user-authentication
Spec file: specs/001-add-user-authentication/spec.md
Plan file: specs/001-add-user-authentication/plan.md
```

**JSON format (`--json`):**
```json
{
  "FEATURE_SPEC": "specs/001-add-user-authentication/spec.md",
  "IMPL_PLAN": "specs/001-add-user-authentication/plan.md",
  "SPECS_DIR": "specs/001-add-user-authentication",
  "BRANCH": "001-add-user-authentication",
  "HAS_GIT": true
}
```

---

## update-agent-context

Update AI agent context files with information from the current feature's plan.

### Usage

```bash
technocrat update-agent-context [agent-type]
```

### Arguments

- `agent-type` (optional) - Specific agent to update

Valid agent types:
- `claude` - Claude Code
- `gemini` - Gemini CLI
- `copilot` - GitHub Copilot
- `cursor` - Cursor IDE
- `qwen` - Qwen Code
- `opencode` - opencode
- `codex` - Codex CLI
- `windsurf` - Windsurf
- `kilocode` - Kilo Code
- `auggie` - Auggie CLI
- `roo` - Roo Code
- `codebuddy` - CodeBuddy
- `q` - Amazon Q Developer CLI

### Prerequisites

- Must be on a feature branch
- `plan.md` must exist in feature directory

### Examples

```bash
# Update all existing agent files
technocrat update-agent-context

# Update specific agent
technocrat update-agent-context claude
technocrat update-agent-context copilot
```

### What It Updates

Extracts from `plan.md`:
- Language/Version
- Primary Dependencies (framework)
- Storage (database)
- Project Type

Updates agent files with:
- Technology stack
- Active feature information
- Recent changes
- Project structure
- Relevant commands

### Output

```
INFO: === Updating agent context files for feature 001-add-user-authentication ===
INFO: Parsing plan data from specs/001-add-user-authentication/plan.md
INFO: Found language: Python 3.11
INFO: Found framework: FastAPI, SQLAlchemy
INFO: Found database: PostgreSQL 15
INFO: Updating Claude Code context file: CLAUDE.md
✓ Updated existing Claude Code context file

INFO: Summary of changes:
  - Added language: Python 3.11
  - Added framework: FastAPI, SQLAlchemy
  - Added database: PostgreSQL 15
```

---

## check

Verify that required development tools are installed.

### Usage

```bash
technocrat check
```

### What It Checks

**Version Control:**
- Git

**AI Assistants (CLI-based):**
- Claude CLI (`claude`)
- Gemini CLI (`gemini`)
- Qwen Code (`qwen`)
- opencode (`opencode`)
- Codex CLI (`codex`)
- Auggie CLI (`auggie`)
- CodeBuddy (`codebuddy`)
- Amazon Q Developer CLI (`q`)

**Code Editors:**
- VS Code (`code`)
- VS Code Insiders (`code-insiders`)
- Cursor (`cursor`)
- Windsurf (`windsurf`)

### Output

```
╭─────────────────────────────────────────╮
│     Tool Availability Check             │
╰─────────────────────────────────────────╯

✓ Git - installed
✓ Claude CLI - installed
✗ Gemini CLI - not installed (optional)
✓ VS Code - installed
✗ Cursor - not installed (optional)

Summary: 3/5 tools available
```

---

## server

Start the Model Context Protocol (MCP) server.

### Usage

```bash
technocrat server [flags]
```

### Flags

```bash
-p, --port int    Port to listen on (default: 8080)
```

### Examples

```bash
# Start server on default port (8080)
technocrat server

# Start on custom port
technocrat server --port 9090
```

### Endpoints

The server implements the MCP protocol with the following endpoints:

- `POST /mcp/v1/initialize` - Initialize MCP connection
- `GET /mcp/v1/tools/list` - List available tools
- `POST /mcp/v1/tools/call` - Execute a tool
- `GET /mcp/v1/resources/list` - List available resources
- `POST /mcp/v1/resources/read` - Read a resource
- `GET /mcp/v1/prompts/list` - List available prompts
- `POST /mcp/v1/prompts/get` - Get a prompt
- `GET /health` - Health check

### Output

```
Starting Technocrat MCP Server on port 8080...
Server listening on :8080
```

See [MCP Server Guide](mcp-server.md) for detailed endpoint documentation.

---

## version

Display version and build information.

### Usage

```bash
technocrat version
```

### Output

```
Technocrat 0.3.0 (commit: abc123)
```

---

## Error Messages

### Common Errors and Solutions

**"failed to get feature paths"**
- **Cause**: Not in a project with `.tchncrt/` directory
- **Solution**: Run `technocrat init` first

**"not on a proper feature branch"**
- **Cause**: Branch name doesn't follow `XXX-feature-name` format
- **Solution**: Use `technocrat create-feature` to create proper branch

**"no plan.md found"**
- **Cause**: Plan file doesn't exist in feature directory
- **Solution**: Run `technocrat setup-plan` first

**"template not found"**
- **Cause**: Templates not installed in `.tchncrt/templates/`
- **Solution**: Run `technocrat init` with templates, or check template location

**"agent CLI not found"**
- **Cause**: Selected AI agent's CLI tool is not installed
- **Solution**: Install the agent's CLI tool, or use `--ignore-agent-tools` flag

---

## Environment Variables

### TCHNCRT_FEATURE

Set automatically by `create-feature` to track the current feature:

```bash
export TCHNCRT_FEATURE="001-add-user-authentication"
```

Can be set manually if needed:

```bash
# Set to match current branch
export TCHNCRT_FEATURE=$(git branch --show-current)
```

---

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Invalid arguments or flags

---

## See Also

- [Quick Start Guide](quickstart.md) - Get started with Technocrat
- [Installation Guide](installation.md) - Install Technocrat
- [Agent Integration](agent-integration.md) - Configure AI agents
- [MCP Server Guide](mcp-server.md) - Use the MCP server
- [Local Development](local-development.md) - Contribute to Technocrat
