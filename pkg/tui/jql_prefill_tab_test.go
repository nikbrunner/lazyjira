package tui

import (
	"testing"

	"github.com/nikbrunner/lazyjira/pkg/config"
	"github.com/nikbrunner/lazyjira/pkg/jira/jiratest"
)

func TestJQLSearch_PrefillsFromFocusedIssueTab(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		focus focusPanel
		setup func(a *App)
		want  string
	}{
		{"issue tabs pane uses tab query", focusIssueTabs, nil, `project = "WEB" AND status = Open`},
		{"issues panel keeps project starter", focusIssues, nil, "project = WEB AND "},
		{"JQL tab uses its query", focusIssueTabs, func(a *App) { a.issuesList.AddJQLTab("type = Bug") }, "type = Bug"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := newAppWithFake(t, &jiratest.FakeClient{T: t})
			a.keymap = DefaultKeymap()
			a.projectKey = "WEB"
			a.jqlFields = nil
			a.issuesList.SetTabs([]config.IssueTabConfig{{Name: "Open", JQL: "project = {{.ProjectKey}} AND status = Open"}})
			if tc.setup != nil {
				tc.setup(a)
			}
			a.side = sideLeft
			a.leftFocus = tc.focus

			pressKey(a, "s")

			if !a.jqlModal.IsVisible() {
				t.Fatal("s should open the JQL search")
			}
			if got := a.jqlModal.InputValue(); got != tc.want {
				t.Errorf("prefill = %q, want %q", got, tc.want)
			}
		})
	}
}
