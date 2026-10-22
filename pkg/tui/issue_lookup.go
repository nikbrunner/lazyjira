package tui

import (
	"context"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/nikbrunner/lazyjira/pkg/jira"
	"github.com/nikbrunner/lazyjira/pkg/tui/components"
)

type issueSuggestionsLoadedMsg struct {
	query       string
	suggestions []jira.IssueSuggestion
}

func (a *App) openIssueLookup() tea.Cmd {
	a.parentLookupFor = ""
	a.linkLookup = nil
	return a.showIssueLookup("Open issue", false)
}

func (a *App) openParentLookup(issueKey string) tea.Cmd {
	a.parentLookupFor = issueKey
	a.linkLookup = nil
	return a.showIssueLookup("Edit Parent", true)
}

func (a *App) showIssueLookup(title string, allowEmpty bool) tea.Cmd {
	if a.searchBar.IsActive() {
		updated, cancelCmd := a.searchBar.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
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
	cmd := a.issueLookup.ShowTitled(title, allowEmpty, prefill, choices)
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

func (a *App) handleIssueLookupSelected(key string) tea.Cmd {
	if pending := a.linkLookup; pending != nil {
		a.linkLookup = nil
		return a.createLink(*pending, key)
	}
	if issueKey := a.parentLookupFor; issueKey != "" {
		a.parentLookupFor = ""
		return a.applyParentEdit(issueKey, key)
	}
	_, cmd := a.startEditIssue(key)
	return cmd
}
