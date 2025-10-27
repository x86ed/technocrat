package ui

import "github.com/pterm/pterm"

// Color scheme constants for consistent styling across the application
var (
	// Primary colors for different message types
	ColorCyan    = pterm.NewStyle(pterm.FgCyan)
	ColorGreen   = pterm.NewStyle(pterm.FgGreen)
	ColorYellow  = pterm.NewStyle(pterm.FgYellow)
	ColorRed     = pterm.NewStyle(pterm.FgRed)
	ColorMagenta = pterm.NewStyle(pterm.FgMagenta)
	ColorWhite   = pterm.NewStyle(pterm.FgWhite)

	// Bright variants
	ColorBrightBlue  = pterm.NewStyle(pterm.FgLightBlue)
	ColorBrightCyan  = pterm.NewStyle(pterm.FgLightCyan)
	ColorBrightWhite = pterm.NewStyle(pterm.FgLightWhite)

	// Dimmed colors (using gray)
	ColorDim = pterm.NewStyle(pterm.FgGray)

	// Text styles
	StyleBold = pterm.NewStyle(pterm.Bold)

	// Banner colors - cycling through these for the ASCII art
	BannerColors = []pterm.Style{
		*ColorBrightBlue,
		*pterm.NewStyle(pterm.FgBlue),
		*ColorCyan,
		*ColorBrightCyan,
		*ColorWhite,
		*ColorBrightWhite,
	}
)

// Status symbols using Unicode characters
const (
	SymbolCheckmark    = "✓"
	SymbolSuccess      = "✓"
	SymbolError        = "✗"
	SymbolWarning      = "⚠"
	SymbolArrow        = "→"
	SymbolCircle       = "○"
	SymbolFilledCircle = "●"
	SymbolPointer      = "▶"
)
