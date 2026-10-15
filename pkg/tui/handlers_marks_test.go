package tui

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nikbrunner/lazyjira/pkg/config"
	"github.com/nikbrunner/lazyjira/pkg/internal/testkit"
	"github.com/nikbrunner/lazyjira/pkg/jira"
	"github.com/nikbrunner/lazyjira/pkg/jira/jiratest"
	"github.com/nikbrunner/lazyjira/pkg/tui/components"
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
func TestYank_MarkedRowsThenSummary(t *testing.T) {
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

	if want := "A-3 Three"; copied != want {
		t.Fatalf("copied %q without marks, want %q", copied, want)
	}
}

//nolint:paralleltest // Replaces the process-wide clipboard command runner.
func TestYank_MarksHiddenByFilterCopySummary(t *testing.T) {
	original := runExternalCommand
	t.Cleanup(func() { runExternalCommand = original })
	var copied string
	runExternalCommand = func(input string, _ bool, _ string, _ ...string) { copied = input }

	app := marksApp(t)
	app.handleKeyMsg(tea.KeyMsg{Type: tea.KeySpace})
	app.issuesList.SetFilter("Three")

	app.handleKeyMsg(keyRunes("y"))

	if want := "A-3 Three"; copied != want {
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

//nolint:paralleltest // Replaces the process-wide clipboard command runner.
func TestCopy_URLAndMarkdownLinkUseCursorIssue(t *testing.T) {
	original := runExternalCommand
	t.Cleanup(func() { runExternalCommand = original })
	var copied string
	runExternalCommand = func(input string, _ bool, _ string, _ ...string) { copied = input }

	app := marksApp(t)
	app.cfg.Sanitize = config.SanitizeConfig{Remove: []string{"[", "]"}}
	app.issuesList.SetIssues([]jira.Issue{{Key: "A-1", Summary: "[web] Fix login"}, {Key: "A-2", Summary: "Two"}})
	app.handleKeyMsg(tea.KeyMsg{Type: tea.KeySpace})
	app.issuesList.Cursor = 1

	app.handleKeyMsg(keyRunes("Y"))
	if want := "https://jira.example/browse/A-2"; copied != want {
		t.Fatalf("Y copied %q, want %q", copied, want)
	}
	if !app.issuesList.HasMarks() {
		t.Fatal("Y cleared the marks")
	}

	app.issuesList.Cursor = 0
	app.handleKeyMsg(tea.KeyMsg{Type: tea.KeyCtrlY})
	if want := "[A-1 web Fix login](https://jira.example/browse/A-1)"; copied != want {
		t.Fatalf("ctrl+y copied %q, want %q", copied, want)
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

func bulkTransitionApp(t *testing.T, fake *jiratest.FakeClient) *App {
	t.Helper()
	app := marksApp(t)
	app.client = fake
	fake.GetTransitionsFunc = func(_ context.Context, key string) ([]jira.Transition, error) {
		done := &jira.Status{ID: "3", Name: "Done"}
		if key == "A-1" {
			return []jira.Transition{
				{ID: "11", Name: "Start", To: &jira.Status{ID: "2", Name: "In Progress"}},
				{ID: "31", Name: "Finish", To: done},
			}, nil
		}
		return []jira.Transition{{ID: "41", Name: "Close", To: done}}, nil
	}
	app.handleKeyMsg(keyRunes("v"))
	app.issuesList.Cursor = 1
	app.handleKeyMsg(keyRunes("v"))
	_, cmd := app.handleKeyMsg(keyRunes("t"))
	app.Update(cmd())
	return app
}

func TestTransition_MarkedIssuesOfferCommonTargets(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.DoTransitionFunc = func(context.Context, string, string) error { return nil }
	app := bulkTransitionApp(t, fake)

	testkit.AssertEqual(t, "modal title", app.modal.Title(), "Transition 2 issues")

	msg := app.onSelect(components.ModalItem{ID: "3"})()
	want := []jiratest.DoTransitionCall{{Key: "A-1", TransitionID: "31"}, {Key: "A-2", TransitionID: "41"}}
	for i, call := range fake.DoTransitionCalls {
		call.Ctx = nil
		testkit.AssertEqual(t, "transition call", call, want[i])
	}
	testkit.AssertEqual(t, "transition calls", len(fake.DoTransitionCalls), 2)

	app.Update(msg)
	if app.issuesList.HasMarks() {
		t.Fatal("marks remain after transition")
	}
}

func TestTransition_MarkedIssueFailureKeepsFailedMarked(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.DoTransitionFunc = func(_ context.Context, key, _ string) error {
		if key == "A-2" {
			return errors.New("forbidden")
		}
		return nil
	}
	app := bulkTransitionApp(t, fake)

	app.Update(app.onSelect(components.ModalItem{ID: "3"})())

	testkit.AssertEqual(t, "transition calls", len(fake.DoTransitionCalls), 2)
	marked := app.issuesList.MarkedIssues()
	testkit.AssertEqual(t, "marked issues", len(marked), 1)
	testkit.AssertEqual(t, "marked issue", marked[0].Key, "A-2")
}
