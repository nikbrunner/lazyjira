package tui

import tea "github.com/charmbracelet/bubbletea"

const focusDetailPane focusPanel = -1

func (a *App) handleSpatialFocus(key string) (tea.Model, tea.Cmd, bool) {
	if key != "H" && key != "J" && key != "K" && key != "L" {
		return nil, nil, false
	}
	current := a.leftFocus
	if a.side == sideRight {
		current = focusDetailPane
	}
	targets := map[focusPanel]map[string]focusPanel{
		focusProjects:   {"J": focusIssueTabs},
		focusIssueTabs:  {"J": focusInfo, "K": focusProjects, "L": focusIssues},
		focusIssues:     {"H": focusIssueTabs, "J": focusDetailPane},
		focusDetailPane: {"H": focusInfo, "K": focusIssues},
		focusInfo:       {"K": focusIssueTabs, "L": focusDetailPane},
	}
	if target, ok := targets[current][key]; ok {
		a.focusPane(target)
	}
	return a, nil, true
}

func (a *App) updateFocusHints() {
	a.issuesList.SetFocusHint(a.keymap.Keys(ActFocusIssues))
	a.detailView.SetFocusHint(a.keymap.Keys(ActFocusDetail))
	a.infoPanel.SetFocusHint(a.keymap.Keys(ActFocusInfo))
	a.projectList.SetFocusHint(a.keymap.Keys(ActFocusProj))
	a.issuesList.SetMaximized(a.maximized && a.maximizedPane == focusIssues)
	a.detailView.SetMaximized(a.maximized && a.maximizedPane == focusDetailPane)
}

func (a *App) toggleMaximize(target focusPanel) {
	if a.maximized && a.maximizedPane == target {
		a.maximized = false
	} else {
		a.maximized = true
		a.maximizedPane = target
	}
	if target == focusDetailPane {
		a.side = sideRight
	} else {
		a.side = sideLeft
		a.leftFocus = target
	}
	a.updateFocusHints()
	a.updateFocusState()
	a.ensureActiveIssueTabVisible()
}

func (a *App) focusPane(target focusPanel) {
	if a.maximized && target != a.maximizedPane {
		a.maximized = false
		a.updateFocusHints()
	}
	if target == focusDetailPane {
		a.side = sideRight
	} else {
		a.side = sideLeft
		a.leftFocus = target
	}
	a.updateFocusState()
	a.ensureActiveIssueTabVisible()
}

func (a *App) switchIssueCollection(direction int) tea.Cmd {
	if len(a.issuesList.Tabs()) == 0 {
		return nil
	}
	if direction < 0 {
		a.issuesList.PrevTab()
	} else {
		a.issuesList.NextTab()
	}
	a.ensureActiveIssueTabVisible()
	return a.activateIssueCollection()
}
