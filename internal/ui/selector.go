package ui

import (
	"fmt"
	"os"

	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	"github.com/pterm/pterm"
)

// SelectOption represents an option in a selection menu
type SelectOption struct {
	Key         string
	Description string
}

// Selector provides an interactive arrow-key selection interface
type Selector struct {
	Prompt       string
	Options      []SelectOption
	DefaultKey   string
	selectedIdx  int
	cancelText   string
	instructions string
}

// NewSelector creates a new interactive selector
func NewSelector(prompt string, options []SelectOption, defaultKey string) *Selector {
	selectedIdx := 0

	// Find the default option index
	if defaultKey != "" {
		for i, opt := range options {
			if opt.Key == defaultKey {
				selectedIdx = i
				break
			}
		}
	}

	return &Selector{
		Prompt:       prompt,
		Options:      options,
		DefaultKey:   defaultKey,
		selectedIdx:  selectedIdx,
		cancelText:   "Selection cancelled",
		instructions: "Use ↑/↓ arrows to navigate, Enter to select, Esc to cancel",
	}
}

// SetInstructions sets custom instructions text
func (s *Selector) SetInstructions(text string) *Selector {
	s.instructions = text
	return s
}

// SetCancelText sets custom cancellation message
func (s *Selector) SetCancelText(text string) *Selector {
	s.cancelText = text
	return s
}

// render creates the current selection display
func (s *Selector) render() string {
	// Create a bordered box with title
	box := pterm.DefaultBox.
		WithTitle(ColorCyan.Sprint(s.Prompt)).
		WithTitleTopCenter().
		WithBoxStyle(pterm.NewStyle(pterm.FgCyan))

	// Build the options list
	var lines []string
	for i, opt := range s.Options {
		var line string
		if i == s.selectedIdx {
			// Highlight selected option
			line = fmt.Sprintf("  %s  %s %s",
				ColorCyan.Sprint(SymbolPointer),
				ColorCyan.Sprint(opt.Key),
				ColorDim.Sprintf("(%s)", opt.Description))
		} else {
			// Regular option
			line = fmt.Sprintf("     %s %s",
				ColorCyan.Sprint(opt.Key),
				ColorDim.Sprintf("(%s)", opt.Description))
		}
		lines = append(lines, line)
	}

	// Add instructions
	lines = append(lines, "")
	lines = append(lines, ColorDim.Sprint(s.instructions))

	content := ""
	for _, line := range lines {
		content += line + "\n"
	}

	return box.Sprint(content)
}

// Select runs the interactive selection and returns the selected key
func (s *Selector) Select() (string, error) {
	// Check if we're in a TTY
	if !IsInteractive() {
		return "", fmt.Errorf("interactive selection requires a TTY")
	}

	// Print initial render
	fmt.Fprintln(os.Stderr)
	area, err := pterm.DefaultArea.Start()
	if err != nil {
		return "", fmt.Errorf("failed to start area: %w", err)
	}
	area.Update(s.render())

	// Listen for keyboard input
	err = keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		switch key.Code {
		case keys.Up, keys.CtrlP:
			// Move selection up
			s.selectedIdx--
			if s.selectedIdx < 0 {
				s.selectedIdx = len(s.Options) - 1
			}
			area.Update(s.render())
			return false, nil

		case keys.Down, keys.CtrlN:
			// Move selection down
			s.selectedIdx++
			if s.selectedIdx >= len(s.Options) {
				s.selectedIdx = 0
			}
			area.Update(s.render())
			return false, nil

		case keys.Enter:
			// Select current option
			area.Stop()
			return true, nil

		case keys.Escape, keys.CtrlC:
			// Cancel selection
			area.Stop()
			fmt.Fprintln(os.Stderr, ColorYellow.Sprint(s.cancelText))
			return true, fmt.Errorf("selection cancelled by user")
		}

		return false, nil
	})

	if err != nil {
		return "", err
	}

	return s.Options[s.selectedIdx].Key, nil
}

// SelectWithArrows provides a convenient wrapper for simple selections
func SelectWithArrows(prompt string, options map[string]string, defaultKey string) (string, error) {
	// Convert map to SelectOptions
	selectOpts := make([]SelectOption, 0, len(options))
	for key, desc := range options {
		selectOpts = append(selectOpts, SelectOption{
			Key:         key,
			Description: desc,
		})
	}

	selector := NewSelector(prompt, selectOpts, defaultKey)
	return selector.Select()
}
