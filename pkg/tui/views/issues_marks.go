package views

import (
	"math"
	"strings"

	"github.com/nikbrunner/lazyjira/v2/pkg/jira"
	"github.com/nikbrunner/lazyjira/v2/pkg/tui/components"
)

// ToggleMark marks or unmarks the issue under the cursor.
func (m *IssuesList) ToggleMark() {
	sel := m.SelectedIssue()
	if sel == nil {
		return
	}
	if m.marked[sel.Key] {
		delete(m.marked, sel.Key)
		return
	}
	if m.marked == nil {
		m.marked = make(map[string]bool)
	}
	m.marked[sel.Key] = true
}

// ToggleVisual starts a range at the cursor, or ends the active range and
// keeps its issues marked.
func (m *IssuesList) ToggleVisual() {
	if m.visualAnchor == "" {
		if sel := m.SelectedIssue(); sel != nil {
			m.visualAnchor = sel.Key
		}
		return
	}
	for _, issue := range m.MarkedIssues() {
		if m.marked == nil {
			m.marked = make(map[string]bool)
		}
		m.marked[issue.Key] = true
	}
	m.visualAnchor = ""
}

// HasMarks reports whether any marked issue is visible in the issue list.
func (m *IssuesList) HasMarks() bool {
	return len(m.MarkedIssues()) > 0
}

func (m *IssuesList) ClearMarks() {
	m.marked = nil
	m.visualAnchor = ""
}

// visualRange returns the inclusive issue-list range between the visual
// anchor and the cursor, or ok=false when no range is active.
func (m *IssuesList) visualRange() (lo, hi int, ok bool) {
	if m.visualAnchor == "" {
		return 0, 0, false
	}
	for i, issue := range m.issues {
		if issue.Key == m.visualAnchor {
			return min(i, m.Cursor), max(i, m.Cursor), true
		}
	}
	return 0, 0, false
}

func (m *IssuesList) isMarked(i int) bool {
	if lo, hi, ok := m.visualRange(); ok && i >= lo && i <= hi {
		return true
	}
	return m.marked[m.issues[i].Key]
}

// MarkedIssues returns the marked issues and the active range in issue-list
// order. Marked issues hidden by the local filter are excluded.
func (m *IssuesList) MarkedIssues() []jira.Issue {
	var out []jira.Issue
	for i, issue := range m.issues {
		if m.isMarked(i) {
			out = append(out, issue)
		}
	}
	return out
}

// MarkedRowsText renders the marked issues as plain-text issue rows, one per
// line, with the Summary column fitted to the longest summary in the issue list.
func (m *IssuesList) MarkedRowsText() string {
	columns := m.issueColumns(math.MaxInt32)
	marked := m.MarkedIssues()
	lines := make([]string, 0, len(marked))
	for _, issue := range marked {
		cells := make([]string, len(columns))
		for i, column := range columns {
			cells[i] = padRight(components.TruncateEnd(m.issueFieldValue(issue, column.field), column.width), column.width)
		}
		lines = append(lines, strings.TrimRight(strings.Join(cells, " "), " "))
	}
	return strings.Join(lines, "\n")
}
