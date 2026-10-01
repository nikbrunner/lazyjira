package tui

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nikbrunner/lazyjira/pkg/internal/testkit"
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

	if err := app.prefillDescriptionMarkdown(fields); err != nil {
		t.Fatal(err)
	}

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

func typeMeta(extra ...string) []jira.CreateMetaField {
	fields := make([]jira.CreateMetaField, 0, 3+len(extra))
	fields = append(fields,
		jira.CreateMetaField{FieldID: "summary", Name: "Summary", Required: true, Schema: jira.CreateMetaSchema{Type: "string", System: "summary"}},
		jira.CreateMetaField{FieldID: fldDescription, Name: "Description", Schema: jira.CreateMetaSchema{Type: "string", System: "description"}},
		jira.CreateMetaField{FieldID: fldPriority, Name: "Priority", Schema: jira.CreateMetaSchema{Type: "priority", System: "priority"}},
	)
	for _, id := range extra {
		fields = append(fields, jira.CreateMetaField{FieldID: id, Name: id, Schema: jira.CreateMetaSchema{Type: "string"}})
	}
	return fields
}

// newCreateTypeApp serves Sub-task, Bug, and Story. Bug and Story each have
// one field the other lacks. metaErr fails createmeta for a type ID.
func newCreateTypeApp(t *testing.T, metaErr map[string]error) *App {
	t.Helper()
	fake := &jiratest.FakeClient{T: t}
	fake.GetIssueTypesFunc = func(context.Context, string) ([]jira.IssueType, error) {
		return []jira.IssueType{
			{ID: "5", Name: "Sub-task", Subtask: true},
			{ID: "1", Name: "Bug"},
			{ID: "2", Name: "Story"},
		}, nil
	}
	fake.GetCreateMetaFunc = func(_ context.Context, _, typeID string) ([]jira.CreateMetaField, error) {
		if err := metaErr[typeID]; err != nil {
			return nil, err
		}
		switch typeID {
		case "1":
			return typeMeta("customfield_bug"), nil
		case "2":
			return typeMeta("customfield_story"), nil
		}
		return typeMeta(), nil
	}
	app := newAppWithFake(t, fake)
	app.projectKey = testProject
	app.projectID = "10000"
	app.usersCache.set(testProject, []jira.User{})
	app.createForm = components.NewCreateForm()
	return app
}

// drive runs a chain of single-message commands through Update.
func drive(app *App, cmd tea.Cmd) {
	for cmd != nil {
		_, cmd = app.Update(cmd())
	}
}

func typeRow(t *testing.T, app *App) *components.CreateFormField {
	t.Helper()
	f := app.createForm.FieldAt(0)
	if f == nil || f.FieldID != fldIssueType {
		t.Fatalf("first field = %+v, want the Type row", f)
	}
	return f
}

func hasField(app *App, id string) bool {
	for i := 0; ; i++ {
		f := app.createForm.FieldAt(i)
		if f == nil {
			return false
		}
		if f.FieldID == id {
			return true
		}
	}
}

func TestCreateIssue_OpensFormWithFirstNonSubtaskType(t *testing.T) {
	t.Parallel()
	app := newCreateTypeApp(t, nil)

	_, cmd := app.startCreateIssue()
	drive(app, cmd)

	if !app.createForm.IsVisible() || app.modal.IsVisible() {
		t.Fatal("n should open the form directly, without a type picker")
	}
	if got := typeRow(t, app).DisplayValue; got != "Bug" {
		t.Errorf("Type = %q, want Bug", got)
	}
	if app.createCtx.issueTypeID != "1" || !hasField(app, "customfield_bug") {
		t.Errorf("form should carry Bug's createmeta, ctx = %+v", app.createCtx)
	}
}

func TestCreateIssue_DefaultsToLastSubmittedTypePerProject(t *testing.T) {
	t.Parallel()
	app := newCreateTypeApp(t, nil)
	app.createCtx = createCtx{projectKey: testProject, issueTypeID: "2"}
	_, _ = app.handleCreateFormSubmit(components.CreateFormSubmitMsg{Fields: map[string]any{}})

	_, cmd := app.startCreateIssue()
	drive(app, cmd)
	if got := typeRow(t, app).DisplayValue; got != "Story" {
		t.Errorf("Type in %s = %q, want the last used Story", testProject, got)
	}

	app.createForm.Hide()
	app.createCtx = createCtx{intent: true, projectKey: "OTHER"}
	_, cmd = app.handleIssueTypesLoaded(issueTypesLoadedMsg{issueTypes: []jira.IssueType{{ID: "1", Name: "Bug"}, {ID: "2", Name: "Story"}}})
	if cmd == nil || app.createCtx.loadingTypeID != "1" {
		t.Errorf("another project should default to its first type, loadingTypeID = %q", app.createCtx.loadingTypeID)
	}
}

