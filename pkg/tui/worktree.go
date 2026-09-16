package tui

import (
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/git"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
)

type gitWorktreeCreatedMsg struct {
	path   string
	branch string
}

func gitCreateWorktree(repoPath, branch, path string) tea.Cmd {
	return func() tea.Msg {
		name, err := git.CreateWorktree(repoPath, branch, path)
		if err != nil {
			return gitErrorMsg{err: err}
		}
		return gitWorktreeCreatedMsg{path: path, branch: name}
	}
}

func (a *App) worktreeName(issue *jira.Issue) (string, error) {
	repoName := ""
	if a.gitRepoPath != "" {
		repoName, _ = git.RepoName(a.gitRepoPath)
	}
	return git.GenerateWorktreeName(repoName, issue.Key, issue.Summary, a.cfg.Git.WorktreeFormat)
}

func (a *App) handleActionCreateWorktree() (tea.Model, tea.Cmd) {
	if a.side == sideLeft && a.leftFocus != focusIssues && a.leftFocus != focusInfo {
		return a, nil
	}
	issue := a.currentIssue()
	if issue == nil {
		return a, nil
	}
	if a.gitRepoPath == "" {
		a.statusPanel.SetError("not a git repository")
		return a, nil
	}
	name, err := a.worktreeName(issue)
	if err != nil {
		a.statusPanel.SetError("worktree name: " + err.Error())
		return a, nil
	}
	parent := a.cfg.Worktree.DefaultPath
	if parent == "" {
		parent = ".."
	}
	parent, err = git.ResolveWorktreePath(a.gitRepoPath, parent)
	if err != nil {
		a.statusPanel.SetError(err.Error())
		return a, nil
	}
	a.inputModal.Show("Create worktree: destination", filepath.Join(parent, name))
	a.editContext = editCtx{kind: editWorktreePath, branchName: name}
	return a, nil
}
