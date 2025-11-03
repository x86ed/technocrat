# Template Authoring Guide

This guide explains how to create and customize prompt templates for the Technocrat MCP server.

## Overview

Technocrat uses Go's `text/template` engine to generate dynamic prompts. Templates combine:

- **User input** - Text provided when invoking the prompt
- **Project metadata** - Auto-detected workspace information
- **Feature artifacts** - Existing spec, plan, and task files
- **Template functions** - Helpers for formatting and data access

## Template Structure

### Basic Template

Every template file has two parts:

1. **YAML Frontmatter** - Metadata about the prompt
2. **Template Body** - The actual prompt content

```markdown
---
description: "Brief description of what this prompt does"
scripts:
  cli: optional-cli-command-to-run
---

# Prompt Title

Your template content here with {{.Variables}}
```

### Minimal Example

```markdown
---
description: "Simple greeting prompt"
---

# Hello World

{{if .Arguments}}
Hello, {{.Arguments}}!
{{else}}
Hello, user!
{{end}}
```

## Available Data

### Core Variables

Every template has access to these variables:

#### `.Arguments` (string)

User-provided input text.

```markdown
User requested: {{.Arguments}}
```

**Example values:**

- `"Create REST API endpoints for user management"`
- `"Focus on security testing"`
- `""` (empty if no input)

#### `.CommandName` (string)

The name of the command being executed.

```markdown
Executing {{.CommandName}} workflow...
```

**Example values:** `"spec"`, `"plan"`, `"implement"`, `"analyze"`

#### `.Timestamp` (time.Time)

Current date/time when the prompt is generated.

```markdown
Generated on: {{.Timestamp.Format "2006-01-02 15:04:05"}}
```

### Project Metadata

Auto-detected from your workspace:

#### `.ProjectName` (string)

Project name from directory or constitution.

```markdown
**Project**: {{.ProjectName}}
```

**Detection order:**

1. First heading in `memory/constitution.md` (if not "Constitution", "About", etc.)
2. Text after "## Project Name" heading in constitution
3. Workspace directory name

#### `.FeatureName` (string)

Current feature name if in a `specs/<feature>/` directory.

```markdown
{{if .FeatureName}}
Working on: {{.FeatureName}}
{{end}}
```

**Example:** If you're in `/project/specs/user-auth/`, `.FeatureName` = `"user-auth"`

#### `.WorkspaceRoot` (string)

Absolute path to the workspace root.

```markdown
Workspace: {{.WorkspaceRoot}}
```

**Detection:** Finds the directory containing `memory/` or `.git/`

### Extra Data

#### `.Extra` (map[string]interface{})

Additional arguments passed to the prompt (advanced usage).

```markdown
{{if .Extra.customField}}
Custom: {{.Extra.customField}}
{{end}}
```

## Template Functions

### String Functions

#### `upper` - Convert to uppercase

```markdown
{{upper .CommandName}}
<!-- SPEC -->
```

#### `lower` - Convert to lowercase

```markdown
{{lower .ProjectName}}
<!-- technosync -->
```

#### `title` - Convert to title case

```markdown
{{title .FeatureName}}
<!-- User-Authentication -->
```

#### `trim` - Remove leading/trailing whitespace

```markdown
{{trim .Arguments}}
<!-- Removes accidental spaces from user input -->
```

### Date/Time Functions

#### `now` - Get current time

```markdown
Generated: {{now.Format "Monday, January 2, 2006"}}
<!-- Generated: Tuesday, October 28, 2025 -->
```

#### `.Timestamp.Format` - Format timestamp

```markdown
Date: {{.Timestamp.Format "2006-01-02"}}
Time: {{.Timestamp.Format "15:04:05"}}
ISO: {{.Timestamp.Format "2006-01-02T15:04:05Z07:00"}}
```

**Common formats:**

- `"2006-01-02"` → `2025-10-28`
- `"15:04:05"` → `14:30:00`
- `"Jan 2, 2006"` → `Oct 28, 2025`
- `"Monday"` → `Tuesday`

### Feature File Functions

Read existing artifacts from the current feature directory:

#### `readSpec` - Read spec.md

```markdown
{{if readSpec}}
## Existing Specification

```markdown
{{readSpec}}
```

{{end}}

```

#### `readPlan` - Read plan.md
```markdown
{{if readPlan}}
Previous implementation plan:
{{readPlan}}
{{end}}
```

#### `readTasks` - Read tasks.md

```markdown
{{if readTasks}}
Current tasks:
{{readTasks}}
{{end}}
```

#### `readFile` - Read any file

```markdown
{{readFile "notes.md"}}
{{readFile "research/api-design.md"}}
```

**Note:** These functions return empty string if:

- Not in a feature directory (`specs/<feature>/`)
- File doesn't exist
- File can't be read

