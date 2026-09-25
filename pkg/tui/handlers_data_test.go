package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nikbrunner/lazyjira/v2/pkg/internal/testkit"
	"github.com/nikbrunner/lazyjira/v2/pkg/jira"
	"github.com/nikbrunner/lazyjira/v2/pkg/jira/jiratest"
	"github.com/nikbrunner/lazyjira/v2/pkg/tui/components"
)

func TestHandleTransitionsLoaded(t *testing.T) {
	t.Parallel()

	t.Run("shows modal and installs handler", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})

		_, cmd := app.handleTransitionsLoaded(transitionsLoadedMsg{
			issueKey:    testKey,
			transitions: []jira.Transition{{ID: "11", Name: "Done", To: &jira.Status{Name: "Closed"}}},
		})

		if cmd != nil {
			t.Errorf("expected nil cmd")
		}
		if !app.modal.IsVisible() {
			t.Error("transition modal should be visible")
		}
		if app.onSelect == nil {
			t.Error("onSelect should be set")
		}
	})

	t.Run("empty is noop", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		_, _ = app.handleTransitionsLoaded(transitionsLoadedMsg{issueKey: testKey})
		if app.modal.IsVisible() {
			t.Error("modal should stay hidden for no transitions")
		}
	})
}

func TestHandleSprintsLoaded(t *testing.T) {
	t.Parallel()

	t.Run("ignores results after create-form cancellation", func(t *testing.T) {
		t.Parallel()
		fake := &jiratest.FakeClient{T: t}
		fake.GetBoardsFunc = func(context.Context) ([]jira.Board, error) { return nil, nil }
		app := newAppWithFake(t, fake)
		app.createForm.ShowForm([]components.CreateFormField{{FieldID: "sprint", Name: "Sprint"}}, "Task", testProject)
		app.createForm.Pause()
		fetchCmd := app.startSprintFetch(sprintPickerTarget{createForm: true, fieldIndex: 0})
		loaded := fetchCmd().(sprintsLoadedMsg)

		app.createForm.Hide()
		updated, cancelCmd := app.Update(components.ModalCancelledMsg{})
		app = updated.(*App)
		if cancelCmd != nil {
			t.Fatal("cancelling the modal should not schedule another command")
		}
		app.createForm.ShowForm([]components.CreateFormField{{FieldID: "priority", Name: "Priority"}}, "Task", testProject)
		updated, sprintCmd := app.handleSprintsLoaded(loaded)
		app = updated.(*App)
		if sprintCmd != nil {
			t.Fatal("ignoring stale sprint results should not schedule another command")
		}

		if app.modal.IsVisible() {
			t.Error("late sprint result should not open a picker for a replacement form")
		}
	})

	t.Run("ignores results after issue selection changes", func(t *testing.T) {
		t.Parallel()
		fake := &jiratest.FakeClient{T: t}
		fake.GetBoardsFunc = func(context.Context) ([]jira.Board, error) { return nil, nil }
		app := newAppWithFake(t, fake)
		app.issuesList.SetIssues([]jira.Issue{{Key: testKey}, {Key: mainKey}})
		fetchCmd := app.startSprintFetch(sprintPickerTarget{issueKey: testKey})
		loaded := fetchCmd().(sprintsLoadedMsg)

		app.issuesList.SelectByKey(mainKey)
		updated, cmd := app.handleSprintsLoaded(loaded)
		app = updated.(*App)
		if cmd != nil {
			t.Fatal("ignoring results for another issue should not schedule another command")
		}

		if app.modal.IsVisible() {
			t.Error("late sprint result should not open a picker for a different issue")
		}
	})

	t.Run("labels sprint by board and keeps its raw name for issue updates", func(t *testing.T) {
		t.Parallel()
		fake := &jiratest.FakeClient{T: t}
		fake.MoveToSprintFunc = func(context.Context, int, string) error { return nil }
		app := newAppWithFake(t, fake)
		issue := &jira.Issue{Key: testKey}
		app.issuesList.SetIssues([]jira.Issue{*issue})
		app.issueCache[testKey] = issue

		_, _ = app.handleSprintsLoaded(sprintsLoadedMsg{
			target: sprintPickerTarget{issueKey: testKey},
			options: []sprintOption{{
				sprint:     jira.Sprint{ID: 1, Name: "Sprint 1", State: "active"},
				boardNames: []string{"Cloud Platform Sprints / CP"},
			}},
		})

		if !app.modal.IsVisible() {
			t.Fatal("sprint modal should be visible")
		}
		app.modal.SetSize(120, 30)
		if view := app.modal.View(); !strings.Contains(view, "Cloud Platform Sprints / CP") {
			t.Errorf("sprint modal does not identify its board: %q", view)
		}
		if app.onSelect == nil {
			t.Fatal("onSelect should be set")
		}
		cmd := app.onSelect(components.ModalItem{ID: "1", Label: "Sprint 1 (active) [Cloud Platform Sprints / CP]"})
		if cmd == nil {
			t.Fatal("selecting a sprint should move the issue")
		}
		cmd()
		if got := app.issueCache[testKey].Sprint.Name; got != "Sprint 1" {
			t.Errorf("cached sprint name = %q, want undecorated name", got)
		}
	})

	t.Run("fetch error is surfaced instead of showing None-only picker", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})

		_, _ = app.handleSprintsLoaded(sprintsLoadedMsg{
			target: sprintPickerTarget{issueKey: testKey},
			err:    errors.New("board does not support sprints"),
		})

		if !app.modal.IsVisible() || !strings.Contains(app.modal.View(), "board does not support sprints") {
			t.Error("fetch error should be visible in an error modal")
		}
		if app.onSelect != nil {
			t.Error("fetch failure should not leave a selection callback")
		}
	})

	t.Run("create form recovers from a fetch error", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		app.overlays = components.OverlayStack{&app.createForm}
		app.createForm.ShowForm([]components.CreateFormField{{FieldID: "sprint", Name: "Sprint"}}, "Task", testProject)
		app.createForm.Pause()

		updated, cmd := app.Update(sprintsLoadedMsg{
			target: sprintPickerTarget{createForm: true, fieldIndex: 0},
			err:    errors.New("board lookup failed"),
		})
		app = updated.(*App)
		if cmd != nil {
			t.Fatal("handling a sprint fetch error should not schedule another command")
		}

		if !app.createForm.IsVisible() {
			t.Fatal("create form should remain visible")
		}
		if app.modal.IsVisible() {
			t.Fatal("create-form errors should render inline, not open a modal behind the form")
		}
		app.createForm.SetSize(80, 24)
		view := app.createForm.Render(testkit.BlankCanvas(80, 24), 80, 24)
		if !strings.Contains(view, "board lookup failed") {
			t.Errorf("create form does not show fetch error: %q", view)
		}
		before := app.createForm.FocusedPanel()
		updated, _ = app.Update(tea.KeyMsg{Type: tea.KeyTab})
		app = updated.(*App)
		if app.createForm.FocusedPanel() == before {
			t.Error("create form should resume keyboard input after fetch failure")
		}
	})

	t.Run("create form picker works without selected issue", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		app.createForm.ShowForm([]components.CreateFormField{{FieldID: "sprint", Name: "Sprint"}}, "Task", testProject)
		app.createForm.Pause()

		_, _ = app.handleSprintsLoaded(sprintsLoadedMsg{
			target: sprintPickerTarget{createForm: true, fieldIndex: 0},
			options: []sprintOption{{
				sprint:     jira.Sprint{ID: 1, Name: "Sprint 1", State: "active"},
				boardNames: []string{"Cloud Platform Sprints / CP"},
			}},
		})

		if !app.modal.IsVisible() || app.onSelect == nil {
			t.Fatal("create-form sprint picker should open without a selected issue")
		}
		app.onSelect(components.ModalItem{ID: "1", Label: "Sprint 1 (active) [Cloud Platform Sprints / CP]"})
		if field := app.createForm.FieldAt(0); field == nil || field.DisplayValue != "Sprint 1 (active) [Cloud Platform Sprints / CP]" {
			t.Errorf("create sprint field = %#v", field)
		}
	})
}

