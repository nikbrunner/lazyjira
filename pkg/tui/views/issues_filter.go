package views

import (
	"fmt"
	"slices"
	"strings"

	"github.com/nikbrunner/lazyjira/pkg/jira"
	"github.com/nikbrunner/lazyjira/pkg/tui/components"
)

// Picker filter groups, in picker order.
const (
	FilterStatus   = "status"
	FilterType     = "type"
	FilterPriority = "priority"
)

var filterGroups = []struct{ key, label string }{
	{FilterStatus, "Status"},
	{FilterType, "Type"},
	{FilterPriority, "Priority"},
}

const filterNone = "None"

type pageInfo struct {
	total   int
	hasMore bool
}

// SetStatusOrder sets the status names issue lists sort by. Matching is
// case-insensitive; statuses missing from the list sort last.
func (m *IssuesList) SetStatusOrder(order []string) {
	m.statusRank = nil
	for _, name := range order {
		if m.statusRank == nil {
			m.statusRank = make(map[string]int, len(order))
		}
		key := strings.ToLower(name)
		if _, ok := m.statusRank[key]; !ok {
			m.statusRank[key] = len(m.statusRank)
		}
	}
	m.applyFilterKeepCursor()
}

func (m *IssuesList) sortedByStatus(issues []jira.Issue) []jira.Issue {
	if m.statusRank == nil || !m.ActiveTab().SortsByStatus() {
		return issues
	}
	rank := func(issue jira.Issue) int {
		if issue.Status != nil {
			if r, ok := m.statusRank[strings.ToLower(issue.Status.Name)]; ok {
				return r
			}
		}
		return len(m.statusRank)
	}
	sorted := slices.Clone(issues)
	slices.SortStableFunc(sorted, func(a, b jira.Issue) int { return rank(a) - rank(b) })
	return sorted
}

func filterKey(group, value string) string { return group + ":" + value }

func issueFilterValue(issue jira.Issue, group string) string {
	var name string
	switch group {
	case FilterStatus:
		if issue.Status != nil {
			name = issue.Status.Name
		}
	case FilterType:
		if issue.IssueType != nil {
			name = issue.IssueType.Name
		}
	case FilterPriority:
		if issue.Priority != nil {
			name = issue.Priority.Name
		}
	}
	if name == "" {
		return filterNone
	}
	return name
}

// PickerFilter returns the active tab's picker filter as a set of
// "group:value" keys.
func (m *IssuesList) PickerFilter() map[string]bool {
	return m.pickerFilters[m.tab]
}

// SetPickerFilter replaces the active tab's picker filter.
func (m *IssuesList) SetPickerFilter(selected map[string]bool) {
	if len(selected) == 0 {
		delete(m.pickerFilters, m.tab)
	} else {
		if m.pickerFilters == nil {
			m.pickerFilters = make(map[int]map[string]bool)
		}
		m.pickerFilters[m.tab] = selected
	}
	m.applyFilterKeepCursor()
}

func (m *IssuesList) IsPickerFiltered() bool { return len(m.pickerFilters[m.tab]) > 0 }

func (m *IssuesList) ClearPickerFilter() { m.SetPickerFilter(nil) }

func (m *IssuesList) matchesPickerFilter(issue jira.Issue) bool {
	return matchesSelection(issue, m.pickerFilters[m.tab], "")
}

// matchesSelection reports whether issue passes every active group of
// selected, ignoring the group named skip.
func matchesSelection(issue jira.Issue, selected map[string]bool, skip string) bool {
	for _, g := range filterGroups {
		if g.key == skip || !groupActive(selected, g.key) {
			continue
		}
		if !selected[filterKey(g.key, issueFilterValue(issue, g.key))] {
			return false
		}
	}
	return true
}

