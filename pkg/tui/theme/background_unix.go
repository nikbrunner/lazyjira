//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd

package theme

import (
	"image/color"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
	"golang.org/x/sys/unix"
)

// backgroundQueryTimeout bounds how long startup waits for a terminal that
// does not answer.
const backgroundQueryTimeout = time.Second

var (
	backgroundReply = regexp.MustCompile(`\x1b\]11;([^\x07\x1b]*)(?:\x07|\x1b\\)`)
	attributesReply = regexp.MustCompile(`\x1b\[\?[0-9;]*c`)
)

func terminalBackground() color.Color {
	fd := os.Stdout.Fd()
	if os.Getenv("CI") != "" || !term.IsTerminal(fd) {
		return nil
	}
	if bg := queryBackground(os.Stdout); bg != nil {
		return bg
	}
	return colorFGBG(os.Getenv("COLORFGBG"))
}

// queryBackground asks the terminal on tty for its background color. It
// skips multiplexers, which can serve several terminals at once, and
// terminals the process does not own in the foreground.
func queryBackground(tty *os.File) color.Color {
	switch t := os.Getenv("TERM"); {
	case strings.HasPrefix(t, "screen"), strings.HasPrefix(t, "tmux"), strings.HasPrefix(t, "dumb"):
		return nil
	}
	fd := tty.Fd()
	if pgrp, err := unix.IoctlGetInt(int(fd), unix.TIOCGPGRP); err != nil || pgrp != unix.Getpgrp() { //nolint:gosec
		return nil
	}

	state, err := term.MakeRaw(fd)
	if err != nil {
		return nil
	}
	defer term.Restore(fd, state) //nolint:errcheck

	rd, err := uv.NewCancelReader(tty)
	if err != nil {
		return nil
	}
	defer rd.Close() //nolint:errcheck
	timer := time.AfterFunc(backgroundQueryTimeout, func() { rd.Cancel() })
	defer timer.Stop()

	// Every terminal answers the device attributes request, so its reply
	// ends the wait whether or not the background request was answered.
	if _, err := io.WriteString(tty, ansi.RequestBackgroundColor+ansi.RequestPrimaryDeviceAttributes); err != nil {
		return nil
	}
	var reply []byte
	buf := make([]byte, 256)
	for !attributesReply.Match(reply) {
		n, err := rd.Read(buf)
		if err != nil {
			break
		}
		reply = append(reply, buf[:n]...)
	}
	if m := backgroundReply.FindSubmatch(reply); m != nil {
		return ansi.XParseColor(string(m[1]))
	}
	return nil
}
