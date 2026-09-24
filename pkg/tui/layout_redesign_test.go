package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/tui/views"
)

func TestIssueTabsPane_CompactViewportKeepsTabsVisibleSelectable(t *testing.T) {
	t.Parallel()
	app := appWithPanelDims(t, 80)
	app.height = 21
	app.layoutPanels()
	app.keymap = DefaultKeymap()
	tabs := make([]config.IssueTabConfig, 5)
	for i := range tabs {
		tabs[i] = config.IssueTabConfig{Name: fmt.Sprintf("Collection %d", i)}
	}
	app.issuesList.SetTabs(tabs)
	app.side, app.leftFocus = sideLeft, focusIssueTabs

	if got := issueTabVisibleCount(len(tabs), app.geometry().tabs.height-2); got != 2 {
		t.Fatalf("80x21 visible tab count=%d, want both content rows", got)
	}
	app.setIssueTabIndex(4)
	view := ansi.Strip(app.renderIssueTabs(app.geometry().tabs.width, app.geometry().tabs.height))
	if !strings.Contains(view, "› Collection 4") || !strings.Contains(view, "Collection 3") {
		t.Fatalf("compact pane hid active/adjacent tabs:\n%s", view)
	}
	if index, overflow := app.issueTabAtRow(2); index != 4 || overflow != 0 {
		t.Fatalf("compact second row maps to index=%d overflow=%d, want active tab 4", index, overflow)
	}

	app.setIssueTabIndex(0)
	_, _ = app.handleKeyMsg(runeKey('j'))
	if app.issuesList.GetTabIndex() != 1 {
		t.Fatalf("j changed active tab to %d, want 1", app.issuesList.GetTabIndex())
	}
	app.setIssueTabIndex(4)
	_, _ = app.mouseClick(panelTabs, 1, 5)
	if app.issuesList.GetTabIndex() != 3 {
		t.Fatalf("mouse click selected tab %d, want visible tab 3", app.issuesList.GetTabIndex())
	}

	app.setIssueTabIndex(4)
	app = resizeViewApp(t, app, 80, 24)
	view = ansi.Strip(app.renderIssueTabs(app.geometry().tabs.width, app.geometry().tabs.height))
	if !strings.Contains(view, "› Collection 4") {
		t.Fatalf("resize hid active tab:\n%s", view)
	}
}

func TestIssueTabsPane_MinimumViewportZeroOneAndManyTabs(t *testing.T) {
	t.Parallel()
	for _, count := range []int{0, 1, 5} {
		t.Run(fmt.Sprintf("%d tabs", count), func(t *testing.T) {
			t.Parallel()
			app := appWithPanelDims(t, 80)
			app.height = 21
			app.layoutPanels()
			app.keymap = DefaultKeymap()
			tabs := make([]config.IssueTabConfig, count)
			for i := range tabs {
				tabs[i] = config.IssueTabConfig{Name: fmt.Sprintf("Collection %d", i)}
			}
			app.issuesList.SetTabs(tabs)

			layout := app.geometry()
			view := app.renderIssueTabs(layout.tabs.width, layout.tabs.height)
			lines := strings.Split(ansi.Strip(view), "\n")
			if len(lines) != layout.tabs.height {
				t.Fatalf("rendered %d lines, want %d: %q", len(lines), layout.tabs.height, view)
			}
			for i, line := range lines {
				if width := lipgloss.Width(line); width != layout.tabs.width {
					t.Errorf("line %d width=%d, want %d", i, width, layout.tabs.width)
				}
			}
			wantIndex := -1
			if count > 0 {
				wantIndex = 0
			}
			if index, overflow := app.issueTabAtRow(1); index != wantIndex || overflow != 0 {
				t.Fatalf("first content row maps to (%d,%d), want (%d,0)", index, overflow, wantIndex)
			}
			app.mouseClick(panelTabs, 1, 5)
			if count > 0 && app.issuesList.GetTabIndex() != 0 {
				t.Fatalf("first-row click selected tab %d, want 0", app.issuesList.GetTabIndex())
			}
		})
	}
}

