package tui

import (
	"slices"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/tui/components"
)

// Action represents a user-triggerable action
type Action string

// Actions each can be remapped to different keys via config
const (
	ActQuit             Action = "quit"
	ActHelp             Action = "help"
	ActSearch           Action = "search"
	ActSwitchPanel      Action = "switchPanel"
	ActFocusRight       Action = "focusRight"
	ActFocusLeft        Action = "focusLeft"
	ActSelect           Action = "select" // primary: mark active + open
	ActOpen             Action = "open"   // secondary: open/preview without marking
	ActPrevTab          Action = "prevTab"
	ActNextTab          Action = "nextTab"
	ActFocusDetail      Action = "focusDetail"
	ActFocusStatus      Action = "focusStatus"
	ActFocusIssues      Action = "focusIssues"
	ActFocusInfo        Action = "focusInfo"
	ActFocusProj        Action = "focusProjects"
	ActToggleMaximize   Action = "toggleMaximize"
	ActCopyURL          Action = "copyURL"
	ActBrowser          Action = "browser"
	ActURLPicker        Action = "urlPicker"
	ActTransition       Action = "transition"
	ActRefresh          Action = "refresh"
	ActRefreshAll       Action = "refreshAll"
	ActInfoTab          Action = "infoTab" // legacy: now focuses Info panel
	ActEdit             Action = "edit"
	ActComments         Action = "comments"
	ActNew              Action = "new"
	ActPriority         Action = "editPriority"
	ActAssignee         Action = "editAssignee"
	ActJQLSearch        Action = "jqlSearch"
	ActCloseJQLTab      Action = "closeJQLTab"
	ActCreateBranch     Action = "createBranch"
	ActCopyBranchName   Action = "copyBranchName"
	ActCopyWorktreeName Action = "copyWorktreeName"
	ActCreateWorktree   Action = "createWorktree"
	ActCreateIssue      Action = "createIssue"
	ActCreateSubtask    Action = "createSubtask"
	ActDuplicateIssue   Action = "duplicateIssue"
	ActShowParent       Action = "showParent"

	ActNavDown     Action = "navDown"
	ActNavUp       Action = "navUp"
	ActNavTop      Action = "navTop"
	ActNavBottom   Action = "navBottom"
	ActNavHalfDown Action = "navHalfPageDown"
	ActNavHalfUp   Action = "navHalfPageUp"

	ActDetailScrollDown Action = "detailScrollDown"
	ActDetailScrollUp   Action = "detailScrollUp"
	ActDetailHalfDown   Action = "detailHalfPageDown"
	ActDetailHalfUp     Action = "detailHalfPageUp"
)

// Keymap maps actions to key strings. Multiple keys can trigger the same action
type Keymap map[Action][]string

// DefaultKeymap returns the default key bindings
func DefaultKeymap() Keymap {
	return Keymap{
		ActQuit:             {"q", "ctrl+c"},
		ActHelp:             {"?"},
		ActSearch:           {"/"},
		ActSwitchPanel:      {},
		ActFocusRight:       {},
		ActFocusLeft:        {"esc"},
		ActSelect:           {" "},
		ActOpen:             {"enter"},
		ActPrevTab:          {"["},
		ActNextTab:          {"]"},
		ActFocusDetail:      {"2"},
		ActFocusStatus:      {"0"},
		ActFocusIssues:      {"1"},
		ActFocusInfo:        {"3"},
		ActFocusProj:        {"4"},
		ActToggleMaximize:   {"+"},
		ActCopyURL:          {"y"},
		ActBrowser:          {"o"},
		ActURLPicker:        {"u"},
		ActTransition:       {"t"},
		ActRefresh:          {"r"},
		ActRefreshAll:       {"R"},
		ActInfoTab:          {"i"},
		ActEdit:             {"e"},
		ActComments:         {"c"},
		ActNew:              {"n"},
		ActPriority:         {"p"},
		ActAssignee:         {"a"},
		ActJQLSearch:        {"s"},
		ActCloseJQLTab:      {"x"},
		ActCreateBranch:     {"B"},
		ActCopyBranchName:   {"b"},
		ActCopyWorktreeName: {"w"},
		ActCreateWorktree:   {"W"},
		ActDuplicateIssue:   {"ctrl+n"},
		ActCreateSubtask:    {"S"},
		ActShowParent:       {"backspace"},

		ActNavDown:     {"j", "down", "ctrl+j"},
		ActNavUp:       {"k", "up", "ctrl+k"},
		ActNavTop:      {"g", "home"},
		ActNavBottom:   {"G", "end"},
		ActNavHalfDown: {"ctrl+d"},
		ActNavHalfUp:   {"ctrl+u"},

		ActDetailScrollDown: {"ctrl+d"},
		ActDetailScrollUp:   {"ctrl+u"},
		ActDetailHalfDown:   {"ctrl+f"},
		ActDetailHalfUp:     {"ctrl+b"},
	}
}

