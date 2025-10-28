package mcp

import (
	"testing"
)

func TestRegisterCommandPrompts(t *testing.T) {
	h := NewHandler()

	// Check that command prompts were registered
	expectedPrompts := []string{
		"tchncrt.spec",
		"tchncrt.plan",
		"tchncrt.tasks",
		"tchncrt.implement",
		"tchncrt.constitution",
		"tchncrt.checklist",
		"tchncrt.clarify",
		"tchncrt.analyze",
	}

	prompts := h.ListPrompts()
	promptMap := make(map[string]bool)
	for _, p := range prompts {
		promptMap[p.Name] = true
	}

	for _, expected := range expectedPrompts {
		if !promptMap[expected] {
			t.Errorf("Expected prompt %s to be registered, but it was not found", expected)
		}
	}
}

func TestPromptExecution(t *testing.T) {
	h := NewHandler()

	tests := []struct {
		name       string
		promptName string
		args       map[string]interface{}
		wantError  bool
	}{
		{
			name:       "spec prompt with user input",
			promptName: "tchncrt.spec",
			args: map[string]interface{}{
				"user_input": "Create a login feature with OAuth support",
			},
			wantError: false,
		},
		{
			name:       "spec prompt without user input",
			promptName: "tchncrt.spec",
			args:       map[string]interface{}{},
			wantError:  false,
		},
		{
			name:       "plan prompt",
			promptName: "tchncrt.plan",
			args:       map[string]interface{}{},
			wantError:  false,
		},
		{
			name:       "nonexistent prompt",
			promptName: "tchncrt.nonexistent",
			args:       map[string]interface{}{},
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.GetPrompt(tt.promptName, tt.args)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify result structure
			resultMap, ok := result.(map[string]interface{})
			if !ok {
				t.Error("Result is not a map")
				return
			}

			messages, ok := resultMap["messages"].([]map[string]interface{})
			if !ok {
				t.Error("messages field is not a slice of maps")
				return
			}

			if len(messages) == 0 {
				t.Error("messages slice is empty")
				return
			}

			// Verify message structure
			firstMsg := messages[0]
			if _, ok := firstMsg["role"]; !ok {
				t.Error("message missing role field")
			}
			if content, ok := firstMsg["content"].(string); !ok || content == "" {
				t.Error("message missing or empty content field")
			}
		})
	}
}

func TestPromptContentStructure(t *testing.T) {
	h := NewHandler()

	result, err := h.GetPrompt("tchncrt.spec", map[string]interface{}{
		"user_input": "Add user authentication",
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap := result.(map[string]interface{})
	messages := resultMap["messages"].([]map[string]interface{})
	content := messages[0]["content"].(string)

	// Verify content includes expected sections
	expectedSections := []string{
		"Technocrat",
		"Workflow",
		"User Input",
		"Add user authentication",
	}

	for _, section := range expectedSections {
		if !contains(content, section) {
			t.Errorf("Expected content to contain '%s', but it doesn't", section)
		}
	}
}

func TestParseCommandTemplate(t *testing.T) {
	testTemplate := `---
description: Test command description
---

# Test Command

## Outline

This is the workflow content.
It should be extracted properly.

## Another Section

More content here.
`

	description, workflow := parseCommandTemplate(testTemplate)

	if description != "Test command description" {
		t.Errorf("Expected description 'Test command description', got '%s'", description)
	}

	if workflow == "" {
		t.Error("Workflow should not be empty")
	}

	if !contains(workflow, "Test Command") {
		t.Error("Workflow should contain 'Test Command'")
	}
}

func TestBuildPromptMessage(t *testing.T) {
	workflow := "Step 1: Do something\nStep 2: Do something else"

	tests := []struct {
		name        string
		commandName string
		userInput   string
		wantContain []string
	}{
		{
			name:        "with user input",
			commandName: "spec",
			userInput:   "Create login feature",
			wantContain: []string{
				"Technocrat",
				"Spec",
				"User Input",
				"Create login feature",
				"Workflow Instructions",
				"Step 1",
			},
		},
		{
			name:        "without user input",
			commandName: "plan",
			userInput:   "",
			wantContain: []string{
				"Technocrat",
				"Plan",
				"Workflow Instructions",
				"Step 1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := buildPromptMessage(tt.commandName, workflow, tt.userInput)

			for _, want := range tt.wantContain {
				if !contains(message, want) {
					t.Errorf("Expected message to contain '%s', but it doesn't", want)
				}
			}

			// If no user input, should not contain "User Input" section
			if tt.userInput == "" && contains(message, "## User Input") {
				t.Error("Message should not contain User Input section when no input provided")
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
