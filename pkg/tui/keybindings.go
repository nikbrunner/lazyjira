package tui

import (
	"slices"

	"github.com/nikbrunner/lazyjira/pkg/config"
	"github.com/nikbrunner/lazyjira/pkg/tui/components"
	"github.com/nikbrunner/lazyjira/pkg/tui/views"
)

// Binding represents a single keybinding with context
type Binding struct {
	Key         string
	Description string
}

func (a *App) bind(action Action, desc string) Binding {
	return Binding{a.keymap.Keys(action), desc}
}

func (a *App) navBindings() []Binding {
	return []Binding{
		a.bind(ActNavDown, "navigate down"),
		a.bind(ActNavUp, "navigate up"),
		a.bind(ActNavTop, "go to top"),
		a.bind(ActNavBottom, "go to bottom"),
		a.bind(ActNavHalfDown, "half-page down"),
		a.bind(ActNavHalfUp, "half-page up"),
	}
}

func (a *App) detailScrollBindings() []Binding {
	return []Binding{
		a.bind(ActDetailScrollDown, "scroll detail down"),
		a.bind(ActDetailScrollUp, "scroll detail up"),
		a.bind(ActDetailHalfDown, "half-page detail down"),
		a.bind(ActDetailHalfUp, "half-page detail up"),
	}
}

// ContextBindings returns the focused pane's bindings, then the global ones.
func (a *App) ContextBindings() []Binding {
	return append(boundOnly(a.localBindings()), boundOnly(a.globalBindings())...)
}

// boundOnly drops actions that have no key in the active keymap.
func boundOnly(bindings []Binding) []Binding {
	return slices.DeleteFunc(bindings, func(b Binding) bool { return b.Key == "" })
}

func (a *App) globalBindings() []Binding {
	km := a.keymap
	return []Binding{
		{km.Keys(ActQuit), "quit"},
		{km.Keys(ActFocusProj), "focus Project selector"},
		{km.Keys(ActFocusIssueTabs), "focus Issue tabs"},
		{km.Keys(ActFocusIssues), "focus Issues"},
		{km.Keys(ActFocusInfo), "focus Issue info"},
		{km.Keys(ActFocusDetail), "focus Issue details"},
		{km.Keys(ActToggleMaximize), "maximize focused pane"},
		{"tab/shift-tab", "switch issue collection"},
		{"H/J/K/L", "move focus"},
		{km.Keys(ActSearch), "search / filter current list"},
		{km.Keys(ActRefresh), "refresh data from Jira"},
		a.bind(ActRefreshAll, "refresh all data"),
		a.bind(ActJQLSearch, "JQL search"),
		a.bind(ActIssueLookup, "open issue by key"),
		{km.Keys(ActHelp), "show all keybindings"},
	}
}

func (a *App) localBindings() []Binding {
	switch {
	case a.side == sideLeft && a.leftFocus == focusIssueTabs:
		return slices.Concat([]Binding{{"enter", "focus Issues"}, {"j/k", "switch issue collection"},
			a.bind(ActJQLSearch, "JQL search from tab query")}, a.navBindings())

	case a.side == sideLeft && a.leftFocus == focusIssues:
		return slices.Concat([]Binding{
			a.bind(ActOpen, "open issue detail"),
			a.bind(ActFocusRight, "open issue detail"),
			a.bind(ActInfoTab, "focus Issue info"),
			a.bind(ActSelect, "mark issue"),
			a.bind(ActVisualSelect, "mark range"),
			a.bind(ActShowChildren, "show children"),
			a.bind(ActShowParent, "show parent"),
			a.bind(ActFilterPicker, "filter by status, type, priority"),
			a.bind(ActCopySummary, "copy key and summary, or marked rows"),
			a.bind(ActCopyMarkdownLink, "copy Markdown link"),
			a.bind(ActCopyURL, "copy URL"),
			a.bind(ActTransition, "transition issue status"),
			a.bind(ActEdit, "edit issue"),
			a.bind(ActComments, "go to comments"),
			a.bind(ActPriority, "change priority"),
			a.bind(ActAssignee, "change assignee"),
			a.bind(ActBrowser, "open issue in browser"),
			a.bind(ActURLPicker, "open URL picker"),
			a.bind(ActCopyBranchName, "copy branch name"),
			a.bind(ActCreateBranch, "create branch"),
			a.bind(ActCopyWorktreeName, "copy worktree name"),
			a.bind(ActCreateWorktree, "create worktree"),
			a.bind(ActNew, "create issue"),
			a.bind(ActCreateSubtask, "create subtask"),
			a.bind(ActFocusLeft, "clear marks, then filters"),
			a.bind(ActDuplicateIssue, "duplicate issue"),
			a.bind(ActCloseJQLTab, "close JQL tab"),
			{"[]", "switch tab"},
		}, a.customCommandBindings(config.CtxIssues), a.navBindings(), a.detailScrollBindings())

	case a.side == sideLeft && a.leftFocus == focusInfo:
		return slices.Concat([]Binding{
			{"[]", "switch tab (Info/Lnk/Sub)"},
			a.bind(ActEdit, "edit field"),
			a.bind(ActTransition, "transition issue status"),
			a.bind(ActPriority, "change priority"),
			a.bind(ActAssignee, "change assignee"),
			a.bind(ActBrowser, "open issue in browser"),
			a.bind(ActURLPicker, "open URL picker"),
			a.bind(ActCopyBranchName, "copy branch name"),
			a.bind(ActCopyWorktreeName, "copy worktree name"),
			a.bind(ActCreateWorktree, "create worktree"),
			a.bind(ActCreateSubtask, "create subtask (Sub tab)"),
			a.bind(ActFocusRight, "next panel"),
			a.bind(ActFocusLeft, "previous panel"),
		}, a.customCommandBindings(config.CtxInfo), a.navBindings(), a.detailScrollBindings())

	case a.side == sideLeft && a.leftFocus == focusProjects:
		return slices.Concat([]Binding{a.bind(ActOpen, "choose project")},
			a.customCommandBindings(config.CtxProjects), a.navBindings())

	case a.side == sideRight:
		bindings := []Binding{
			{"[]", "previous/next tab"},
			a.bind(ActFocusLeft, "back to left panel"),
			a.bind(ActInfoTab, "focus info panel"),
			a.bind(ActPriority, "change priority"),
			a.bind(ActAssignee, "change assignee"),
			a.bind(ActBrowser, "open in browser"),
			a.bind(ActURLPicker, "open URL picker"),
			a.bind(ActCopyBranchName, "copy branch name"),
			a.bind(ActCopyWorktreeName, "copy worktree name"),
			a.bind(ActCreateWorktree, "create worktree"),
		}
		if a.detailView.ActiveTab() == views.TabComments {
			bindings = append(bindings,
				a.bind(ActEdit, "edit comment"),
				a.bind(ActNew, "new comment"),
			)
		} else {
			bindings = append(bindings,
				a.bind(ActEdit, "edit issue"),
			)
		}
		bindings = append(bindings, a.customCommandBindings(config.CtxDetail)...)
		if a.detailView.ActiveTab() == views.TabComments {
			bindings = append(bindings, a.customCommandBindings(config.CtxDetailComments)...)
		}
		return append(bindings, a.navBindings()...)
	}

	return nil
}

