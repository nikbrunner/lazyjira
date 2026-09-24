package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/nikbrunner/lazyjira/v2/pkg/tui/theme"
)

// HelpItem represents a single keybinding hint.
type HelpItem struct {
	Key         string
	Description string
}

// HelpBar displays context-sensitive keybinding hints at the bottom of the screen.
type HelpBar struct {
	items     []HelpItem
	statusMsg string // transient message shown left of hints (green)
	width     int
}

// NewHelpBar creates a help bar with the given items.
func NewHelpBar(items []HelpItem) HelpBar {
	return HelpBar{items: items}
}

// SetItems replaces the current help items.
func (h *HelpBar) SetItems(items []HelpItem) {
	h.items = items
}

// SetStatusMsg sets a transient message shown left of hints.
func (h *HelpBar) SetStatusMsg(msg string) {
	h.statusMsg = msg
}

// SetWidth updates the help bar width.
func (h *HelpBar) SetWidth(w int) {
	h.width = w
}

func (h HelpBar) Init() tea.Cmd {
	return nil
}

func (h HelpBar) Update(msg tea.Msg) (HelpBar, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		h.width = msg.Width
	}
	return h, nil
}

func (h HelpBar) View() string {
	blueStyle := lipgloss.NewStyle().Foreground(theme.ColorBlue)
	sep := blueStyle.Render(" | ")

	// Status message (green) on the left.
	prefix := " "
	if h.statusMsg != "" {
		greenStyle := lipgloss.NewStyle().Foreground(theme.ColorGreen).Bold(true)
		prefix = " " + greenStyle.Render(h.statusMsg) + " "
	}
	availW := h.width - lipgloss.Width(prefix)

	var parts []string
	totalWidth := 0
	truncated := false
	ellipsis := blueStyle.Render(" ...")
	ellipsisWidth := lipgloss.Width(ellipsis)
	for i, item := range h.items {
		part := blueStyle.Render(item.Description + ": " + item.Key)
		separatorWidth := 0
		if len(parts) > 0 {
			separatorWidth = lipgloss.Width(sep)
		}
		nextWidth := totalWidth + separatorWidth + lipgloss.Width(part)
		if availW > 0 && (nextWidth > availW || (i < len(h.items)-1 && nextWidth+ellipsisWidth > availW)) {
			truncated = true
			break
		}
		parts = append(parts, part)
		totalWidth = nextWidth
	}

	result := strings.Join(parts, sep)
	if truncated {
		if availW > 0 && ellipsisWidth > availW {
			result += blueStyle.Render(TruncateEnd("...", availW))
		} else {
			result += ellipsis
		}
	}
	return prefix + result
}
