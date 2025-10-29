package mcp

import (
	"os"
	"path/filepath"
	"strings"
)

// WorkspaceContext holds detected workspace information
type WorkspaceContext struct {
	Root        string // Absolute path to workspace root
	ProjectName string // Name of the project (from directory name or memory/constitution.md)
	FeatureName string // Name of current feature (if in specs/<feature>/)
}

// DetectWorkspaceContext analyzes the current working directory to extract
// project and feature context information
func DetectWorkspaceContext() WorkspaceContext {
	ctx := WorkspaceContext{}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return ctx // Return empty context on error
	}

	// Find workspace root (look for memory/ directory or .git/)
	ctx.Root = findWorkspaceRoot(cwd)
	if ctx.Root == "" {
		ctx.Root = cwd // Fallback to current directory
	}

	// Extract project name from workspace root directory name
	ctx.ProjectName = filepath.Base(ctx.Root)

	// Check if we're in a feature directory (specs/<feature>/)
	ctx.FeatureName = extractFeatureName(cwd, ctx.Root)

	// Try to get project name from constitution if it exists
	if constitutionName := getProjectNameFromConstitution(ctx.Root); constitutionName != "" {
		ctx.ProjectName = constitutionName
	}

	return ctx
}

// findWorkspaceRoot searches upward from the current directory to find
// the workspace root (indicated by memory/ directory or .git/)
func findWorkspaceRoot(startDir string) string {
	dir := startDir
	for {
		// Check for memory/ directory (technocrat project marker)
		memoryPath := filepath.Join(dir, "memory")
		if stat, err := os.Stat(memoryPath); err == nil && stat.IsDir() {
			return dir
		}

		// Check for .git/ directory (git project marker)
		gitPath := filepath.Join(dir, ".git")
		if stat, err := os.Stat(gitPath); err == nil && stat.IsDir() {
			return dir
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root of filesystem
			return ""
		}
		dir = parent
	}
}

// extractFeatureName checks if the current working directory is inside
// a specs/<feature>/ directory and returns the feature name
func extractFeatureName(cwd, workspaceRoot string) string {
	if workspaceRoot == "" {
		return ""
	}

	// Get relative path from workspace root
	relPath, err := filepath.Rel(workspaceRoot, cwd)
	if err != nil {
		return ""
	}

	// Check if path starts with specs/
	parts := strings.Split(filepath.ToSlash(relPath), "/")
	if len(parts) >= 2 && parts[0] == "specs" {
		return parts[1]
	}

	return ""
}

// getProjectNameFromConstitution attempts to read the project name
// from memory/constitution.md if it exists
func getProjectNameFromConstitution(workspaceRoot string) string {
	if workspaceRoot == "" {
		return ""
	}

	constitutionPath := filepath.Join(workspaceRoot, "memory", "constitution.md")
	content, err := os.ReadFile(constitutionPath)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(content), "\n")
	
	// First pass: Look for "Project Name" heading with value on next line
	for i, line := range lines {
		line = strings.TrimSpace(line)
		
		// Match headings like "## Project Name" (not "# My Project Name")
		if (strings.HasPrefix(line, "## Project") || 
		    strings.ToLower(line) == "## project name") {
			// Find next non-empty line that's not a heading
			for j := i + 1; j < len(lines); j++ {
				nextLine := strings.TrimSpace(lines[j])
				if nextLine != "" && !strings.HasPrefix(nextLine, "#") {
					return nextLine
				}
			}
		}
	}
	
	// Second pass: Check if the very first heading is the project name itself
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			heading := strings.TrimPrefix(line, "# ")
			// Avoid generic headings like "Constitution", "About", etc.
			lower := strings.ToLower(heading)
			if !strings.Contains(lower, "constitution") && 
			   !strings.Contains(lower, "about") &&
			   !strings.Contains(lower, "overview") {
				return heading
			}
			// Stop after first heading
			break
		}
	}

	return ""
}
