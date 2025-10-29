package mcp

import (
	"strings"
	"testing"
)

// TestConstitutionPromptAcceptsArguments verifies constitution prompt handles user input
func TestConstitutionPromptAcceptsArguments(t *testing.T) {
	handler := NewHandler()
	if err := handler.RegisterCommandPrompts(); err != nil {
		t.Fatalf("Failed to register prompts: %v", err)
	}
	
	// Get constitution prompt
	prompt, exists := handler.prompts["constitution"]
	if !exists {
		t.Fatal("constitution prompt not registered")
	}
	
	// Verify it has user_input argument
	if len(prompt.Arguments) == 0 {
		t.Fatal("constitution prompt has no arguments")
	}
	
	found := false
	for _, arg := range prompt.Arguments {
		if arg.Name == "user_input" {
			found = true
			if arg.Required {
				t.Error("user_input should not be required")
			}
		}
	}
	
	if !found {
		t.Error("constitution prompt missing user_input argument")
	}
	
	// Test with user input
	userInput := "Focus on security and testing principles"
	result, err := prompt.Handler(map[string]interface{}{
		"user_input": userInput,
	})
	if err != nil {
		t.Fatalf("Prompt handler failed: %v", err)
	}
	
	// Extract message content
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}
	
	messages, ok := resultMap["messages"].([]map[string]interface{})
	if !ok || len(messages) == 0 {
		t.Fatal("No messages in result")
	}
	
	content, ok := messages[0]["content"].(string)
	if !ok {
		t.Fatal("Message content is not a string")
	}
	
	// Verify user input appears in the prompt
	if !strings.Contains(content, userInput) {
		t.Errorf("Constitution prompt should contain user input %q, but got:\n%s", userInput, content)
	}
	
	// Verify it doesn't show the "no input" fallback
	if strings.Contains(content, "No specific user input provided") {
		t.Error("Should not show 'no input' message when input is provided")
	}
}

// TestConstitutionPromptWithoutArguments verifies fallback message
func TestConstitutionPromptWithoutArguments(t *testing.T) {
	handler := NewHandler()
	if err := handler.RegisterCommandPrompts(); err != nil {
		t.Fatalf("Failed to register prompts: %v", err)
	}
	
	prompt := handler.prompts["constitution"]
	
	// Test without user input
	result, err := prompt.Handler(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Prompt handler failed: %v", err)
	}
	
	resultMap := result.(map[string]interface{})
	messages := resultMap["messages"].([]map[string]interface{})
	content := messages[0]["content"].(string)
	
	// Should show the fallback message
	if !strings.Contains(content, "No specific user input provided") {
		t.Error("Should show 'no input' fallback message when no input provided")
	}
}

// TestAllPromptsAcceptArguments verifies all prompts accept user_input
func TestAllPromptsAcceptArguments(t *testing.T) {
	handler := NewHandler()
	if err := handler.RegisterCommandPrompts(); err != nil {
		t.Fatalf("Failed to register prompts: %v", err)
	}
	
	expectedPrompts := []string{
		"constitution",
		"spec",
		"plan",
		"tasks",
		"implement",
		"analyze",
		"clarify",
		"checklist",
	}
	
	for _, promptName := range expectedPrompts {
		t.Run(promptName, func(t *testing.T) {
			prompt, exists := handler.prompts[promptName]
			if !exists {
				t.Fatalf("Prompt %s not registered", promptName)
			}
			
			// Check for user_input argument
			hasUserInput := false
			for _, arg := range prompt.Arguments {
				if arg.Name == "user_input" {
					hasUserInput = true
					break
				}
			}
			
			if !hasUserInput {
				t.Errorf("Prompt %s should have user_input argument", promptName)
			}
			
			// Test that it accepts input
			result, err := prompt.Handler(map[string]interface{}{
				"user_input": "test input",
			})
			if err != nil {
				t.Errorf("Prompt %s failed with user input: %v", promptName, err)
			}
			
			if result == nil {
				t.Errorf("Prompt %s returned nil result", promptName)
			}
		})
	}
}
