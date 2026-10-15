package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/nikbrunner/lazyjira/pkg/internal/testkit"
)

func TestConfirmModal_Answers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		key  tea.KeyMsg
		want tea.Msg
	}{
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}, ConfirmAcceptedMsg{}},
		{tea.KeyMsg{Type: tea.KeyEnter}, ConfirmAcceptedMsg{}},
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}, ConfirmCancelledMsg{}},
		{tea.KeyMsg{Type: tea.KeyEsc}, ConfirmCancelledMsg{}},
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}, ConfirmCancelledMsg{}},
	}
	for _, tt := range tests {
		t.Run(tt.key.String(), func(t *testing.T) {
			t.Parallel()
			c := NewConfirmModal()
			c.Show("Open 23 browser tabs?")

			cmd, ok := c.Intercept(tt.key)

			testkit.AssertEqual(t, "consumed", ok, true)
			testkit.AssertEqual(t, "visible", c.IsVisible(), false)
			if cmd == nil {
				t.Fatal("no answer message")
			}
			testkit.AssertEqual(t, "answer", cmd(), tt.want)
		})
	}
}

func TestConfirmModal_OtherKeysAreSwallowed(t *testing.T) {
	t.Parallel()
	c := NewConfirmModal()
	c.Show("Open 23 browser tabs?")

	cmd, ok := c.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

	testkit.AssertEqual(t, "consumed", ok, true)
	testkit.AssertEqual(t, "still visible", c.IsVisible(), true)
	if cmd != nil {
		t.Fatal("unexpected command")
	}
}

func TestConfirmModal_HiddenIgnoresKeys(t *testing.T) {
	t.Parallel()
	c := NewConfirmModal()

	_, ok := c.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})

	testkit.AssertEqual(t, "consumed", ok, false)
}

func TestConfirmModal_ViewShowsQuestion(t *testing.T) {
	t.Parallel()
	c := NewConfirmModal()
	c.SetSize(80, 24)
	c.Show("Open 23 browser tabs?")

	view := c.View()

	if !strings.Contains(stripANSI(view), "Open 23 browser tabs?") {
		t.Fatalf("question missing from %q", stripANSI(view))
	}
	lines := strings.Split(view, "\n")
	for _, line := range lines {
		testkit.AssertEqual(t, "line width", lipgloss.Width(line), 30)
	}
}
