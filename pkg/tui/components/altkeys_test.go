package components

import (
	"testing"

	"github.com/nikbrunner/lazyjira/pkg/internal/testkit"
	"github.com/nikbrunner/lazyjira/pkg/tui/keyname/keytest"
)

// Text inputs read alt combinations as the plain key.

func TestTextInput_AltKeysActLikePlainKeys(t *testing.T) {
	t.Parallel()
	input := NewTextInput()
	input.SetValue("abc")

	input, _ = input.Update(keytest.Press("alt+backspace"))
	input, _ = input.Update(keytest.Press("alt+x"))

	testkit.AssertEqual(t, "value", input.Value(), "abx")
}

func TestSearchBar_AltEnterConfirms(t *testing.T) {
	t.Parallel()
	bar := NewSearchBar()
	bar.Activate()

	bar, cmd := bar.Update(keytest.Press("alt+enter"))

	testkit.AssertEqual(t, "closed", bar.IsActive(), false)
	if _, ok := cmd().(SearchConfirmedMsg); !ok {
		t.Error("alt+enter should confirm the search")
	}
}

func TestInputModal_AltKeysActLikePlainKeys(t *testing.T) {
	t.Parallel()
	m := NewInputModal()
	m.Show("Branch", "feat/ab")

	m.Intercept(keytest.Press("alt+backspace"))
	cmd, _ := m.Intercept(keytest.Press("alt+enter"))

	confirmed, ok := cmd().(InputConfirmedMsg)
	if !ok || confirmed.Text != "feat/a" {
		t.Errorf("alt+enter = %#v, want confirmation of feat/a", cmd())
	}
}

func TestJQLModal_AltEnterSubmits(t *testing.T) {
	t.Parallel()
	m := NewJQLModal()
	m.Show("project = X", nil)

	cmd, _ := m.Intercept(keytest.Press("alt+enter"))

	if _, ok := cmd().(JQLSubmitMsg); !ok {
		t.Errorf("alt+enter = %#v, want JQLSubmitMsg", cmd())
	}
}

func TestCreateForm_SummaryAltKeysActLikePlainKeys(t *testing.T) {
	t.Parallel()
	form := NewCreateForm()
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.setFocus(CreatePanelSummary)
	before := string(form.summaryText)
	for _, key := range keytest.Type("ab") {
		form.Intercept(key)
	}

	form.Intercept(keytest.Press("alt+backspace"))
	form.Intercept(keytest.Press("alt+enter"))

	testkit.AssertEqual(t, "summary", string(form.summaryText), before+"a")
	testkit.AssertEqual(t, "focus", form.FocusedPanel(), CreatePanelDescription)
}
