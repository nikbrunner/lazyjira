package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nikbrunner/lazyjira/v2/pkg/config"
	"github.com/nikbrunner/lazyjira/v2/pkg/jira"
	"github.com/nikbrunner/lazyjira/v2/pkg/jira/jiratest"
	"github.com/nikbrunner/lazyjira/v2/pkg/tui/components"
)

func newFilterPickerApp(t *testing.T) *App {
	t.Helper()
	a := newAppWithFake(t, &jiratest.FakeClient{T: t})
	a.keymap = DefaultKeymap()
	a.side = sideLeft
	a.leftFocus = focusIssues
	a.issuesList.SetTabs([]config.IssueTabConfig{{Name: "All"}})
	a.issuesList.SetIssues([]jira.Issue{
		{Key: "A-1", Summary: "one", Status: &jira.Status{Name: "To Do"}},
		{Key: "A-2", Summary: "two", Status: &jira.Status{Name: "In Progress"}, Subtasks: []jira.Issue{
			{Key: "S-1", Summary: "sub", Status: &jira.Status{Name: "Done"}},
		}},
		{Key: "A-3", Summary: "three", Status: &jira.Status{Name: "In Progress"}},
	})
	return a
}

func pressKey(a *App, key string) {
	if key == "esc" {
		a.handleKeyMsg(tea.KeyMsg{Type: tea.KeyEsc})
		return
	}
	a.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
}

func TestFilterPicker_AppliesCheckedValues(t *testing.T) {
	t.Parallel()
	a := newFilterPickerApp(t)

	pressKey(a, "f")
	if !a.modal.IsVisible() || !a.modal.IsChecklist() {
		t.Fatal("f should open the filter checklist")
	}
	var picked []components.ModalItem
	for _, item := range a.issuesList.FilterPickerItems(nil) {
		if item.ID == "status:In Progress" {
			picked = append(picked, item)
		}
	}
	if len(picked) != 1 {
		t.Fatal("picker items missing status:In Progress")
	}
	a.modal.Hide()
	a.handleChecklistConfirmed(components.ChecklistConfirmedMsg{Selected: picked})

	if got := a.issuesList.ItemCount(); got != 2 {
		t.Errorf("filtered issues = %d, want 2", got)
	}
}

func TestFilterPicker_EscClearsMarksThenTextThenPicker(t *testing.T) {
	t.Parallel()
	a := newFilterPickerApp(t)
	a.issuesList.SetPickerFilter(map[string]bool{"status:In Progress": true})
	a.issuesList.SetFilter("t")
	a.issuesList.ToggleMark()

	pressKey(a, "esc")
	if a.issuesList.HasMarks() || !a.issuesList.IsFiltered() || !a.issuesList.IsPickerFiltered() {
		t.Fatal("first esc should clear only marks")
	}
	pressKey(a, "esc")
	if a.issuesList.IsFiltered() || !a.issuesList.IsPickerFiltered() {
		t.Fatal("second esc should clear only the text filter")
	}
	pressKey(a, "esc")
	if a.issuesList.IsPickerFiltered() {
		t.Fatal("third esc should clear the picker filter")
	}
	if got := a.issuesList.ItemCount(); got != 3 {
		t.Errorf("issues after clearing = %d, want 3", got)
	}
}

func TestFilterPicker_ChildrenStartUnfilteredAndBackRestores(t *testing.T) {
	t.Parallel()
	a := newFilterPickerApp(t)
	a.issuesList.SetPickerFilter(map[string]bool{"status:In Progress": true})

	pressKey(a, ">")
	if !a.issuesList.IsHierarchyTab() {
		t.Fatal("> should open the children tab")
	}
	if a.issuesList.IsPickerFiltered() {
		t.Error("children tab kept the picker filter")
	}
	if got := a.issuesList.ItemCount(); got != 1 {
		t.Errorf("children shown = %d, want 1", got)
	}

	pressKey(a, "esc")
	if a.issuesList.IsHierarchyTab() {
		t.Fatal("esc should leave the children tab")
	}
	if !a.issuesList.IsPickerFiltered() {
		t.Error("picker filter lost after going back")
	}
	if sel := a.issuesList.SelectedIssue(); sel == nil || sel.Key != "A-2" {
		t.Errorf("SelectedIssue() after back = %+v, want A-2", sel)
	}
}

func TestIssuesLoaded_RecordsLoadedHint(t *testing.T) {
	t.Parallel()
	a := newFilterPickerApp(t)
	a.issuesList.SetSize(80, 10)
	a.handleIssuesLoaded(issuesLoadedMsg{
		issues:  []jira.Issue{{Key: "A-1", Summary: "one"}},
		tab:     0,
		total:   312,
		hasMore: true,
	})
	if view := a.issuesList.View(); !strings.Contains(view, "1/312 loaded") {
		t.Errorf("Issues panel missing loaded hint:\n%s", view)
	}
}

func TestFilterPicker_TitleCountsResultingIssues(t *testing.T) {
	t.Parallel()
	a := newFilterPickerApp(t)
	a.modal.SetSize(80, 24)

	pressKey(a, "f")
	if got := a.modal.Title(); got != "Filter issues · 3 of 3" {
		t.Errorf("title = %q, want Filter issues · 3 of 3", got)
	}

	a.modal, _ = a.modal.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
	if got := a.modal.Title(); got != "Filter issues · 2 of 3" {
		t.Errorf("title after checking In Progress = %q, want Filter issues · 2 of 3", got)
	}
}
