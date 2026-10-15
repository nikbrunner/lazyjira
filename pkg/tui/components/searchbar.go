package components

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/nikbrunner/lazyjira/pkg/tui/keyname"
	"github.com/nikbrunner/lazyjira/pkg/tui/theme"
)

// SearchBar is a bottom search input that filters the current panel
type SearchBar struct {
	input  TextInput
	active bool
	prefix string
}

func NewSearchBar() SearchBar {
	return SearchBar{}
}

func (s *SearchBar) SetWidth(w int) { s.input.SetWidth(w) }
func (s *SearchBar) IsActive() bool { return s.active }
func (s *SearchBar) Query() string  { return s.input.Value() }
func (s *SearchBar) Activate()      { s.ActivateWithPrefix("/") }
func (s *SearchBar) Deactivate()    { s.active = false; s.input.SetValue("") }

// ActivateWithPrefix opens the bar with prefix shown before the query.
func (s *SearchBar) ActivateWithPrefix(prefix string) {
	s.active = true
	s.prefix = prefix
	s.input.SetValue("")
}

// SearchChangedMsg is sent when the search query changes
type SearchChangedMsg struct{ Query string }

// SearchConfirmedMsg is sent when the user presses Enter.
type SearchConfirmedMsg struct{ Query string }

// SearchCancelledMsg is sent when the user presses Esc.
type SearchCancelledMsg struct{}

func (s *SearchBar) Update(msg tea.Msg) (SearchBar, tea.Cmd) {
	if !s.active {
		return *s, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch keyname.Of(keyname.WithoutAlt(msg)) {
		case keyEnter:
			s.active = false
			q := s.input.Value()
			return *s, func() tea.Msg { return SearchConfirmedMsg{Query: q} }
		case keyEsc:
			s.active = false
			s.input.SetValue("")
			return *s, func() tea.Msg { return SearchCancelledMsg{} }
		}
	case tea.PasteMsg:
	default:
		return *s, nil
	}
	updated, changed := s.input.Update(msg)
	s.input = updated
	if changed {
		q := s.input.Value()
		return *s, func() tea.Msg { return SearchChangedMsg{Query: q} }
	}
	return *s, nil
}

func (s *SearchBar) View() string {
	if !s.active {
		return ""
	}
	return renderFilterBar(s.prefix, &s.input)
}

// RenderFilterBarInput renders filter bar using TextInput (with cursor positioning)
func RenderFilterBarInput(input *TextInput) string {
	return renderFilterBar("/", input)
}

func renderFilterBar(prefix string, input *TextInput) string {
	prefixStyle := lipgloss.NewStyle().
		Foreground(theme.ColorCyan).
		Bold(true)

	return prefixStyle.Render(" "+prefix) + input.View()
}
