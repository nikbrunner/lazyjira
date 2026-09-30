package config

import "strings"

// SanitizeConfig lists text removed from copied values: yanked issue rows,
// copied summaries and Markdown links, and the `sanitize` custom command helper.
type SanitizeConfig struct {
	Remove []string `yaml:"remove"`
}

// Apply removes every configured string from s.
func (c SanitizeConfig) Apply(s string) string {
	for _, r := range c.Remove {
		s = strings.ReplaceAll(s, r, "")
	}
	return s
}
