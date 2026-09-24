# Ubiquitous Language

## Jira issues

| Term | Definition | Aliases to avoid |
| --- | --- | --- |
| **Issue** | A Jira work item with a unique issue key and associated fields. | Ticket, item without qualification |
| **Issue key** | The human-readable identifier of an issue, such as `WEBSDK-205`. | Issue ID, number |
| **Summary** | The short title stored in an issue's Jira summary field. | Description, body |
| **Status** | An issue's named position in its Jira workflow, such as Open or In Progress. | State without qualification |
| **Status indicator** | The displayed representation of a status, such as a symbol, abbreviation, or configured text label. | Status color, status emoji when the value is text |
| **Issue type** | The Jira classification of an issue, such as Task or Story. | Type without qualification |
| **Priority** | The Jira priority assigned to an issue. | Status, severity |
| **Assignee** | The Jira user assigned to an issue. | Owner |
| **Updated age** | The elapsed time since an issue's last update, displayed as a relative value such as `2h`. | Issue age, creation age |

## Git integration

| Term | Definition | Aliases to avoid |
| --- | --- | --- |
| **Repository name** | The main checkout's directory name. Linked worktrees share this name. Available as the `RepoName` slug in the worktree format. | Project name when referring to a Git repository |
| **Branch format rule** | A `git.branchFormat` entry that selects a naming template by issue type. Copying and creating a branch use the same rule. | Copy-only format |
| **Worktree format** | The `git.worktreeFormat` template for the name copied by `w` and used as the branch and default directory name by `W`. Its default includes the repository name. | Branch format rule |
| **Worktree location** | The parent directory configured by `worktree.defaultPath`. Relative locations resolve from the main checkout. | Branch name |
| **Worktree destination** | The full path confirmed when creating a worktree, combining its location and directory name by default. | Worktree location |

## Issue-list layout

| Term | Definition | Aliases to avoid |
| --- | --- | --- |
| **Issues panel** | The bordered screen region containing the column header row, issue rows, and footer for the active issue tab. | Issue list when referring to the whole panel |
| **Issue tab** | A named issue collection shown in the Issue tabs pane, usually defined by a JQL query. | Column, panel |
| **Issue tabs pane** | The focusable vertical pane listing issue tabs beside the Issues panel. Focusing it leaves the active issue tab unchanged. | Issues panel, issue list |
| **JQL query** | A Jira Query Language expression defining which issues Jira returns and their ordering. | Local filter |
| **Local filter** | A text filter applied to issues already loaded for the active issue tab. | JQL query, search without qualification |
| **Issue list** | The ordered collection of issues in the active issue tab after any local filter is applied. | Visible rows when referring to all matching issues |
| **Issue row** | The table-like line representing one issue through its configured columns. | Current line, cell |
| **Column** | An aligned vertical slot for one configured issue field, in `issueListFields` order. | Field when referring to layout |
| **Cell** | The displayed value at the intersection of an issue row and a column. | Column, row |
| **Column header row** | The fixed row of field labels above the issue rows inside the Issues panel. | Header without qualification, table-header-like row |
| **Viewport** | The portion of the issue list that fits in the panel's available row space. | Issue list when referring only to the on-screen subset |
| **Display width** | The number of terminal character cells occupied by text, accounting for wide characters and excluding ANSI escape sequences. | Byte length, character count |
| **Summary width** | The shared display width of the Summary column, fitted to the longest summary in the issue list, capped by available space, and at least as wide as its header. | Remaining width, per-row width |
| **Trailing space** | The unused horizontal area between the final column and the panel's right border. | Summary padding, column gap |

## Selection and colors

| Term | Definition | Aliases to avoid |
| --- | --- | --- |
| **Selected issue** | The issue targeted by issue-list navigation and issue actions. | Current line, highlighted cell |
| **Selection highlight** | The background marking the selected row, including its cells, inter-column spaces, and trailing space. | Column color, current-line color |
| **Column color** | The stable ANSI foreground color assigned to a field's header and values, independent of column order or issue status. | Selection color, status-dependent color |
| **ANSI palette** | The terminal-defined indexed colors used for column text. | ASCII colors |
| **Terminal background** | The terminal's default background color used as the basis for the selection highlight. | Highlight background, ANSI accent |
| **Highlight override** | An explicit `highlight` value in the theme configuration that takes precedence over the background-derived selection color. | Column color setting |

## Relationships

- The **Issue tabs pane** lists the **Issue tabs**; its active tab determines the **Issue list** shown in the **Issues panel**.
- An **Issue tab** contains zero or more **Issues**; an **Issue** can appear in more than one tab.
- A **Local filter** narrows the active tab's loaded issues into the current **Issue list**.
- Each **Issue row** represents one **Issue** and contains one **Cell** per configured **Column**.
- The **Column header row** and all **Issue rows** share column order and widths.
- The **Viewport** contains a subset of the **Issue rows**; **Summary width** accounts for the full **Issue list**, including off-screen issues.
- **Trailing space** belongs to the row after its final **Column**, not to the Summary column.
- A focused **Issue list** uses a **Selection highlight** for its **Selected issue**, while retaining each **Column color**.
- The **Selection highlight** is derived from the **Terminal background** unless a **Highlight override** applies: 8% toward white on dark backgrounds, or 8% toward black on light backgrounds.

## Example dialogue

> **Dev:** "Should **Summary width** use only the **Issue rows** inside the **Viewport**?"
>
> **Domain expert:** "No. Use every **Summary** in the current **Issue list**, including off-screen issues, so columns stay aligned while scrolling."
>
> **Dev:** "If a **Local filter** removes the longest summary, can the following **Columns** move left?"
>
> **Domain expert:** "Yes. The Summary column shrinks, and the released width becomes **Trailing space** after the final column."
>
> **Dev:** "Does the **Selection highlight** replace the **Column colors**?"
>
> **Domain expert:** "No. Column colors belong to the text; the highlight covers the selected row's entire background, including gaps and trailing space."

## Flagged ambiguities

- "Header" can mean a panel title, the issue-detail heading, or the table labels; use **Column header row** for the labels above issue rows.
- "Issue list" can mean the entire bordered panel or only the rows on screen; use **Issues panel**, **Issue list**, and **Viewport** for those distinct scopes.
- "Field" is Jira data, while **Column** is its presentation in the issue list; not every field is displayed as a column.
- "Summary" means the issue title, not its longer description; use **Summary width** when discussing the column's layout.
- "Longest displayed summary" is ambiguous about scrolling; calculate **Summary width** from the current **Issue list**, not just the **Viewport**.
- "Current line highlight" suggested a cell-level effect; use **Selected issue** for the navigation target and **Selection highlight** for the full-row background.
- "Different column" was used to request different colors; use **Column color** for foreground styling and "column separator" for a divider.
- "ASCII colors" means **ANSI palette** here; the background-derived **Selection highlight** is an RGB shade, approximated by the available palette on terminals without truecolor.
- "Updated" is the UI label, while **Updated age** describes the relative value beneath it; it is not the time since issue creation.
- UI labels such as `Key`, `Type`, and `Updated`, and configuration identifiers such as `issueListFields`, retain their existing spelling; use the qualified glossary terms in discussion.
