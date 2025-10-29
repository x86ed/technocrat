package mcp

import (
	"strings"
	"testing"
)

// TestTemplateSubstitutionIntegration demonstrates the full workflow
// of user input being substituted into command templates
func TestTemplateSubstitutionIntegration(t *testing.T) {
	tests := []struct {
		name           string
		promptName     string
		userInput      string
		shouldContain  []string
		shouldNotContain []string
	}{
		{
			name:       "Constitution with user input",
			promptName: "constitution",
			userInput:  "Add security and privacy principles",
			shouldContain: []string{
				"Constitution Workflow",
				"Add security and privacy principles",
				"You **MUST** consider the user input before proceeding",
				"You are updating the project constitution",
			},
			shouldNotContain: []string{
				"$ARGUMENTS",
				"_No specific user input provided",
			},
		},
		{
			name:       "Constitution without user input",
			promptName: "constitution",
			userInput:  "",
			shouldContain: []string{
				"Constitution Workflow",
				"_No specific user input provided",
				"You are updating the project constitution",
			},
			shouldNotContain: []string{
				"$ARGUMENTS",
				"You **MUST** consider the user input before proceeding",
			},
		},
		{
			name:       "Spec with feature description",
			promptName: "spec",
			userInput:  "Create a user authentication system with OAuth2",
			shouldContain: []string{
				"Spec Workflow",
				"Create a user authentication system with OAuth2",
				"You **MUST** consider the user input",
			},
			shouldNotContain: []string{
				"$ARGUMENTS",
				"_No specific feature description provided",
			},
		},
		{
			name:       "Plan with guidance",
			promptName: "plan",
			userInput:  "Focus on database schema design",
			shouldContain: []string{
				"Plan Workflow",
				"Focus on database schema design",
			},
			shouldNotContain: []string{
				"$ARGUMENTS",
			},
		},
		{
			name:       "Analyze without specific focus",
			promptName: "analyze",
			userInput:  "",
			shouldContain: []string{
				"Analyze Workflow",
				"_No specific analysis focus provided",
			},
			shouldNotContain: []string{
				"$ARGUMENTS",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler()

			args := map[string]interface{}{}
			if tt.userInput != "" {
				args["user_input"] = tt.userInput
			}

			result, err := handler.GetPrompt(tt.promptName, args)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			resultMap, ok := result.(map[string]interface{})
			if !ok {
				t.Fatal("Result is not a map")
			}

			messages, ok := resultMap["messages"].([]map[string]interface{})
			if !ok || len(messages) == 0 {
				t.Fatal("Invalid messages structure")
			}

			content, ok := messages[0]["content"].(string)
			if !ok {
				t.Fatal("Content is not a string")
			}

			// Check that expected strings are present
			for _, expected := range tt.shouldContain {
				if !strings.Contains(content, expected) {
					t.Errorf("Expected content to contain %q, but it didn't.\nContent:\n%s",
						expected, content)
				}
			}

			// Check that unwanted strings are not present
			for _, notExpected := range tt.shouldNotContain {
				if strings.Contains(content, notExpected) {
					t.Errorf("Expected content NOT to contain %q, but it did.\nContent:\n%s",
						notExpected, content)
				}
			}

			// Verify no raw template syntax leaked through
			if strings.Contains(content, "$ARGUMENTS") {
				t.Error("Template syntax $ARGUMENTS was not replaced!")
			}
		})
	}
}

// TestBackwardCompatibility ensures old $ARGUMENTS syntax still works
func TestBackwardCompatibility(t *testing.T) {
	// Test that PrepareTemplateContent converts legacy format
	legacyTemplate := "User input: $ARGUMENTS\n\nMore content with $ARGUMENTS"
	converted := PrepareTemplateContent(legacyTemplate)

	if strings.Contains(converted, "$ARGUMENTS") {
		t.Error("$ARGUMENTS was not converted to {{.Arguments}}")
	}

	if !strings.Contains(converted, "{{.Arguments}}") {
		t.Error("Template was not converted to {{.Arguments}}")
	}

	// Verify it processes correctly
	data := TemplateData{
		Arguments: "test input",
	}

	result, err := ProcessTemplate(converted, data)
	if err != nil {
		t.Fatalf("Failed to process converted template: %v", err)
	}

	if !strings.Contains(result, "test input") {
		t.Error("Converted template did not substitute arguments correctly")
	}
}

// TestConditionalRendering verifies if/else logic works correctly
func TestConditionalRendering(t *testing.T) {
	handler := NewHandler()

	// Test with input
	resultWithInput, err := handler.GetPrompt("spec", map[string]interface{}{
		"user_input": "test feature",
	})
	if err != nil {
		t.Fatalf("Error with input: %v", err)
	}

	contentWithInput := resultWithInput.(map[string]interface{})["messages"].([]map[string]interface{})[0]["content"].(string)

	// Test without input
	resultWithoutInput, err := handler.GetPrompt("spec", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Error without input: %v", err)
	}

	contentWithoutInput := resultWithoutInput.(map[string]interface{})["messages"].([]map[string]interface{})[0]["content"].(string)

	// They should be different
	if contentWithInput == contentWithoutInput {
		t.Error("Content should differ based on presence of user input")
	}

	// With input should have the actual input
	if !strings.Contains(contentWithInput, "test feature") {
		t.Error("Content with input should contain the actual input text")
	}

	// Without input should have fallback message
	if !strings.Contains(contentWithoutInput, "_No specific") {
		t.Error("Content without input should contain fallback message")
	}
}