## Control Flow

### Conditionals

#### Basic if/else

```markdown
{{if .Arguments}}
User input: {{.Arguments}}
{{else}}
No input provided.
{{end}}
```

#### Check for non-empty strings

```markdown
{{if .FeatureName}}
Feature: {{.FeatureName}}
{{end}}
```

#### Multiple conditions

```markdown
{{if .ProjectName}}
Project: {{.ProjectName}}
  {{if .FeatureName}}
  Feature: {{.FeatureName}}
  {{end}}
{{end}}
```

#### Check function results

```markdown
{{if readSpec}}
A spec exists for this feature.
{{else}}
No spec found - create one first.
{{end}}
```

### Comparisons

```markdown
{{if eq .CommandName "spec"}}
Creating specification...
{{end}}

{{if ne .Arguments ""}}
User provided guidance.
{{end}}
```

**Operators:** `eq` (equal), `ne` (not equal), `lt` (less than), `gt` (greater than)

### Logical Operators

```markdown
{{if and .ProjectName .FeatureName}}
Project: {{.ProjectName}}, Feature: {{.FeatureName}}
{{end}}

{{if or .Arguments (readSpec)}}
Context is available.
{{end}}
```

## Practical Examples

### Example 1: User Input with Fallback

```markdown
---
description: "Process user request with sensible defaults"
---

# Task Processor

## Guidance

{{if .Arguments}}
{{trim .Arguments}}
{{else}}
_No specific guidance provided. Follow standard workflow._
{{end}}
```

### Example 2: Show Project Context

```markdown
---
description: "Display project and feature information"
---

# Context Summary

{{if .ProjectName}}
**Project**: {{.ProjectName}}
{{end}}

{{if .FeatureName}}
**Feature**: {{.FeatureName}}
**Location**: `{{.WorkspaceRoot}}/specs/{{.FeatureName}}/`
{{else}}
**Location**: {{.WorkspaceRoot}}
_(Not in a feature directory)_
{{end}}
```

### Example 3: Reference Previous Artifacts

```markdown
---
description: "Implementation with full context"
---

# Implementation Guide

## User Guidance

{{if .Arguments}}
{{.Arguments}}
{{else}}
Follow the implementation plan below.
{{end}}

{{if readSpec}}
## Feature Specification

The following specification should guide your implementation:

```markdown
{{readSpec}}
```

{{end}}
{{if readPlan}}

## Implementation Plan

Follow this plan:

```markdown
{{readPlan}}
```

{{end}}
{{if readTasks}}

## Task Checklist

```markdown
{{readTasks}}
```

{{else}}
**Note**: No tasks file found. Create one with `/tasks` first.
{{end}}

```

### Example 4: Conditional Workflow

```markdown
---
description: "Adaptive workflow based on available context"
---

# Feature Analysis

{{if readSpec}}
## Review Existing Spec

Please review the current specification:

{{readSpec}}

{{if .Arguments}}
Focus your analysis on: {{.Arguments}}
{{end}}

{{else}}
## Create New Spec

No specification exists yet.

{{if .Arguments}}
Create a spec based on: {{.Arguments}}
{{else}}
Create a spec for this feature from scratch.
{{end}}
{{end}}
```

### Example 5: Advanced Formatting

```markdown
---
description: "Report with formatted metadata"
---

# {{upper .CommandName}} Report

**Generated**: {{now.Format "Monday, January 2, 2006 at 3:04 PM"}}
**Project**: {{.ProjectName | title}}
{{if .FeatureName}}**Feature**: {{.FeatureName}}{{end}}

---

{{if .Arguments}}
## Analysis Focus

{{trim .Arguments}}

{{end}}
## Artifacts Status

- **Specification**: {{if readSpec}}✓ Exists{{else}}✗ Missing{{end}}
- **Plan**: {{if readPlan}}✓ Exists{{else}}✗ Missing{{end}}
- **Tasks**: {{if readTasks}}✓ Exists{{else}}✗ Missing{{end}}
```

## Best Practices

### 1. Always Provide Fallbacks

❌ **Bad:**

```markdown
User input: {{.Arguments}}
```

✅ **Good:**

```markdown
{{if .Arguments}}
User input: {{.Arguments}}
{{else}}
_No specific input provided._
{{end}}
```

### 2. Check Before Reading Files

❌ **Bad:**

```markdown
## Specification

{{readSpec}}
```

✅ **Good:**

```markdown
{{if readSpec}}
## Specification

{{readSpec}}
{{end}}
```

### 3. Trim User Input

❌ **Bad:**

```markdown
Task: {{.Arguments}}
```

✅ **Good:**

```markdown
Task: {{trim .Arguments}}
```

### 4. Use Descriptive Section Headers

❌ **Bad:**

```markdown
{{if .Arguments}}
{{.Arguments}}
{{end}}
```

