package cmd

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestAgentConfigs verifies that all agent configurations are properly defined
func TestAgentConfigs(t *testing.T) {
	expectedAgents := []string{
		"copilot", "claude", "gemini", "cursor-agent", "qwen",
		"opencode", "codex", "windsurf", "kilocode", "auggie",
		"codebuddy", "roo", "q",
	}

	for _, agent := range expectedAgents {
		t.Run(agent, func(t *testing.T) {
			config, ok := agentConfigs[agent]
			if !ok {
				t.Errorf("Agent %s not found in agentConfigs", agent)
				return
			}

			if config.Name == "" {
				t.Errorf("Agent %s has empty Name", agent)
			}
			if config.Folder == "" {
				t.Errorf("Agent %s has empty Folder", agent)
			}

			// If RequiresCLI is true, InstallURL should typically be set (unless IDE-based)
			if config.RequiresCLI && config.InstallURL == "" {
				t.Logf("Note: Agent %s requires CLI but has no InstallURL", agent)
			}
		})
	}

	// Verify count
	if len(agentConfigs) != len(expectedAgents) {
		t.Errorf("Expected %d agents, got %d", len(expectedAgents), len(agentConfigs))
	}
}

// TestCheckToolInstalled tests the checkToolInstalled function
func TestCheckToolInstalled(t *testing.T) {
	t.Run("git should be available", func(t *testing.T) {
		// Git is typically available in CI environments
		if checkToolInstalled("git") {
			t.Log("git is installed")
		} else {
			t.Log("git is not installed (this is okay for some environments)")
		}
	})

	t.Run("nonexistent tool", func(t *testing.T) {
		if checkToolInstalled("this-tool-definitely-does-not-exist-12345") {
			t.Error("Expected nonexistent tool to return false")
		}
	})

	t.Run("claude special path", func(t *testing.T) {
		// This test verifies the special handling for Claude CLI
		// We create a mock file to test the logic
		tmpDir := t.TempDir()
		claudeDir := filepath.Join(tmpDir, ".claude", "local")
		claudePath := filepath.Join(claudeDir, "claude")

		// Set HOME to temp directory for this test
		origHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", origHome)

		// Create the directory structure
		if err := os.MkdirAll(claudeDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create the claude file
		if err := os.WriteFile(claudePath, []byte("#!/bin/sh\necho test"), 0755); err != nil {
			t.Fatal(err)
		}

		// Now check should find it
		if !checkToolInstalled("claude") {
			t.Error("Expected to find claude at special path")
		}
	})
}

// TestGetAgentList tests the getAgentList function
func TestGetAgentList(t *testing.T) {
	list := getAgentList()

	if list == "" {
		t.Error("getAgentList returned empty string")
	}

	// Should contain at least some known agents
	expectedAgents := []string{"copilot", "claude", "gemini"}
	for _, agent := range expectedAgents {
		if !strings.Contains(list, agent) {
			t.Errorf("Expected agent list to contain %s, got: %s", agent, list)
		}
	}
}

// TestGetDefaultScriptType tests platform-specific defaults
func TestGetDefaultScriptType(t *testing.T) {
	scriptType := getDefaultScriptType()

	if runtime.GOOS == "windows" {
		if scriptType != "ps" {
			t.Errorf("Expected 'ps' for Windows, got %s", scriptType)
		}
	} else {
		if scriptType != "sh" {
			t.Errorf("Expected 'sh' for non-Windows, got %s", scriptType)
		}
	}
}

// TestFormatBytes tests the byte formatting function
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int
		expected string
	}{
		{"zero", 0, "0 B"},
		{"bytes", 500, "500 B"},
		{"1KB", 1024, "1.0 KB"},
		{"1.5KB", 1536, "1.5 KB"},
		{"1MB", 1048576, "1.0 MB"},
		{"2.5MB", 2621440, "2.5 MB"},
		{"1GB", 1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %s, want %s", tt.bytes, result, tt.expected)
			}
		})
	}
}

// TestIsGitRepo tests the isGitRepo function
func TestIsGitRepo(t *testing.T) {
	t.Run("valid git repository", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Initialize git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Skipf("Skipping test: git not available: %v", err)
		}

		if !isGitRepo(tmpDir) {
			t.Error("Expected isGitRepo to return true for valid git repo")
		}
	})

	t.Run("not a git repository", func(t *testing.T) {
		tmpDir := t.TempDir()

		if isGitRepo(tmpDir) {
			t.Error("Expected isGitRepo to return false for non-git directory")
		}
	})

	t.Run("has .git directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")

		if err := os.MkdirAll(gitDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Should return true even without valid git repo
		if !isGitRepo(tmpDir) {
			t.Error("Expected isGitRepo to return true when .git directory exists")
		}
	})
}

