package tui

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
)

func appWithPanelDims(t *testing.T, width int) *App {
	t.Helper()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.width = width
	app.height = 40
	app.layoutPanels()
	return app
}

func TestHitTest_UsesSharedLayout(t *testing.T) {
	t.Parallel()
	app := appWithPanelDims(t, 120)

	tests := []struct {
		name     string
		x, y     int
		want     panelID
		wantRelY int
	}{
		{"status", 5, 0, panelStatus, 0},
		{"issue tabs", 5, 4, panelTabs, 1},
		{"issues", 30, 3, panelIssues, 0},
		{"info", 5, 14, panelInfo, 1},
		{"projects", 5, 25, panelProjects, 2},
		{"detail", 30, 18, panelDetail, 5},
		{"log", 30, 35, panelLog, 1},
		{"help bar is not a pane", 5, 39, panelNone, 0},
		{"outside terminal", 120, 10, panelNone, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel, relY := app.hitTest(tt.x, tt.y)
			testkit.AssertEqual(t, "panel", panel, tt.want)
			testkit.AssertEqual(t, "relative Y", relY, tt.wantRelY)
		})
	}
}

func TestHitTest_TooSmallTerminalHasNoPaneTargets(t *testing.T) {
	t.Parallel()
	app := appWithPanelDims(t, 60)
	panel, relY := app.hitTest(2, 2)
	testkit.AssertEqual(t, "panel", panel, panelNone)
	testkit.AssertEqual(t, "relative Y", relY, 0)
}

func TestMouseScroll_FocusesPanel(t *testing.T) {
	t.Parallel()

	t.Run("issues panel scroll takes focus", func(t *testing.T) {
		t.Parallel()
		app := appWithPanelDims(t, 120)
		app.keymap = DefaultKeymap()
		app.side = sideRight
		app.issuesList.SetIssues([]jira.Issue{{Key: testKey}, {Key: "PLAT-2"}})

		_, _ = app.mouseScroll(panelIssues, 3)

		if app.side != sideLeft || app.leftFocus != focusIssues {
			t.Errorf("focus = (%v,%v), want left/issues", app.side, app.leftFocus)
		}
	})

	t.Run("detail panel scroll switches to right side", func(t *testing.T) {
		t.Parallel()
		app := appWithPanelDims(t, 120)
		app.keymap = DefaultKeymap()
		app.side = sideLeft

		_, _ = app.mouseScroll(panelDetail, 3)

		if app.side != sideRight {
			t.Errorf("side = %v, want right", app.side)
		}
	})
}
