package components

import (
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/nikbrunner/lazyjira/pkg/internal/testkit"
	"github.com/nikbrunner/lazyjira/pkg/tui/keyname/keytest"
)

const (
	wtHome   = "/home/nik"
	wtParent = "/home/nik/repos"
	wtName   = "web-ui-proj-1-task"
)

func newWorktreeModal(existing ...string) *WorktreeModal {
	m := NewWorktreeModal()
	m.SetSize(160, 45)
	m.Show(wtName, wtParent, wtHome, func(name string) bool { return slices.Contains(existing, name) })
	return &m
}

func intercept(t *testing.T, m *WorktreeModal, msgs ...tea.Msg) tea.Cmd {
	t.Helper()
	var cmd tea.Cmd
	for _, msg := range msgs {
		c, ok := m.Intercept(msg)
		if !ok {
			t.Fatalf("message %v not consumed", msg)
		}
		cmd = c
	}
	return cmd
}

func TestWorktreeModal_RendersNamePathBranch(t *testing.T) {
	t.Parallel()
	view := stripANSI(newWorktreeModal().View())
	for _, want := range []string{
		"New worktree",
		"Name", wtName,
		"Path", "~/repos/" + wtName,
		"Branch", wtName + " (new)",
		"enter create · esc cancel",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q:\n%s", want, view)
		}
	}
}

func TestWorktreeModal_PreviewFollowsTyping(t *testing.T) {
	t.Parallel()
	m := newWorktreeModal("web-ui-proj-1-taskx")
	if !strings.Contains(stripANSI(m.View()), wtName+" (new)") {
		t.Fatal("generated name should be a new branch")
	}
	intercept(t, m, keytest.Press("x"))
	view := stripANSI(m.View())
	if !strings.Contains(view, "~/repos/"+wtName+"x") {
		t.Errorf("path did not follow the name:\n%s", view)
	}
	if !strings.Contains(view, wtName+"x (existing)") {
		t.Errorf("existing branch not marked:\n%s", view)
	}
	intercept(t, m, tea.PasteMsg{Content: "-y"})
	if !strings.Contains(stripANSI(m.View()), wtName+"x-y (new)") {
		t.Errorf("paste did not update the preview:\n%s", stripANSI(m.View()))
	}
}

func TestWorktreeModal_HomeAbbreviation(t *testing.T) {
	t.Parallel()
	m := NewWorktreeModal()
	m.SetSize(160, 45)
	m.Show("a", "/home/nikx", wtHome, func(string) bool { return false })
	if view := stripANSI(m.View()); !strings.Contains(view, "/home/nikx/a") {
		t.Errorf("sibling of home must not be abbreviated:\n%s", view)
	}
}

func TestWorktreeModal_WidthFitsLongestLine(t *testing.T) {
	t.Parallel()
	m := newWorktreeModal()
	lines := strings.Split(m.View(), "\n")
	pathLine := ""
	for _, l := range lines {
		if strings.Contains(stripANSI(l), "Path") {
			pathLine = stripANSI(l)
		}
	}
	if !strings.HasSuffix(pathLine, "~/repos/"+wtName+" │") {
		t.Errorf("modal should hug the longest line: %q", pathLine)
	}
	if w := lipgloss.Width(lines[0]); w >= 160-4 {
		t.Errorf("modal width %d should not fill the terminal", w)
	}
}

func TestWorktreeModal_NarrowTerminalTruncatesPathMiddle(t *testing.T) {
	t.Parallel()
	const width = 50
	m := NewWorktreeModal()
	m.SetSize(width, 40)
	m.Show("imfusion-portal-cp-159-portal-users-get-a-polished", "/home/nik/repos/imfusion", wtHome, func(string) bool { return false })
	var pathLine string
	for _, l := range strings.Split(m.View(), "\n") {
		if w := lipgloss.Width(l); w > width {
			t.Errorf("line width %d exceeds terminal width %d: %q", w, width, stripANSI(l))
		}
		if strings.Contains(stripANSI(l), "Path") {
			pathLine = stripANSI(l)
		}
	}
	if !strings.Contains(pathLine, "~/repos/") || !strings.Contains(pathLine, "polished") || !strings.Contains(pathLine, "...") {
		t.Errorf("path must keep start and end: %q", pathLine)
	}
}

func TestWorktreeModal_EnterConfirms(t *testing.T) {
	t.Parallel()
	m := newWorktreeModal()
	cmd := intercept(t, m, keytest.Press(" "), keytest.Press("enter"))
	if cmd == nil {
		t.Fatal("enter must confirm")
	}
	testkit.AssertEqual(t, "msg", cmd(), tea.Msg(WorktreeConfirmedMsg{Name: wtName, Path: wtParent + "/" + wtName}))
	testkit.AssertEqual(t, "visible", m.IsVisible(), false)
}

func TestWorktreeModal_EmptyNameIgnoresEnter(t *testing.T) {
	t.Parallel()
	m := newWorktreeModal()
	intercept(t, m, keytest.Press("ctrl+u"), keytest.Press(" "))
	if cmd := intercept(t, m, keytest.Press("enter")); cmd != nil {
		t.Errorf("enter with a blank name must do nothing, got %v", cmd())
	}
	testkit.AssertEqual(t, "visible", m.IsVisible(), true)
}

func TestWorktreeModal_EscCancels(t *testing.T) {
	t.Parallel()
	m := newWorktreeModal()
	cmd := intercept(t, m, keytest.Press("q"), keytest.Press("esc"))
	testkit.AssertEqual(t, "visible", m.IsVisible(), false)
	if cmd == nil {
		t.Fatal("esc must send a cancel message")
	}
	testkit.AssertEqual(t, "msg", cmd(), tea.Msg(WorktreeCancelledMsg{}))
}
