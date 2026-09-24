# Vertical issue tabs handoff

## Purpose

Continue the bounded design and implementation of opt-in vertical issue tabs in lazyjira. The design is not yet approved, so implementation has not started.

## Repository state

- Main repository: `/Users/brunner/repos/nikbrunner/lazyjira`
- Current branch: `integration` at `bcc013c`
- `main`, `origin/main`, and `upstream/main` match at `8c859b6`.
- The binary at `/Users/brunner/repos/nikbrunner/lazyjira/lazyjira` is built from `integration`.
- `~/.local/bin/lazyjira` points to that binary. Herdr runs it through `command = "lazyjira"`.
- The working tree is clean.
- This handoff is tracked only on the personal `integration` branch. `.superpowers/` remains locally ignored and neither artifact belongs in an upstream PR.

### Active branches and pull requests

- `fix/selected-issue-tabs` at `8bb38d8`
  - Upstream PR #130
  - Adds an active-tab marker and makes horizontal title rendering and mouse hit regions share one layout.
- `feat/detail-summary-header` at `ee78a51`
  - Kept for local use only because upstream PR #108 already implements the same feature
- `integration` merges both topic branches and is local only.

The vertical-tab work naturally builds on `fix/selected-issue-tabs`. Decide whether to use a stacked local branch now and rebase it onto upstream after PR #130 merges, or implement independently from `upstream/main`. Do not open a vertical-tabs PR with #130 commits included.

## Brainstorm classification

Bounded. This changes the existing issue-tab presentation and mouse handling. Follow the `brainstorming` skill, finish the short design in chat, and obtain explicit approval before implementation.

## Approved product decisions

- Layout: a labeled vertical rail inside the Issues panel, to the left of the issue table.
- Availability: opt-in configuration; horizontal tabs remain the default.
- Narrow panels: vertical mode remains vertical and truncates tab labels rather than changing orientation.
- Interaction: `[` and `]` keep switching tabs; `j/k` keep navigating issues; the rail never gains focus.
- Mouse: visible tab rows are clickable.
- Active row: full-row highlight plus a `›` marker.
- Vertical overflow:
  - The rail reserves fixed top and bottom indicator rows, even when either row is blank, so tab positions do not jump.
  - Indicators read `↑ N more` and `↓ N more`.
  - Clicking an indicator scrolls only the rail. It does not change the active tab.
  - Keyboard tab changes keep the active tab visible.

## Canonical mockups

The visual companion session is persisted at:

`/Users/brunner/repos/nikbrunner/lazyjira/.superpowers/brainstorm/1273-1789133940/`

Use these files:

- Placement comparison, layout A selected:
  `.superpowers/brainstorm/1273-1789133940/content/vertical-tabs-placement-v2.html`
- Active-row comparison, highlight selected and later amended with a `›` marker:
  `.superpowers/brainstorm/1273-1789133940/content/vertical-tabs-active-style.html`

The first `vertical-tabs-placement.html` is an unsuccessful narrow-card draft and is not design input.

The companion server is stopped. Restart it with the same project directory before using the visual companion again. Read the refreshed `server-info` file for the current URL. The persisted HTML files remain usable without the server.

## Relevant code

- `pkg/tui/views/issues.go`
  - `IssuesList.View`
  - horizontal `issueTabLayout`
  - `ClickTabAt`
- `pkg/tui/mouse.go:157`
  - maps title-bar clicks into `IssuesList.ClickTabAt`
  - vertical mode needs two-dimensional hit-testing for rail rows and overflow controls
- `pkg/config/config.go:236`
  - `GUIConfig`
- `docs/Config.md:178`
  - GUI configuration reference
- Tests:
  - `pkg/tui/views/issues_render_test.go`
  - `pkg/tui/views/issues_tab_title_test.go`
  - mouse handler tests near `pkg/tui/mouse_test.go`

## Open design points

Resolve these before presenting the short design:

1. Configuration name and values. A likely shape is `gui.issueTabLayout: horizontal | vertical`, defaulting to `horizontal`.
2. Rail width calculation and cap. It should fit the longest label when possible, reserve the marker column, and truncate by display width when constrained.
3. Minimum width retained for the issue table before the rail begins truncating more aggressively.
4. Rail viewport state for mouse-only scrolling, including how keyboard tab changes recenter the active row.
5. Collapsed Issues panels should retain the normal `[2] Issues` bar without rendering the rail.

## Expected implementation scope

Keep horizontal mode unchanged. Add a vertical renderer and its small layout model rather than branching throughout issue-row rendering. The vertical layout model should own rail width, visible tab indices, fixed overflow slots, and mouse hit regions.

Tests should cover:

- horizontal default and existing behavior
- config parsing and vertical opt-in
- display-width-safe truncation
- fixed overflow slots and counts
- active row visibility after keyboard navigation
- mouse selection of visible tabs
- overflow-row scrolling without tab activation
- collapsed panel behavior

Update `docs/Config.md` with the new GUI setting. Run `make check` before any commit or PR.

## Suggested next skills

1. `brainstorming` to finish and approve the bounded design
2. `dev-style-tdd` or `test-driven-development` for implementation
3. `dev-commit` when Nik explicitly approves committing
4. `verification-before-completion` before reporting completion
