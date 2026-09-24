package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nikbrunner/lazyjira/v2/pkg/internal/testkit"
	"github.com/nikbrunner/lazyjira/v2/pkg/jira/jiratest"
)

func TestHandleResize_SetsWidthAndHeight(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()

	_, _ = app.handleResize(tea.WindowSizeMsg{Width: 140, Height: 50})

	testkit.AssertEqual(t, "width", app.width, 140)
	testkit.AssertEqual(t, "height", app.height, 50)
}

func TestHandleResize_PanelsLaidOut(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()

	_, _ = app.handleResize(tea.WindowSizeMsg{Width: 120, Height: 40})

	layout := app.geometry()
	if layout.tabs.width == 0 {
		t.Error("Issue tabs pane should be laid out after resize")
	}
	if layout.detail.height == 0 {
		t.Error("Details pane should be laid out after resize")
	}
}

func TestHandleResize_NarrowTerminalShowsTooSmallLayout(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()

	_, _ = app.handleResize(tea.WindowSizeMsg{Width: 60, Height: 40})

	if !app.geometry().tooSmall {
		t.Error("60-column terminal should require the too-small message")
	}
}
