package tui

import (
	"context"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nikbrunner/lazyjira/pkg/jira"
	"github.com/nikbrunner/lazyjira/pkg/tui/components"
)

// lookupReturn is the focus and maximize state restored when an issue opened
// by key is closed.
type lookupReturn struct {
	side          focusSide
	leftFocus     focusPanel
	maximized     bool
	maximizedPane focusPanel
}

type issueSuggestionsLoadedMsg struct {
	query       string
	suggestions []jira.IssueSuggestion
}

func (a *App) openIssueLookup() tea.Cmd {
	if a.searchBar.IsActive() {
		updated, cancelCmd := a.searchBar.Update(tea.KeyMsg{Type: tea.KeyEsc})
		a.searchBar = updated
		if cancelCmd != nil {
			a.handleSearchCancelled()
		}
	}
	choices := make([]components.ProjectChoice, 0, len(a.projectList.AllProjects()))
	for _, project := range a.projectList.AllProjects() {
		choices = append(choices, components.ProjectChoice{Key: project.Key, Name: project.Name})
	}
	prefill := ""
	if a.projectKey != "" {
		prefill = a.projectKey + "-"
	}
	cmd := a.issueLookup.Show(prefill, choices)
	a.overlays.SetSize(a.width, a.height)
	return cmd
}

// handleIssueLookupQuery shows matching issues lazyjira already has, then asks
// Jira for more.
func (a *App) handleIssueLookupQuery(query string) tea.Cmd {
	a.issueLookup.SetIssueSuggestions(query, a.loadedLookupSuggestions(query, nil))
	client := a.client
	return func() tea.Msg {
		suggestions, err := client.SuggestIssues(context.Background(), query)
		if err != nil {
			return nil
		}
		return issueSuggestionsLoadedMsg{query: query, suggestions: suggestions}
	}
}

func (a *App) handleIssueSuggestionsLoaded(msg issueSuggestionsLoadedMsg) {
	items := make([]components.LookupSuggestion, 0, len(msg.suggestions))
	for _, s := range msg.suggestions {
		items = append(items, components.LookupSuggestion{Key: s.Key, Label: s.Summary})
	}
	a.issueLookup.SetIssueSuggestions(msg.query, a.loadedLookupSuggestions(msg.query, items))
}

// loadedLookupSuggestions appends cached issues whose key starts with query to
// items, skipping keys already present.
func (a *App) loadedLookupSuggestions(query string, items []components.LookupSuggestion) []components.LookupSuggestion {
	seen := map[string]bool{}
	for _, item := range items {
		seen[item.Key] = true
	}
	prefix := strings.ToUpper(query)
	var loaded []components.LookupSuggestion
	for key, issue := range a.issueCache {
		if issue == nil || seen[key] || !strings.HasPrefix(strings.ToUpper(key), prefix) {
			continue
		}
		loaded = append(loaded, components.LookupSuggestion{Key: key, Label: issue.Summary})
	}
	sort.Slice(loaded, func(i, j int) bool {
		if len(loaded[i].Key) != len(loaded[j].Key) {
			return len(loaded[i].Key) < len(loaded[j].Key)
		}
		return loaded[i].Key < loaded[j].Key
	})
	return append(items, loaded...)
}

// openLookupIssue shows key in the maximized Issue details pane.
func (a *App) openLookupIssue(key string) tea.Cmd {
	if a.lookupReturn == nil {
		a.lookupReturn = &lookupReturn{
			side:          a.side,
			leftFocus:     a.leftFocus,
			maximized:     a.maximized,
			maximizedPane: a.maximizedPane,
		}
	}
	a.previewEpoch++
	a.previewKey = key
	if cached, ok := a.issueCache[key]; ok && cached != nil {
		a.detailView.SetIssue(cached)
	} else {
		a.detailView.SetIssue(&jira.Issue{Key: key})
	}
	a.side = sideRight
	a.maximized = true
	a.maximizedPane = focusDetailPane
	a.updateFocusHints()
	a.updateFocusState()
	return fetchIssueDetail(a.client, key)
}

// closeIssueLookup restores the layout from before the lookup and shows the
// selected issue again.
func (a *App) closeIssueLookup() tea.Cmd {
	ret := a.lookupReturn
	a.lookupReturn = nil
	a.side = ret.side
	a.leftFocus = ret.leftFocus
	a.maximized = ret.maximized
	a.maximizedPane = ret.maximizedPane
	a.updateFocusHints()
	a.updateFocusState()
	return a.restoreSelectionPreview()
}

// endIssueLookupIfLeft drops the lookup once focus leaves the maximized Issue
// details pane by other means than esc.
func (a *App) endIssueLookupIfLeft() {
	if a.lookupReturn == nil || (a.side == sideRight && a.maximized && a.maximizedPane == focusDetailPane) {
		return
	}
	a.lookupReturn = nil
	a.restoreSelectionPreview()
}

func (a *App) restoreSelectionPreview() tea.Cmd {
	if a.issuesList.SelectedIssue() == nil {
		a.clearIssuePreview()
		return nil
	}
	return a.previewSelectedIssue()
}
