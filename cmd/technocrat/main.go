//go:generate go run ../../build.go -build

package main

import (
	"os"
	"technocrat/internal/cmd"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, commit)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
