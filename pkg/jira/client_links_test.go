package jira

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/nikbrunner/lazyjira/pkg/internal/testkit"
)

func TestClient_GetIssueLinkTypes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		opts     ClientOpts
		wantPath string
	}{
		{"cloud", cloudOpts(), "/rest/api/3/issueLinkType"},
		{"server", serverOpts(), "/rest/api/2/issueLinkType"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client, recorded := newRecordingClient(t, tc.opts, testkit.StubResponse{Status: http.StatusOK, Body: `{"issueLinkTypes":[
				{"id":"1000","name":"Duplicate","inward":"is duplicated by","outward":"duplicates"},
				{"id":"1010","name":"Blocks","inward":"is blocked by","outward":"blocks"}]}`})

			types, err := client.GetIssueLinkTypes(t.Context())
			if err != nil {
				t.Fatalf("GetIssueLinkTypes: %v", err)
			}

			testkit.AssertEqual(t, "method", recorded.Method, http.MethodGet)
			testkit.AssertEqual(t, "path", recorded.Path, tc.wantPath)
			if len(types) != 2 {
				t.Fatalf("len(types) = %d, want 2", len(types))
			}
			testkit.AssertEqual(t, "types[1]", types[1], IssueLinkType{Name: "Blocks", Inward: "is blocked by", Outward: "blocks"})
		})
	}
}

func TestClient_CreateIssueLink(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		opts     ClientOpts
		wantPath string
	}{
		{"cloud", cloudOpts(), "/rest/api/3/issueLink"},
		{"server", serverOpts(), "/rest/api/2/issueLink"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client, recorded := newRecordingClient(t, tc.opts, testkit.StubResponse{Status: http.StatusCreated})

			if err := client.CreateIssueLink(t.Context(), "Blocks", "SHOP-12", "SHOP-2"); err != nil {
				t.Fatalf("CreateIssueLink: %v", err)
			}

			testkit.AssertEqual(t, "method", recorded.Method, http.MethodPost)
			testkit.AssertEqual(t, "path", recorded.Path, tc.wantPath)
			var body map[string]map[string]string
			if err := json.Unmarshal(recorded.Body, &body); err != nil {
				t.Fatalf("decode body %s: %v", recorded.Body, err)
			}
			testkit.AssertEqual(t, "type.name", body["type"]["name"], "Blocks")
			testkit.AssertEqual(t, "inwardIssue.key", body["inwardIssue"]["key"], "SHOP-12")
			testkit.AssertEqual(t, "outwardIssue.key", body["outwardIssue"]["key"], "SHOP-2")
			testkit.AssertEqual(t, "body keys", len(body), 3)
		})
	}
}

func TestClient_DeleteIssueLink(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		opts     ClientOpts
		wantPath string
	}{
		{"cloud", cloudOpts(), "/rest/api/3/issueLink/10042"},
		{"server", serverOpts(), "/rest/api/2/issueLink/10042"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client, recorded := newRecordingClient(t, tc.opts, testkit.StubResponse{Status: http.StatusNoContent})

			if err := client.DeleteIssueLink(t.Context(), "10042"); err != nil {
				t.Fatalf("DeleteIssueLink: %v", err)
			}

			testkit.AssertEqual(t, "method", recorded.Method, http.MethodDelete)
			testkit.AssertEqual(t, "path", recorded.Path, tc.wantPath)
		})
	}
}

func TestClient_IssueLinkErrors(t *testing.T) {
	t.Parallel()

	client, _ := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusNotFound, Body: `{"errorMessages":["Issue Does Not Exist"]}`})

	if _, err := client.GetIssueLinkTypes(t.Context()); err == nil {
		t.Error("GetIssueLinkTypes: want error")
	}
	if err := client.CreateIssueLink(t.Context(), "Blocks", "SHOP-1", "SHOP-2"); err == nil {
		t.Error("CreateIssueLink: want error")
	}
	if err := client.DeleteIssueLink(t.Context(), "1"); err == nil {
		t.Error("DeleteIssueLink: want error")
	}
}
