package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestCopyToastDoesNotMoveHelpBarOrCaptureKeys(t *testing.T) {
	t.Parallel()
	app := appForView(t)
	before := strings.Split(ansi.Strip(app.View()), "\n")

	_, cmd := app.Update(customCommandFinishedMsg{output: "Copied: ABC-42 summary"})
	if cmd == nil || app.toastText != "Copied: ABC-42 summary" {
		t.Fatalf("toast = %q, command = %v", app.toastText, cmd)
	}
	if got := app.helpBar.View(); strings.Contains(got, "Copied:") {
		t.Fatalf("copy output moved into help bar: %q", got)
	}
	lines := strings.Split(ansi.Strip(app.View()), "\n")
	if lines[len(lines)-1] != before[len(before)-1] {
		t.Errorf("help bar moved: before %q, after %q", before[len(before)-1], lines[len(lines)-1])
	}
	if !strings.Contains(strings.Join(lines, "\n"), "Copied: ABC-42 summary") {
		t.Error("toast not visible")
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if !app.showHelp {
		t.Error("toast captured the help key")
	}
}

func TestCopyToastExpiryIgnoresOlderTimeout(t *testing.T) {
	t.Parallel()
	app := appForView(t)
	_, _ = app.Update(customCommandFinishedMsg{output: "Copied: first"})
	first := app.toastEpoch
	_, _ = app.Update(customCommandFinishedMsg{output: "Copied: second"})
	if app.toastEpoch != first+1 {
		t.Fatalf("toast generation = %d, want %d", app.toastEpoch, first+1)
	}
	_, _ = app.Update(copyToastExpiredMsg{epoch: first})
	if app.toastText != "Copied: second" {
		t.Errorf("old timeout removed newer toast: %q", app.toastText)
	}
	_, _ = app.Update(copyToastExpiredMsg{epoch: app.toastEpoch})
	if app.toastText != "" {
		t.Errorf("toast did not expire: %q", app.toastText)
	}
}

func TestCopyToastFitsNarrowTerminalAndWideText(t *testing.T) {
	t.Parallel()
	app := appForView(t)
	app.width, app.height = 20, 6
	app.toastText = "Copied: 你好世界-long summary"
	base := strings.Repeat(strings.Repeat(".", app.width)+"\n", app.height-1) + strings.Repeat("K", app.width)
	out := ansi.Strip(app.renderCopyToast(base))
	lines := strings.Split(out, "\n")
	if len(lines) != app.height {
		t.Fatalf("rendered %d rows, want %d", len(lines), app.height)
	}
	for i, line := range lines {
		if width := ansi.StringWidth(line); width != app.width {
			t.Errorf("row %d width = %d, want %d: %q", i, width, app.width, line)
		}
	}
	if lines[app.height-1] != strings.Repeat("K", app.width) {
		t.Errorf("keybindings row changed: %q", lines[app.height-1])
	}
	if !strings.Contains(out, "Copied: 你好…") {
		t.Errorf("toast did not truncate wide text: %q", out)
	}

	app.toastText = "Copied: \x1b[31m你好世界-long summary\x1b[0m"
	styled := app.renderCopyToast(base)
	if !strings.Contains(ansi.Strip(styled), "Copied: 你好…") || !strings.Contains(styled, "\x1b[31m") {
		t.Errorf("styled toast was not truncated safely: %q", styled)
	}
}
