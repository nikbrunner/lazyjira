package tui

const (
	statusHeight = 3
	logHeight    = 5
	helpHeight   = 1
	minWidth     = 80
	minHeight    = 21
)

type rect struct {
	x, y, width, height int
}

type appLayout struct {
	status, tabs, issues, info, projects, detail, log, help rect
	requiredWidth, requiredHeight                           int
	tooSmall                                                bool
}

func (a *App) configuredSideWidth() int {
	width := a.cfg.GUI.SidePanelWidth
	if width <= 0 {
		width = 22
	}
	return max(18, width)
}

func sideWidthForTerminal(configuredWidth, terminalWidth int) int {
	if terminalWidth < 120 {
		configuredWidth = max(18, min(configuredWidth, terminalWidth*35/100))
	}
	return configuredWidth
}

func (a *App) sideWidth() int {
	width := sideWidthForTerminal(a.configuredSideWidth(), a.width)
	return min(width, max(18, a.width-a.issuesList.MinimumWidth()))
}

func (a *App) minimumTerminalWidth() int {
	issuesWidth := a.issuesList.MinimumWidth()
	for width := minWidth; width < 120; width++ {
		if width >= sideWidthForTerminal(a.configuredSideWidth(), width)+issuesWidth {
			return width
		}
	}
	return max(120, a.configuredSideWidth()+issuesWidth)
}

func (a *App) geometry() appLayout {
	layout := appLayout{}
	if a.width <= 0 || a.height <= 0 {
		return layout
	}
	layout.help = rect{0, a.height - helpHeight, a.width, helpHeight}
	layout.requiredWidth = a.minimumTerminalWidth()
	layout.requiredHeight = minHeight
	layout.tooSmall = a.width < layout.requiredWidth || a.height < layout.requiredHeight
	if layout.tooSmall {
		return layout
	}
	if a.maximized {
		maximized := rect{0, 0, a.width, a.height - helpHeight}
		if a.maximizedPane == focusIssues {
			layout.issues = maximized
		} else {
			layout.detail = maximized
		}
		return layout
	}

	sideWidth := a.sideWidth()
	mainWidth := a.width - sideWidth
	layout.projects = rect{0, 0, sideWidth, statusHeight}
	layout.status = rect{sideWidth, 0, mainWidth, statusHeight}
	layout.log = rect{0, a.height - helpHeight - logHeight, a.width, logHeight}
	bodyY := statusHeight
	bodyHeight := layout.log.y - bodyY
	issuesHeight := bodyHeight / 3
	detailHeight := bodyHeight - issuesHeight
	layout.tabs = rect{0, bodyY, sideWidth, issuesHeight}
	layout.issues = rect{sideWidth, bodyY, mainWidth, issuesHeight}
	layout.info = rect{0, bodyY + issuesHeight, sideWidth, detailHeight}
	layout.detail = rect{sideWidth, bodyY + issuesHeight, mainWidth, detailHeight}
	return layout
}

func (a *App) layoutPanels() {
	layout := a.geometry()
	if layout.tooSmall {
		return
	}
	if layout.status.width > 0 {
		a.statusPanel.SetSize(layout.status.width, layout.status.height)
	}
	if layout.issues.width > 0 {
		a.issuesList.SetSize(layout.issues.width, layout.issues.height)
	}
	if layout.info.width > 0 {
		a.infoPanel.SetSize(layout.info.width, layout.info.height)
	}
	if layout.projects.width > 0 {
		a.projectList.SetSize(layout.projects.width, layout.projects.height)
	}
	if layout.detail.width > 0 {
		a.detailView.SetSize(layout.detail.width, layout.detail.height)
	}
	if layout.log.width > 0 {
		a.logPanel.SetSize(layout.log.width, layout.log.height)
	}
}
