package components

import (
	"testing"

	"charm.land/bubbles/v2/textarea"
	"charm.land/lipgloss/v2"

	"github.com/nikbrunner/lazyjira/pkg/tui/theme"
)

//nolint:paralleltest // Switches the process-wide theme.
func TestRefreshStylesFollowsThemeSwitch(t *testing.T) {
	t.Cleanup(func() { theme.Init(theme.Options{}) })
	theme.Init(theme.Options{Background: lipgloss.Color("#20262c")})
	f := NewCreateForm()
	f.ShowForm(nil, "Task", "PLAT")

	theme.SetBackground(lipgloss.Color("#e0e0e0"))
	f.RefreshStyles()

	want := textarea.DefaultStyles(false).Focused.EndOfBuffer.GetForeground()
	if got := f.desc.Styles().Focused.EndOfBuffer.GetForeground(); got != want {
		t.Errorf("Description end-of-buffer = %v, want the light %v", got, want)
	}
}
