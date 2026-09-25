package tui

import (
	"context"
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nikbrunner/lazyjira/v2/pkg/jira"
	"github.com/nikbrunner/lazyjira/v2/pkg/jira/jiratest"
	"github.com/nikbrunner/lazyjira/v2/pkg/tui/components"
)

func sprintLoadedFromCmd(t *testing.T, cmd tea.Cmd) sprintsLoadedMsg {
	t.Helper()
	batch, ok := cmd().(tea.BatchMsg)
	if !ok {
		t.Fatal("sprint fetch command did not return a batch")
	}
	for _, child := range batch {
		if loaded, ok := child().(sprintsLoadedMsg); ok {
			return loaded
		}
	}
	t.Fatal("sprint fetch batch did not return a sprint result")
	return sprintsLoadedMsg{}
}

func TestSprintPickerLoadingIsReadOnlyAndCancellable(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetBoardsFunc = func(context.Context) ([]jira.Board, error) { return nil, nil }
	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})

	cmd := app.startSprintFetch(sprintPickerTarget{issueKey: testKey})
	if !app.modal.IsVisible() || app.modal.Title() != "Loading sprints" {
		t.Fatal("picker should immediately show the loading modal")
	}
	if app.onSelect != nil {
		t.Fatal("loading modal must not have a selection callback")
	}

	updated, cancel := app.modal.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app.modal = updated
	if cancel == nil {
		t.Fatal("Enter in a read-only loading modal should cancel, not select")
	}
	if _, ok := cancel().(components.ModalCancelledMsg); !ok {
		t.Fatal("loading modal Enter should emit ModalCancelledMsg")
	}
	app.handleModalCancelled()

	loaded := sprintLoadedFromCmd(t, cmd)
	_, _ = app.handleSprintsLoaded(loaded)
	if app.modal.IsVisible() {
		t.Fatal("cancelled request reopened a picker")
	}
}

func TestExpiredSprintOptionsShowLoadingInsteadOfStaleChoices(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetBoardsFunc = func(context.Context) ([]jira.Board, error) { return nil, nil }
	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})
	start := time.Date(2026, 9, 25, 10, 0, 0, 0, time.UTC)
	app.sprintsCache.now = func() time.Time { return start }
	app.sprintsCache.set("all", []sprintOption{{sprint: jira.Sprint{ID: 1, Name: "Expired", State: "future"}}})
	start = start.Add(app.sprintsCache.ttl)

	cmd := app.startSprintFetch(sprintPickerTarget{issueKey: testKey})
	if cmd == nil || !app.modal.IsVisible() || app.modal.Title() != "Loading sprints" {
		t.Fatal("expired options should show a loading state and fetch again")
	}
	if app.onSelect != nil {
		t.Fatal("expired options must not remain selectable during refresh")
	}
}

func TestSprintPickerDoesNotReplaceSameTitleModal(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetBoardsFunc = func(context.Context) ([]jira.Board, error) { return nil, nil }
	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})

	cmd := app.startSprintFetch(sprintPickerTarget{issueKey: testKey})
	app.modal.ShowReadOnly("Loading sprints", []components.ModalItem{{Label: "Replacement loading state"}})
	replacementGeneration := app.modal.Generation()
	app.onSelect = func(components.ModalItem) tea.Cmd { return nil }
	loaded := sprintLoadedFromCmd(t, cmd)
	_, _ = app.handleSprintsLoaded(loaded)

	if !app.modal.IsVisible() || app.modal.Title() != "Loading sprints" || app.modal.Generation() != replacementGeneration {
		t.Fatal("late sprint result replaced a same-title modal")
	}
	if app.onSelect == nil {
		t.Fatal("late sprint result cleared the replacement modal callback")
	}
}

