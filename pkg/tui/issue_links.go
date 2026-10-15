package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/nikbrunner/lazyjira/pkg/jira"
	"github.com/nikbrunner/lazyjira/pkg/tui/components"
	"github.com/nikbrunner/lazyjira/pkg/tui/views"
)

// linkDirection is one way to read a link type from the issue being linked,
// such as "blocks" or "is blocked by".
type linkDirection struct {
	typeName string
	phrase   string
	outward  bool
}

// pendingLink is a link waiting for its target in the issue lookup.
type pendingLink struct {
	issueKey string
	dir      linkDirection
}

type linkTypesLoadedMsg struct {
	issueKey     string
	types        []jira.IssueLinkType
	cacheVersion uint64
}

// linkEnds maps "from <phrase> to" onto Jira's link ends. The outward issue
// is the one the outward phrase describes: "A blocks B" makes A outward.
func linkEnds(from, to string, outward bool) (inwardKey, outwardKey string) {
	if outward {
		return to, from
	}
	return from, to
}

// linkDirections lists each type's outward then inward phrase, once when both
// read the same.
func linkDirections(types []jira.IssueLinkType) []linkDirection {
	dirs := make([]linkDirection, 0, 2*len(types))
	for _, t := range types {
		dirs = append(dirs, linkDirection{typeName: t.Name, phrase: t.Outward, outward: true})
		if t.Inward != t.Outward {
			dirs = append(dirs, linkDirection{typeName: t.Name, phrase: t.Inward})
		}
	}
	return dirs
}

func (a *App) onLinksTab() bool {
	return a.side == sideLeft && a.leftFocus == focusInfo &&
		a.infoPanel.ActiveTab() == views.InfoTabLinks && a.infoPanel.IssueKey() != ""
}

func (a *App) startAddLink() tea.Cmd {
	issueKey := a.infoPanel.IssueKey()
	if types, ok := a.linkTypesCache.get(""); ok {
		a.showLinkTypePicker(issueKey, types)
		return nil
	}
	return fetchIssueLinkTypes(a.client, issueKey, a.referenceCacheVersion)
}

func (a *App) handleLinkTypesLoaded(msg linkTypesLoadedMsg) {
	if msg.cacheVersion == a.referenceCacheVersion {
		a.linkTypesCache.set("", msg.types)
	}
	if msg.issueKey == a.infoPanel.IssueKey() && a.onLinksTab() && !a.modal.IsVisible() {
		a.showLinkTypePicker(msg.issueKey, msg.types)
	}
}

func (a *App) showLinkTypePicker(issueKey string, types []jira.IssueLinkType) {
	dirs := linkDirections(types)
	if len(dirs) == 0 {
		a.statusPanel.SetError("Jira has no issue link types")
		return
	}
	items := make([]components.ModalItem, len(dirs))
	for i, d := range dirs {
		items[i] = components.ModalItem{ID: strconv.Itoa(i), Label: d.phrase}
	}
	a.onSelect = func(item components.ModalItem) tea.Cmd {
		i, err := strconv.Atoi(item.ID)
		if err != nil || i < 0 || i >= len(dirs) {
			return nil
		}
		a.parentLookupFor = ""
		a.linkLookup = &pendingLink{issueKey: issueKey, dir: dirs[i]}
		return a.showIssueLookup(issueKey+" "+dirs[i].phrase, false)
	}
	a.modal.Show("Link "+issueKey, items)
}

func (a *App) createLink(p pendingLink, target string) tea.Cmd {
	target = strings.ToUpper(strings.TrimSpace(target))
	if !parentKeyRegex.MatchString(target) {
		return func() tea.Msg {
			return errorMsg{err: fmt.Errorf("invalid issue key %q (expected PROJ-123)", target)}
		}
	}
	if target == p.issueKey {
		return func() tea.Msg { return errorMsg{err: fmt.Errorf("can't link %s to itself", target)} }
	}
	inward, outward := linkEnds(p.issueKey, target, p.dir.outward)
	delete(a.issueCache, target)
	*a.logFlag = true
	return createIssueLink(a.client, p.issueKey, p.dir.typeName, inward, outward)
}

func (a *App) confirmRemoveLink() {
	row, ok := a.infoPanel.SelectedLinkRow()
	if !ok || row.LinkID == "" {
		return
	}
	issueKey := a.infoPanel.IssueKey()
	a.confirm(fmt.Sprintf("Remove link: %s %s?", row.Phrase, row.Issue.Key), func() tea.Cmd {
		delete(a.issueCache, row.Issue.Key)
		*a.logFlag = true
		return deleteIssueLink(a.client, issueKey, row.LinkID)
	})
}

func fetchIssueLinkTypes(client jira.ClientInterface, issueKey string, cacheVersion uint64) tea.Cmd {
	return func() tea.Msg {
		types, err := client.GetIssueLinkTypes(context.Background())
		if err != nil {
			return errorMsg{err: err}
		}
		return linkTypesLoadedMsg{issueKey: issueKey, types: types, cacheVersion: cacheVersion}
	}
}

func createIssueLink(client jira.ClientInterface, issueKey, typeName, inwardKey, outwardKey string) tea.Cmd {
	return func() tea.Msg {
		if err := client.CreateIssueLink(context.Background(), typeName, inwardKey, outwardKey); err != nil {
			return errorMsg{err: err}
		}
		return issueUpdatedMsg{issueKey: issueKey, field: "issuelinks"}
	}
}

func deleteIssueLink(client jira.ClientInterface, issueKey, linkID string) tea.Cmd {
	return func() tea.Msg {
		if err := client.DeleteIssueLink(context.Background(), linkID); err != nil {
			return errorMsg{err: err}
		}
		return issueUpdatedMsg{issueKey: issueKey, field: "issuelinks"}
	}
}