✅ **Good:**

```markdown
{{if .Arguments}}
## User Guidance

{{.Arguments}}
{{end}}
```

### 5. Document Available Variables

Add comments at the top of complex templates:

```markdown
---
description: "My custom workflow"
---

<!-- Available variables:
     - .Arguments: User input
     - .ProjectName: Project name
     - .FeatureName: Current feature
     - .WorkspaceRoot: Workspace path
     
     Available functions:
     - readSpec, readPlan, readTasks
     - upper, lower, trim, title
-->

# Workflow
...
```

### 6. Test Edge Cases

Test your template with:

- ✅ User input provided
- ✅ No user input (empty)
- ✅ In feature directory
- ✅ Outside feature directory (no .FeatureName)
- ✅ Files exist (readSpec returns content)
- ✅ Files missing (readSpec returns empty)

## Common Patterns

### Pattern: Optional Section

```markdown
{{if .FeatureName}}
## Feature Context

Working on feature: **{{.FeatureName}}**
{{end}}
```

### Pattern: Either/Or Display

```markdown
{{if .Arguments}}
Focus: {{.Arguments}}
{{else if readSpec}}
Review existing spec for context.
{{else}}
Create a new specification.
{{end}}
```

### Pattern: List with Conditionals

```markdown
## Available Context

{{if .ProjectName}}- Project: {{.ProjectName}}{{end}}
{{if .FeatureName}}- Feature: {{.FeatureName}}{{end}}
{{if readSpec}}- Specification exists{{end}}
{{if readPlan}}- Implementation plan exists{{end}}
```

### Pattern: Nested Markdown Code Blocks

```markdown
{{if readSpec}}
## Current Spec

\`\`\`markdown
{{readSpec}}
\`\`\`
{{end}}
```

## Troubleshooting

### Template Parse Errors

**Error:** `template: workflow:5: unexpected "}" in command`

**Cause:** Syntax error in template (unclosed block, typo)

**Fix:** Check for:

- Matching `{{if}}...{{end}}` pairs
- Correct function names
- Proper variable access (`.FieldName` not `FieldName`)

### Empty Output

**Problem:** Template renders but shows nothing

**Causes:**

1. All conditionals are false
2. Variables are empty
3. File reading returns nothing

**Fix:** Add fallback messages:

```markdown
{{if .Arguments}}
{{.Arguments}}
{{else}}
_DEBUG: No arguments provided_
{{end}}
```

### Variables Not Substituted

**Problem:** Seeing `{{.ProjectName}}` in output

**Cause:** Variable doesn't exist or typo

**Fix:** Check variable names (case-sensitive):

- ✅ `.ProjectName`
- ❌ `.projectName`
- ❌ `.Project_Name`

## Advanced Topics

### Custom Extra Fields

Pass custom data via the `Extra` map:

```go
// In handler code
templateData := TemplateData{
    Arguments: userInput,
    Extra: map[string]interface{}{
        "priority": "high",
        "deadline": "2025-11-01",
    },
}
```

```markdown
<!-- In template -->
{{if .Extra.priority}}
Priority: {{.Extra.priority}}
{{end}}
```

### Pipelines

Chain functions together:

```markdown
{{.ProjectName | lower | title}}
{{.Arguments | trim | upper}}
```

### Range Over Lists

(Advanced - requires Extra map with lists)

```markdown
{{range .Extra.items}}
- {{.}}
{{end}}
```

## Reference

### Quick Variable Reference

| Variable | Example Value |
|----------|---------------|
| `.Arguments` | `"Add REST API"` |
| `.CommandName` | `"spec"` |
| `.Timestamp` | `2025-10-28 14:30:00` |
| `.ProjectName` | `"TechnoSync"` |
| `.FeatureName` | `"user-auth"` |
| `.WorkspaceRoot` | `"/home/dev/project"` |

### Quick Function Reference

| Function | Usage | Result |
|----------|-------|--------|
| `upper` | `{{upper "hello"}}` | `HELLO` |
| `lower` | `{{lower "HELLO"}}` | `hello` |
| `title` | `{{title "hello world"}}` | `Hello World` |
| `trim` | `{{trim "  hi  "}}` | `hi` |
| `now` | `{{now.Format "2006"}}` | `2025` |
| `readSpec` | `{{readSpec}}` | _(file contents)_ |
| `readPlan` | `{{readPlan}}` | _(file contents)_ |
| `readTasks` | `{{readTasks}}` | _(file contents)_ |
| `readFile` | `{{readFile "x.md"}}` | _(file contents)_ |

## See Also

- [MCP Server Guide](mcp-server.md) - Server setup and API
- [Commands Reference](commands-reference.md) - CLI commands
- [Go text/template docs](https://pkg.go.dev/text/template) - Full template syntax
