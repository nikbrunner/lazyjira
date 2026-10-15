//go:build !(darwin || dragonfly || freebsd || linux || netbsd || openbsd)

package theme

import "image/color"

func terminalBackground() color.Color { return nil }
