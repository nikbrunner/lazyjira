package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
	"github.com/textfuel/lazyjira/v2/pkg/tui/components"
	"github.com/textfuel/lazyjira/v2/pkg/tui/views"
)

const jiraProjectKey = "JIRA"

func TestProjectSelectorPickerConfirmAndCancel(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.width, app.height = 120, 40
	app.layoutPanels()
	app.overlays = components.OverlayStack{&app.projectPicker}
	app.issuesList.SetTabs(nil)
	app.projectList.SetProjects([]jira.Project{
		{Key: "OLD", ID: "1", Name: "Old Project"},
		{Key: jiraProjectKey, ID: "2", Name: "Jira Platform"},
	})

	// Direct focus key reaches the selector; Enter opens the picker and focus
	// remains there whether the overlay is confirmed or cancelled.
	_, _ = app.Update(runeKey('0'))
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !app.projectPicker.IsVisible() {
		t.Fatal("Enter on selector did not open picker")
	}
	_, _ = app.Update(runeKey('j'))
	pickerView := ansi.Strip(app.projectPicker.View())
	if !strings.Contains(pickerView, "Filter: j") || !strings.Contains(pickerView, jiraProjectKey+"  Jira Platform") {
		t.Fatalf("picker query/results = %q, want j/JIRA", pickerView)
	}
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyDown})
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter in picker should emit selection")
	}
	selection, ok := cmd().(components.ProjectPickerSelectedMsg)
	if !ok || selection.Project.Key != jiraProjectKey {
		t.Fatalf("picker command = %#v, want JIRA selection", selection)
	}
	_, _ = app.Update(selection)
	if app.projectKey != jiraProjectKey || app.leftFocus != focusProjects || app.projectPicker.IsVisible() {
		t.Fatalf("confirmed state: key=%q focus=%d visible=%v", app.projectKey, app.leftFocus, app.projectPicker.IsVisible())
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_, cmd = app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("Esc in picker should emit cancellation")
	}
	_, _ = app.Update(cmd())
	if app.projectKey != jiraProjectKey || app.leftFocus != focusProjects || app.projectPicker.IsVisible() {
		t.Fatalf("cancelled state: key=%q focus=%d visible=%v", app.projectKey, app.leftFocus, app.projectPicker.IsVisible())
	}
}

func TestProjectSelectorClickCancelsActiveSearch(t *testing.T) {
	t.Parallel()
	for _, finish := range []string{"confirm", "cancel"} {
		t.Run(finish, func(t *testing.T) {
			t.Parallel()
			app := appWithPanelDims(t, 120)
			app.keymap = DefaultKeymap()
			app.overlays = components.OverlayStack{&app.projectPicker}
			app.projectList.SetProjects([]jira.Project{{Key: jiraProjectKey, ID: "1", Name: "Jira Platform"}})
			app.issuesList.SetIssues([]jira.Issue{{Key: "OLD-1", Summary: "Alpha"}})
			app.side, app.leftFocus = sideLeft, focusIssues
			app.searchBar.Activate()
			_, cmd := app.Update(runeKey('z'))
			if cmd == nil {
				t.Fatal("typing into active search should emit its changed message")
			}
			_, _ = app.Update(cmd())
			if app.issuesList.SelectedIssue() != nil {
				t.Fatal("expected active search to filter out issue before selector click")
			}

			_, _ = app.Update(tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress, X: 5, Y: 0})
			if !app.projectPicker.IsVisible() || app.searchBar.IsActive() || app.searchBar.Query() != "" || app.issuesList.SelectedIssue() == nil {
				t.Fatalf("opening picker did not cancel search: picker=%v search=%v query=%q selected=%v", app.projectPicker.IsVisible(), app.searchBar.IsActive(), app.searchBar.Query(), app.issuesList.SelectedIssue())
			}
			_, _ = app.Update(runeKey('j'))
			if !strings.Contains(ansi.Strip(app.projectPicker.View()), "Filter: j") {
				t.Fatal("typing after selector click was not consumed by picker")
			}
			if finish == "confirm" {
				_, cmd = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
				if cmd == nil {
					t.Fatal("picker Enter did not emit selected project")
				}
				_, _ = app.Update(cmd())
				if app.projectKey != jiraProjectKey {
					t.Fatalf("projectKey=%q, want %q", app.projectKey, jiraProjectKey)
				}
			} else {
				_, cmd = app.Update(tea.KeyMsg{Type: tea.KeyEsc})
				if cmd == nil {
					t.Fatal("picker Esc did not emit cancellation")
				}
				_, _ = app.Update(cmd())
			}
			if app.projectPicker.IsVisible() || app.searchBar.IsActive() || app.searchBar.Query() != "" || app.issuesList.SelectedIssue() == nil {
				t.Fatalf("search unexpectedly resumed after picker close: picker=%v search=%v query=%q selected=%v", app.projectPicker.IsVisible(), app.searchBar.IsActive(), app.searchBar.Query(), app.issuesList.SelectedIssue())
			}
		})
	}
}

