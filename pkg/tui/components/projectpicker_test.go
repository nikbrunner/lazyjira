package components

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/nikbrunner/lazyjira/pkg/tui/keyname/keytest"
)

func TestProjectPickerFiltersKeyAndNameImmediately(t *testing.T) {
	t.Parallel()
	picker := NewProjectPicker()
	picker.Show([]ProjectChoice{
		{Key: "OPS", Name: "Operations"},
		{Key: "WEB", Name: "Web portal"},
	})

	picker = typeInto(picker, "op")
	if got := len(picker.filtered); got != 1 || picker.filtered[0].Key != "OPS" {
		t.Fatalf("filter by key got %#v", picker.filtered)
	}
	picker.Show([]ProjectChoice{
		{Key: "OPS", Name: "Operations"},
		{Key: "WEB", Name: "Web portal"},
	})
	picker = typeInto(picker, "web")
	if got := len(picker.filtered); got != 1 || picker.filtered[0].Key != "WEB" {
		t.Fatalf("filter by name got %#v", picker.filtered)
	}
}

func TestProjectPickerNavigateAndConfirm(t *testing.T) {
	t.Parallel()
	picker := NewProjectPicker()
	projects := []ProjectChoice{{Key: "OPS", Name: "Operations"}, {Key: "WEB", Name: "Web"}}
	picker.Show(projects)

	picker, _ = picker.Update(keytest.Press("down"))
	picker, _ = picker.Update(keytest.Press("o"))
	if picker.cursor != 0 {
		t.Fatalf("filter should reset cursor, got %d", picker.cursor)
	}
	// Arrow navigation remains available and Enter confirms the selected item.
	picker.query = ""
	picker.applyFilter()
	picker, _ = picker.Update(keytest.Press("down"))
	_, cmd := picker.Update(keytest.Press("enter"))
	if picker.IsVisible() || cmd == nil {
		t.Fatal("confirmation should close picker and emit a selection")
	}
	msg, ok := cmd().(ProjectPickerSelectedMsg)
	if !ok || msg.Project.Key != "WEB" {
		t.Fatalf("confirmation = %#v, want WEB selection", msg)
	}
}

func TestProjectPickerFilterSupportsKeySpace(t *testing.T) {
	t.Parallel()
	picker := NewProjectPicker()
	picker.Show([]ProjectChoice{{Key: "ABC", Name: "Alpha Beta"}, {Key: "XYZ", Name: "AlphaGamma"}})
	picker = typeInto(picker, "Alpha")
	picker, _ = picker.Update(keytest.Press(" "))
	picker = typeInto(picker, "Beta")
	if got := len(picker.filtered); got != 1 || picker.filtered[0].Key != "ABC" {
		t.Fatalf("space-containing filter results = %#v, want ABC", picker.filtered)
	}
}

func TestProjectPickerFilterLineFitsNarrowDisplayWidth(t *testing.T) {
	t.Parallel()
	picker := NewProjectPicker()
	picker.SetSize(40, 20)
	picker.Show([]ProjectChoice{{Key: "ABC", Name: "Alpha"}})
	picker = typeInto(picker, strings.Repeat("界", 40))
	lines := strings.Split(ansi.Strip(picker.View()), "\n")
	for i, line := range lines {
		if width := lipgloss.Width(line); width != 30 {
			t.Errorf("line %d width=%d, want 30-cell picker width: %q", i, width, line)
		}
	}
	if !strings.Contains(lines[1], "…") {
		t.Fatalf("long query not truncated with ellipsis: %q", lines[1])
	}
}

func TestProjectPickerFilterCanIncludeNavigationLetters(t *testing.T) {
	t.Parallel()
	picker := NewProjectPicker()
	picker.Show([]ProjectChoice{{Key: "JIRA", Name: "Jira"}, {Key: "OPS", Name: "Operations"}})
	for _, r := range []rune{'j', 'k'} {
		picker, _ = picker.Update(keytest.Press(string(r)))
	}
	if picker.query != "jk" {
		t.Fatalf("query = %q, want j/k accepted as filter text", picker.query)
	}
}

func TestProjectPickerCancelAndOverlayInterception(t *testing.T) {
	t.Parallel()
	picker := NewProjectPicker()
	key := keytest.Press("x")
	if _, ok := picker.Intercept(key); ok {
		t.Fatal("hidden picker should not intercept")
	}
	picker.Show([]ProjectChoice{{Key: "OPS", Name: "Operations"}})
	cmd, ok := picker.Intercept(key)
	if !ok || cmd != nil || picker.query != "x" {
		t.Fatalf("open picker did not consume/filter printable input: ok=%v query=%q cmd=%v", ok, picker.query, cmd)
	}
	cmd, ok = picker.Intercept(keytest.Press("esc"))
	if !ok || picker.IsVisible() || cmd == nil {
		t.Fatal("escape should be consumed and close picker")
	}
	if _, ok := cmd().(ProjectPickerCancelledMsg); !ok {
		t.Fatalf("cancel command returned %T", cmd())
	}
}

func TestProjectPickerInterceptsMouseWhileVisible(t *testing.T) {
	t.Parallel()
	picker := NewProjectPicker()
	picker.Show([]ProjectChoice{{Key: "A", Name: "Alpha"}, {Key: "B", Name: "Beta"}})
	for _, msg := range []tea.Msg{
		tea.MouseClickMsg{Button: tea.MouseLeft, X: 10, Y: 5},
		tea.MouseWheelMsg{Button: tea.MouseWheelDown, X: 10, Y: 5},
	} {
		cmd, handled := picker.Intercept(msg)
		if !handled || cmd != nil || picker.cursor != 0 || picker.query != "" {
			t.Fatalf("mouse escaped visible picker: handled=%v cmd=%v cursor=%d query=%q", handled, cmd, picker.cursor, picker.query)
		}
	}
}

func TestProjectPickerViewShowsFilterAndEmptyState(t *testing.T) {
	t.Parallel()
	picker := NewProjectPicker()
	picker.SetSize(80, 24)
	picker.Show([]ProjectChoice{{Key: "OPS", Name: "Operations"}})
	picker, _ = picker.Update(keytest.Press("z"))
	view := ansi.Strip(picker.View())
	if !strings.Contains(view, "Filter: z") || !strings.Contains(view, "No projects match") {
		t.Fatalf("picker view missing query/empty state: %q", view)
	}
}

func typeInto(picker ProjectPicker, text string) ProjectPicker {
	for _, key := range keytest.Type(text) {
		picker, _ = picker.Update(key)
	}
	return picker
}
