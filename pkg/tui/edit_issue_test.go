package tui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/nikbrunner/lazyjira/pkg/jira"
	"github.com/nikbrunner/lazyjira/pkg/jira/jiratest"
	"github.com/nikbrunner/lazyjira/pkg/tui/components"
	"github.com/nikbrunner/lazyjira/pkg/tui/views"
)

const (
	editKey = "PLAT-3"
	bugType = "Bug"
)

func editableIssue() *jira.Issue {
	return &jira.Issue{
		Key:         editKey,
		Summary:     "Login fails",
		Description: "old body",
		IssueType:   &jira.IssueType{ID: "1", Name: bugType},
		Priority:    &jira.Priority{ID: "2", Name: "High"},
	}
}

// newEditApp previews issue in the Issues list and serves Bug's createmeta.
func newEditApp(t *testing.T, issue *jira.Issue, updateErr error) (*App, *jiratest.FakeClient) {
	t.Helper()
	fake := &jiratest.FakeClient{T: t}
	stubFullIssueFetch(fake, issue)
	fake.GetCreateMetaFunc = func(context.Context, string, string) ([]jira.CreateMetaField, error) {
		return typeMeta("customfield_bug"), nil
	}
	fake.UpdateIssueFunc = func(context.Context, string, map[string]any) error { return updateErr }
	app := newAppWithFake(t, fake)
	app.converter = taggingConverter{}
	app.projectKey = testProject
	app.usersCache.set(testProject, []jira.User{soloUser()})
	app.createForm = components.NewCreateForm()
	app.createForm.SetSize(120, 40)
	app.issuesList.SetIssues([]jira.Issue{*issue})
	app.previewKey = issue.Key
	app.issueCache[issue.Key] = issue
	app.side = sideLeft
	app.leftFocus = focusIssues
	return app, fake
}

func openEdit(t *testing.T, app *App) {
	t.Helper()
	_, cmd := app.handleActionEdit()
	drive(app, cmd)
	if !app.createForm.IsVisible() || app.createForm.IsLoading() {
		t.Fatal("the Issue Edit View should be open")
	}
}

func save(app *App) {
	cmd, _ := app.createForm.Intercept(tea.KeyMsg{Type: tea.KeyCtrlS})
	drive(app, cmd)
}

func TestEditAction_OpensPrefilledEditView(t *testing.T) {
	t.Parallel()
	for _, pane := range []string{"issues", "details"} {
		app, _ := newEditApp(t, editableIssue(), nil)
		if pane == "details" {
			app.side = sideRight
			app.detailView.SetActiveTab(views.TabDetails)
		}
		openEdit(t, app)

		row := typeRow(t, app)
		if row.DisplayValue != bugType || len(row.AllowedValues) != 0 {
			t.Errorf("%s: Type row = %+v, want read-only Bug", pane, row)
		}
		got := app.createForm.Values()
		if got["summary"] != "Login fails" || got[fldDescription] != "old body" ||
			!reflect.DeepEqual(got[fldPriority], map[string]string{"id": "2"}) {
			t.Errorf("%s: prefill = %#v", pane, got)
		}
		if !strings.Contains(formText(app), "Edit PLAT-3 · *Summary") {
			t.Errorf("%s: title should name the edited issue:\n%s", pane, formText(app))
		}
	}
}

func TestEditSave_SendsOnlyChangedFields(t *testing.T) {
	t.Parallel()
	app, fake := newEditApp(t, editableIssue(), nil)
	openEdit(t, app)
	app.createForm.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	app.createForm.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" now")})

	save(app)

	if len(fake.UpdateIssueCalls) != 1 {
		t.Fatalf("UpdateIssue calls = %d, want 1", len(fake.UpdateIssueCalls))
	}
	call := fake.UpdateIssueCalls[0]
	want := map[string]any{"summary": "Login fails now"}
	if call.Key != editKey || !reflect.DeepEqual(call.Fields, want) {
		t.Errorf("update = %s %#v, want %#v", call.Key, call.Fields, want)
	}
	if app.createForm.IsVisible() {
		t.Error("a successful save should close the view")
	}
}

