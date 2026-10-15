package components

import (
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/nikbrunner/lazyjira/pkg/tui/keyname"
	"github.com/nikbrunner/lazyjira/pkg/tui/theme"
)

// WorktreeConfirmedMsg carries the trimmed name, used as the branch, and the absolute destination path.
type WorktreeConfirmedMsg struct {
	Name string
	Path string
}

// WorktreeCancelledMsg is sent when the user dismisses the modal.
type WorktreeCancelledMsg struct{}

const (
	worktreeTitle  = "New worktree"
	worktreeFooter = "enter create · esc cancel"
	worktreeLabelW = 6
	// worktreePrefixW is the leading space, label, and gap before each value.
	worktreePrefixW = 1 + worktreeLabelW + 2
)

// WorktreeModal edits the name of a new worktree and previews its path and branch.
type WorktreeModal struct {
	input        TextInput
	parent       string
	home         string
	branchExists func(string) bool
	existing     bool
	checked      string
	visible      bool
	width        int
	height       int
}

func NewWorktreeModal() WorktreeModal {
	return WorktreeModal{input: NewTextInput()}
}

// Show opens the modal with a pre-filled name. parent is the absolute directory
// the worktree goes into; home is abbreviated to ~ in the path preview.
func (m *WorktreeModal) Show(name, parent, home string, branchExists func(string) bool) {
	m.input.SetValue(name)
	m.parent = parent
	m.home = home
	m.branchExists = branchExists
	m.checked = ""
	m.refreshBranch()
	m.visible = true
}

func (m *WorktreeModal) Hide()           { m.visible = false }
func (m *WorktreeModal) IsVisible() bool { return m.visible }
func (m *WorktreeModal) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *WorktreeModal) name() string { return strings.TrimSpace(m.input.Value()) }

func (m *WorktreeModal) path() string { return filepath.Join(m.parent, m.name()) }

func (m *WorktreeModal) refreshBranch() {
	name := m.name()
	if name == m.checked {
		return
	}
	m.checked = name
	m.existing = name != "" && m.branchExists != nil && m.branchExists(name)
}

func (m *WorktreeModal) Update(msg tea.Msg) (WorktreeModal, tea.Cmd) {
	if !m.visible {
		return *m, nil
	}
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch keyname.Of(keyname.WithoutAlt(key)) {
		case keyEnter:
			name := m.name()
			if name == "" {
				return *m, nil
			}
			m.visible = false
			confirmed := WorktreeConfirmedMsg{Name: name, Path: m.path()}
			return *m, func() tea.Msg { return confirmed }
		case keyEsc:
			m.visible = false
			return *m, func() tea.Msg { return WorktreeCancelledMsg{} }
		}
	}
	if _, changed := m.input.Update(msg); changed {
		m.refreshBranch()
	}
	return *m, nil
}

func (m *WorktreeModal) displayPath() string {
	p := m.path()
	if m.home == "" {
		return p
	}
	if p == m.home {
		return "~"
	}
	if rest, ok := strings.CutPrefix(p, m.home+string(filepath.Separator)); ok {
		return "~/" + rest
	}
	return p
}

func (m *WorktreeModal) marker() string {
	switch {
	case m.name() == "":
		return ""
	case m.existing:
		return " (existing)"
	default:
		return " (new)"
	}
}

func (m *WorktreeModal) contentWidth(path string) int {
	nameW := lipgloss.Width(m.input.Value()) + 1
	valueW := max(nameW, lipgloss.Width(path), lipgloss.Width(m.name()+m.marker()))
	natural := max(worktreePrefixW+valueW+1+2, lipgloss.Width(worktreeFooter)+4, lipgloss.Width(worktreeTitle)+4)
	return max(min(natural, m.width-4), worktreePrefixW+8)
}

func worktreeRow(label, value string) string {
	muted := lipgloss.NewStyle().Foreground(theme.ColorGray)
	return " " + muted.Render(label+strings.Repeat(" ", worktreeLabelW-lipgloss.Width(label))) + "  " + value
}

func (m *WorktreeModal) View() string {
	if !m.visible {
		return ""
	}
	path := m.displayPath()
	contentW := m.contentWidth(path)
	budget := max(contentW-2-worktreePrefixW-1, 1)

	m.input.SetWidth(budget)
	marker := m.marker()
	branch := TruncateMiddle(m.name(), max(budget-lipgloss.Width(marker), 1)) +
		lipgloss.NewStyle().Foreground(theme.ColorGray).Render(marker)

	body := strings.Join([]string{
		worktreeRow("Name", m.input.View()),
		worktreeRow("Path", TruncateMiddle(path, budget)),
		worktreeRow("Branch", branch),
	}, "\n")
	return RenderPanelFull(worktreeTitle, worktreeFooter, body, contentW, 3, true, nil)
}

// Intercept handles key, paste, and mouse messages while the modal is visible.
func (m *WorktreeModal) Intercept(msg tea.Msg) (tea.Cmd, bool) {
	if !m.visible {
		return nil, false
	}
	switch msg.(type) {
	case tea.KeyPressMsg, tea.PasteMsg:
		updated, cmd := m.Update(msg)
		*m = updated
		return cmd, true
	case tea.MouseMsg:
		return nil, true
	}
	return nil, false
}

// Render draws the modal centered on bg.
func (m *WorktreeModal) Render(bg string, w, h int) string {
	if !m.visible {
		return bg
	}
	return centerOverlay(bg, m.View(), w, h)
}
