package components

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/nikbrunner/lazyjira/pkg/tui/keyname"
)

// ConfirmAcceptedMsg is sent when the user answers yes.
type ConfirmAcceptedMsg struct{}

// ConfirmCancelledMsg is sent when the user answers no.
type ConfirmCancelledMsg struct{}

// ConfirmModal asks a yes/no question. While visible it consumes every key.
type ConfirmModal struct {
	question string
	visible  bool
	width    int
	height   int
}

func NewConfirmModal() ConfirmModal {
	return ConfirmModal{}
}

func (c *ConfirmModal) Show(question string) {
	c.question = question
	c.visible = true
}

func (c *ConfirmModal) Hide()           { c.visible = false }
func (c *ConfirmModal) IsVisible() bool { return c.visible }
func (c *ConfirmModal) SetSize(w, h int) {
	c.width = w
	c.height = h
}

func (c *ConfirmModal) Update(msg tea.Msg) (ConfirmModal, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !c.visible || !ok {
		return *c, nil
	}
	switch keyname.Of(key) {
	case "y", keyEnter:
		c.visible = false
		return *c, func() tea.Msg { return ConfirmAcceptedMsg{} }
	case "n", keyEsc, "q":
		c.visible = false
		return *c, func() tea.Msg { return ConfirmCancelledMsg{} }
	}
	return *c, nil
}

func (c *ConfirmModal) View() string {
	if !c.visible {
		return ""
	}
	contentW := min(max(lipgloss.Width(c.question)+4, 30), c.width-4)
	body := " " + TruncateEnd(c.question, max(contentW-3, 1))
	return RenderPanelFull("Confirm", "y yes · n no", body, contentW, 1, true, nil)
}

// Intercept handles key and mouse messages if the modal is visible.
func (c *ConfirmModal) Intercept(msg tea.Msg) (tea.Cmd, bool) {
	if !c.visible {
		return nil, false
	}
	switch msg.(type) {
	case tea.KeyPressMsg:
		updated, cmd := c.Update(msg)
		*c = updated
		return cmd, true
	case tea.PasteMsg, tea.MouseMsg:
		return nil, true
	}
	return nil, false
}

// Render draws the modal centered on bg.
func (c *ConfirmModal) Render(bg string, w, h int) string {
	if !c.visible {
		return bg
	}
	return centerOverlay(bg, c.View(), w, h)
}