func TestIssueTabsPane_TitleShowsConfiguredFocusKey(t *testing.T) {
	t.Parallel()
	app := appWithPanelDims(t, 120)
	app.keymap = DefaultKeymap()
	app.issuesList.SetTabs([]config.IssueTabConfig{{Name: "All"}})
	view := ansi.Strip(app.renderIssueTabs(app.geometry().tabs.width, app.geometry().tabs.height))
	if !strings.Contains(view, "[1] Issue tabs") {
		t.Fatalf("default title missing focus key: %q", view)
	}

	app.width, app.height = 80, 21
	app.layoutPanels()
	app.keymap[ActFocusIssueTabs] = []string{"F"}
	view = ansi.Strip(app.renderIssueTabs(app.geometry().tabs.width, app.geometry().tabs.height))
	if !strings.Contains(view, "[F] Issue tabs") {
		t.Fatalf("compact title missing remapped focus key: %q", view)
	}
}

func TestIssueTabsPane_OverflowIndicatorsAndClicks(t *testing.T) {
	t.Parallel()
	app := appWithPanelDims(t, 120)
	tabs := make([]config.IssueTabConfig, 12)
	for i := range tabs {
		tabs[i] = config.IssueTabConfig{Name: "Collection " + string(rune('A'+i))}
	}
	app.issuesList.SetTabs(tabs)

	view := app.renderIssueTabs(app.geometry().tabs.width, app.geometry().tabs.height)
	lines := strings.Split(ansi.Strip(view), "\n")
	if lipgloss.Width(lines[0]) != app.geometry().tabs.width {
		t.Fatalf("top border width = %d", lipgloss.Width(lines[0]))
	}
	if len(lines) != app.geometry().tabs.height || !strings.Contains(view, "↓ 6 more") {
		t.Fatalf("overflow pane = %q", view)
	}

	before := app.issuesList.GetTabIndex()
	app.mouseClick(panelTabs, app.geometry().tabs.height-2, 5)
	scrolledOffset := app.tabOffset
	if before != app.issuesList.GetTabIndex() || scrolledOffset == 0 {
		t.Fatalf("bottom overflow click activated tab or failed to scroll: index=%d offset=%d", app.issuesList.GetTabIndex(), scrolledOffset)
	}
	app.renderIssueTabs(app.geometry().tabs.width, app.geometry().tabs.height)
	if app.issuesList.GetTabIndex() != before || app.tabOffset != scrolledOffset {
		t.Fatalf("render changed overflow selection or offset: index=%d offset=%d", app.issuesList.GetTabIndex(), app.tabOffset)
	}
	app.mouseClick(panelTabs, 1, 5)
	if app.issuesList.GetTabIndex() != before || app.tabOffset != 0 {
		t.Fatalf("top overflow click activated tab or failed to scroll back: index=%d offset=%d", app.issuesList.GetTabIndex(), app.tabOffset)
	}
	app.switchIssueCollection(-1)
	if app.issuesList.GetTabIndex() != len(tabs)-1 || app.tabOffset != len(tabs)-6 {
		t.Fatalf("keyboard collection change did not reveal active tab: index=%d offset=%d", app.issuesList.GetTabIndex(), app.tabOffset)
	}
	lastWindow := ansi.Strip(app.renderIssueTabs(app.geometry().tabs.width, app.geometry().tabs.height))
	if strings.Contains(lastWindow, "↓ 0 more") {
		t.Fatalf("final overflow row advertises an empty scroll: %s", lastWindow)
	}
	if _, overflow := app.issueTabAtRow(app.geometry().tabs.height - 2); overflow != 0 {
		t.Fatalf("empty bottom overflow row is still clickable: overflow=%d", overflow)
	}
}

