package views

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/nikbrunner/lazyjira/v2/pkg/tui/components"
)

type StatusPanel struct {
	project   string
	user      string
	host      string
	auth      string
	version   string
	online    bool
	errText   string
	width     int
	height    int
	focused   bool
	focusHint string
}

func NewStatusPanel(project, user, host string) *StatusPanel {
	return &StatusPanel{
		project: project,
		user:    user,
		host:    host,
		online:  true,
	}
}

func (s *StatusPanel) SetProject(project string) { s.project = project }
func (s *StatusPanel) SetOnline(online bool)     { s.online = online }
func (s *StatusPanel) SetError(err string)       { s.errText = err }
func (s *StatusPanel) SetAuthMethod(auth string) { s.auth = auth }
func (s *StatusPanel) SetVersion(version string) { s.version = version }
func (s *StatusPanel) ErrorMessage() string      { return s.errText }
func (s *StatusPanel) SetSize(w, h int)          { s.width = w; s.height = h }
func (s *StatusPanel) SetFocused(focused bool)   { s.focused = focused }
func (s *StatusPanel) SetFocusHint(hint string)  { s.focusHint = hint }

func (s *StatusPanel) title(name string) string {
	if s.focusHint == "" {
		return name
	}
	return "[" + s.focusHint + "] " + name
}

func (s *StatusPanel) Init() tea.Cmd                              { return nil }
func (s *StatusPanel) Update(msg tea.Msg) (*StatusPanel, tea.Cmd) { return s, nil }

func (s *StatusPanel) View() string {
	statusTitle := s.title("App Status")
	state := "✓ Connected"
	if !s.online {
		state = "✗ Disconnected"
	}
	if s.height <= 1 {
		return components.RenderCollapsedBar(statusTitle, state, s.width, s.focused)
	}

	innerHeight := max(s.height-2, 1)
	contentWidth := max(s.width-2, 1)
	account := s.user
	if account == "" {
		account = s.host
	}
	if s.height <= 3 {
		parts := []string{state}
		if s.errText != "" {
			parts = append(parts, "Error: "+s.errText)
		}
		if account != "" {
			parts = append(parts, "Acct: "+account)
		}
		if s.host != "" {
			parts = append(parts, "Host: "+s.host)
		}
		if s.auth != "" {
			parts = append(parts, "Auth: "+s.auth)
		}
		if s.version != "" {
			parts = append(parts, s.version)
		}
		return components.RenderPanel(statusTitle, ansi.Truncate(strings.Join(parts, " · "), contentWidth, "…"), s.width, innerHeight, s.focused)
	}
	lines := []string{state}
	if s.errText != "" {
		lines = append(lines, "Error: "+s.errText)
	}
	lines = append(lines,
		"Account: "+account,
		"Host: "+s.host,
		"Auth: "+s.auth,
		"Version: "+s.version,
	)
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, contentWidth, "…")
	}
	content := strings.Join(lines[:min(innerHeight, len(lines))], "\n")
	return components.RenderPanel(statusTitle, content, s.width, innerHeight, s.focused)
}
