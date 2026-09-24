package tui

import (
	"context"
	"testing"

	"github.com/nikbrunner/lazyjira/v2/pkg/config"
	"github.com/nikbrunner/lazyjira/v2/pkg/jira"
	"github.com/nikbrunner/lazyjira/v2/pkg/jira/jiratest"
	"github.com/nikbrunner/lazyjira/v2/pkg/tui/views"
)

func setupPreviewedSub(t *testing.T, fake *jiratest.FakeClient, sub *jira.Issue) *App {
	t.Helper()
	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{Key: mainKey, Summary: "main summary"}})
	app.previewKey = subKey1
	app.issueCache[subKey1] = sub
	return app
}

func TestEditAction_TargetsPreviewedIssue(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{Key: mainKey, Summary: "main summary"}})
	app.previewKey = subKey1
	app.issueCache[subKey1] = &jira.Issue{Key: subKey1, Summary: "sub summary"}
	app.side = sideLeft
	app.leftFocus = focusIssues

	_, _ = app.handleActionEdit()

	if got := app.editContext.issueKey; got != subKey1 {
		t.Errorf("editContext.issueKey = %q, want %q", got, subKey1)
	}
}

func TestCustomCommand_TargetsPreviewedIssue(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{Key: mainKey, Summary: "main"}})
	app.previewKey = subKey1
	app.issueCache[subKey1] = &jira.Issue{Key: subKey1, Summary: "sub"}

	rc := resolvedCommand(t, "x", "test", "echo {{.Key}}", config.CtxIssues)

	data, ok := app.buildCommandData(rc)
	if !ok {
		t.Fatal("buildCommandData returned ok=false")
	}
	if data["Key"] != subKey1 {
		t.Errorf("data[Key] = %q, want %q", data["Key"], subKey1)
	}
}

func TestCurrentIssue_StubWhenPreviewKeyUncached(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{Key: mainKey}})
	app.previewKey = subKey1

	cur := app.currentIssue()
	if cur == nil {
		t.Fatal("currentIssue() returned nil with previewKey set")
	}
	if cur.Key != subKey1 {
		t.Errorf("currentIssue().Key = %q, want %q (stub for previewKey)", cur.Key, subKey1)
	}
}

func TestCurrentIssue_FallsBackToListWhenNoPreview(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{Key: mainKey}})

	cur := app.currentIssue()
	if cur == nil {
		t.Fatal("currentIssue() returned nil with list selection present")
	}
	if cur.Key != mainKey {
		t.Errorf("currentIssue().Key = %q, want %q", cur.Key, mainKey)
	}
}

func TestEditAction_OnInfoSubTab_EditsPreviewedIssueSummary(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := setupPreviewedSub(t, fake, &jira.Issue{Key: subKey1, Summary: "sub summary"})
	app.side = sideLeft
	app.leftFocus = focusInfo
	main := &jira.Issue{Key: mainKey, Subtasks: []jira.Issue{{Key: subKey1}}}
	app.infoPanel.SetIssue(main)
	app.infoPanel.NextTab()
	app.infoPanel.NextTab()

	_, _ = app.handleActionEdit()

	if got := app.editContext.issueKey; got != subKey1 {
		t.Errorf("editContext.issueKey = %q, want %q", got, subKey1)
	}
	if app.editContext.kind != editSummary {
		t.Errorf("editContext.kind = %v, want editSummary", app.editContext.kind)
	}
}

func TestEditAction_Description_TargetsPreviewedIssue(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := setupPreviewedSub(t, fake, &jira.Issue{Key: subKey1, Description: "sub desc"})
	app.side = sideRight
	app.detailView.SetActiveTab(views.TabDetails)

	_, _ = app.handleActionEdit()

	if got := app.editContext.issueKey; got != subKey1 {
		t.Errorf("editContext.issueKey = %q, want %q", got, subKey1)
	}
	if app.editContext.kind != editDesc {
		t.Errorf("editContext.kind = %v, want editDesc", app.editContext.kind)
	}
}

