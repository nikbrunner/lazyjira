package tui

import (
	"reflect"
	"slices"
	"testing"

	"github.com/nikbrunner/lazyjira/pkg/jira"
	"github.com/nikbrunner/lazyjira/pkg/jira/jiratest"
	"github.com/nikbrunner/lazyjira/pkg/tui/navstack"
	"github.com/nikbrunner/lazyjira/pkg/tui/views"
)

func appForKeybindings(t *testing.T) *App {
	t.Helper()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	return app
}

func TestContextBindings_ContainsQuit(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		setup func(*App)
	}{
		{
			name:  "issues focus",
			setup: func(app *App) { app.side = sideLeft; app.leftFocus = focusIssues },
		},
		{
			name:  "info focus",
			setup: func(app *App) { app.side = sideLeft; app.leftFocus = focusInfo },
		},
		{
			name:  "projects focus",
			setup: func(app *App) { app.side = sideLeft; app.leftFocus = focusProjects },
		},
		{
			name:  "status focus",
			setup: func(app *App) { app.side = sideLeft; app.leftFocus = focusProjects },
		},
		{
			name:  "detail side",
			setup: func(app *App) { app.side = sideRight },
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			app := appForKeybindings(t)
			testCase.setup(app)
			bindings := app.ContextBindings()
			if len(bindings) == 0 {
				t.Fatal("expected non-empty bindings")
			}
			found := false
			for _, binding := range bindings {
				if binding.Description == string(ActQuit) {
					found = true
					break
				}
			}
			if !found {
				t.Error("quit binding missing from context bindings")
			}
		})
	}
}

func TestContextBindings_RefreshAllIsGlobal(t *testing.T) {
	t.Parallel()
	app := appForKeybindings(t)
	for _, focus := range []focusPanel{focusIssues, focusInfo, focusProjects} {
		app.leftFocus = focus
		found := false
		for _, binding := range app.ContextBindings() {
			if binding.Description == "refresh all data" && binding.Key == "R" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Refresh All binding missing for focus %v", focus)
		}
	}
}

func TestContextBindings_DetailCommentsIncludesEdit(t *testing.T) {
	t.Parallel()
	app := appForKeybindings(t)
	app.side = sideRight
	app.detailView.SetIssue(&jira.Issue{Key: testKey})
	app.detailView.SetActiveTab(views.TabComments)

	bindings := app.ContextBindings()

	found := false
	for _, binding := range bindings {
		if binding.Description == "edit comment" {
			found = true
			break
		}
	}
	if !found {
		t.Error("edit comment binding missing when on comments tab")
	}
}

func TestHelpBarItems_NotEmpty(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		setup func(*App)
	}{
		{
			name:  "issues panel",
			setup: func(app *App) { app.side = sideLeft; app.leftFocus = focusIssues },
		},
		{
			name:  "info panel",
			setup: func(app *App) { app.side = sideLeft; app.leftFocus = focusInfo },
		},
		{
			name:  "projects panel",
			setup: func(app *App) { app.side = sideLeft; app.leftFocus = focusProjects },
		},
		{
			name:  "status panel",
			setup: func(app *App) { app.side = sideLeft; app.leftFocus = focusProjects },
		},
		{
			name:  "detail right panel",
			setup: func(app *App) { app.side = sideRight },
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			app := appForKeybindings(t)
			testCase.setup(app)
			items := app.helpBarItems()
			if len(items) == 0 {
				t.Error("expected non-empty help bar items")
			}
		})
	}
}

func TestNavBindings_HasSixEntries(t *testing.T) {
	t.Parallel()
	app := appForKeybindings(t)
	bindings := app.navBindings()
	if len(bindings) != 6 {
		t.Errorf("navBindings len = %d, want 6", len(bindings))
	}
}

func TestDetailScrollBindings_HasFourEntries(t *testing.T) {
	t.Parallel()
	app := appForKeybindings(t)
	bindings := app.detailScrollBindings()
	if len(bindings) != 4 {
		t.Errorf("detailScrollBindings len = %d, want 4", len(bindings))
	}
}

func TestBind_ReturnsBindingWithDescription(t *testing.T) {
	t.Parallel()
	app := appForKeybindings(t)
	b := app.bind(ActQuit, "quit the app")
	if b.Description != "quit the app" {
		t.Errorf("description = %q, want %q", b.Description, "quit the app")
	}
	if b.Key == "" {
		t.Error("key should not be empty for ActQuit")
	}
}

func TestHelpBarItems_GlobalInEveryPane(t *testing.T) {
	t.Parallel()
	app := appForKeybindings(t)
	app.issuesList.SetIssues([]jira.Issue{{Key: "A1", Parent: &jira.Issue{Key: "P1"}, Subtasks: []jira.Issue{{Key: "S1"}}}})
	want := app.helpBarItems()
	for _, focus := range []focusPanel{focusIssueTabs, focusIssues, focusInfo, focusProjects} {
		app.leftFocus = focus
		if got := app.helpBarItems(); !reflect.DeepEqual(got, want) {
			t.Errorf("focus %v: help bar = %v, want %v", focus, got, want)
		}
	}
}

func TestFilteredHelpSections_LocalFirst(t *testing.T) {
	t.Parallel()
	app := appForKeybindings(t)
	app.side = sideLeft
	app.leftFocus = focusIssues

	bindings, localN := app.filteredHelpSections()
	if localN == 0 || localN >= len(bindings) {
		t.Fatalf("localN = %d of %d, want both sections", localN, len(bindings))
	}
	if bindings[0].Description != "open issue detail" {
		t.Errorf("first binding = %+v, want the pane's own", bindings[0])
	}
	if !slices.Contains(bindings[:localN], Binding{"backspace", "show parent"}) {
		t.Error("show parent missing from the Local section")
	}
	if !slices.Contains(bindings[localN:], Binding{"q", "quit"}) {
		t.Error("quit missing from the Global section")
	}

	app.helpFilter = "quit"
	if _, localN := app.filteredHelpSections(); localN != 0 {
		t.Errorf("localN = %d with filter quit, want 0", localN)
	}
}

func TestHelpScrollOffset(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name                              string
		offset, cursorLine, height, total int
		want                              int
	}{
		{"fits", 0, 5, 20, 10, 0},
		{"cursor near bottom scrolls", 0, 19, 20, 40, 1},
		{"cursor up keeps two lines above", 10, 11, 20, 40, 9},
		{"stays put inside window", 5, 12, 20, 40, 5},
		{"clamps to end", 30, 39, 20, 40, 20},
	}
	for _, tc := range cases {
		if got := helpScrollOffset(tc.offset, tc.cursorLine, tc.height, tc.total); got != tc.want {
			t.Errorf("%s: offset = %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestContextBindings_SkipsUnboundActions(t *testing.T) {
	t.Parallel()
	app := appForKeybindings(t)
	app.side = sideLeft
	app.leftFocus = focusIssues
	for _, b := range app.ContextBindings() {
		if b.Key == "" {
			t.Errorf("binding %q has no key", b.Description)
		}
	}
}

func TestTitleForNavFrame_ChildrenNamesParent(t *testing.T) {
	t.Parallel()
	frame := navstack.NavFrame{Source: navstack.SourceFromList, ParentKey: "PLAT-1"}
	if got := titleForNavFrame(frame); got != "Children of PLAT-1" {
		t.Errorf("title = %q, want %q", got, "Children of PLAT-1")
	}
}
