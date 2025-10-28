package mcp

import (
	"fmt"
	"strings"

	"technocrat/internal/templates"
)

// Handler manages MCP protocol operations
type Handler struct {
	tools     map[string]Tool
	resources map[string]Resource
	prompts   map[string]Prompt
}

// Tool represents an MCP tool
type Tool struct {
	Name        string                                            `json:"name"`
	Description string                                            `json:"description"`
	InputSchema map[string]interface{}                            `json:"inputSchema"`
	Handler     func(map[string]interface{}) (interface{}, error) `json:"-"`
}

// Resource represents an MCP resource
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

// Prompt represents an MCP prompt
type Prompt struct {
	Name        string                                            `json:"name"`
	Description string                                            `json:"description"`
	Arguments   []PromptArgument                                  `json:"arguments"`
	Handler     func(map[string]interface{}) (interface{}, error) `json:"-"`
}

// PromptArgument represents a prompt argument
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// NewHandler creates a new MCP handler
func NewHandler() *Handler {
	h := &Handler{
		tools:     make(map[string]Tool),
		resources: make(map[string]Resource),
		prompts:   make(map[string]Prompt),
	}

	// Register default tools
	h.registerDefaultTools()
	h.registerDefaultResources()
	h.registerDefaultPrompts()

	// Register command prompts from templates
	if err := h.RegisterCommandPrompts(); err != nil {
		// Log error but don't fail - default prompts still available
		fmt.Printf("Warning: Failed to register command prompts: %v\n", err)
	}

	return h
}

// registerDefaultTools registers the default tools
func (h *Handler) registerDefaultTools() {
	// Echo tool - simple example
	h.tools["echo"] = Tool{
		Name:        "echo",
		Description: "Echoes back the input message",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type":        "string",
					"description": "The message to echo",
				},
			},
			"required": []string{"message"},
		},
		Handler: func(args map[string]interface{}) (interface{}, error) {
			message, ok := args["message"].(string)
			if !ok {
				return nil, fmt.Errorf("message must be a string")
			}
			return map[string]interface{}{
				"echoed": message,
			}, nil
		},
	}

	// System info tool
	h.tools["system_info"] = Tool{
		Name:        "system_info",
		Description: "Returns basic system information",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"server":  "technocrat",
				"version": "1.0.0",
				"status":  "running",
			}, nil
		},
	}
}

// registerDefaultResources registers the default resources
func (h *Handler) registerDefaultResources() {
	h.resources["info://server"] = Resource{
		URI:         "info://server",
		Name:        "Server Information",
		Description: "Information about the Technocrat MCP server",
		MimeType:    "application/json",
	}
}

// registerDefaultPrompts registers the default prompts
func (h *Handler) registerDefaultPrompts() {
	h.prompts["welcome"] = Prompt{
		Name:        "welcome",
		Description: "A welcome message for new users",
		Arguments: []PromptArgument{
			{
				Name:        "name",
				Description: "User's name",
				Required:    false,
			},
		},
		Handler: func(args map[string]interface{}) (interface{}, error) {
			name := "there"
			if n, ok := args["name"].(string); ok && n != "" {
				name = n
			}
			return map[string]interface{}{
				"messages": []map[string]string{
					{
						"role":    "user",
						"content": fmt.Sprintf("Hello, %s! Welcome to Technocrat MCP Server.", name),
					},
				},
			}, nil
		},
	}
}

// ListTools returns all registered tools
func (h *Handler) ListTools() []Tool {
	tools := make([]Tool, 0, len(h.tools))
	for _, tool := range h.tools {
		// Don't include the handler in the response
		tools = append(tools, Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
		})
	}
	return tools
}

// CallTool executes a tool by name
func (h *Handler) CallTool(name string, args map[string]interface{}) (interface{}, error) {
	tool, exists := h.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return tool.Handler(args)
}

// ListResources returns all registered resources
func (h *Handler) ListResources() []Resource {
	resources := make([]Resource, 0, len(h.resources))
	for _, resource := range h.resources {
		resources = append(resources, resource)
	}
	return resources
}

// ReadResource reads a resource by URI
func (h *Handler) ReadResource(uri string) (interface{}, error) {
	resource, exists := h.resources[uri]
	if !exists {
		return nil, fmt.Errorf("resource not found: %s", uri)
	}

	// For this example, return basic info
	return map[string]interface{}{
		"uri":      resource.URI,
		"name":     resource.Name,
		"mimeType": resource.MimeType,
		"text":     "This is the Technocrat MCP server, a Spec Driven Development Framework.",
	}, nil
}

