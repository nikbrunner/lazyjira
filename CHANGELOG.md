# Changelog

lazyjira continues [textfuel/lazyjira](https://github.com/textfuel/lazyjira) from its `2.19.2` release. Changes up to that
point are in the [upstream changelog](https://github.com/textfuel/lazyjira/blob/main/CHANGELOG.md).

## [0.14.0](https://github.com/nikbrunner/lazyjira/compare/v0.13.0...v0.14.0) (2026-10-15)

### Highlights

#### Marked issues work like a single issue

Mark issues with `space` or a `v` range, and the copy keys act on each one: `y` copies `KEY Summary`, `ctrl+y` a Markdown
link and `Y` the URL, one line per issue. `o` opens every marked issue in the browser and asks first above 10. The padded
table of the visible columns is now on `alt+y`, and works on the cursor row without marks. On macOS, `alt+y` needs the
terminal's Option-as-Alt setting.

#### Links and parents from the Issue info pane

On the Links tab, `n` adds a link: pick the link type and direction, then the target issue with the same lookup as `#`.
`d` removes the selected link after a confirmation. Editing Parent opens that lookup too, prefilled with the project
key, and an empty prompt clears the parent.

#### One name for a new worktree

`W` opens a small dialog with the worktree name, prefilled from the issue. It shows the folder under
`worktree.defaultPath` and the branch while you type, and the name sets both, so folder and branch always match. A branch
that already exists is marked and gets checked out.

#### Issue tabs for one project

An issue tab with `projects` appears only while one of those projects is active:

```yaml
issueTabs:
  - name: "Web SDK releases"
    jql: "project = WEBSDK AND fixVersion in unreleasedVersions()"
    projects: [WEBSDK]
```

#### A closer fit with the terminal

lazyjira now runs on Bubble Tea v2. The colors follow light and dark theme switches while it runs, the terminal title
shows the active project, and `shift+enter` adds a line in the Description. Copies also go through the terminal (OSC 52),
so `y` and friends work over SSH. Your keybindings keep working; `space` and `" "` both name the space key.

#### Smaller changes

- `@` hides or shows the Command log, and `gui.showCommandLog` sets how it starts.
- While issues are marked or the list is filtered, the help bar starts with what `esc` clears next.
- `lazyjira --version` reports the release after `go install`.

#### Upgrading from 0.13

With marks, `y` copies `KEY Summary` per issue. The padded rows it copied before are on `alt+y`. While issues are marked,
custom commands bound to `ctrl+y`, `Y`, `alt+y` or `o` give way to the built-in action, as for `y` and `t`.

Building from source needs Go 1.26. Release binaries are not affected.

### ⚠ BREAKING CHANGES

* **tui:** with marks, y copies key and summary per issue. The padded rows it copied are on alt+y.

### Features

* **config:** show issue tabs only for their configured projects ([086d2d5](https://github.com/nikbrunner/lazyjira/commit/086d2d5d50e488d44146d0c6b0133b5dfc038c3a))
* **tui:** add and remove issue links from the Links tab ([86bcc4d](https://github.com/nikbrunner/lazyjira/commit/86bcc4d112c0ea9871a3c0879c2cea12d3fbe0ef))
* **tui:** complete the parent key when editing Parent ([e799774](https://github.com/nikbrunner/lazyjira/commit/e79977437cc073a5a2d480dbb1fa2be6871fa44b))
* **tui:** copy and open marked issues like a single issue ([269a9ce](https://github.com/nikbrunner/lazyjira/commit/269a9ce9917634e9a66168038ef48857d74accb5))
* **tui:** copy through the terminal with OSC 52 as well ([280c55b](https://github.com/nikbrunner/lazyjira/commit/280c55b234ffb16c6cedf181206ac22f03372133))
* **tui:** follow light and dark switches of the terminal while running ([729defb](https://github.com/nikbrunner/lazyjira/commit/729defb245885628013324159ede63bf1dd9911e))
* **tui:** insert a new line with shift+enter in the Description ([4cadba7](https://github.com/nikbrunner/lazyjira/commit/4cadba7ee7f5e6dccfa860c70141fa4050b6a7ab))
* **tui:** name a new worktree in a modal that previews path and branch ([c400343](https://github.com/nikbrunner/lazyjira/commit/c4003437e66dd90a73cd781344a9a9d44592f4df))
* **tui:** show the project in the terminal title and progress while dialogs load ([8521233](https://github.com/nikbrunner/lazyjira/commit/852123344e52238fa5caf2e2451eb5f66b0dea8f))
* **tui:** show what esc clears in the help bar ([5402f2e](https://github.com/nikbrunner/lazyjira/commit/5402f2e7f34f15bc7c0d5335cc6eaed83c221449))
* **tui:** toggle the command log with @ ([9ca4455](https://github.com/nikbrunner/lazyjira/commit/9ca44556f617e42b89b06b17a7a5506aa8976047))

### Refactors

* **tui:** move to Bubble Tea v2, Lip Gloss v2 and Bubbles v2 ([4513644](https://github.com/nikbrunner/lazyjira/commit/45136440ec778a86c8102cf4d480ce82fa1c5e6d)), closes [#25](https://github.com/nikbrunner/lazyjira/issues/25)

### Bug Fixes

* report the release version after go install ([11b8108](https://github.com/nikbrunner/lazyjira/commit/11b8108817564463cd45005310699616a20104b7))
* **tui:** draw the confirmation above other open dialogs ([a16ff48](https://github.com/nikbrunner/lazyjira/commit/a16ff48049f82fd5a868700a415a8817eae566d8))

## [0.13.0](https://github.com/nikbrunner/lazyjira/compare/v0.12.0...v0.13.0) (2026-10-15)

### Highlights

#### Transition several issues at once

Mark issues with `space` or a `v` range and press `t` to move them all in one go. The picker lists only the statuses every
marked issue can reach, and each issue takes its own workflow transition to get there. When a transition fails, the error
names the issue and only that issue stays marked, so a second `t` retries just the failures. While issues are marked, `t`
keeps this behavior even if a custom command is bound to it.

### Features

* **tui:** transition all marked issues at once ([c65f1b0](https://github.com/nikbrunner/lazyjira/commit/c65f1b063210929bd25b2a5726d828a350115afd))

## [0.12.0](https://github.com/nikbrunner/lazyjira/compare/v0.11.0...v0.12.0) (2026-10-14)

### Highlights

#### Full-text filter

Press `ctrl+g` in the Issues list to filter by description too. It opens the same filter bar as `/` and matches key,
summary, assignee, and the issue description. The bar and the panel title show `~/query`, so you can tell the two filters
apart. Like `/`, it searches the issues the tab has loaded, and comments are left out. Remap it with
`keybinding.issues.bodySearch`.

#### Upgrading from 0.11

Jump to top is now `gg`, and a single `g` waits for the second one. To keep the single key:

```diff
 keybinding:
   navigation:
+    top: g
```

### Features

* **tui:** filter the Issues list by description with ctrl+g ([971dfa0](https://github.com/nikbrunner/lazyjira/commit/971dfa002c3a6986ed46da873310bc64cdcc084b))
* **tui:** jump to the top with gg ([4fb7f3c](https://github.com/nikbrunner/lazyjira/commit/4fb7f3c4084d1942a4d51435ea632d52472b3bb7))

## [0.11.0](https://github.com/nikbrunner/lazyjira/compare/v0.10.0...v0.11.0) (2026-10-13)

### Highlights

#### A sprint column, and a filter picker that follows your columns

Add `sprint` to `issueListFields` to see each issue's sprint in the Issues list. Long names are cut at 16 cells, and issues
without a sprint stay blank. One tab can then hold sprint and backlog issues side by side.

```yaml
gui:
  issueListFields: [key, summary, status, type, priority, updated, sprint]
```

`f` now offers a group for each status, type, priority, assignee, and sprint column you show, in column order. Assignee and
sprint are new filters. Issues without a value show up as `None`.

#### Upgrading from 0.10

The filter picker only shows groups for configured columns. With the default `key, status, summary`, that is Status alone.
Add the columns you filter by to `issueListFields`:

```diff
- issueListFields: [key, status, summary]
+ issueListFields: [key, status, type, priority, summary]
```

### Features

* **tui:** add a sprint column to the Issues list ([3982d61](https://github.com/nikbrunner/lazyjira/commit/3982d61666e47738e9ecbeec58e1b6592bba9d38))
* **tui:** offer filter picker groups for the configured columns ([6a94e92](https://github.com/nikbrunner/lazyjira/commit/6a94e92bdd37fa0a1814e01dc6c7831ee71c179c))

## [0.10.0](https://github.com/nikbrunner/lazyjira/compare/v0.9.0...v0.10.0) (2026-10-08)

### Highlights

#### Create issues inside an epic

Open an epic's children with `>` and press `n`, and the new issue becomes a child of that epic. The Issue Edit View shows
a read-only Parent row under Type with the epic's key and summary, so you can see where the issue lands before you save.
Subtasks created with `S` show their parent the same way. The children view tab names its parent too:
`Children of PLAT-1`.

#### A help popup for the pane you are in

`?` lists the focused pane's keys first, under Local, and the keys that work everywhere after them, under Global. The
popup scrolls with the cursor, so long lists reach their end, and actions without a key stay out of it. The help bar
along the bottom shows only the global keys, the same in every pane.

### Features

* **tui:** create new issues under the epic of the open children view ([8513d26](https://github.com/nikbrunner/lazyjira/commit/8513d2678846abddb2abb79ce3941ee5ef12ae2a))
* **tui:** list the focused pane's keys first in the help popup ([e788b9d](https://github.com/nikbrunner/lazyjira/commit/e788b9dc84d04ba688dadf2ddefde056050dc9ae))
* **tui:** name the parent in the children view tab ([13e480a](https://github.com/nikbrunner/lazyjira/commit/13e480aad4411cd2b7f19489cd1c30eb3246991e))

## [0.9.0](https://github.com/nikbrunner/lazyjira/compare/v0.8.0...v0.9.0) (2026-10-01)

### Highlights

#### See which project you are writing to

The Issue Edit View shows the project above the form, next to the App status panel, so you can tell at a glance where a
new issue lands. For a subtask that is the parent's project, even when it differs from the one you have open. The Summary
title names it too: `Create issue in PLAT`.

![The Issue Edit View with the project shown above the form](https://raw.githubusercontent.com/nikbrunner/lazyjira/v0.9.0/docs/assets/recordings/create-issue.gif)

The view now opens with Fields focused. `tab` moves to Summary, then Description, then back to Fields.

#### Important fixes

- `e` on Sprint in the Issue Edit View opens the sprint picker on real Jira sites, and the chosen sprint is saved with
  the issue.

## [0.8.0](https://github.com/nikbrunner/lazyjira/compare/v0.7.0...v0.8.0) (2026-09-30)

### Highlights

#### One view to create and edit issues

`n` and `e` open the same Issue Edit View. Fields sit in a column on the left, Summary and a full-height Description on
the right, and the Description is a real text area: type, paste, and press `enter` for a new line. `ctrl+s` saves from
any panel, and `ctrl+g` opens the Description in `$EDITOR` when you want your own editor.

Creating an issue is one step. Type is the first row in Fields and starts on the type you used last in the project.
Changing it keeps what you wrote. Editing an existing issue sends only the fields you changed, and `e` in the Issue info
pane still edits a single field in place.

![Creating an issue with a pasted screenshot in the Issue Edit View](https://raw.githubusercontent.com/nikbrunner/lazyjira/v0.8.0/docs/assets/recordings/create-issue.gif)

#### Paste screenshots into issues

`ctrl+v` in the Description pastes the image on your clipboard as `[Image #1]`, and dragging image files into the terminal
does the same. The images upload as attachments when you save, for new and existing issues alike. The clipboard needs
`osascript` on macOS, `wl-paste` on Wayland, or `xclip` on X11.

#### Upgrading from 0.7

Creating an issue skips the type picker, and only `ctrl+s` saves it:

```diff
- n, pick a type, enter
+ n, ctrl+s
```

`e` in the Issues panel and the Issue details pane opens the Issue Edit View, and `ctrl+g` inside it opens `$EDITOR`:

```diff
- e   edit the summary, or the description in $EDITOR
+ e   open the Issue Edit View; ctrl+g for $EDITOR
```

### Added

- Documentation — nbr <nikolaus.brunner@protonmail.ch>
  - The README recommends Jira CLIs for scripts and AI agents, in order: TWG CLI, Atlassian CLI, and jira-cli.

## [0.7.0](https://github.com/nikbrunner/lazyjira/compare/v0.6.5...v0.7.0) (2026-09-30)

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

## [0.6.5](https://github.com/nikbrunner/lazyjira/releases/tag/v0.6.5) (2026-09-29)

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
