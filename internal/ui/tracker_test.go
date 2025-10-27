package ui

import (
	"strings"
	"testing"
)

func TestNewStepTracker(t *testing.T) {
	tracker := NewStepTracker("Test Progress")

	if tracker.Title != "Test Progress" {
		t.Errorf("Expected title 'Test Progress', got '%s'", tracker.Title)
	}

	if len(tracker.Steps) != 0 {
		t.Errorf("Expected 0 steps, got %d", len(tracker.Steps))
	}
}

func TestStepTrackerAdd(t *testing.T) {
	tracker := NewStepTracker("Test")

	tracker.Add("step1", "First step")
	tracker.Add("step2", "Second step")

	if len(tracker.Steps) != 2 {
		t.Fatalf("Expected 2 steps, got %d", len(tracker.Steps))
	}

	if tracker.Steps[0].Key != "step1" {
		t.Errorf("Expected step1, got %s", tracker.Steps[0].Key)
	}

	if tracker.Steps[0].Status != StatusPending {
		t.Errorf("Expected pending status, got %s", tracker.Steps[0].Status)
	}

	// Test duplicate add (should be ignored)
	tracker.Add("step1", "Duplicate")
	if len(tracker.Steps) != 2 {
		t.Errorf("Expected 2 steps after duplicate add, got %d", len(tracker.Steps))
	}
}

func TestStepTrackerStart(t *testing.T) {
	tracker := NewStepTracker("Test")
	tracker.Add("step1", "First step")

	tracker.Start("step1", "Starting...")

	if tracker.Steps[0].Status != StatusRunning {
		t.Errorf("Expected running status, got %s", tracker.Steps[0].Status)
	}

	if tracker.Steps[0].Detail != "Starting..." {
		t.Errorf("Expected detail 'Starting...', got '%s'", tracker.Steps[0].Detail)
	}
}

func TestStepTrackerComplete(t *testing.T) {
	tracker := NewStepTracker("Test")
	tracker.Add("step1", "First step")
	tracker.Start("step1")

	tracker.Complete("step1", "Done!")

	if tracker.Steps[0].Status != StatusDone {
		t.Errorf("Expected done status, got %s", tracker.Steps[0].Status)
	}

	if tracker.Steps[0].Detail != "Done!" {
		t.Errorf("Expected detail 'Done!', got '%s'", tracker.Steps[0].Detail)
	}
}

func TestStepTrackerError(t *testing.T) {
	tracker := NewStepTracker("Test")
	tracker.Add("step1", "First step")
	tracker.Start("step1")

	tracker.Error("step1", "Failed: timeout")

	if tracker.Steps[0].Status != StatusError {
		t.Errorf("Expected error status, got %s", tracker.Steps[0].Status)
	}

	if tracker.Steps[0].Detail != "Failed: timeout" {
		t.Errorf("Expected detail 'Failed: timeout', got '%s'", tracker.Steps[0].Detail)
	}
}

func TestStepTrackerSkip(t *testing.T) {
	tracker := NewStepTracker("Test")
	tracker.Add("step1", "First step")

	tracker.Skip("step1", "Not needed")

	if tracker.Steps[0].Status != StatusSkipped {
		t.Errorf("Expected skipped status, got %s", tracker.Steps[0].Status)
	}

	if tracker.Steps[0].Detail != "Not needed" {
		t.Errorf("Expected detail 'Not needed', got '%s'", tracker.Steps[0].Detail)
	}
}

func TestStepTrackerUpdateNonExistent(t *testing.T) {
	tracker := NewStepTracker("Test")

	// Update a step that doesn't exist yet
	tracker.Complete("step1", "Done")

	if len(tracker.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(tracker.Steps))
	}

	if tracker.Steps[0].Key != "step1" {
		t.Errorf("Expected key 'step1', got '%s'", tracker.Steps[0].Key)
	}

	if tracker.Steps[0].Status != StatusDone {
		t.Errorf("Expected done status, got %s", tracker.Steps[0].Status)
	}
}

func TestStepTrackerRender(t *testing.T) {
	tracker := NewStepTracker("Test Progress")
	tracker.Add("step1", "First step")
	tracker.Add("step2", "Second step")
	tracker.Start("step1")
	tracker.Complete("step2")

	rendered := tracker.Render()

	// Check that title is present
	if !strings.Contains(rendered, "Test Progress") {
		t.Error("Expected rendered output to contain title")
	}

	// Check that steps are present
	if !strings.Contains(rendered, "First step") {
		t.Error("Expected rendered output to contain 'First step'")
	}

	if !strings.Contains(rendered, "Second step") {
		t.Error("Expected rendered output to contain 'Second step'")
	}
}

func TestStepTrackerSummary(t *testing.T) {
	tracker := NewStepTracker("Test")
	tracker.Add("step1", "Step 1")
	tracker.Add("step2", "Step 2")
	tracker.Add("step3", "Step 3")
	tracker.Add("step4", "Step 4")

	tracker.Complete("step1")
	tracker.Error("step2")
	tracker.Skip("step3")
	// step4 remains pending

	summary := tracker.Summary()

	// Check summary contains expected counts
	if !strings.Contains(summary, "1 completed") {
		t.Errorf("Expected summary to contain '1 completed', got: %s", summary)
	}

	if !strings.Contains(summary, "1 failed") {
		t.Errorf("Expected summary to contain '1 failed', got: %s", summary)
	}

	if !strings.Contains(summary, "1 skipped") {
		t.Errorf("Expected summary to contain '1 skipped', got: %s", summary)
	}

	if !strings.Contains(summary, "1 pending") {
		t.Errorf("Expected summary to contain '1 pending', got: %s", summary)
	}
}

func TestStepTrackerRefreshCallback(t *testing.T) {
	tracker := NewStepTracker("Test")

	callCount := 0
	tracker.AttachRefresh(func() {
		callCount++
	})

	tracker.Add("step1", "Step 1")
	if callCount != 1 {
		t.Errorf("Expected refresh to be called once, got %d", callCount)
	}

	tracker.Start("step1")
	if callCount != 2 {
		t.Errorf("Expected refresh to be called twice, got %d", callCount)
	}

	tracker.Complete("step1")
	if callCount != 3 {
		t.Errorf("Expected refresh to be called three times, got %d", callCount)
	}
}