// ListPrompts returns all registered prompts
func (h *Handler) ListPrompts() []Prompt {
	prompts := make([]Prompt, 0, len(h.prompts))
	for _, prompt := range h.prompts {
		// Don't include the handler in the response
		prompts = append(prompts, Prompt{
			Name:        prompt.Name,
			Description: prompt.Description,
			Arguments:   prompt.Arguments,
		})
	}
	return prompts
}

// GetPrompt retrieves and executes a prompt by name
func (h *Handler) GetPrompt(name string, args map[string]interface{}) (interface{}, error) {
	prompt, exists := h.prompts[name]
	if !exists {
		return nil, fmt.Errorf("prompt not found: %s", name)
	}
	return prompt.Handler(args)
}

// RegisterTool registers a new tool
func (h *Handler) RegisterTool(tool Tool) {
	h.tools[tool.Name] = tool
}

// RegisterResource registers a new resource
func (h *Handler) RegisterResource(resource Resource) {
	h.resources[resource.URI] = resource
}

// RegisterPrompt registers a new prompt
func (h *Handler) RegisterPrompt(prompt Prompt) {
	h.prompts[prompt.Name] = prompt
}

// RegisterCommandPrompts registers all workflow prompts from embedded templates
func (h *Handler) RegisterCommandPrompts() error {
	// Get list of all command templates
	commands, err := templates.ListCommands()
	if err != nil {
		return fmt.Errorf("failed to list command templates: %w", err)
	}

	// Register each command as a prompt
	for _, cmdName := range commands {
		if err := h.registerCommandPrompt(cmdName); err != nil {
			// Log error but continue with other commands
			continue
		}
	}

	return nil
}

// registerCommandPrompt registers a single command prompt from template
func (h *Handler) registerCommandPrompt(commandName string) error {
	// Load template from embedded filesystem
	content, err := templates.GetCommandTemplate(commandName + ".md")
	if err != nil {
		return fmt.Errorf("failed to load template for %s: %w", commandName, err)
	}

	// Parse template to extract description and workflow
	description, workflow := parseCommandTemplate(string(content))

	// Create prompt
	prompt := Prompt{
		Name:        "tchncrt." + commandName,
		Description: description,
		Arguments: []PromptArgument{
			{
				Name:        "user_input",
				Description: "Optional user input to guide the workflow",
				Required:    false,
			},
		},
		Handler: func(args map[string]interface{}) (interface{}, error) {
			userInput := ""
			if input, ok := args["user_input"].(string); ok {
				userInput = input
			}

			// Build prompt message with workflow instructions
			message := buildPromptMessage(commandName, workflow, userInput)

			return map[string]interface{}{
				"messages": []map[string]interface{}{
					{
						"role":    "user",
						"content": message,
					},
				},
			}, nil
		},
	}

	h.RegisterPrompt(prompt)
	return nil
}

// parseCommandTemplate extracts description and workflow from command template
func parseCommandTemplate(content string) (description string, workflow string) {
	lines := strings.Split(content, "\n")

	// Extract description from YAML frontmatter
	inFrontmatter := false
	inWorkflow := false
	var workflowLines []string

	for i, line := range lines {
		// Check for frontmatter
		if i == 0 && strings.TrimSpace(line) == "---" {
			inFrontmatter = true
			continue
		}
		if inFrontmatter && strings.TrimSpace(line) == "---" {
			inFrontmatter = false
			continue
		}
		if inFrontmatter && strings.HasPrefix(line, "description:") {
			description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			description = strings.Trim(description, "\"")
			continue
		}

		// Extract workflow content (everything after frontmatter)
		if !inFrontmatter && !inWorkflow && len(strings.TrimSpace(line)) > 0 {
			inWorkflow = true
		}
		if inWorkflow {
			workflowLines = append(workflowLines, line)
		}
	}

	workflow = strings.Join(workflowLines, "\n")

	// Default description if not found
	if description == "" {
		description = "Execute workflow command"
	}

	return description, workflow
}

// buildPromptMessage constructs the prompt message for the AI
func buildPromptMessage(commandName, workflow, userInput string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Technocrat %s Workflow\n\n", strings.Title(commandName)))

	if userInput != "" {
		sb.WriteString("## User Input\n\n")
		sb.WriteString(userInput)
		sb.WriteString("\n\n")
		sb.WriteString("---\n\n")
	}

	sb.WriteString("## Workflow Instructions\n\n")
	sb.WriteString(workflow)

	return sb.String()
}
