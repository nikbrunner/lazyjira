package tui

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nikbrunner/lazyjira/v2/pkg/jira"
	"github.com/nikbrunner/lazyjira/v2/pkg/jira/jiratest"
	"github.com/nikbrunner/lazyjira/v2/pkg/tui/components"
)

func TestOlderSprintRequestCannotReplaceCurrentLoadingModal(t *testing.T) {
	for _, oldRequestFails := range []bool{false, true} {
		name := "success"
		if oldRequestFails {
			name = "error"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			boardCalls := 0
			fake := &jiratest.FakeClient{T: t}
			fake.GetBoardsFunc = func(context.Context) ([]jira.Board, error) {
				boardCalls++
				if oldRequestFails && boardCalls == 1 {
					return nil, errors.New("old request failed")
				}
				return nil, nil
			}
			app := newAppWithFake(t, fake)
			app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})

			oldCmd := app.startSprintFetch(sprintPickerTarget{issueKey: testKey})
			app.startSprintFetch(sprintPickerTarget{issueKey: testKey})
			currentModalID := app.sprintLoadingModalID
			oldResult := sprintLoadedFromCmd(t, oldCmd)
			_, _ = app.handleSprintsLoaded(oldResult)

			if !app.modal.IsVisible() || app.modal.Title() != "Loading sprints" || app.modal.Generation() != currentModalID {
				t.Fatal("older response changed the current loading modal")
			}
			if app.sprintLoadingID != app.sprintFetchID {
				t.Fatal("older response invalidated the current request")
			}
		})
	}
}

func TestExpiredBoardCacheFetchesBoardsAgain(t *testing.T) {
	t.Parallel()
	boardCalls := 0
	fake := &jiratest.FakeClient{T: t}
	fake.GetBoardsFunc = func(context.Context) ([]jira.Board, error) {
		boardCalls++
		return nil, nil
	}
	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})
	start := time.Date(2026, 9, 25, 10, 0, 0, 0, time.UTC)
	app.boardsCache.now = func() time.Time { return start }
	app.boardsCache.set("all", []jira.Board{{ID: 1}})
	start = start.Add(app.boardsCache.ttl)

	cmd := app.startSprintFetch(sprintPickerTarget{issueKey: testKey})
	if cmd == nil {
		t.Fatal("expired boards should trigger a fetch")
	}
	loaded := sprintLoadedFromCmd(t, cmd)
	_, _ = app.handleSprintsLoaded(loaded)
	if boardCalls != 1 {
		t.Fatalf("GetBoards calls = %d, want 1", boardCalls)
	}
}

func TestSuccessfulEmptySprintResponseIsCached(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetBoardsFunc = func(context.Context) ([]jira.Board, error) { return nil, nil }
	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})

	cmd := app.startSprintFetch(sprintPickerTarget{issueKey: testKey})
	loaded := sprintLoadedFromCmd(t, cmd)
	_, _ = app.handleSprintsLoaded(loaded)
	options, ok := app.sprintsCache.get("all")
	if !ok || len(options) != 0 {
		t.Fatalf("cached options = (%v, %v), want an empty cached result", options, ok)
	}
}

func TestPartialSprintResponseIsNotCached(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetBoardsFunc = func(context.Context) ([]jira.Board, error) {
		return []jira.Board{{ID: 1, Name: "First", Type: "scrum"}, {ID: 2, Name: "Second", Type: "scrum"}}, nil
	}
	fake.GetSprintsFunc = func(_ context.Context, boardID int) ([]jira.Sprint, error) {
		if boardID == 1 {
			return []jira.Sprint{{ID: 8, Name: "Partial", State: "active"}}, nil
		}
		return nil, errors.New("second board failed")
	}
	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})

	cmd := app.startSprintFetch(sprintPickerTarget{issueKey: testKey})
	loaded := sprintLoadedFromCmd(t, cmd)
	if len(loaded.options) != 1 || loaded.err == nil {
		t.Fatalf("fetch result = options:%d err:%v, want partial options and an error", len(loaded.options), loaded.err)
	}
	_, _ = app.handleSprintsLoaded(loaded)
	if _, ok := app.sprintsCache.get("all"); ok {
		t.Fatal("partial sprint response was cached")
	}
}

func TestSprintRequestInvalidationClosesOnlyOwnedLoadingModal(t *testing.T) {
	t.Run("project change closes owned modal", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		app.demoMode = true
		app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})
		app.startSprintFetch(sprintPickerTarget{issueKey: testKey})
		app.selectProject(&jira.Project{Key: "OTHER"})
		if app.modal.IsVisible() {
			t.Fatal("project change left the loading modal open")
		}
	})

	t.Run("project change preserves a replacement modal", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		app.demoMode = true
		app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})
		app.startSprintFetch(sprintPickerTarget{issueKey: testKey})
		app.modal.Show("Another picker", []components.ModalItem{{Label: "Keep me"}})
		app.selectProject(&jira.Project{Key: "OTHER"})
		if !app.modal.IsVisible() || app.modal.Title() != "Another picker" {
			t.Fatal("project change closed a modal it did not own")
		}
	})

	t.Run("create form cancellation closes owned modal", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		app.createForm = components.NewCreateForm(nil)
		app.createForm.ShowForm(nil, "Task", testProject)
		app.startSprintFetch(sprintPickerTarget{createForm: true})
		app.Update(components.CreateFormCancelMsg{})
		if app.modal.IsVisible() {
			t.Fatal("create-form cancellation left the loading modal open")
		}
	})

	t.Run("metadata completion closes owned modal", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		app.createForm = components.NewCreateForm(nil)
		app.createForm.ShowForm(nil, "Task", testProject)
		app.createCtx = createCtx{projectKey: testProject, issueTypeID: "1", issueTypeName: "Task"}
		app.startSprintFetch(sprintPickerTarget{createForm: true})
		app.handleCreateMetaLoaded(createMetaLoadedMsg{projectKey: testProject, issueTypeID: "1", cacheVersion: app.referenceCacheVersion})
		if app.modal.IsVisible() || !app.createForm.IsVisible() {
			t.Fatal("metadata completion did not replace its owned loading modal with the form")
		}
	})
}

func TestReferenceRefreshRejectsPreRefreshSprintResponse(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetBoardsFunc = func(context.Context) ([]jira.Board, error) {
		return []jira.Board{{ID: 1, Name: "Cloud", Type: "scrum"}}, nil
	}
	fake.GetSprintsFunc = func(context.Context, int) ([]jira.Sprint, error) {
		return []jira.Sprint{{ID: 8, Name: "Sprint 8", State: "active"}}, nil
	}
	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})
	cmd := app.startSprintFetch(sprintPickerTarget{issueKey: testKey})
	loaded := sprintLoadedFromCmd(t, cmd)
	app.invalidateReferenceCaches()
	_, _ = app.handleSprintsLoaded(loaded)

	if _, ok := app.boardsCache.get("all"); ok {
		t.Fatal("pre-refresh response repopulated the board cache")
	}
	if _, ok := app.sprintsCache.get("all"); ok {
		t.Fatal("pre-refresh response repopulated the sprint cache")
	}
	if app.modal.IsVisible() {
		t.Fatal("pre-refresh response left its loading modal open")
	}
}
