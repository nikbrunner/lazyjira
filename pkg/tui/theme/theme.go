package theme

import (
	"fmt"
	"image/color"
	"log/slog"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lucasb-eyer/go-colorful"
)

// ColorPalette holds the semantic color values for a theme. The defaults are
// ANSI 0-15 codes so the terminal's own color scheme decides how they look.
type ColorPalette struct {
	Green     color.Color
	Blue      color.Color
	Red       color.Color
	Yellow    color.Color
	Cyan      color.Color
	Magenta   color.Color
	White     color.Color
	Gray      color.Color
	Orange    color.Color
	None      color.Color
	Highlight color.Color // selection/cursor background
}

// Package-level color variables. These are kept in sync with Default.Colors
// by Init so that existing call sites (theme.ColorBlue, etc.) continue
// to work without changes.
var (
	ColorGreen                 = lipgloss.Color("2")  // ANSI green — active borders, accents
	ColorBlue                  = lipgloss.Color("4")  // ANSI blue — help bar, selected bg
	ColorRed                   = lipgloss.Color("1")  // ANSI red — errors, unstaged
	ColorYellow                = lipgloss.Color("3")  // ANSI yellow — warnings, in-progress
	ColorCyan                  = lipgloss.Color("6")  // ANSI cyan — search mode
	ColorMagenta               = lipgloss.Color("5")  // ANSI magenta — JQL keywords
	ColorWhite                 = lipgloss.Color("7")  // ANSI white (light gray)
	ColorGray                  = lipgloss.Color("8")  // ANSI bright black (dark gray)
	ColorOrange                = lipgloss.Color("11") // ANSI bright yellow — secondary accent (names, metadata)
	ColorNone      color.Color = lipgloss.NoColor{}   // default terminal color
	ColorHighlight             = lipgloss.Color("4")  // selection/cursor background (same as blue by default)
)

type Theme struct {
	Title          lipgloss.Style
	Subtitle       lipgloss.Style
	HintBar        lipgloss.Style
	SelectedItem   lipgloss.Style
	NormalItem     lipgloss.Style
	ActiveBorder   lipgloss.Style
	InactiveBorder lipgloss.Style
	ErrorText      lipgloss.Style
	SuccessText    lipgloss.Style
	WarningText    lipgloss.Style
	KeyStyle       lipgloss.Style
	ValueStyle     lipgloss.Style
	PriorityHigh   lipgloss.Style
	PriorityMedium lipgloss.Style
	PriorityLow    lipgloss.Style

	Colors        ColorPalette
	AuthorPalette []color.Color
	Border        lipgloss.Border
}

// Default is the singleton theme instance
var Default = defaultTheme()

// DefaultTheme returns the singleton theme. Kept for compatibility
func DefaultTheme() *Theme { return Default }

// defaultPalette returns the ANSI 16 color palette.
func defaultPalette() ColorPalette {
	return ColorPalette{
		Green:     lipgloss.Color("2"),
		Blue:      lipgloss.Color("4"),
		Red:       lipgloss.Color("1"),
		Yellow:    lipgloss.Color("3"),
		Cyan:      lipgloss.Color("6"),
		Magenta:   lipgloss.Color("5"),
		White:     lipgloss.Color("7"),
		Gray:      lipgloss.Color("8"),
		Orange:    lipgloss.Color("11"),
		None:      lipgloss.NoColor{},
		Highlight: lipgloss.Color("4"), // same as blue for default theme
	}
}

// defaultAuthorPalette returns the ANSI colors that tell comment authors apart.
func defaultAuthorPalette() []color.Color {
	codes := []string{
		"1", "2", "3", "4", "5", "6",
		"9", "10", "11", "12", "13", "14",
	}
	palette := make([]color.Color, 0, len(codes))
	for _, code := range codes {
		palette = append(palette, lipgloss.Color(code))
	}
	return palette
}

func defaultTheme() *Theme {
	return buildTheme(defaultPalette(), defaultAuthorPalette())
}

// buildTheme constructs a Theme from a color palette and author palette.
func buildTheme(p ColorPalette, authors []color.Color) *Theme {
	return &Theme{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.Green),

		Subtitle: lipgloss.NewStyle().
			Foreground(p.Gray),

		HintBar: lipgloss.NewStyle().
			Foreground(p.Gray),

		SelectedItem: lipgloss.NewStyle().
			Bold(true).
			Background(p.Highlight),

		NormalItem: lipgloss.NewStyle(),

		ActiveBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.Green),

		InactiveBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.None),

		ErrorText: lipgloss.NewStyle().
			Foreground(p.Red).
			Bold(true),

		SuccessText: lipgloss.NewStyle().
			Foreground(p.Green),

		WarningText: lipgloss.NewStyle().
			Foreground(p.Yellow),

		KeyStyle: lipgloss.NewStyle().
			Foreground(p.Green),

		ValueStyle: lipgloss.NewStyle(),

		PriorityHigh: lipgloss.NewStyle().
			Foreground(p.Red),

		PriorityMedium: lipgloss.NewStyle().
			Foreground(p.Yellow),

		PriorityLow: lipgloss.NewStyle().
			Foreground(p.Green),

		Colors:        p,
		AuthorPalette: authors,
		Border:        lipgloss.RoundedBorder(),
	}
}

