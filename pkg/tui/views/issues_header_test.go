package views

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/tui/components"
)

func TestIssuesList_HeaderFollowsFields(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		fields []string
		want   string
	}{
		{"defaults", nil, "Key Status Summary"},
		{"configured", []string{"key", "status", "type", "priority", "summary", "updated"}, "Key Status Type Priority Summary Updated"},
		{"reordered", []string{"updated", "summary", "assignee", "key"}, "Updated Summary Assignee Key"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			list := NewIssuesList()
			list.SetFields(tt.fields)
			list.SetSize(120, 8)
			lines := strings.Split(ansi.Strip(list.View()), "\n")
			got := strings.Join(strings.Fields(strings.Trim(lines[1], "│")), " ")
			if got != tt.want {
				t.Fatalf("header = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIssuesList_HeaderAlignsWithMixedIconsAndFallbacks(t *testing.T) {
	t.Parallel()
	list := NewIssuesList()
	list.SetFields([]string{"key", "status", "type", "priority", "summary", "assignee", "updated"})
	list.SetTypeIcons(map[string]string{"Story": "文"})
	list.SetPriorityIcons(map[string]string{"High": "!"})
	list.SetStatusIcons(map[string]string{"Open": "→"})
	list.SetIssues([]jira.Issue{
		{Key: "A-1", Summary: "First", Status: &jira.Status{Name: "Open"}, IssueType: &jira.IssueType{Name: "Story"}, Priority: &jira.Priority{Name: "High"}, Assignee: &jira.User{DisplayName: "Alice"}, Updated: time.Now().Add(-2 * time.Hour)},
		{Key: "LONG-22", Summary: "Second", Status: &jira.Status{CategoryKey: "done"}, IssueType: &jira.IssueType{Name: "Task"}, Priority: &jira.Priority{Name: "Low"}, Assignee: &jira.User{DisplayName: "Bob"}, Updated: time.Now().Add(-3 * time.Hour)},
	})
	list.SetSize(110, 8)
	lines := strings.Split(ansi.Strip(list.View()), "\n")
	for _, tt := range []struct {
		header string
		first  string
		second string
	}{
		{"Key", "A-1", "LONG-22"},
		{"Status", "→", "✓"},
		{"Type", "文", "Task"},
		{"Priority", "!", "Low"},
		{"Summary", "First", "Second"},
		{"Assignee", "Alice", "Bob"},
		{"Updated", "2h", "3h"},
	} {
		want := displayColumn(t, lines[1], tt.header)
		for i, value := range []string{tt.first, tt.second} {
			if got := displayColumn(t, lines[i+2], value); got != want {
				t.Errorf("%s starts at %d, header %s at %d", value, got, tt.header, want)
			}
		}
	}
}

func TestIssuesList_HeaderReservesScrollingAndClickSpace(t *testing.T) {
	t.Parallel()
	list := NewIssuesList()
	issues := make([]jira.Issue, 0, 10)
	for i := range 10 {
		issues = append(issues, jira.Issue{Key: fmt.Sprintf("ROW-%d", i)})
	}
	list.SetIssues(issues)
	list.SetSize(60, 6)
	list.ScrollBy(9)
	if list.VisibleRows() != 3 || list.Offset != 7 {
		t.Fatalf("visible/offset = %d/%d, want 3/7", list.VisibleRows(), list.Offset)
	}
	lines := strings.Split(ansi.Strip(list.View()), "\n")
	if len(lines) != 6 || !strings.Contains(lines[1], "Key") || !strings.Contains(lines[4], "ROW-9") {
		t.Fatalf("header or last visible issue missing:\n%s", list.View())
	}
	if list.ClickAt(1) || list.Cursor != 9 {
		t.Fatal("header click selected an issue")
	}
	if list.ClickAt(2) || list.Cursor != 7 {
		t.Fatalf("first issue click: cursor = %d, want 7", list.Cursor)
	}
	if !list.ClickAt(2) {
		t.Fatal("double-click on first visible issue was not detected")
	}
	if list.ClickAt(5) || list.Cursor != 7 {
		t.Fatal("bottom border click selected an issue")
	}
	if got := list.ContentHeight(); got != 13 {
		t.Errorf("natural height = %d, want 13", got)
	}
}

func TestIssuesList_HeaderReservesKeyboardPageAndResizeSpace(t *testing.T) {
	t.Parallel()
	list := NewIssuesList()
	issues := make([]jira.Issue, 20)
	list.SetIssues(issues)
	list.SetSize(60, 10)
	list.ResolveNav = func(string) components.NavAction { return components.NavHalfDown }
	list.KeyNav("d")
	if list.Cursor != 3 {
		t.Errorf("half-page cursor = %d, want 3", list.Cursor)
	}
	list.SetSize(60, 4)
	if list.Offset != list.Cursor {
		t.Errorf("resize offset = %d, want selected row %d", list.Offset, list.Cursor)
	}
}

func TestIssuesList_HeaderFitsSmallPanels(t *testing.T) {
	t.Parallel()
	for _, height := range []int{1, 3, 4, 6} {
		t.Run(strconv.Itoa(height), func(t *testing.T) {
			t.Parallel()
			list := NewIssuesList()
			list.SetFields([]string{"key", "status", "type", "priority", "summary", "updated"})
			list.SetIssues([]jira.Issue{{Key: "A-1", Summary: "A long summary"}})
			list.SetSize(24, height)
			lines := strings.Split(ansi.Strip(list.View()), "\n")
			if len(lines) != height {
				t.Fatalf("height = %d, want %d", len(lines), height)
			}
			for _, line := range lines {
				if got := ansi.StringWidth(line); got != 24 {
					t.Errorf("line width = %d, want 24: %q", got, line)
				}
			}
			if height >= 4 && !strings.Contains(lines[1], "Key") {
				t.Fatal("header missing")
			}
			if height == 3 && !strings.Contains(lines[1], "A-1") {
				t.Fatal("tiny panel should retain its issue row")
			}
		})
	}
}

//nolint:paralleltest // The terminal color profile is process-wide.
func TestIssuesList_ColumnColorsUseANSIForHeadersAndValues(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })
	list := NewIssuesList()
	list.SetFields([]string{"key", "status", "type", "priority", "summary", "assignee", "updated"})
	list.SetIssues([]jira.Issue{{Key: "A-1", Summary: "Example", Status: &jira.Status{CategoryKey: "indeterminate"}, IssueType: &jira.IssueType{Name: "Task"}, Priority: &jira.Priority{Name: "High"}, Assignee: &jira.User{DisplayName: "Alice"}, Updated: time.Now().Add(-2 * time.Hour)}})
	list.SetSize(120, 6)
	for _, focused := range []bool{false, true} {
		list.SetFocused(focused)
		lines := strings.Split(list.View(), "\n")
		for _, tt := range []struct {
			header string
			value  string
			color  int
		}{
			{"Key", "A-1", 36},
			{"Status", "→", 32},
			{"Type", "Task", 35},
			{"Priority", "High", 33},
			{"Summary", "Example", 37},
			{"Assignee", "Alice", 96},
			{"Updated", "2h", 94},
		} {
			for i, text := range []string{tt.header, tt.value} {
				if got := foregroundAt(t, lines[i+1], text); got != tt.color {
					t.Errorf("focused=%v %s foreground = %d, want ANSI %d", focused, text, got, tt.color)
				}
			}
		}
		wantBackground := ""
		if focused {
			wantBackground = "44"
		}
		assertRowBackground(t, lines[2], 120, wantBackground)
	}
}

func displayColumn(t *testing.T, line, text string) int {
	t.Helper()
	idx := strings.Index(line, text)
	if idx < 0 {
		t.Fatalf("%q missing from %q", text, line)
	}
	return ansi.StringWidth(line[:idx])
}

func foregroundAt(t *testing.T, line, text string) int {
	t.Helper()
	idx := strings.Index(line, text)
	if idx < 0 {
		t.Fatalf("%q missing from %q", text, line)
	}
	color := 0
	for _, seq := range regexp.MustCompile(`\x1b\[([0-9;]*)m`).FindAllStringSubmatch(line[:idx], -1) {
		for _, value := range strings.Split(seq[1], ";") {
			code, _ := strconv.Atoi(value)
			if code == 0 || code == 39 {
				color = 0
			} else if code >= 30 && code <= 37 || code >= 90 && code <= 97 {
				color = code
			}
		}
	}
	return color
}
