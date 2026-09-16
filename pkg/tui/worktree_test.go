package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/git"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/tui/components"
)

const worktreeMainBranch = "main"

func TestCreateWorktreeGuards(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"no repo", "no issue", "projects panel", "invalid format"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			app := focusApp(t)
			app.side = sideLeft
			app.leftFocus = focusIssues
			app.gitRepoPath = gitTestRepo(t)
			app.issuesList.SetIssues([]jira.Issue{{Key: "PROJ-1", Summary: "Task"}})
			switch scenario {
			case "no repo":
				app.gitRepoPath = ""
			case "no issue":
				app.issuesList.SetIssues(nil)
			case "projects panel":
				app.leftFocus = focusProjects
			case "invalid format":
				app.cfg.Git.WorktreeFormat = "{{.Unknown}}"
			}
			_, cmd := app.handleKeyMsg(runeKey('W'))
			if cmd != nil || app.inputModal.IsVisible() {
				t.Error("invalid context must not start creation")
			}
			if (scenario == "no repo" || scenario == "invalid format") && app.statusPanel.ErrorMessage() == "" {
				t.Error("expected a visible error")
			}
		})
	}
}

func TestCreateWorktreeCancelAndFailure(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"cancel", "occupied branch", "occupied path"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			app := focusApp(t)
			app.side = sideLeft
			app.leftFocus = focusIssues
			app.gitRepoPath = gitTestRepo(t)
			app.gitBranch = worktreeMainBranch
			app.issuesList.SetIssues([]jira.Issue{{Key: "PROJ-1", Summary: "Task"}})
			repo := app.gitRepoPath
			path := filepath.Join(t.TempDir(), "new-worktree")

			app.cfg.Git.WorktreeFormat = "proj-1-task"
			if scenario == "occupied branch" {
				app.cfg.Git.WorktreeFormat = worktreeMainBranch
			}
			_, _ = app.handleKeyMsg(runeKey('W'))
			if scenario == "cancel" {
				_, cancel := app.inputModal.Update(tea.KeyMsg{Type: tea.KeyEsc})
				_, _ = app.Update(cancel())
			} else {
				if scenario == "occupied path" {
					if err := os.Mkdir(path, 0o755); err != nil {
						t.Fatal(err)
					}
				}
				_, confirm := app.inputModal.Update(tea.KeyMsg{Type: tea.KeyEnter})
				msg := confirm().(components.InputConfirmedMsg)
				msg.Text = path
				_, cmd := app.Update(msg)
				if cmd == nil {
					t.Fatal("expected creation command")
				}
				_, _ = app.Update(cmd())
				if app.statusPanel.ErrorMessage() == "" {
					t.Error("creation failure must surface")
				}
				if strings.Contains(app.helpBar.View(), "Creating worktree") {
					t.Error("failure must clear progress status")
				}
			}
			if app.gitRepoPath != repo || app.gitBranch != worktreeMainBranch || app.editContext.kind != editNone {
				t.Error("cancel/failure changed Git context or left pending input state")
			}
			if git.BranchExists(repo, "proj-1-task") {
				t.Error("cancel/failure created a branch")
			}
		})
	}
}

