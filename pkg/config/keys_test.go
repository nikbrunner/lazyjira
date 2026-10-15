package config

import "testing"

func TestNormalizeKey(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"space":      " ",
		" ":          " ",
		"alt+space":  "alt+ ",
		"ctrl+space": "ctrl+@",
		"ctrl+@":     "ctrl+@",
		"j":          "j",
		"gg":         "gg",
		"alt+y":      "alt+y",
		"enter":      "enter",
		"backspace":  "backspace",
	}
	for in, want := range tests {
		if got := NormalizeKey(in); got != want {
			t.Errorf("NormalizeKey(%q) = %q, want %q", in, got, want)
		}
	}
}
