package config

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func resolveCommand(t *testing.T, shell, command string) (ResolvedCustomCommand, error) {
	t.Helper()
	cfg := &Config{CustomCommands: []CustomCommandConfig{{Key: "x", Name: "test", Command: command}}}
	resolved, err := cfg.ResolveCustomCommandsForShell(shell)
	if err != nil {
		return ResolvedCustomCommand{}, err
	}
	return resolved[0], nil
}

func runShellOutput(t *testing.T, shell, command string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, shell, "-c", command).CombinedOutput()
	if err != nil {
		t.Fatalf("execute: %v; output: %s; command: %s", err, out, command)
	}
	return out
}

func TestResolveCustomCommandsRejectsUnsupportedInterpolation(t *testing.T) {
	t.Parallel()
	marker := filepath.Join(t.TempDir(), "executed")
	tests := []struct {
		name    string
		command string
	}{
		{"Go template range", `echo {{range .Values}}{{.}}{{end}}`},
		{"double quotes", `printf %s "{{.Value}}"`},
		{"single quotes", `printf %s '{{.Value}}'`},
		{"escaped placeholder", `printf %s \{{.Value}}`},
		{"static marker beside quoted placeholder", `printf %s __LAZYJIRA_VALUE_0__ '{{.Value}}'`},
		{"command name", `{{.Value}} --flag`},
		{"command substitution as command name", `printf %s $({{.Value}})`},
		{"backtick substitution", "printf %s `echo {{.Value}}`"},
		{"arithmetic", `printf %s $(({{.Value}}))`},
		{"heredoc", "cat <<EOF\n{{.Value}}\nEOF"},
		{"format function", `printf %s {{printf "%q" .Value}}`},
		{"format pipeline", `printf %s {{.Value | printf "%q"}}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cmd := fmt.Sprintf("touch %q; %s", marker, tc.command)
			if _, err := resolveCommand(t, "sh", cmd); err == nil {
				t.Fatalf("expected command %q to be rejected", cmd)
			}
			info, err := os.Stat(marker)
			if err == nil {
				t.Fatalf("marker exists after validation (%s)", info.Mode())
			}
			if !os.IsNotExist(err) {
				t.Fatalf("stat marker: %v", err)
			}
		})
	}
}

func TestEscapedPlaceholderCannotExecuteJiraValue(t *testing.T) {
	t.Parallel()
	shell, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh is not installed")
	}
	marker := filepath.Join(t.TempDir(), "executed")
	resolved, resolveErr := resolveCommand(t, shell, `printf %s \{{.Value}}`)
	if resolveErr == nil {
		rendered, renderErr := resolved.Render(map[string]string{"Value": "$(touch " + marker + ") #"})
		if renderErr == nil {
			_ = runShellOutput(t, shell, rendered)
		}
	}
	if _, statErr := os.Stat(marker); !os.IsNotExist(statErr) {
		t.Fatalf("injected command ran: %v", statErr)
	}
	if resolveErr == nil {
		t.Fatal("expected escaped placeholder to be rejected")
	}
}

func TestResolvedCustomCommandEscapesArguments(t *testing.T) {
	t.Parallel()
	for _, shellName := range []string{"sh", "zsh"} {
		t.Run(shellName, func(t *testing.T) {
			t.Parallel()
			shell, err := exec.LookPath(shellName)
			if err != nil {
				t.Skipf("%s is not installed", shellName)
			}
			marker := filepath.Join(t.TempDir(), "executed")
			value := "quotes ' \"; $(touch " + marker + "); `touch " + marker + "`\\\n雪"
			resolved, err := resolveCommand(t, shell, `printf '%s' {{.Value}}`)
			if err != nil {
				t.Fatalf("resolve: %v", err)
			}

			rendered, err := resolved.Render(map[string]string{"Value": value})
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			out := runShellOutput(t, shell, rendered)
			if string(out) != value {
				t.Fatalf("output = %q, want %q", out, value)
			}
			info, err := os.Stat(marker)
			if err == nil {
				t.Fatalf("injected command ran (%s)", info.Mode())
			}
			if !os.IsNotExist(err) {
				t.Fatalf("stat marker: %v", err)
			}
		})
	}
}

func TestResolvedCommandAcceptsUnicodeBeforeProtectedValue(t *testing.T) {
	t.Parallel()
	shell, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh is not installed")
	}
	tests := []struct {
		name, command, want string
		values              map[string]string
	}{
		{
			name:    "template source",
			command: `printf 'é:%s' {{.Value}}`,
			values:  map[string]string{"Value": "value"},
			want:    "é:value",
		},
		{
			name:    "shellraw expansion",
			command: `printf '%s:%s' {{.Prefix | shellraw}} {{.Value}}`,
			values:  map[string]string{"Prefix": "é", "Value": "value"},
			want:    "é:value",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resolved, err := resolveCommand(t, shell, tc.command)
			if err != nil {
				t.Fatalf("resolve: %v", err)
			}
			rendered, err := resolved.Render(tc.values)
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if got := string(runShellOutput(t, shell, rendered)); got != tc.want {
				t.Fatalf("output = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestResolvedCustomCommandEscapesEmptyValue(t *testing.T) {
	t.Parallel()
	shell, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh is not installed")
	}
	resolved, err := resolveCommand(t, shell, `set -- {{.Value}}; printf '%s:%s' "$#" "$1"`)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	rendered, err := resolved.Render(map[string]string{"Value": ""})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	out := runShellOutput(t, shell, rendered)
	if string(out) != "1:" {
		t.Fatalf("output = %q, want one empty argument", out)
	}
}

func TestResolvedCustomCommandShellrawIsExplicit(t *testing.T) {
	t.Parallel()
	shell, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh is not installed")
	}
	resolved, err := resolveCommand(t, shell, `printf '%s' {{.Value | shellraw}}`)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	rendered, err := resolved.Render(map[string]string{"Value": "$(printf raw)"})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	out := runShellOutput(t, shell, rendered)
	if string(out) != "raw" {
		t.Fatalf("output = %q, want raw", out)
	}
}

func TestResolvedCopyCommandBodiesRenderInCommandSubstitutions(t *testing.T) {
	t.Parallel()
	shell, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh is not installed")
	}
	marker := filepath.Join(t.TempDir(), "executed")
	key := "WEBSDK-42$(touch " + marker + ")"
	summary := "Fix [login] $(touch " + marker + ") 'quoted'"
	host := "jira.example$(touch " + marker + ")"
	summaryText := strings.NewReplacer("[", "", "]", "").Replace(summary)
	stubDir := t.TempDir()
	for name, script := range map[string]string{
		"pbcopy":  "#!/bin/sh\ncat\n",
		"wl-copy": "#!/bin/sh\ncat\n",
		"herdr":   "#!/bin/sh\nexit 0\n",
	} {
		path := filepath.Join(stubDir, name)
		if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
			t.Fatalf("write %s stub: %v", name, err)
		}
	}
	stubs := "PATH=" + shellescape(stubDir) + ":$PATH\n"
	tests := []struct {
		name, command, want string
	}{
		{
			name: "markdown link",
			command: `s=$(printf %s {{.Summary | shellescape}} | tr -d '[]')
link=$(printf '[%s %s](%s/browse/%s)' {{.Key | shellescape}} "$s" {{.JiraHost | shellescape}} {{.Key | shellescape}})
printf '%s' "$link" | (command -v pbcopy >/dev/null && pbcopy || wl-copy)
command -v herdr >/dev/null && herdr notification show "Copied markdown link" --body "$link" --sound done`,
			want: fmt.Sprintf("[%s %s](%s/browse/%s)", key, summaryText, host, key),
		},
		{
			name: "key and summary",
			command: `s=$(printf %s {{.Summary | shellescape}} | tr -d '[]')
out=$(printf '%s %s' {{.Key | shellescape}} "$s")
printf '%s' "$out" | (command -v pbcopy >/dev/null && pbcopy || wl-copy)
command -v herdr >/dev/null && herdr notification show "Copied key + summary" --body "$out" --sound done`,
			want: fmt.Sprintf("%s %s", key, summaryText),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{CustomCommands: []CustomCommandConfig{{Key: "ctrl+y", Name: tc.name, Command: tc.command}}}
			resolved, err := cfg.ResolveCustomCommandsForShell(shell)
			if err != nil {
				t.Fatalf("resolve: %v", err)
			}
			rendered, err := resolved[0].Render(map[string]string{
				"Key": key, "Summary": summary, "JiraHost": host,
			})
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if got := string(runShellOutput(t, shell, stubs+rendered)); got != tc.want {
				t.Fatalf("output = %q, want %q", got, tc.want)
			}
			info, statErr := os.Stat(marker)
			if statErr == nil {
				t.Fatalf("injected command ran (%s)", info.Mode())
			}
			if !os.IsNotExist(statErr) {
				t.Fatalf("stat marker: %v", statErr)
			}
		})
	}
}

func TestResolvedCustomCommandHelpersPreserveArguments(t *testing.T) {
	t.Parallel()
	shell, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh is not installed")
	}
	tests := []struct {
		name, command, value, want string
	}{
		{"shellescape", `printf '%s' {{.Value | shellescape}}`, "it's safe; $(false)", "it's safe; $(false)"},
		{"slugify", `printf '%s' {{.Value | slugify}}`, "Fix Login Bug #42!", "fix-login-bug-42"},
		{"empty slugify", `set -- {{.Value | slugify}}; printf '%s:%s' "$#" "$1"`, "!!!", "1:"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resolved, err := resolveCommand(t, shell, tc.command)
			if err != nil {
				t.Fatalf("resolve: %v", err)
			}
			rendered, err := resolved.Render(map[string]string{"Value": tc.value})
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			out := runShellOutput(t, shell, rendered)
			if string(out) != tc.want {
				t.Fatalf("output = %q, want %q", out, tc.want)
			}
		})
	}
}

func TestRenderedRawValueCannotChangeProtectedArgumentContext(t *testing.T) {
	t.Parallel()
	shell, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh is not installed")
	}
	marker := filepath.Join(t.TempDir(), "executed")
	resolved, err := resolveCommand(t, shell, `printf '%s\n' {{.Raw | shellraw}} {{.Value}}`)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	_, err = resolved.Render(map[string]string{
		"Raw":   "$(touch " + marker + ") #",
		"Value": "safe",
	})
	if err == nil {
		t.Fatal("expected raw expansion to invalidate the protected argument")
	}
	if _, statErr := os.Stat(marker); !os.IsNotExist(statErr) {
		t.Fatalf("marker exists after render rejection: %v", statErr)
	}
}

func TestResolveCustomCommandsSupportsIfBranches(t *testing.T) {
	t.Parallel()
	if _, err := resolveCommand(t, "sh", `printf %s {{if .Value}}{{.Value}}{{else}}empty{{end}}`); err != nil {
		t.Fatalf("resolve: %v", err)
	}
}

func TestResolveCustomCommandsRejectsUnknownShell(t *testing.T) {
	t.Parallel()
	if _, err := resolveCommand(t, "fish", `printf %s {{.Value}}`); err == nil {
		t.Fatal("expected unsupported shell to be rejected")
	}
}

func TestResolveCustomCommandsSupportsMatchedShellVariants(t *testing.T) {
	t.Parallel()
	for _, shell := range []string{"sh", "bash", "zsh", "mksh"} {
		t.Run(shell, func(t *testing.T) {
			t.Parallel()
			if _, err := resolveCommand(t, shell, `printf %s {{.Value}}`); err != nil {
				t.Fatalf("resolve: %v", err)
			}
		})
	}
}
