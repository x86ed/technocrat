package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
)

// BorderStyle defines the type of border to use for panels
type BorderStyle string

const (
	BorderStyleCyan    BorderStyle = "cyan"
	BorderStyleRed     BorderStyle = "red"
	BorderStyleYellow  BorderStyle = "yellow"
	BorderStyleGreen   BorderStyle = "green"
	BorderStyleMagenta BorderStyle = "magenta"
)

// Panel represents a bordered text panel with a title
type Panel struct {
	Title       string
	Content     string
	BorderStyle BorderStyle
	Padding     int
}

// RenderPanel creates a bordered panel with a title and content
func RenderPanel(title, content string, borderStyle BorderStyle) string {
	return RenderPanelWithPadding(title, content, borderStyle, 1)
}

// RenderPanelWithPadding creates a bordered panel with custom padding
func RenderPanelWithPadding(title, content string, borderStyle BorderStyle, padding int) string {
	// Get color based on border style
	var titleColor pterm.Style
	var borderColor pterm.Color

	switch borderStyle {
	case BorderStyleCyan:
		titleColor = *ColorCyan
		borderColor = pterm.FgCyan
	case BorderStyleRed:
		titleColor = *ColorRed
		borderColor = pterm.FgRed
	case BorderStyleYellow:
		titleColor = *ColorYellow
		borderColor = pterm.FgYellow
	case BorderStyleGreen:
		titleColor = *ColorGreen
		borderColor = pterm.FgGreen
	case BorderStyleMagenta:
		titleColor = *ColorMagenta
		borderColor = pterm.FgMagenta
	default:
		titleColor = *ColorCyan
		borderColor = pterm.FgCyan
	}

	// Create the panel using pterm
	panel := pterm.DefaultBox.
		WithTitle(titleColor.Sprint(title)).
		WithTitleTopCenter().
		WithBoxStyle(pterm.NewStyle(borderColor))

	// Add padding by adding blank lines
	paddedContent := content
	if padding > 0 {
		paddingStr := strings.Repeat("\n", padding-1)
		paddedContent = paddingStr + content + paddingStr
	}

	return panel.Sprint(paddedContent)
}

// ShowPanel prints a panel to stderr
func ShowPanel(title, content string, borderStyle BorderStyle) {
	fmt.Fprintln(os.Stderr, RenderPanel(title, content, borderStyle))
}

// ShowError displays an error message in a red panel
func ShowError(title, message string) {
	ShowPanel(title, message, BorderStyleRed)
}

// ShowWarning displays a warning message in a yellow panel
func ShowWarning(title, message string) {
	ShowPanel(title, message, BorderStyleYellow)
}

// ShowInfo displays an informational message in a cyan panel
func ShowInfo(title, message string) {
	ShowPanel(title, message, BorderStyleCyan)
}

// ShowSuccess displays a success message in a green panel
func ShowSuccess(title, message string) {
	ShowPanel(title, message, BorderStyleGreen)
}

// FormatKeyValue formats a key-value pair with proper spacing
func FormatKeyValue(key, value string, keyWidth int) string {
	return fmt.Sprintf("%-*s  %s", keyWidth, key, value)
}

// CreateTable creates a simple aligned table from key-value pairs
func CreateTable(pairs map[string]string, keyWidth int) string {
	var lines []string
	for key, value := range pairs {
		lines = append(lines, FormatKeyValue(key, value, keyWidth))
	}
	return strings.Join(lines, "\n")
}