// TestInitGitRepo tests the initGitRepo function
func TestInitGitRepo(t *testing.T) {
	// Check if git is available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Skipping test: git not available")
	}

	// Configure git for testing (required for git commit)
	gitConfigured := false
	if output, err := exec.Command("git", "config", "--global", "user.email").CombinedOutput(); err == nil && len(strings.TrimSpace(string(output))) > 0 {
		gitConfigured = true
	}

	if !gitConfigured {
		// Set temporary git config
		exec.Command("git", "config", "--global", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "--global", "user.name", "Test User").Run()
	}

	tmpDir := t.TempDir()

	// Create a dummy file so we have something to commit
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := initGitRepo(tmpDir); err != nil {
		t.Errorf("initGitRepo failed: %v", err)
	}

	// Verify .git directory exists
	gitDir := filepath.Join(tmpDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Error(".git directory was not created")
	}

	// Verify commit was made
	cmd := exec.Command("git", "log", "--oneline")
	cmd.Dir = tmpDir
	output, err := cmd.Output()
	if err != nil {
		t.Errorf("Failed to get git log: %v", err)
	}

	if !strings.Contains(string(output), "Initial commit from Technocrat template") {
		t.Errorf("Initial commit not found in git log: %s", string(output))
	}
}

// TestMakeScriptsExecutable tests the makeScriptsExecutable function
func TestMakeScriptsExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	tmpDir := t.TempDir()
	scriptsDir := filepath.Join(tmpDir, ".tchncrt", "scripts")

	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test scripts
	tests := []struct {
		name       string
		content    string
		shouldExec bool
	}{
		{"valid-script.sh", "#!/bin/bash\necho test", true},
		{"no-shebang.sh", "echo test", false},
		{"not-a-script.txt", "just text", false},
		{"python-script.sh", "#!/usr/bin/env python3\nprint('test')", true},
	}

	for _, tt := range tests {
		scriptPath := filepath.Join(scriptsDir, tt.name)
		if err := os.WriteFile(scriptPath, []byte(tt.content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Run makeScriptsExecutable
	if err := makeScriptsExecutable(tmpDir); err != nil {
		t.Errorf("makeScriptsExecutable failed: %v", err)
	}

	// Verify permissions
	for _, tt := range tests {
		if !strings.HasSuffix(tt.name, ".sh") {
			continue // Only check .sh files
		}

		scriptPath := filepath.Join(scriptsDir, tt.name)
		info, err := os.Stat(scriptPath)
		if err != nil {
			t.Fatal(err)
		}

		mode := info.Mode()
		isExecutable := mode&0111 != 0

		if tt.shouldExec && !isExecutable {
			t.Errorf("Script %s should be executable but is not (mode: %v)", tt.name, mode)
		}
		if !tt.shouldExec && isExecutable {
			t.Errorf("Script %s should not be executable but is (mode: %v)", tt.name, mode)
		}
	}
}

// TestMakeScriptsExecutableNoScriptsDir tests behavior when scripts directory doesn't exist
func TestMakeScriptsExecutableNoScriptsDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	tmpDir := t.TempDir()

	// Should not error when scripts directory doesn't exist
	if err := makeScriptsExecutable(tmpDir); err != nil {
		t.Errorf("Expected no error when scripts directory doesn't exist, got: %v", err)
	}
}

// TestExtractZipWithStats tests ZIP extraction with statistics
func TestExtractZipWithStats(t *testing.T) {
	// Create a temporary ZIP file
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test.zip")
	extractDir := filepath.Join(tmpDir, "extract")

	// Create test ZIP
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}

	w := zip.NewWriter(zipFile)

	// Add some test files
	files := map[string]string{
		"root-dir/file1.txt": "content1",
		"root-dir/file2.txt": "content2",
		"root-dir/sub/file3.txt": "content3",
	}

	for name, content := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	zipFile.Close()

	// Extract ZIP
	count, size, err := extractZipWithStats(zipPath, extractDir, false, nil)
	if err != nil {
		t.Errorf("extractZipWithStats failed: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 files extracted, got %d", count)
	}

	expectedSize := len("content1") + len("content2") + len("content3")
	if size != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, size)
	}

	// Verify files were extracted (with root-dir stripped)
	for name, content := range files {
		// Strip "root-dir/" prefix
		strippedName := strings.TrimPrefix(name, "root-dir/")
		extractedPath := filepath.Join(extractDir, strippedName)

		data, err := os.ReadFile(extractedPath)
		if err != nil {
			t.Errorf("Failed to read extracted file %s: %v", strippedName, err)
			continue
		}

		if string(data) != content {
			t.Errorf("File %s: expected content %q, got %q", strippedName, content, string(data))
		}
	}
}

