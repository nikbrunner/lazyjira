package views

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"

	"github.com/nikbrunner/lazyjira/v2/pkg/jira"
)

//nolint:paralleltest // The terminal color profile is process-wide.
func TestIssuesList_SelectedBackgroundCoversRow(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	list := NewIssuesList()
	th := *list.theme
	th.SelectedItem = th.SelectedItem.Background(lipgloss.Color("#303840"))
	list.theme = &th
	list.SetFields([]string{"key", "type", "summary", "updated"})
	list.SetTypeIcons(map[string]string{"Task": "文"})
	list.SetIssues([]jira.Issue{
		{Key: "A-1", IssueType: &jira.IssueType{Name: "Task"}, Summary: "A long summary to truncate"},
		{Key: "A-2", Summary: "Other"},
	})
	list.SetFocused(true)
	for _, width := range []int{24, 80} {
		list.SetSize(width, 6)
		lines := strings.Split(list.View(), "\n")
		assertRowBackground(t, lines[1], width, "")
		assertRowBackground(t, lines[2], width, "")
		assertRowBackground(t, lines[3], width, "48;2;48;56;64")
		assertRowBackground(t, lines[4], width, "")
	}
}

func assertRowBackground(t *testing.T, line string, width int, want string) {
	t.Helper()
	background := ""
	column := 0
	end := 0
	check := func(text string) {
		t.Helper()
		for _, char := range text {
			cells := ansi.StringWidth(string(char))
			if column > 0 && column < width-1 && background != want {
				t.Errorf("column %d (%q): background = %q, want %q", column, char, background, want)
			}
			column += cells
		}
	}
	for _, match := range regexp.MustCompile(`\x1b\[([0-9;]*)m`).FindAllStringSubmatchIndex(line, -1) {
		check(line[end:match[0]])
		codes := strings.Split(line[match[2]:match[3]], ";")
		for i := 0; i < len(codes); i++ {
			code, _ := strconv.Atoi(codes[i])
			switch {
			case code == 0 || code == 49:
				background = ""
			case code >= 40 && code <= 47 || code >= 100 && code <= 107:
				background = codes[i]
			case (code == 38 || code == 48) && i+1 < len(codes):
				count := 2
				if codes[i+1] == "2" {
					count = 4
				}
				if i+count >= len(codes) {
					t.Fatalf("incomplete color sequence in %q", line)
				}
				if code == 48 {
					background = strings.Join(codes[i:i+count+1], ";")
				}
				i += count
			}
		}
		end = match[1]
	}
	check(line[end:])
	if column != width {
		t.Errorf("row width = %d, want %d", column, width)
	}
}
