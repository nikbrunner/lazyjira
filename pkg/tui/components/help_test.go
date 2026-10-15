package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/nikbrunner/lazyjira/pkg/internal/testkit"
)

func TestHelpBar_ViewRendersItems(t *testing.T) {
	t.Parallel()
	bar := NewHelpBar([]HelpItem{
		{Key: "q", Description: "quit"},
		{Key: "?", Description: "help"},
	})
	bar.SetWidth(120)
	out := bar.View()
	plain := stripANSI(out)
	if !strings.Contains(plain, "quit: q") {
		t.Errorf("expected 'quit: q' in output, got %q", plain)
	}
	if !strings.Contains(plain, "help: ?") {
		t.Errorf("expected 'help: ?' in output, got %q", plain)
	}
}

func TestHelpBar_SetItemsReplacesItems(t *testing.T) {
	t.Parallel()
	bar := NewHelpBar([]HelpItem{{Key: "q", Description: "quit"}})
	bar.SetWidth(120)
	bar.SetItems([]HelpItem{{Key: "r", Description: "refresh"}})
	out := stripANSI(bar.View())
	if strings.Contains(out, "quit") {
		t.Error("old item should be replaced")
	}
	if !strings.Contains(out, "refresh: r") {
		t.Errorf("expected new item in output, got %q", out)
	}
}

func TestHelpBar_SetStatusMsgAppearsInView(t *testing.T) {
	t.Parallel()
	bar := NewHelpBar(nil)
	bar.SetWidth(120)
	bar.SetStatusMsg("saved!")
	out := stripANSI(bar.View())
	if !strings.Contains(out, "saved!") {
		t.Errorf("expected status msg in output, got %q", out)
	}
}

func TestHelpBar_SetWidthUpdatesWidth(t *testing.T) {
	t.Parallel()
	bar := NewHelpBar(nil)
	bar.SetWidth(80)
	testkit.AssertEqual(t, "width", bar.width, 80)
}

func TestHelpBar_UpdateHandlesWindowSize(t *testing.T) {
	t.Parallel()
	bar := NewHelpBar(nil)
	updated, _ := bar.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	testkit.AssertEqual(t, "width after window size", updated.width, 100)
}

func TestHelpBar_TruncatesWhenTooNarrow(t *testing.T) {
	t.Parallel()
	items := make([]HelpItem, 10)
	for i := range items {
		items[i] = HelpItem{Key: "x", Description: "longdescription"}
	}
	bar := NewHelpBar(items)
	bar.SetWidth(20)
	out := bar.View()
	if !strings.Contains(stripANSI(out), "...") {
		t.Errorf("expected truncation marker '...' in narrow bar, got %q", stripANSI(out))
	}
}

func TestHelpBar_ViewFitsConfiguredWidth(t *testing.T) {
	t.Parallel()
	items := make([]HelpItem, 10)
	for i := range items {
		items[i] = HelpItem{Key: "ctrl+x", Description: "long action"}
	}
	for _, width := range []int{20, 80, 160} {
		bar := NewHelpBar(items)
		bar.SetWidth(width)
		bar.SetStatusMsg("saved")
		if got := lipgloss.Width(bar.View()); got > width {
			t.Fatalf("help bar width = %d, want at most %d: %q", got, width, stripANSI(bar.View()))
		}
	}
}

func TestHelpBar_InitReturnsNil(t *testing.T) {
	t.Parallel()
	bar := NewHelpBar(nil)
	cmd := bar.Init()
	testkit.AssertEqual(t, "Init returns nil", cmd == nil, true)
}

func TestHelpBar_PinnedItemSurvivesTruncation(t *testing.T) {
	t.Parallel()
	bar := NewHelpBar([]HelpItem{
		{Key: "/", Description: "search"},
		{Key: "s", Description: "JQL search"},
		{Key: "r", Description: "refresh"},
		{Key: "?", Description: "keybindings", Pinned: true},
	})
	bar.SetWidth(40)
	out := stripANSI(bar.View())
	if !strings.HasSuffix(out, "keybindings: ?") {
		t.Errorf("pinned item missing at the end: %q", out)
	}
	if !strings.Contains(out, "search: /") || strings.Contains(out, "refresh") {
		t.Errorf("expected leading hints kept and later ones dropped: %q", out)
	}
	testkit.AssertEqual(t, "fits width", lipgloss.Width(out) <= 40, true)
}

func TestHelpBar_PinnedItemAloneWhenNothingElseFits(t *testing.T) {
	t.Parallel()
	bar := NewHelpBar([]HelpItem{
		{Key: "/", Description: "search"},
		{Key: "?", Description: "keybindings", Pinned: true},
	})
	bar.SetWidth(22)
	out := stripANSI(bar.View())
	testkit.AssertEqual(t, "only pinned", strings.TrimSpace(out), "... | keybindings: ?")
}
