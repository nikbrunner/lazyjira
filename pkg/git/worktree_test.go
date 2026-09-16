package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveWorktreePath(t *testing.T) {
	t.Parallel()
	repo := filepath.Join(t.TempDir(), "web-ui")
	if err := os.Rename(initRepo(t), repo); err != nil {
		t.Fatal(err)
	}
	repo, err := filepath.EvalSymlinks(repo)
	if err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(repo, "src")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	linked := filepath.Join(filepath.Dir(repo), "linked")
	gitRun(t, repo, "worktree", "add", "--detach", linked)
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{repo, nested, linked} {
		for _, tc := range []struct{ path, want string }{
			{"../trees/issue", filepath.Join(filepath.Dir(repo), "trees", "issue")},
			{filepath.Join(filepath.Dir(repo), "absolute"), filepath.Join(filepath.Dir(repo), "absolute")},
			{"~/worktrees/issue", filepath.Join(home, "worktrees", "issue")},
			{"~", home},
		} {
			got, err := ResolveWorktreePath(dir, tc.path)
			if err != nil || got != tc.want {
				t.Errorf("ResolveWorktreePath(%q, %q) = %q, %v; want %q", dir, tc.path, got, err, tc.want)
			}
		}
	}
}

func TestCreateWorktreeUsesCurrentHead(t *testing.T) {
	t.Parallel()
	repo := initRepo(t)
	linked := filepath.Join(t.TempDir(), "linked")
	gitRun(t, repo, "worktree", "add", "-b", "base", linked)
	if err := os.WriteFile(filepath.Join(linked, "from-linked"), []byte("linked HEAD"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, linked, "add", "from-linked")
	gitRun(t, linked, "commit", "-m", "linked base")
	path := filepath.Join(t.TempDir(), "new")
	if _, err := CreateWorktree(linked, "next", path); err != nil {
		t.Fatal(err)
	}
	if content, err := os.ReadFile(filepath.Join(path, "from-linked")); err != nil || string(content) != "linked HEAD" { //nolint:gosec // The worktree is created under t.TempDir.
		t.Errorf("new worktree did not start at current checkout HEAD: %q, %v", content, err)
	}
}

func TestCreateWorktree(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"new", "existing", "name matches remote", "occupied branch", "occupied path"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			repo := initRepo(t)
			path := filepath.Join(t.TempDir(), "worktree with spaces")
			branch := "issue-123"
			wantBranch := branch
			wantErr := false
			switch scenario {
			case "existing":
				createBranch(t, repo, branch)
			case "name matches remote":
				gitRun(t, repo, "remote", "add", "origin", repo)
				addRemoteRef(t, repo, "origin/"+branch)
				branch = "origin/" + branch
				wantBranch = branch
			case "occupied branch":
				branch = "main"
				wantErr = true
			case "occupied path":
				if err := os.Mkdir(path, 0o755); err != nil {
					t.Fatal(err)
				}
				wantErr = true
			}

			gotBranch, err := CreateWorktree(repo, branch, path)

			if (err != nil) != wantErr {
				t.Fatalf("error = %v, want error %v", err, wantErr)
			}
			if !wantErr {
				if gotBranch != wantBranch {
					t.Errorf("branch = %q, want %q", gotBranch, wantBranch)
				}
				if got, err := CurrentBranch(path); err != nil || got != wantBranch {
					t.Errorf("worktree branch = %q, %v", got, err)
				}
			}
			if got, err := CurrentBranch(repo); err != nil || got != "main" {
				t.Errorf("original checkout = %q, %v", got, err)
			}
			if scenario == "occupied path" && BranchExists(repo, branch) {
				t.Error("destination failure created a branch")
			}
		})
	}
}
