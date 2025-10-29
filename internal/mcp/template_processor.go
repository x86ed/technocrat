package mcp

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// TemplateData holds all data available for template substitution
type TemplateData struct {
	Arguments     string                 // User input text
	CommandName   string                 // Name of the command being executed
	Timestamp     time.Time              // Current timestamp
	ProjectName   string                 // Name of the current project
	FeatureName   string                 // Name of the current feature (if in specs/<feature>/)
	WorkspaceRoot string                 // Absolute path to workspace root
	Extra         map[string]interface{} // Additional arguments from the request
}

// ProcessTemplate executes Go template substitution on workflow content
func ProcessTemplate(workflowContent string, data TemplateData) (string, error) {
	// Create template with custom functions
	tmpl, err := template.New("workflow").
		Funcs(templateFuncs()).
		Parse(workflowContent)
	if err != nil {
		return "", enhanceTemplateError("parse", err, workflowContent)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", enhanceTemplateError("execute", err, workflowContent)
	}

	return buf.String(), nil
}

// PrepareTemplateContent converts legacy $ARGUMENTS to {{.Arguments}} format
// This maintains backward compatibility while enabling template features
func PrepareTemplateContent(content string) string {
	// Replace $ARGUMENTS with {{.Arguments}}
	content = strings.ReplaceAll(content, "$ARGUMENTS", "{{.Arguments}}")
	return content
}

// ReadFeatureFile reads a file from the current feature directory
// Returns empty string if file doesn't exist or can't be read
func ReadFeatureFile(workspaceRoot, featureName, filename string) string {
	if workspaceRoot == "" || featureName == "" {
		return ""
	}

	filePath := filepath.Join(workspaceRoot, "specs", featureName, filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}

	return string(content)
}

// templateFuncs returns custom template functions available in all templates
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": strings.Title,
		"trim":  strings.TrimSpace,
		"now":   time.Now,
	}
}

// makeTemplateFuncsWithContext creates template functions that have access to workspace context
func makeTemplateFuncsWithContext(workspaceRoot, featureName string) template.FuncMap {
	funcs := templateFuncs()
	
	// Add feature file reading functions
	funcs["readSpec"] = func() string {
		return ReadFeatureFile(workspaceRoot, featureName, "spec.md")
	}
	funcs["readPlan"] = func() string {
		return ReadFeatureFile(workspaceRoot, featureName, "plan.md")
	}
	funcs["readTasks"] = func() string {
		return ReadFeatureFile(workspaceRoot, featureName, "tasks.md")
	}
	funcs["readFile"] = func(filename string) string {
		return ReadFeatureFile(workspaceRoot, featureName, filename)
	}
	
	return funcs
}

// ProcessTemplateWithContext executes Go template substitution with workspace context
// This variant provides additional functions for reading feature files
func ProcessTemplateWithContext(workflowContent string, data TemplateData) (string, error) {
	// Create template with context-aware functions
	funcs := makeTemplateFuncsWithContext(data.WorkspaceRoot, data.FeatureName)
	
	tmpl, err := template.New("workflow").
		Funcs(funcs).
		Parse(workflowContent)
	if err != nil {
		return "", enhanceTemplateError("parse", err, workflowContent)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", enhanceTemplateError("execute", err, workflowContent)
	}

	return buf.String(), nil
}

// enhanceTemplateError provides more context for template errors
func enhanceTemplateError(phase string, err error, template string) error {
	errMsg := err.Error()
	
	// Extract line number if present
	lineInfo := ""
	if strings.Contains(errMsg, ":") {
		parts := strings.Split(errMsg, ":")
		if len(parts) >= 2 {
			lineInfo = " at " + parts[1]
		}
	}
	
	// Add helpful hints based on error type
	hint := getErrorHint(errMsg)
	
	baseMsg := fmt.Sprintf("template %s failed%s: %v", phase, lineInfo, err)
	if hint != "" {
		return fmt.Errorf("%s\n\nHint: %s", baseMsg, hint)
	}
	
	return fmt.Errorf("%s", baseMsg)
}

// getErrorHint provides helpful suggestions based on error patterns
func getErrorHint(errMsg string) string {
	switch {
	case strings.Contains(errMsg, "unexpected EOF"):
		return "Check for unclosed {{if}} or {{range}} blocks. Every {{if}} needs a {{end}}."
	case strings.Contains(errMsg, "unexpected \"}\""):
		return "Check for unclosed {{if}} or {{range}} blocks. Every {{if}} needs a {{end}}."
	case strings.Contains(errMsg, "unexpected \"<\""):
		return "Template syntax error. Make sure you're using {{.Variable}} not <.Variable>."
	case strings.Contains(errMsg, "function") && strings.Contains(errMsg, "not defined"):
		return "Unknown function. Available functions: upper, lower, title, trim, now, readSpec, readPlan, readTasks, readFile."
	case strings.Contains(errMsg, "can't evaluate field"):
		return "Unknown variable. Available variables: .Arguments, .CommandName, .Timestamp, .ProjectName, .FeatureName, .WorkspaceRoot."
	case strings.Contains(errMsg, "nil pointer"):
		return "Trying to access a field that doesn't exist. Use {{if .Field}} to check before accessing."
	case strings.Contains(errMsg, "unexpected \"(\""):
		return "Function call syntax error. Use {{functionName}} or {{functionName \"arg\"}}."
	case strings.Contains(errMsg, "unclosed action"):
		return "Missing closing braces. Make sure every {{ has a matching }}."
	default:
		return ""
	}
}
