package components

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/nikbrunner/lazyjira/pkg/internal/testkit"
	"github.com/nikbrunner/lazyjira/pkg/tui/keyname/keytest"
)

func TestJQLModal_PasteInsertsIntoInput(t *testing.T) {
	t.Parallel()
	m := NewJQLModal()
	m.Show("project = ", nil)

	cmd, handled := m.Intercept(tea.PasteMsg{Content: "PLAT"})

	testkit.AssertEqual(t, "handled", handled, true)
	testkit.AssertEqual(t, "input", m.InputValue(), "project = PLAT")
	if cmd == nil {
		t.Error("paste should report the changed input")
	}
}

func TestInputModal_PasteInsertsText(t *testing.T) {
	t.Parallel()
	m := NewInputModal()
	m.Show("Branch", "feat/")

	_, handled := m.Intercept(tea.PasteMsg{Content: "x y"})

	testkit.AssertEqual(t, "handled", handled, true)
	testkit.AssertEqual(t, "text", string(m.text), "feat/x y")
}

func TestModal_PasteFiltersWhileSearching(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.Show(testTitle, []ModalItem{{ID: testItem1ID, Label: testItem1Label}, {ID: testItem2ID, Label: testItem2Label}})
	m.Intercept(keytest.Press("/"))

	_, handled := m.Intercept(tea.PasteMsg{Content: "Second"})

	testkit.AssertEqual(t, "handled", handled, true)
	testkit.AssertEqual(t, "filter", m.filterInput.Value(), "Second")
	testkit.AssertEqual(t, "matches", len(m.items), 1)
}

func TestModal_PasteDoesNotTriggerKeys(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.Show(testTitle, []ModalItem{{ID: testItem1ID, Label: testItem1Label}})

	cmd, handled := m.Intercept(tea.PasteMsg{Content: "q"})

	testkit.AssertEqual(t, "handled", handled, true)
	testkit.AssertEqual(t, "still visible", m.IsVisible(), true)
	if cmd != nil {
		t.Error("pasted text must not run a key binding")
	}
}

func TestCreateForm_PasteIntoSummary(t *testing.T) {
	t.Parallel()
	form := NewCreateForm()
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.setFocus(CreatePanelSummary)
	before := string(form.summaryText)

	_, handled := form.Intercept(tea.PasteMsg{Content: "q crash"})

	testkit.AssertEqual(t, "handled", handled, true)
	testkit.AssertEqual(t, "summary", string(form.summaryText), before+"q crash")
}