func TestSprintCacheHitAndRefreshAllBypassesTTL(t *testing.T) {
	t.Parallel()
	boardCalls, sprintCalls := 0, 0
	fake := &jiratest.FakeClient{T: t}
	fake.GetBoardsFunc = func(context.Context) ([]jira.Board, error) {
		boardCalls++
		return []jira.Board{{ID: 2043, Name: "Cloud Platform Sprints", Type: "scrum"}}, nil
	}
	fake.GetSprintsFunc = func(context.Context, int) ([]jira.Sprint, error) {
		sprintCalls++
		return []jira.Sprint{{ID: 8, Name: "Sprint 8", State: "active"}}, nil
	}
	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})

	cmd := app.startSprintFetch(sprintPickerTarget{issueKey: testKey})
	if !app.modal.IsVisible() || app.modal.Title() != "Loading sprints" {
		t.Fatal("cache miss should show loading immediately")
	}
	loaded := sprintLoadedFromCmd(t, cmd)
	_, _ = app.handleSprintsLoaded(loaded)
	if boardCalls != 1 || sprintCalls != 1 {
		t.Fatalf("first fetch calls = boards:%d sprints:%d, want 1 each", boardCalls, sprintCalls)
	}

	app.modal.Hide()
	app.handleModalCancelled()
	if cmd := app.startSprintFetch(sprintPickerTarget{issueKey: testKey}); cmd != nil {
		t.Fatal("fresh cache hit should not start a network command")
	}
	if !app.modal.IsVisible() || app.modal.Title() != "Sprint: "+testKey {
		t.Fatal("fresh cached options should appear immediately")
	}
	if boardCalls != 1 || sprintCalls != 1 {
		t.Fatalf("cache hit calls = boards:%d sprints:%d, want 1 each", boardCalls, sprintCalls)
	}

	app.modal.Hide()
	app.handleModalCancelled()
	app.boardsCache.set("all", []jira.Board{{ID: 9}})
	app.usersCache.set(testProject, []jira.User{{AccountID: "u1"}})
	app.createMetaCache.set(testProject+":1", []jira.CreateMetaField{{FieldID: "summary"}})
	app.keymap = DefaultKeymap()
	app.overlays = components.OverlayStack{&app.modal}
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}})
	if _, ok := app.boardsCache.get("all"); ok {
		t.Fatal("Refresh All did not clear board cache")
	}
	if _, ok := app.sprintsCache.get("all"); ok {
		t.Fatal("Refresh All did not clear sprint cache")
	}
	if _, ok := app.usersCache.get(testProject); ok {
		t.Fatal("Refresh All did not clear user cache")
	}
	if _, ok := app.createMetaCache.get(testProject + ":1"); ok {
		t.Fatal("Refresh All did not clear create metadata cache")
	}

	cmd = app.startSprintFetch(sprintPickerTarget{issueKey: testKey})
	if cmd == nil {
		t.Fatal("Refresh All should make the next sprint lookup fetch again")
	}
	loaded = sprintLoadedFromCmd(t, cmd)
	_, _ = app.handleSprintsLoaded(loaded)
	if boardCalls != 2 || sprintCalls != 2 {
		t.Fatalf("post-refresh calls = boards:%d sprints:%d, want 2 each", boardCalls, sprintCalls)
	}
}

func TestRefreshAllDoesNotInterceptSearchOrEditorTyping(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.boardsCache.set("all", []jira.Board{{ID: 1}})
	app.searchBar = components.NewSearchBar()
	app.searchBar.Activate()
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}})
	if app.searchBar.Query() != "R" {
		t.Fatalf("search query = %q, want R", app.searchBar.Query())
	}
	if _, ok := app.boardsCache.get("all"); !ok {
		t.Fatal("Refresh All intercepted search typing")
	}

	app.searchBar.Deactivate()
	app.inputModal = components.NewInputModal()
	app.inputModal.Show("Edit", "")
	app.overlays = components.OverlayStack{&app.inputModal, &app.modal}
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}})
	if _, ok := app.boardsCache.get("all"); !ok {
		t.Fatal("Refresh All intercepted editor input")
	}
}

