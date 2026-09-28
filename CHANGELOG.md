# Changelog

lazyjira continues [textfuel/lazyjira](https://github.com/textfuel/lazyjira) from its `2.19.2` release. Changes up to that
point are in the [upstream changelog](https://github.com/textfuel/lazyjira/blob/main/CHANGELOG.md).

## [Unreleased]

### Added

- Issues panel ([#7](https://github.com/nikbrunner/lazyjira/issues/7)) — nbr <nikolaus.brunner@protonmail.ch>
  - `statusOrder` groups issue lists by status, keeping JQL order within each status and in tabs with `sortByStatus: false`.
  - `f` opens a per-tab filter picker for status, issue type, and priority, and the panel title shows the active selection.
  - Filter picker counts update as values are toggled, and values that would leave no issues are dimmed and cannot be checked.
  - The filter picker title shows how many loaded issues the current selection leaves, e.g. `Filter issues · 4 of 50`.
  - The help bar and `?` help list `f` filter, and `?` help lists `i` for the Issue info pane.

### Changed

- Workspace — nbr <nikolaus.brunner@protonmail.ch>
  - The Issues row takes two-fifths of the workspace height and the Issue details row three-fifths.

### Fixed

- Workspace — nbr <nikolaus.brunner@protonmail.ch>
  - Scrolling the Issue tabs pane by mouse wheel or overflow click stops at the last full window.
  - `esc` clears marks, then the local filter, then the picker filter.
  - The footer shows a hint such as `50/312 loaded` when Jira holds more results than the tab loaded.

---

## `0.6.5` &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 2026.09.28

### Highlights

This is the first release since the fork from [textfuel/lazyjira](https://github.com/textfuel/lazyjira) `2.19.2`. Everything
below is new in the fork: a rebuilt workspace, Git worktrees from an issue, shell-safe custom commands, a reworked Issues
panel with multi-issue copy, and cached Jira data.

#### A fixed workspace

The workspace keeps one layout regardless of focus. The Project selector and the App status panel share the top row, the
Issue tabs pane sits beside the Issues panel, and the Issue info and Issue details panes share the row below. Jump to a pane
with `0` to `4`, move between neighbours with `H` `J` `K` `L`, and press `+` to maximize Issues or Details. `Enter` on the
Project selector opens a searchable Project picker. The Issue details pane shows the issue summary above its tabs.

#### Branches and worktrees from an issue

From the Issues panel, the Issue info pane, or the Issue details pane, `b` copies the branch name and `B` creates that
branch. `w` copies a worktree name, and `W` creates the worktree: it uses that name for the branch and the directory, lets you
confirm or edit the destination, then switches lazyjira's Git context to it.

```yaml
git:
  worktreeFormat: "{{if .RepoName}}{{.RepoName}}-{{end}}{{.Key}}-{{.Summary}}"
worktree:
  defaultPath: ".."
```

Relative locations resolve from the main checkout, so worktrees land beside it by default. Existing directories are never
overwritten.

#### Shell-safe custom commands

Custom command placeholders are shell-escaped automatically and must stand as whole, unquoted arguments. `{{.X | shellraw}}`
inserts a trusted value unescaped, and `$(...)` substitutions pass startup validation. A background command's last line of
output shows in a toast for three seconds.

#### Mark issues and copy their rows

In the Issues panel, `space` marks single issues and `v` marks a range from the cursor. Press `v` again to close the range
and keep it marked, then add or remove single issues with `space`. `y` copies the marked issue rows as plain text, one line
per issue with your configured columns, and `esc` clears the marks. Without marks, `y` copies the issue URL.

Copied rows drop square brackets, so a summary like `[web-ui] Release` does not turn into a Markdown link when pasted.
`sanitize.remove` sets which strings are removed, and custom commands use the same list through `{{.Summary | sanitize}}`:

```yaml
sanitize:
  remove: ["[", "]"]
```

While issues are marked, these keys keep their built-in meaning even if a custom command uses the same key. Remap the range
key with `keybinding.issues.visualSelect`.

#### A clearer Issues panel

The Issues panel has a column header row that labels each `issueListFields` column, gives each column its own color, and
sits above a separator line. The Summary column fits the longest summary in the whole issue list. A local filter stays
applied after you confirm the search, the panel title shows it, and `esc` clears it.

#### Faster pickers

Boards, sprint options, project users, and issue-creation metadata are cached in memory for five minutes. `R` (Refresh All)
clears the cache.

```yaml
cache:
  enabled: true
  ttl: 5m
```

#### Sharp borders

Panels, modals, and toasts can drop the rounded corners. Rounded stays the default
([#6](https://github.com/nikbrunner/lazyjira/issues/6)).

```yaml
gui:
  borders: sharp
```

#### Important fixes

- The sprint picker follows Jira pagination across Scrum boards and lists every available sprint.
- The selection highlight spans the full row, derived from the terminal background.
- The active issue tab is marked as `[Name]`, and clicks land on the tab under the cursor.

#### Upgrading from lazyjira 2.19.2

The Nix flake and package are gone. Install from GitHub Releases, or with Go under the new module path:

```sh
go install github.com/nikbrunner/lazyjira/v2/cmd/lazyjira@main
```

The focus keys follow the new pane order. `focusStatus` is gone; rename it in your config and check the other focus keys:

```diff
 keybinding:
   universal:
-    focusDetail: "0"
-    focusStatus: "1"
-    focusProjects: "4"
+    focusProjects: "0"
+    focusIssueTabs: "1"
+    focusDetail: "4"
```

`b` copies the branch name, and creating the branch moves to `B`:

```diff
 keybinding:
   issues:
-    createBranch: b
+    copyBranchName: b
+    createBranch: B
```

Branch format templates use lowercase variables, and `{{.Summary}}` is already a slug:

```diff
-      template: "{{.Key}}-{{.Summary | slugify}}"
+      template: "{{.Key}}-{{.Summary}}"
```

Quoted placeholders in custom commands fail validation. Remove the quotes around them; lazyjira escapes the value itself.

In the Issues panel, `space` marks the issue, `enter` opens its detail, and `>` opens its children
([#3](https://github.com/nikbrunner/lazyjira/issues/3)).

In the Issue details pane, `ctrl+d` and `ctrl+u` scroll, and `H` `J` `K` `L` move focus. `sidePanelWidth` defaults to `22`,
and `collapsedPanelHeight` has no effect.

### Added

- Workspace — nbr <nikolaus.brunner@protonmail.ch>
  - `Tab` and `Shift+Tab` switch issue tabs from the main workspace.
  - The App status panel shows a colored connection indicator and cyan field labels.
- Configuration — nbr <nikolaus.brunner@protonmail.ch>
  - `auth.json` is written atomically, so an interrupted save keeps the previous credentials.
- Agent skills — nbr <nikolaus.brunner@protonmail.ch>
  - `lj-changelog`, `lj-commit`, and `lj-release` cover changelog entries, commits, and
    releases.
- CI and releases — nbr <nikolaus.brunner@protonmail.ch>
  - release-please proposes versions from Conventional Commits, and GoReleaser publishes binaries with the curated
    changelog section as release notes.

### Changed

- Issues panel — nbr <nikolaus.brunner@protonmail.ch>
  - Summaries use the terminal's default foreground.
  - `Enter` on a maximized Issues panel opens the issue in maximized Details. `Esc` returns to the maximized Issues panel, and
    a second `Esc` restores the split layout.
- Issue info pane — nbr <nikolaus.brunner@protonmail.ch>
  - Assignee and reporter use the regular text color.
- Documentation — nbr <nikolaus.brunner@protonmail.ch>
  - The README covers installing with Go on macOS, Linux, and Windows.
  - `GLOSSARY.md` defines the workspace, issue list, and Git integration terms.
