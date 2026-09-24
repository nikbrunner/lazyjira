package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/textfuel/lazyjira/v2/pkg/tui/components"
	"github.com/textfuel/lazyjira/v2/pkg/tui/theme"
)

type copyToastExpiredMsg struct{ epoch int }

func (a *App) showCopyToast(text string) tea.Cmd {
	a.toastEpoch++
	a.toastText = text
	epoch := a.toastEpoch
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return copyToastExpiredMsg{epoch: epoch}
	})
}

func (a *App) handleCopyToastExpired(msg copyToastExpiredMsg) (tea.Model, tea.Cmd) {
	if msg.epoch == a.toastEpoch {
		a.toastText = ""
	}
	return a, nil
}

func (a *App) renderCopyToast(base string) string {
	if a.toastText == "" || a.width < 8 || a.height < 5 {
		return base
	}

	contentWidth := min(60, a.width-2) - 4
	style := lipgloss.NewStyle().
		Foreground(theme.ColorGreen).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorGreen).
		Padding(0, 1)
	popup := style.Render(ansi.Truncate(a.toastText, contentWidth, "…"))
	x := a.width - lipgloss.Width(popup) - 1
	y := a.height - lipgloss.Height(popup) - 2
	return components.OverlayAt(base, popup, x, y, a.width, a.height)
}
