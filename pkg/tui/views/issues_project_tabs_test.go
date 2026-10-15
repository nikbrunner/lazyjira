package views

import (
	"testing"

	"github.com/nikbrunner/lazyjira/pkg/config"
	"github.com/nikbrunner/lazyjira/pkg/internal/testkit"
	"github.com/nikbrunner/lazyjira/pkg/jira"
)

const (
	scopedProject = "WEBSDK"
	otherProject  = "PLAT"
)

func projectTabsList(scopedFirst bool) *IssuesList {
	scoped := config.IssueTabConfig{Name: "Web", JQL: "project = WEBSDK", Projects: []string{"websdk"}}
	tabs := []config.IssueTabConfig{
		{Name: "All", JQL: "project = {{.ProjectKey}}"},
		scoped,
		{Name: "Mine", JQL: "assignee = currentUser()"},
	}
	if scopedFirst {
		tabs = []config.IssueTabConfig{scoped, tabs[0], tabs[2]}
	}
	m := NewIssuesList()
	m.SetTabs(tabs)
	return m
}

func tabNames(tabs []config.IssueTabConfig) []string {
	names := make([]string, len(tabs))
	for i, tab := range tabs {
		names[i] = tab.Name
	}
	return names
}

func TestProjectScopedTabsFollowActiveProject(t *testing.T) {
	t.Parallel()
	m := projectTabsList(false)

	m.SetProject(scopedProject)
	testkit.AssertSliceEqual(t, "tabs in scoped project", tabNames(m.Tabs()), []string{"All", "Web", "Mine"})

	m.SetProject(otherProject)
	testkit.AssertSliceEqual(t, "tabs in other project", tabNames(m.Tabs()), []string{"All", "Mine"})

	m.SetProject("")
	testkit.AssertSliceEqual(t, "tabs without project", tabNames(m.Tabs()), []string{"All", "Mine"})
}

func TestHidingActiveTabSelectsFirstVisibleTab(t *testing.T) {
	t.Parallel()
	m := projectTabsList(true)
	m.SetProject(scopedProject)
	m.SetTabIndex(0)

	m.SetProject(otherProject)

	testkit.AssertEqual(t, "active tab", m.ActiveTab().Name, "All")
	testkit.AssertEqual(t, "tab position", m.TabPosition(), 0)
}

func TestHidingEarlierTabKeepsActiveTab(t *testing.T) {
	t.Parallel()
	m := projectTabsList(true)
	m.SetProject(scopedProject)
	m.SetTabIndex(2)

	m.SetProject(otherProject)

	testkit.AssertEqual(t, "active tab", m.ActiveTab().Name, "Mine")
	testkit.AssertEqual(t, "tab position", m.TabPosition(), 1)
}

func TestTabStateStaysWithItsTabAcrossHideAndShow(t *testing.T) {
	t.Parallel()
	m := projectTabsList(false)
	m.SetProject(scopedProject)
	m.SetIssuesForTab(0, []jira.Issue{{Key: "ALL-1"}})
	m.SetIssuesForTab(1, []jira.Issue{{Key: "WEB-1"}})
	m.SetIssuesForTab(2, []jira.Issue{{Key: "MINE-1"}})
	m.SetTabIndex(1)
	m.SetPickerFilter(map[string]bool{"status:Open": true})

	m.SetProject(otherProject)
	m.SetTabPosition(1)
	testkit.AssertEqual(t, "second visible tab", m.ActiveTab().Name, "Mine")
	testkit.AssertEqual(t, "second visible tab issue", m.CurrentIssues()[0].Key, "MINE-1")
	testkit.AssertEqual(t, "second visible tab filter", len(m.PickerFilter()), 0)

	m.SetProject(scopedProject)
	m.SetTabPosition(1)
	testkit.AssertEqual(t, "shown tab", m.ActiveTab().Name, "Web")
	testkit.AssertEqual(t, "shown tab issue", m.CurrentIssues()[0].Key, "WEB-1")
	testkit.AssertEqual(t, "shown tab filter", m.PickerFilter()["status:Open"], true)
}

func TestTabCyclingSkipsHiddenTabs(t *testing.T) {
	t.Parallel()
	m := projectTabsList(false)
	m.SetProject(otherProject)

	m.NextTab()
	testkit.AssertEqual(t, "next", m.ActiveTab().Name, "Mine")
	m.NextTab()
	testkit.AssertEqual(t, "next wraps", m.ActiveTab().Name, "All")
	m.PrevTab()
	testkit.AssertEqual(t, "prev wraps", m.ActiveTab().Name, "Mine")
}

func TestHiddenTabsAreNotActivatedOrSearched(t *testing.T) {
	t.Parallel()
	m := projectTabsList(false)
	m.SetProject(otherProject)
	m.SetIssuesForTab(1, []jira.Issue{{Key: "WEB-1"}})

	m.SetTabIndex(1)
	testkit.AssertEqual(t, "active tab", m.ActiveTab().Name, "All")
	if tab, found := m.FindInAnyTab("WEB-1"); found {
		t.Fatalf("FindInAnyTab found hidden tab %d", tab)
	}
}

func TestTransientTabsShowInEveryProject(t *testing.T) {
	t.Parallel()
	m := projectTabsList(false)
	m.SetProject(otherProject)

	m.AddJQLTab("project = PLAT")

	testkit.AssertSliceEqual(t, "tabs", tabNames(m.Tabs()), []string{"All", "Mine", "JQL"})
	testkit.AssertEqual(t, "tab position", m.TabPosition(), 2)

	m.RemoveJQLTab()
	testkit.AssertEqual(t, "after closing JQL tab", m.ActiveTab().Name, "All")
}

func TestNoVisibleTabsLeavesNoActiveTab(t *testing.T) {
	t.Parallel()
	m := NewIssuesList()
	m.SetTabs([]config.IssueTabConfig{{Name: "Web", JQL: "project = WEBSDK", Projects: []string{scopedProject}}})

	m.SetProject(otherProject)

	testkit.AssertEqual(t, "tabs", len(m.Tabs()), 0)
	testkit.AssertEqual(t, "active tab JQL", m.ActiveTab().JQL, "")
}
