package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/nikbrunner/lazyjira/pkg/tui/keyname/keytest"
)

func TestUpdate_ActiveSearchBarTypesSpaceAndAt(t *testing.T) {
	t.Parallel()
	app := routingApp(t)
	logHidden := app.logHidden
	app.searchBar.Activate()

	for _, key := range keytest.Type("a b@") {
		_, _ = app.Update(key)
	}

	if got := app.searchBar.Query(); got != "a b@" {
		t.Errorf("search query = %q, want %q", got, "a b@")
	}
	if app.logHidden != logHidden {
		t.Error("@ typed into the search bar must not toggle the Command log")
	}
}

func TestUpdate_ActiveSearchBarTakesPaste(t *testing.T) {
	t.Parallel()
	app := routingApp(t)
	app.searchBar.Activate()

	_, _ = app.Update(tea.PasteMsg{Content: "q @x"})

	if got := app.searchBar.Query(); got != "q @x" {
		t.Errorf("search query = %q, want %q", got, "q @x")
	}
}

func TestUpdate_PasteOutsideTextInputRunsNoBinding(t *testing.T) {
	t.Parallel()
	app := routingApp(t)
	logHidden := app.logHidden

	for _, text := range []string{"q", "@", "?"} {
		if _, cmd := app.Update(tea.PasteMsg{Content: text}); cmd != nil {
			t.Errorf("paste %q returned a command", text)
		}
	}
	if app.showHelp || app.logHidden != logHidden {
		t.Error("pasted text must not run key bindings")
	}
}

func TestUpdate_HelpSearchTakesPaste(t *testing.T) {
	t.Parallel()
	app := routingApp(t)
	app.showHelp = true
	app.helpSearching = true

	_, _ = app.Update(tea.PasteMsg{Content: "quit"})

	if app.helpFilter != "quit" {
		t.Errorf("help filter = %q, want quit", app.helpFilter)
	}
}
