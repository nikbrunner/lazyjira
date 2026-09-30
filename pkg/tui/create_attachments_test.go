package tui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/nikbrunner/lazyjira/pkg/internal/testkit"
	"github.com/nikbrunner/lazyjira/pkg/jira"
	"github.com/nikbrunner/lazyjira/pkg/jira/jiratest"
	"github.com/nikbrunner/lazyjira/pkg/tui/components"
)

func pasteClipboard(t *testing.T, read func() (string, error)) *App {
	t.Helper()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.readClipboardImage = read
	app.createForm = formWithFields([]components.CreateFormField{{FieldID: "summary"}, {FieldID: fldDescription}})
	app.createForm.SetSize(120, 40)
	drive(app, func() tea.Msg { return components.CreateFormPasteImageMsg{} })
	return app
}

func formText(app *App) string {
	return ansi.Strip(app.createForm.Render(testkit.BlankCanvas(120, 40), 120, 40))
}

func TestClipboardPaste_AttachesImage(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "lazyjira-paste-1.png")
	if err := os.WriteFile(path, make([]byte, 2048), 0o600); err != nil {
		t.Fatal(err)
	}

	app := pasteClipboard(t, func() (string, error) { return path, nil })

	if got := app.createForm.DescriptionText(); got != "[Image #1]" {
		t.Errorf("description = %q, want the image token", got)
	}
	if !strings.Contains(formText(app), "[Image #1] image-1.png  2 KB") {
		t.Errorf("attachment line missing:\n%s", formText(app))
	}
}

func TestClipboardPaste_ReportsNoImageAndMissingTool(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		err  error
		want string
	}{
		{errNoClipboardImage, "No image in clipboard"},
		{errors.New("image paste needs xclip, which was not found"), "image paste needs xclip, which was not found"},
	} {
		app := pasteClipboard(t, func() (string, error) { return "", tc.err })
		if app.createForm.DescriptionText() != "" {
			t.Errorf("%v: description should stay empty", tc.err)
		}
		if !strings.Contains(formText(app), tc.want) {
			t.Errorf("form should show %q:\n%s", tc.want, formText(app))
		}
	}
}

func TestUploadAttachments_UploadsInOrderAndReportsPartialFailure(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	temp := filepath.Join(dir, "clip.png")
	own := filepath.Join(dir, "own.png")
	_ = os.WriteFile(temp, []byte("CLIP"), 0o600)
	_ = os.WriteFile(own, []byte("OWN"), 0o600)
	fake := &jiratest.FakeClient{T: t}
	fake.AddAttachmentFunc = func(_ context.Context, _, filename string, _ []byte) error {
		if filename == "own.png" {
			return errors.New("413 too large")
		}
		return nil
	}
	fake.GetIssueFunc = func(context.Context, string) (*jira.Issue, error) { return &jira.Issue{Key: "PLAT-7"}, nil }
	app := newAppWithFake(t, fake)

	msg := uploadAttachments(fake, "PLAT-7", "Created", []components.CreateAttachment{
		{Path: temp, Name: "image-1.png", Temp: true},
		{Path: own, Name: "own.png"},
	})().(attachmentsUploadedMsg)

	if len(fake.AddAttachmentCalls) != 2 || fake.AddAttachmentCalls[0].Filename != "image-1.png" ||
		string(fake.AddAttachmentCalls[0].Data) != "CLIP" || fake.AddAttachmentCalls[1].Key != "PLAT-7" {
		t.Fatalf("calls = %+v, want both files uploaded in order", fake.AddAttachmentCalls)
	}
	if _, err := os.Stat(temp); !os.IsNotExist(err) {
		t.Error("clipboard temp file should be deleted after upload")
	}
	if _, err := os.Stat(own); err != nil {
		t.Error("the user's own file must stay")
	}

	_, cmd := app.handleAttachmentsUploaded(msg)
	want := "Created PLAT-7; 1 of 2 images failed to upload: 413 too large"
	if got := app.statusPanel.ErrorMessage(); got != want {
		t.Errorf("status = %q, want %q", got, want)
	}
	if app.createForm.IsVisible() {
		t.Error("an upload failure must not reopen the form")
	}
	if !app.modal.IsVisible() {
		t.Error("the failure should stay visible in an error dialog")
	}
	if cmd == nil {
		t.Error("uploads should refresh the issue detail")
	}
}

func TestHandleIssueCreated_UploadsSubmittedImagesBeforeRefresh(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "x.png")
	_ = os.WriteFile(path, []byte("X"), 0o600)
	fake := &jiratest.FakeClient{T: t}
	fake.AddAttachmentFunc = func(context.Context, string, string, []byte) error { return nil }
	app := newAppWithFake(t, fake)
	app.createCtx = createCtx{projectKey: testProject, issueTypeID: "1"}
	_, _ = app.handleCreateFormSubmit(components.CreateFormSubmitMsg{
		Fields:      map[string]any{},
		Attachments: []components.CreateAttachment{{Path: path, Name: "x.png"}},
	})

	_, cmd := app.handleIssueCreated(issueCreatedMsg{issue: &jira.Issue{Key: "PLAT-1"}})

	msg, ok := cmd().(attachmentsUploadedMsg)
	if !ok || msg.issueKey != "PLAT-1" || msg.total != 1 || len(msg.errs) != 0 {
		t.Fatalf("created issue should upload first, got %#v", msg)
	}
	if !*app.logFlag {
		t.Error("uploads should be logged")
	}
}

func TestClipboardPaste_DroppedWhileSubmitting(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "late.png")
	_ = os.WriteFile(path, []byte("X"), 0o600)
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.createForm = formWithFields([]components.CreateFormField{{FieldID: fldDescription}})
	app.createForm.SetLoading(true)

	_, _ = app.handleClipboardImage(clipboardImageMsg{path: path})
	_, _ = app.handleClipboardImage(clipboardImageMsg{err: errNoClipboardImage})

	if !app.createForm.IsLoading() || app.createForm.DescriptionText() != "" {
		t.Error("a late clipboard result must leave the submitting form alone")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("a dropped clipboard file should be deleted")
	}
}

func TestHandleCreateFormSubmit_EscapesImageTokensOnServer(t *testing.T) {
	t.Parallel()
	got := submitCreateDescription(t, false, "see [Image #1] and [Image #12], keep [link|http://x]")
	if got != `see \[Image #1\] and \[Image #12\], keep [link|http://x]` {
		t.Errorf("description = %q, want only image tokens escaped", got)
	}
	cloud := submitCreateDescription(t, true, "see [Image #1]")
	if m, _ := cloud.(map[string]any); m["md"] != "see [Image #1]" {
		t.Errorf("cloud description = %#v, want the token untouched", cloud)
	}
}
