package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/nikbrunner/lazyjira/pkg/tui/keyname/keytest"

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

func TestIssueLookupOpensIssueEditViewForUnloadedIssue(t *testing.T) {
	t.Parallel()
	other := editableIssue()
	other.Key = "WEBSDK-205"
	app, _ := newEditApp(t, other, nil)
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})
	delete(app.issueCache, other.Key)

	drive(app, app.handleIssueLookupSelected(other.Key))

	if !app.createForm.IsVisible() || app.createForm.IsLoading() {
		t.Fatal("the Issue Edit View should be open")
	}
	if app.createCtx.editKey != other.Key || app.createCtx.projectKey != "WEBSDK" {
		t.Fatalf("editKey = %q, projectKey = %q, want %s in WEBSDK", app.createCtx.editKey, app.createCtx.projectKey, other.Key)
	}
	if app.side != sideLeft || app.leftFocus != focusIssues || app.maximized {
		t.Fatalf("side = %v, focus = %v, maximized = %v, want the workspace layout unchanged", app.side, app.leftFocus, app.maximized)
	}
}

func TestIssueLookupMergesJiraAndLoadedSuggestions(t *testing.T) {
	t.Parallel()
	app := newLookupApp(t)
	app.issueCache["PLAT-11"] = &jira.Issue{Key: "PLAT-11", Summary: "cached"}
	pressKey(app, "#")
	_, _ = app.Update(keytest.Press("1"))

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
