package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/nikbrunner/lazyjira/v2/pkg/jira"
	"github.com/nikbrunner/lazyjira/v2/pkg/tui/components"
)

const fieldSummary = "summary"

type issueColumn struct {
	field string
	label string
	width int
	color lipgloss.Color
}

func (m *IssuesList) MinimumWidth() int {
	columns := m.issueColumns(0)
	width := len(columns) + 2
	for _, column := range columns {
		width += column.width
	}
	return width
}

func (m *IssuesList) issueColumns(width int) []issueColumn {
	fields := m.fields
	if len(fields) == 0 {
		fields = []string{"key", fieldStatus, fieldSummary}
	}
	columns := make([]issueColumn, 0, len(fields))
	fixedWidth := 0
	for _, field := range fields {
		column := issueColumn{field: field}
		switch field {
		case "key":
			column.label, column.width, column.color = "Key", m.keyColWidth, "6"
		case fieldStatus:
			column.label, column.width, column.color = "Status", m.statusIconCols, "2"
		case "type":
			column.label, column.width, column.color = "Type", m.typeIconCols, "5"
			if len(m.typeIcons) == 0 {
				column.width = 10
			}
			for _, issue := range m.allIssues {
				if typeIcon(m.typeIcons, issue.IssueType) == "" {
					column.width = max(column.width, 10)
				}
			}
		case "priority":
			column.label, column.width, column.color = "Priority", m.priorityIconCols, "3"
			if len(m.priorityIcons) == 0 {
				column.width = 8
			}
			for _, issue := range m.allIssues {
				if priorityIcon(m.priorityIcons, issue.Priority) == "" {
					column.width = max(column.width, 8)
				}
			}
		case fieldSummary:
			column.label, column.color = "Summary", "7"
		case "assignee":
			column.label, column.width, column.color = "Assignee", 12, "14"
		case "updated":
			column.label, column.width, column.color = "Updated", 8, "12"
		default:
			continue
		}
		if field != fieldSummary {
			column.width = max(column.width, len(column.label))
			fixedWidth += column.width
		}
		columns = append(columns, column)
	}
	fixedWidth += len(columns)
	for i := range columns {
		if columns[i].field == fieldSummary {
			summaryWidth := len(columns[i].label)
			for _, issue := range m.issues {
				summaryWidth = max(summaryWidth, ansi.StringWidth(issue.Summary))
			}
			columns[i].width = max(min(summaryWidth, width-fixedWidth), len(columns[i].label))
		}
	}
	return columns
}

func renderIssueHeader(columns []issueColumn, width int) string {
	parts := make([]string, len(columns))
	for i, column := range columns {
		parts[i] = lipgloss.NewStyle().Foreground(column.color).Bold(true).
			Render(padRight(column.label, column.width))
	}
	return ansi.Truncate(" "+strings.Join(parts, " "), width, "")
}

func (m *IssuesList) issueFieldValue(issue jira.Issue, field string) string {
	switch field {
	case "key":
		return issue.Key
	case fieldSummary:
		return issue.Summary
	case fieldStatus:
		if icon := statusIcon(m.statusIcons, issue.Status); icon != "" {
			return icon
		}
		return statusEmojiPlain(issue.Status)
	case "type":
		if icon := typeIcon(m.typeIcons, issue.IssueType); icon != "" {
			return icon
		}
		if issue.IssueType != nil {
			return components.TruncateEnd(issue.IssueType.Name, 10)
		}
	case "priority":
		if icon := priorityIcon(m.priorityIcons, issue.Priority); icon != "" {
			return icon
		}
		if issue.Priority != nil {
			return components.TruncateEnd(issue.Priority.Name, 8)
		}
	case "assignee":
		if issue.Assignee != nil {
			return issue.Assignee.DisplayName
		}
	case "updated":
		return issueTimeAgo(issue.Updated)
	}
	return ""
}
