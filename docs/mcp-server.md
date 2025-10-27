# MCP Server Guide

The Technocrat MCP (Model Context Protocol) server provides a standardized HTTP API for AI agents to interact with your development workflow. This allows Claude Desktop, Cline, and other MCP-compatible tools to access Technocrat's capabilities programmatically.

## What is MCP?

The Model Context Protocol is an open standard that enables AI assistants to:
- **Discover and call tools** (functions the AI can execute)
- **Read resources** (access to project context, specs, plans)
- **Use prompts** (predefined conversation starters)

Technocrat implements MCP to bridge AI coding assistants with Spec-Driven Development workflows.

## Starting the Server

### Basic Usage

```bash
# Start on default port 8080
technocrat server

# Start on custom port
technocrat server --port 9090
```

### Expected Output

```
Starting Technocrat MCP Server on port 8080...
Server listening on :8080
```

The server runs in the foreground. Press `Ctrl+C` to stop it.

### Running in Background

```bash
# Run in background (macOS/Linux)
technocrat server &

# Run with nohup
nohup technocrat server > technocrat.log 2>&1 &

# Check if running
curl http://localhost:8080/health
```

## Server Endpoints

### Health Check

**GET** `/health`

Check if the server is running.

```bash
curl http://localhost:8080/health
```

**Response (200 OK):**
```json
{
  "status": "ok"
}
```

---

### Initialize Connection

**POST** `/mcp/v1/initialize`

Initialize an MCP client connection. Required before using other MCP endpoints.

```bash
curl -X POST http://localhost:8080/mcp/v1/initialize \
  -H "Content-Type: application/json" \
  -d '{
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "tools": {},
      "resources": {},
      "prompts": {}
    },
    "clientInfo": {
      "name": "my-client",
      "version": "1.0.0"
    }
  }'
```

**Response (200 OK):**
```json
{
  "protocolVersion": "2024-11-05",
  "capabilities": {
    "tools": {},
    "resources": {},
    "prompts": {}
  },
  "serverInfo": {
    "name": "technocrat",
    "version": "0.3.0"
  }
}
```

---

## Tools API

Tools are functions that AI agents can call to perform actions.

### List Available Tools

**GET** `/mcp/v1/tools/list`

Get a list of all available tools.

```bash
curl http://localhost:8080/mcp/v1/tools/list
```

**Response (200 OK):**
```json
{
  "tools": [
    {
      "name": "echo",
      "description": "Echo back the input text",
      "inputSchema": {
        "type": "object",
        "properties": {
          "text": {
            "type": "string",
            "description": "Text to echo"
          }
        },
        "required": ["text"]
      }
    },
    {
      "name": "list_features",
      "description": "List all features in the specs directory",
      "inputSchema": {
        "type": "object",
        "properties": {}
      }
    }
  ]
}
```

### Call a Tool

**POST** `/mcp/v1/tools/call`

Execute a tool with provided arguments.

```bash
curl -X POST http://localhost:8080/mcp/v1/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "echo",
    "arguments": {
      "text": "Hello from MCP!"
    }
  }'
```

**Response (200 OK):**
```json
{
  "content": [
    {
      "type": "text",
      "text": "Hello from MCP!"
    }
  ]
}
```

**Example: List Features**

```bash
curl -X POST http://localhost:8080/mcp/v1/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "list_features",
    "arguments": {}
  }'
```

**Response:**
```json
{
  "content": [
    {
      "type": "text",
      "text": "001-add-user-authentication\n002-add-dashboard\n003-add-reporting"
    }
  ]
}
```

---

## Resources API

Resources provide read-only access to project context and specifications.

### List Available Resources

**GET** `/mcp/v1/resources/list`

Get a list of all available resources.

```bash
curl http://localhost:8080/mcp/v1/resources/list
```

**Response (200 OK):**
```json
{
  "resources": [
    {
      "uri": "spec://001-add-user-authentication",
      "name": "Feature Spec: 001-add-user-authentication",
      "description": "Specification for feature 001-add-user-authentication",
      "mimeType": "text/markdown"
    },
    {
      "uri": "constitution://memory",
      "name": "Project Constitution",
      "description": "Project principles and guidelines",
      "mimeType": "text/markdown"
    }
  ]
}
```

### Read a Resource

**POST** `/mcp/v1/resources/read`

Read the contents of a specific resource.

```bash
curl -X POST http://localhost:8080/mcp/v1/resources/read \
  -H "Content-Type: application/json" \
  -d '{
    "uri": "spec://001-add-user-authentication"
  }'
```

**Response (200 OK):**
```json
{
  "contents": [
    {
      "uri": "spec://001-add-user-authentication",
      "mimeType": "text/markdown",
      "text": "# Feature Specification\n\n## Overview\n..."
    }
  ]
}
```

**Example: Read Constitution**

```bash
curl -X POST http://localhost:8080/mcp/v1/resources/read \
  -H "Content-Type: application/json" \
  -d '{
    "uri": "constitution://memory"
  }'
```

---

## Prompts API

Prompts are reusable templates for starting AI conversations with specific context.

### List Available Prompts

**GET** `/mcp/v1/prompts/list`

