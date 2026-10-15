package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/nikbrunner/lazyjira/pkg/internal/testkit"
)

func TestViewWindowTitle(t *testing.T) {
	t.Parallel()

	t.Run("names the active project", func(t *testing.T) {
		t.Parallel()
		app := focusApp(t)
		app.projectKey = "PLAT"
		testkit.AssertEqual(t, "title", app.View().WindowTitle, "lazyjira · PLAT")
	})

	t.Run("plain without a project", func(t *testing.T) {
		t.Parallel()
		app := focusApp(t)
		app.projectKey = ""
		testkit.AssertEqual(t, "title", app.View().WindowTitle, "lazyjira")
	})
}

func TestViewProgressBar(t *testing.T) {
	t.Parallel()

	t.Run("nil when idle", func(t *testing.T) {
		t.Parallel()
		app := focusApp(t)
		if pb := app.View().ProgressBar; pb != nil {
			t.Fatalf("idle ProgressBar = %+v, want nil", pb)
		}
	})

	t.Run("indeterminate while a modal loads, nil once it closes", func(t *testing.T) {
		t.Parallel()
		app := focusApp(t)
		_ = app.modal.ShowLoading("Loading", "Loading…")
		pb := app.View().ProgressBar
		if pb == nil || pb.State != tea.ProgressBarIndeterminate {
			t.Fatalf("ProgressBar = %+v, want indeterminate", pb)
		}
		app.modal.Hide()
		if pb := app.View().ProgressBar; pb != nil {
			t.Fatalf("ProgressBar after hide = %+v, want nil", pb)
		}
	})

	t.Run("create form clears after an error ends loading", func(t *testing.T) {
		t.Parallel()
		app := focusApp(t)
		app.createForm.SetLoading(true)
		if pb := app.View().ProgressBar; pb == nil {
			t.Fatal("loading form should set ProgressBar")
		}
		app.createForm.SetError("boom")
		if pb := app.View().ProgressBar; pb != nil {
			t.Fatalf("ProgressBar after error = %+v, want nil", pb)
		}
	})

	t.Run("create form clears when hidden while loading", func(t *testing.T) {
		t.Parallel()
		app := focusApp(t)
		app.createForm.SetLoading(true)
		app.createForm.Hide()
		if pb := app.View().ProgressBar; pb != nil {
			t.Fatalf("ProgressBar after hide = %+v, want nil", pb)
		}
	})

	t.Run("jql modal clears after a search error", func(t *testing.T) {
		t.Parallel()
		app := focusApp(t)
		app.jqlModal.Show("", nil)
		app.jqlModal.SetLoading(true)
		if pb := app.View().ProgressBar; pb == nil {
			t.Fatal("loading JQL modal should set ProgressBar")
		}
		_, _ = app.handleJQLSearchError(jqlSearchErrorMsg{err: "bad jql"})
		if pb := app.View().ProgressBar; pb != nil {
			t.Fatalf("ProgressBar after error = %+v, want nil", pb)
		}
	})

	t.Run("jql modal clears when hidden while loading", func(t *testing.T) {
		t.Parallel()
		app := focusApp(t)
		app.jqlModal.Show("", nil)
		app.jqlModal.SetLoading(true)
		app.jqlModal.Hide()
		if pb := app.View().ProgressBar; pb != nil {
			t.Fatalf("ProgressBar after hide = %+v, want nil", pb)
		}
	})
}