func TestIssueTabsPane_ResizeKeepsActiveCollectionVisible(t *testing.T) {
	t.Parallel()
	app := appWithPanelDims(t, 120)
	tabs := make([]config.IssueTabConfig, 12)
	for i := range tabs {
		tabs[i] = config.IssueTabConfig{Name: "Collection " + string(rune('A'+i))}
	}
	app.issuesList.SetTabs(tabs)
	app.tabOffset = 6
	app.renderIssueTabs(app.geometry().tabs.width, app.geometry().tabs.height)
	if app.tabOffset != 6 {
		t.Fatalf("initial render changed offset to %d", app.tabOffset)
	}

	app = resizeViewApp(t, app, 120, 24)
	if app.tabOffset != 0 {
		t.Fatalf("resize left active tab offscreen at offset %d", app.tabOffset)
	}
	view := ansi.Strip(app.renderIssueTabs(app.geometry().tabs.width, app.geometry().tabs.height))
	if !strings.Contains(view, "Collection A") {
		t.Fatalf("resized tab window does not show the active collection:\n%s", view)
	}
}

func TestIssueTabsOffsetSurvivesHiddenResizeUntilRestore(t *testing.T) {
	t.Parallel()
	app := appWithPanelDims(t, 120)
	tabs := make([]config.IssueTabConfig, 12)
	for i := range tabs {
		tabs[i] = config.IssueTabConfig{Name: "Collection " + string(rune('A'+i))}
	}
	app.issuesList.SetTabs(tabs)
	app.tabOffset = 6
	app.toggleMaximize(focusDetailPane)
	app = resizeViewApp(t, app, 120, 24)
	if app.tabOffset != 6 {
		t.Fatalf("resize changed the hidden tab offset to %d", app.tabOffset)
	}
	app.toggleMaximize(focusDetailPane)
	if app.tabOffset != 0 {
		t.Fatalf("restoring split did not reveal active collection: offset=%d", app.tabOffset)
	}
}

func TestIssueTabsPane_LongUnicodeLabelsFit(t *testing.T) {
	t.Parallel()
	app := appWithPanelDims(t, 120)
	app.issuesList.SetTabs([]config.IssueTabConfig{{Name: "界界界界界界界界界界"}, {Name: "Mine"}})
	view := app.renderIssueTabs(app.geometry().tabs.width, app.geometry().tabs.height)
	for i, line := range strings.Split(view, "\n") {
		if width := lipgloss.Width(line); width != app.geometry().tabs.width {
			t.Errorf("line %d width = %d, want %d", i, width, app.geometry().tabs.width)
		}
	}
}

func TestIssueTabsPane_KeyboardChangesCollectionAndEnterFocusesIssues(t *testing.T) {
	t.Parallel()
	app := appWithPanelDims(t, 120)
	app.keymap = DefaultKeymap()
	app.issuesList.SetTabs([]config.IssueTabConfig{{Name: "All"}, {Name: "Mine"}})
	app.side = sideLeft
	app.leftFocus = focusIssueTabs

	_, _ = app.handleKeyMsg(runeKey('j'))
	if app.issuesList.GetTabIndex() != 1 || app.leftFocus != focusIssueTabs {
		t.Fatalf("j changed tab/focus to (%d,%d)", app.issuesList.GetTabIndex(), app.leftFocus)
	}
	_, _ = app.handleKeyMsg(tea.KeyMsg{Type: tea.KeyEnter})
	if app.issuesList.GetTabIndex() != 1 || app.leftFocus != focusIssues {
		t.Fatalf("enter changed tab/focus to (%d,%d)", app.issuesList.GetTabIndex(), app.leftFocus)
	}
}