func TestEditSave_ConvertsDescription(t *testing.T) {
	t.Parallel()
	t.Run("cloud converts ADF with mentions and the prefill state", func(t *testing.T) {
		t.Parallel()
		issue := editableIssue()
		issue.DescriptionADF = map[string]any{"md": "old body"}
		app, fake := newEditApp(t, issue, nil)
		app.isCloud = true
		openEdit(t, app)
		app.createForm.SetDescriptionText("ping @Solo_One")

		save(app)

		want := map[string]any{fldDescription: map[string]any{"md": "ping [@Solo One](accountid:s1)", "state": "state-from-adf"}}
		if len(fake.UpdateIssueCalls) != 1 || !reflect.DeepEqual(fake.UpdateIssueCalls[0].Fields, want) {
			t.Errorf("update = %#v, want %#v", fake.UpdateIssueCalls, want)
		}
	})
	t.Run("server sends raw text with escaped image tokens", func(t *testing.T) {
		t.Parallel()
		app, fake := newEditApp(t, editableIssue(), nil)
		openEdit(t, app)
		app.createForm.SetDescriptionText("see [Image #1]")

		save(app)

		want := map[string]any{fldDescription: `see \[Image #1\]`}
		if len(fake.UpdateIssueCalls) != 1 || !reflect.DeepEqual(fake.UpdateIssueCalls[0].Fields, want) {
			t.Errorf("update = %#v, want %#v", fake.UpdateIssueCalls, want)
		}
	})
}

func TestEditSave_NothingChangedClosesWithoutRequest(t *testing.T) {
	t.Parallel()
	app, fake := newEditApp(t, editableIssue(), nil)
	openEdit(t, app)

	save(app)

	if len(fake.UpdateIssueCalls) != 0 || app.createForm.IsVisible() {
		t.Errorf("no-op save: %d updates, visible %v; want none and closed", len(fake.UpdateIssueCalls), app.createForm.IsVisible())
	}
}

func TestEditSave_UpdateErrorKeepsViewOpen(t *testing.T) {
	t.Parallel()
	app, _ := newEditApp(t, editableIssue(), errors.New("field priority is not on the screen"))
	openEdit(t, app)
	app.createForm.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	app.createForm.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("!")})

	save(app)

	if !app.createForm.IsVisible() || app.createForm.IsLoading() {
		t.Fatal("a failed update should keep the view open and editable")
	}
	if !strings.Contains(formText(app), "field priority is not on the screen") {
		t.Error("the update error should show inline")
	}
}

func TestEditSave_UploadsImagesAfterUpdate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	good := filepath.Join(dir, "good.png")
	bad := filepath.Join(dir, "bad.png")
	_ = os.WriteFile(good, []byte("G"), 0o600)
	_ = os.WriteFile(bad, []byte("B"), 0o600)
	app, fake := newEditApp(t, editableIssue(), nil)
	fake.AddAttachmentFunc = func(_ context.Context, key, filename string, _ []byte) error {
		if filename == "bad.png" {
			return errors.New("413 too large")
		}
		return nil
	}
	openEdit(t, app)
	app.createForm.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	app.createForm.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	app.createForm.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(good + " " + bad), Paste: true})

	save(app)

	if len(fake.UpdateIssueCalls) != 1 || !reflect.DeepEqual(fake.UpdateIssueCalls[0].Fields, map[string]any{fldDescription: `old body\[Image #1\] \[Image #2\]`}) {
		t.Errorf("update = %#v, want only the description with its tokens", fake.UpdateIssueCalls)
	}
	if len(fake.AddAttachmentCalls) != 2 || fake.AddAttachmentCalls[0].Key != editKey {
		t.Fatalf("uploads = %+v, want both images on %s", fake.AddAttachmentCalls, editKey)
	}
	app.modal.SetSize(120, 40)
	if view := ansi.Strip(app.modal.View()); !strings.Contains(view, "Updated PLAT-3; 1 of 2 images failed to upload") {
		t.Errorf("the failure dialog should outlive the refresh:\n%s", view)
	}
	if app.createForm.IsVisible() {
		t.Error("an upload failure must not reopen the view")
	}
}

