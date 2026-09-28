package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/nikbrunner/lazyjira/v2/pkg/config"
)

func TestJQLSearchResult_ScrollsIssueTabsToJQLTab(t *testing.T) {
	t.Parallel()
	app := appWithPanelDims(t, 120)
	app.keymap = DefaultKeymap()
	tabs := make([]config.IssueTabConfig, 12)
	for i := range tabs {
		tabs[i] = config.IssueTabConfig{Name: "Collection " + string(rune('A'+i))}
	}
	app.issuesList.SetTabs(tabs)

	app.handleJQLSearchResult(jqlSearchResultMsg{jql: "type = Bug"})
	view := ansi.Strip(app.renderIssueTabs(app.geometry().tabs.width, app.geometry().tabs.height))
	if !strings.Contains(view, "› JQL") {
		t.Fatalf("Issue tabs pane does not show the active JQL tab:\n%s", view)
	}

	app.side = sideLeft
	app.leftFocus = focusIssues
	pressKey(app, "x")
	view = ansi.Strip(app.renderIssueTabs(app.geometry().tabs.width, app.geometry().tabs.height))
	if !strings.Contains(view, "› Collection A") {
		t.Fatalf("Issue tabs pane does not show the first tab after closing JQL:\n%s", view)
	}
}
