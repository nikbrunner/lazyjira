package views

import (
	"testing"

	"github.com/nikbrunner/lazyjira/pkg/internal/testkit"
	"github.com/nikbrunner/lazyjira/pkg/jira"
)

func TestInfoPanel_SelectedLinkRow(t *testing.T) {
	t.Parallel()
	blocks := &jira.IssueLinkType{Name: "Blocks", Inward: "is blocked by", Outward: "blocks"}
	p := makeInfoPanelFocused()
	p.SetIssue(&jira.Issue{Key: "MAIN-1", IssueLinks: []jira.IssueLink{
		{ID: "10", Type: blocks, OutwardIssue: &jira.Issue{Key: "SHOP-12"}},
		{ID: "11", Type: blocks, InwardIssue: &jira.Issue{Key: "SHOP-3"}},
	}})

	if _, ok := p.SelectedLinkRow(); ok {
		t.Error("Info tab: want no link row")
	}

	p.SetActiveTab(InfoTabLinks)
	p.syncItemCount()
	row, ok := p.SelectedLinkRow()
	testkit.AssertEqual(t, "ok", ok, true)
	testkit.AssertEqual(t, "first id", row.LinkID, "10")
	testkit.AssertEqual(t, "first phrase", row.Phrase, "blocks")
	testkit.AssertEqual(t, "first key", row.Issue.Key, "SHOP-12")

	p.Update(pressJ())
	row, _ = p.SelectedLinkRow()
	testkit.AssertEqual(t, "second id", row.LinkID, "11")
	testkit.AssertEqual(t, "second phrase", row.Phrase, "is blocked by")
	testkit.AssertEqual(t, "second key", row.Issue.Key, "SHOP-3")
	testkit.AssertEqual(t, "SelectedLinkKey", p.SelectedLinkKey(), "SHOP-3")
}
