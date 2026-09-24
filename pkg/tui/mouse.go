package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/nikbrunner/lazyjira/v2/pkg/tui/views"
)

type panelID int

const (
	panelNone panelID = iota
	panelStatus
	panelTabs
	panelIssues
	panelInfo
	panelProjects
	panelDetail
	panelLog
)

func (a *App) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	x, y := msg.X, msg.Y
	panel, relY := a.hitTest(x, y)

	switch {
	case msg.Button == tea.MouseButtonWheelUp:
		return a.mouseScroll(panel, -3)
	case msg.Button == tea.MouseButtonWheelDown:
		return a.mouseScroll(panel, 3)
	case msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft:
		return a.mouseClick(panel, relY, x)
	}
	return a, nil
}

// hitTest uses the same rectangles as composition and sizing.
func (a *App) hitTest(x, y int) (panelID, int) {
	layout := a.geometry()
	panels := []struct {
		id   panelID
		area rect
	}{
		{panelStatus, layout.status},
		{panelTabs, layout.tabs},
		{panelIssues, layout.issues},
		{panelInfo, layout.info},
		{panelProjects, layout.projects},
		{panelDetail, layout.detail},
		{panelLog, layout.log},
	}
	for _, panel := range panels {
		area := panel.area
		if area.width > 0 && area.height > 0 && x >= area.x && x < area.x+area.width && y >= area.y && y < area.y+area.height {
			return panel.id, y - area.y
		}
	}
	return panelNone, 0
}

func (a *App) mouseScroll(panel panelID, delta int) (tea.Model, tea.Cmd) {
	switch panel { //nolint:exhaustive
	case panelTabs:
		a.side = sideLeft
		a.leftFocus = focusIssueTabs
		a.tabOffset = max(0, a.tabOffset+delta)
		a.updateFocusState()
	case panelIssues:
		if a.side != sideLeft || a.leftFocus != focusIssues {
			a.side = sideLeft
			a.leftFocus = focusIssues
			a.updateFocusState()
		}
		if delta > 0 {
			a.issuesList.ScrollBy(1)
		} else {
			a.issuesList.ScrollBy(-1)
		}
		if sel := a.issuesList.SelectedIssue(); sel != nil {
			return a.Update(views.IssueSelectedMsg{Issue: sel})
		}
	case panelInfo:
		if a.side != sideLeft || a.leftFocus != focusInfo {
			a.side = sideLeft
			a.leftFocus = focusInfo
			a.updateFocusState()
		}
		if delta > 0 {
			a.infoPanel.ScrollBy(1)
		} else {
			a.infoPanel.ScrollBy(-1)
		}
	case panelProjects:
		// The selector opens a picker on click; wheel input has no selection preview.
	case panelDetail:
		if a.side != sideRight {
			a.side = sideRight
			a.updateFocusState()
		}
		if delta > 0 {
			a.detailView.ScrollBy(1)
		} else {
			a.detailView.ScrollBy(-1)
		}
	}
	return a, nil
}

func (a *App) mouseClick(panel panelID, relY int, x int) (tea.Model, tea.Cmd) {
	switch panel { //nolint:exhaustive
	case panelStatus:
		// App status is display-only.

	case panelTabs:
		a.side = sideLeft
		a.leftFocus = focusIssueTabs
		a.updateFocusState()
		if index, overflow := a.issueTabAtRow(relY); overflow != 0 {
			a.tabOffset = max(0, a.tabOffset+overflow)
		} else if index >= 0 && index != a.issuesList.GetTabIndex() {
			a.setIssueTabIndex(index)
			return a, a.activateIssueCollection()
		}

	case panelIssues:
		a.side = sideLeft
		a.leftFocus = focusIssues
		a.updateFocusState()
		if relY == 0 {
			if a.issuesList.ClickMaximizeAt(x - a.geometry().issues.x) {
				a.toggleMaximize(focusIssues)
			}
		} else if dbl := a.issuesList.ClickAt(relY); dbl {
			return a.openIssueDetail()
		} else if sel := a.issuesList.SelectedIssue(); sel != nil {
			return a.Update(views.IssueSelectedMsg{Issue: sel})
		}

	case panelInfo:
		a.side = sideLeft
		a.leftFocus = focusInfo
		a.updateFocusState()
		if relY == 0 {
			a.infoPanel.ClickTabAt(x)
			return a, tea.Batch(a.previewForInfoTab(), a.infoPanel.MaybeChildrenRequest())
		}
		a.infoPanel.ClickAt(relY)

	case panelProjects:
		a.focusPane(focusProjects)
		a.openProjectPicker()

	case panelDetail:
		a.side = sideRight
		a.updateFocusState()
		if relY == 0 {
			relX := x - a.geometry().detail.x
			if a.detailView.ClickMaximizeAt(relX) {
				a.toggleMaximize(focusDetailPane)
			} else {
				a.detailView.ClickTab(relX)
			}
		} else {
			if cmd := a.detailView.ClickItem(relY); cmd != nil {
				return a, cmd
			}
		}
	}
	return a, nil
}
