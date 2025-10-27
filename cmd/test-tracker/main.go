package main

import (
	"fmt"
	"time"

	"technocrat/internal/ui"
)

func main() {
	fmt.Println("StepTracker Demo - Watch live progress updates!\n")

	// Create a new tracker
	tracker := ui.NewStepTracker("Installation Progress")

	// Add steps
	tracker.Add("download", "Downloading template")
	tracker.Add("extract", "Extracting files")
	tracker.Add("permissions", "Setting permissions")
	tracker.Add("git", "Initializing git repository")

	// Start live rendering
	if err := tracker.StartLive(); err != nil {
		fmt.Printf("Warning: Live mode not available: %v\n", err)
		// Fallback to static rendering
		tracker.Print()
	}

	// Simulate download
	tracker.Start("download", "Fetching from GitHub...")
	time.Sleep(1 * time.Second)
	tracker.Complete("download", "2.3 MB downloaded")
	time.Sleep(500 * time.Millisecond)

	// Simulate extraction
	tracker.Start("extract", "Unpacking archive...")
	time.Sleep(800 * time.Millisecond)
	tracker.Complete("extract", "15 files extracted")
	time.Sleep(500 * time.Millisecond)

	// Simulate permissions
	tracker.Start("permissions", "Making scripts executable...")
	time.Sleep(600 * time.Millisecond)
	tracker.Complete("permissions", "3 files updated")
	time.Sleep(500 * time.Millisecond)

	// Simulate git init
	tracker.Start("git", "Running git init...")
	time.Sleep(700 * time.Millisecond)
	tracker.Complete("git", "Repository initialized")
	time.Sleep(500 * time.Millisecond)

	// Stop live rendering
	tracker.StopLive()

	// Show summary
	fmt.Println("\n" + ui.ColorGreen.Sprint("✓") + " " + tracker.Summary())
	fmt.Println()
}