func TestTransitionAction_TargetsPreviewedIssue(t *testing.T) {
	t.Parallel()
	var calledKey string
	fake := &jiratest.FakeClient{T: t}
	fake.GetTransitionsFunc = func(_ context.Context, key string) ([]jira.Transition, error) {
		calledKey = key
		return nil, nil
	}
	app := setupPreviewedSub(t, fake, &jira.Issue{Key: subKey1})

	_, cmd, handled := app.handleIssueAction(ActTransition)
	if !handled {
		t.Fatal("ActTransition not handled")
	}
	if cmd == nil {
		t.Fatal("expected cmd, got nil")
	}
	cmd()

	if calledKey != subKey1 {
		t.Errorf("GetTransitions called with %q, want %q", calledKey, subKey1)
	}
}

func TestAssigneeAction_TargetsPreviewedIssue(t *testing.T) {
	t.Parallel()
	var calledProject string
	fake := &jiratest.FakeClient{T: t}
	fake.GetUsersFunc = func(_ context.Context, projectKey string) ([]jira.User, error) {
		calledProject = projectKey
		return nil, nil
	}
	app := setupPreviewedSub(t, fake, &jira.Issue{Key: subKey1})
	app.projectKey = "SUB"

	_, cmd, handled := app.handleIssueAction(ActAssignee)
	if !handled {
		t.Fatal("ActAssignee not handled")
	}
	if cmd == nil {
		t.Fatal("expected cmd, got nil")
	}
	cmd()

	if app.onSelect == nil {
		t.Error("onSelect was not installed (handler needs a previewed issue)")
	}
	if calledProject != "SUB" {
		t.Errorf("GetUsers called for project %q, want SUB", calledProject)
	}
}

func TestCommentsAction_TargetsPreviewedIssue(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := setupPreviewedSub(t, fake, &jira.Issue{Key: subKey1, Summary: "cached"})
	app.detailView.SetIssue(app.issueCache[subKey1])

	_, cmd, handled := app.handleIssueAction(ActComments)
	if !handled {
		t.Fatal("ActComments not handled")
	}
	if cmd != nil {
		cmd()
	}
	if app.detailView.ActiveTab() != views.TabComments {
		t.Errorf("detailView tab = %v, want Comments", app.detailView.ActiveTab())
	}
	if app.detailView.IssueKey() != subKey1 {
		t.Errorf("detailView.IssueKey() = %q, want %q", app.detailView.IssueKey(), subKey1)
	}
}

func TestNewCommentAction_TargetsPreviewedIssue(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := setupPreviewedSub(t, fake, &jira.Issue{Key: subKey1})
	app.side = sideRight
	app.detailView.SetActiveTab(views.TabComments)

	_, _, handled := app.handleIssueAction(ActNew)
	if !handled {
		t.Fatal("ActNew not handled")
	}
	if got := app.editContext.issueKey; got != subKey1 {
		t.Errorf("editContext.issueKey = %q, want %q", got, subKey1)
	}
	if app.editContext.kind != editCommentNew {
		t.Errorf("editContext.kind = %v, want editCommentNew", app.editContext.kind)
	}
}

func TestDuplicateIssueAction_TargetsPreviewedIssue(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetIssueTypesFunc = func(_ context.Context, _ string) ([]jira.IssueType, error) {
		return nil, nil
	}
	app := setupPreviewedSub(t, fake, &jira.Issue{Key: subKey1, Summary: "sub summary"})
	app.side = sideLeft
	app.leftFocus = focusIssues
	app.projectKey = "SUB"

	_, _, handled := app.handleIssueAction(ActDuplicateIssue)
	if !handled {
		t.Fatal("ActDuplicateIssue not handled")
	}
	if app.createCtx.duplicateFrom == nil {
		t.Fatal("duplicateFrom not set")
	}
	if got := app.createCtx.duplicateFrom.Key; got != subKey1 {
		t.Errorf("duplicateFrom.Key = %q, want %q", got, subKey1)
	}
}
