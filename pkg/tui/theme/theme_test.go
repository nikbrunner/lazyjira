package theme

import (
	"image/color"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestSelectionBackground(t *testing.T) {
	for _, tt := range []struct {
		name       string
		background color.Color
		want       string
	}{
		{"dark terminal", lipgloss.Color("#20262c"), "#32373d"},
		{"light terminal", lipgloss.Color("#e0e0e0"), "#cecece"},
		{"black", lipgloss.Color("#000000"), "#141414"},
		{"white", lipgloss.Color("#ffffff"), "#ebebeb"},
		{"ANSI fallback", ansi.IndexedColor(0), "#141414"},
		{"no answer", nil, "#141414"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := selectionBackground(tt.background); got != lipgloss.Color(tt.want) {
				t.Errorf("highlight = %v, want %s", got, tt.want)
			}
		})
	}
}

func TestInitDerivesSelectionBackground(t *testing.T) {
	t.Cleanup(func() { Init(Options{}) })
	Init(Options{})
	if ColorHighlight != lipgloss.Color("#141414") {
		t.Errorf("highlight = %v, want neutral fallback #141414", ColorHighlight)
	}
	if Default.SelectedItem.GetBackground() != lipgloss.Color("#141414") {
		t.Error("selected row style does not use the derived highlight")
	}
}

func TestInitUsesTerminalPalette(t *testing.T) {
	Init(Options{})
	want := map[string]color.Color{
		"green": lipgloss.Green, "blue": lipgloss.Blue, "red": lipgloss.Red, "yellow": lipgloss.Yellow, "cyan": lipgloss.Cyan,
		"magenta": lipgloss.Magenta, "white": lipgloss.White, "gray": lipgloss.BrightBlack, "orange": lipgloss.BrightYellow,
	}
	got := map[string]color.Color{
		"green": ColorGreen, "blue": ColorBlue, "red": ColorRed, "yellow": ColorYellow, "cyan": ColorCyan,
		"magenta": ColorMagenta, "white": ColorWhite, "gray": ColorGray, "orange": ColorOrange,
	}
	for key, c := range want {
		if got[key] != c {
			t.Errorf("%s = %v, want %v", key, got[key], c)
		}
	}
	for _, c := range append(authorPalette, got["orange"]) {
		if n, ok := c.(ansi.BasicColor); !ok || n > 15 {
			t.Errorf("color %v is not an ANSI 16 color", c)
		}
	}
}