func TestFocusMap_DirectsUppercaseHJKLWithoutWrapping(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		startSide focusSide
		start     focusPanel
		key       string
		wantSide  focusSide
		want      focusPanel
	}{
		{"selector down", sideLeft, focusProjects, "J", sideLeft, focusIssueTabs},
		{"tabs down", sideLeft, focusIssueTabs, "J", sideLeft, focusInfo},
		{"tabs up", sideLeft, focusIssueTabs, "K", sideLeft, focusProjects},
		{"tabs right", sideLeft, focusIssueTabs, "L", sideLeft, focusIssues},
		{"issues left", sideLeft, focusIssues, "H", sideLeft, focusIssueTabs},
		{"issues down", sideLeft, focusIssues, "J", sideRight, focusIssues},
		{"issues up is no-op", sideLeft, focusIssues, "K", sideLeft, focusIssues},
		{"details left", sideRight, focusIssues, "H", sideLeft, focusInfo},
		{"details up", sideRight, focusIssues, "K", sideLeft, focusIssues},
		{"info up", sideLeft, focusInfo, "K", sideLeft, focusIssueTabs},
		{"info right", sideLeft, focusInfo, "L", sideRight, focusInfo},
		{"selector left is no-op", sideLeft, focusProjects, "H", sideLeft, focusProjects},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			app := appWithPanelDims(t, 120)
			app.keymap = DefaultKeymap()
			app.side, app.leftFocus = tt.startSide, tt.start
			_, _ = app.handleKeyMsg(runeKey(rune(tt.key[0])))
			testkit.AssertEqual(t, "side", app.side, tt.wantSide)
			testkit.AssertEqual(t, "left focus", app.leftFocus, tt.want)
		})
	}
}

func TestIssueCollectionsSwitchFromAnyPaneWithoutMovingFocus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		side focusSide
		pane focusPanel
	}{
		{"issue tabs", sideLeft, focusIssueTabs},
		{"issues", sideLeft, focusIssues},
		{"info", sideLeft, focusInfo},
		{"projects", sideLeft, focusProjects},
		{"details", sideRight, focusIssues},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			app := appWithPanelDims(t, 120)
			app.keymap = DefaultKeymap()
			app.issuesList.SetTabs([]config.IssueTabConfig{{Name: "All"}, {Name: "Mine"}})
			app.side, app.leftFocus = tt.side, tt.pane
			app.previewKey = "OLD-1"
			app.detailView.SetIssue(&jira.Issue{Key: "OLD-1"})
			app.maximized = tt.name == "details"
			if app.maximized {
				app.maximizedPane = focusDetailPane
			}

			_, _ = app.Update(tea.KeyMsg{Type: tea.KeyTab})

			if app.issuesList.GetTabIndex() != 1 || app.side != tt.side || app.leftFocus != tt.pane || app.maximized != (tt.name == "details") {
				t.Fatalf("tab/focus/max = (%d,%d,%d,%v)", app.issuesList.GetTabIndex(), app.side, app.leftFocus, app.maximized)
			}
			if app.previewKey != "" || app.detailView.IssueKey() != "" {
				t.Fatalf("old issue preview remains: key=%q detail=%q", app.previewKey, app.detailView.IssueKey())
			}
		})
	}
}

func TestFocusActionRestoresSplitWhenTargetIsHidden(t *testing.T) {
	t.Parallel()
	app := appWithPanelDims(t, 120)
	app.keymap = DefaultKeymap()
	app.maximized = true
	app.maximizedPane = focusIssues
	app.side = sideLeft
	app.leftFocus = focusIssues

	_, _ = app.handleKeyMsg(runeKey('4'))

	if app.maximized || app.side != sideRight {
		t.Fatalf("target detail focus left maximized=%v side=%v", app.maximized, app.side)
	}
}

func TestInfoMouseAndKeyboardTabsRequestSamePreview(t *testing.T) {
	t.Parallel()
	issue := &jira.Issue{
		Key: "MAIN-1",
		IssueLinks: []jira.IssueLink{{
			Type:         &jira.IssueLinkType{Outward: "blocks"},
			OutwardIssue: &jira.Issue{Key: "LINK-1", Summary: "linked"},
		}},
	}
	makeApp := func(t *testing.T) *App {
		t.Helper()
		app := appWithPanelDims(t, 120)
		app.keymap = DefaultKeymap()
		app.infoPanel.SetIssue(issue)
		app.side = sideLeft
		app.leftFocus = focusInfo
		return app
	}
	request := func(cmd tea.Cmd) string {
		t.Helper()
		if cmd == nil {
			t.Fatal("tab change returned no preview request")
		}
		msg, ok := cmd().(views.PreviewRequestMsg)
		if !ok {
			t.Fatalf("preview request = %T", cmd())
		}
		return msg.Key
	}

	mouseApp := makeApp(t)
	_, mouseCmd := mouseApp.mouseClick(panelInfo, 0, 13)
	keyboardApp := makeApp(t)
	_, keyboardCmd := keyboardApp.handleKeyMsg(runeKey(']'))
	if mouseKey, keyboardKey := request(mouseCmd), request(keyboardCmd); mouseKey != keyboardKey || mouseKey != "LINK-1" {
		t.Fatalf("mouse/keyboard preview = (%q,%q)", mouseKey, keyboardKey)
	}
}