func TestProjectSelectorAndStatusRenderInAlignedHeader(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.width, app.height = 120, 40
	app.projectKey = jiraProjectKey
	app.projectList.SetProjects([]jira.Project{{Key: jiraProjectKey, Name: "Jira Platform"}})
	app.statusPanel = views.NewStatusPanel("", "user@example.com", "jira.example.com")
	app.statusPanel.SetAuthMethod("Environment")
	app.statusPanel.SetVersion("v1")
	app.statusPanel.SetSize(app.geometry().status.width, 3)
	app.projectList.SetActiveKey(app.projectKey)
	view := ansi.Strip(app.View())
	for _, want := range []string{"JIRA", "Jira Platf…", "Connected", "Acct:", "Host:", "Auth:", "v1"} {
		if !strings.Contains(view, want) {
			t.Errorf("header rendering missing %q in %q", want, view)
		}
	}
}

func TestProjectPickerConsumesWorkspaceMouseEvents(t *testing.T) {
	t.Parallel()
	app := appWithPanelDims(t, 120)
	app.overlays = components.OverlayStack{&app.projectPicker}
	app.projectPicker.Show([]components.ProjectChoice{{Key: "A", Name: "Alpha"}, {Key: "B", Name: "Beta"}})
	app.issuesList.SetIssues([]jira.Issue{{Key: "ISSUE-1"}, {Key: "ISSUE-2"}})
	app.issuesList.Cursor = 1
	app.projectList.SetProjects([]jira.Project{{Key: "A", Name: "Alpha"}, {Key: "B", Name: "Beta"}})
	app.projectList.Cursor = 1
	app.detailView.SetIssue(&jira.Issue{Key: "ISSUE-1"})
	app.side, app.leftFocus = sideRight, focusInfo

	for _, msg := range []tea.MouseMsg{
		{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress, X: 60, Y: 5},
		{Button: tea.MouseButtonWheelDown, Action: tea.MouseActionPress, X: 5, Y: 0},
	} {
		_, _ = app.Update(msg)
	}
	if app.side != sideRight || app.leftFocus != focusInfo || app.issuesList.Cursor != 1 || app.projectList.Cursor != 1 || app.detailView.IssueKey() != "ISSUE-1" || !app.projectPicker.IsVisible() {
		t.Fatalf("workspace mouse escaped picker: side=%v focus=%v issueCursor=%d projectCursor=%d detail=%q picker=%v", app.side, app.leftFocus, app.issuesList.Cursor, app.projectList.Cursor, app.detailView.IssueKey(), app.projectPicker.IsVisible())
	}
}

func TestProjectPickerMessagesAreOverlayOwned(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.overlays = components.OverlayStack{&app.projectPicker}
	app.projectPicker.Show([]components.ProjectChoice{{Key: "ABC", Name: "Alpha"}})
	cmd, handled := app.overlays.Intercept(runeKey('a'))
	if !handled || cmd != nil || !strings.Contains(ansi.Strip(app.projectPicker.View()), "Filter: a") {
		t.Fatalf("picker did not intercept key: handled=%v cmd=%v view=%q", handled, cmd, ansi.Strip(app.projectPicker.View()))
	}
}
