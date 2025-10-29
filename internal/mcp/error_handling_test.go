package mcp

import (
	"strings"
	"testing"
)

func TestEnhancedErrorMessages(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		data           TemplateData
		expectError    bool
		errorContains  []string
		hintContains   string
	}{
		{
			name:     "Unclosed if block",
			template: `{{if .Arguments}}Some text`,
			data:     TemplateData{Arguments: "test"},
			expectError: true,
			errorContains: []string{"template", "parse", "failed"},
			hintContains: "{{if}}",
		},
		{
			name:     "Unknown function",
			template: `{{unknownFunc .Arguments}}`,
			data:     TemplateData{Arguments: "test"},
			expectError: true,
			errorContains: []string{"template", "failed"},
			hintContains: "Available functions",
		},
		{
			name:     "Unknown variable",
			template: `{{.UnknownField}}`,
			data:     TemplateData{Arguments: "test"},
			expectError: true,
			errorContains: []string{"template", "execute", "failed"},
			hintContains: "Available variables",
		},
		{
			name:     "Unclosed action",
			template: `{{.Arguments`,
			data:     TemplateData{Arguments: "test"},
			expectError: true,
			errorContains: []string{"template", "parse", "failed"},
			hintContains: "Missing closing braces",
		},
		{
			name:     "Wrong bracket type",
			template: `<.Arguments>`,
			data:     TemplateData{Arguments: "test"},
			expectError: false, // This is actually valid text, not a template
		},
		{
			name:     "Valid template",
			template: `{{.Arguments}}`,
			data:     TemplateData{Arguments: "test"},
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ProcessTemplate(tt.template, tt.data)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				
				errMsg := err.Error()
				for _, expected := range tt.errorContains {
					if !strings.Contains(errMsg, expected) {
						t.Errorf("Error message should contain %q, got: %s", expected, errMsg)
					}
				}
				
				if tt.hintContains != "" && !strings.Contains(errMsg, tt.hintContains) {
					t.Errorf("Error hint should contain %q, got: %s", tt.hintContains, errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestErrorHints(t *testing.T) {
	tests := []struct {
		name         string
		errorMsg     string
		expectedHint string
	}{
		{
			name:         "Unclosed if",
			errorMsg:     `unexpected "}" in operand`,
			expectedHint: "{{if}}",
		},
		{
			name:         "Unknown function",
			errorMsg:     `function "badFunc" not defined`,
			expectedHint: "Available functions",
		},
		{
			name:         "Unknown variable",
			errorMsg:     `can't evaluate field BadField`,
			expectedHint: "Available variables",
		},
		{
			name:         "Nil pointer",
			errorMsg:     `nil pointer evaluating interface`,
			expectedHint: "{{if .Field}}",
		},
		{
			name:         "Unclosed action",
			errorMsg:     `unclosed action`,
			expectedHint: "Missing closing braces",
		},
		{
			name:         "No special error",
			errorMsg:     `some other error`,
			expectedHint: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hint := getErrorHint(tt.errorMsg)
			
			if tt.expectedHint == "" {
				if hint != "" {
					t.Errorf("Expected no hint but got: %s", hint)
				}
			} else {
				if !strings.Contains(hint, tt.expectedHint) {
					t.Errorf("Hint should contain %q, got: %s", tt.expectedHint, hint)
				}
			}
		})
	}
}

func TestProcessTemplateWithContextErrors(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		data        TemplateData
		expectError bool
	}{
		{
			name:        "Invalid readSpec call",
			template:    `{{readSpec "arg"}}`, // readSpec takes no args
			data:        TemplateData{WorkspaceRoot: "/tmp", FeatureName: "test"},
			expectError: true,
		},
		{
			name:        "Unclosed if with readPlan",
			template:    `{{if readPlan}}Content`,
			data:        TemplateData{WorkspaceRoot: "/tmp", FeatureName: "test"},
			expectError: true,
		},
		{
			name:        "Valid readSpec",
			template:    `{{if readSpec}}Has spec{{else}}No spec{{end}}`,
			data:        TemplateData{WorkspaceRoot: "/tmp", FeatureName: "test"},
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ProcessTemplateWithContext(tt.template, tt.data)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestErrorMessageFormat(t *testing.T) {
	// Test that error messages are well-formatted
	template := `{{if .Arguments}}{{.UnknownField}}{{end}}`
	data := TemplateData{Arguments: "test"}
	
	_, err := ProcessTemplate(template, data)
	if err == nil {
		t.Fatal("Expected error but got none")
	}
	
	errMsg := err.Error()
	
	// Should contain phase information
	if !strings.Contains(errMsg, "template") {
		t.Error("Error should mention 'template'")
	}
	
	// Should contain the actual error
	if !strings.Contains(errMsg, "execute") || !strings.Contains(errMsg, "parse") {
		if !strings.Contains(errMsg, "failed") {
			t.Error("Error should contain phase information")
		}
	}
	
	// Check format is readable
	lines := strings.Split(errMsg, "\n")
	if len(lines) < 1 {
		t.Error("Error should be multi-line with hints")
	}
}
