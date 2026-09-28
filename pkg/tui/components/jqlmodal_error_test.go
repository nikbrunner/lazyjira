package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestJQLModal_ErrorWrapsIntoPanelAndShrinksHistory(t *testing.T) {
	t.Parallel()
	const width, height = 60, 30
	m := NewJQLModal()
	m.SetSize(width, height)
	m.Show("", []string{"project = A", "project = B"})
	before := strings.Split(m.View(), "\n")
	listBefore := m.listHeight()

	msg := "Error in the JQL Query: Expecting either 'ASC' or 'DESC' but got 'and'. (line 1, character 201)"
	m.SetError(msg)
	lines := strings.Split(m.View(), "\n")

	if len(lines) != len(before) {
		t.Errorf("modal height = %d lines, want %d", len(lines), len(before))
	}
	for i, line := range lines {
		if got := lipgloss.Width(line); got != width {
			t.Errorf("line %d occupies %d cells, want %d: %q", i, got, width, stripANSI(line))
		}
	}
	errLines := m.errorLines()
	if len(errLines) < 2 {
		t.Fatalf("error lines = %q, want the message wrapped over several lines", errLines)
	}
	if got, want := m.listHeight(), listBefore-len(errLines)-2; got != want {
		t.Errorf("history height = %d, want %d", got, want)
	}
	plain := stripANSI(strings.Join(lines, "\n"))
	if !strings.Contains(plain, "Error") || !strings.Contains(plain, "character 201)") {
		t.Errorf("modal does not show the whole message:\n%s", plain)
	}
}

func TestJQLModal_ErrorCappedToKeepHistoryVisible(t *testing.T) {
	t.Parallel()
	const width, height = 40, 20
	m := NewJQLModal()
	m.SetSize(width, height)
	m.Show("", nil)
	before := len(strings.Split(m.View(), "\n"))
	m.SetError(strings.Repeat("very long jira error message ", 40))

	lines := strings.Split(m.View(), "\n")
	if len(lines) != before {
		t.Errorf("modal height = %d lines, want %d", len(lines), before)
	}
	if got := m.listHeight(); got < 3 {
		t.Errorf("history height = %d, want at least 3", got)
	}
	errLines := m.errorLines()
	if last := errLines[len(errLines)-1]; !strings.HasSuffix(last, "…") {
		t.Errorf("last kept error line = %q, want an ellipsis", last)
	}
}

func TestJQLModal_ShowPlacesCursorBeforeOrderBy(t *testing.T) {
	t.Parallel()
	cases := []struct {
		prefill string
		want    string
	}{
		{"statusCategory ≠ Done ORDER BY status DESC", "statusCategory ≠ Done AND status = Accepted ORDER BY status DESC"},
		{"type = Bug order  by created", "type = Bug AND status = Accepted order  by created"},
		{"project = WEB AND ", "project = WEB AND  AND status = Accepted"},
	}
	for _, tc := range cases {
		m := NewJQLModal()
		m.SetSize(80, 30)
		m.Show(tc.prefill, nil)
		m.input.InsertAtCursor(" AND status = Accepted")
		if got := m.InputValue(); got != tc.want {
			t.Errorf("Show(%q) then typing = %q, want %q", tc.prefill, got, tc.want)
		}
	}
}
