package views

import (
	"github.com/nikbrunner/lazyjira/pkg/config"
	"github.com/nikbrunner/lazyjira/pkg/tui/theme"
)

const (
	glamourStyleDark  = "dark"
	glamourStyleLight = "light"
	glamourStyleNoTTY = "notty"
)

func ResolveGlamourStyle(style string) string {
	return resolveGlamourStyle(style, theme.IsDark)
}

func resolveGlamourStyle(style string, hasDarkBackground func() bool) string {
	switch style {
	case config.RendererStyleDark:
		return glamourStyleDark
	case config.RendererStyleLight:
		return glamourStyleLight
	case config.RendererStyleNoTTY:
		return glamourStyleNoTTY
	}
	if hasDarkBackground() {
		return glamourStyleDark
	}
	return glamourStyleLight
}
