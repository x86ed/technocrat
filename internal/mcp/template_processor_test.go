package mcp

import (
	"strings"
	"testing"
	"time"
)

func TestProcessTemplate(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		data        TemplateData
		expected    string
		expectError bool
	}{
		{
			name:    "Simple argument substitution",
			content: "User input: {{.Arguments}}",
			data: TemplateData{
				Arguments: "test input",
			},
			expected:    "User input: test input",
			expectError: false,
		},
		{
			name:    "Empty arguments",
			content: "User input: {{.Arguments}}",
			data: TemplateData{
				Arguments: "",
			},
			expected:    "User input: ",
			expectError: false,
		},
		{
			name:    "Conditional with arguments",
			content: "{{if .Arguments}}Has input: {{.Arguments}}{{else}}No input{{end}}",
			data: TemplateData{
				Arguments: "some input",
			},
			expected:    "Has input: some input",
			expectError: false,
		},
		{
			name:    "Conditional without arguments",
			content: "{{if .Arguments}}Has input: {{.Arguments}}{{else}}No input{{end}}",
			data: TemplateData{
				Arguments: "",
			},
			expected:    "No input",
			expectError: false,
		},
		{
			name:    "Command name substitution",
			content: "Command: {{.CommandName}}",
			data: TemplateData{
				CommandName: "constitution",
			},
			expected:    "Command: constitution",
			expectError: false,
		},
		{
			name:    "Multiple substitutions",
			content: "{{.CommandName}}: {{.Arguments}}",
			data: TemplateData{
				CommandName: "spec",
				Arguments:   "feature description",
			},
			expected:    "spec: feature description",
			expectError: false,
		},
		{
			name: "Complex template with conditionals",
			content: `## User Input
{{if .Arguments}}

` + "```text" + `
{{.Arguments}}
` + "```" + `

You **MUST** consider the user input.
{{else}}
_No user input provided._
{{end}}`,
			data: TemplateData{
				Arguments: "Add security principle",
			},
			expected: `## User Input


` + "```text" + `
Add security principle
` + "```" + `

You **MUST** consider the user input.
`,
			expectError: false,
		},
		{
			name:    "Template with upper function",
			content: "Command: {{upper .CommandName}}",
			data: TemplateData{
				CommandName: "constitution",
			},
			expected:    "Command: CONSTITUTION",
			expectError: false,
		},
		{
			name:    "Template with lower function",
			content: "{{lower .Arguments}}",
			data: TemplateData{
				Arguments: "TEST INPUT",
			},
			expected:    "test input",
			expectError: false,
		},
		{
			name:    "Template with title function",
			content: "{{title .CommandName}}",
			data: TemplateData{
				CommandName: "constitution",
			},
			expected:    "Constitution",
			expectError: false,
		},
		{
			name:    "Template with trim function",
			content: "{{trim .Arguments}}",
			data: TemplateData{
				Arguments: "  spaced text  ",
			},
			expected:    "spaced text",
			expectError: false,
		},
		{
			name:        "Invalid template syntax",
			content:     "{{.Arguments",
			data:        TemplateData{},
			expected:    "",
			expectError: true,
		},
		{
			name:        "Invalid field reference",
			content:     "{{.NonExistentField}}",
			data:        TemplateData{},
			expected:    "",
			expectError: true,
		},
		{
			name: "Extra arguments access",
			content: "Feature: {{index .Extra \"feature_name\"}}",
			data: TemplateData{
				Extra: map[string]interface{}{
					"feature_name": "login system",
				},
			},
			expected:    "Feature: login system",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProcessTemplate(tt.content, tt.data)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected:\n%s\n\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestProcessTemplateWithTimestamp(t *testing.T) {
	now := time.Now()
	content := "Date: {{.Timestamp.Format \"2006-01-02\"}}"
	data := TemplateData{
		Timestamp: now,
	}

	result, err := ProcessTemplate(content, data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "Date: " + now.Format("2006-01-02")
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestPrepareTemplateContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Replace single $ARGUMENTS",
			input:    "User input: $ARGUMENTS",
			expected: "User input: {{.Arguments}}",
		},
		{
			name:     "Replace multiple $ARGUMENTS",
			input:    "$ARGUMENTS and $ARGUMENTS again",
			expected: "{{.Arguments}} and {{.Arguments}} again",
		},
		{
			name: "Replace in markdown code block",
			input: "```text\n$ARGUMENTS\n```",
			expected: "```text\n{{.Arguments}}\n```",
		},
		{
			name: "Replace in complex template",
			input: `## User Input

` + "```text" + `
$ARGUMENTS
` + "```" + `

You **MUST** consider the user input before proceeding (if not empty).`,
			expected: `## User Input

` + "```text" + `
{{.Arguments}}
` + "```" + `

You **MUST** consider the user input before proceeding (if not empty).`,
		},
		{
			name:     "No replacement needed",
			input:    "No placeholders here",
			expected: "No placeholders here",
		},
		{
			name:     "Already converted format",
			input:    "User input: {{.Arguments}}",
			expected: "User input: {{.Arguments}}",
		},
		{
			name:     "Mixed old and new format",
			input:    "$ARGUMENTS and {{.CommandName}}",
			expected: "{{.Arguments}} and {{.CommandName}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrepareTemplateContent(tt.input)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\n\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestPrepareTemplateContentIdempotent(t *testing.T) {
	// Test that preparing already-prepared content doesn't change it
	original := "User input: {{.Arguments}}"
	first := PrepareTemplateContent(original)
	second := PrepareTemplateContent(first)

	if first != second {
		t.Errorf("PrepareTemplateContent is not idempotent. First: %s, Second: %s", first, second)
	}
}

func TestProcessTemplateRealWorldExample(t *testing.T) {
	// Simulate the constitution.md template after conversion
	content := `# Constitution

## User Input

{{if .Arguments}}
` + "```text" + `
{{.Arguments}}
` + "```" + `

You **MUST** consider the user input before proceeding.
{{else}}
_No specific user input provided._
{{end}}

## Outline

You are updating the project constitution at ` + "`/memory/constitution.md`" + `.`

	tests := []struct {
		name      string
		arguments string
		contains  []string
		notContains []string
	}{
		{
			name:      "With user input",
			arguments: "Add security principle",
			contains: []string{
				"Add security principle",
				"You **MUST** consider the user input before proceeding.",
				"## Outline",
			},
			notContains: []string{
				"_No specific user input provided._",
			},
		},
		{
			name:      "Without user input",
			arguments: "",
			contains: []string{
				"_No specific user input provided._",
				"## Outline",
			},
			notContains: []string{
				"You **MUST** consider the user input before proceeding.",
				"```text",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := TemplateData{
				Arguments:   tt.arguments,
				CommandName: "constitution",
				Timestamp:   time.Now(),
			}

			result, err := ProcessTemplate(content, data)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			for _, substring := range tt.contains {
				if !strings.Contains(result, substring) {
					t.Errorf("Expected result to contain %q, but it didn't.\nResult:\n%s", substring, result)
				}
			}

			for _, substring := range tt.notContains {
				if strings.Contains(result, substring) {
					t.Errorf("Expected result NOT to contain %q, but it did.\nResult:\n%s", substring, result)
				}
			}
		})
	}
}

func TestTemplateFuncs(t *testing.T) {
	funcs := templateFuncs()

	// Test that all expected functions are registered
	expectedFuncs := []string{"upper", "lower", "title", "trim", "now"}
	for _, funcName := range expectedFuncs {
		if _, exists := funcs[funcName]; !exists {
			t.Errorf("Expected template function %q to be registered", funcName)
		}
	}
}

func BenchmarkProcessTemplate(b *testing.B) {
	content := `## User Input

{{if .Arguments}}
{{.Arguments}}
{{else}}
No input
{{end}}

Command: {{.CommandName}}`

	data := TemplateData{
		Arguments:   "test input for benchmarking",
		CommandName: "constitution",
		Timestamp:   time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ProcessTemplate(content, data)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkPrepareTemplateContent(b *testing.B) {
	content := `## User Input

` + "```text" + `
$ARGUMENTS
` + "```" + `

You **MUST** consider $ARGUMENTS before proceeding.`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = PrepareTemplateContent(content)
	}
}