func TestEditAction_ConversionFailureKeepsViewClosed(t *testing.T) {
	t.Parallel()
	issue := editableIssue()
	issue.DescriptionADF = map[string]any{"type": "doc"}
	app, _ := newEditApp(t, issue, nil)
	app.isCloud = true
	app.converter = failingConverter{}

	_, cmd := app.handleActionEdit()
	drive(app, cmd)

	if app.createForm.IsVisible() {
		t.Error("a description that can't become Markdown must not open for editing")
	}
	if app.statusPanel.ErrorMessage() == "" {
		t.Error("the conversion failure should show in the status panel")
	}
}

func TestCreateIssue_TitleNamesCreateAndProject(t *testing.T) {
	t.Parallel()
	app := newCreateTypeApp(t, nil)
	_, cmd := app.startCreateIssue()
	drive(app, cmd)
	app.createForm.SetSize(120, 40)

	if !strings.Contains(formText(app), "Create issue in "+testProject+" · *Summary") {
		t.Errorf("the view should say it creates an issue in the active project:\n%s", formText(app))
	}
}

func TestCreateIssue_HeaderShowsProjectAboveForm(t *testing.T) {
	t.Parallel()
	app := newCreateTypeApp(t, nil)
	app.projectList.SetProjects([]jira.Project{{Key: testProject, ID: "10000", Name: "Platform"}})
	app.overlays = components.OverlayStack{&app.createForm}
	_, _ = app.handleResize(tea.WindowSizeMsg{Width: 120, Height: 40})
	_, cmd := app.startCreateIssue()
	drive(app, cmd)

	lines := strings.Split(ansi.Strip(app.View()), "\n")
	header := strings.Join(lines[:3], "\n")
	if !strings.Contains(lines[0], "Project") || !strings.Contains(lines[1], testProject+" · Platform") {
		t.Errorf("the header should name the project:\n%s", header)
	}
	if strings.Contains(header, "↵") || strings.Contains(header, "[") {
		t.Errorf("the header should show no key hints inside the form:\n%s", header)
	}
	if !strings.Contains(lines[3], "Summary") {
		t.Errorf("the form should start below the header:\n%s", strings.Join(lines[:5], "\n"))
	}
}

func TestCreateHeader_NamesTargetProject(t *testing.T) {
	t.Parallel()
	app := newCreateTypeApp(t, nil)
	app.projectList.SetProjects([]jira.Project{{Key: testProject, ID: "10000"}})
	_, _ = app.handleResize(tea.WindowSizeMsg{Width: 120, Height: 40})
	app.createCtx.projectKey = "DSOTEST"

	if header := ansi.Strip(app.renderCreateHeader()); !strings.Contains(header, "DSOTEST") || strings.Contains(header, testProject) {
		t.Errorf("the header should name the subtask's project, not the active one:\n%s", header)
	}
}

func TestCreateHeader_FullWidthAfterResizeWhileMaximized(t *testing.T) {
	t.Parallel()
	app := newCreateTypeApp(t, nil)
	_, _ = app.handleResize(tea.WindowSizeMsg{Width: 120, Height: 40})
	app.toggleMaximize(focusIssues)
	_, _ = app.handleResize(tea.WindowSizeMsg{Width: 150, Height: 40})
	app.createCtx.projectKey = testProject

	for i, line := range strings.Split(app.renderCreateHeader(), "\n") {
		if w := ansi.StringWidth(line); w != 150 {
			t.Errorf("header line %d is %d wide, want 150", i, w)
		}
	}
}
