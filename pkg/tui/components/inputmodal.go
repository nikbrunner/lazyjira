package components

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/nikbrunner/lazyjira/pkg/tui/keyname"
	"github.com/nikbrunner/lazyjira/pkg/tui/theme"
)

// InputConfirmedMsg is sent when the user presses Enter in the input modal
type InputConfirmedMsg struct{ Text string }

// InputCancelledMsg is sent when the user presses Esc
type InputCancelledMsg struct{}

// InputModal is a single-line text input popup with an optional hints list below the input
type InputModal struct {
	title      string
	text       []rune
	cursor     int
	visible    bool
	focusInput bool
	hints      []string
	hintCursor int
	width      int
	height     int
}

func NewInputModal() InputModal {
	return InputModal{}
}

// Show opens the input modal with a title and pre-filled text
func (m *InputModal) Show(title, prefill string) {
	m.title = title
	m.text = []rune(prefill)
	m.cursor = len(m.text)
	m.visible = true
	m.focusInput = true
	m.hints = nil
	m.hintCursor = 0
}

// SetHints sets the optional hint items shown below the input
func (m *InputModal) SetHints(hints []string) {
	m.hints = hints
	m.hintCursor = 0
}

func (m *InputModal) Hide()           { m.visible = false }
func (m *InputModal) IsVisible() bool { return m.visible }
func (m *InputModal) HasHints() bool  { return len(m.hints) > 0 }
func (m *InputModal) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *InputModal) Update(msg tea.Msg) (InputModal, tea.Cmd) {
	if !m.visible {
		return *m, nil
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if keyname.Of(keyname.WithoutAlt(msg)) == keyTab && len(m.hints) > 0 {
			m.focusInput = !m.focusInput
			return *m, nil
		}
		if !m.focusInput {
			return m.handleHintKeys(msg)
		}
		return m.handleTextInput(msg)
	case tea.PasteMsg:
		if m.focusInput {
			m.insert([]rune(msg.Content))
		}
	}
	return *m, nil
}

func (m *InputModal) handleHintKeys(msg tea.KeyPressMsg) (InputModal, tea.Cmd) {
	switch keyname.Of(keyname.WithoutAlt(msg)) {
	case keyEnter:
		if m.hintCursor >= 0 && m.hintCursor < len(m.hints) {
			m.visible = false
			text := m.hints[m.hintCursor]
			return *m, func() tea.Msg { return InputConfirmedMsg{Text: text} }
		}
	case keyEsc:
		m.visible = false
		return *m, func() tea.Msg { return InputCancelledMsg{} }
	case keyDown, KeyCtrlJ:
		if m.hintCursor < len(m.hints)-1 {
			m.hintCursor++
		}
	case "up", KeyCtrlK:
		if m.hintCursor > 0 {
			m.hintCursor--
		}
	default:
		switch keyname.Of(msg) {
		case "q":
			m.visible = false
			return *m, func() tea.Msg { return InputCancelledMsg{} }
		case "j":
			if m.hintCursor < len(m.hints)-1 {
				m.hintCursor++
			}
		case "k":
			if m.hintCursor > 0 {
				m.hintCursor--
			}
		}
	}
	return *m, nil
}

func (m *InputModal) handleTextInput(msg tea.KeyPressMsg) (InputModal, tea.Cmd) {
	msg = keyname.WithoutAlt(msg)
	switch keyname.Of(msg) {
	case keyEnter:
		m.visible = false
		text := string(m.text)
		return *m, func() tea.Msg { return InputConfirmedMsg{Text: text} }
	case keyEsc:
		m.visible = false
		return *m, func() tea.Msg { return InputCancelledMsg{} }
	case keyBackspace:
		if m.cursor > 0 {
			m.text = append(m.text[:m.cursor-1], m.text[m.cursor:]...)
			m.cursor--
		}
	case "delete":
		if m.cursor < len(m.text) {
			m.text = append(m.text[:m.cursor], m.text[m.cursor+1:]...)
		}
	case "left":
		if m.cursor > 0 {
			m.cursor--
		}
	case "right":
		if m.cursor < len(m.text) {
			m.cursor++
		}
	case "home", "ctrl+a":
		m.cursor = 0
	case "end", "ctrl+e":
		m.cursor = len(m.text)
	case keyCtrlU:
		m.text = m.text[m.cursor:]
		m.cursor = 0
	case KeyCtrlK:
		m.text = m.text[:m.cursor]
	default:
		m.insert([]rune(keyname.Text(msg)))
	}
	return *m, nil
}

func (m *InputModal) insert(runes []rune) {
	newText := make([]rune, 0, len(m.text)+len(runes))
	newText = append(newText, m.text[:m.cursor]...)
	newText = append(newText, runes...)
	newText = append(newText, m.text[m.cursor:]...)
	m.text = newText
	m.cursor += len(runes)
}

