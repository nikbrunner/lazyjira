package keyname

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"

	"github.com/nikbrunner/lazyjira/pkg/config"
)

// decode returns the key press Bubble Tea delivers for the bytes a terminal sends.
func decode(t *testing.T, input string) tea.KeyPressMsg {
	t.Helper()
	var d uv.EventDecoder
	n, ev := d.Decode([]byte(input))
	if n != len(input) {
		t.Fatalf("decode(%q) consumed %d of %d bytes", input, n, len(input))
	}
	k, ok := ev.(uv.KeyPressEvent)
	if !ok {
		t.Fatalf("decode(%q) = %T, want a key press", input, ev)
	}
	return tea.KeyPressMsg(k)
}

func TestOf_LegacyTerminal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"j", "j"},
		{"G", "G"},
		{"@", "@"},
		{"#", "#"},
		{"ü", "ü"},
		{" ", " "},
		{"\x00", "ctrl+@"},
		{"\x15", "ctrl+u"},
		{"\x19", "ctrl+y"},
		{"\x0a", "ctrl+j"},
		{"\x08", "ctrl+h"},
		{"\x1by", "alt+y"},
		{"\x1bY", "alt+Y"},
		{"\x1b ", "alt+ "},
		{"\x1b\x19", "alt+ctrl+y"},
		{"\r", "enter"},
		{"\t", "tab"},
		{"\x1b[Z", "shift+tab"},
		{"\x7f", "backspace"},
		{"\x1b", "esc"},
		{"\x1b[A", "up"},
		{"\x1b[B", "down"},
		{"\x1b[1;2A", "shift+up"},
		{"\x1b[H", "home"},
		{"\x1b[F", "end"},
		{"\x1b[3~", "delete"},
		{"\x1b[5~", "pgup"},
		{"\x1bOP", "f1"},
	}
	for _, tt := range tests {
		if got := Of(decode(t, tt.input)); got != tt.want {
			t.Errorf("Of(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// A terminal with the kitty keyboard protocol reports the same key presses
// with more detail. Each one must keep its legacy notation.
func TestOf_KittyKeyboardMatchesLegacy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		kitty  string
		legacy string
	}{
		{"\x1b[121;3u", "\x1by"},    // alt+y
		{"\x1b[121;4u", "\x1bY"},    // alt+shift+y
		{"\x1b[121;7u", "\x1b\x19"}, // ctrl+alt+y
		{"\x1b[117;5u", "\x15"},     // ctrl+u
		{"\x1b[117;6u", "\x15"},     // ctrl+shift+u
		{"\x1b[106;5u", "\x0a"},     // ctrl+j
		{"\x1b[105;5u", "\t"},       // ctrl+i
		{"\x1b[109;5u", "\r"},       // ctrl+m
		{"\x1b[32;5u", "\x00"},      // ctrl+space
		{"\x1b[32;2u", " "},         // shift+space
		{"\x1b[13u", "\r"},          // enter
		{"\x1b[13;2u", "\r"},        // shift+enter
		{"\x1b[27u", "\x1b"},        // esc
		{"\x1b[127u", "\x7f"},       // backspace
		{"\x1b[127;2u", "\x7f"},     // shift+backspace
		{"\x1b[9;2u", "\x1b[Z"},     // shift+tab
		{"\x1b[27;5;117~", "\x15"},  // ctrl+u via modifyOtherKeys
		{"\x1b[27;3;121~", "\x1by"}, // alt+y via modifyOtherKeys
		{"\x1b[27;2;13~", "\r"},     // shift+enter via modifyOtherKeys
		{"\x1b[127;1:1u", "\x7f"},   // backspace press event
		{"\x1b[103;1u", "g"},
		{"\x1b[103;2u", "G"},
		{"\x1b[27;2;71~", "G"},
		{"\x1b[27;2;32~", " "},         // g
		{"\x1b[1;2A", "\x1b[1;2A"},     // shift+up
		{"\x1b[121;67u", "\x1by"},      // alt+y with caps lock
		{"\x1b[32;129u", " "},          // space with num lock
		{"\x1b[117;133u", "\x15"},      // ctrl+u with num lock
		{"\x1b[57399u", "\x1b[57399u"}, // keypad 0 stays as reported
	}
	for _, tt := range tests {
		got, want := Of(decode(t, tt.kitty)), Of(decode(t, tt.legacy))
		if got != want {
			t.Errorf("Of(%q) = %q, want %q like legacy %q", tt.kitty, got, want, tt.legacy)
		}
	}
}

func TestText(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input  string
		want   string
		isText bool
	}{
		{"a", "a", true},
		{"G", "G", true},
		{"@", "@", true},
		{"ü", "ü", true},
		{" ", " ", false},
		{"\x1by", "", false},
		{"\x15", "", false},
		{"\r", "", false},
		{"\x1b[A", "", false},
	}
	for _, tt := range tests {
		msg := decode(t, tt.input)
		if got := Text(msg); got != tt.want {
			t.Errorf("Text(%q) = %q, want %q", tt.input, got, tt.want)
		}
		if got := IsText(msg); got != tt.isText {
			t.Errorf("IsText(%q) = %v, want %v", tt.input, got, tt.isText)
		}
	}
}

// Folding keeps the key press a component matches with its own keymap the
// same in every terminal.
func TestFold(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		input  string
		legacy string
	}{
		{"\x1b[13;2u", "\r"},
		{"\x1b[27;2;13~", "\r"},
		{"\x1b[127;2u", "\x7f"},
		{"\x1b[32;2u", " "},
		{"\x1b[27;2;32~", " "},
		{"\x1b[105;5u", "\t"},
		{"\x1b[117;6u", "\x15"},
	} {
		got, want := Fold(decode(t, tt.input)), Fold(decode(t, tt.legacy))
		if got.String() != want.String() || Text(got) != Text(want) {
			t.Errorf("Fold(%q) = %q/%q, want %q/%q", tt.input, got.String(), Text(got), want.String(), Text(want))
		}
		if again := Fold(got); again != got {
			t.Errorf("Fold(%q) is not stable: %+v then %+v", tt.input, got, again)
		}
	}
}

func TestWithoutAlt(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		input string
		key   string
		text  string
	}{
		{"\x1b\x7f", "backspace", ""},
		{"\x1b\r", "enter", ""},
		{"\x1b[127;3u", "backspace", ""},
		{"\x1b[13;3u", "enter", ""},
		{"\x1bx", "x", "x"},
		{"\x1bX", "X", "X"},
		{"\x1b ", " ", " "},
		{"\x1b\x15", "ctrl+u", ""},
		{"\x1b[1;3D", "left", ""},
		{"x", "x", "x"},
	} {
		got := WithoutAlt(decode(t, tt.input))
		if Of(got) != tt.key || Text(got) != tt.text {
			t.Errorf("WithoutAlt(%q) = %q/%q, want %q/%q", tt.input, Of(got), Text(got), tt.key, tt.text)
		}
	}
}

// A config key written either way matches the key press Of reports.
func TestConfigKeysMatchOf(t *testing.T) {
	t.Parallel()
	for key, input := range map[string]string{
		"space":      " ",
		" ":          " ",
		"alt+space":  "\x1b ",
		"ctrl+space": "\x00",
	} {
		if got, want := config.NormalizeKey(key), Of(decode(t, input)); got != want {
			t.Errorf("NormalizeKey(%q) = %q, want %q", key, got, want)
		}
	}
}
