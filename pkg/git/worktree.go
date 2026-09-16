package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ResolveWorktreePath(repoPath, path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	root, err := MainWorktreePath(repoPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, path), nil
}

func CreateWorktree(repoPath, branch, path string) (string, error) {
	if _, err := os.Lstat(path); err == nil {
		return "", fmt.Errorf("destination already exists: %s", path)
	} else if !os.IsNotExist(err) {
		return "", err
	}
	args := []string{"-C", repoPath, "worktree", "add"}
	if BranchExists(repoPath, branch) {
		args = append(args, "--", path, branch)
	} else {
		args = append(args, "-b", branch, "--", path)
	}
	cmd := exec.CommandContext(context.Background(), "git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return branch, nil
}
