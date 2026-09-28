package views

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"

	"github.com/nikbrunner/lazyjira/v2/pkg/config"
	"github.com/nikbrunner/lazyjira/v2/pkg/jira"
)

func marksFixture() *IssuesList {
	list := NewIssuesList()
	list.SetFields([]string{"key", "summary"})
	list.SetIssues([]jira.Issue{
		{Key: "A-1", Summary: "First"},
		{Key: "A-2", Summary: "Second, longer summary"},
		{Key: "A-3", Summary: "Third"},
		{Key: "A-10", Summary: "Fourth"},
	})
	return list
}

func markedKeys(list *IssuesList) string {
	marked := list.MarkedIssues()
	keys := make([]string, 0, len(marked))
	for _, issue := range marked {
		keys = append(keys, issue.Key)
	}
	return strings.Join(keys, ",")
}

func TestIssuesList_MarksCombineRangeAndSingles(t *testing.T) {
	t.Parallel()
	list := marksFixture()

	list.Cursor = 3
	list.ToggleMark()
	list.Cursor = 0
	list.ToggleVisual()
	list.Cursor = 1
	if got := markedKeys(list); got != "A-1,A-2,A-10" {
		t.Fatalf("marked during range = %q", got)
	}

	list.ToggleVisual()
	list.Cursor = 2
	if got := markedKeys(list); got != "A-1,A-2,A-10" {
		t.Fatalf("marked after ending range = %q", got)
	}

	list.Cursor = 1
	list.ToggleMark()
	if got := markedKeys(list); got != "A-1,A-10" {
		t.Fatalf("marked after unmark = %q", got)
	}

	list.ClearMarks()
	if list.HasMarks() {
		t.Fatal("HasMarks() after ClearMarks = true")
	}
}

func TestIssuesList_RangeFollowsAnchorKeyAcrossRefresh(t *testing.T) {
	t.Parallel()
	list := marksFixture()
	list.Cursor = 2
	list.ToggleVisual()

	list.SetIssues([]jira.Issue{{Key: "A-0"}, {Key: "A-1"}, {Key: "A-2"}, {Key: "A-3"}})
	list.Cursor = 1

	if got := markedKeys(list); got != "A-1,A-2,A-3" {
		t.Fatalf("marked after refresh = %q", got)
	}
}

func TestIssuesList_TabSwitchClearsMarks(t *testing.T) {
	t.Parallel()
	list := marksFixture()
	list.SetTabs([]config.IssueTabConfig{{Name: "One"}, {Name: "Two"}})
	list.ToggleMark()
	list.ToggleVisual()

	list.NextTab()

	if list.HasMarks() {
		t.Fatal("HasMarks() after tab switch = true")
	}
}

func TestIssuesList_MarkedRowsTextKeepsFullSummaries(t *testing.T) {
	t.Parallel()
	list := marksFixture()
	list.SetSize(12, 8)
	list.Cursor = 1
	list.ToggleMark()
	list.Cursor = 3
	list.ToggleMark()

	want := "A-2  Second, longer summary\nA-10 Fourth"
	if got := list.MarkedRowsText(); got != want {
		t.Fatalf("MarkedRowsText() = %q, want %q", got, want)
	}
}

//nolint:paralleltest // The terminal color profile is process-wide.
func TestIssuesList_MarkedRowsShowGlyphAndKeepWidth(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	list := marksFixture()
	th := *list.theme
	th.SelectedItem = th.SelectedItem.Background(lipgloss.Color("#303840"))
	list.theme = &th
	list.SetFocused(true)
	list.SetSize(40, 8)
	list.ToggleMark()
	list.Cursor = 2
	list.ToggleMark()

	lines := strings.Split(list.View(), "\n")
	for _, i := range []int{3, 4, 5} {
		if w := ansi.StringWidth(lines[i]); w != 40 {
			t.Errorf("line %d width = %d, want 40", i, w)
		}
	}
	if got := ansi.Strip(lines[3]); !strings.HasPrefix(got, "│"+markGlyph+"A-1") {
		t.Errorf("marked row = %q, want mark glyph before key", got)
	}
	if got := ansi.Strip(lines[4]); !strings.HasPrefix(got, "│ A-2") {
		t.Errorf("unmarked row = %q, want plain lead", got)
	}
	assertRowBackground(t, lines[5], 40, "48;2;48;56;64")
}
