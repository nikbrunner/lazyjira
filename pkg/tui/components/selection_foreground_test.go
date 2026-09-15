package components

import (
	"strconv"
	"strings"
	"testing"
)

func TestInputModal_SelectedHintUsesTerminalForeground(t *testing.T) {
	t.Parallel()
	m := NewInputModal()
	m.SetSize(80, 24)
	m.Show(testTitle, "")
	m.SetHints([]string{testBranchName1})
	m.focusInput = false
	assertTerminalForeground(t, m.HintView(), testBranchName1)
}

func TestCreateForm_SelectedFieldUsesTerminalForeground(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.focusedPanel = CreatePanelFields
	assertTerminalForeground(t, form.renderFields(80, 8), testFieldName)
}

func assertTerminalForeground(t *testing.T, output, text string) {
	t.Helper()
	idx := strings.Index(output, text)
	if idx < 0 {
		t.Fatalf("%q missing from %q", text, output)
	}
	foreground := 0
	for _, seq := range ansiEscapeRe.FindAllString(output[:idx], -1) {
		for _, value := range strings.Split(seq[2:len(seq)-1], ";") {
			code, _ := strconv.Atoi(value)
			if code == 0 || code == 39 {
				foreground = 0
			} else if code >= 30 && code <= 37 || code >= 90 && code <= 97 {
				foreground = code
			}
		}
	}
	if foreground != 0 {
		t.Errorf("%q forces foreground %d instead of the terminal default", text, foreground)
	}
}
