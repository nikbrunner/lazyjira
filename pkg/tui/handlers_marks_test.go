package tui

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/nikbrunner/lazyjira/pkg/tui/keyname/keytest"

	"github.com/nikbrunner/lazyjira/pkg/config"
	"github.com/nikbrunner/lazyjira/pkg/internal/testkit"
	"github.com/nikbrunner/lazyjira/pkg/jira"
	"github.com/nikbrunner/lazyjira/pkg/jira/jiratest"
	"github.com/nikbrunner/lazyjira/pkg/tui/components"
)

func keyRunes(s string) tea.KeyPressMsg { return keytest.Press(s) }

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
	app.handleKeyMsg(keytest.Press(" "))
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
	app.handleKeyMsg(keytest.Press(" "))
	app.issuesList.SetFilter("Three")

	app.handleKeyMsg(keyRunes("y"))

	if want := "A-3 Three"; copied != want {
		t.Fatalf("copied %q, want %q", copied, want)
	}

	app.handleKeyMsg(keytest.Press("esc"))

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

	app.handleKeyMsg(keytest.Press(" "))
	app.handleKeyMsg(keyRunes("y"))

	if want := "A-1 One"; copied != want {
		t.Fatalf("yanked %q, want %q", copied, want)
	}
}

type externalCalls struct {
	copied []string
	opened []string
}

func stubExternal(t *testing.T) *externalCalls {
	t.Helper()
	original := runExternalCommand
	t.Cleanup(func() { runExternalCommand = original })
	calls := &externalCalls{}
	runExternalCommand = func(input string, waitForExit bool, _ string, args ...string) {
		if waitForExit {
			calls.copied = append(calls.copied, input)
			return
		}
		calls.opened = append(calls.opened, args[len(args)-1])
	}
	return calls
}

func (c *externalCalls) lastCopied(t *testing.T) string {
	t.Helper()
	if len(c.copied) == 0 {
		t.Fatal("nothing copied")
	}
	return c.copied[len(c.copied)-1]
}

var altY = keytest.Press("alt+y")

func copyApp(t *testing.T) *App {
	t.Helper()
	app := marksApp(t)
	app.cfg.Sanitize = config.SanitizeConfig{Remove: []string{"[", "]"}}
	app.issuesList.SetFields([]string{"key", "type", "status", "summary"})
	app.issuesList.SetIssues([]jira.Issue{
		{Key: "A-205", Summary: "[web] Release notes", IssueType: &jira.IssueType{Name: "Task"}, Status: &jira.Status{Name: "Open"}},
		{Key: "A-206", Summary: "Changelog", IssueType: &jira.IssueType{Name: "Story"}, Status: &jira.Status{Name: "Done"}},
		{Key: "A-1000", Summary: "Unmarked", IssueType: &jira.IssueType{Name: "Bug"}, Status: &jira.Status{Name: "Open"}},
	})
	return app
}

func markFirstTwo(app *App) {
	app.handleKeyMsg(keytest.Press(" "))
	app.issuesList.Cursor = 1
	app.handleKeyMsg(keytest.Press(" "))
	app.issuesList.Cursor = 2
}

//nolint:paralleltest // Replaces the process-wide clipboard command runner.
func TestCopy_MarkedIssuesOneLinePerIssue(t *testing.T) {
	tests := []struct {
		name   string
		key    tea.KeyPressMsg
		want   string
		status string
	}{
		{"y copies key and summary only", keyRunes("y"), "A-205 web Release notes\nA-206 Changelog", "Copied 2 issues"},
		{"ctrl+y copies Markdown links", keytest.Press("ctrl+y"),
			"[A-205 web Release notes](https://jira.example/browse/A-205)\n[A-206 Changelog](https://jira.example/browse/A-206)",
			"Copied 2 Markdown links"},
		{"Y copies URLs", keyRunes("Y"), "https://jira.example/browse/A-205\nhttps://jira.example/browse/A-206", "Copied 2 URLs"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := stubExternal(t)
			app := copyApp(t)
			markFirstTwo(app)

			app.handleKeyMsg(tt.key)

			testkit.AssertEqual(t, "copied", calls.lastCopied(t), tt.want)
			if !strings.Contains(app.helpBar.View(), tt.status) {
				t.Errorf("status %q missing from %q", tt.status, app.helpBar.View())
			}
			if app.issuesList.HasMarks() {
				t.Error("marks remain after copy")
			}
		})
	}
}

