package config

import "strings"

// NormalizeKey returns the key notation the TUI matches for a key written in
// the config. It accepts "space" as another spelling of " ".
func NormalizeKey(key string) string {
	switch {
	case key == "space":
		return " "
	case key == "ctrl+space":
		return "ctrl+@"
	case strings.HasSuffix(key, "+space"):
		return strings.TrimSuffix(key, "space") + " "
	}
	return key
}
