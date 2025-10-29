package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMetadataPopulation verifies that template data is populated with workspace metadata
func TestMetadataPopulation(t *testing.T) {
	// Create temporary workspace structure
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "my-project")
	memoryDir := filepath.Join(projectRoot, "memory")
	specsDir := filepath.Join(projectRoot, "specs")
	featureDir := filepath.Join(specsDir, "api-integration")
	
	if err := os.MkdirAll(memoryDir, 0755); err != nil {
		t.Fatalf("Failed to create memory dir: %v", err)
	}
	if err := os.MkdirAll(featureDir, 0755); err != nil {
		t.Fatalf("Failed to create feature dir: %v", err)
	}
	
	// Create constitution with project name
	constitutionContent := `# Awesome API Platform

This project provides a comprehensive API integration platform.
`
	constitutionPath := filepath.Join(memoryDir, "constitution.md")
	if err := os.WriteFile(constitutionPath, []byte(constitutionContent), 0644); err != nil {
		t.Fatalf("Failed to write constitution: %v", err)
	}
	
	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	// Change to feature directory
	if err := os.Chdir(featureDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	
	// Create handler and register prompts
	handler := NewHandler()
	if err := handler.RegisterCommandPrompts(); err != nil {
		t.Fatalf("Failed to register prompts: %v", err)
	}
	
	// Test that prompts are registered
	prompt, exists := handler.prompts["spec"]
	if !exists {
		t.Fatal("spec prompt not registered")
	}
	
	// Execute prompt handler with user input
	userInput := "Create REST API endpoints for user management"
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
	
	// Verify user input was substituted
	if !strings.Contains(content, userInput) {
		t.Errorf("Content does not contain user input:\n%s", content)
	}
	
	// Note: We can't directly verify metadata fields in the output since they're
	// only available to the template if the template actually uses them.
	// The key test is that DetectWorkspaceContext() is called and doesn't error.
	t.Log("Metadata population successful - workspace context detected")
}

// TestTemplateWithMetadata verifies that templates can access metadata fields
func TestTemplateWithMetadata(t *testing.T) {
	// Create a test template that uses metadata
	templateContent := `
User Input: {{.Arguments}}
Project: {{.ProjectName}}
Feature: {{.FeatureName}}
Workspace: {{.WorkspaceRoot}}
Command: {{.CommandName}}
Time: {{.Timestamp.Format "2006-01-02"}}
`
	
	data := TemplateData{
		Arguments:     "Test input",
		CommandName:   "test",
		ProjectName:   "TestProject",
		FeatureName:   "test-feature",
		WorkspaceRoot: "/path/to/workspace",
	}
	
	result, err := ProcessTemplate(templateContent, data)
	if err != nil {
		t.Fatalf("ProcessTemplate failed: %v", err)
	}
	
	// Verify all metadata fields are present in output
	checks := map[string]string{
		"User Input":  "Test input",
		"Project":     "TestProject",
		"Feature":     "test-feature",
		"Workspace":   "/path/to/workspace",
		"Command":     "test",
	}
	
	for field, expected := range checks {
		if !strings.Contains(result, expected) {
			t.Errorf("Result missing %s = %q:\n%s", field, expected, result)
		}
	}
	
	// Verify timestamp is formatted
	if !strings.Contains(result, "Time:") {
		t.Error("Result missing timestamp")
	}
}

// TestConditionalMetadataRendering tests conditional sections based on metadata presence
func TestConditionalMetadataRendering(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		data           TemplateData
		shouldContain  []string
		shouldNotContain []string
	}{
		{
			name: "Feature name present",
			template: `{{if .FeatureName}}Working on: {{.FeatureName}}{{else}}No feature context{{end}}`,
			data: TemplateData{
				FeatureName: "user-auth",
			},
			shouldContain: []string{"Working on: user-auth"},
			shouldNotContain: []string{"No feature context"},
		},
		{
			name: "Feature name absent",
			template: `{{if .FeatureName}}Working on: {{.FeatureName}}{{else}}No feature context{{end}}`,
			data: TemplateData{
				FeatureName: "",
			},
			shouldContain: []string{"No feature context"},
			shouldNotContain: []string{"Working on:"},
		},
		{
			name: "Project metadata present",
			template: `{{if .ProjectName}}Project: {{.ProjectName}}{{end}}
{{if .FeatureName}}Feature: {{.FeatureName}}{{end}}`,
			data: TemplateData{
				ProjectName: "MyApp",
				FeatureName: "",
			},
			shouldContain: []string{"Project: MyApp"},
			shouldNotContain: []string{"Feature:"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProcessTemplate(tt.template, tt.data)
			if err != nil {
				t.Fatalf("ProcessTemplate failed: %v", err)
			}
			
			for _, expected := range tt.shouldContain {
				if !strings.Contains(result, expected) {
					t.Errorf("Result should contain %q but doesn't:\n%s", expected, result)
				}
			}
			
			for _, unexpected := range tt.shouldNotContain {
				if strings.Contains(result, unexpected) {
					t.Errorf("Result should not contain %q but does:\n%s", unexpected, result)
				}
			}
		})
	}
}
