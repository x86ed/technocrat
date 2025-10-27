package ui

import (
	"os"

	"golang.org/x/term"
)

// IsTTY checks if the given file descriptor is a terminal
func IsTTY(fd uintptr) bool {
	return term.IsTerminal(int(fd))
}

// IsStdinTTY checks if stdin is a terminal (interactive)
func IsStdinTTY() bool {
	return IsTTY(os.Stdin.Fd())
}

// IsStdoutTTY checks if stdout is a terminal
func IsStdoutTTY() bool {
	return IsTTY(os.Stdout.Fd())
}

// IsStderrTTY checks if stderr is a terminal
func IsStderrTTY() bool {
	return IsTTY(os.Stderr.Fd())
}

// IsInteractive returns true if both stdin and stderr are terminals
func IsInteractive() bool {
	return IsStdinTTY() && IsStderrTTY()
}

// GetTerminalWidth returns the width of the terminal, or a default if not available
func GetTerminalWidth() int {
	if !IsStdoutTTY() {
		return 80 // Default width for non-TTY
	}

	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80 // Default width on error
	}

	if width <= 0 {
		return 80
	}

	return width
}

// GetTerminalSize returns both width and height of the terminal
func GetTerminalSize() (width, height int) {
	if !IsStdoutTTY() {
		return 80, 24 // Default dimensions
	}

	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80, 24 // Default dimensions on error
	}

	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}

	return w, h
}
