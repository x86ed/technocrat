package ui

import (
	"strings"
	"testing"
)

func TestRenderPanel(t *testing.T) {
	title := "Test Panel"
	content := "This is test content"
	
	result := RenderPanel(title, content, BorderStyleCyan)
	
	if result == "" {
		t.Error("Expected non-empty panel output")
	}
	
	// Check that title and content appear in output
	if !strings.Contains(result, title) {
		t.Errorf("Expected panel to contain title %q", title)
	}
	if !strings.Contains(result, content) {
		t.Errorf("Expected panel to contain content %q", content)
	}
}

func TestRenderPanelWithPadding(t *testing.T) {
	result := RenderPanelWithPadding("Title", "Content", BorderStyleRed, 2)
	
	if result == "" {
		t.Error("Expected non-empty panel output")
	}
}

func TestFormatKeyValue(t *testing.T) {
	tests := []struct {
		key      string
		value    string
		keyWidth int
		contains string
	}{
		{"Name", "Test", 10, "Name"},
		{"Project", "MyProject", 15, "Project"},
	}
	
	for _, tt := range tests {
		result := FormatKeyValue(tt.key, tt.value, tt.keyWidth)
		if !strings.Contains(result, tt.contains) {
			t.Errorf("Expected result to contain %q, got %q", tt.contains, result)
		}
		if !strings.Contains(result, tt.value) {
			t.Errorf("Expected result to contain value %q, got %q", tt.value, result)
		}
	}
}

func TestCreateTable(t *testing.T) {
	pairs := map[string]string{
		"Project": "technocrat",
		"Author":  "test",
	}
	
	result := CreateTable(pairs, 15)
	
	if result == "" {
		t.Error("Expected non-empty table output")
	}
	
	// Check that all keys and values appear
	for key, value := range pairs {
		if !strings.Contains(result, key) {
			t.Errorf("Expected table to contain key %q", key)
		}
		if !strings.Contains(result, value) {
			t.Errorf("Expected table to contain value %q", value)
		}
	}
}
