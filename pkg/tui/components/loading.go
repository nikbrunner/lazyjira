package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/nikbrunner/lazyjira/v2/pkg/tui/theme"
)

const loadingIndicatorFrames = "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"

const loadingIndicatorInterval = 120 * time.Millisecond

// LoadingIndicatorTickMsg advances one loading indicator animation.
type LoadingIndicatorTickMsg struct {
	indicator  *LoadingIndicator
	generation uint64
}

// LoadingIndicator renders an animated dot glyph beside a label.
type LoadingIndicator struct {
	label      string
	frame      int
	generation uint64
	active     bool
}

func NewLoadingIndicator(label string) *LoadingIndicator {
	return &LoadingIndicator{label: label}
}

func (l *LoadingIndicator) Start() tea.Cmd {
	if l.active {
		return nil
	}
	l.active = true
	l.generation++
	return l.nextTick()
}

func (l *LoadingIndicator) Stop() {
	l.active = false
	l.generation++
}

func (l *LoadingIndicator) Update(msg tea.Msg) tea.Cmd {
	if !l.active || !l.owns(msg) {
		return nil
	}
	l.frame = (l.frame + 1) % len([]rune(loadingIndicatorFrames))
	return l.nextTick()
}

func (l *LoadingIndicator) View() string {
	frame := []rune(loadingIndicatorFrames)[l.frame]
	icon := lipgloss.NewStyle().Foreground(theme.ColorOrange).Render(string(frame))
	return icon + " " + l.label
}

func (l *LoadingIndicator) owns(msg tea.Msg) bool {
	tick, ok := msg.(LoadingIndicatorTickMsg)
	return ok && tick.indicator == l && tick.generation == l.generation
}

func (l *LoadingIndicator) nextTick() tea.Cmd {
	generation := l.generation
	return tea.Tick(loadingIndicatorInterval, func(time.Time) tea.Msg {
		return LoadingIndicatorTickMsg{indicator: l, generation: generation}
	})
}
