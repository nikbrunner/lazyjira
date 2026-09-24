package tui

import (
	"strings"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/tui/views"
)

func TestApp_SideWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		width      int
		configured int
		want       int
	}{
		{"default when unset", 200, 0, 22},
		{"configured width on wide terminal", 200, 50, 50},
		{"wide configured width is not capped at half", 200, 150, 150},
		{"caps on narrow terminal", 100, 40, 35},
		{"minimum sidebar width", 100, 10, 18},
		{"narrow terminal cap", 60, 40, 21},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			app := &App{width: tt.width, cfg: &config.Config{}, issuesList: views.NewIssuesList()}
			app.cfg.GUI.SidePanelWidth = tt.configured
			testkit.AssertEqual(t, "sideWidth", app.sideWidth(), tt.want)
		})
	}
}

func TestApp_MinimumWidthAccountsForNarrowSidebarCap(t *testing.T) {
	t.Parallel()
	issues := views.NewIssuesList()
	issues.SetFields([]string{"key", "status", "type", "priority", "summary", "assignee", "updated"})
	issues.SetIssues([]jira.Issue{{Key: strings.Repeat("A", 40)}})
	app := &App{height: 24, cfg: &config.Config{}, issuesList: issues}
	app.cfg.GUI.SidePanelWidth = 50
	if got := sideWidthForTerminal(50, 119); got != 41 {
		t.Fatalf("sidebar at 119 columns = %d, want 41", got)
	}
	if got := sideWidthForTerminal(50, 120); got != 50 {
		t.Fatalf("sidebar at 120 columns = %d, want 50", got)
	}
	required := 50 + issues.MinimumWidth()
	if got := app.minimumTerminalWidth(); got != required || got < 120 {
		t.Fatalf("minimum terminal width = %d, want wide-layout threshold %d", got, required)
	}
	for _, width := range []int{119, 120} {
		app.width = width
		if layout := app.geometry(); !layout.tooSmall || layout.requiredWidth != required {
			t.Fatalf("%d-column layout = %+v, want too small until %d", width, layout, required)
		}
	}
	app.width = required
	if layout := app.geometry(); layout.tooSmall {
		t.Fatalf("%d-column layout is too small: %+v", required, layout)
	}
}

func TestApp_GeometryUsesStableSplit(t *testing.T) {
	t.Parallel()
	app := &App{width: 80, height: 24, cfg: &config.Config{}, issuesList: views.NewIssuesList()}
	layout := app.geometry()
	if layout.tooSmall {
		t.Fatalf("80×24 marked too small, requires %d×%d", layout.requiredWidth, layout.requiredHeight)
	}
	if layout.status != (rect{0, 0, 80, 3}) || layout.tabs != (rect{0, 3, 22, 5}) || layout.issues != (rect{22, 3, 58, 5}) || layout.detail != (rect{22, 8, 58, 10}) || layout.log != (rect{0, 18, 80, 5}) || layout.help != (rect{0, 23, 80, 1}) {
		t.Fatalf("unexpected 80×24 layout: %+v", layout)
	}

	app.side = sideRight
	app.leftFocus = focusInfo
	if got := app.geometry(); got != layout {
		t.Errorf("focus changed layout: got %+v, want %+v", got, layout)
	}
}

func TestApp_GeometryMaximizeAndTooSmallResize(t *testing.T) {
	t.Parallel()
	app := &App{width: 79, height: 24, cfg: &config.Config{}, issuesList: views.NewIssuesList()}
	layout := app.geometry()
	if !layout.tooSmall || layout.requiredWidth != minWidth {
		t.Fatalf("small geometry = %+v, want too small requiring %d columns", layout, minWidth)
	}

	app.width = 100
	app.maximized = true
	app.maximizedPane = focusIssues
	layout = app.geometry()
	if layout.issues != (rect{0, 0, 100, 23}) || layout.detail.width != 0 || layout.help != (rect{0, 23, 100, 1}) {
		t.Fatalf("maximized geometry = %+v", layout)
	}
}