//nolint:paralleltest // Replaces the process-wide clipboard command runner.
func TestCopy_OneMarkReportsSingular(t *testing.T) {
	tests := []struct {
		key    tea.KeyPressMsg
		status string
	}{
		{keyRunes("y"), "Copied 1 issue"},
		{keytest.Press("ctrl+y"), "Copied 1 Markdown link"},
		{keyRunes("Y"), "Copied 1 URL"},
		{altY, "Copied 1 issue row"},
	}
	for _, tt := range tests {
		t.Run(tt.key.String(), func(t *testing.T) {
			stubExternal(t)
			app := copyApp(t)
			app.handleKeyMsg(keytest.Press(" "))

			app.handleKeyMsg(tt.key)

			view := app.helpBar.View()
			if !strings.Contains(view, tt.status) || strings.Contains(view, tt.status+"s") {
				t.Errorf("status %q not reported singular in %q", tt.status, view)
			}
		})
	}
}

//nolint:paralleltest // Replaces the process-wide clipboard command runner.
func TestCopy_WithoutMarksUsesCursorIssue(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyPressMsg
		want string
	}{
		{"y", keyRunes("y"), "A-206 Changelog"},
		{"ctrl+y", keytest.Press("ctrl+y"), "[A-206 Changelog](https://jira.example/browse/A-206)"},
		{"Y", keyRunes("Y"), "https://jira.example/browse/A-206"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := stubExternal(t)
			app := copyApp(t)
			app.issuesList.Cursor = 1

			app.handleKeyMsg(tt.key)

			testkit.AssertEqual(t, "copied", calls.lastCopied(t), tt.want)
			if !strings.Contains(app.helpBar.View(), "Copied: "+tt.want) {
				t.Errorf("status missing from %q", app.helpBar.View())
			}
		})
	}
}

//nolint:paralleltest // Replaces the process-wide clipboard command runner.
func TestCopyRows_MarkedRowsArePadded(t *testing.T) {
	calls := stubExternal(t)
	app := marksApp(t)
	app.issuesList.SetFields([]string{"key", "assignee", "summary"})
	app.issuesList.SetIssues([]jira.Issue{
		{Key: "A-1", Summary: "One", Assignee: &jira.User{DisplayName: "Ann"}},
		{Key: "A-10", Summary: "Longer summary", Assignee: &jira.User{DisplayName: "Bob Builder"}},
		{Key: "A-100", Summary: "Unmarked"},
	})
	app.handleKeyMsg(keyRunes("v"))
	app.issuesList.Cursor = 1

	app.handleKeyMsg(altY)

	want := "A-1   Ann          One\nA-10  Bob Builder  Longer summary"
	testkit.AssertEqual(t, "copied rows", calls.lastCopied(t), want)
	if !strings.Contains(app.helpBar.View(), "Copied 2 issue rows") {
		t.Errorf("status missing from %q", app.helpBar.View())
	}
	if app.issuesList.HasMarks() {
		t.Error("marks remain after copying rows")
	}
}

//nolint:paralleltest // Replaces the process-wide clipboard command runner.
func TestCopyRows_WithoutMarksCopiesCursorRow(t *testing.T) {
	calls := stubExternal(t)
	app := marksApp(t)
	app.issuesList.SetFields([]string{"key", "assignee", "summary"})
	app.issuesList.SetIssues([]jira.Issue{
		{Key: "A-1", Summary: "One", Assignee: &jira.User{DisplayName: "Ann"}},
		{Key: "A-10", Summary: "Longer summary", Assignee: &jira.User{DisplayName: "Bob Builder"}},
	})

	app.handleKeyMsg(altY)

	testkit.AssertEqual(t, "copied row", calls.lastCopied(t), "A-1  Ann          One")
}

func TestKeymap_AltYMatchesCopyRows(t *testing.T) {
	t.Parallel()
	testkit.AssertEqual(t, "alt+y key string", altY.String(), "alt+y")
	testkit.AssertEqual(t, "alt+y action", DefaultKeymap().Match(altY.String()), ActCopyRows)
}

func browserApp(t *testing.T, n int) *App {
	t.Helper()
	app := marksApp(t)
	issues := make([]jira.Issue, n)
	for i := range issues {
		issues[i] = jira.Issue{Key: fmt.Sprintf("A-%d", i+1), Summary: "Issue"}
	}
	app.issuesList.SetIssues(issues)
	return app
}

func markAll(app *App, n int) {
	app.handleKeyMsg(keyRunes("v"))
	app.issuesList.Cursor = n - 1
}

