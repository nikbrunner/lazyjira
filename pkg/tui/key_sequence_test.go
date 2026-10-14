package tui

import (
	"testing"

	"github.com/nikbrunner/lazyjira/pkg/config"
	"github.com/nikbrunner/lazyjira/pkg/jira"
	"github.com/nikbrunner/lazyjira/pkg/jira/jiratest"
	"github.com/nikbrunner/lazyjira/pkg/tui/components"
)

func sequenceApp(t *testing.T) *App {
	t.Helper()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.issuesList.ResolveNav = func(key string) components.NavAction { return app.keymap.MatchNav(key) }
	app.issuesList.SetFocused(true)
	app.issuesList.SetIssues([]jira.Issue{{Key: "PLAT-1"}, {Key: "PLAT-2"}, {Key: "PLAT-3"}, {Key: "PLAT-4"}})
	app.side = sideLeft
	app.leftFocus = focusIssues
	app.issuesList.Cursor = 2
	return app
}

func TestKeySequence_GGJumpsToTop(t *testing.T) {
	t.Parallel()
	app := sequenceApp(t)

	app.Update(runeKey('g'))
	if app.issuesList.Cursor != 2 {
		t.Fatalf("cursor after g = %d, want it to wait for the second key", app.issuesList.Cursor)
	}
	app.Update(runeKey('g'))
	if app.issuesList.Cursor != 0 {
		t.Errorf("cursor after gg = %d, want 0", app.issuesList.Cursor)
	}
}

func TestKeySequence_OtherKeyDropsPendingPrefix(t *testing.T) {
	t.Parallel()
	app := sequenceApp(t)

	app.Update(runeKey('g'))
	app.Update(runeKey('j'))
	if app.issuesList.Cursor != 3 {
		t.Errorf("cursor after g j = %d, want j handled on its own", app.issuesList.Cursor)
	}
	app.Update(runeKey('g'))
	if app.issuesList.Cursor != 3 {
		t.Errorf("cursor after g j g = %d, want the earlier g dropped", app.issuesList.Cursor)
	}
}

func TestKeySequence_ConfiguredSingleKeyActsImmediately(t *testing.T) {
	t.Parallel()
	app := sequenceApp(t)
	app.keymap = KeymapFromConfig(config.KeybindingConfig{Navigation: config.NavigationKeys{Top: "g"}})

	app.Update(runeKey('g'))
	if app.issuesList.Cursor != 0 {
		t.Errorf("cursor after configured g = %d, want 0", app.issuesList.Cursor)
	}
}

func TestKeySequence_CustomCommandKeyIsNotHeld(t *testing.T) {
	t.Parallel()
	app := sequenceApp(t)
	app.customCmds = []config.ResolvedCustomCommand{resolvedCommand(t, "g", "go", "true", config.CtxIssues)}

	if _, buffered := app.resolveKeySequence(runeKey('g')); buffered {
		t.Error("g bound to a custom command should not wait for a sequence")
	}
}
