package ui

import (
	"testing"
)

func TestIsTTY(t *testing.T) {
	// Basic test that the function doesn't panic
	// We can't reliably test the actual TTY status in CI/tests
	_ = IsStdinTTY()
	_ = IsStdoutTTY()
	_ = IsStderrTTY()
	_ = IsInteractive()
}

func TestGetTerminalWidth(t *testing.T) {
	width := GetTerminalWidth()
	if width <= 0 {
		t.Errorf("Expected positive terminal width, got %d", width)
	}
	// Should return at least the default width
	if width < 80 {
		t.Errorf("Expected minimum width of 80, got %d", width)
	}
}

func TestGetTerminalSize(t *testing.T) {
	width, height := GetTerminalSize()
	if width <= 0 {
		t.Errorf("Expected positive terminal width, got %d", width)
	}
	if height <= 0 {
		t.Errorf("Expected positive terminal height, got %d", height)
	}
	// Should return at least the default dimensions
	if width < 80 {
		t.Errorf("Expected minimum width of 80, got %d", width)
	}
	if height < 24 {
		t.Errorf("Expected minimum height of 24, got %d", height)
	}
}
