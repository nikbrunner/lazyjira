package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestLoadingIndicatorAnimatesDotsBeforeLabel(t *testing.T) {
	t.Parallel()
	indicator := NewLoadingIndicator("Loading sprint options")
	if indicator.Start() == nil {
		t.Fatal("starting the indicator should schedule a tick")
	}
	first := indicator.View()
	if !strings.Contains(first, "⠋") {
		t.Fatalf("indicator view should contain the dot glyph: %q", first)
	}
	if strings.Index(first, "⠋") >= strings.Index(first, "Loading sprint options") {
		t.Fatalf("indicator view should place the dots before the label: %q", first)
	}

	tick := LoadingIndicatorTickMsg{indicator: indicator, generation: indicator.generation}
	if indicator.Update(tick) == nil {
		t.Fatal("an active indicator should schedule its next tick")
	}
	second := indicator.View()
	if first == second {
		t.Fatal("indicator frame did not advance")
	}
	if lipgloss.Width(first) != lipgloss.Width(second) {
		t.Fatalf("animation changed the label position: widths %d and %d", lipgloss.Width(first), lipgloss.Width(second))
	}
	if strings.Index(first, "Loading sprint options") != strings.Index(second, "Loading sprint options") {
		t.Fatal("animation moved the label")
	}

	indicator.Stop()
	if indicator.Update(tick) != nil {
		t.Fatal("a stopped indicator scheduled another tick")
	}
	if indicator.Start() == nil {
		t.Fatal("restarting the indicator should schedule a tick")
	}
	restarted := indicator.View()
	if indicator.Update(tick) != nil || indicator.View() != restarted {
		t.Fatal("a previous generation advanced the restarted indicator")
	}

	other := NewLoadingIndicator("Other")
	if other.Start() == nil {
		t.Fatal("starting another indicator should schedule a tick")
	}
	otherTick := LoadingIndicatorTickMsg{indicator: other, generation: other.generation}
	if indicator.Update(otherTick) != nil || indicator.View() != restarted {
		t.Fatal("indicator accepted another instance's tick")
	}
	other.Stop()
}

func TestModalLoadingIndicatorHasPadding(t *testing.T) {
	t.Parallel()
	modal := NewModal()
	modal.SetSize(80, 24)
	modal.ShowLoading("Loading sprints", "Loading sprint options")
	lines := strings.Split(stripANSI(modal.View()), "\n")
	if len(lines) != 5 {
		t.Fatalf("loading modal has %d rows, want borders plus one blank row above and below", len(lines))
	}
	if strings.Trim(lines[1], "│ ") != "" || strings.Trim(lines[3], "│ ") != "" {
		t.Fatal("loading modal is missing vertical padding")
	}
	if !strings.HasPrefix(lines[2], "│  ⠋") || !strings.HasSuffix(lines[2], "  │") {
		t.Fatalf("loading line is missing horizontal padding: %q", lines[2])
	}
}

func TestModalLoadingIndicatorWrapsToNarrowWidth(t *testing.T) {
	t.Parallel()
	modal := NewModal()
	modal.SetSize(24, 12)
	if modal.ShowLoading("Loading sprints", "Fetching sprint options, 読み込み中 🌱") == nil {
		t.Fatal("loading modal should start the indicator")
	}

	for _, line := range strings.Split(modal.View(), "\n") {
		if width := lipgloss.Width(line); width > 24 {
			t.Fatalf("loading modal line width = %d, exceeds terminal width 24: %q", width, line)
		}
	}
}

func TestModalLoadingIndicatorFitsShortTerminalWithLongLabel(t *testing.T) {
	t.Parallel()
	const width, height = 24, 5
	modal := NewModal()
	modal.SetSize(width, height)
	modal.ShowLoading("Loading sprints", "Fetching a very long result from the Jira sprint board cache")

	lines := strings.Split(modal.View(), "\n")
	if len(lines) > height {
		t.Fatalf("loading modal has %d rows, exceeds terminal height %d", len(lines), height)
	}
	for _, line := range lines {
		if lineWidth := lipgloss.Width(line); lineWidth > width {
			t.Fatalf("loading modal line width = %d, exceeds terminal width %d: %q", lineWidth, width, line)
		}
	}
}

func TestModalCancellationStopsLoadingIndicator(t *testing.T) {
	t.Parallel()
	modal := NewModal()
	modal.SetSize(80, 24)
	modal.ShowLoading("Loading sprints", "Fetching sprint options")
	indicator := modal.loading

	updated, cmd := modal.Update(tea.KeyMsg{Type: tea.KeyEsc})
	modal = updated
	if modal.IsVisible() || indicator.active {
		t.Fatal("cancelling the modal did not stop its loading indicator")
	}
	if cmd == nil {
		t.Fatal("cancelling the loading modal should emit a cancellation message")
	}
}

func TestModalReplacementRejectsOldLoadingTick(t *testing.T) {
	t.Parallel()
	modal := NewModal()
	modal.SetSize(80, 24)
	modal.ShowLoading("Loading sprints", "Fetching sprint options")
	oldIndicator := modal.loading
	oldTick := LoadingIndicatorTickMsg{indicator: oldIndicator, generation: oldIndicator.generation}

	modal.Show("Replacement", []ModalItem{{Label: "Keep me"}})
	if oldIndicator.active {
		t.Fatal("replacing a loading modal did not stop its indicator")
	}
	cmd, handled := modal.Intercept(oldTick)
	if handled || cmd != nil {
		t.Fatal("replacement modal consumed a tick from the old indicator")
	}
}
