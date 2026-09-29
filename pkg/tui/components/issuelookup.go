package components

import (
	"regexp"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/nikbrunner/lazyjira/pkg/tui/theme"
)

// LookupSuggestion is a project or issue offered by the issue lookup.
type LookupSuggestion struct {
	Key     string
	Label   string
	Project bool
}

// IssueLookupSelectedMsg is sent when an issue key is confirmed.
type IssueLookupSelectedMsg struct{ Key string }

// IssueLookupCancelledMsg is sent when the lookup is closed without a key.
type IssueLookupCancelledMsg struct{}

// IssueLookupQueryMsg asks for issue suggestions for a query that contains a
// project key and a dash.
type IssueLookupQueryMsg struct{ Query string }

var issueKeyPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*-[0-9]+$`)

// IssueLookup is a prompt that completes a project key, then an issue key.
type IssueLookup struct {
	projects    []ProjectChoice
	issues      []LookupSuggestion
	issueQuery  string
	suggestions []LookupSuggestion
	query       string
	cursor      int
	navigated   bool
	visible     bool
	width       int
	height      int
}

func NewIssueLookup() IssueLookup { return IssueLookup{} }

// Show opens the lookup with prefill as the query. The returned command asks
// for issue suggestions when prefill already contains a dash.
func (l *IssueLookup) Show(prefill string, projects []ProjectChoice) tea.Cmd {
	l.projects = append([]ProjectChoice(nil), projects...)
	l.issues = nil
	l.issueQuery = ""
	l.visible = true
	return l.setQuery(prefill)
}

func (l *IssueLookup) IsVisible() bool { return l.visible }

func (l *IssueLookup) Query() string { return l.query }

func (l *IssueLookup) SetSize(w, h int) { l.width, l.height = w, h }

// SetIssueSuggestions shows issues for query. Results for an outdated query
// are ignored.
func (l *IssueLookup) SetIssueSuggestions(query string, issues []LookupSuggestion) {
	if !l.visible || !strings.EqualFold(query, l.query) {
		return
	}
	l.issueQuery = query
	l.issues = issues
	l.refresh()
}

func (l *IssueLookup) Update(msg tea.Msg) (IssueLookup, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !l.visible || !ok {
		return *l, nil
	}
	switch key.String() {
	case keyEsc:
		l.visible = false
		return *l, func() tea.Msg { return IssueLookupCancelledMsg{} }
	case keyEnter:
		return *l, l.confirm()
	case "tab":
		if s, ok := l.selected(); ok {
			return *l, l.complete(s)
		}
	case "up", "ctrl+p":
		l.move(-1)
	case "down", "ctrl+n":
		l.move(1)
	case "backspace", "⌫":
		if l.query != "" {
			runes := []rune(l.query)
			return *l, l.setQuery(string(runes[:len(runes)-1]))
		}
	case "ctrl+u":
		return *l, l.setQuery("")
	default:
		if key.Type == tea.KeyRunes {
			var typed strings.Builder
			for _, r := range key.Runes {
				if unicode.IsPrint(r) && !unicode.IsSpace(r) {
					typed.WriteRune(r)
				}
			}
			if typed.Len() > 0 {
				return *l, l.setQuery(l.query + typed.String())
			}
		}
	}
	return *l, nil
}

func (l *IssueLookup) confirm() tea.Cmd {
	s, hasSelection := l.selected()
	switch {
	case hasSelection && l.navigated:
		return l.choose(s)
	case issueKeyPattern.MatchString(l.query):
		return l.selectKey(l.query)
	case hasSelection:
		return l.choose(s)
	}
	return nil
}

func (l *IssueLookup) choose(s LookupSuggestion) tea.Cmd {
	if s.Project {
		return l.complete(s)
	}
	return l.selectKey(s.Key)
}

func (l *IssueLookup) complete(s LookupSuggestion) tea.Cmd {
	if s.Project {
		return l.setQuery(s.Key + "-")
	}
	return l.setQuery(s.Key)
}

func (l *IssueLookup) selectKey(key string) tea.Cmd {
	l.visible = false
	key = strings.ToUpper(key)
	return func() tea.Msg { return IssueLookupSelectedMsg{Key: key} }
}

func (l *IssueLookup) setQuery(query string) tea.Cmd {
	l.query = query
	l.navigated = false
	l.refresh()
	if !strings.Contains(query, "-") {
		return nil
	}
	return func() tea.Msg { return IssueLookupQueryMsg{Query: query} }
}

func (l *IssueLookup) selected() (LookupSuggestion, bool) {
	if l.cursor < 0 || l.cursor >= len(l.suggestions) {
		return LookupSuggestion{}, false
	}
	return l.suggestions[l.cursor], true
}

func (l *IssueLookup) move(delta int) {
	if len(l.suggestions) == 0 {
		return
	}
	l.navigated = true
	l.cursor = (l.cursor + delta + len(l.suggestions)) % len(l.suggestions)
}

func (l *IssueLookup) refresh() {
	l.cursor = 0
	l.suggestions = l.suggestions[:0]
	if strings.Contains(l.query, "-") {
		if strings.EqualFold(l.issueQuery, l.query) {
			l.suggestions = append(l.suggestions, l.issues...)
		}
		return
	}
	query := strings.ToLower(l.query)
	var byName []LookupSuggestion
	for _, project := range l.projects {
		s := LookupSuggestion{Key: project.Key, Label: project.Name, Project: true}
		switch {
		case strings.HasPrefix(strings.ToLower(project.Key), query):
			l.suggestions = append(l.suggestions, s)
		case strings.Contains(strings.ToLower(project.Name), query):
			byName = append(byName, s)
		}
	}
	l.suggestions = append(l.suggestions, byName...)
}

// View renders the popup content without compositing it over a background.
func (l *IssueLookup) View() string {
	if !l.visible {
		return ""
	}
	width := min(max(l.width*7/10, 30), 80)
	if l.width > 0 && width > l.width-4 {
		width = max(l.width-4, 1)
	}
	innerWidth := max(width-4, 1)
	maxRows := max(min(l.height-8, 12), 1)
	maxRows = min(maxRows, max(len(l.suggestions), 1))
	start := 0
	if l.cursor >= maxRows {
		start = l.cursor - maxRows + 1
	}
	end := min(start+maxRows, len(l.suggestions))

	prompt := ansi.Truncate("# "+l.query, max(innerWidth-1, 1), "…") + theme.Default.SelectedItem.Render(" ")
	lines := []string{prompt}
	keyWidth := 0
	for _, s := range l.suggestions[start:end] {
		keyWidth = max(keyWidth, len(s.Key))
	}
	for i := start; i < end; i++ {
		s := l.suggestions[i]
		label := s.Key
		if s.Label != "" {
			label += strings.Repeat(" ", keyWidth-len(s.Key)+2) + s.Label
		}
		label = lipgloss.NewStyle().Width(innerWidth).MaxWidth(innerWidth).Render(label)
		if i == l.cursor {
			label = theme.Default.SelectedItem.Render(label)
		}
		lines = append(lines, label)
	}
	if len(l.suggestions) == 0 {
		hint := "No matches"
		if issueKeyPattern.MatchString(l.query) {
			hint = "enter opens " + strings.ToUpper(l.query)
		}
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.ColorGray).Render(hint))
	}
	return RenderPanel("Open issue", strings.Join(lines, "\n"), width, len(lines), true)
}

// Intercept routes all input to the lookup while it is open.
func (l *IssueLookup) Intercept(msg tea.Msg) (tea.Cmd, bool) {
	if !l.visible {
		return nil, false
	}
	if _, ok := msg.(tea.MouseMsg); ok {
		return nil, true
	}
	if _, ok := msg.(tea.KeyMsg); !ok {
		return nil, false
	}
	updated, cmd := l.Update(msg)
	*l = updated
	return cmd, true
}

// Render draws the lookup centered over bg.
func (l *IssueLookup) Render(bg string, w, h int) string {
	if !l.visible {
		return bg
	}
	l.SetSize(w, h)
	return centerOverlay(bg, l.View(), w, h)
}