func TestCreateForm_TypeChangeKeepsTextAndMatchingFields(t *testing.T) {
	t.Parallel()
	app := newCreateTypeApp(t, nil)
	_, cmd := app.startCreateIssue()
	drive(app, cmd)
	app.createForm.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	app.createForm.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Crash on login")})
	app.createForm.SetDescriptionText("steps")
	for i := 0; app.createForm.FieldAt(i) != nil; i++ {
		switch app.createForm.FieldAt(i).FieldID {
		case fldPriority:
			app.createForm.SetFieldValue(i, map[string]string{"id": "3"}, "High")
		case "customfield_bug":
			app.createForm.SetFieldValue(i, "v1", "v1")
		}
	}

	drive(app, func() tea.Msg { return components.CreateFormTypeSelectedMsg{TypeID: "2", TypeName: "Story"} })

	if got := typeRow(t, app).DisplayValue; got != "Story" {
		t.Errorf("Type = %q, want Story", got)
	}
	if hasField(app, "customfield_bug") || !hasField(app, "customfield_story") {
		t.Error("fields should follow Story's createmeta")
	}
	cmd, _ = app.createForm.Intercept(tea.KeyMsg{Type: tea.KeyCtrlS})
	submit, ok := cmd().(components.CreateFormSubmitMsg)
	if !ok {
		t.Fatalf("ctrl+s = %T, want submit", cmd())
	}
	want := map[string]any{"summary": "Crash on login", fldDescription: "steps", fldPriority: map[string]string{"id": "3"}}
	if !reflect.DeepEqual(submit.Fields, want) {
		t.Errorf("submitted %#v, want %#v", submit.Fields, want)
	}
}

func TestCreateForm_TypeChangeIgnoresStaleResponse(t *testing.T) {
	t.Parallel()
	app := newCreateTypeApp(t, nil)
	_, cmd := app.startCreateIssue()
	drive(app, cmd)

	_, storyCmd := app.Update(components.CreateFormTypeSelectedMsg{TypeID: "2", TypeName: "Story"})
	stale := storyCmd()
	_, _ = app.Update(components.CreateFormTypeSelectedMsg{TypeID: "5", TypeName: "Task"})
	_, _ = app.Update(stale)

	if got := typeRow(t, app).DisplayValue; got != "Bug" {
		t.Errorf("Type = %q, want Bug until the latest request resolves", got)
	}
	if app.createCtx.loadingTypeID != "5" {
		t.Errorf("loadingTypeID = %q, want the latest request", app.createCtx.loadingTypeID)
	}
}

func TestCreateForm_TypeChangeErrorKeepsPreviousType(t *testing.T) {
	t.Parallel()
	app := newCreateTypeApp(t, map[string]error{"2": errors.New("createmeta unavailable")})
	_, cmd := app.startCreateIssue()
	drive(app, cmd)

	drive(app, func() tea.Msg { return components.CreateFormTypeSelectedMsg{TypeID: "2", TypeName: "Story"} })

	if !app.createForm.IsVisible() || typeRow(t, app).DisplayValue != "Bug" || !hasField(app, "customfield_bug") {
		t.Fatal("a failed type change should keep the form on Bug")
	}
	if app.createCtx.issueTypeID != "1" || app.createCtx.loadingTypeID != "" {
		t.Errorf("ctx = %+v, want Bug and nothing loading", app.createCtx)
	}
	out := app.createForm.Render(testkit.BlankCanvas(120, 40), 120, 40)
	if !strings.Contains(out, "createmeta unavailable") {
		t.Error("the form should show the error")
	}
}

func TestCreateForm_TypePickerOffersFilteredTypes(t *testing.T) {
	t.Parallel()
	app := newCreateTypeApp(t, nil)
	_, cmd := app.startCreateIssue()
	drive(app, cmd)
	row := typeRow(t, app)

	_, _ = app.handleCreateFormPicker(components.CreateFormPickerMsg{FieldIndex: 0, Items: row.AllowedValues})

	if !app.modal.IsVisible() || app.onSelect == nil {
		t.Fatal("editing Type should open the type picker")
	}
	want := []components.ModalItem{{ID: "1", Label: "Bug"}, {ID: "2", Label: "Story"}}
	if !reflect.DeepEqual(row.AllowedValues, want) {
		t.Errorf("choices = %v, want %v", row.AllowedValues, want)
	}
	if app.onSelect(components.ModalItem{ID: "1", Label: "Bug"}) != nil {
		t.Error("picking the current type should do nothing")
	}
	sel := app.onSelect(components.ModalItem{ID: "2", Label: "Story"})
	if msg, ok := sel().(components.CreateFormTypeSelectedMsg); !ok || msg.TypeID != "2" {
		t.Errorf("picking Story = %#v, want a type change", sel())
	}
}

func TestCreateSubtask_TypeRowIsReadOnly(t *testing.T) {
	t.Parallel()
	app := newCreateTypeApp(t, nil)
	app.createCtx = createCtx{intent: true, projectKey: testProject, parentKey: testKey}
	drive(app, fetchIssueTypes(app.client, "10000"))

	row := typeRow(t, app)
	if row.DisplayValue != "Sub-task" || len(row.AllowedValues) != 0 {
		t.Fatalf("Type row = %+v, want Sub-task without choices", row)
	}
	_, cmd := app.handleCreateFormPicker(components.CreateFormPickerMsg{FieldIndex: 0})
	if cmd != nil || app.modal.IsVisible() {
		t.Error("editing the subtask Type should do nothing")
	}
	_, _ = app.handleCreateFormSubmit(components.CreateFormSubmitMsg{Fields: map[string]any{}})
	if _, ok := app.lastCreateType[testProject]; ok {
		t.Error("a subtask type must not become the project's default")
	}
}

func TestDuplicateIssue_DefaultsToSourceType(t *testing.T) {
	t.Parallel()
	app := newCreateTypeApp(t, nil)
	app.createCtx = createCtx{intent: true, projectKey: testProject, duplicateFrom: &jira.Issue{
		Summary:   "Original",
		IssueType: &jira.IssueType{ID: "2", Name: "Story"},
	}}
	drive(app, fetchIssueTypes(app.client, "10000"))

	if got := typeRow(t, app).DisplayValue; got != "Story" {
		t.Errorf("Type = %q, want the source's Story", got)
	}
}