// TestExtractZipWithStatsCurrentDir tests extraction to current directory
func TestExtractZipWithStatsCurrentDir(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test.zip")

	// Create test ZIP
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}

	w := zip.NewWriter(zipFile)
	f, _ := w.Create("test.txt")
	f.Write([]byte("content"))
	w.Close()
	zipFile.Close()

	// Extract to current dir (tmpDir in this case)
	_, _, err = extractZipWithStats(zipPath, tmpDir, true, nil)
	if err != nil {
		t.Errorf("extractZipWithStats with inCurrentDir failed: %v", err)
	}

	// Verify file exists
	extractedPath := filepath.Join(tmpDir, "test.txt")
	if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
		t.Error("Expected file to be extracted to current directory")
	}
}

// TestExtractZip tests legacy extractZip function
func TestExtractZip(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test.zip")
	extractDir := filepath.Join(tmpDir, "extract")

	// Create test ZIP
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}

	w := zip.NewWriter(zipFile)
	f, _ := w.Create("test.txt")
	f.Write([]byte("content"))
	w.Close()
	zipFile.Close()

	// Extract using legacy function
	if err := extractZip(zipPath, extractDir, false); err != nil {
		t.Errorf("extractZip failed: %v", err)
	}

	// Verify file exists
	extractedPath := filepath.Join(extractDir, "test.txt")
	if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
		t.Error("Expected file to be extracted")
	}
}

// TestMin tests the min helper function
func TestMin(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{1, 2, 1},
		{5, 3, 3},
		{10, 10, 10},
		{0, 100, 0},
		{-5, 5, -5},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("min(%d,%d)", tt.a, tt.b), func(t *testing.T) {
			result := min(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestCenterText tests the text centering function
func TestCenterText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected string
	}{
		{
			name:     "short text",
			text:     "hello",
			width:    11,
			expected: "   hello",
		},
		{
			name:     "exact width",
			text:     "hello",
			width:    5,
			expected: "hello",
		},
		{
			name:     "text longer than width",
			text:     "hello world",
			width:    5,
			expected: "hello world",
		},
		{
			name:     "zero width",
			text:     "test",
			width:    0,
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := centerText(tt.text, tt.width)
			if result != tt.expected {
				t.Errorf("centerText(%q, %d) = %q, want %q", tt.text, tt.width, result, tt.expected)
			}
		})
	}
}

// TestGetTerminalWidth tests the getTerminalWidth function
func TestGetTerminalWidth(t *testing.T) {
	// This test may fail in non-terminal environments
	width, err := getTerminalWidth()

	if err != nil {
		t.Logf("getTerminalWidth returned error (expected in non-terminal): %v", err)
		return
	}

	if width <= 0 {
		t.Errorf("Expected positive width, got %d", width)
	}

	t.Logf("Terminal width: %d", width)
}

// MockGitHubServer creates a test HTTP server that mocks GitHub API responses
func MockGitHubServer(t *testing.T, releaseResponse string, assetContent []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/releases/latest"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(releaseResponse))
		case strings.Contains(r.URL.Path, "/download/"):
			w.Header().Set("Content-Type", "application/zip")
			w.WriteHeader(http.StatusOK)
			w.Write(assetContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// TestDownloadAndExtractTemplateIntegration is an integration test for download functionality
func TestDownloadAndExtractTemplateIntegration(t *testing.T) {
	// Create a mock ZIP file
	tmpDir := t.TempDir()
	mockZip := filepath.Join(tmpDir, "mock.zip")

	zipFile, err := os.Create(mockZip)
	if err != nil {
		t.Fatal(err)
	}

	w := zip.NewWriter(zipFile)
	f, _ := w.Create("root-dir/test.txt")
	f.Write([]byte("test content"))
	w.Close()
	zipFile.Close()

	// Read the ZIP content
	zipContent, err := os.ReadFile(mockZip)
	if err != nil {
		t.Fatal(err)
	}

	// Create mock GitHub API response
	releaseResponse := map[string]interface{}{
		"tag_name": "v1.0.0",
		"assets": []map[string]interface{}{
			{
				"name":                 "technocrat-template-copilot-sh.zip",
				"browser_download_url": "http://example.com/download/asset.zip",
				"size":                 len(zipContent),
			},
		},
	}

	releaseJSON, _ := json.Marshal(releaseResponse)

	// Create mock server
	server := MockGitHubServer(t, string(releaseJSON), zipContent)
	defer server.Close()

	t.Logf("Note: This test would require modifying the download function to accept a custom API URL")
	t.Logf("Mock server running at: %s", server.URL)

	// This is a placeholder - actual integration would require refactoring
	// downloadAndExtractTemplate to accept custom URLs for testing
}

// TestShowBanner tests that the banner display doesn't panic
func TestShowBanner(t *testing.T) {
	// Redirect stderr to prevent output during test
	origStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("showBanner panicked: %v", r)
		}
	}()

	showBanner()

	// Close writer and read output
	w.Close()
	os.Stderr = origStderr

	output := make([]byte, 4096)
	n, _ := r.Read(output)

	if n == 0 {
		t.Error("Expected banner output, got none")
	}

	// Check for key elements - tagline contains "Technocrat"
	bannerStr := string(output[:n])
	if !strings.Contains(bannerStr, "Spec-Driven Development") {
		t.Error("Banner output should contain 'Spec-Driven Development'")
	}
}