// syncColors updates the package-level color variables and the author palette
// to match the current Default theme.
func syncColors() {
	ColorGreen = Default.Colors.Green
	ColorBlue = Default.Colors.Blue
	ColorRed = Default.Colors.Red
	ColorYellow = Default.Colors.Yellow
	ColorCyan = Default.Colors.Cyan
	ColorMagenta = Default.Colors.Magenta
	ColorWhite = Default.Colors.White
	ColorGray = Default.Colors.Gray
	ColorOrange = Default.Colors.Orange
	ColorNone = Default.Colors.None
	ColorHighlight = Default.Colors.Highlight

	authorMutex.Lock()
	authorPalette = Default.AuthorPalette
	authorCache = make(map[string]lipgloss.Style)
	authorMutex.Unlock()
}

// Options configures the theme system. Values map 1:1 to the GUIConfig
// fields in pkg/config; the indirection keeps the theme package free of
// any config-package dependency.
//
// Precedence (low to high):
//  1. The ANSI 16 palette.
//  2. Colors (shared overrides).
//  3. ColorsDark or ColorsLight, whichever matches the terminal background.
//
// Override map keys are lowercase palette field names: "green", "blue",
// "red", "yellow", "cyan", "magenta", "white", "gray", "orange", "highlight".
// Unknown keys and empty values are silently ignored so configs stay
// forward-compatible.
//
// Borders selects the border shape: "sharp" uses square corners, any other
// value keeps the rounded default.
//
// Background is the terminal background from TerminalBackground. It decides
// between dark and light and is the basis of the selection highlight. nil
// counts as black.
type Options struct {
	Colors      map[string]string
	ColorsDark  map[string]string
	ColorsLight map[string]string
	Borders     string
	Background  color.Color
}

// dark reports whether the terminal background Init received is dark.
var dark = true

// current holds the options of the last Init, so SetBackground can rebuild
// the theme with a new background and the same overrides.
var current Options

// IsDark reports whether the terminal background is dark.
func IsDark() bool { return dark }

// TerminalBackground asks the terminal for its background color, falling
// back to the COLORFGBG environment variable. It returns nil when neither
// answers, and on platforms without the query. Call it before the TUI owns
// terminal input.
func TerminalBackground() color.Color {
	return terminalBackground()
}

// colorFGBG returns the background from a COLORFGBG value such as "15;0".
func colorFGBG(fgbg string) color.Color {
	if !strings.Contains(fgbg, ";") {
		return nil
	}
	parts := strings.Split(fgbg, ";")
	if i, err := strconv.Atoi(parts[len(parts)-1]); err == nil && i >= 0 && i <= 255 {
		return ansi.IndexedColor(i)
	}
	return nil
}

// Init applies any user overrides to the ANSI palette and refreshes the
// global Default theme plus the package-level color variables. Must be
// called before the TUI starts. Default is updated in place, so holders of
// the pointer see the new styles.
func Init(opts Options) {
	current = opts
	palette := defaultPalette()
	palette.Highlight = selectionBackground(opts.Background)
	palette = applyOverrides(palette, opts.Colors, "themeColors")
	dark = isDark(opts.Background)
	if dark {
		palette = applyOverrides(palette, opts.ColorsDark, "themeDark")
	} else {
		palette = applyOverrides(palette, opts.ColorsLight, "themeLight")
	}

	*Default = *buildTheme(palette, defaultAuthorPalette())
	if opts.Borders == "sharp" {
		Default.Border = lipgloss.NormalBorder()
	}
	syncColors()
}

// SetBackground rebuilds the theme for a terminal background reported
// while the TUI runs, keeping the overrides of the last Init. It reports
// whether the theme changed: a switch between dark and light, or a new
// selection highlight. While the background is still a startup fallback
// (no answer, or COLORFGBG), only a dark/light switch counts, so the first
// real answer does not shift the highlight of an unchanged terminal. nil is
// ignored.
func SetBackground(bg color.Color) bool {
	if bg == nil {
		return false
	}
	changed := isDark(bg) != dark ||
		!isFallback(current.Background) && selectionHex(bg) != selectionHex(current.Background)
	if !changed {
		current.Background = bg
		return false
	}
	opts := current
	opts.Background = bg
	Init(opts)
	return true
}