// KeymapFromConfig builds a Keymap starting from defaults and overriding
// with any non-empty values from the user's keybinding config.
func KeymapFromConfig(kcfg config.KeybindingConfig) Keymap {
	km := DefaultKeymap()
	configuredActions := make(map[Action]bool)
	configuredKeys := make(map[string]bool)
	set := func(action Action, val string) {
		if val != "" {
			km[action] = []string{val}
			configuredActions[action] = true
			configuredKeys[val] = true
		}
	}
	// Universal
	set(ActQuit, kcfg.Universal.Quit)
	set(ActHelp, kcfg.Universal.Help)
	set(ActSearch, kcfg.Universal.Search)
	set(ActSwitchPanel, kcfg.Universal.SwitchPanel)
	set(ActRefresh, kcfg.Universal.Refresh)
	set(ActRefreshAll, kcfg.Universal.RefreshAll)
	set(ActPrevTab, kcfg.Universal.PrevTab)
	set(ActNextTab, kcfg.Universal.NextTab)
	set(ActFocusDetail, kcfg.Universal.FocusDetail)
	set(ActFocusStatus, kcfg.Universal.FocusStatus)
	set(ActFocusIssues, kcfg.Universal.FocusIssues)
	set(ActFocusInfo, kcfg.Universal.FocusInfo)
	set(ActFocusProj, kcfg.Universal.FocusProj)
	set(ActToggleMaximize, kcfg.Universal.ToggleMaximize)
	set(ActJQLSearch, kcfg.Universal.JQLSearch)
	// Issues (Select, Open, FocusRight are shared with Projects panel)
	set(ActSelect, kcfg.Issues.Select)
	set(ActOpen, kcfg.Issues.Open)
	set(ActFocusRight, kcfg.Issues.FocusRight)
	set(ActTransition, kcfg.Issues.Transition)
	set(ActBrowser, kcfg.Issues.Browser)
	set(ActURLPicker, kcfg.Issues.URLPicker)
	set(ActCopyURL, kcfg.Issues.CopyURL)
	set(ActCloseJQLTab, kcfg.Issues.CloseJQLTab)
	set(ActCreateBranch, kcfg.Issues.CreateBranch)
	set(ActCopyBranchName, kcfg.Issues.CopyBranchName)
	set(ActCopyWorktreeName, kcfg.Issues.CopyWorktreeName)
	set(ActCreateWorktree, kcfg.Issues.CreateWorktree)
	set(ActCreateIssue, kcfg.Issues.CreateIssue)
	set(ActCreateSubtask, kcfg.Issues.CreateSubtask)
	// Detail
	set(ActFocusLeft, kcfg.Detail.FocusLeft)
	set(ActInfoTab, kcfg.Detail.InfoTab)
	set(ActDetailScrollDown, kcfg.Detail.ScrollDown)
	set(ActDetailScrollUp, kcfg.Detail.ScrollUp)
	set(ActDetailHalfDown, kcfg.Detail.HalfPageDown)
	set(ActDetailHalfUp, kcfg.Detail.HalfPageUp)
	// Navigation
	set(ActNavDown, kcfg.Navigation.Down)
	set(ActNavUp, kcfg.Navigation.Up)
	set(ActNavTop, kcfg.Navigation.Top)
	set(ActNavBottom, kcfg.Navigation.Bottom)
	set(ActNavHalfDown, kcfg.Navigation.HalfDown)
	set(ActNavHalfUp, kcfg.Navigation.HalfUp)
	for action, keys := range km {
		if !configuredActions[action] {
			km[action] = slices.DeleteFunc(keys, func(key string) bool { return configuredKeys[key] })
		}
	}
	return km
}

// Match returns the action for the given key, or "" if none
func (km Keymap) Match(key string) Action {
	for action, keys := range km {
		if slices.Contains(keys, key) {
			return action
		}
	}
	return ""
}

var navActionMap = map[Action]components.NavAction{
	ActNavDown:     components.NavDown,
	ActNavUp:       components.NavUp,
	ActNavTop:      components.NavTop,
	ActNavBottom:   components.NavBottom,
	ActNavHalfDown: components.NavHalfDown,
	ActNavHalfUp:   components.NavHalfUp,
}

func (km Keymap) MatchNav(key string) components.NavAction {
	for action, nav := range navActionMap {
		if slices.Contains(km[action], key) {
			return nav
		}
	}
	return components.NavNone
}

// Keys returns the first key bound to the action (for display in help)
func (km Keymap) Keys(action Action) string {
	if keys, ok := km[action]; ok && len(keys) > 0 {
		k := keys[0]
		if k == " " {
			return "space"
		}
		return k
	}
	return ""
}
