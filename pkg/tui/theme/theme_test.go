package theme

import (
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestSelectionBackground(t *testing.T) {
	for _, tt := range []struct {
		name       string
		background termenv.Color
		want       lipgloss.Color
	}{
		{"dark terminal", termenv.RGBColor("#20262c"), "#32373d"},
		{"light terminal", termenv.RGBColor("#e0e0e0"), "#cecece"},
		{"black", termenv.RGBColor("#000000"), "#141414"},
		{"white", termenv.RGBColor("#ffffff"), "#ebebeb"},
		{"ANSI fallback", termenv.ANSIColor(0), "#141414"},
		{"no color", termenv.NoColor{}, "#141414"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := selectionBackground(tt.background); got != tt.want {
				t.Errorf("highlight = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInitDerivesSelectionBackground(t *testing.T) {
	renderer := lipgloss.DefaultRenderer()
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(io.Discard))
	t.Cleanup(func() {
		lipgloss.SetDefaultRenderer(renderer)
		Init(Options{})
	})
	Init(Options{})
	if ColorHighlight != lipgloss.Color("#141414") {
		t.Errorf("highlight = %q, want neutral fallback #141414", ColorHighlight)
	}
	if Default.SelectedItem.GetBackground() != lipgloss.Color("#141414") {
		t.Error("selected row style does not use the derived highlight")
	}
}

func TestInitUsesTerminalPalette(t *testing.T) {
	Init(Options{})
	want := map[string]lipgloss.Color{
		"green": "2", "blue": "4", "red": "1", "yellow": "3", "cyan": "6",
		"magenta": "5", "white": "7", "gray": "8", "orange": "11",
	}
	got := map[string]lipgloss.Color{
		"green": ColorGreen, "blue": ColorBlue, "red": ColorRed, "yellow": ColorYellow, "cyan": ColorCyan,
		"magenta": ColorMagenta, "white": ColorWhite, "gray": ColorGray, "orange": ColorOrange,
	}
	for key, color := range want {
		if got[key] != color {
			t.Errorf("%s = %q, want %q", key, got[key], color)
		}
	}
	for _, color := range append(authorPalette, got["orange"]) {
		if n, err := strconv.Atoi(string(color)); err != nil || n < 0 || n > 15 {
			t.Errorf("color %q is not an ANSI 16 color", color)
		}
	}
}

func TestIgnoredThemeWarning(t *testing.T) {
	for _, name := range []string{"", "default", " Default "} {
		if got := IgnoredThemeWarning(name); got != "" {
			t.Errorf("IgnoredThemeWarning(%q) = %q, want no warning", name, got)
		}
	}
	for _, name := range []string{"auto", "catppuccin-mocha"} {
		got := IgnoredThemeWarning(name)
		if !strings.Contains(got, name) || !strings.Contains(got, "themeColors") {
			t.Errorf("IgnoredThemeWarning(%q) = %q, want the name and the themeColors hint", name, got)
		}
	}
}

func TestInitResetsAuthorCache(t *testing.T) {
	Init(Options{})
	_ = AuthorStyle("Alice")
	if len(authorCache) == 0 {
		t.Fatal("author cache should have an entry")
	}

	Init(Options{Colors: map[string]string{"green": "#abcdef"}})
	if len(authorCache) != 0 {
		t.Error("author cache should be empty after Init")
	}
	Init(Options{})
}

func TestInitAppliesSharedOverrides(t *testing.T) {
	Init(Options{
		Colors: map[string]string{
			"green":     "#abcdef",
			"highlight": "#123456",
			"bogus":     "ignored",
			"red":       "", // empty value must be skipped
		},
	})
	if Default.Colors.Green != lipgloss.Color("#abcdef") {
		t.Errorf("Green = %q, want #abcdef", Default.Colors.Green)
	}
	if ColorGreen != lipgloss.Color("#abcdef") {
		t.Errorf("ColorGreen not synced: %q", ColorGreen)
	}
	if Default.Colors.Highlight != lipgloss.Color("#123456") {
		t.Errorf("Highlight = %q, want #123456", Default.Colors.Highlight)
	}
	// Empty value must leave Red at the default.
	if Default.Colors.Red != lipgloss.Color("1") {
		t.Errorf("Red = %q, want default 1", Default.Colors.Red)
	}
	Init(Options{})
}

func TestInitAppliesOverridesMatchingTerminalBackground(t *testing.T) {
	t.Cleanup(func() {
		lipgloss.SetHasDarkBackground(true)
		Init(Options{})
	})
	opts := Options{
		ColorsDark:  map[string]string{"green": "#111111"},
		ColorsLight: map[string]string{"green": "#999999"},
	}

	lipgloss.SetHasDarkBackground(true)
	Init(opts)
	if Default.Colors.Green != lipgloss.Color("#111111") {
		t.Errorf("dark override not applied: Green = %q", Default.Colors.Green)
	}

	lipgloss.SetHasDarkBackground(false)
	Init(opts)
	if Default.Colors.Green != lipgloss.Color("#999999") {
		t.Errorf("light override not applied: Green = %q", Default.Colors.Green)
	}
}

func TestInitFallsBackOnInvalidColorValues(t *testing.T) {
	Init(Options{
		Colors: map[string]string{
			"green": "not-a-color",
			"blue":  "#zzzzzz",
			"red":   "totally bogus value",
		},
	})

	// Each invalid value falls back to lipgloss.Color("-1"), the terminal
	// default sentinel used elsewhere in the package (theme.go:39, 82).
	fallback := lipgloss.Color("-1")
	if Default.Colors.Green != fallback {
		t.Errorf("Green = %q, want %q (terminal default)", Default.Colors.Green, fallback)
	}
	if Default.Colors.Blue != fallback {
		t.Errorf("Blue = %q, want %q (terminal default)", Default.Colors.Blue, fallback)
	}
	if Default.Colors.Red != fallback {
		t.Errorf("Red = %q, want %q (terminal default)", Default.Colors.Red, fallback)
	}

	// Package-level color vars must be synced to the fallback too.
	if ColorGreen != fallback {
		t.Errorf("ColorGreen not synced: %q, want %q", ColorGreen, fallback)
	}
	if ColorBlue != fallback {
		t.Errorf("ColorBlue not synced: %q, want %q", ColorBlue, fallback)
	}
	if ColorRed != fallback {
		t.Errorf("ColorRed not synced: %q, want %q", ColorRed, fallback)
	}

	// Other keys must keep their defaults.
	if Default.Colors.Yellow != lipgloss.Color("3") {
		t.Errorf("Yellow = %q, want default 3", Default.Colors.Yellow)
	}

	// Rendering with the fallback color does not panic.
	_ = Default.Title.Render("smoke test")
	Init(Options{})
}

func TestValidColor(t *testing.T) {
	cases := []struct {
		val  string
		want bool
	}{
		// Valid hex (case-insensitive).
		{"#abc", true},
		{"#abcdef", true},
		{"#ABCDEF", true},
		{"#abcdef12", true},
		// Valid ANSI decimal and terminal-default sentinel.
		{"-1", true},
		{"0", true},
		{"15", true},
		{"208", true},
		{"255", true},
		// Invalid.
		{"", false},
		{"#", false},
		{"#xyz", false},
		{"#abcd", false},
		{"#abcde", false},
		{"#zzzzzz", false},
		{"-2", false},
		{"256", false},
		{"not-a-color", false},
		{"#abcdef ", false}, // trailing space — we do not trim
	}
	for _, c := range cases {
		t.Run(c.val, func(t *testing.T) {
			if got := ValidColor(c.val); got != c.want {
				t.Errorf("ValidColor(%q) = %v, want %v", c.val, got, c.want)
			}
		})
	}
}

func TestInitSelectsBorderShape(t *testing.T) {
	t.Cleanup(func() { Init(Options{}) })
	for _, tt := range []struct {
		borders string
		want    lipgloss.Border
	}{
		{"", lipgloss.RoundedBorder()},
		{"rounded", lipgloss.RoundedBorder()},
		{"sharp", lipgloss.NormalBorder()},
	} {
		Init(Options{Borders: tt.borders})
		if Default.Border != tt.want {
			t.Errorf("borders %q: Border = %+v, want %+v", tt.borders, Default.Border, tt.want)
		}
	}
}