func TestIgnoredThemeWarning(t *testing.T) {
	for _, name := range []string{"", "default", " Default "} {
		if got := IgnoredThemeWarning(name); got != "" {
			t.Errorf("IgnoredThemeWarning(%q) = %v, want no warning", name, got)
		}
	}
	for _, name := range []string{"auto", "catppuccin-mocha"} {
		got := IgnoredThemeWarning(name)
		if !strings.Contains(got, name) || !strings.Contains(got, "themeColors") {
			t.Errorf("IgnoredThemeWarning(%q) = %v, want the name and the themeColors hint", name, got)
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
		t.Errorf("Green = %v, want #abcdef", Default.Colors.Green)
	}
	if ColorGreen != lipgloss.Color("#abcdef") {
		t.Errorf("ColorGreen not synced: %v", ColorGreen)
	}
	if Default.Colors.Highlight != lipgloss.Color("#123456") {
		t.Errorf("Highlight = %v, want #123456", Default.Colors.Highlight)
	}
	// Empty value must leave Red at the default.
	if Default.Colors.Red != lipgloss.Color("1") {
		t.Errorf("Red = %v, want default 1", Default.Colors.Red)
	}
	Init(Options{})
}

func TestInitAppliesOverridesMatchingTerminalBackground(t *testing.T) {
	t.Cleanup(func() { Init(Options{}) })
	opts := Options{
		ColorsDark:  map[string]string{"green": "#111111"},
		ColorsLight: map[string]string{"green": "#999999"},
	}

	opts.Background = lipgloss.Color("#20262c")
	Init(opts)
	if Default.Colors.Green != lipgloss.Color("#111111") {
		t.Errorf("dark override not applied: Green = %q", Default.Colors.Green)
	}

	opts.Background = lipgloss.Color("#e0e0e0")
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

	// Each invalid value falls back to the terminal default color.
	fallback := color.Color(lipgloss.NoColor{})
	if Default.Colors.Green != fallback {
		t.Errorf("Green = %v, want %v (terminal default)", Default.Colors.Green, fallback)
	}
	if Default.Colors.Blue != fallback {
		t.Errorf("Blue = %v, want %v (terminal default)", Default.Colors.Blue, fallback)
	}
	if Default.Colors.Red != fallback {
		t.Errorf("Red = %v, want %v (terminal default)", Default.Colors.Red, fallback)
	}

	// Package-level color vars must be synced to the fallback too.
	if ColorGreen != fallback {
		t.Errorf("ColorGreen not synced: %v, want %v", ColorGreen, fallback)
	}
	if ColorBlue != fallback {
		t.Errorf("ColorBlue not synced: %v, want %v", ColorBlue, fallback)
	}
	if ColorRed != fallback {
		t.Errorf("ColorRed not synced: %v, want %v", ColorRed, fallback)
	}

	// Other keys must keep their defaults.
	if Default.Colors.Yellow != lipgloss.Color("3") {
		t.Errorf("Yellow = %v, want default 3", Default.Colors.Yellow)
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

func TestColorTerminalDefault(t *testing.T) {
	if got := Color("-1"); got != color.Color(lipgloss.NoColor{}) {
		t.Errorf(`Color("-1") = %v, want the terminal default color`, got)
	}
	if got := Color("4"); got != lipgloss.Color("4") {
		t.Errorf(`Color("4") = %v, want ANSI 4`, got)
	}
}

func TestInitDetectsDarkBackground(t *testing.T) {
	t.Cleanup(func() { Init(Options{}) })
	for _, tt := range []struct {
		background color.Color
		dark       bool
	}{
		{nil, true},
		{lipgloss.Color("#20262c"), true},
		{lipgloss.Color("#e0e0e0"), false},
	} {
		Init(Options{Background: tt.background})
		if IsDark() != tt.dark {
			t.Errorf("background %v: IsDark = %v, want %v", tt.background, IsDark(), tt.dark)
		}
	}
}

func TestColorFGBG(t *testing.T) {
	for _, tt := range []struct {
		value string
		want  color.Color
	}{
		{"15;0", ansi.IndexedColor(0)},
		{"0;default;15", ansi.IndexedColor(15)},
		{"", nil},
		{"15", nil},
		{"15;default", nil},
	} {
		if got := colorFGBG(tt.value); got != tt.want {
			t.Errorf("colorFGBG(%q) = %v, want %v", tt.value, got, tt.want)
		}
	}
}

func TestSetBackgroundSwitchesDarkToLight(t *testing.T) {
	t.Cleanup(func() { Init(Options{}) })
	Init(Options{Background: lipgloss.Color("#20262c")})
	held := Default

	if !SetBackground(lipgloss.Color("#e0e0e0")) {
		t.Fatal("dark to light should re-init the theme")
	}
	if IsDark() {
		t.Error("IsDark after switching to a light background")
	}
	if ColorHighlight != lipgloss.Color("#cecece") {
		t.Errorf("highlight = %v, want #cecece", ColorHighlight)
	}
	if held.SelectedItem.GetBackground() != lipgloss.Color("#cecece") {
		t.Error("a view holding Default still renders the old selection background")
	}
}

func TestSetBackgroundIgnoresTheSameColor(t *testing.T) {
	t.Cleanup(func() { Init(Options{}) })
	Init(Options{Background: lipgloss.Color("#20262c")})
	theme := *Default

	for _, same := range []color.Color{
		lipgloss.Color("#20262c"),
		ansi.XParseColor("rgb:2020/2626/2c2c"),
	} {
		if SetBackground(same) {
			t.Errorf("SetBackground(%v) re-initialized an unchanged theme", same)
		}
	}
	if SetBackground(nil) {
		t.Error("SetBackground(nil) re-initialized the theme")
	}
	if Default.SelectedItem.GetBackground() != theme.SelectedItem.GetBackground() {
		t.Error("selection background changed")
	}
}

func TestSetBackgroundKeepsOverrides(t *testing.T) {
	t.Cleanup(func() { Init(Options{}) })
	Init(Options{
		Colors:      map[string]string{"green": "#00ff00"},
		ColorsLight: map[string]string{"red": "#aa0000"},
		Borders:     "sharp",
		Background:  lipgloss.Color("#20262c"),
	})

	SetBackground(lipgloss.Color("#e0e0e0"))
	if ColorGreen != lipgloss.Color("#00ff00") || Default.Title.GetForeground() != lipgloss.Color("#00ff00") {
		t.Errorf("green = %v, want the themeColors override", ColorGreen)
	}
	if ColorRed != lipgloss.Color("#aa0000") {
		t.Errorf("red = %v, want the themeLight override", ColorRed)
	}
	if Default.Border != lipgloss.NormalBorder() {
		t.Error("border shape reset after the switch")
	}
}

func TestSetBackgroundFirstAnswerAfterFallback(t *testing.T) {
	t.Cleanup(func() { Init(Options{}) })
	Init(Options{})

	if SetBackground(lipgloss.Color("#20262c")) {
		t.Error("a dark answer after the black fallback re-initialized the theme")
	}
	if ColorHighlight != lipgloss.Color("#141414") {
		t.Errorf("highlight = %v, want the startup #141414", ColorHighlight)
	}
	if !SetBackground(lipgloss.Color("#30363c")) {
		t.Error("a later change of the dark background should re-init")
	}

	Init(Options{})
	if !SetBackground(lipgloss.Color("#e0e0e0")) {
		t.Error("a light answer after the black fallback should re-init")
	}

	Init(Options{Background: ansi.IndexedColor(0)})
	if SetBackground(lipgloss.Color("#20262c")) {
		t.Error("a dark answer after the COLORFGBG fallback re-initialized the theme")
	}
}

func TestSetBackgroundFirstAnswerAfterStartupAnswer(t *testing.T) {
	t.Cleanup(func() { Init(Options{}) })
	Init(Options{Background: lipgloss.Color("#20262c")})
	if !SetBackground(lipgloss.Color("#30363c")) {
		t.Error("another dark background after a real startup answer should re-init")
	}
}
