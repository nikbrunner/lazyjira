package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nikbrunner/lazyjira/pkg/config"
	"github.com/nikbrunner/lazyjira/pkg/jira"
	"github.com/nikbrunner/lazyjira/pkg/jira/jiratest"
	"github.com/nikbrunner/lazyjira/pkg/tui/components"
)

func TestApplyParentEdit_SetSendsKey(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.UpdateIssueFunc = func(_ context.Context, _ string, _ map[string]any) error { return nil }

	app := newAppWithFake(t, fake)
	app.issueCache[mainKey] = &jira.Issue{Key: mainKey}

	cmd := app.applyParentEdit(mainKey, "NEW-1")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	_ = cmd()

	if got := app.issueCache[mainKey].Parent; got == nil || got.Key != "NEW-1" {
		t.Errorf("optimistic Parent = %+v, want Key=NEW-1", got)
	}
	if len(fake.UpdateIssueCalls) != 1 {
		t.Fatalf("UpdateIssue calls = %d, want 1", len(fake.UpdateIssueCalls))
	}
	parent, _ := fake.UpdateIssueCalls[0].Fields["parent"].(map[string]string)
	if parent["key"] != "NEW-1" {
		t.Errorf("UpdateIssue fields.parent.key = %v, want NEW-1", parent)
	}
}

func TestApplyParentEdit_EmptyCallsRemove(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.RemoveIssueParentFunc = func(_ context.Context, _ string) error { return nil }

	app := newAppWithFake(t, fake)
	app.issueCache[mainKey] = &jira.Issue{Key: mainKey, Parent: &jira.Issue{Key: "OLD-1"}}

	cmd := app.applyParentEdit(mainKey, "")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	_ = cmd()

	if got := app.issueCache[mainKey].Parent; got != nil {
		t.Errorf("optimistic Parent = %+v, want nil", got)
	}
	if len(fake.RemoveIssueParentCalls) != 1 {
		t.Errorf("RemoveIssueParent calls = %d, want 1", len(fake.RemoveIssueParentCalls))
	}
	if len(fake.UpdateIssueCalls) != 0 {
		t.Errorf("UpdateIssue should not be called for unset; got %d calls", len(fake.UpdateIssueCalls))
	}
}

func TestApplyParentEdit_InvalidKeyShortCircuits(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	app.issueCache[mainKey] = &jira.Issue{Key: mainKey}

	cmd := app.applyParentEdit(mainKey, "not-a-key")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	emsg, ok := msg.(errorMsg)
	if !ok {
		t.Fatalf("msg = %T, want errorMsg", msg)
	}
	if emsg.err == nil || !strings.Contains(emsg.err.Error(), "invalid parent key") {
		t.Errorf("err = %v, want 'invalid parent key ...'", emsg.err)
	}
	if len(fake.UpdateIssueCalls)+len(fake.RemoveIssueParentCalls) != 0 {
		t.Errorf("no client call expected on validation failure")
	}
}

func newParentLookupApp(t *testing.T, fake *jiratest.FakeClient) *App {
	t.Helper()
	app := editFlowApp(t, fake)
	app.overlays = components.OverlayStack{&app.issueLookup}
	app.projectKey = testProject
	issue := &jira.Issue{Key: mainKey, IssueType: &jira.IssueType{ID: "10001", Name: "Story"}, Parent: &jira.Issue{Key: "OLD-1"}}
	app.issueCache[mainKey] = issue
	selectInfoField(t, app, issue, config.FieldConfig{ID: "parent"})
	return app
}

func sendLookupKey(app *App, key tea.KeyMsg) {
	_, cmd := app.Update(key)
	if cmd == nil {
		return
	}
	if msg := cmd(); msg != nil {
		_, cmd = app.Update(msg)
		if cmd != nil {
			_ = cmd()
		}
	}
}

