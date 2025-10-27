package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
)

// StepStatus represents the status of a step
type StepStatus string

const (
	StatusPending StepStatus = "pending"
	StatusRunning StepStatus = "running"
	StatusDone    StepStatus = "done"
	StatusError   StepStatus = "error"
	StatusSkipped StepStatus = "skipped"
)

// Step represents a single step in the tracker
type Step struct {
	Key    string
	Label  string
	Status StepStatus
	Detail string
}

// StepTracker tracks and renders hierarchical steps with live updates
type StepTracker struct {
	Title     string
	Steps     []Step
	area      *pterm.AreaPrinter
	refreshCb func()
}

// NewStepTracker creates a new step tracker with the given title
func NewStepTracker(title string) *StepTracker {
	return &StepTracker{
		Title: title,
		Steps: []Step{},
	}
}

// AttachRefresh attaches a callback for live refresh
func (t *StepTracker) AttachRefresh(cb func()) {
	t.refreshCb = cb
}

// Add adds a new step to the tracker in pending status
func (t *StepTracker) Add(key, label string) {
	// Check if step already exists
	for i := range t.Steps {
		if t.Steps[i].Key == key {
			return // Already exists
		}
	}

	t.Steps = append(t.Steps, Step{
		Key:    key,
		Label:  label,
		Status: StatusPending,
		Detail: "",
	})
	t.maybeRefresh()
}

// Start marks a step as running
func (t *StepTracker) Start(key string, detail ...string) {
	detailStr := ""
	if len(detail) > 0 {
		detailStr = detail[0]
	}
	t.update(key, StatusRunning, detailStr)
}

// Complete marks a step as completed
func (t *StepTracker) Complete(key string, detail ...string) {
	detailStr := ""
	if len(detail) > 0 {
		detailStr = detail[0]
	}
	t.update(key, StatusDone, detailStr)
}

// Error marks a step as failed
func (t *StepTracker) Error(key string, detail ...string) {
	detailStr := ""
	if len(detail) > 0 {
		detailStr = detail[0]
	}
	t.update(key, StatusError, detailStr)
}

// Skip marks a step as skipped
func (t *StepTracker) Skip(key string, detail ...string) {
	detailStr := ""
	if len(detail) > 0 {
		detailStr = detail[0]
	}
	t.update(key, StatusSkipped, detailStr)
}

// update updates a step's status and detail
func (t *StepTracker) update(key string, status StepStatus, detail string) {
	for i := range t.Steps {
		if t.Steps[i].Key == key {
			t.Steps[i].Status = status
			if detail != "" {
				t.Steps[i].Detail = detail
			}
			t.maybeRefresh()
			return
		}
	}

	// Step doesn't exist, create it
	t.Steps = append(t.Steps, Step{
		Key:    key,
		Label:  key,
		Status: status,
		Detail: detail,
	})
	t.maybeRefresh()
}

// maybeRefresh calls the refresh callback if set
func (t *StepTracker) maybeRefresh() {
	if t.refreshCb != nil {
		t.refreshCb()
	}
}

// Render creates a tree-style rendering of the steps
func (t *StepTracker) Render() string {
	// Create title
	title := ColorCyan.Sprint(t.Title)

	// Build tree structure
	root := pterm.TreeNode{
		Text: title,
	}

	// Add each step as a child node
	for _, step := range t.Steps {
		node := t.renderStep(step)
		root.Children = append(root.Children, node)
	}

	// Render the tree
	tree := pterm.DefaultTree.WithRoot(root)
	rendered, _ := tree.Srender()
	return rendered
}

// renderStep renders a single step with appropriate symbol and styling
func (t *StepTracker) renderStep(step Step) pterm.TreeNode {
	var symbol string
	var labelStyle pterm.Style
	var detailStyle pterm.Style

	switch step.Status {
	case StatusDone:
		symbol = ColorGreen.Sprint(SymbolFilledCircle)
		labelStyle = *ColorWhite
		detailStyle = *ColorDim
	case StatusPending:
		symbol = ColorDim.Sprint(SymbolCircle)
		labelStyle = *ColorDim
		detailStyle = *ColorDim
	case StatusRunning:
		symbol = ColorCyan.Sprint(SymbolCircle)
		labelStyle = *ColorWhite
		detailStyle = *ColorDim
	case StatusError:
		symbol = ColorRed.Sprint(SymbolFilledCircle)
		labelStyle = *ColorWhite
		detailStyle = *ColorDim
	case StatusSkipped:
		symbol = ColorYellow.Sprint(SymbolCircle)
		labelStyle = *ColorWhite
		detailStyle = *ColorDim
	default:
		symbol = " "
		labelStyle = *ColorWhite
		detailStyle = *ColorDim
	}

	// Build the text
	var text string
	if step.Detail != "" {
		text = fmt.Sprintf("%s %s %s",
			symbol,
			labelStyle.Sprint(step.Label),
			detailStyle.Sprintf("(%s)", step.Detail))
	} else {
		text = fmt.Sprintf("%s %s",
			symbol,
			labelStyle.Sprint(step.Label))
	}

	return pterm.TreeNode{
		Text: text,
	}
}

// Print prints the tracker to stderr
func (t *StepTracker) Print() {
	fmt.Fprintln(os.Stderr, t.Render())
}

// StartLive starts live rendering of the tracker
func (t *StepTracker) StartLive() error {
	if !IsInteractive() {
		// For non-interactive mode, just print once
		return nil
	}

	area, err := pterm.DefaultArea.Start()
	if err != nil {
		return fmt.Errorf("failed to start area: %w", err)
	}

	t.area = area
	t.AttachRefresh(func() {
		if t.area != nil {
			t.area.Update(t.Render())
		}
	})

	// Initial render
	t.area.Update(t.Render())
	return nil
}

// StopLive stops live rendering
func (t *StepTracker) StopLive() {
	if t.area != nil {
		t.area.Stop()
		t.area = nil
	}
}

// Summary returns a summary of the tracker status
func (t *StepTracker) Summary() string {
	var done, errors, skipped, pending int
	for _, step := range t.Steps {
		switch step.Status {
		case StatusDone:
			done++
		case StatusError:
			errors++
		case StatusSkipped:
			skipped++
		case StatusPending, StatusRunning:
			pending++
		}
	}

	parts := []string{}
	if done > 0 {
		parts = append(parts, ColorGreen.Sprintf("%d completed", done))
	}
	if errors > 0 {
		parts = append(parts, ColorRed.Sprintf("%d failed", errors))
	}
	if skipped > 0 {
		parts = append(parts, ColorYellow.Sprintf("%d skipped", skipped))
	}
	if pending > 0 {
		parts = append(parts, ColorDim.Sprintf("%d pending", pending))
	}

	return strings.Join(parts, ", ")
}