func TestHandleLabelsLoaded_ShowsChecklist(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey, Labels: []string{"backend"}}})

	_, _ = app.handleLabelsLoaded(labelsLoadedMsg{labels: []string{"backend", "frontend"}})

	if !app.modal.IsVisible() {
		t.Error("labels checklist should be visible")
	}
}

func TestHandleComponentsLoaded_ShowsChecklist(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})

	_, _ = app.handleComponentsLoaded(componentsLoadedMsg{components: []jira.Component{{ID: "10", Name: "backend"}}})

	if !app.modal.IsVisible() {
		t.Error("components checklist should be visible")
	}
}

func TestHandleIssuePrefetched_CachesIssue(t *testing.T) {
	t.Parallel()

	t.Run("caches the issue", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})

		_, _ = app.handleIssuePrefetched(issuePrefetchedMsg{issue: &jira.Issue{Key: testKey}})

		if app.issueCache[testKey] == nil {
			t.Error("issue not cached")
		}
	})

	t.Run("nil issue is noop", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		_, _ = app.handleIssuePrefetched(issuePrefetchedMsg{})
		if len(app.issueCache) != 0 {
			t.Error("nil issue should not populate cache")
		}
	})
}

func TestHandleBatchPrefetched_CachesAll(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})

	_, _ = app.handleBatchPrefetched(batchPrefetchedMsg{issues: []jira.Issue{{Key: "A-1"}, {Key: "B-2"}}})

	if app.issueCache["A-1"] == nil || app.issueCache["B-2"] == nil {
		t.Errorf("batch not fully cached: %v", app.issueCache)
	}
}