Get a list of all available prompts.

```bash
curl http://localhost:8080/mcp/v1/prompts/list
```

**Response (200 OK):**
```json
{
  "prompts": [
    {
      "name": "review_spec",
      "description": "Review a feature specification for completeness",
      "arguments": [
        {
          "name": "feature_id",
          "description": "Feature ID to review (e.g., 001-add-user-auth)",
          "required": true
        }
      ]
    },
    {
      "name": "implement_feature",
      "description": "Guide implementation of a feature",
      "arguments": [
        {
          "name": "feature_id",
          "description": "Feature ID to implement",
          "required": true
        }
      ]
    }
  ]
}
```

### Get a Prompt

**POST** `/mcp/v1/prompts/get`

Get a specific prompt with provided arguments.

```bash
curl -X POST http://localhost:8080/mcp/v1/prompts/get \
  -H "Content-Type: application/json" \
  -d '{
    "name": "review_spec",
    "arguments": {
      "feature_id": "001-add-user-authentication"
    }
  }'
```

**Response (200 OK):**
```json
{
  "messages": [
    {
      "role": "user",
      "content": {
        "type": "text",
        "text": "Please review the specification for feature 001-add-user-authentication...\n\n[spec contents included here]"
      }
    }
  ]
}
```

---

## Integration with AI Tools

### Claude Desktop

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "technocrat": {
      "command": "technocrat",
      "args": ["server", "--port", "8080"]
    }
  }
}
```

Restart Claude Desktop. You should see "technocrat" in the MCP section.

### Cline (VS Code Extension)

Add to Cline's MCP settings:

```json
{
  "mcpServers": {
    "technocrat": {
      "url": "http://localhost:8080"
    }
  }
}
```

### Custom Integration

Any HTTP client can connect to the server:

```python
import requests

# Initialize connection
response = requests.post('http://localhost:8080/mcp/v1/initialize', json={
    'protocolVersion': '2024-11-05',
    'capabilities': {'tools': {}, 'resources': {}, 'prompts': {}},
    'clientInfo': {'name': 'my-app', 'version': '1.0.0'}
})

# List available tools
tools = requests.get('http://localhost:8080/mcp/v1/tools/list').json()

# Call a tool
result = requests.post('http://localhost:8080/mcp/v1/tools/call', json={
    'name': 'echo',
    'arguments': {'text': 'Hello!'}
})
```

---

## Extending the MCP Server

The MCP handler is implemented in `internal/mcp/handler.go`.

### Adding a New Tool

1. **Define the tool schema** in `ListTools()`:

```go
{
    Name:        "my_new_tool",
    Description: "Description of what this tool does",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "param1": map[string]interface{}{
                "type":        "string",
                "description": "First parameter",
            },
        },
        "required": []string{"param1"},
    },
}
```

2. **Implement the tool logic** in `CallTool()`:

```go
case "my_new_tool":
    param1, ok := req.Arguments["param1"].(string)
    if !ok {
        return CallToolResult{
            Content: []Content{{Type: "text", Text: "param1 is required"}},
        }, fmt.Errorf("invalid arguments")
    }
    
    // Your tool logic here
    result := doSomething(param1)
    
    return CallToolResult{
        Content: []Content{{Type: "text", Text: result}},
    }, nil
```

### Adding a New Resource

1. **Define the resource** in `ListResources()`:

```go
{
    URI:         "myresource://identifier",
    Name:        "My Resource Name",
    Description: "What this resource provides",
    MimeType:    "text/plain",
}
```

2. **Implement resource reading** in `ReadResource()`:

```go
case "myresource":
    content := loadMyResource(uri)
    return ReadResourceResult{
        Contents: []ResourceContents{{
            URI:      uri,
            MimeType: "text/plain",
            Text:     content,
        }},
    }, nil
```

### Adding a New Prompt

1. **Define the prompt** in `ListPrompts()`:

```go
{
    Name:        "my_prompt",
    Description: "What this prompt helps with",
    Arguments: []PromptArgument{{
        Name:        "arg1",
        Description: "Description of argument",
        Required:    true,
    }},
}
```

2. **Implement prompt generation** in `GetPrompt()`:

```go
case "my_prompt":
    arg1 := req.Arguments["arg1"]
    text := fmt.Sprintf("Prompt template with %s", arg1)
    
    return GetPromptResult{
        Messages: []PromptMessage{{
            Role: "user",
            Content: MessageContent{
                Type: "text",
                Text: text,
            },
        }},
    }, nil
```

---

## Troubleshooting

### Server won't start

**"address already in use"**
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process or use a different port
technocrat server --port 9090
```

### Connection refused

```bash
# Verify server is running
ps aux | grep technocrat

# Check if port is accessible
curl http://localhost:8080/health
```

### Tools not appearing in Claude Desktop

1. Verify Claude Desktop config path is correct
2. Restart Claude Desktop completely
3. Check Claude Desktop logs for errors
4. Ensure `technocrat` binary is in PATH

---

## See Also

- [Command Reference](commands-reference.md) - All CLI commands
- [Agent Integration](agent-integration.md) - Configure AI agents
- [Local Development](local-development.md) - Extend the MCP server
