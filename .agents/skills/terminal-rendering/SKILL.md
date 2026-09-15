---
name: terminal-rendering
description: Use when implementing or fixing terminal text rendering involving ANSI styles, display widths, Unicode truncation, padding, alignment, or selection backgrounds. Applies across TUI views, not navigation, Jira operations, or application-state design.
---

# Terminal rendering

Treat terminal output as styled display cells, not a byte string. Use the project's glossary for the names of the UI elements involved.

## Measure and compose

- Check the rendering-library versions in `go.mod` before borrowing an API from current documentation.
- Measure visible width with ANSI-aware terminal helpers such as `ansi.StringWidth` or `lipgloss.Width`. Byte length and rune count are not display width.
- Use Unicode-aware, ANSI-aware truncation. Preserve escape sequences and wide characters when fitting a line to its width budget.
- Define which data determines a shared width: the full collection, filtered collection, or viewport. Headers and values use the same width calculation; scrolling alone should not change alignment unless requested.
- Account for borders, leading spaces, separators, and trailing padding in the width budget.
- Compose each colored cell from its complete style, including any selection background. Inner ANSI resets can clear a background applied only around the assembled row.
- Style separators and trailing padding as part of the selection. Keep the background inside the intended row, away from borders and neighboring rows.

For example, a selected row with cyan and magenta cells uses the same background on both cells and their separator:

```go
selection := lipgloss.NewStyle().Background(background)
left := selection.Foreground(lipgloss.Color("6")).Render("Key ")
right := selection.Foreground(lipgloss.Color("5")).Render("Type")
row := selection.Width(width).Render(left + selection.Render(" ") + right)
```

## Colors

- ANSI palette indices refer to terminal-defined colors; they are not fixed RGB values.
- Derive a background shade from the reported terminal background when that is the requested behavior. Keep explicit configuration overrides authoritative.
- Query terminal colors before the interactive event loop owns terminal input, not inside rendering functions. Define a fallback for terminals that cannot report their colors.
- Preserve readable text on both dark and light backgrounds; a subtle light selection cannot rely on forced white text.

## Regression proof

- Assert visible column positions and final line widths using literal fixture expectations. Include wide Unicode text, narrow widths, and empty values.
- For styling, inspect effective ANSI state across every relevant display cell. The presence of one background escape sequence does not prove full-row coverage.
- Cover selected and unselected rows, inter-column spaces, trailing padding, and style resets at row boundaries.
- Keep tests that modify a process-wide renderer or color profile serial, and restore the original state afterward.
- Use a pseudo-terminal when verifying color-query behavior. A non-TTY unit test proves only the fallback path.

## Sources of truth

- [Project dependencies](../../../go.mod) and [domain terms](../../../GLOSSARY.md)
- [Lip Gloss styling](https://pkg.go.dev/github.com/charmbracelet/lipgloss#Style.Render)
- [ANSI display-width helpers](https://pkg.go.dev/github.com/charmbracelet/x/ansi#StringWidth)
- [Terminal background detection](https://pkg.go.dev/github.com/muesli/termenv#Output.BackgroundColor)

Use documentation matching the installed dependency versions when implementing a pattern.
