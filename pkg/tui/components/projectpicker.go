package components

import (
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/nikbrunner/lazyjira/v2/pkg/tui/theme"
)

// ProjectChoice is a project available in the project picker.
type ProjectChoice struct {
	Key  string
	Name string
}

// ProjectPickerSelectedMsg is sent after a project is confirmed.
type ProjectPickerSelectedMsg struct{ Project ProjectChoice }

// ProjectPickerCancelledMsg is sent after the picker is cancelled.
type ProjectPickerCancelledMsg struct{}

// ProjectPicker is an immediate-filter project selection overlay.
type ProjectPicker struct {
	projects []ProjectChoice
	filtered []ProjectChoice
	query    string
	cursor   int
	visible  bool
	width    int
	height   int
}

func NewProjectPicker() ProjectPicker { return ProjectPicker{} }

// Show opens the picker with the given projects and clears any previous filter.
func (p *ProjectPicker) Show(projects []ProjectChoice) {
	p.projects = append([]ProjectChoice(nil), projects...)
	p.filtered = append(p.filtered[:0], p.projects...)
	p.query = ""
	p.cursor = 0
	p.visible = true
}

func (p *ProjectPicker) IsVisible() bool { return p.visible }

func (p *ProjectPicker) SetSize(w, h int) {
	p.width, p.height = w, h
}

func (p *ProjectPicker) Update(msg tea.Msg) (ProjectPicker, tea.Cmd) {
	if !p.visible {
		return *p, nil
	}
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return *p, nil
	}
	if key.Type == tea.KeySpace {
		p.query += " "
		p.applyFilter()
		return *p, nil
	}
	switch key.String() {
	case keyEsc:
		p.visible = false
		return *p, func() tea.Msg { return ProjectPickerCancelledMsg{} }
	case keyEnter:
		if p.cursor >= 0 && p.cursor < len(p.filtered) {
			project := p.filtered[p.cursor]
			p.visible = false
			return *p, func() tea.Msg { return ProjectPickerSelectedMsg{Project: project} }
		}
	case "up":
		p.move(-1)
	case "down":
		p.move(1)
	case "backspace", "⌫":
		if p.query != "" {
			runes := []rune(p.query)
			p.query = string(runes[:len(runes)-1])
			p.applyFilter()
		}
	default:
		if key.Type == tea.KeyRunes && len(key.Runes) > 0 {
			for _, r := range key.Runes {
				if unicode.IsPrint(r) {
					p.query += string(r)
				}
			}
			p.applyFilter()
		}
	}
	return *p, nil
}

func (p *ProjectPicker) move(delta int) {
	if len(p.filtered) == 0 {
		return
	}
	p.cursor = (p.cursor + delta + len(p.filtered)) % len(p.filtered)
}

func (p *ProjectPicker) applyFilter() {
	query := strings.ToLower(p.query)
	p.filtered = p.filtered[:0]
	for _, project := range p.projects {
		if strings.Contains(strings.ToLower(project.Key), query) || strings.Contains(strings.ToLower(project.Name), query) {
			p.filtered = append(p.filtered, project)
		}
	}
	p.cursor = 0
}

// View renders the popup content without compositing it over a background.
func (p *ProjectPicker) View() string {
	if !p.visible {
		return ""
	}
	width := min(max(p.width*7/10, 30), 70)
	if p.width > 0 && width > p.width-4 {
		width = max(p.width-4, 1)
	}
	innerWidth := max(width-4, 1)
	maxRows := max(min(p.height-8, 12), 1)
	maxRows = min(maxRows, max(len(p.filtered), 1))
	start := 0
	if p.cursor >= maxRows {
		start = p.cursor - maxRows + 1
	}
	end := min(start+maxRows, len(p.filtered))
	filterLine := ansi.Truncate("Filter: "+p.query, innerWidth, "…")
	lines := []string{lipgloss.NewStyle().Foreground(theme.ColorGray).Render(filterLine)}
	for i := start; i < end; i++ {
		project := p.filtered[i]
		label := project.Key
		if project.Name != "" {
			label += "  " + project.Name
		}
		label = lipgloss.NewStyle().Width(innerWidth).MaxWidth(innerWidth).Render(label)
		if i == p.cursor {
			label = theme.Default.SelectedItem.Render(label)
		}
		lines = append(lines, label)
	}
	if len(p.filtered) == 0 {
		lines = append(lines, "No projects match")
	}
	content := strings.Join(lines, "\n")
	return RenderPanel("Projects", content, width, len(lines), true)
}

// Intercept routes all input to this picker while it is open.
func (p *ProjectPicker) Intercept(msg tea.Msg) (tea.Cmd, bool) {
	if !p.visible {
		return nil, false
	}
	if _, ok := msg.(tea.MouseMsg); ok {
		return nil, true
	}
	if _, ok := msg.(tea.KeyMsg); !ok {
		return nil, false
	}
	updated, cmd := p.Update(msg)
	*p = updated
	return cmd, true
}

// Render draws the picker centered over bg.
func (p *ProjectPicker) Render(bg string, w, h int) string {
	if !p.visible {
		return bg
	}
	p.SetSize(w, h)
	return centerOverlay(bg, p.View(), w, h)
}
