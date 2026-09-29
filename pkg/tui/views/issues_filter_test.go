package views

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/nikbrunner/lazyjira/pkg/config"
	"github.com/nikbrunner/lazyjira/pkg/jira"
	"github.com/nikbrunner/lazyjira/pkg/tui/components"
)

func filterIssue(key, status, typ, prio string) jira.Issue {
	issue := jira.Issue{Key: key, Summary: key}
	if status != "" {
		issue.Status = &jira.Status{Name: status}
	}
	if typ != "" {
		issue.IssueType = &jira.IssueType{Name: typ}
	}
	if prio != "" {
		issue.Priority = &jira.Priority{Name: prio}
	}
	return issue
}

func visibleKeys(m *IssuesList) string {
	keys := make([]string, len(m.issues))
	for i, issue := range m.issues {
		keys[i] = issue.Key
	}
	return strings.Join(keys, " ")
}

func filterTestIssues() []jira.Issue {
	return []jira.Issue{
		filterIssue("A-1", "Done", "Bug", "High"),
		filterIssue("A-2", "In Progress", "Story", "Low"),
		filterIssue("A-3", "Blocked", "Bug", ""),
		filterIssue("A-4", "to do", "Story", "High"),
		filterIssue("A-5", "In Progress", "Bug", "Low"),
	}
}

func TestIssuesList_StatusOrderSortsStablyWithUnlistedLast(t *testing.T) {
	t.Parallel()
	m := makeIssuesListWithTabs(80, 12, "All")
	m.SetStatusOrder([]string{"To Do", "In Progress", "Done"})
	m.SetIssues(filterTestIssues())

	if got, want := visibleKeys(m), "A-4 A-2 A-5 A-1 A-3"; got != want {
		t.Errorf("order = %q, want %q", got, want)
	}
}

func TestIssuesList_StatusOrderIgnoresDuplicateNames(t *testing.T) {
	t.Parallel()
	m := makeIssuesListWithTabs(80, 12, "All")
	m.SetStatusOrder([]string{"Done", "done", "DONE", "In Progress"})
	m.SetIssues(filterTestIssues())

	if got, want := visibleKeys(m), "A-1 A-2 A-5 A-3 A-4"; got != want {
		t.Errorf("order = %q, want %q", got, want)
	}
}

func TestIssuesList_StatusOrderRespectsTabOptOut(t *testing.T) {
	t.Parallel()
	off := false
	m := NewIssuesList()
	m.SetTabs([]config.IssueTabConfig{{Name: "Recent", SortByStatus: &off}})
	m.SetSize(80, 12)
	m.SetStatusOrder([]string{"To Do", "In Progress", "Done"})
	m.SetIssues(filterTestIssues())

	if got, want := visibleKeys(m), "A-1 A-2 A-3 A-4 A-5"; got != want {
		t.Errorf("order = %q, want JQL order %q", got, want)
	}
}

func TestIssuesList_PickerFilterCombinesGroupsAndTextFilter(t *testing.T) {
	t.Parallel()
	m := makeIssuesListWithTabs(80, 12, "All")
	m.SetIssues(filterTestIssues())

	m.SetPickerFilter(map[string]bool{"status:In Progress": true, "status:Done": true})
	if got, want := visibleKeys(m), "A-1 A-2 A-5"; got != want {
		t.Errorf("status OR = %q, want %q", got, want)
	}

	m.SetPickerFilter(map[string]bool{"status:In Progress": true, "status:Done": true, "type:Bug": true})
	if got, want := visibleKeys(m), "A-1 A-5"; got != want {
		t.Errorf("status AND type = %q, want %q", got, want)
	}

	m.SetFilter("A-5")
	if got, want := visibleKeys(m), "A-5"; got != want {
		t.Errorf("with text filter = %q, want %q", got, want)
	}
}

func TestIssuesList_PickerFilterMatchesMissingPriorityAsNone(t *testing.T) {
	t.Parallel()
	m := makeIssuesListWithTabs(80, 12, "All")
	m.SetIssues(filterTestIssues())

	m.SetPickerFilter(map[string]bool{"priority:None": true})
	if got, want := visibleKeys(m), "A-3"; got != want {
		t.Errorf("priority None = %q, want %q", got, want)
	}
}

func TestIssuesList_PickerFilterIsPerTab(t *testing.T) {
	t.Parallel()
	m := makeIssuesListWithTabs(80, 12, "All", "Mine")
	m.SetIssues(filterTestIssues())
	m.SetPickerFilter(map[string]bool{"type:Story": true})

	m.NextTab()
	m.SetIssues(filterTestIssues())
	if m.IsPickerFiltered() {
		t.Fatal("second tab inherited the picker filter")
	}
	if got := len(m.issues); got != 5 {
		t.Errorf("second tab shows %d issues, want 5", got)
	}

	m.PrevTab()
	if got, want := visibleKeys(m), "A-2 A-4"; got != want {
		t.Errorf("first tab after switching back = %q, want %q", got, want)
	}
}

