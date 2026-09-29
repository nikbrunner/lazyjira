package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/nikbrunner/lazyjira/pkg/jira"
	"github.com/nikbrunner/lazyjira/pkg/jira/jiratest"
	"github.com/nikbrunner/lazyjira/pkg/tui/components"
)

func newLookupApp(t *testing.T) *App {
	t.Helper()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.overlays = components.OverlayStack{&app.issueLookup}
	app.projectKey = testProject
	_, _ = app.handleIssuesLoaded(issuesLoadedMsg{tab: 0, issues: []jira.Issue{{Key: testKey}, {Key: testKey2}}})
	app.focusPane(focusIssues)
	return app
}

func TestIssueLookupOpensWithProjectPrefix(t *testing.T) {
	t.Parallel()
	app := newLookupApp(t)

	pressKey(app, "#")

	if !app.issueLookup.IsVisible() || app.issueLookup.Query() != testProject+"-" {
		t.Fatalf("lookup visible = %v, query = %q, want %s-", app.issueLookup.IsVisible(), app.issueLookup.Query(), testProject)
	}
}

func TestIssueLookupShowsIssueMaximizedAndEscRestores(t *testing.T) {
	t.Parallel()
	app := newLookupApp(t)

	_, _ = app.Update(components.IssueLookupSelectedMsg{Key: "WEBSDK-205"})

	if app.side != sideRight || !app.maximized || app.maximizedPane != focusDetailPane {
		t.Fatalf("side = %v, maximized = %v/%v, want maximized Issue details", app.side, app.maximized, app.maximizedPane)
	}
	if app.previewKey != "WEBSDK-205" || app.currentIssue().Key != "WEBSDK-205" {
		t.Fatalf("previewKey = %q, want WEBSDK-205", app.previewKey)
	}

	_, _ = app.handleIssuesLoaded(issuesLoadedMsg{tab: 0, issues: []jira.Issue{{Key: testKey}, {Key: testKey2}}})
	if app.previewKey != "WEBSDK-205" {
		t.Fatalf("a list refresh moved the preview to %q", app.previewKey)
	}

	pressKey(app, "esc")

	if app.side != sideLeft || app.leftFocus != focusIssues || app.maximized {
		t.Fatalf("side = %v, focus = %v, maximized = %v, want the split layout on Issues", app.side, app.leftFocus, app.maximized)
	}
	if app.previewKey != testKey {
		t.Fatalf("previewKey = %q, want the selected %s", app.previewKey, testKey)
	}
}

func TestIssueLookupEndsWhenFocusLeaves(t *testing.T) {
	t.Parallel()
	app := newLookupApp(t)
	_, _ = app.Update(components.IssueLookupSelectedMsg{Key: "WEBSDK-205"})

	app.focusPane(focusIssues)

	if app.lookupReturn != nil || app.previewKey != testKey {
		t.Fatalf("lookup = %v, previewKey = %q, want the lookup ended on %s", app.lookupReturn, app.previewKey, testKey)
	}
}

func TestIssueLookupMergesJiraAndLoadedSuggestions(t *testing.T) {
	t.Parallel()
	app := newLookupApp(t)
	app.issueCache["PLAT-11"] = &jira.Issue{Key: "PLAT-11", Summary: "cached"}
	pressKey(app, "#")
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})

	app.handleIssueSuggestionsLoaded(issueSuggestionsLoadedMsg{
		query:       testProject + "-1",
		suggestions: []jira.IssueSuggestion{{Key: "PLAT-100", Summary: "remote"}, {Key: "PLAT-10", Summary: "remote"}},
	})

	app.issueLookup.SetSize(120, 40)
	view := ansi.Strip(app.issueLookup.View())
	for _, want := range []string{"PLAT-100", "PLAT-10 ", "PLAT-11", "cached"} {
		if !strings.Contains(view, want) {
			t.Errorf("lookup view is missing %q:\n%s", want, view)
		}
	}
}