func TestMaximizeBindingCanBeRemapped(t *testing.T) {
	t.Parallel()
	km := KeymapFromConfig(config.KeybindingConfig{Universal: config.UniversalKeys{ToggleMaximize: "ctrl+m"}})
	if got := km.Match("ctrl+m"); got != ActToggleMaximize {
		t.Fatalf("ctrl+m action = %q", got)
	}
}

func TestEmptyIssueCollectionClearsPreview(t *testing.T) {
	t.Parallel()
	app := appWithPanelDims(t, 120)
	app.keymap = DefaultKeymap()
	app.issuesList.SetTabs([]config.IssueTabConfig{{Name: "All"}, {Name: "Empty"}})
	app.issuesList.SetIssues([]jira.Issue{{Key: "OLD-1"}})
	app.detailView.SetIssue(&jira.Issue{Key: "OLD-1"})
	app.previewKey = "OLD-1"
	app.issuesList.NextTab()
	app.issuesList.SetIssues(nil)

	_, _ = app.handleIssuesLoaded(issuesLoadedMsg{tab: 1})

	if app.previewKey != "" || app.detailView.IssueKey() != "" {
		t.Fatalf("empty collection retained preview %q", app.previewKey)
	}
}

func TestIssueColumnMinimumWidthKeepsEveryConfiguredColumnVisible(t *testing.T) {
	t.Parallel()
	app := appWithPanelDims(t, 80)
	fields := []string{"key", "status", "type", "priority", "summary", "assignee", "updated"}
	app.issuesList.SetFields(fields)
	app.issuesList.SetIssues([]jira.Issue{{
		Key:       "PROJECT-123",
		Summary:   "A wide summary",
		Status:    &jira.Status{Name: "In Progress"},
		IssueType: &jira.IssueType{Name: "Story"},
		Priority:  &jira.Priority{Name: "High"},
		Assignee:  &jira.User{DisplayName: "Ada"},
	}})
	layout := app.geometry()
	if !layout.tooSmall || layout.requiredWidth <= 80 {
		t.Fatalf("all-column layout = %+v, want a larger required width", layout)
	}
	app.width = layout.requiredWidth
	app.layoutPanels()
	view := ansi.Strip(app.issuesList.View())
	for _, field := range []string{"Key", "Status", "Type", "Priority", "Summary", "Assignee", "Updated", "PROJECT-123", "Ada"} {
		if !strings.Contains(view, field) {
			t.Errorf("issue view missing %q:\n%s", field, view)
		}
	}
}

