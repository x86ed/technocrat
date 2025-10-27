package tchncrt

import (
	"testing"
)

func TestGetFeatureName(t *testing.T) {
	tests := []struct {
		branch   string
		expected string
	}{
		{"feature/add-user-auth", "add-user-auth"},
		{"feat/new-api", "new-api"},
		{"bugfix/fix-login", "fix-login"},
		{"fix/broken-test", "broken-test"},
		{"hotfix/critical-bug", "critical-bug"},
		{"my-feature", "my-feature"},
		{"main", "main"},
	}

	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			result := getFeatureName(tt.branch)
			if result != tt.expected {
				t.Errorf("getFeatureName(%q) = %q, want %q", tt.branch, result, tt.expected)
			}
		})
	}
}

func TestCheckFeatureBranch(t *testing.T) {
	tests := []struct {
		branch    string
		hasGit    bool
		shouldErr bool
	}{
		{"feature/test", true, false},
		{"main", true, true},
		{"master", true, true},
		{"", true, true},
		{"feature/test", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			err := CheckFeatureBranch(tt.branch, tt.hasGit)
			if tt.shouldErr && err == nil {
				t.Errorf("CheckFeatureBranch(%q, %v) expected error, got nil", tt.branch, tt.hasGit)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("CheckFeatureBranch(%q, %v) expected no error, got %v", tt.branch, tt.hasGit, err)
			}
		})
	}
}
