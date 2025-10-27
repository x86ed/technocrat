# Technocrat

*Spec-Driven Development with AI Agent Integration*

**A comprehensive framework that combines structured feature development with AI coding assistants through the Model Context Protocol.**

## What is Technocrat?

Technocrat is a dual-purpose tool that provides:

1. **Spec-Driven Development CLI**: A complete workflow for managing software features with structured specifications, implementation plans, and automated context management
2. **MCP Server**: A fully-compliant Model Context Protocol server that exposes tools, resources, and prompts for AI agent integration

Built entirely in Go, Technocrat delivers a single-binary solution that works seamlessly with 13+ AI coding assistants including Claude, GitHub Copilot, Gemini, Cursor, Windsurf, and more.

## Why Spec-Driven Development?

Spec-Driven Development emphasizes creating clear, executable specifications before implementation. This approach:

- **Increases clarity** by defining the "_what_" before the "_how_"
- **Improves AI collaboration** by providing structured context that AI agents can understand
- **Enables better planning** through multi-step refinement rather than one-shot generation
- **Maintains consistency** across features with templates and conventions
- **Supports iteration** through structured feedback loops

## Getting Started

- [Installation Guide](installation.md) - Set up Technocrat on your system
- [Quick Start Guide](quickstart.md) - Your first feature with Technocrat
- [Command Reference](commands-reference.md) - Complete CLI documentation
- [Agent Integration](agent-integration.md) - Configure your AI coding assistant
- [Local Development](local-development.md) - Contributing to Technocrat

## Key Features

### For Developers

- **Single Binary**: Pure Go implementation, no runtime dependencies
- **Cross-Platform**: Works on Linux, macOS, and Windows
- **Git Integration**: Automatic branch management and feature tracking
- **Template System**: Customizable templates for specs, plans, and tasks
- **Multi-Agent Support**: Works with your preferred AI coding assistant

### For AI Agents

- **Automated Context**: Keep AI agents informed about project structure and features
- **MCP Protocol**: Standard protocol for tool and resource access
- **Rich Specifications**: Structured format that AI can understand and act on
- **Live Updates**: Agent context files automatically sync with feature development

### For Teams

- **Consistent Workflow**: Standardized approach across all team members
- **Technology Independent**: Use any programming language or framework
- **Organizational Constraints**: Incorporate company standards and practices
- **Audit Trail**: Git-based history of all specifications and plans

## Development Workflow

Technocrat supports a structured workflow for feature development:

1. **Initialize Project**: Set up templates and agent configuration
2. **Create Feature**: Generate feature branch and specification directory
3. **Define Specification**: Write clear requirements and acceptance criteria
4. **Plan Implementation**: Break down into actionable tasks
5. **Update Context**: Sync information to AI agent configuration
6. **Implement**: Build with AI assistant support
7. **Review and Iterate**: Refine based on feedback

Each step maintains consistency through templates and automated tooling, while giving developers and AI agents the context they need to be effective.

## Use Cases

### Greenfield Development
Start new projects with structured specifications, automated branching, and AI-ready context from day one.

### Feature Addition
Add new capabilities to existing projects with consistent workflows and clear documentation.

### Team Onboarding
Bring new team members (human or AI) up to speed quickly with standardized processes and well-maintained context.

### Multi-Agent Workflows
Work with different AI assistants on different parts of your codebase while maintaining consistent project understanding.

## Architecture

- **CLI Layer**: Cobra-based commands for all user interactions
- **Core Logic**: Feature management, template processing, git integration
- **MCP Server**: HTTP endpoints implementing Model Context Protocol
- **Agent Integration**: Automated context file management for 13+ AI agents
- **Template System**: Customizable templates for all project artifacts

## Next Steps

Ready to get started? Check out the [Quick Start Guide](quickstart.md) to create your first feature with Technocrat.

## Experimental Goals

Our research and experimentation focus on:

### Technology Independence
- Create applications using diverse technology stacks
- Validate the hypothesis that Spec-Driven Development is a process not tied to specific technologies, programming languages, or frameworks

### Enterprise Constraints
- Demonstrate mission-critical application development
- Incorporate organizational constraints (cloud providers, tech stacks, engineering practices)
- Support enterprise design systems and compliance requirements

### User-Centric Development
- Build applications for different user cohorts and preferences
- Support various development approaches (from vibe-coding to AI-native development)

### Creative & Iterative Processes
- Validate the concept of parallel implementation exploration
- Provide robust iterative feature development workflows
- Extend processes to handle upgrades and modernization tasks

## Contributing

Please see our [Contributing Guide](https://github.com/github/spec-kit/blob/main/CONTRIBUTING.md) for information on how to contribute to this project.

## Support

For support, please check our [Support Guide](https://github.com/github/spec-kit/blob/main/SUPPORT.md) or open an issue on GitHub.