func groupActive(selected map[string]bool, group string) bool {
	prefix := group + ":"
	for key := range selected {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

// FilterPickerItems lists each group's values found in the loaded issues, plus
// selected values no loaded issue carries anymore. A value's count is the
// number of issues that pass the local filter and the other groups' selections
// and carry that value; unselected values without issues are disabled.
func (m *IssuesList) FilterPickerItems(selected map[string]bool) []components.ModalItem {
	q := strings.ToLower(m.filter)
	var textMatches []jira.Issue
	for _, issue := range m.allIssues {
		if matchesText(issue, q) {
			textMatches = append(textMatches, issue)
		}
	}
	var items []components.ModalItem
	for _, g := range filterGroups {
		counts := make(map[string]int)
		seen := make(map[string]bool)
		var values []string
		for _, issue := range m.allIssues {
			if v := issueFilterValue(issue, g.key); !seen[v] {
				seen[v] = true
				values = append(values, v)
			}
		}
		for _, issue := range textMatches {
			if matchesSelection(issue, selected, g.key) {
				counts[issueFilterValue(issue, g.key)]++
			}
		}
		prefix := g.key + ":"
		for key := range selected {
			if v, ok := strings.CutPrefix(key, prefix); ok && !seen[v] {
				values = append(values, v)
			}
		}
		if len(values) == 0 {
			continue
		}
		if g.key == FilterStatus {
			values = m.sortedStatusNames(values)
		} else {
			slices.Sort(values)
		}
		items = append(items, components.ModalItem{Label: g.label, Separator: true})
		for _, v := range values {
			id := filterKey(g.key, v)
			items = append(items, components.ModalItem{
				ID:       id,
				Label:    fmt.Sprintf("%s (%d)", v, counts[v]),
				Disabled: counts[v] == 0 && !selected[id],
			})
		}
	}
	return items
}

// PickerResultCount returns how many loaded issues pass the local filter and
// selected, out of all loaded issues.
func (m *IssuesList) PickerResultCount(selected map[string]bool) (matching, loaded int) {
	q := strings.ToLower(m.filter)
	for _, issue := range m.allIssues {
		if matchesText(issue, q) && matchesSelection(issue, selected, "") {
			matching++
		}
	}
	return matching, len(m.allIssues)
}

func (m *IssuesList) sortedStatusNames(names []string) []string {
	slices.Sort(names)
	if m.statusRank == nil {
		return names
	}
	rank := func(name string) int {
		if r, ok := m.statusRank[strings.ToLower(name)]; ok {
			return r
		}
		return len(m.statusRank)
	}
	slices.SortStableFunc(names, func(a, b string) int { return rank(a) - rank(b) })
	return names
}

// pickerFilterLabel renders the active picker filter as "status: A, B type: Bug".
func (m *IssuesList) pickerFilterLabel() string {
	selected := m.pickerFilters[m.tab]
	var parts []string
	for _, g := range filterGroups {
		var values []string
		prefix := g.key + ":"
		for key := range selected {
			if v, ok := strings.CutPrefix(key, prefix); ok {
				values = append(values, v)
			}
		}
		if len(values) == 0 {
			continue
		}
		if g.key == FilterStatus {
			values = m.sortedStatusNames(values)
		} else {
			slices.Sort(values)
		}
		parts = append(parts, g.key+": "+strings.Join(values, ", "))
	}
	return strings.Join(parts, " ")
}

// SetTabPageInfo records whether Jira holds more results for a tab than it loaded.
func (m *IssuesList) SetTabPageInfo(tab, total int, hasMore bool) {
	if !hasMore {
		delete(m.pageInfo, tab)
		return
	}
	if m.pageInfo == nil {
		m.pageInfo = make(map[int]pageInfo)
	}
	m.pageInfo[tab] = pageInfo{total: total, hasMore: true}
}

// loadedHint renders "50/312 loaded", or "50+ loaded" when Jira reports no
// total, for tabs with results beyond the loaded page.
func (m *IssuesList) loadedHint() string {
	info, ok := m.pageInfo[m.tab]
	if !ok || !info.hasMore {
		return ""
	}
	if info.total > len(m.allIssues) {
		return fmt.Sprintf("%d/%d loaded", len(m.allIssues), info.total)
	}
	return fmt.Sprintf("%d+ loaded", len(m.allIssues))
}

// dropTabState forgets the per-tab filter and page info of a removed tab.
func (m *IssuesList) dropTabState(tab int) {
	delete(m.pickerFilters, tab)
	delete(m.pageInfo, tab)
}
