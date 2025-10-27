//go:build ignore

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	version    = "0.2.0"
	commit     = "unknown"
	buildDir   = "bin"
	binary     = "technocrat"
	installDir = "/usr/local/bin"
)

func main() {
	var (
		buildCmd = flag.Bool("build", false, "Build the binary")
		cleanCmd = flag.Bool("clean", false, "Clean build artifacts")
		testCmd  = flag.Bool("test", false, "Run tests")
		fmtCmd   = flag.Bool("fmt", false, "Format code")
		vetCmd   = flag.Bool("vet", false, "Run go vet")
		depsCmd  = flag.Bool("deps", false, "Download and tidy dependencies")
		allCmd   = flag.Bool("all", false, "Build everything")
	)

	flag.Parse()

	// Get version info from git if available
	if output, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output(); err == nil {
		commit = strings.TrimSpace(string(output))
	}

	if *cleanCmd {
		clean()
		return
	}

	if *depsCmd {
		deps()
		return
	}

	if *fmtCmd {
		formatCode()
		return
	}

	if *vetCmd {
		vet()
		return
	}

	if *testCmd {
		test()
		return
	}

	if *buildCmd || *allCmd {
		build()
		return
	}

	// Default: show help
	fmt.Println("Technocrat Build Tool")
	fmt.Println("\nUsage: go run build.go [options]")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nExamples:")
	fmt.Println("  go run build.go -build              # Build the binary")
	fmt.Println("  go run build.go -install            # Build and install")
	fmt.Println("  go run build.go -test               # Run tests")
	fmt.Println("  go run build.go -clean              # Clean build artifacts")
	fmt.Println("  go generate && go install ./cmd/technocrat  # Using go generate")
}

func build() {
	info("Building technocrat...")

	// Get the project root directory
	wd, err := os.Getwd()
	if err != nil {
		fatal("Failed to get working directory: %v", err)
	}

	// If we're in a subdirectory, find the project root
	for !fileExists(filepath.Join(wd, "go.mod")) {
		parent := filepath.Dir(wd)
		if parent == wd {
			fatal("Could not find project root (go.mod not found)")
		}
		wd = parent
	}

	buildPath := filepath.Join(wd, buildDir)
	if err := os.MkdirAll(buildPath, 0755); err != nil {
		fatal("Failed to create build directory: %v", err)
	}

	ldflags := fmt.Sprintf("-X main.version=%s -X main.commit=%s", version, commit)
	outputPath := filepath.Join(buildPath, binary)

	cmd := exec.Command("go", "build",
		"-ldflags", ldflags,
		"-o", outputPath,
		"./cmd/technocrat")

	cmd.Dir = wd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fatal("Build failed: %v", err)
	}

	success("Build complete: %s", outputPath)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func clean() {
	info("Cleaning build artifacts...")

	if err := os.RemoveAll(buildDir); err != nil && !os.IsNotExist(err) {
		fatal("Failed to clean: %v", err)
	}

	success("Clean complete!")
}

func test() {
	info("Running tests...")

	cmd := exec.Command("go", "test", "-v", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fatal("Tests failed: %v", err)
	}

	success("Tests passed!")
}

func formatCode() {
	info("Formatting code...")

	cmd := exec.Command("go", "fmt", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fatal("Format failed: %v", err)
	}

	success("Code formatted!")
}

func vet() {
	info("Running go vet...")

	cmd := exec.Command("go", "vet", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fatal("Vet found issues: %v", err)
	}

	success("Vet passed!")
}

func deps() {
	info("Downloading dependencies...")

	cmd := exec.Command("go", "mod", "download")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fatal("Failed to download dependencies: %v", err)
	}

	info("Tidying dependencies...")
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fatal("Failed to tidy dependencies: %v", err)
	}

	success("Dependencies updated!")
}

// Helper functions
func info(format string, args ...interface{}) {
	fmt.Printf("\033[0;36m[INFO]\033[0m "+format+"\n", args...)
}

func success(format string, args ...interface{}) {
	fmt.Printf("\033[0;32m[SUCCESS]\033[0m "+format+"\n", args...)
}

func fatal(format string, args ...interface{}) {
	fmt.Printf("\033[0;31m[ERROR]\033[0m "+format+"\n", args...)
	os.Exit(1)
}