func TestMaximizeClickAndKeyPreserveSelection(t *testing.T) {
	t.Parallel()
	app := appWithPanelDims(t, 120)
	app.keymap = DefaultKeymap()
	issues := make([]jira.Issue, 60)
	for i := range issues {
		issues[i] = jira.Issue{Key: fmt.Sprintf("A-%d", i)}
	}
	app.issuesList.SetIssues(issues)
	app.issuesList.ScrollBy(30)
	cursor, offset := app.issuesList.Cursor, app.issuesList.Offset
	app.side, app.leftFocus = sideLeft, focusInfo
	app.updateFocusHints()
	app.updateFocusState()
	layout := app.geometry()
	title := ansi.Strip(app.issuesList.View())
	buttonPrefix, _, hasButton := strings.Cut(title, "[+]")
	if !hasButton {
		t.Fatalf("Issues title has no maximize button: %q", title)
	}
	buttonX := layout.issues.x + lipgloss.Width(buttonPrefix)
	_, _ = app.handleMouse(tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress, X: buttonX, Y: layout.issues.y})
	if !app.maximized || app.maximizedPane != focusIssues {
		panel, relY := app.hitTest(buttonX, layout.issues.y)
		t.Fatalf("clicking the unfocused Issues maximize target did not maximize Issues: panel=%d relY=%d x=%d layout=%+v", panel, relY, buttonX, layout)
	}
	_, _ = app.handleKeyMsg(runeKey('+'))
	if app.maximized || app.issuesList.Cursor != cursor || app.issuesList.Offset != offset {
		t.Fatalf("same-size restore changed issue position: max=%v cursor=%d offset=%d, want cursor=%d offset=%d", app.maximized, app.issuesList.Cursor, app.issuesList.Offset, cursor, offset)
	}
	_, _ = app.handleKeyMsg(runeKey('+'))
	if !app.maximized {
		t.Fatal("plus did not maximize Issues")
	}
	app = resizeViewApp(t, app, 120, 35)
	_, _ = app.handleKeyMsg(runeKey('+'))
	visible := app.issuesList.VisibleRows()
	if app.maximized || app.issuesList.Cursor != cursor || app.issuesList.Offset == 0 || cursor < app.issuesList.Offset || cursor >= app.issuesList.Offset+visible {
		t.Fatalf("resize/restore lost visible issue position: max=%v cursor=%d offset=%d visible=%d, want cursor=%d in view", app.maximized, app.issuesList.Cursor, app.issuesList.Offset, visible, cursor)
	}
	if selected := app.issuesList.SelectedIssue(); selected == nil || selected.Key != fmt.Sprintf("A-%d", cursor) {
		t.Fatalf("selected issue after resize/restore = %v, want A-%d", selected, cursor)
	}
}

func TestDetailMaximizeTitleClick(t *testing.T) {
	t.Parallel()
	app := appWithPanelDims(t, 120)
	app.keymap = DefaultKeymap()
	app.detailView.SetIssue(&jira.Issue{Key: "A-1"})
	app.side, app.leftFocus = sideLeft, focusIssues
	app.updateFocusHints()
	app.updateFocusState()
	layout := app.geometry()
	title := ansi.Strip(app.detailView.View())
	buttonPrefix, _, hasButton := strings.Cut(title, "[+]")
	if !hasButton {
		t.Fatalf("Details title has no maximize button: %q", title)
	}
	x := layout.detail.x + lipgloss.Width(buttonPrefix)
	_, _ = app.handleMouse(tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress, X: x, Y: layout.detail.y})
	if !app.maximized || app.maximizedPane != focusDetailPane {
		t.Fatal("clicking the Details maximize target did not maximize Details")
	}
}

func TestMaximizedIssuesDoesNotScrollHiddenDetails(t *testing.T) {
	t.Parallel()
	app := appWithPanelDims(t, 120)
	app.keymap = DefaultKeymap()
	app.detailView.SetIssue(&jira.Issue{Key: "A-1", Description: strings.Repeat("line\n", 30)})
	app.detailView.ScrollBy(2)
	before := app.detailView.View()
	app.maximized = true
	app.maximizedPane = focusIssues
	app.side = sideLeft
	app.leftFocus = focusIssues

	_, _ = app.handleKeyMsg(tea.KeyMsg{Type: tea.KeyCtrlD})

	if after := app.detailView.View(); after != before {
		t.Fatal("Ctrl+D scrolled hidden Details while Issues was maximized")
	}
}

func TestFocusRestorationAfterMaximizedIssueOpen(t *testing.T) {
	t.Parallel()
	app := appWithPanelDims(t, 120)
	app.keymap = DefaultKeymap()
	app.issuesList.SetIssues([]jira.Issue{{Key: "OLD-1"}})
	app.maximized = true
	app.maximizedPane = focusIssues
	app.leftFocus = focusIssues

	_, _ = app.handleKeyMsg(tea.KeyMsg{Type: tea.KeyEnter})

	if app.maximized || app.side != sideRight {
		t.Fatalf("open left maximized=%v side=%v", app.maximized, app.side)
	}
}
