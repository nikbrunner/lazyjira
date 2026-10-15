package tui

import (
	"reflect"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/nikbrunner/lazyjira/pkg/config"
	"github.com/nikbrunner/lazyjira/pkg/tui/theme"
	"github.com/nikbrunner/lazyjira/pkg/tui/views"
)

func TestViewReportsFocus(t *testing.T) {
	t.Parallel()
	if !focusApp(t).View().ReportFocus {
		t.Error("View should enable focus reporting")
	}
}

func TestFocusRequestsBackgroundColor(t *testing.T) {
	t.Parallel()
	_, cmd := focusApp(t).Update(tea.FocusMsg{})
	if cmd == nil || reflect.ValueOf(cmd).Pointer() != reflect.ValueOf(tea.RequestBackgroundColor).Pointer() {
		t.Error("FocusMsg should return tea.RequestBackgroundColor")
	}
}

//nolint:paralleltest // Switches the process-wide theme.
func TestBackgroundColorMsgSwitchesTheme(t *testing.T) {
	t.Cleanup(func() { theme.Init(theme.Options{}) })
	theme.Init(theme.Options{Background: lipgloss.Color("#20262c")})
	app := focusApp(t)

	app.Update(tea.BackgroundColorMsg{Color: lipgloss.Color("#e0e0e0")})
	if theme.IsDark() {
		t.Fatal("theme stayed dark after a light background answer")
	}
	if theme.ColorHighlight != lipgloss.Color("#cecece") {
		t.Errorf("highlight = %v, want #cecece", theme.ColorHighlight)
	}
}

func TestGlamourRendererKeepsAutoStyle(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{Renderer: config.RendererGlamour, RendererStyle: config.RendererStyleAuto}
	if got := newADFRenderer(cfg); got != (views.GlamourRenderer{Style: config.RendererStyleAuto}) {
		t.Errorf("renderer = %#v, want auto resolved per render", got)
	}
}