func (m *InputModal) View() string {
	if !m.visible {
		return ""
	}

	contentW := min(max(m.width*6/10, 30), m.width-4)
	innerW := contentW - 2

	cursorStyle := lipgloss.NewStyle().Foreground(theme.ColorCyan)
	allRunes := append([]rune{' '}, m.text...)
	cursorPos := m.cursor + 1

	type wrappedLine struct {
		runes []rune
		start int
	}
	var wrapped []wrappedLine
	off := 0
	for off < len(allRunes) {
		cut := 0
		w := 0
		for i := off; i < len(allRunes); i++ {
			rw := lipgloss.Width(string(allRunes[i]))
			if w+rw > innerW {
				break
			}
			w += rw
			cut = i + 1
		}
		if cut <= off {
			cut = off + 1
		}
		wrapped = append(wrapped, wrappedLine{runes: allRunes[off:cut], start: off})
		off = cut
	}
	if len(wrapped) == 0 {
		wrapped = append(wrapped, wrappedLine{})
	}

	var lines []string
	cursorPlaced := false
	for _, wl := range wrapped {
		lineEnd := wl.start + len(wl.runes)
		lineW := lipgloss.Width(string(wl.runes))

		if !cursorPlaced && (cursorPos < lineEnd || (cursorPos == lineEnd && cursorPos >= len(allRunes))) {
			col := cursorPos - wl.start
			switch {
			case col >= len(wl.runes) && lineW >= innerW:
				lines = append(lines, string(wl.runes))
				lines = append(lines, cursorStyle.Render("█")+strings.Repeat(" ", innerW-1))
			case col >= len(wl.runes):
				rendered := string(wl.runes) + cursorStyle.Render("█")
				if w := lipgloss.Width(rendered); w < innerW {
					rendered += strings.Repeat(" ", innerW-w)
				}
				lines = append(lines, rendered)
			default:
				before := string(wl.runes[:col])
				at := string(wl.runes[col : col+1])
				after := string(wl.runes[col+1:])
				rendered := before + cursorStyle.Render(at) + after
				if w := lipgloss.Width(rendered); w < innerW {
					rendered += strings.Repeat(" ", innerW-w)
				}
				lines = append(lines, rendered)
			}
			cursorPlaced = true
		} else {
			line := string(wl.runes)
			if w := lipgloss.Width(line); w < innerW {
				line += strings.Repeat(" ", innerW-w)
			}
			lines = append(lines, line)
		}
	}

	body := strings.Join(lines, "\n")

	return RenderPanelFull(m.title, "", body, contentW, len(lines), m.focusInput, nil)
}

// Intercept handles a message if the modal is visible
func (m *InputModal) Intercept(msg tea.Msg) (tea.Cmd, bool) {
	if !m.visible {
		return nil, false
	}
	switch msg.(type) {
	case tea.KeyPressMsg, tea.PasteMsg:
		updated, cmd := m.Update(msg)
		*m = updated
		return cmd, true
	}
	return nil, false
}

// Render draws the input modal centered on bg with an optional hint panel
func (m *InputModal) Render(bg string, w, h int) string {
	if !m.visible {
		return bg
	}
	popup := m.View()
	return centerOverlayWithHint(bg, popup, m.HintView(), w, h)
}

// HintView returns a bordered panel with hint items or an empty string if no hints are set
func (m *InputModal) HintView() string {
	if !m.visible || len(m.hints) == 0 {
		return ""
	}

	contentW := min(max(m.width*6/10, 30), m.width-4)
	innerW := contentW - 2

	selStyle := lipgloss.NewStyle().Background(theme.ColorHighlight)
	normalStyle := lipgloss.NewStyle()

	maxHints := min(5, len(m.hints))
	start := 0
	if m.hintCursor >= maxHints {
		start = m.hintCursor - maxHints + 1
	}
	end := start + maxHints
	if end > len(m.hints) {
		end = len(m.hints)
		start = max(0, end-maxHints)
	}

	var lines []string
	for i := start; i < end; i++ {
		line := " " + m.hints[i]
		line = TruncateEnd(line, innerW)
		if lineW := lipgloss.Width(line); lineW < innerW {
			line += strings.Repeat(" ", innerW-lineW)
		}
		if i == m.hintCursor && !m.focusInput {
			lines = append(lines, selStyle.Render(line))
		} else {
			lines = append(lines, normalStyle.Render(line))
		}
	}

	body := strings.Join(lines, "\n")
	title := "Existing branches"
	footer := fmt.Sprintf("%d of %d", m.hintCursor+1, len(m.hints))
	return RenderPanelFull(title, footer, body, contentW, len(lines), !m.focusInput, nil)
}
