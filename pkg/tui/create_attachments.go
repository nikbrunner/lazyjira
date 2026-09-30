package tui

import (
	"context"
	"fmt"
	"os"
	"regexp"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nikbrunner/lazyjira/pkg/jira"
	"github.com/nikbrunner/lazyjira/pkg/tui/components"
)

// attachmentsUploadedMsg reports the uploads for a newly created issue.
type attachmentsUploadedMsg struct {
	issueKey string
	total    int
	errs     []error
}

var imageTokenRe = regexp.MustCompile(`\[(Image #\d+)\]`)

// escapeImageTokensForWiki keeps [Image #N] literal in Jira wiki markup, where
// square brackets make a link.
func escapeImageTokensForWiki(text string) string {
	return imageTokenRe.ReplaceAllString(text, `\[$1\]`)
}

// uploadAttachments uploads the images one after another and deletes the
// clipboard temp files afterwards.
func uploadAttachments(client jira.ClientInterface, issueKey string, attachments []components.CreateAttachment) tea.Cmd {
	return func() tea.Msg {
		defer components.RemoveTempAttachments(attachments)
		msg := attachmentsUploadedMsg{issueKey: issueKey, total: len(attachments)}
		for _, att := range attachments {
			data, err := os.ReadFile(att.Path)
			if err == nil {
				err = client.AddAttachment(context.Background(), issueKey, att.Name, data)
			}
			if err != nil {
				msg.errs = append(msg.errs, err)
			}
		}
		return msg
	}
}

// handleAttachmentsUploaded reports failures against the created issue, which
// already exists, so the form never reopens. The error dialog outlives the
// refresh, which clears the status error.
func (a *App) handleAttachmentsUploaded(msg attachmentsUploadedMsg) (tea.Model, tea.Cmd) {
	if len(msg.errs) > 0 {
		text := fmt.Sprintf("Created %s; %d of %d images failed to upload: %v",
			msg.issueKey, len(msg.errs), msg.total, msg.errs[0])
		a.statusPanel.SetError(text)
		a.modal.ShowError("Upload failed", []components.ModalItem{{Label: text}})
	} else {
		noun := "images"
		if msg.total == 1 {
			noun = "image"
		}
		a.helpBar.SetStatusMsg(fmt.Sprintf("Created %s with %d %s", msg.issueKey, msg.total, noun))
	}
	return a, tea.Batch(a.fetchActiveTab(), fetchIssueDetail(a.client, msg.issueKey))
}