//nolint:paralleltest // Replaces the process-wide command runner.
func TestBrowser_OpensEachMarkedIssue(t *testing.T) {
	calls := stubExternal(t)
	app := browserApp(t, 10)
	markAll(app, 10)

	app.handleKeyMsg(keyRunes("o"))

	testkit.AssertEqual(t, "opened tabs", len(calls.opened), 10)
	testkit.AssertEqual(t, "first tab", calls.opened[0], "https://jira.example/browse/A-1")
	if app.confirmModal.IsVisible() {
		t.Error("confirmation shown for 10 issues")
	}
	if !strings.Contains(app.helpBar.View(), "Opened 10 issues") {
		t.Errorf("status missing from %q", app.helpBar.View())
	}
	if app.issuesList.HasMarks() {
		t.Error("marks remain after opening")
	}
}

//nolint:paralleltest // Replaces the process-wide command runner.
func TestBrowser_WithoutMarksOpensCursorIssue(t *testing.T) {
	calls := stubExternal(t)
	app := browserApp(t, 3)
	app.issuesList.Cursor = 1

	app.handleKeyMsg(keyRunes("o"))

	testkit.AssertSliceEqual(t, "opened tabs", calls.opened, []string{"https://jira.example/browse/A-2"})
}

//nolint:paralleltest // Replaces the process-wide command runner.
func TestBrowser_ManyMarkedIssuesAskFirst(t *testing.T) {
	tests := []struct {
		name       string
		answer     tea.KeyPressMsg
		wantOpened int
		wantMarks  bool
	}{
		{"accept", keyRunes("y"), 11, false},
		{"cancel", keytest.Press("esc"), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := stubExternal(t)
			app := browserApp(t, 11)
			app.overlays = components.OverlayStack{&app.confirmModal}
			app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
			markAll(app, 11)

			app.Update(keyRunes("o"))

			testkit.AssertEqual(t, "opened before answer", len(calls.opened), 0)
			if !app.confirmModal.IsVisible() {
				t.Fatal("no confirmation for 11 issues")
			}
			if view := app.View().Content; !strings.Contains(view, "Open 11 browser tabs?") {
				t.Fatal("confirmation question missing from the view")
			}

			_, cmd := app.Update(tt.answer)
			if cmd != nil {
				app.Update(cmd())
			}

			testkit.AssertEqual(t, "opened tabs", len(calls.opened), tt.wantOpened)
			testkit.AssertEqual(t, "copied", len(calls.copied), 0)
			testkit.AssertEqual(t, "marks kept", app.issuesList.HasMarks(), tt.wantMarks)
			testkit.AssertEqual(t, "confirmation visible", app.confirmModal.IsVisible(), false)
			if tt.wantOpened > 0 && !strings.Contains(app.helpBar.View(), "Opened 11 issues") {
				t.Errorf("status missing from %q", app.helpBar.View())
			}
		})
	}
}

func TestMarkingKeys_IncludeCopyAndBrowserActions(t *testing.T) {
	t.Parallel()
	app := marksApp(t)
	for _, key := range []string{"ctrl+y", "Y", "o", "alt+y"} {
		testkit.AssertEqual(t, key+" without marks", app.isMarkingKey(key), false)
	}
	app.handleKeyMsg(keytest.Press(" "))
	for _, key := range []string{"y", "ctrl+y", "Y", "o", "alt+y"} {
		testkit.AssertEqual(t, key+" with marks", app.isMarkingKey(key), true)
	}
}

func TestEsc_ClearsMarksBeforeFilter(t *testing.T) {
	t.Parallel()
	app := marksApp(t)
	app.issuesList.SetFilter("A")
	app.handleKeyMsg(keytest.Press(" "))

	app.handleKeyMsg(keytest.Press("esc"))

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

func TestConfirm_DrawnAboveLaterOverlays(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.SetOnRequestFunc = func(func(jira.RequestLog)) {}
	app := NewApp(&config.Config{}, fake)
	app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	app.confirm("Remove link: blocks SHOP-12?", func() tea.Cmd { return nil })
	items := make([]components.ModalItem, 0, 12)
	for i := range 12 {
		items = append(items, components.ModalItem{ID: strconv.Itoa(i), Label: strings.Repeat("blocks ", 8)})
	}
	app.modal.Show("Link type", items)

	if !strings.Contains(app.View().Content, "Remove link: blocks SHOP-12?") {
		t.Fatal("confirmation hidden under a later overlay")
	}
}