func TestEditParent_OpensLookupWithProjectPrefix(t *testing.T) {
	t.Parallel()
	app := newParentLookupApp(t, &jiratest.FakeClient{T: t})

	_, cmd := app.editInfoField(app.issueCache[mainKey])

	if !app.issueLookup.IsVisible() || app.issueLookup.Query() != testProject+"-" {
		t.Fatalf("lookup visible = %v, query = %q, want %s-", app.issueLookup.IsVisible(), app.issueLookup.Query(), testProject)
	}
	if app.inputModal.IsVisible() {
		t.Error("the plain input modal should stay closed")
	}
	if cmd == nil {
		t.Error("expected a suggestion query for the prefilled key")
	}
}

func TestEditParent_SuggestionSetsParent(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.UpdateIssueFunc = func(context.Context, string, map[string]any) error { return nil }
	app := newParentLookupApp(t, fake)
	_, _ = app.editInfoField(app.issueCache[mainKey])
	app.handleIssueSuggestionsLoaded(issueSuggestionsLoadedMsg{
		query:       testProject + "-",
		suggestions: []jira.IssueSuggestion{{Key: "PLAT-205", Summary: "Release notes"}},
	})

	sendLookupKey(app, tea.KeyMsg{Type: tea.KeyDown})
	sendLookupKey(app, tea.KeyMsg{Type: tea.KeyEnter})

	if len(fake.UpdateIssueCalls) != 1 {
		t.Fatalf("UpdateIssue calls = %d, want 1", len(fake.UpdateIssueCalls))
	}
	parent, _ := fake.UpdateIssueCalls[0].Fields["parent"].(map[string]string)
	if parent["key"] != "PLAT-205" {
		t.Errorf("parent key = %v, want PLAT-205", parent)
	}
}

func TestEditParent_TypedKeyIsAccepted(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.SuggestIssuesFunc = func(context.Context, string) ([]jira.IssueSuggestion, error) { return nil, nil }
	fake.UpdateIssueFunc = func(context.Context, string, map[string]any) error { return nil }
	app := newParentLookupApp(t, fake)
	_, _ = app.editInfoField(app.issueCache[mainKey])

	sendLookupKey(app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("9")})
	sendLookupKey(app, tea.KeyMsg{Type: tea.KeyEnter})

	if len(fake.UpdateIssueCalls) != 1 {
		t.Fatalf("UpdateIssue calls = %d, want 1", len(fake.UpdateIssueCalls))
	}
	parent, _ := fake.UpdateIssueCalls[0].Fields["parent"].(map[string]string)
	if parent["key"] != testProject+"-9" {
		t.Errorf("parent key = %v, want %s-9", parent, testProject)
	}
}

func TestEditParent_EmptyPromptClearsParent(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.RemoveIssueParentFunc = func(context.Context, string) error { return nil }
	app := newParentLookupApp(t, fake)
	_, _ = app.editInfoField(app.issueCache[mainKey])

	sendLookupKey(app, tea.KeyMsg{Type: tea.KeyCtrlU})
	sendLookupKey(app, tea.KeyMsg{Type: tea.KeyEnter})

	if len(fake.RemoveIssueParentCalls) != 1 {
		t.Fatalf("RemoveIssueParent calls = %d, want 1", len(fake.RemoveIssueParentCalls))
	}
}

func TestEditParent_EscCancelsWithoutAPICall(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newParentLookupApp(t, fake)
	_, _ = app.editInfoField(app.issueCache[mainKey])

	sendLookupKey(app, tea.KeyMsg{Type: tea.KeyEsc})

	if app.issueLookup.IsVisible() || app.parentLookupFor != "" {
		t.Errorf("lookup visible = %v, parentLookupFor = %q, want closed and cleared", app.issueLookup.IsVisible(), app.parentLookupFor)
	}
	if len(fake.UpdateIssueCalls)+len(fake.RemoveIssueParentCalls) != 0 {
		t.Error("esc must not call the API")
	}
	if got := app.issueCache[mainKey].Parent; got == nil || got.Key != "OLD-1" {
		t.Errorf("Parent = %+v, want unchanged OLD-1", got)
	}
}
