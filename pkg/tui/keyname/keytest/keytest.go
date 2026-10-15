// Package keytest builds the key presses a terminal reports, for tests.
package keytest

import (
	"fmt"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
)

var named = map[string]string{
	"enter":     "\r",
	"tab":       "\t",
	"shift+tab": "\x1b[Z",
	"esc":       "\x1b",
	"backspace": "\x7f",
	"delete":    "\x1b[3~",
	"up":        "\x1b[A",
	"down":      "\x1b[B",
	"right":     "\x1b[C",
	"left":      "\x1b[D",
	"home":      "\x1b[H",
	"end":       "\x1b[F",
	"pgup":      "\x1b[5~",
	"pgdown":    "\x1b[6~",
	"f1":        "\x1bOP",
	"ctrl+@":    "\x00",
}

// Press returns the key press Bubble Tea delivers when a legacy terminal
// sends key. key uses the keyname notation: "j", "G", " ", "enter",
// "ctrl+u", "alt+y", "shift+tab".
func Press(key string) tea.KeyPressMsg {
	return Decode(encode(key))
}

// Decode returns the key press Bubble Tea delivers for the bytes a terminal
// sends, such as "\x1b[13;2u" for shift+enter with the kitty keyboard
// protocol.
func Decode(input string) tea.KeyPressMsg {
	var d uv.EventDecoder
	n, ev := d.Decode([]byte(input))
	k, ok := ev.(uv.KeyPressEvent)
	if !ok || n != len(input) {
		panic(fmt.Sprintf("keytest: %q does not decode to one key press", input))
	}
	return tea.KeyPressMsg(k)
}

// Type returns the key presses for typing text, one per character.
func Type(text string) []tea.KeyPressMsg {
	presses := make([]tea.KeyPressMsg, 0, len(text))
	for _, r := range text {
		presses = append(presses, Press(string(r)))
	}
	return presses
}

func encode(key string) string {
	if seq, ok := named[key]; ok {
		return seq
	}
	if utf8.RuneCountInString(key) == 1 {
		return key
	}
	if rest, ok := strings.CutPrefix(key, "alt+"); ok {
		return "\x1b" + encode(rest)
	}
	if rest, ok := strings.CutPrefix(key, "ctrl+"); ok && len(rest) == 1 && rest[0] >= 'a' && rest[0] <= 'z' {
		return string(rest[0] - 'a' + 1)
	}
	panic(fmt.Sprintf("keytest: unknown key %q", key))
}