func (a *App) helpBarItems() []components.HelpItem {
	// Overlay-specific hints take priority over panel hints
	switch {
	case a.createForm.IsVisible():
		save := "create"
		if a.createCtx.editKey != "" {
			save = "save"
		}
		items := []components.HelpItem{
			{Key: "tab", Description: "next panel"},
			{Key: "ctrl+s", Description: save},
			{Key: "esc", Description: "cancel"},
		}
		switch a.createForm.FocusedPanel() { //nolint:exhaustive
		case components.CreatePanelFields:
			items = append(items,
				components.HelpItem{Key: "e", Description: "edit"},
				components.HelpItem{Key: "/", Description: "filter"},
			)
		case components.CreatePanelDescription:
			items = append(items,
				components.HelpItem{Key: "ctrl+g", Description: "edit in $EDITOR"},
				components.HelpItem{Key: "ctrl+v", Description: "paste image"},
			)
		}
		return items
	case a.jqlModal.IsVisible():
		return []components.HelpItem{
			{Key: "enter", Description: "search"},
			{Key: "tab", Description: "switch focus"},
			{Key: "esc", Description: "cancel"},
		}
	case a.showHelp:
		return []components.HelpItem{
			{Key: "/", Description: "filter"},
			{Key: "esc", Description: "close"},
		}
	case a.diffView.IsVisible():
		return []components.HelpItem{
			{Key: "enter", Description: "confirm"},
			{Key: "esc", Description: "cancel"},
		}
	case a.inputModal.IsVisible():
		items := []components.HelpItem{
			{Key: "enter", Description: "confirm"},
			{Key: "esc", Description: "cancel"},
		}
		if a.inputModal.HasHints() {
			items = append(items, components.HelpItem{Key: "tab", Description: "existing branches"})
		}
		return items
	case a.modal.IsVisible() && a.modal.IsChecklist():
		return []components.HelpItem{
			{Key: "space", Description: "toggle"},
			{Key: "/", Description: "search"},
			{Key: "enter", Description: "confirm"},
			{Key: "esc", Description: "cancel"},
		}
	case a.modal.IsVisible():
		return []components.HelpItem{
			{Key: "/", Description: "search"},
			{Key: "enter", Description: "select"},
			{Key: "esc", Description: "cancel"},
		}
	}

	km := a.keymap
	return []components.HelpItem{
		{Key: km.Keys(ActSearch), Description: "search"},
		{Key: km.Keys(ActJQLSearch), Description: "JQL search"},
		{Key: km.Keys(ActIssueLookup), Description: "open by key"},
		{Key: km.Keys(ActRefresh), Description: "refresh"},
		{Key: km.Keys(ActHelp), Description: "keybindings"},
		{Key: km.Keys(ActQuit), Description: "quit"},
	}
}