func isFallback(background color.Color) bool {
	_, indexed := background.(ansi.IndexedColor)
	return background == nil || indexed
}

// IgnoredThemeWarning returns the startup warning for a gui.theme value, or
// "" when the value selects the terminal palette.
func IgnoredThemeWarning(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "default":
		return ""
	}
	return fmt.Sprintf("gui.theme %q is ignored: colors come from the terminal; use gui.themeColors to override them", name)
}

func toRGB(background color.Color) colorful.Color {
	if background == nil {
		return colorful.Color{}
	}
	c, _ := colorful.MakeColor(background)
	return c
}

func isDark(background color.Color) bool {
	_, _, lightness := toRGB(background).Hsl()
	return lightness < 0.5
}

func selectionBackground(background color.Color) color.Color {
	return lipgloss.Color(selectionHex(background))
}

func selectionHex(background color.Color) string {
	rgb := toRGB(background)
	_, _, lightness := rgb.Hsl()
	target := 1.0
	if lightness >= 0.5 {
		target = 0
	}
	rgb.R += (target - rgb.R) * 0.08
	rgb.G += (target - rgb.G) * 0.08
	rgb.B += (target - rgb.B) * 0.08
	return rgb.Hex()
}

// Color returns the color for a config value that ValidColor accepts. "-1"
// is the terminal's default color.
func Color(val string) color.Color {
	if val == "-1" {
		return lipgloss.NoColor{}
	}
	return lipgloss.Color(val)
}

// ValidColor reports whether val is a color string lipgloss will
// render correctly. Accepted forms:
//   - hex: "#rgb", "#rrggbb", "#rrggbbaa" (case-insensitive)
//   - ANSI decimal: "0".."255"
//   - terminal default sentinel: "-1"
//
// Empty strings are rejected here; callers (e.g. applyOverrides) skip
// empty values before calling ValidColor so users can use "" to mean
// "keep the default".
//
// Exported so other packages (e.g. pkg/tui/views/adf.go) can guard
// dynamic color strings from untrusted sources (Jira ADF marks, etc.)
// against malformed values.
func ValidColor(val string) bool {
	if val == "" {
		return false
	}
	if strings.HasPrefix(val, "#") {
		hex := val[1:]
		switch len(hex) {
		case 3, 6, 8:
		default:
			return false
		}
		for i := range len(hex) {
			c := hex[i]
			switch {
			case c >= '0' && c <= '9':
			case c >= 'a' && c <= 'f':
			case c >= 'A' && c <= 'F':
			default:
				return false
			}
		}
		return true
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return false
	}
	return n == -1 || (n >= 0 && n <= 255)
}

// applyOverrides patches a ColorPalette with any non-empty values from the
// provided map. Unknown keys are silently ignored for forward-compat. Values
// that are not a recognized color format fall back to the terminal default
// color and emit a slog.Warn.
//
// scope identifies the source map ("themeColors", "themeDark",
// "themeLight") so log lines can pinpoint the bad entry.
func applyOverrides(p ColorPalette, m map[string]string, scope string) ColorPalette {
	for key, val := range m {
		if val == "" {
			continue
		}
		c := Color(val)
		if !ValidColor(val) {
			slog.Warn("ignoring invalid theme color; falling back to terminal default",
				"scope", scope, "key", key, "value", val)
			c = lipgloss.NoColor{}
		}
		switch strings.ToLower(key) {
		case "green":
			p.Green = c
		case "blue":
			p.Blue = c
		case "red":
			p.Red = c
		case "yellow":
			p.Yellow = c
		case "cyan":
			p.Cyan = c
		case "magenta":
			p.Magenta = c
		case "white":
			p.White = c
		case "gray":
			p.Gray = c
		case "orange":
			p.Orange = c
		case "highlight":
			p.Highlight = c
		}
	}
	return p
}

// PriorityStyled applies priority color based on name
func PriorityStyled(name string) string {
	switch strings.ToLower(name) {
	case "highest", "high", "critical", "blocker":
		return Default.PriorityHigh.Render(name)
	case "medium":
		return Default.PriorityMedium.Render(name)
	default:
		return Default.PriorityLow.Render(name)
	}
}

func StatusColor(categoryKey string) lipgloss.Style {
	switch categoryKey {
	case "done":
		return lipgloss.NewStyle().Foreground(ColorGreen)
	case "indeterminate":
		return lipgloss.NewStyle().Foreground(ColorYellow)
	case "new":
		return lipgloss.NewStyle().Foreground(ColorBlue)
	default:
		return lipgloss.NewStyle().Foreground(ColorGray)
	}
}
