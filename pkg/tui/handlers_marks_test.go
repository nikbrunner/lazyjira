package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nikbrunner/lazyjira/v2/pkg/config"
	"github.com/nikbrunner/lazyjira/v2/pkg/jira"
)

func keyRunes(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

func marksApp(t *testing.T) *App {
	t.Helper()
	app := focusApp(t)
	app.side = sideLeft
	app.leftFocus = focusIssues
	app.cfg.Jira.Host = "https://jira.example"
	app.issuesList.SetFields([]string{"key", "summary"})
	app.issuesList.SetIssues([]jira.Issue{
		{Key: "A-1", Summary: "One"},
		{Key: "A-2", Summary: "Two"},
		{Key: "A-3", Summary: "Three"},
	})
	return app
}

//nolint:paralleltest // Replaces the process-wide clipboard command runner.
func TestYank_MarkedRowsThenURL(t *testing.T) {
	original := runExternalCommand
	t.Cleanup(func() { runExternalCommand = original })
	var copied string
	runExternalCommand = func(input string, _ bool, _ string, _ ...string) { copied = input }

	app := marksApp(t)
	app.handleKeyMsg(keyRunes("v"))
	app.handleKeyMsg(keyRunes("j"))
	app.issuesList.Cursor = 1
	app.handleKeyMsg(keyRunes("v"))
	app.issuesList.Cursor = 2
	app.handleKeyMsg(tea.KeyMsg{Type: tea.KeySpace})
	app.handleKeyMsg(keyRunes("y"))

	if want := "A-1 One\nA-2 Two\nA-3 Three"; copied != want {
		t.Fatalf("yanked %q, want %q", copied, want)
	}
	if app.issuesList.HasMarks() {
		t.Fatal("marks remain after yank")
	}

	app.handleKeyMsg(keyRunes("y"))

	if want := "https://jira.example/browse/A-3"; copied != want {
		t.Fatalf("copied %q without marks, want %q", copied, want)
	}
}

//nolint:paralleltest // Replaces the process-wide clipboard command runner.
func TestYank_MarksHiddenByFilterCopyURL(t *testing.T) {
	original := runExternalCommand
	t.Cleanup(func() { runExternalCommand = original })
	var copied string
	runExternalCommand = func(input string, _ bool, _ string, _ ...string) { copied = input }

	app := marksApp(t)
	app.handleKeyMsg(tea.KeyMsg{Type: tea.KeySpace})
	app.issuesList.SetFilter("Three")

	app.handleKeyMsg(keyRunes("y"))

	if want := "https://jira.example/browse/A-3"; copied != want {
		t.Fatalf("copied %q, want %q", copied, want)
	}

	app.handleKeyMsg(tea.KeyMsg{Type: tea.KeyEsc})

	if app.issuesList.IsFiltered() {
		t.Fatal("esc did not clear the filter")
	}
	if !app.issuesList.HasMarks() {
		t.Fatal("hidden mark lost after clearing the filter")
	}
}

//nolint:paralleltest // Replaces the process-wide clipboard command runner.
func TestYank_MarkingModeBypassesCustomCommand(t *testing.T) {
	original := runExternalCommand
	t.Cleanup(func() { runExternalCommand = original })
	var copied string
	runExternalCommand = func(input string, _ bool, _ string, _ ...string) { copied = input }

	app := marksApp(t)
	app.ctx = t.Context()
	app.customCmds = []config.ResolvedCustomCommand{resolvedCommand(t, "y", "copy", "true", config.CtxIssues)}

	if _, cmd := app.handleKeyMsg(keyRunes("y")); cmd == nil {
		t.Fatal("y without marks should run the custom command")
	}

	app.handleKeyMsg(tea.KeyMsg{Type: tea.KeySpace})
	app.handleKeyMsg(keyRunes("y"))

	if want := "A-1 One"; copied != want {
		t.Fatalf("yanked %q, want %q", copied, want)
	}
}

func TestEsc_ClearsMarksBeforeFilter(t *testing.T) {
	t.Parallel()
	app := marksApp(t)
	app.issuesList.SetFilter("A")
	app.handleKeyMsg(tea.KeyMsg{Type: tea.KeySpace})

	app.handleKeyMsg(tea.KeyMsg{Type: tea.KeyEsc})

	if app.issuesList.HasMarks() {
		t.Fatal("marks remain after esc")
	}
	if !app.issuesList.IsFiltered() {
		t.Fatal("first esc cleared the filter too")
	}
}
