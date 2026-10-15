# Keybindings

Press `?` inside lazyjira to see all available keys. The help popup lists the focused pane's keys under Local, then the keys that work everywhere under Global. Use `/` inside it to filter keybindings. The help bar along the bottom shows the global keys, with `?` always last. While issues are marked or the Issues list is filtered, it starts with what `esc` clears next.

Actions exposed under `keybinding` can be remapped in `config.yml`. Main-workspace `Tab`/`Shift+Tab` collection switching, uppercase `H`/`J`/`K`/`L` focus movement, and `j`/`k` collection switching in Issue tabs are fixed.

## Navigation

| Key | Action |
|-----|--------|
| `j` / `k` | Move down / up in focused lists; switch collections in the Issue tabs pane |
| `gg` / `G` | Jump to top / bottom |
| `ctrl+d` / `ctrl+u` | Scroll Details by one line from Issues or Info; half-page movement in Details and help |
| `ctrl+f` / `ctrl+b` | Half-page detail scroll from Issues or Info |
| `tab` / `shift+tab` | Switch issue collection without changing focus |
| `H` / `J` / `K` / `L` | Move focus according to the pane map below |
| `0` `1` `2` `3` `4` | Focus Project selector, Issue tabs, Issues, Issue info, Issue details |
| `+` | Maximize or restore focused Issues or Details pane; `enter` in maximized Issues opens maximized Details, and `esc` steps back to maximized Issues, then to the split layout |
| `@` | Show or hide the Command log for the session; the Issues and Details panes take the freed rows |
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
| `f` | Filter the loaded issues by the configured `status`, `type`, `priority`, `assignee`, and `sprint` columns, in column order; `space` toggles a value, `enter` applies. Counts show the issues each value would leave, and values that would leave none are dimmed. Each tab keeps its own filter for the session |
| `ctrl+g` | Filter the loaded issues by text like `/`, also matching descriptions. The filter shows as `~/query` |
| `t` | Transition status. With marks, transition every marked issue to a status they can all reach; only issues whose transition failed stay marked |
| `e` | Open the selected issue in the Issue Edit View (issues list and details pane), edit the focused field (Issue info pane), or edit the selected comment (comments tab) |
| `p` | Change priority |
| `a` | Change assignee |
| `n` | Create new issue (issues list), new comment (comments tab), or new link (the info panel's Lnk tab). In an epic's children view, the new issue becomes a child of that epic. A new link starts with a picker of every link type in both directions, such as `blocks` and `is blocked by`, then the `#` issue picker selects the other issue by key, number, or summary; `enter` creates the link. |
| `d` | Remove the selected link after confirmation (the info panel's Lnk tab) |
| `ctrl+n` | Duplicate issue |
| `S` | Create a subtask under the selected issue (issues list or the info panel's Sub tab). Parent and project are taken from the selection; not available when the selection is itself a subtask or an epic. |
| `c` | View comments |
| `o` | Open in browser. With marks, open each marked issue in its own browser tab and clear the marks; above 10 marked issues, ask first |
| `u` | Pick URL from description |
| `y` | Copy the issue key and summary, with `sanitize.remove` applied to the summary. With marks, copy one line per marked issue and clear the marks |
| `ctrl+y` | Copy the issue as a Markdown link, `[KEY Summary](URL)`, with `sanitize.remove` applied to the summary. With marks, copy one link per marked issue and clear the marks |
| `Y` | Copy the issue URL. With marks, copy one URL per marked issue and clear the marks |
| `alt+y` | Copy the issue row as plain text with its columns padded to width, with `sanitize.remove` applied to each cell. With marks, copy the marked issue rows, one per line, and clear the marks. On macOS, `alt+y` needs the terminal to send Option as Alt |
| `b` | Copy branch name (issues, info, or detail panel) |
| `B` | Create branch from issue using the same name as `b` |
| `w` | Copy repository-prefixed worktree name (issues, info, or detail panel) |
| `W` | Create a worktree (issues, info, or detail panel). The New worktree modal starts with the `w` string as the name, which sets both the directory under `worktree.defaultPath` and the branch |
| `s` | JQL search |
| `x` | Close JQL tab |
| `esc` | Clear marks, then the local filter, then the picker filter |

## Help popup

| Key | Action |
|-----|--------|
| `/` | Filter keybindings |
| `j` / `k` | Navigate up / down |
| `gg` / `G` | Jump to top / bottom |
| `esc` | Clear filter or close |
| `q` / `?` | Close |

## Issue Edit View

`n` opens the Issue Edit View to create an issue, and `e` opens it on the selected issue to edit it. The view shows Fields beside Summary and Description, or stacks them on narrow terminals. Its keys are fixed.

Type is the first row in Fields. When creating, it starts on the type last used in the project this session, or on the first type Jira lists, and editing it switches the view to another type. A subtask's Type and an edited issue's Type can't be changed.

When the new issue has a parent, a read-only Parent row below Type shows the parent's key and summary.

Editing starts from the issue's current values, and saving sends only the fields you changed.

| Key | Action |
|-----|--------|
| `tab` / `shift+tab` | Next / previous panel |
| `ctrl+s` | Create or save the issue from any panel |
| `enter` / `shift+enter` | Move from Summary to Description; start a new line in Description; edit the selected field in Fields |
| `ctrl+g` | Open the Description in `$EDITOR`; the saved text returns to the view |
| `ctrl+v` | Paste an image from the clipboard into the Description |
| `e` / `space` | Edit the selected field (Fields) |
| `/` | Filter fields (Fields) |
| `q` | Cancel (Fields) |
| `esc` | Cancel |

Dropping image files (PNG, JPEG, GIF, WebP) on the terminal while the Description is focused attaches them too. Each image adds an `[Image #N]` token at the cursor and a line under the Description, which lists only the images added in this view. The images upload as attachments once the issue is created or saved, and the tokens stay as text in the description. An image stays attached until the view closes, even if its token is deleted from the text.

Reading the clipboard needs `osascript` on macOS, `wl-paste` on Wayland, or `xclip` on X11.

## General

| Key | Action |
|-----|--------|
| `?` | Help |
| `/` | Search |
| `r` | Refresh current view |
| `R` | Refresh all data |
| `[` / `]` | Switch tabs for the focused pane |
| `q` / `ctrl+c` | Quit |
