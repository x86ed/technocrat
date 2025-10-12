package mcp

import (
	"testing"
)

func TestNewHandler(t *testing.T) {
	handler := NewHandler()
	if handler == nil {
		t.Fatal("NewHandler() returned nil")
	}

	if handler.tools == nil {
		t.Error("tools map is nil")
	}

	if handler.resources == nil {
		t.Error("resources map is nil")
	}

	if handler.prompts == nil {
		t.Error("prompts map is nil")
	}
}

func TestListTools(t *testing.T) {
	handler := NewHandler()
	tools := handler.ListTools()

	if len(tools) == 0 {
		t.Error("Expected some default tools, got none")
	}

	// Check for default tools
	found := false
	for _, tool := range tools {
		if tool.Name == "echo" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected 'echo' tool to be registered by default")
	}
}

func TestCallTool(t *testing.T) {
	handler := NewHandler()

	// Test echo tool
	result, err := handler.CallTool("echo", map[string]interface{}{
		"message": "test message",
	})

	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	echoed, ok := resultMap["echoed"].(string)
	if !ok {
		t.Fatal("echoed field is not a string")
	}

	if echoed != "test message" {
		t.Errorf("Expected 'test message', got '%s'", echoed)
	}

	// Test non-existent tool
	_, err = handler.CallTool("nonexistent", map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for non-existent tool")
	}
}

func TestListResources(t *testing.T) {
	handler := NewHandler()
	resources := handler.ListResources()

	if len(resources) == 0 {
		t.Error("Expected some default resources, got none")
	}
}

func TestListPrompts(t *testing.T) {
	handler := NewHandler()
	prompts := handler.ListPrompts()

	if len(prompts) == 0 {
		t.Error("Expected some default prompts, got none")
	}

	// Check for welcome prompt
	found := false
	for _, prompt := range prompts {
		if prompt.Name == "welcome" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected 'welcome' prompt to be registered by default")
	}
}
