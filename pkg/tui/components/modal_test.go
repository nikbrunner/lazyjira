package components

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModal_ShowAndSelect(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	items := []ModalItem{
		{ID: "1", Label: "First"},
		{ID: "2", Label: "Second"},
		{ID: "3", Label: "Third"},
	}
	m.Show("Pick", items)

	if !m.IsVisible() {
		t.Fatal("modal should be visible after Show")
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if m.IsVisible() {
		t.Error("modal should hide after selection")
	}
	if cmd == nil {
		t.Fatal("expected a command from selection")
	}
	msg := cmd()
	sel, ok := msg.(ModalSelectedMsg)
	if !ok {
		t.Fatalf("expected ModalSelectedMsg, got %T", msg)
	}
	if sel.Item.ID != "3" {
		t.Errorf("selected ID = %q, want %q", sel.Item.ID, "3")
	}
}

func TestModal_Cancel(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	m.Show("Pick", []ModalItem{{ID: "1", Label: "One"}})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if m.IsVisible() {
		t.Error("modal should hide after esc")
	}
	if cmd == nil {
		t.Fatal("expected cancel command")
	}
	msg := cmd()
	if _, ok := msg.(ModalCancelledMsg); !ok {
		t.Fatalf("expected ModalCancelledMsg, got %T", msg)
	}
}

func TestModal_SkipsSeparators(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	items := []ModalItem{
		{Label: "Section", Separator: true},
		{ID: "1", Label: "First"},
		{ID: "2", Label: "Second"},
	}
	m.Show("Pick", items)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	msg := cmd()
	sel := msg.(ModalSelectedMsg)
	if sel.Item.ID != "1" {
		t.Errorf("selected ID = %q, want %q (should skip separator)", sel.Item.ID, "1")
	}
}

func TestModal_InterceptWhenVisible(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)

	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	_, ok := m.Intercept(key)
	if ok {
		t.Error("hidden modal should not intercept")
	}

	m.Show("Pick", []ModalItem{{ID: "1", Label: "One"}})
	_, ok = m.Intercept(key)
	if !ok {
		t.Error("visible modal should intercept key messages")
	}
}

func TestModal_SearchFilter(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	m.Show("Pick", []ModalItem{
		{ID: "1", Label: "Apple"},
		{ID: "2", Label: "Banana"},
		{ID: "3", Label: "Avocado"},
	})

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !m.IsSearching() {
		t.Fatal("search should be active after /")
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	msg := cmd()
	sel := msg.(ModalSelectedMsg)
	if sel.Item.ID != "3" {
		t.Errorf("selected ID = %q, want %q (Avocado)", sel.Item.ID, "3")
	}
}

func TestModal_Checklist(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	items := []ModalItem{
		{ID: "a", Label: "Alpha"},
		{ID: "b", Label: "Beta"},
		{ID: "c", Label: "Gamma"},
	}
	m.ShowChecklist("Select", items, nil)

	if !m.IsChecklist() {
		t.Fatal("should be in checklist mode")
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	msg := cmd()
	confirmed, ok := msg.(ChecklistConfirmedMsg)
	if !ok {
		t.Fatalf("expected ChecklistConfirmedMsg, got %T", msg)
	}
	if len(confirmed.Selected) != 2 {
		t.Errorf("selected %d items, want 2", len(confirmed.Selected))
	}
}

func TestModal_GroupedChecklistRelabelsAndBlocksDisabled(t *testing.T) {
	t.Parallel()
	relabel := func(selected map[string]bool) (string, []ModalItem) {
		return fmt.Sprintf("Filter %d", len(selected)), []ModalItem{
			{Label: "Group", Separator: true},
			{ID: "a", Label: map[bool]string{true: "A on", false: "A off"}[selected["a"]]},
			{ID: "b", Label: "B", Disabled: selected["a"]},
		}
	}
	m := NewModal()
	m.SetSize(80, 24)
	m.ShowGroupedChecklist(nil, relabel)
	if m.Title() != "Filter 0" {
		t.Errorf("title = %q, want Filter 0", m.Title())
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
	if m.items[1].Label != "A on" || !m.items[2].Disabled {
		t.Fatalf("items after toggling a = %+v, want relabelled with b disabled", m.items)
	}
	if m.Title() != "Filter 1" {
		t.Errorf("title after toggle = %q, want Filter 1", m.Title())
	}
	if m.cursor != 1 {
		t.Errorf("cursor = %d, want it to stay on a", m.cursor)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got, ok := cmd().(ChecklistConfirmedMsg)
	if !ok {
		t.Fatal("enter should confirm the checklist")
	}
	if len(got.Selected) != 1 || got.Selected[0].ID != "a" {
		t.Errorf("confirmed = %+v, want only a (b is disabled)", got.Selected)
	}
}
