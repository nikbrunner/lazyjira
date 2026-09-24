package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nikbrunner/lazyjira/v2/pkg/config"
	"github.com/nikbrunner/lazyjira/v2/pkg/git"
	"github.com/nikbrunner/lazyjira/v2/pkg/jira"
	"github.com/nikbrunner/lazyjira/v2/pkg/tui/components"
)

//nolint:paralleltest // Replaces the process-wide clipboard command runner.
func TestBranchCopyAndCreateUseSameName(t *testing.T) {
	original := runExternalCommand
	t.Cleanup(func() { runExternalCommand = original })
	var copied string
	runExternalCommand = func(input string, _ bool, _ string, _ ...string) { copied = input }

	tests := []struct {
		name    string
		rules   []config.BranchFormatRule
		summary string
		want    string
	}{
		{name: "default", want: "websdk-218-web-ui-audit-skills"},
		{name: "UTF-8 truncation boundary", summary: strings.Repeat("a", 48) + "é", want: "websdk-218-" + strings.Repeat("a", 48)},
		{name: "type rule", rules: []config.BranchFormatRule{
			{When: config.BranchFormatCondition{Type: "Bug"}, Template: "fix/{{.Key}}"},
			{When: config.BranchFormatCondition{Type: "Story"}, Template: "story/{{.Key}}-{{.Summary}}"},
			{When: config.BranchFormatCondition{Type: "*"}, Template: "fallback/{{.Key}}"},
		}, want: "story/websdk-218-web-ui-audit-skills"},
		{name: "catch all", rules: []config.BranchFormatRule{
			{When: config.BranchFormatCondition{Type: "*"}, Template: "{{.ParentKey}}/{{.ProjectKey}}-{{.Number}}"},
		}, want: "websdk-100/websdk-218"},
		{name: "unmatched rule", rules: []config.BranchFormatRule{
			{When: config.BranchFormatCondition{Type: "Bug"}, Template: "fix/{{.Key}}"},
		}, want: "websdk-218-web-ui-audit-skills"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := focusApp(t)
			app.side = sideLeft
			app.leftFocus = focusIssues
			app.gitRepoPath = gitTestRepo(t)
			app.cfg.Git.BranchFormat = tt.rules
			app.issuesList.SetIssues([]jira.Issue{{Key: "OTHER-1", Summary: "List selection"}})
			app.previewKey = "WEBSDK-218"
			app.issueCache[app.previewKey] = &jira.Issue{
				Key: app.previewKey, Summary: "Web UI audit skills", IssueType: &jira.IssueType{Name: "Story"},
				Parent: &jira.Issue{Key: "WEBSDK-100"},
			}
			if tt.summary != "" {
				app.issueCache[app.previewKey].Summary = tt.summary
			}
			copied = ""

			_, _ = app.handleKeyMsg(runeKey('b'))
			if copied != tt.want {
				t.Errorf("clipboard = %q, want %q", copied, tt.want)
			}
			_, _ = app.handleKeyMsg(runeKey('B'))
			_, cmd := app.inputModal.Update(tea.KeyMsg{Type: tea.KeyEnter})
			if cmd == nil {
				t.Fatal("expected branch creation input")
			}
			name := cmd().(components.InputConfirmedMsg).Text
			if name != copied {
				t.Errorf("creation name = %q, clipboard = %q", name, copied)
			}
		})
	}
}

//nolint:paralleltest // Replaces the process-wide clipboard command runner.
func TestCopyWorktreeNameConfigAndContext(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LAZYJIRA_CONFIG_DIR", dir)
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(`git:
  worktreeFormat: "{{.Summary}}/{{.Key}}"
keybinding:
  issues:
    copyWorktreeName: "v"
`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	original := runExternalCommand
	t.Cleanup(func() { runExternalCommand = original })
	var copied string
	runExternalCommand = func(input string, _ bool, _ string, _ ...string) { copied = input }

	for _, context := range []string{"issues", "info", "detail", "projects", "status", "no issue", "bad template"} {
		t.Run(context, func(t *testing.T) {
			app := focusApp(t)
			cfgCopy := *cfg
			app.cfg = &cfgCopy
			app.keymap = KeymapFromConfig(cfg.Keybinding)
			app.side = sideLeft
			app.leftFocus = focusIssues
			app.issuesList.SetIssues([]jira.Issue{{Key: "WEBSDK-218", Summary: "List title"}})
			app.previewKey = "WEBSDK-9"
			app.issueCache[app.previewKey] = &jira.Issue{Key: app.previewKey, Summary: "Preview title"}
			want := "preview-title/websdk-9"
			switch context {
			case "info":
				app.leftFocus = focusInfo
			case "detail":
				app.side = sideRight
			case "projects":
				app.leftFocus = focusProjects
				want = ""
			case "status":
				app.leftFocus = focusProjects
				want = ""
			case "no issue":
				app.previewKey = ""
				app.issuesList.SetIssues(nil)
				want = ""
			case "bad template":
				app.cfg.Git.WorktreeFormat = "{{.Unknown}}"
				want = ""
			}
			copied = ""

			_, _ = app.handleKeyMsg(runeKey('v'))

			if copied != want {
				t.Errorf("clipboard = %q, want %q", copied, want)
			}
			if context == "bad template" && app.statusPanel.ErrorMessage() == "" {
				t.Error("invalid template must surface an error")
			}
		})
	}
}

//nolint:paralleltest // Replaces the process-wide clipboard command runner.
func TestCopyWorktreeNameKey(t *testing.T) {
	original := runExternalCommand
	t.Cleanup(func() { runExternalCommand = original })
	var copied string
	runExternalCommand = func(input string, _ bool, _ string, _ ...string) {
		copied = input
	}

	for _, location := range []string{"main", "nested", "worktree", "outside git"} {
		t.Run(location, func(t *testing.T) {
			repo := filepath.Join(t.TempDir(), "web-ui")
			if err := os.Rename(gitTestRepo(t), repo); err != nil {
				t.Fatal(err)
			}
			dir := repo
			want := "web-ui-websdk-218-web-ui-audit-skills-with-skill-c"
			switch location {
			case "nested":
				dir = filepath.Join(repo, "src")
				if err := os.Mkdir(dir, 0o755); err != nil {
					t.Fatal(err)
				}
			case "worktree":
				dir = filepath.Join(filepath.Dir(repo), "websdk-218-sibling")
				cmd := exec.CommandContext(t.Context(), "git", "-C", repo, "worktree", "add", "--detach", dir)
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("add worktree: %v: %s", err, out)
				}
			case "outside git":
				dir = ""
				want = "websdk-218-web-ui-audit-skills-with-skill-creator"
			}
			app := focusApp(t)
			app.gitRepoPath = dir
			app.issuesList.SetIssues([]jira.Issue{{Key: "WEBSDK-218", Summary: "Web UI audit skills with skill creator"}})
			copied = ""

			_, _ = app.handleKeyMsg(runeKey('w'))

			if copied != want {
				t.Errorf("clipboard = %q, want %q", copied, want)
			}
			if app.inputModal.IsVisible() {
				t.Error("copy must not open branch creation")
			}
			if !strings.Contains(app.helpBar.View(), "Copied worktree name") {
				t.Error("expected copy confirmation")
			}
			branch, err := git.CurrentBranch(repo)
			if err != nil || branch != "main" {
				t.Errorf("copy changed current branch: %q, %v", branch, err)
			}
		})
	}
}
