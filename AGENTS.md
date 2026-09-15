# lazyjira

## Project context

- This is an independently maintained Jira terminal UI. The repository is `nikbrunner/lazyjira`; GitHub operations target that repository.
- The Go module still uses `github.com/textfuel/lazyjira/v2`. Changing that identity is a separate migration, not routine cleanup.
- Preserve the original MIT copyright notice in `LICENSE` and the README attribution.
- Use `GLOSSARY.md` for domain terms. Keep it aligned when terminology changes.
- Treat `handoffs/` as historical task context; check its assumptions against current code and Git state.

## Agent configuration

- `AGENTS.md` and `.agents/skills/` are canonical. `CLAUDE.md` and `.claude` are relative symlinks; keep them as symlinks.
- Load the `terminal-rendering` skill for ANSI styling, display-width calculations, truncation, padding, alignment, and selection-background changes.

## Local preview

- On Nik's machine, `~/.local/bin/lazyjira` points to this checkout's root `lazyjira` executable. Confirm the symlink before changing the installed preview.
- A build in another worktree does not update that launcher. Build the intended checkout before asking Nik to reopen the app.

## Publishing

- Installation links and release configuration include original-project destinations, including its Homebrew tap and shared AUR package names.
- Review publishing destinations and credentials before creating release tags or invoking release automation; repository ownership alone does not establish ownership of those destinations.