func TestCreateWorktreeConfiguredNamesFromLinkedCheckout(t *testing.T) {
	t.Parallel()
	repo := filepath.Join(t.TempDir(), "web-ui")
	if err := os.Rename(gitTestRepo(t), repo); err != nil {
		t.Fatal(err)
	}
	repo, err := filepath.EvalSymlinks(repo)
	if err != nil {
		t.Fatal(err)
	}
	linked := filepath.Join(filepath.Dir(repo), "linked")
	if out, err := exec.CommandContext(t.Context(), "git", "-C", repo, "worktree", "add", "--detach", linked).CombinedOutput(); err != nil {
		t.Fatalf("linked checkout: %v: %s", err, out)
	}
	app := focusApp(t)
	app.gitRepoPath = linked
	app.side = sideRight
	app.cfg.Git.BranchFormat = []config.BranchFormatRule{{When: config.BranchFormatCondition{Type: "*"}, Template: "task/{{.Key}}"}}
	app.cfg.Git.WorktreeFormat = "{{.RepoName}}_{{.Key}}"
	app.issuesList.SetIssues([]jira.Issue{{Key: "PROJ-1", Summary: "Task"}})
	_, _ = app.handleKeyMsg(runeKey('W'))
	_, confirm := app.inputModal.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if confirm == nil {
		t.Fatal("expected destination prompt in detail panel")
	}
	msg := confirm().(components.InputConfirmedMsg)
	if want := filepath.Join(filepath.Dir(repo), "web-ui_proj-1"); msg.Text != want {
		t.Fatalf("destination = %q, want %q", msg.Text, want)
	}
	_, cmd := app.Update(msg)
	if cmd == nil {
		t.Fatal("destination confirmation must create the worktree")
	}
	_, _ = app.Update(cmd())
	if branch, err := git.CurrentBranch(msg.Text); err != nil || branch != "web-ui_proj-1" {
		t.Errorf("worktree must use the w name as its branch: %q, %v", branch, err)
	}
}

func TestCreateWorktreeFlow(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("LAZYJIRA_CONFIG_DIR", configDir)
	if err := os.WriteFile(filepath.Join(configDir, "config.yml"), []byte("worktree:\n  defaultPath: ../trees\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	repo := filepath.Join(t.TempDir(), "web-ui")
	if err := os.Rename(gitTestRepo(t), repo); err != nil {
		t.Fatal(err)
	}
	repo, err = filepath.EvalSymlinks(repo)
	if err != nil {
		t.Fatal(err)
	}
	app := focusApp(t)
	app.ctx = t.Context()
	app.cfg = cfg
	app.side = sideLeft
	app.leftFocus = focusIssues
	app.gitRepoPath = repo
	app.gitBranch = worktreeMainBranch
	app.issuesList.SetIssues([]jira.Issue{{Key: "WEBSDK-218", Summary: "Audit skills"}})

	const worktreeBranch = "web-ui-websdk-218-audit-skills"
	_, _ = app.handleKeyMsg(runeKey('W'))
	if !app.inputModal.IsVisible() {
		t.Fatal("W must open the destination prompt")
	}
	if git.BranchExists(repo, worktreeBranch) {
		t.Fatal("branch created before destination confirmation")
	}
	_, confirm := app.inputModal.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if confirm == nil {
		t.Fatal("expected destination prompt")
	}
	pathMsg := confirm().(components.InputConfirmedMsg)
	wantPath := filepath.Join(filepath.Dir(repo), "trees", "web-ui-websdk-218-audit-skills")
	if pathMsg.Text != wantPath {
		t.Fatalf("destination = %q, want %q", pathMsg.Text, wantPath)
	}
	_, cmd := app.handleInputConfirmed(pathMsg)
	if cmd == nil {
		t.Fatal("destination confirmation must create the worktree")
	}
	_, _ = app.Update(cmd())
	if app.gitRepoPath != wantPath || app.gitBranch != worktreeBranch {
		t.Errorf("Git context = %q / %q", app.gitRepoPath, app.gitBranch)
	}
	if branch, err := git.CurrentBranch(wantPath); err != nil || branch != worktreeBranch {
		t.Errorf("worktree branch = %q, %v", branch, err)
	}
	if branch, err := git.CurrentBranch(repo); err != nil || branch != worktreeMainBranch {
		t.Errorf("original checkout changed: %q, %v", branch, err)
	}
	if !strings.Contains(app.helpBar.View(), "Created worktree") {
		t.Error("missing creation confirmation")
	}
	background := false
	cfg.CustomCommands = []config.CustomCommandConfig{{Key: "z", Name: "Current branch", Command: "git branch --show-current", Suspend: &background}}
	commands, err := cfg.ResolveCustomCommands()
	if err != nil {
		t.Fatal(err)
	}
	result := app.executeCustomCommand(commands[0], nil)().(customCommandFinishedMsg)
	if result.err != nil || strings.TrimSpace(result.output) != worktreeBranch {
		t.Errorf("custom command still targets old context: %q, %v", result.output, result.err)
	}
}