// TestAgentConfigValues verifies specific agent configuration values
func TestAgentConfigValues(t *testing.T) {
	tests := []struct {
		agent       string
		folder      string
		requiresCLI bool
	}{
		{"copilot", ".github/", false},
		{"claude", ".claude/", true},
		{"gemini", ".gemini/", true},
		{"cursor-agent", ".cursor/", false},
		{"windsurf", ".windsurf/", false},
		{"q", ".amazonq/", true},
	}

	for _, tt := range tests {
		t.Run(tt.agent, func(t *testing.T) {
			config, ok := agentConfigs[tt.agent]
			if !ok {
				t.Fatalf("Agent %s not found", tt.agent)
			}

			if config.Folder != tt.folder {
				t.Errorf("Expected folder %s, got %s", tt.folder, config.Folder)
			}

			if config.RequiresCLI != tt.requiresCLI {
				t.Errorf("Expected RequiresCLI=%v, got %v", tt.requiresCLI, config.RequiresCLI)
			}
		})
	}
}

// TestInitCmdFlags verifies that all expected flags are registered
func TestInitCmdFlags(t *testing.T) {
	expectedFlags := []string{
		"ai",
		"script",
		"ignore-agent-tools",
		"no-git",
		"here",
		"force",
		"github-token",
		"skip-tls",
		"debug",
	}

	for _, flagName := range expectedFlags {
		flag := initCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag --%s to be registered", flagName)
		}
	}
}

// TestInitCmdMetadata verifies command metadata
func TestInitCmdMetadata(t *testing.T) {
	if initCmd.Use == "" {
		t.Error("initCmd.Use should not be empty")
	}

	if initCmd.Short == "" {
		t.Error("initCmd.Short should not be empty")
	}

	if initCmd.Long == "" {
		t.Error("initCmd.Long should not be empty")
	}

	if initCmd.RunE == nil {
		t.Error("initCmd.RunE should be set")
	}
}

// TestBannerAndTaglineConstants verifies constants are defined
func TestBannerAndTaglineConstants(t *testing.T) {
	if banner == "" {
		t.Error("banner constant should not be empty")
	}

	if tagline == "" {
		t.Error("tagline constant should not be empty")
	}

	// Banner is ASCII art
	if !strings.Contains(banner, "█") {
		t.Error("banner should contain ASCII art characters")
	}

	if !strings.Contains(tagline, "Technocrat") {
		t.Error("tagline should contain 'Technocrat'")
	}

	if !strings.Contains(tagline, "Spec-Driven Development") {
		t.Error("tagline should contain 'Spec-Driven Development'")
	}
}

// TestExtractZipNestedDirectories tests extraction with nested directory structure
func TestExtractZipNestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test.zip")
	extractDir := filepath.Join(tmpDir, "extract")

	// Create test ZIP with nested directories
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}

	w := zip.NewWriter(zipFile)

	// Add nested structure
	files := []string{
		"root/a/b/c/deep.txt",
		"root/x/y/z/file.txt",
		"root/level1/level2/test.txt",
	}

	for _, name := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		f.Write([]byte("content"))
	}

	w.Close()
	zipFile.Close()

	// Extract
	_, _, err = extractZipWithStats(zipPath, extractDir, false, nil)
	if err != nil {
		t.Errorf("extractZipWithStats failed: %v", err)
	}

	// Verify nested files exist (with root/ stripped)
	for _, name := range files {
		strippedName := strings.TrimPrefix(name, "root/")
		extractedPath := filepath.Join(extractDir, strippedName)

		if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
			t.Errorf("Expected nested file %s to exist", strippedName)
		}
	}
}

// BenchmarkFormatBytes benchmarks the formatBytes function
func BenchmarkFormatBytes(b *testing.B) {
	sizes := []int{0, 1024, 1048576, 1073741824}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				formatBytes(size)
			}
		})
	}
}

// BenchmarkCheckToolInstalled benchmarks the checkToolInstalled function
func BenchmarkCheckToolInstalled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		checkToolInstalled("git")
	}
}

// BenchmarkCenterText benchmarks the centerText function
func BenchmarkCenterText(b *testing.B) {
	text := "Technocrat - Spec-Driven Development"
	width := 80

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centerText(text, width)
	}
}
