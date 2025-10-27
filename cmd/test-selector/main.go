package main

import (
	"fmt"
	"os"

	"technocrat/internal/ui"
)

func main() {
	fmt.Println("Testing Interactive Selector")
	fmt.Println("=============================\n")

	// Test 1: AI Assistant Selection
	fmt.Println("Test 1: Select an AI assistant")
	options := map[string]string{
		"copilot":      "GitHub Copilot",
		"claude":       "Claude Code",
		"gemini":       "Gemini CLI",
		"cursor-agent": "Cursor",
	}

	selected, err := ui.SelectWithArrows("Choose your AI assistant", options, "copilot")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ You selected: %s (%s)\n\n", selected, options[selected])

	// Test 2: Script Type Selection
	fmt.Println("Test 2: Select script type")
	scriptOptions := map[string]string{
		"sh": "POSIX Shell (bash/zsh)",
		"ps": "PowerShell",
	}

	scriptSelected, err := ui.SelectWithArrows("Choose script type", scriptOptions, "sh")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ You selected: %s (%s)\n", scriptSelected, scriptOptions[scriptSelected])
}
