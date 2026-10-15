//go:build demo

package jira

import (
	"testing"
)

func linkTo(t *testing.T, iss *Issue, key string) *IssueLink {
	t.Helper()
	for i, link := range iss.IssueLinks {
		if (link.OutwardIssue != nil && link.OutwardIssue.Key == key) || (link.InwardIssue != nil && link.InwardIssue.Key == key) {
			return &iss.IssueLinks[i]
		}
	}
	return nil
}

func TestDemoServer_CreatesAndDeletesIssueLinks(t *testing.T) {
	t.Parallel()
	srv, err := NewDemoServer()
	if err != nil {
		t.Fatalf("NewDemoServer: %v", err)
	}
	t.Cleanup(func() { _ = srv.Close() })
	client := NewClientWithOpts(ClientOpts{Host: srv.URL, Email: "demo@lazyjira.dev", Token: "demo", IsCloud: true})

	types, err := client.GetIssueLinkTypes(t.Context())
	if err != nil || len(types) == 0 {
		t.Fatalf("GetIssueLinkTypes = %v, %v", types, err)
	}

	if err := client.CreateIssueLink(t.Context(), "Blocks", "SHOP-7", "SHOP-3"); err != nil {
		t.Fatalf("CreateIssueLink: %v", err)
	}
	from, _ := client.GetIssue(t.Context(), "SHOP-3")
	to, _ := client.GetIssue(t.Context(), "SHOP-7")
	fromLink, toLink := linkTo(t, from, "SHOP-7"), linkTo(t, to, "SHOP-3")
	if fromLink == nil || fromLink.OutwardIssue == nil || fromLink.Type.Outward != "blocks" {
		t.Fatalf("SHOP-3 link = %+v, want outward SHOP-7 (blocks)", fromLink)
	}
	if toLink == nil || toLink.InwardIssue == nil || toLink.ID != fromLink.ID {
		t.Fatalf("SHOP-7 link = %+v, want inward SHOP-3 with ID %s", toLink, fromLink.ID)
	}

	if err := client.DeleteIssueLink(t.Context(), fromLink.ID); err != nil {
		t.Fatalf("DeleteIssueLink: %v", err)
	}
	from, _ = client.GetIssue(t.Context(), "SHOP-3")
	to, _ = client.GetIssue(t.Context(), "SHOP-7")
	if linkTo(t, from, "SHOP-7") != nil || linkTo(t, to, "SHOP-3") != nil {
		t.Error("link still listed after delete")
	}
	if err := client.DeleteIssueLink(t.Context(), fromLink.ID); err == nil {
		t.Error("deleting a missing link: want error")
	}
}