func TestHandleTransitionDone(t *testing.T) {
	t.Parallel()

	t.Run("refetches for selected issue", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})

		_, cmd := app.handleTransitionDone()
		if cmd == nil {
			t.Error("expected refetch command after transition")
		}
	})

	t.Run("no selection is noop", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		_, cmd := app.handleTransitionDone()
		if cmd != nil {
			t.Error("expected nil cmd without a selected issue")
		}
	})
}

func TestHandleUsersLoaded_ShowsAssigneeModal(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.usersCache = map[string][]jira.User{}
	app.projectKey = testProject
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})

	_, _ = app.handleUsersLoaded(usersLoadedMsg{
		users:    []jira.User{{AccountID: "u1", DisplayName: "Ann"}},
		issueKey: testKey,
	})

	if !app.modal.IsVisible() {
		t.Error("assignee modal should be visible")
	}
	if len(app.usersCache[testProject]) != 1 {
		t.Errorf("users not cached: %v", app.usersCache)
	}
}

func TestBuildUserItems(t *testing.T) {
	t.Parallel()

	t.Run("prepends me and None and dedups self", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		app.currentUser = &jira.User{AccountID: "me", DisplayName: "Me"}

		items := app.buildUserItems([]jira.User{{AccountID: "me", DisplayName: "Me"}, {AccountID: "u1", DisplayName: "Ann"}})

		if len(items) != 3 {
			t.Fatalf("items = %d, want 3 (me, None, Ann)", len(items))
		}
		testkit.AssertEqual(t, "me label", items[0].Label, "Me (me)")
		testkit.AssertEqual(t, "None label", items[1].Label, "None")
		testkit.AssertEqual(t, "other label", items[2].Label, "Ann")
	})

	t.Run("without current user starts with None", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})

		items := app.buildUserItems([]jira.User{{AccountID: "u1", DisplayName: "Ann"}})

		if len(items) != 2 {
			t.Fatalf("items = %d, want 2 (None, Ann)", len(items))
		}
		testkit.AssertEqual(t, "first label", items[0].Label, "None")
	})
}
