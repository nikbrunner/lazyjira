package jira

import (
	"net/http"
	"testing"

	"github.com/nikbrunner/lazyjira/pkg/internal/testkit"
)

func TestClient_SuggestIssues(t *testing.T) {
	t.Parallel()

	body := `{"sections":[
		{"issues":[{"key":"PLAT-12","summaryText":"Rate limiter"},{"key":"PLAT-120","summaryText":"Tracing"}]},
		{"issues":[{"key":"PLAT-12","summaryText":"Rate limiter"},{"key":"PLAT-1","summaryText":"OAuth"}]}
	]}`

	for _, tt := range []struct {
		name string
		opts ClientOpts
		path string
	}{
		{"cloud", cloudOpts(), "/rest/api/3/issue/picker"},
		{"server", serverOpts(), "/rest/api/2/issue/picker"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client, recorded := newRecordingClient(t, tt.opts, testkit.StubResponse{Status: http.StatusOK, Body: body})

			got, err := client.SuggestIssues(t.Context(), "PLAT-1")
			if err != nil {
				t.Fatalf("SuggestIssues: %v", err)
			}

			testkit.AssertEqual(t, "path", recorded.Path, tt.path)
			testkit.AssertEqual(t, "query", recorded.Query.Get("query"), "PLAT-1")
			want := []IssueSuggestion{{"PLAT-12", "Rate limiter"}, {"PLAT-120", "Tracing"}, {"PLAT-1", "OAuth"}}
			if len(got) != len(want) {
				t.Fatalf("suggestions = %v, want %v", got, want)
			}
			for i := range want {
				testkit.AssertEqual(t, "suggestion", got[i], want[i])
			}
		})
	}
}
