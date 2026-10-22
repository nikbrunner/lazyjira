package tui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/nikbrunner/lazyjira/pkg/config"
	"github.com/nikbrunner/lazyjira/pkg/internal/testkit"
	"github.com/nikbrunner/lazyjira/pkg/jira"
	"github.com/nikbrunner/lazyjira/pkg/jira/jiratest"
	"github.com/nikbrunner/lazyjira/pkg/tui/components"
)

const scopedTabProject = "WEBSDK"

func projectTabsApp(t *testing.T, fake *jiratest.FakeClient) *App {
	t.Helper()
	app := appWithPanelDims(t, 120)
	app.keymap = DefaultKeymap()
	app.client = fake
	app.demoMode = true
	app.issuesList.SetTabs([]config.IssueTabConfig{
		{Name: "Web", JQL: "project = WEBSDK", Projects: []string{"websdk"}},
		{Name: "All", JQL: "project = {{.ProjectKey}}"},
		{Name: "Mine", JQL: "assignee = currentUser()"},
	})
	app.projectList.SetProjects([]jira.Project{{Key: scopedTabProject, ID: "1"}, {Key: testProject, ID: "2"}})
	app.usersCache.set(scopedTabProject, []jira.User{})
	app.usersCache.set(testProject, []jira.User{})
	app.selectProject(&jira.Project{Key: scopedTabProject, ID: "1"})
	app.issuesList.SetTabIndex(0)
	return app
}

func pickProject(app *App, key string) {
	_, cmd := app.Update(components.ProjectPickerSelectedMsg{Project: components.ProjectChoice{Key: key}})
	runCmd(app, cmd)
}

func runCmd(app *App, cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			runCmd(app, c)
		}
		return
	}
	_, _ = app.Update(msg)
}

func TestIssueTabsPaneListsTabsForActiveProject(t *testing.T) {
	t.Parallel()
	app := projectTabsApp(t, &jiratest.FakeClient{T: t})

	pane := ansi.Strip(app.renderIssueTabs(30, 8))
	for _, name := range []string{"Web", "All", "Mine"} {
		if !strings.Contains(pane, name) {
			t.Fatalf("pane in %s misses %s:\n%s", scopedTabProject, name, pane)
		}
	}

	app.selectProject(&jira.Project{Key: testProject, ID: "2"})
	pane = ansi.Strip(app.renderIssueTabs(30, 8))
	if strings.Contains(pane, "Web") || !strings.Contains(pane, "› All") || !strings.Contains(pane, "Mine") {
		t.Fatalf("pane in %s:\n%s", testProject, pane)
	}
}

func TestSwitchingProjectLoadsFirstVisibleTabWhenActiveTabHides(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.SearchIssuesFunc = func(_ context.Context, _ string, _, _ int) (*jira.SearchResult, error) {
		return &jira.SearchResult{Issues: []jira.Issue{{Key: "PLAT-1"}}}, nil
	}
	app := projectTabsApp(t, fake)
	testkit.AssertEqual(t, "initial tab", app.issuesList.ActiveTab().Name, "Web")

	pickProject(app, testProject)

	testkit.AssertEqual(t, "active tab", app.issuesList.ActiveTab().Name, "All")
	testkit.AssertEqual(t, "search calls", len(fake.SearchIssuesCalls), 1)
	testkit.AssertEqual(t, "searched JQL", fake.SearchIssuesCalls[0].JQL, `project = "PLAT"`)
	testkit.AssertEqual(t, "loaded issue", app.issuesList.CurrentIssues()[0].Key, "PLAT-1")
}

func TestStaleFetchForHiddenTabStaysOutOfVisibleTabs(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.SearchIssuesFunc = func(_ context.Context, _ string, _, _ int) (*jira.SearchResult, error) {
		return &jira.SearchResult{Issues: []jira.Issue{{Key: "WEBSDK-1"}}}, nil
	}
	app := projectTabsApp(t, fake)
	stale := app.fetchActiveTab()

	app.selectProject(&jira.Project{Key: testProject, ID: "2"})
	_, _ = app.Update(stale())

	testkit.AssertEqual(t, "active tab", app.issuesList.ActiveTab().Name, "All")
	testkit.AssertEqual(t, "active issues", len(app.issuesList.CurrentIssues()), 0)
	app.issuesList.NextTab()
	testkit.AssertEqual(t, "next tab", app.issuesList.ActiveTab().Name, "Mine")
	testkit.AssertEqual(t, "next tab issues", len(app.issuesList.CurrentIssues()), 0)
	if _, found := app.issuesList.FindInAnyTab("WEBSDK-1"); found {
		t.Fatal("stale issue reachable from a visible tab")
	}
}

func TestMouseClickActivatesVisibleTabRow(t *testing.T) {
	t.Parallel()
	app := projectTabsApp(t, &jiratest.FakeClient{T: t})
	app.selectProject(&jira.Project{Key: testProject, ID: "2"})

	_, _ = app.mouseClick(panelTabs, 2, 5)

	testkit.AssertEqual(t, "active tab", app.issuesList.ActiveTab().Name, "Mine")
}

func TestFirstLoadedProjectActivatesItsFirstTab(t *testing.T) {
	t.Parallel()
	app := appWithPanelDims(t, 120)
	app.demoMode = true
	app.issuesList.SetTabs([]config.IssueTabConfig{
		{Name: "Web", Projects: []string{scopedTabProject}},
		{Name: "All"},
	})
	app.issuesList.SetProject("")
	testkit.AssertEqual(t, "tab before projects load", app.issuesList.ActiveTab().Name, "All")

	_, _ = app.Update(projectsLoadedMsg{projects: []jira.Project{{Key: scopedTabProject, ID: "1"}}})

	testkit.AssertEqual(t, "tab after projects load", app.issuesList.ActiveTab().Name, "Web")
}

func TestReachingTheLastIssueLoadsTheNextPage(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.SearchIssuesFunc = func(_ context.Context, _ string, _, maxResults int) (*jira.SearchResult, error) {
		issues := make([]jira.Issue, min(maxResults, 120))
		for i := range issues {
			issues[i] = jira.Issue{Key: fmt.Sprintf("PLAT-%d", i+1)}
		}
		return &jira.SearchResult{Issues: issues, HasMore: len(issues) < 120}, nil
	}
	app := projectTabsApp(t, fake)
	pickProject(app, testProject)
	testkit.AssertEqual(t, "first page", len(app.issuesList.CurrentIssues()), 50)
	staleRefresh := app.fetchActiveTab()

	if app.loadMoreIssues() != nil {
		t.Fatal("load more fired before the cursor reached the last issue")
	}
	app.issuesList.SelectByKey("PLAT-50")
	loadMore := app.loadMoreIssues()
	if loadMore == nil {
		t.Fatal("reaching the last issue should load the next page")
	}
	if app.loadMoreIssues() != nil {
		t.Fatal("a second load started while the first is in flight")
	}
	_, _ = app.Update(loadMore())
	_, _ = app.Update(staleRefresh())

	testkit.AssertEqual(t, "loaded", len(app.issuesList.CurrentIssues()), 100)
	testkit.AssertEqual(t, "selected", app.issuesList.SelectedIssue().Key, "PLAT-50")

	_ = app.fetchActiveTab()()
	testkit.AssertEqual(t, "refresh keeps the limit", fake.SearchIssuesCalls[len(fake.SearchIssuesCalls)-1].MaxResults, 100)
}
