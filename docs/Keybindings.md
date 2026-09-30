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
| `+` | Maximize or restore focused Issues or Details pane; `enter` in maximized Issues opens maximized Details, and `esc` steps back to maximized Issues, then to the split layout |
| `enter` in Issue tabs | Focus Issues without changing collection |
| `s` in Issue tabs | Open JQL search starting from the focused tab's query |
| `#` | Open an issue by key. The prompt starts with the active project key, completes project keys and issue numbers, and opens the issue in maximized Issue details; `esc` returns to the previous layout |

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
| `enter` | Open issue detail |
| `i` | Focus the Issue info pane for the selected issue; `esc` returns to the Issues panel |
| `space` | Mark or unmark the issue |
| `>` | Open the issue's children in a temporary tab; `backspace` opens the parent and `esc` goes back |
| `v` | Start a marked range at the cursor; press again to end it and keep the range marked |
| `f` | Filter the loaded issues by status, issue type, and priority; `space` toggles a value, `enter` applies. Counts show the issues each value would leave, and values that would leave none are dimmed. Each tab keeps its own filter for the session |
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
| `y` | Copy the issue key and summary. With marks, copy the marked issue rows as plain text, one per line, and clear the marks. `sanitize.remove` applies to both |
| `ctrl+y` | Copy the issue as a Markdown link, `[KEY Summary](URL)`, with `sanitize.remove` applied to the summary |
| `Y` | Copy the issue URL |
| `b` | Copy branch name (issues, info, or detail panel) |
| `B` | Create branch from issue using the same name as `b` |
| `w` | Copy repository-prefixed worktree name (issues, info, or detail panel) |
| `W` | Create a worktree using the `w` string for its branch and default directory name (issues, info, or detail panel) |
| `s` | JQL search |
| `x` | Close JQL tab |
| `esc` | Clear marks, then the local filter, then the picker filter |

## Help popup

| Key | Action |
|-----|--------|
| `/` | Filter keybindings |
| `j` / `k` | Navigate up / down |
| `g` / `G` | Jump to top / bottom |
| `esc` | Clear filter or close |
| `q` / `?` | Close |

## Create issue form

The form shows Fields beside Summary and Description, or stacks them on narrow terminals. Type is the first row in Fields. It starts on the type last used in the project this session, or on the first type Jira lists, and editing it switches the form to another type. A subtask's Type can't be changed. Its keys are fixed.

| Key | Action |
|-----|--------|
| `tab` / `shift+tab` | Next / previous panel |
| `ctrl+s` | Create the issue from any panel |
| `enter` | Move from Summary to Description; start a new line in Description; edit the selected field in Fields |
| `ctrl+g` | Open the Description in `$EDITOR`; the saved text returns to the form |
| `e` / `space` | Edit the selected field (Fields) |
| `/` | Filter fields (Fields) |
| `q` | Cancel (Fields) |
| `esc` | Cancel |

## General

| Key | Action |
|-----|--------|
| `?` | Help |
| `/` | Search |
| `r` | Refresh current view |
| `R` | Refresh all data |
| `[` / `]` | Switch tabs for the focused pane |
| `q` / `ctrl+c` | Quit |
