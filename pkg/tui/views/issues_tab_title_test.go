package views

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/textfuel/lazyjira/v2/pkg/config"
)

func topBorderLine(m *IssuesList) string {
	v := m.View()
	if v == "" {
		return ""
	}
	return strings.SplitN(v, "\n", 2)[0]
}

func makeIssuesListWithTabs(width, height int, tabNames ...string) *IssuesList {
	m := NewIssuesList()
	tabs := make([]config.IssueTabConfig, len(tabNames))
	for i, name := range tabNames {
		tabs[i] = config.IssueTabConfig{Name: name}
	}
	m.SetTabs(tabs)
	m.SetFocusHint("1")
	m.SetSize(width, height)
	return m
}

func TestIssuesList_TitleShowsOnlyActiveCollection(t *testing.T) {
	t.Parallel()
	m := makeIssuesListWithTabs(50, 8, "All", "Mine")
	m.SetTabIndex(1)

	plain := stripANSI(topBorderLine(m))
	if !strings.Contains(plain, "[1] [Mine]") || strings.Contains(plain, "All") {
		t.Errorf("title = %q, want only the active collection", plain)
	}
}

func TestIssuesList_TitleFitsPanelWidth(t *testing.T) {
	t.Parallel()
	for _, width := range []int{24, 50, 80} {
		m := makeIssuesListWithTabs(width, 8, "A Very Long 界 Collection", "Done")
		line := topBorderLine(m)
		if got := lipgloss.Width(line); got != width {
			t.Errorf("width %d title line occupies %d cells", width, got)
		}
	}
}

func TestIssuesList_MaximizeTitleHitTarget(t *testing.T) {
	t.Parallel()
	m := makeIssuesListWithTabs(50, 8, "All")
	line := stripANSI(topBorderLine(m))
	buttonPrefix, _, hasButton := strings.Cut(line, "[+]")
	buttonX := -1
	if hasButton {
		buttonX = lipgloss.Width(buttonPrefix)
	}
	if buttonX < 0 || !m.ClickMaximizeAt(buttonX) {
		t.Fatalf("maximize button in %q was not clickable", line)
	}
	m.SetMaximized(true)
	if !strings.Contains(stripANSI(topBorderLine(m)), "[−]") {
		t.Fatal("maximized title does not show restore control")
	}
}
