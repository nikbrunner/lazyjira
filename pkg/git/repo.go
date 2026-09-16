package git

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitAvailable returns true if git is found in PATH
func GitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// IsRepo returns true if dir is inside a git repository
func IsRepo(dir string) bool {
	cmd := exec.CommandContext(context.Background(), "git", "-C", dir, "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

func RepoName(dir string) (string, error) {
	path, err := MainWorktreePath(dir)
	if err != nil {
		return "", err
	}
	return filepath.Base(path), nil
}

func MainWorktreePath(dir string) (string, error) {
	cmd := exec.CommandContext(context.Background(), "git", "-C", dir, "worktree", "list", "--porcelain", "-z")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// Git lists the main worktree first, including from a linked worktree.
	first, _, _ := strings.Cut(string(out), "\x00")
	return strings.TrimPrefix(first, "worktree "), nil
}

// CurrentBranch returns the current branch name, or an empty string when in detached HEAD state
func CurrentBranch(dir string) (string, error) {
	cmd := exec.CommandContext(context.Background(), "git", "-C", dir, "symbolic-ref", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
