package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nikbrunner/lazyjira/pkg/tui/components"
)

var errNoClipboardImage = errors.New("no image in clipboard")

// clipboardImageMsg carries the temp PNG a clipboard read produced.
type clipboardImageMsg struct {
	path string
	err  error
}

// clipboardPNGScript writes the clipboard's PNG data to the file named by its
// first argument and prints "none" when the clipboard holds no image.
var clipboardPNGScript = []string{
	"on run argv",
	"try",
	"set img to the clipboard as «class PNGf»",
	"on error",
	`return "none"`,
	"end try",
	"set f to open for access (POSIX file (item 1 of argv)) with write permission",
	"set eof f to 0",
	"write img to f",
	"close access f",
	`return "ok"`,
	"end run",
}

// readClipboardImage writes the clipboard image to a temp PNG and returns its
// path. The caller owns the file.
func readClipboardImage() (string, error) {
	tmp, err := os.CreateTemp("", "lazyjira-paste-*.png")
	if err != nil {
		return "", err
	}
	path := tmp.Name()
	_ = tmp.Close()
	if err := writeClipboardPNG(path); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	return path, nil
}

func writeClipboardPNG(path string) error {
	ctx := context.Background()
	switch runtime.GOOS {
	case "darwin":
		if _, err := exec.LookPath("osascript"); err != nil {
			return errors.New("image paste needs osascript, which was not found")
		}
		args := make([]string, 0, 2*len(clipboardPNGScript)+1)
		for _, line := range clipboardPNGScript {
			args = append(args, "-e", line)
		}
		out, err := exec.CommandContext(ctx, "osascript", append(args, path)...).Output()
		if err != nil {
			return fmt.Errorf("read clipboard image: %w", err)
		}
		if strings.TrimSpace(string(out)) != "ok" {
			return errNoClipboardImage
		}
		return nil
	case "linux":
		name, args := "xclip", []string{"-selection", "clipboard", "-t", "image/png", "-o"}
		if os.Getenv("WAYLAND_DISPLAY") != "" {
			name, args = "wl-paste", []string{"--type", "image/png"}
		}
		if _, err := exec.LookPath(name); err != nil {
			return fmt.Errorf("image paste needs %s, which was not found", name)
		}
		// Both tools exit non-zero when the clipboard holds no PNG.
		data, err := exec.CommandContext(ctx, name, args...).Output()
		if err != nil || len(data) == 0 {
			return errNoClipboardImage
		}
		return os.WriteFile(path, data, 0o600)
	default:
		return fmt.Errorf("image paste is not supported on %s", runtime.GOOS)
	}
}

// pasteClipboardImage reads the clipboard off the update loop.
func (a *App) pasteClipboardImage() tea.Cmd {
	read := a.readClipboardImage
	return func() tea.Msg {
		path, err := read()
		return clipboardImageMsg{path: path, err: err}
	}
}

// A read that lands while the form is submitting is dropped, so it can neither
// re-enable the form nor miss the upload.
func (a *App) handleClipboardImage(msg clipboardImageMsg) (tea.Model, tea.Cmd) {
	if !a.createForm.IsVisible() || a.createForm.IsLoading() {
		if msg.err == nil {
			_ = os.Remove(msg.path)
		}
		return a, nil
	}
	if msg.err != nil {
		text := msg.err.Error()
		if errors.Is(msg.err, errNoClipboardImage) {
			text = "No image in clipboard"
		}
		a.createForm.SetError(text)
		return a, nil
	}
	info, err := os.Stat(msg.path)
	if err != nil {
		_ = os.Remove(msg.path)
		return a, nil
	}
	a.createForm.AttachImage(components.CreateAttachment{Path: msg.path, Size: info.Size(), Temp: true})
	return a, nil
}
