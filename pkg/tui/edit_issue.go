package tui

import (
	"context"
	"errors"
	"reflect"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nikbrunner/lazyjira/pkg/jira"
	"github.com/nikbrunner/lazyjira/pkg/tui/components"
)

var errMissingIssueType = errors.New("issue has no issue type")

// editIssueLoadedMsg carries the fresh issue the Issue Edit View prefills from.
type editIssueLoadedMsg struct {
	key   string
	issue *jira.Issue
	err   error
}

// issueSavedMsg reports a successful update from the Issue Edit View.
type issueSavedMsg struct{ key string }

func fetchIssueForEdit(client jira.ClientInterface, key string) tea.Cmd {
	return func() tea.Msg {
		issue, err := client.GetIssue(context.Background(), key)
		return editIssueLoadedMsg{key: key, issue: issue, err: err}
	}
}

func saveIssue(client jira.ClientInterface, key string, fields map[string]any) tea.Cmd {
	return func() tea.Msg {
		if len(fields) > 0 {
			if err := client.UpdateIssue(context.Background(), key, fields); err != nil {
				return createErrorMsg{err: err}
			}
		}
		return issueSavedMsg{key: key}
	}
}

// startEditIssue opens the Issue Edit View for key, loading the issue first so
// the prefill carries its full description and fields.
func (a *App) startEditIssue(key string) (tea.Model, tea.Cmd) {
	a.createCtx = createCtx{projectKey: projectKeyFromIssueKey(key), editKey: key}
	a.createForm.SetTitle("Edit " + key)
	a.createForm.SetLoading(true)
	*a.logFlag = true
	return a, fetchIssueForEdit(a.client, key)
}

func (a *App) handleEditIssueLoaded(msg editIssueLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.key != a.createCtx.editKey {
		return a, nil
	}
	if msg.err == nil && (msg.issue == nil || msg.issue.IssueType == nil) {
		msg.err = errMissingIssueType
	}
	if msg.err != nil {
		text := "Cannot edit " + msg.key + ": " + msg.err.Error()
		a.createForm.Hide()
		a.createCtx = createCtx{}
		a.statusPanel.SetError(text)
		a.modal.ShowError("Error", []components.ModalItem{{Label: text}})
		return a, nil
	}
	a.createCtx.editFrom = msg.issue
	t := msg.issue.IssueType
	return a.handleCreateFormTypeSelected(components.CreateFormTypeSelectedMsg{TypeID: t.ID, TypeName: t.Name})
}

// applyEditPrefill fills the form with the issue's current values. Unlike a
// duplicate, the summary stays as is and the description keeps its media.
func applyEditPrefill(fields []components.CreateFormField, src *jira.Issue, isCloud bool) {
	applyDuplicatePrefill(fields, src, isCloud)
	for i := range fields {
		switch fields[i].FieldID {
		case fldSummary:
			fields[i].DisplayValue = src.Summary
			fields[i].Value = src.Summary
		case fldDescription:
			if src.DescriptionADF != nil {
				fields[i].Value = src.DescriptionADF
			}
		}
	}
}

// changedFields returns the values that differ from the view's initial ones.
// A value the user cleared is sent as nil.
func changedFields(initial, current map[string]any) map[string]any {
	changed := make(map[string]any)
	for id, v := range current {
		if !reflect.DeepEqual(initial[id], v) {
			changed[id] = v
		}
	}
	for id := range initial {
		if _, ok := current[id]; !ok {
			changed[id] = nil
		}
	}
	return changed
}

// handleEditSave sends only the changed fields. Saving with nothing changed
// and no images closes the view without a request.
func (a *App) handleEditSave(msg components.CreateFormSubmitMsg) (tea.Model, tea.Cmd) {
	fields := changedFields(a.createCtx.editInitial, msg.Fields)
	if len(fields) == 0 && len(msg.Attachments) == 0 {
		a.createForm.Hide()
		a.createCtx = createCtx{}
		return a, nil
	}
	a.createCtx.attachments = msg.Attachments
	a.createForm.SetLoading(true)
	*a.logFlag = true
	return a, a.submitIssueForm(fields)
}

func (a *App) handleIssueSaved(msg issueSavedMsg) (tea.Model, tea.Cmd) {
	attachments := a.createCtx.attachments
	a.createForm.Hide()
	a.createCtx = createCtx{}
	a.helpBar.SetStatusMsg("Updated " + msg.key)
	if len(attachments) > 0 {
		return a, uploadAttachments(a.client, msg.key, "Updated", attachments)
	}
	return a, tea.Batch(a.fetchActiveTab(), fetchIssueDetail(a.client, msg.key))
}
