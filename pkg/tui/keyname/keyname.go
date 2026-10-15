// Package keyname turns Bubble Tea key presses into the key notation that the
// keymap and the user config use.
//
// The notation is the one legacy terminals produce: "j", "G", " " for space,
// "ctrl+u", "alt+y", "alt+Y", "shift+tab", "enter". Bubble Tea also turns on
// the kitty keyboard protocol where the terminal supports it, which reports
// distinctions a legacy terminal cannot send, such as shift+enter or ctrl+i
// apart from tab. Of folds those back so a key means the same thing in every
// terminal.
package keyname

import (
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
)

// Space is the notation for the space bar.
const Space = " "

const lockMods = tea.ModCapsLock | tea.ModNumLock | tea.ModScrollLock

// Fold returns msg without the detail only some terminals report: lock
// modifiers, shift on enter, backspace, esc, and space, shift on ctrl and alt
// letter combinations, and ctrl+i, ctrl+m, and ctrl+[ apart from tab, enter,
// and esc. Bubble Tea components that match keys themselves, such as the
// textarea, then see a key press the same way in every terminal.
func Fold(msg tea.KeyPressMsg) tea.KeyPressMsg {
	k := msg.Key()
	k.Mod &^= lockMods
	k.BaseCode = 0

	switch k.Code {
	case tea.KeyEnter, tea.KeyBackspace, tea.KeyEscape, tea.KeySpace:
		k.Mod &^= tea.ModShift
	}
	if k.Mod.Contains(tea.ModShift) && k.Mod&(tea.ModCtrl|tea.ModAlt) != 0 && isASCIILetter(k.Code) {
		k.Mod &^= tea.ModShift
		k.ShiftedCode = 0
		if !k.Mod.Contains(tea.ModCtrl) {
			k.Code = unicode.ToUpper(k.Code)
		}
	}

	if k.Mod.Contains(tea.ModCtrl) {
		switch k.Code {
		case 'i':
			k.Code, k.Mod = tea.KeyTab, k.Mod&^tea.ModCtrl
		case 'm':
			k.Code, k.Mod = tea.KeyEnter, k.Mod&^tea.ModCtrl
		case '[':
			k.Code, k.Mod = tea.KeyEscape, k.Mod&^tea.ModCtrl
		case tea.KeySpace:
			k.Code = '@'
		}
	}

	switch {
	case k.Mod&(tea.ModCtrl|tea.ModAlt) != 0:
		k.Text = ""
	case k.Code == tea.KeySpace:
		k.Text = Space
	}
	return tea.KeyPressMsg(k)
}

// Of returns the notation for a key press.
func Of(msg tea.KeyPressMsg) string {
	k := Fold(msg)
	s := k.String()
	if k.Code == tea.KeySpace {
		s = strings.TrimSuffix(k.Keystroke(), "space") + Space
	}
	if k.Mod.Contains(tea.ModCtrl | tea.ModAlt) {
		s = "alt+" + strings.Replace(s, "ctrl+alt+", "ctrl+", 1)
	}
	return s
}

// WithoutAlt returns msg folded and with alt released. Text inputs read keys
// this way: alt+backspace deletes like backspace, alt+enter confirms like
// enter, and alt+x types x.
func WithoutAlt(msg tea.KeyPressMsg) tea.KeyPressMsg {
	k := Fold(msg)
	if !k.Mod.Contains(tea.ModAlt) {
		return k
	}
	k.Mod &^= tea.ModAlt
	if !k.Mod.Contains(tea.ModCtrl) && unicode.IsPrint(k.Code) {
		k.Text = string(k.Code)
	}
	return k
}

// Text returns the characters a key press types, or "" when the press is a
// command such as enter, an arrow, or a ctrl or alt combination.
func Text(msg tea.KeyPressMsg) string {
	if msg.Mod&(tea.ModCtrl|tea.ModAlt|tea.ModMeta|tea.ModSuper|tea.ModHyper) != 0 {
		return ""
	}
	return msg.Text
}

// IsText reports whether a key press types text other than a space. Only
// such a key can start a key sequence such as "gg".
func IsText(msg tea.KeyPressMsg) bool {
	t := Text(msg)
	return t != "" && t != Space
}

func isASCIILetter(r rune) bool {
	return r >= 'a' && r <= 'z'
}
