package views

import (
	"testing"

	"github.com/nikbrunner/lazyjira/pkg/internal/testkit"
	"github.com/nikbrunner/lazyjira/pkg/tui/theme"
)

func TestNoneStyle_GrayForeground(t *testing.T) {
	t.Parallel()
	wantGray := theme.ColorGray
	testkit.AssertEqual(t, "foreground", noneStyle().GetForeground(), wantGray)
}
