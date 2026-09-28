package tui

import (
	"errors"
	"testing"

	"github.com/nikbrunner/lazyjira/v2/pkg/jira"
)

func TestFormatJQLError_UsesJiraMessages(t *testing.T) {
	t.Parallel()
	apiErr := &jira.APIError{
		Method:   "GET",
		Path:     "/search/jql?jql=very+long",
		Status:   400,
		Body:     `{"errorMessages":["Error in the JQL Query: bad token."]}`,
		Messages: []string{"Error in the JQL Query: bad token.", "Field 'x' does not exist."},
	}
	if got, want := formatJQLError(apiErr), "Error in the JQL Query: bad token.\nField 'x' does not exist."; got != want {
		t.Errorf("formatJQLError = %q, want %q", got, want)
	}
	if got := formatJQLError(errors.New("dial tcp: timeout")); got != "dial tcp: timeout" {
		t.Errorf("non-API error = %q, want raw message", got)
	}
}