func TestIssuesList_FilterPickerItemsKeepSelectedValuesWithoutIssues(t *testing.T) {
	t.Parallel()
	m := makeIssuesListWithTabs(80, 12, "All")
	m.SetStatusOrder([]string{"To Do", "In Progress", "Done"})
	m.SetIssues(filterTestIssues())
	m.SetPickerFilter(map[string]bool{"status:QA": true})

	items := m.FilterPickerItems(m.PickerFilter())
	labels := make([]string, 0, len(items))
	for _, item := range items {
		labels = append(labels, item.Label)
	}
	got := strings.Join(labels, "|")
	want := "Status|to do (1)|In Progress (2)|Done (1)|Blocked (1)|QA (0)|" +
		"Type|Bug (0)|Story (0)|" +
		"Priority|High (0)|Low (0)|None (0)"
	if got != want {
		t.Errorf("items = %q\nwant    %q", got, want)
	}
}

func TestIssuesList_TitleShowsPickerFilter(t *testing.T) {
	t.Parallel()
	m := makeIssuesListWithTabs(80, 8, "Sprint")
	m.SetClearFilterKey("esc")
	m.SetIssues(filterTestIssues())
	m.SetFilter("A")
	m.SetPickerFilter(map[string]bool{"status:In Progress": true, "type:Bug": true})

	plain := stripANSI(topBorderLine(m))
	if !strings.Contains(plain, "─ /A status: In Progress type: Bug (esc to clear) ─") {
		t.Errorf("title = %q, want text and picker filters", plain)
	}
}

func TestIssuesList_FooterShowsLoadedHint(t *testing.T) {
	t.Parallel()
	cases := []struct {
		total   int
		hasMore bool
		want    string
	}{
		{312, true, "1 of 5 · 5/312 loaded"},
		{0, true, "1 of 5 · 5+ loaded"},
		{5, false, "1 of 5"},
	}
	for _, tc := range cases {
		m := makeIssuesListWithTabs(80, 12, "All")
		m.SetTabPageInfo(0, tc.total, tc.hasMore)
		m.SetIssues(filterTestIssues())
		if got := m.footer(80); got != tc.want {
			t.Errorf("footer(total=%d, more=%v) = %q, want %q", tc.total, tc.hasMore, got, tc.want)
		}
	}
}

func TestIssuesList_LoadedHintKeepsPanelWidth(t *testing.T) {
	t.Parallel()
	for _, width := range []int{14, 20, 30, 40} {
		m := makeIssuesListWithTabs(width, 8, "All")
		m.SetTabPageInfo(0, 312, true)
		m.SetIssues(filterTestIssues())
		lines := strings.Split(m.View(), "\n")
		for _, line := range lines {
			if got := lipgloss.Width(line); got != width {
				t.Errorf("width %d line %q occupies %d cells", width, stripANSI(line), got)
			}
		}
		bottom := stripANSI(lines[len(lines)-1])
		if fits := width >= 26; strings.Contains(bottom, "loaded") != fits {
			t.Errorf("width %d footer = %q, want hint shown only when it fits", width, bottom)
		}
	}
}

func TestIssuesList_CollapsedBarDropsLoadedHintWhenTight(t *testing.T) {
	t.Parallel()
	m := makeIssuesListWithTabs(40, 1, "All")
	m.SetTabPageInfo(0, 312, true)
	m.SetIssues(filterTestIssues())
	if got := lipgloss.Width(m.View()); got != 40 {
		t.Errorf("collapsed bar occupies %d cells, want 40", got)
	}
}

func TestIssuesList_NewJQLSearchStartsWithoutPickerFilter(t *testing.T) {
	t.Parallel()
	m := makeIssuesListWithTabs(80, 12, "All")
	m.AddJQLTab("type = Bug")
	m.SetIssues(filterTestIssues())
	m.SetPickerFilter(map[string]bool{"status:Done": true})

	m.AddJQLTab("type = Story")
	m.SetIssues(filterTestIssues())
	if m.IsPickerFiltered() {
		t.Error("new JQL search kept the previous picker filter")
	}
}

func pickerLabels(items []components.ModalItem) (labels, disabled string) {
	var all, off []string
	for _, item := range items {
		if item.Separator {
			continue
		}
		all = append(all, item.Label)
		if item.Disabled {
			off = append(off, item.ID)
		}
	}
	return strings.Join(all, "|"), strings.Join(off, "|")
}

func TestIssuesList_FilterPickerCountsFacetAgainstOtherGroups(t *testing.T) {
	t.Parallel()
	m := makeIssuesListWithTabs(80, 12, "All")
	m.SetIssues(filterTestIssues())

	labels, disabled := pickerLabels(m.FilterPickerItems(map[string]bool{"type:Story": true, "status:Done": true}))
	wantLabels := "Blocked (0)|Done (0)|In Progress (1)|to do (1)|" +
		"Bug (1)|Story (0)|" +
		"High (0)|Low (0)|None (0)"
	if labels != wantLabels {
		t.Errorf("labels = %q\nwant     %q", labels, wantLabels)
	}
	wantDisabled := "status:Blocked|priority:High|priority:Low|priority:None"
	if disabled != wantDisabled {
		t.Errorf("disabled = %q, want %q", disabled, wantDisabled)
	}
}

func TestIssuesList_FilterPickerCountsRespectTextFilter(t *testing.T) {
	t.Parallel()
	m := makeIssuesListWithTabs(80, 12, "All")
	m.SetIssues(filterTestIssues())
	m.SetFilter("A-1")

	labels, _ := pickerLabels(m.FilterPickerItems(nil))
	want := "Blocked (0)|Done (1)|In Progress (0)|to do (0)|Bug (1)|Story (0)|High (1)|Low (0)|None (0)"
	if labels != want {
		t.Errorf("labels = %q\nwant     %q", labels, want)
	}
}
