package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSanitize_DefaultRemovesBrackets(t *testing.T) {
	t.Parallel()
	if got := DefaultConfig().Sanitize.Apply("[web-ui] Release [0.7.0]"); got != "web-ui Release 0.7.0" {
		t.Fatalf("Apply() = %q", got)
	}
}

func TestLoad_SanitizeRemoveFromYAML(t *testing.T) {
	for name, tc := range map[string]struct {
		yaml, want string
	}{
		"custom list replaces default": {yaml: "sanitize:\n  remove: [\"`\", \"[\"]\n", want: "a] b"},
		"empty list disables":          {yaml: "sanitize:\n  remove: []\n", want: "[a] `b`"},
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("LAZYJIRA_CONFIG_DIR", dir)
			if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(tc.yaml), 0o644); err != nil {
				t.Fatal(err)
			}

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if got := cfg.Sanitize.Apply("[a] `b`"); got != tc.want {
				t.Fatalf("Apply() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCustomCommand_SanitizeHelper(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.CustomCommands = []CustomCommandConfig{{Key: "y", Name: "copy", Command: "printf %s {{.Summary | sanitize}}"}}
	commands, err := cfg.ResolveCustomCommands()
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	got, err := commands[0].Render(map[string]string{"Summary": "[web-ui] it's done"})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if want := `printf %s 'web-ui it'\''s done'`; got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}
