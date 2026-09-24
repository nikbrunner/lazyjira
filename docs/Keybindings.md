# Keybindings

Press `?` inside lazyjira to see all available keys. Use `/` inside the help popup to filter keybindings.

Actions exposed under `keybinding` can be remapped in `config.yml`. Main-workspace `Tab`/`Shift+Tab` collection switching, uppercase `H`/`J`/`K`/`L` focus movement, and `j`/`k` collection switching in Issue tabs are fixed.

## Navigation

| Key | Action |
|-----|--------|
| `j` / `k` | Move down / up in focused lists; switch collections in the Issue tabs pane |
| `g` / `G` | Jump to top / bottom |
| `ctrl+d` / `ctrl+u` | Scroll Details by one line from Issues or Info; half-page movement in Details and help |
| `ctrl+f` / `ctrl+b` | Half-page detail scroll from Issues or Info |
| `tab` / `shift+tab` | Switch issue collection without changing focus |
| `H` / `J` / `K` / `L` | Move focus according to the pane map below |
| `0` `1` `2` `3` `4` | Focus Project selector, Issue tabs, Issues, Issue info, Issue details |
| `+` | Maximize or restore focused Issues or Details pane |
| `enter` in Issue tabs | Focus Issues without changing collection |

The Project selector opens a searchable picker with `enter` or a click. The Issue tabs pane is directly focusable; `enter` there focuses Issues.

| Focused pane | `H` | `J` | `K` | `L` |
| --- | --- | --- | --- | --- |
| Project selector | — | Issue tabs | — | — |
| Issue tabs | — | Info | Project selector | Issues |
| Issues | Issue tabs | Details | — | — |
| Details | Info | — | Issues | — |
| Info | — | — | Issue tabs | Details |

`[` / `]` switch issue collections from Issues or the Issue tabs pane, Details tabs from Details, and Info tabs from Info. Detail scroll keys can be remapped via `keybinding.detail`; list navigation keys can be remapped via `keybinding.navigation`.

## Issues

| Key | Action |
|-----|--------|
| `enter` / `space` | Open issue detail |
| `t` | Transition status |
| `e` | Edit (summary, description, or focused field) |
| `p` | Change priority |
| `a` | Change assignee |
| `n` | Create new issue (issues list) or new comment (comments tab) |
| `ctrl+n` | Duplicate issue |
| `S` | Create a subtask under the selected issue (issues list or the info panel's Sub tab). Parent and project are taken from the selection; not available when the selection is itself a subtask or an epic. |
| `c` | View comments |
| `o` | Open in browser |
| `u` | Pick URL from description |
| `y` | Copy issue URL |
| `b` | Copy branch name (issues, info, or detail panel) |
| `B` | Create branch from issue using the same name as `b` |
| `w` | Copy repository-prefixed worktree name (issues, info, or detail panel) |
| `W` | Create a worktree using the `w` string for its branch and default directory name (issues, info, or detail panel) |
| `s` | JQL search |
| `x` | Close JQL tab |

## Help popup

| Key | Action |
|-----|--------|
| `/` | Filter keybindings |
| `j` / `k` | Navigate up / down |
| `g` / `G` | Jump to top / bottom |
| `esc` | Clear filter or close |
| `q` / `?` | Close |

## General

| Key | Action |
|-----|--------|
| `?` | Help |
| `/` | Search |
| `r` | Refresh current view |
| `R` | Refresh all data |
| `[` / `]` | Switch tabs for the focused pane |
| `q` / `ctrl+c` | Quit |