func TestSprintFailureIsNotCachedAndCanRetry(t *testing.T) {
	t.Parallel()
	boardCalls, sprintCalls := 0, 0
	fake := &jiratest.FakeClient{T: t}
	fake.GetBoardsFunc = func(context.Context) ([]jira.Board, error) {
		boardCalls++
		return []jira.Board{{ID: 2043, Name: "Cloud", Type: "scrum"}}, nil
	}
	fake.GetSprintsFunc = func(context.Context, int) ([]jira.Sprint, error) {
		sprintCalls++
		if sprintCalls == 1 {
			return nil, errors.New("temporary failure")
		}
		return []jira.Sprint{{ID: 8, Name: "Sprint 8", State: "future"}}, nil
	}
	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})

	cmd := app.startSprintFetch(sprintPickerTarget{issueKey: testKey})
	loaded := sprintLoadedFromCmd(t, cmd)
	_, _ = app.handleSprintsLoaded(loaded)
	if _, ok := app.sprintsCache.get("all"); ok {
		t.Fatal("failed sprint fetch was cached as empty options")
	}
	if _, ok := app.boardsCache.get("all"); !ok {
		t.Fatal("successful board fetch should be cached despite sprint failure")
	}
	if !app.modal.IsVisible() || app.modal.Title() != "Sprint picker" {
		t.Fatal("sprint failure should remain visible in an error modal")
	}

	app.modal.Hide()
	app.handleModalCancelled()
	cmd = app.startSprintFetch(sprintPickerTarget{issueKey: testKey})
	if cmd == nil {
		t.Fatal("failed sprint fetch must be retried rather than treated as an empty cache hit")
	}
	loaded = sprintLoadedFromCmd(t, cmd)
	_, _ = app.handleSprintsLoaded(loaded)
	if boardCalls != 1 || sprintCalls != 2 {
		t.Fatalf("retry calls = boards:%d sprints:%d, want 1 and 2", boardCalls, sprintCalls)
	}
	if _, ok := app.sprintsCache.get("all"); !ok {
		t.Fatal("successful retry was not cached")
	}
}

func TestSprintLoadingTickRoutesThroughAppUpdate(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetBoardsFunc = func(context.Context) ([]jira.Board, error) { return nil, nil }
	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})
	app.modal.SetSize(80, 24)
	app.overlays = components.OverlayStack{&app.modal}

	cmd := app.startSprintFetch(sprintPickerTarget{issueKey: testKey})
	before := app.modal.View()
	loadingGeneration := app.modal.Generation()
	batch, ok := cmd().(tea.BatchMsg)
	if !ok {
		t.Fatal("sprint fetch should batch its request and animation")
	}
	var tick tea.Msg
	var loaded sprintsLoadedMsg
	gotLoaded := false
	for _, child := range batch {
		msg := child()
		switch msg := msg.(type) {
		case components.LoadingIndicatorTickMsg:
			tick = msg
		case sprintsLoadedMsg:
			loaded = msg
			gotLoaded = true
		}
	}
	if tick == nil || !gotLoaded {
		t.Fatal("loading batch did not return both the fetch result and an animation tick")
	}

	updated, nextTick := app.Update(tick)
	app = updated.(*App)
	if nextTick == nil || app.modal.View() == before || app.modal.Generation() != loadingGeneration {
		t.Fatal("App.Update did not advance the loading indicator without replacing the modal")
	}
	updated, _ = app.Update(loaded)
	app = updated.(*App)
	completedGeneration := app.modal.Generation()
	if completedGeneration == loadingGeneration || app.modal.Title() == "Loading sprints" {
		t.Fatal("fetch result did not replace the loading modal")
	}

	updated, _ = app.Update(nextTick())
	app = updated.(*App)
	if app.modal.Generation() != completedGeneration || app.modal.Title() == "Loading sprints" {
		t.Fatal("a pending tick restarted the completed loading modal")
	}
}
