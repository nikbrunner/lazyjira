package tui

import (
	"context"
	"reflect"
	"testing"

	"github.com/nikbrunner/lazyjira/pkg/jira"
	"github.com/nikbrunner/lazyjira/pkg/jira/jiratest"
	"github.com/nikbrunner/lazyjira/pkg/tui/components"
)

func TestHandleCreateFormEditExternal_NilFieldIsNoop(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()

	_, cmd := app.handleCreateFormEditExternal(components.CreateFormEditExternalMsg{FieldIndex: 99})

	if cmd != nil {
		t.Error("expected nil cmd with nil field")
	}
}

func TestHandleCreateFormEditExternal_LaunchesEditorForDescription(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.projectKey = testProject
	app.createCtx = createCtx{
		intent:     true,
		projectKey: testProject,
	}
	form := components.NewCreateForm()
	form.ShowForm([]components.CreateFormField{
		{
			FieldID:      fldDescription,
			Name:         "Description",
			Type:         components.CFFieldMultiText,
			DisplayValue: "initial description",
		},
	}, "Story", testProject)
	app.createForm = form

	_, cmd := app.handleCreateFormEditExternal(components.CreateFormEditExternalMsg{FieldIndex: 0})

	if cmd == nil {
		t.Error("expected editor launch cmd for description field")
	}
}

// taggingConverter wraps Markdown so tests can tell converted values from raw ones.
type taggingConverter struct{}

func (taggingConverter) ToMarkdown(adf any) (string, any, error) {
	m, _ := adf.(map[string]any)
	md, _ := m["md"].(string)
	return md, "state-from-adf", nil
}

func (taggingConverter) FromMarkdown(md string, state any) (any, error) {
	return map[string]any{"md": md, "state": state}, nil
}

func submitCreateDescription(t *testing.T, isCloud bool, desc string) any {
	t.Helper()
	var created map[string]any
	fake := &jiratest.FakeClient{T: t, CreateIssueFunc: func(_ context.Context, fields map[string]any) (*jira.Issue, error) {
		created = fields
		return &jira.Issue{Key: "PLAT-1"}, nil
	}}
	app := newAppWithFake(t, fake)
	app.converter = taggingConverter{}
	app.isCloud = isCloud
	app.projectKey = testProject
	app.usersCache.set(testProject, []jira.User{soloUser()})
	app.createCtx = createCtx{projectKey: testProject, issueTypeID: "10001", descConvState: "prefill-state"}
	app.createForm = formWithFields([]components.CreateFormField{{FieldID: "summary"}, {FieldID: fldDescription}})

	_, cmd := app.handleCreateFormSubmit(components.CreateFormSubmitMsg{Fields: map[string]any{fldDescription: desc}})
	if cmd == nil {
		t.Fatal("expected create command")
	}
	if app.pendingMention != nil {
		t.Fatal("warm cache must not defer the submit")
	}
	cmd()
	return created[fldDescription]
}

func TestHandleCreateFormSubmit_CloudConvertsDescriptionWithMentions(t *testing.T) {
	t.Parallel()
	got := submitCreateDescription(t, true, "**hi** @Solo_One")
	want := map[string]any{"md": "**hi** [@Solo One](accountid:s1)", "state": "prefill-state"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("description = %#v, want %#v", got, want)
	}
}

func TestHandleCreateFormSubmit_ServerSendsRawDescription(t *testing.T) {
	t.Parallel()
	got := submitCreateDescription(t, false, "**hi** @Solo_One")
	if got != "**hi** @Solo_One" {
		t.Errorf("description = %#v, want the raw text", got)
	}
}

func TestPrefillDescriptionMarkdown_ConvertsADF(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.converter = taggingConverter{}
	fields := []components.CreateFormField{
		{FieldID: "summary"},
		{FieldID: fldDescription, DisplayValue: "plain text", Value: map[string]any{"md": "**rich**"}},
	}

	app.prefillDescriptionMarkdown(fields)

	if fields[1].DisplayValue != "**rich**" || fields[1].Value != "**rich**" {
		t.Errorf("description = %q / %#v, want Markdown", fields[1].DisplayValue, fields[1].Value)
	}
	if app.createCtx.descConvState != "state-from-adf" {
		t.Errorf("descConvState = %#v, want the converter state", app.createCtx.descConvState)
	}
}

func TestHandleCreateFormEditExternal_SeedsEditorFromTextarea(t *testing.T) {
	t.Setenv("EDITOR", "true")
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.createForm = formWithFields([]components.CreateFormField{{FieldID: fldDescription}})
	app.createForm.SetDescriptionText("typed in the form")

	_, cmd := app.handleCreateFormEditExternal(components.CreateFormEditExternalMsg{FieldIndex: 0})

	if cmd == nil {
		t.Fatal("expected editor launch cmd")
	}
	if app.editContext.kind != editCreateDesc {
		t.Errorf("editContext.kind = %v, want editCreateDesc", app.editContext.kind)
	}
}
