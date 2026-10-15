package tui

import (
	"strings"
	"testing"

	"github.com/nikbrunner/lazyjira/pkg/tui/components"
	"github.com/nikbrunner/lazyjira/pkg/tui/keyname/keytest"
)

// Terminals with the kitty keyboard protocol or modifyOtherKeys report
// shifted keys in detail. Text inputs treat them like the plain keys a
// legacy terminal sends.
func TestUpdate_EnhancedKeysEditDescriptionLikeLegacyKeys(t *testing.T) {
	t.Parallel()
	app, _ := newEditApp(t, editableIssue(), nil)
	openEdit(t, app)
	app.overlays = components.OverlayStack{&app.createForm}
	for app.createForm.FocusedPanel() != components.CreatePanelDescription {
		_, _ = app.Update(keytest.Press("tab"))
	}

	for _, input := range []string{
		"\x1b[13;2u", "x", // shift+enter, kitty
		"\x1b[27;2;13~", "y", // shift+enter, modifyOtherKeys
		"\x1b[32;2u", "z", // shift+space, kitty
		"\x1b[127;2u", // shift+backspace, kitty
	} {
		_, _ = app.Update(keytest.Decode(input))
	}

	if desc := app.createForm.DescriptionText(); !strings.Contains(desc, "\nx\ny ") || strings.Contains(desc, "z") {
		t.Errorf("description = %q, want the typed lines %q without z", desc, "\nx\ny ")
	}
}

func TestUpdate_EnhancedKeysTypeIntoSearchBar(t *testing.T) {
	t.Parallel()
	app := routingApp(t)
	app.searchBar.Activate()

	for _, input := range []string{
		"\x1b[27;2;32~", // shift+space, modifyOtherKeys
		"\x1b[103;2u",   // shift+g, kitty
		"\x1b[27;2;71~", // shift+g, modifyOtherKeys
	} {
		_, _ = app.Update(keytest.Decode(input))
	}

	if got := app.searchBar.Query(); got != " GG" {
		t.Errorf("search query = %q, want %q", got, " GG")
	}
}
