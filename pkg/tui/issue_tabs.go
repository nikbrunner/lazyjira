package tui

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/nikbrunner/lazyjira/v2/pkg/tui/components"
	"github.com/nikbrunner/lazyjira/v2/pkg/tui/theme"
)

func issueTabVisibleCount(tabCount, innerHeight int) int {
	visible := innerHeight
	if tabCount > visible && innerHeight > 2 {
		visible = innerHeight - 2
	}
	return visible
}

func (a *App) setIssueTabIndex(index int) {
	a.issuesList.SetTabIndex(index)
	a.ensureActiveIssueTabVisible()
}

func (a *App) ensureActiveIssueTabVisible() {
	tabs := a.issuesList.Tabs()
	if len(tabs) == 0 {
		return
	}
	layout := a.geometry()
	if layout.tooSmall || layout.tabs.height == 0 {
		return
	}
	innerHeight := max(0, layout.tabs.height-2)
	visible := issueTabVisibleCount(len(tabs), innerHeight)
	if len(tabs) <= innerHeight {
		a.tabOffset = 0
		return
	}
	index := a.issuesList.GetTabIndex()
	lastOffset := max(0, len(tabs)-visible)
	a.tabOffset = min(max(a.tabOffset, 0), lastOffset)
	a.tabOffset = min(a.tabOffset, index)
	if index >= a.tabOffset+visible {
		a.tabOffset = index - visible + 1
	}
}

func (a *App) issueTabsTitle(width int) string {
	title := "Issue tabs"
	if hint := a.keymap.Keys(ActFocusIssueTabs); hint != "" {
		title = "[" + hint + "] " + title
	}
	return ansi.Truncate(title, max(width-3, 0), "…")
}

func (a *App) renderIssueTabs(width, height int) string {
	innerWidth := max(0, width-2)
	title := a.issueTabsTitle(width)
	innerHeight := max(0, height-2)
	tabs := a.issuesList.Tabs()
	rows := make([]string, innerHeight)
	if innerHeight == 0 {
		return ""
	}
	if len(tabs) == 0 {
		rows[0] = "No issue tabs"
	} else {
		visible := issueTabVisibleCount(len(tabs), innerHeight)
		lastOffset := max(0, len(tabs)-visible)
		a.tabOffset = min(max(a.tabOffset, 0), lastOffset)
		if innerHeight <= 2 {
			for row := range min(visible, innerHeight, len(tabs)-a.tabOffset) {
				index := a.tabOffset + row
				label := "  " + tabs[index].Name
				if index == a.issuesList.GetTabIndex() {
					label = theme.Default.Title.Render("› " + tabs[index].Name)
				}
				rows[row] = ansi.Truncate(label, innerWidth, "…")
			}
			focused := a.side == sideLeft && a.leftFocus == focusIssueTabs
			return components.RenderPanel(title, strings.Join(rows, "\n"), width, innerHeight, focused)
		}
		row := 0
		if len(tabs) > visible {
			if a.tabOffset > 0 {
				rows[row] = "↑ " + strconv.Itoa(a.tabOffset) + " more"
			}
			row++
		}
		for i := a.tabOffset; i < min(len(tabs), a.tabOffset+visible) && row < innerHeight; i++ {
			label := "  " + tabs[i].Name
			if i == a.issuesList.GetTabIndex() {
				label = theme.Default.Title.Render("› " + tabs[i].Name)
			}
			rows[row] = ansi.Truncate(label, innerWidth, "…")
			row++
		}
		remaining := len(tabs) - a.tabOffset - visible
		if remaining > 0 && row < innerHeight {
			rows[innerHeight-1] = "↓ " + strconv.Itoa(remaining) + " more"
		}
	}
	focused := a.side == sideLeft && a.leftFocus == focusIssueTabs
	return components.RenderPanel(title, strings.Join(rows, "\n"), width, innerHeight, focused)
}

func (a *App) issueTabAtRow(y int) (index, overflow int) {
	layout := a.geometry()
	innerHeight := max(0, layout.tabs.height-2)
	tabs := a.issuesList.Tabs()
	if innerHeight <= 2 {
		index = a.tabOffset + y - 1
		if index >= 0 && index < len(tabs) && index < a.tabOffset+issueTabVisibleCount(len(tabs), innerHeight) {
			return index, 0
		}
		return -1, 0
	}
	if len(tabs) <= innerHeight {
		index = a.tabOffset + y - 1
		if index >= 0 && index < len(tabs) {
			return index, 0
		}
		return -1, 0
	}
	visible := issueTabVisibleCount(len(tabs), innerHeight)
	lastOffset := max(0, len(tabs)-visible)
	offset := min(max(a.tabOffset, 0), lastOffset)
	contentRow := y - 1
	if contentRow == 0 {
		if offset > 0 {
			return -1, -visible
		}
		return -1, 0
	}
	if contentRow == innerHeight-1 {
		if offset < lastOffset {
			return -1, visible
		}
		return -1, 0
	}
	index = offset + contentRow - 1
	if index >= 0 && index < len(tabs) && index < offset+visible {
		return index, 0
	}
	return -1, 0
}

func (a *App) activateIssueCollection() tea.Cmd {
	if selected := a.issuesList.SelectedIssue(); selected != nil {
		cmd := a.previewSelectedIssue()
		if !a.issuesList.HasCachedTab() {
			return tea.Batch(a.fetchActiveTab(), cmd)
		}
		return cmd
	}
	a.clearIssuePreview()
	if !a.issuesList.HasCachedTab() {
		return a.fetchActiveTab()
	}
	return nil
}

func (a *App) clearIssuePreview() {
	a.previewEpoch++
	a.previewKey = ""
	a.detailView.SetIssue(nil)
	a.infoPanel.SetIssue(nil)
}
