# Quick Start Guide

This guide will walk you through creating your first feature using Technocrat's Spec-Driven Development workflow.

## Prerequisites

- Technocrat installed ([Installation Guide](installation.md))
- Git installed and configured
- An AI coding assistant (optional but recommended)

## The Technocrat Workflow

Technocrat provides a structured 5-step process for feature development:

1. **Initialize Project** - Set up templates and agent configuration
2. **Create Feature** - Generate branch and spec directory
3. **Setup Plan** - Create implementation plan
4. **Update Context** - Sync info to AI agent
5. **Implement** - Build with AI assistance

Let's walk through each step with a practical example.

## Example: Building a Task Manager

### Step 1: Initialize Your Project

First, create a new project with your preferred AI agent:

```bash
# Initialize with Claude
technocrat init task-manager --ai claude

# Or with GitHub Copilot
technocrat init task-manager --ai copilot

# Supported agents: claude, copilot, gemini, cursor-agent, qwen, opencode,
#                   codex, windsurf, kilocode, auggie, roo, codebuddy, q
```

This creates:

```
task-manager/
├── .tchncrt/
│   └── templates/           # Spec, plan, and task templates
├── specs/                   # Feature specifications directory
├── memory/
│   └── constitution.md      # Project principles
└── CLAUDE.md               # Agent context file (or corresponding agent file)
```

### Step 2: Create a New Feature

Create a feature branch and specification directory:

```bash
cd task-manager
technocrat create-feature "add user authentication system"
```

from. There will be no password required. When you click on a user, you go into the main view, which displays the list of

This creates:projects. When you click on a project, you open the Kanban board for that project. You're going to see the columns.

You'll be able to drag and drop cards back and forth between different columns. You will see any cards that are

- `specs/001-add-user-authentication-system/` directoryassigned to you, the currently logged in user, in a different color from all the other ones, so you can quickly

- Git branch: `001-add-user-authentication-system`see yours. You can edit any comments that you make, but you can't edit comments that other people made. You can

- `spec.md` from templatedelete any comments that you made, but you can't delete comments anybody else made.

```

### Step 3: Setup Implementation Plan

### Step 2: Refine the Specification

Generate the plan.md file:

After the initial specification is created, clarify any missing requirements:

```bash

technocrat setup-plan```text

```For each sample project or project that you create there should be a variable number of tasks between 5 and 15

tasks for each one randomly distributed into different states of completion. Make sure that there's at least

This automatically:
- Creates `specs/001-add-user-authentication-system/` directory
- Creates git branch `001-add-user-authentication-system`
- Copies `spec.md` from template
- Sets `TCHNCRT_FEATURE` environment variable

Your directory now looks like:

```
specs/
└── 001-add-user-authentication-system/
    └── spec.md              # Feature specification template
```

### Step 3: Setup Implementation Plan

Edit `spec.md` with your feature requirements, then generate the plan:

```bash
technocrat setup-plan
```

This creates `plan.md` in your feature directory:

```
specs/001-add-user-authentication-system/
├── spec.md                  # Your requirements
└── plan.md                  # Implementation plan (to be filled in)
```

Edit `specs/001-add-user-authentication-system/plan.md` with technical details:

- Language and version
- Framework and dependencies
- Database choice
- Architecture decisions

### Step 4: Update Agent Context

After filling in your plan, sync the information to your AI agent:

```bash
technocrat update-agent-context
```

This updates your agent's context file (e.g., `CLAUDE.md`) with:
- Current feature information
- Technology stack
- Project structure
- Recent changes

### Step 5: Implement with AI Assistance

Now your AI agent has all the context needed. Start implementing:

1. Open your IDE/editor with the AI agent
2. Ask the agent to implement the feature according to the spec
3. The agent will use information from:
   - `spec.md` (requirements)
   - `plan.md` (technical approach)
   - Agent context file (project info)
   - `memory/constitution.md` (project principles)

## Tips for Success

### Write Clear Specifications

- Focus on **what** the feature does, not **how** to implement it
- Include success criteria
- Specify user stories or scenarios
- List acceptance criteria

### Detail Your Plans

- Choose specific technologies (don't say "a database", say "PostgreSQL 15")
- Explain architectural decisions
- Note any constraints or requirements
- Document dependencies

### Keep Context Updated

Run `technocrat update-agent-context` whenever you:
- Change technology stack
- Update the plan
- Switch to a new feature
- Make significant architectural changes

### Use Templates

The templates in `.tchncrt/templates/` guide you through:
- Writing complete specifications
- Creating detailed plans
- Breaking down tasks
- Tracking progress

## Next Steps

- Read the [Command Reference](commands-reference.md) for all available commands
- Learn about [Agent Integration](agent-integration.md) to customize your AI assistant setup
- Check out [MCP Server](mcp-server.md) to use Technocrat's API programmatically
- See [Local Development](local-development.md) to contribute to Technocrat itself
