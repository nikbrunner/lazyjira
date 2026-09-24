package tui

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
)

var ansiRe = regexp.MustCompile("\x1b\\[[0-9;]*m")

func plain(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

func appForView(t *testing.T) *App {
	t.Helper()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	app.layoutPanels()
	app.updateFocusState()
	return app
}

func TestView_LoadingBeforeWidthSet(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})

	output := app.View()

	if output != "Loading..." {
		t.Errorf("expected Loading..., got %q", output)
	}
}

func TestView_HorizontalLayout(t *testing.T) {
	t.Parallel()
	app := appForView(t)
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey, Summary: testSummary}})

	output := app.View()

	if output == "" {
		t.Fatal("View() returned empty string")
	}
	stripped := plain(output)
	if stripped == "" {
		t.Error("stripped output should not be empty")
	}
}

func resizeViewApp(t *testing.T, app *App, width, height int) *App {
	t.Helper()
	model, cmd := app.handleResize(tea.WindowSizeMsg{Width: width, Height: height})
	if cmd != nil {
		t.Fatal("resize returned an unexpected command")
	}
	resized, ok := model.(*App)
	if !ok {
		t.Fatalf("resize returned %T, want *App", model)
	}
	return resized
}

func TestView_UsableTerminalRendersEveryCell(t *testing.T) {
	t.Parallel()
	for _, size := range []struct{ width, height int }{{80, 21}, {80, 24}, {160, 50}} {
		t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
			t.Parallel()
			app := newAppWithFake(t, &jiratest.FakeClient{T: t})
			app.keymap = DefaultKeymap()
			app.issuesList.SetIssues([]jira.Issue{{Key: "RENDER-1", Summary: "Layout sample"}})
			app = resizeViewApp(t, app, size.width, size.height)
			output := ansi.Strip(app.View())
			if strings.Contains(output, "terminal too small") {
				t.Fatalf("%dx%d render unexpectedly reports too small", size.width, size.height)
			}
			lines := strings.Split(output, "\n")
			if len(lines) != size.height {
				t.Fatalf("rendered %d lines, want %d", len(lines), size.height)
			}
			for i, line := range lines {
				if width := ansi.StringWidth(line); width != size.width {
					t.Errorf("line %d width = %d, want %d: %q", i, width, size.width, line)
				}
			}
			for _, want := range []string{"Key", "Status", "Summary", "RENDER-1", "Layout sample"} {
				if !strings.Contains(output, want) {
					t.Errorf("%dx%d render missing %q:\n%s", size.width, size.height, want, output)
				}
			}

		})
	}
}

func TestView_IssueColumnsAtMeasuredMinimum(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		fields []string
		want   []string
	}{
		{"default", []string{"key", "status", "summary"}, []string{"Key", "Status", "Summary", "PROJECT-123", "Short summary"}},
		{"selected columns", []string{"key", "summary", "assignee"}, []string{"Key", "Summary", "Assignee", "PROJECT-123", "Ada"}},
		{"all columns", []string{"key", "status", "type", "priority", "summary", "assignee", "updated"}, []string{"Key", "Status", "Type", "Priority", "Summary", "Assignee", "Updated", "PROJECT-123", "Story", "High", "Short", "Ada", "5h"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := newAppWithFake(t, &jiratest.FakeClient{T: t})
			app.keymap = DefaultKeymap()
			app.issuesList.SetFields(tc.fields)
			app.issuesList.SetIssues([]jira.Issue{{
				Key:       "PROJECT-123",
				Summary:   "Short summary",
				Status:    &jira.Status{Name: "In Progress"},
				IssueType: &jira.IssueType{Name: "Story"},
				Priority:  &jira.Priority{Name: "High"},
				Assignee:  &jira.User{DisplayName: "Ada"},
				Updated:   time.Now().Add(-5 * time.Hour),
			}})
			required := app.minimumTerminalWidth()
			app = resizeViewApp(t, app, required-1, 24)
			if output := ansi.Strip(app.View()); !strings.Contains(output, "terminal too small") || !strings.Contains(output, fmt.Sprintf("%d×%d", required, minHeight)) {
				t.Fatalf("render below minimum = %q, want too-small message for %d×%d", output, required, minHeight)
			}
			app = resizeViewApp(t, app, required, 24)
			output := ansi.Strip(app.View())
			if strings.Contains(output, "terminal too small") {
				t.Fatalf("render at minimum still reports too small: %q", output)
			}
			for _, want := range tc.want {
				if !strings.Contains(output, want) {
					t.Errorf("minimum render missing %q:\n%s", want, output)
				}
			}
		})
	}
}

func TestView_TooSmallTerminalAndResizeRecovery(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app = resizeViewApp(t, app, 60, 40)

	if output := plain(app.View()); !strings.Contains(output, "terminal too small") {
		t.Fatalf("narrow terminal view = %q, want size message", output)
	}

	for _, height := range []int{19, 20} {
		app = resizeViewApp(t, app, 80, height)
		if output := plain(app.View()); !strings.Contains(output, "terminal too small") {
			t.Fatalf("%d-row terminal view = %q, want size message", height, output)
		}
	}
	app = resizeViewApp(t, app, 80, 21)
	if output := plain(app.View()); strings.Contains(output, "terminal too small") {
		t.Fatalf("view did not recover at 21 rows: %q", output)
	}
	app = resizeViewApp(t, app, 80, 24)
	if output := app.View(); strings.Contains(plain(output), "terminal too small") {
		t.Fatalf("view did not recover after resize: %q", output)
	}
}

func TestView_ShowHelpOverlay(t *testing.T) {
	t.Parallel()
	app := appForView(t)
	app.showHelp = true

	output := app.View()

	stripped := plain(output)
	if stripped == "" {
		t.Error("help overlay output should not be empty")
	}
}

func TestView_ShowHelpOverlay_WithFilter(t *testing.T) {
	t.Parallel()
	app := appForView(t)
	app.showHelp = true
	app.helpFilter = string(ActQuit)

	output := app.View()

	stripped := plain(output)
	if stripped == "" {
		t.Error("filtered help overlay output should not be empty")
	}
}

func TestView_NoPanic_AllFocusStates(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		setup func(*App)
	}{
		{
			name:  "left issues",
			setup: func(app *App) { app.side = sideLeft; app.leftFocus = focusIssues },
		},
		{
			name:  "left info",
			setup: func(app *App) { app.side = sideLeft; app.leftFocus = focusInfo },
		},
		{
			name:  "left projects",
			setup: func(app *App) { app.side = sideLeft; app.leftFocus = focusProjects },
		},
		{
			name:  "left status",
			setup: func(app *App) { app.side = sideLeft; app.leftFocus = focusProjects },
		},
		{
			name:  "right side",
			setup: func(app *App) { app.side = sideRight },
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			app := appForView(t)
			testCase.setup(app)
			app.updateFocusState()

			output := app.View()
			if output == "" {
				t.Error("View() returned empty string")
			}
		})
	}
}
