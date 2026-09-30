# Changelog

lazyjira continues [textfuel/lazyjira](https://github.com/textfuel/lazyjira) from its `2.19.2` release. Changes up to that
point are in the [upstream changelog](https://github.com/textfuel/lazyjira/blob/main/CHANGELOG.md).

## [Unreleased]

### Breaking

- Create issue form — nbr <nikolaus.brunner@protonmail.ch>
  - **Breaking:** `ctrl+g` opens the Description in `$EDITOR`. `e` types a letter there.
  - **Breaking:** Only `ctrl+s` creates the issue, from any panel. `enter` moves from Summary to Description, starts a
    new line in the Description, and edits the selected field in Fields.
  - **Breaking:** `n` opens the form directly, without an issue type picker first. Type is the first row in Fields and
    starts on the type last used in that project this session, or on the first one Jira lists. Changing it keeps
    Summary, Description, and the values of fields the new type shares. A subtask keeps its subtask type, and a
    duplicate starts on its source's type.

### Added

- Create issue form — nbr <nikolaus.brunner@protonmail.ch>
  - The Description is a text area in the form. Pasted text keeps its line breaks.
  - `ctrl+v` in the Description pastes a clipboard image, and dropped image files attach too. Each image adds an
    `[Image #N]` token and uploads as an attachment after the issue is created. The clipboard needs `osascript` on
    macOS, `wl-paste` on Wayland, or `xclip` on X11.
- Documentation — nbr <nikolaus.brunner@protonmail.ch>
  - The README recommends Jira CLIs for scripts and AI agents, in order: TWG CLI, Atlassian CLI, and jira-cli.

### Changed

- Create issue form — nbr <nikolaus.brunner@protonmail.ch>
  - The form fills the workspace, with Fields in a left column beside Summary and Description. Terminals narrower than
    68 columns stack the three panels.

---

## `0.7.0` &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 2026.09.30

### Highlights

lazyjira is a keyboard-driven terminal UI for Jira. Browse issues through your own JQL tabs, edit them in place, and turn an
issue into a Git branch or worktree. It works with Jira Cloud and Jira Server / Data Center.

![lazyjira workspace with issue tabs, the Issues panel, Issue info, and an issue description with a code block](https://raw.githubusercontent.com/nikbrunner/lazyjira/v0.7.0/docs/assets/recordings/screenshot.png)

#### Install

On macOS or Linux, the install script downloads the release, checks it against `checksums.txt`, and installs it to
`~/.local/bin`. mise and Go work too:

```sh
curl -fsSL https://raw.githubusercontent.com/nikbrunner/lazyjira/main/install.sh | sh
mise use -g github:nikbrunner/lazyjira
go install github.com/nikbrunner/lazyjira/cmd/lazyjira@latest
```

#### A fixed workspace

Five panes keep their place. `0` to `4` jump to a pane, `H` `J` `K` `L` move between them, and `+` maximizes Issues or
Details. `f` filters the loaded issues by status, type, and priority, and `s` runs a JQL search into a temporary tab.

![Maximized Issues panel, maximized issue details, filter picker, comments, JQL search, and project picker](https://raw.githubusercontent.com/nikbrunner/lazyjira/v0.7.0/docs/assets/recordings/preview.gif)

#### Open any issue by key

`#` opens a prompt that starts with the active project key, completes other project keys, and suggests issues as you type
the number.

![Opening PLAT-3 by key with project and issue completion](https://raw.githubusercontent.com/nikbrunner/lazyjira/v0.7.0/docs/assets/recordings/issue-lookup.gif)

#### Three ways to copy an issue

From the Issues panel, `y` copies the key and summary, `ctrl+y` copies a Markdown link, and `Y` copies the URL. With issues
marked, `y` copies the marked rows. The status bar shows what was copied, and `sanitize.remove` cleans the summary.

```text
y       SHOP-1 Implement shopping cart persistence
ctrl+y  [SHOP-1 Implement shopping cart persistence](https://example.atlassian.net/browse/SHOP-1)
Y       https://example.atlassian.net/browse/SHOP-1
```

The new keys are remappable as `copySummary` and `copyMarkdownLink`.

The [`0.6.5` notes](https://github.com/nikbrunner/lazyjira/releases/tag/v0.6.5) and the
[README](https://github.com/nikbrunner/lazyjira#features) list everything else.

### Changed

- Documentation — nbr <nikolaus.brunner@protonmail.ch>
  - The README opens with a screenshot, then the description, a Walkthrough of the three recordings, the demo, and
    installation.
  - The README feature list is ordered by impact, one line per feature, and covers copying, parent and child issues,
    custom fields, caching, and remappable keys.
  - The README describes `--dry-run` as simulated edits that are logged and never sent to Jira.
  - The README screenshot is re-recorded with the GIFs on every release, at twice their resolution.

---

## `0.6.5` &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 2026.09.29

### Highlights

This is the first release since the fork from [textfuel/lazyjira](https://github.com/textfuel/lazyjira) `2.19.2`. Everything
below is new in the fork: a rebuilt workspace, opening any issue by key, a filterable Issues panel, colors from your
terminal, Git worktrees from an issue, and three ways to install.

![lazyjira: browsing issues, switching issue tabs, and opening an issue in the maximized Issues panel](https://raw.githubusercontent.com/nikbrunner/lazyjira/v0.6.5/docs/assets/recordings/hero.gif)

#### Install

On macOS or Linux, the install script downloads the release, checks it against `checksums.txt`, and installs it to
`~/.local/bin`. mise and Go work too ([#5](https://github.com/nikbrunner/lazyjira/issues/5)):

```sh
curl -fsSL https://raw.githubusercontent.com/nikbrunner/lazyjira/main/install.sh | sh
mise use -g github:nikbrunner/lazyjira
go install github.com/nikbrunner/lazyjira/cmd/lazyjira@latest
```

#### A fixed workspace

The workspace keeps one layout regardless of focus. The Project selector and the App status panel share the top row, the
Issue tabs pane sits beside the Issues panel, and the Issue info and Issue details panes share the row below. Jump to a pane
with `0` to `4`, move between neighbours with `H` `J` `K` `L`, and switch issue tabs with `Tab` and `Shift+Tab` from
anywhere. `Enter` on the Project selector opens a searchable Project picker.

`+` maximizes Issues or Details. `Enter` on a maximized Issues panel opens the issue in maximized Details, and `esc` steps
back to the list, then to the split layout.

![Maximized Issues panel, maximized issue details, filter picker, comments, JQL search, and project picker](https://raw.githubusercontent.com/nikbrunner/lazyjira/v0.6.5/docs/assets/recordings/preview.gif)

#### Open any issue by key

`#` opens a prompt that starts with the active project key, such as `PLAT-`. Type the number and it suggests matching
issues from Jira with their summaries. Delete the prefix to switch projects: `web` completes to `WEBSDK-`. The issue opens
in maximized Issue details, and `esc` returns to where you were
([#20](https://github.com/nikbrunner/lazyjira/issues/20)).

![Opening PLAT-3 by key with project and issue completion](https://raw.githubusercontent.com/nikbrunner/lazyjira/v0.6.5/docs/assets/recordings/issue-lookup.gif)

#### Filter and group the Issues panel

`f` filters the loaded issues of a tab by status, issue type, and priority. Each value shows how many issues it would
leave, and values that would leave none are dimmed. `statusOrder` groups every list by your workflow and keeps the JQL
order within each status; set `sortByStatus: false` on a tab to keep pure JQL order
([#7](https://github.com/nikbrunner/lazyjira/issues/7)):

```yaml
statusOrder: [To Do, In Progress, In Review, Done]
```

When Jira holds more results than a tab loaded, the footer says so, for example `50/312 loaded`. `>` opens an issue's
children in a temporary tab ([#3](https://github.com/nikbrunner/lazyjira/issues/3)), and `s` in the Issue tabs pane starts a
JQL search from the focused tab's query ([#11](https://github.com/nikbrunner/lazyjira/issues/11)).

#### Colors from your terminal

lazyjira uses your terminal's 16 ANSI colors, so it matches your color scheme from the first start, code blocks included.
`themeColors` overrides single colors, and `themeDark` and `themeLight` apply by terminal background:

```yaml
gui:
  themeColors:
    green: "#a6e3a1"
```

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

#### Shell-safe custom commands

Custom command placeholders are shell-escaped automatically and must stand as whole, unquoted arguments. `{{.X | shellraw}}`
inserts a trusted value unescaped, and `$(...)` substitutions pass startup validation. A background command's last line of
output shows in a toast for three seconds.

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

- After creating an issue, Issue details shows the new issue ([#16](https://github.com/nikbrunner/lazyjira/issues/16)).
- A failed JQL search shows Jira's error messages above the search history.
- The sprint picker follows Jira pagination across Scrum boards and lists every available sprint.
- The selection highlight spans the full row, derived from the terminal background.
- The active issue tab is marked as `[Name]`, and clicks land on the tab under the cursor.

### Added

- Workspace — nbr <nikolaus.brunner@protonmail.ch>
  - The App status panel shows a colored connection indicator and cyan field labels.
- Issues panel — nbr <nikolaus.brunner@protonmail.ch>
  - The filter picker title shows how many loaded issues the selection leaves, e.g. `Filter issues · 4 of 50`, and the
    panel title shows the active filter.
  - `esc` clears marks, then the local filter, then the picker filter.
  - The help bar and `?` help list `f` filter, and `?` help lists `i` for the Issue info pane.
- Configuration — nbr <nikolaus.brunner@protonmail.ch>
  - `auth.json` is written atomically, so an interrupted save keeps the previous credentials.
- Demo mode — nbr <nikolaus.brunner@protonmail.ch>
  - `e2e/demo-config/config.yml` gives the demo five issue tabs, seven columns, and a status order. Start it with
    `LAZYJIRA_CONFIG_DIR=e2e/demo-config ./lazyjira-demo --demo`.
  - SHOP-1 has a description with a heading, a list, and a code block.
  - The SHOP project has 42 issues, enough to scroll a maximized Issues panel.
  - `make build-demo` writes `lazyjira-demo`, so it never replaces the regular `lazyjira` build
    ([#19](https://github.com/nikbrunner/lazyjira/issues/19)).
- Agent skills — nbr <nikolaus.brunner@protonmail.ch>
  - `lj-changelog`, `lj-commit`, and `lj-release` cover changelog entries, commits, and
    releases.
- CI and releases — nbr <nikolaus.brunner@protonmail.ch>
  - release-please proposes versions from Conventional Commits, and GoReleaser publishes binaries with the curated
    changelog section as release notes.

### Changed

- Workspace — nbr <nikolaus.brunner@protonmail.ch>
  - The Issues row takes two-fifths of the workspace height and the Issue details row three-fifths.
- Issues panel — nbr <nikolaus.brunner@protonmail.ch>
  - Summaries use the terminal's default foreground.
- Issue info pane — nbr <nikolaus.brunner@protonmail.ch>
  - Assignee and reporter use the regular text color.
- Documentation — nbr <nikolaus.brunner@protonmail.ch>
  - The README covers installing with Go on macOS, Linux, and Windows.
  - `GLOSSARY.md` defines the workspace, issue list, and Git integration terms.
  - The README describes the current workspace and features, follows the introduction with a short recording, and shows
    captioned recordings of browsing, creating an issue, and opening an issue by key.
  - The VHS tapes use the current keys, record from the demo config at a single width, and store the GIFs in
    `docs/assets/recordings/`.
  - The release process re-records the GIFs.
  - The README states that lazyjira only talks to your Jira host, and explains how to open an unsigned binary on macOS.
  - `docs/Config.md` lists the JQL history file as `jql_history`.

### Fixed

- Workspace — nbr <nikolaus.brunner@protonmail.ch>
  - Scrolling the Issue tabs pane by mouse wheel or overflow click stops at the last full window.
  - Closing a JQL or children tab keeps the other one with its issues and filters.
  - Opening or closing a JQL tab scrolls the Issue tabs pane to the active tab.
  - The text cursor stays visible when it sits inside the text, including on spaces.
- Demo mode — nbr <nikolaus.brunner@protonmail.ch>
  - Demo tabs apply `statusCategory != Done` and `issuetype = …` from their JQL.
  - Issues created in the demo have Demo User as reporter.
