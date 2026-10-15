package components

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/nikbrunner/lazyjira/pkg/tui/keyname/keytest"

	"github.com/nikbrunner/lazyjira/pkg/internal/testkit"
)

func TestDiffView_ShowAndIsVisible(t *testing.T) {
	t.Parallel()
	d := NewDiffView()
	testkit.AssertEqual(t, "initially invisible", d.IsVisible(), false)
	d.Show(testTitle, "old text", "new text")
	testkit.AssertEqual(t, "visible after Show", d.IsVisible(), true)
}

func TestDiffView_HideHidesView(t *testing.T) {
	t.Parallel()
	d := NewDiffView()
	d.Show(testTitle, "old", "new")
	d.Hide()
	testkit.AssertEqual(t, "hidden after Hide", d.IsVisible(), false)
}

func TestDiffView_SetSizeStoresSize(t *testing.T) {
	t.Parallel()
	d := NewDiffView()
	d.SetSize(80, 24)
	testkit.AssertEqual(t, "width stored", d.width, 80)
	testkit.AssertEqual(t, "height stored", d.height, 24)
}

func TestDiffView_ViewRendersTitle(t *testing.T) {
	t.Parallel()
	d := NewDiffView()
	d.SetSize(80, 24)
	d.Show(testTitle, "line1\nline2", "line1\nline3")
	out := d.View()
	plain := stripANSI(out)
	if !strings.Contains(plain, testTitle) {
		t.Errorf("expected title %q in view, got %q", testTitle, plain)
	}
}

func TestDiffView_ViewEmptyWhenInvisible(t *testing.T) {
	t.Parallel()
	d := NewDiffView()
	d.SetSize(80, 24)
	testkit.AssertEqual(t, "empty view when invisible", d.View(), "")
}

func TestDiffView_UpdateScrollsWithJK(t *testing.T) {
	t.Parallel()
	d := NewDiffView()
	d.SetSize(80, 10)
	d.Show(testTitle, strings.Repeat("line\n", 30), strings.Repeat("line\n", 30)+"extra")
	d, _ = d.Update(keytest.Press("j"))
	testkit.AssertEqual(t, "offset increased", d.offset, 1)
	d, _ = d.Update(keytest.Press("k"))
	testkit.AssertEqual(t, "offset back to zero", d.offset, 0)
}

func TestDiffView_UpdateConfirmSendsMsg(t *testing.T) {
	t.Parallel()
	d := NewDiffView()
	d.SetSize(80, 24)
	d.Show(testTitle, "old", "new")
	d, cmd := d.Update(keytest.Press("enter"))
	testkit.AssertEqual(t, "hidden after confirm", d.IsVisible(), false)
	if cmd == nil {
		t.Fatal("expected command from confirm")
	}
	msg := cmd()
	confirmed, ok := msg.(DiffConfirmedMsg)
	if !ok {
		t.Fatalf("expected DiffConfirmedMsg, got %T", msg)
	}
	testkit.AssertEqual(t, "confirmed content", confirmed.Content, "new")
}

func TestDiffView_UpdateCancelSendsMsg(t *testing.T) {
	t.Parallel()
	d := NewDiffView()
	d.SetSize(80, 24)
	d.Show(testTitle, "old", "new")
	d, cmd := d.Update(keytest.Press("esc"))
	testkit.AssertEqual(t, "hidden after cancel", d.IsVisible(), false)
	if cmd == nil {
		t.Fatal("expected command from cancel")
	}
	if _, ok := cmd().(DiffCancelledMsg); !ok {
		t.Error("expected DiffCancelledMsg")
	}
}

func TestDiffView_UpdateScrollCtrlDU(t *testing.T) {
	t.Parallel()
	d := NewDiffView()
	d.SetSize(80, 20)
	d.Show(testTitle, strings.Repeat("x\n", 50), strings.Repeat("x\n", 50)+"y")
	d, _ = d.Update(keytest.Press("ctrl+d"))
	testkit.AssertEqual(t, "ctrl+d increases offset", d.offset > 0, true)
	saved := d.offset
	d, _ = d.Update(keytest.Press("ctrl+u"))
	testkit.AssertEqual(t, "ctrl+u decreases offset", d.offset < saved, true)
}

func TestDiffView_UpdateMouseWheel(t *testing.T) {
	t.Parallel()
	d := NewDiffView()
	d.SetSize(80, 20)
	d.Show(testTitle, strings.Repeat("a\n", 30), strings.Repeat("a\n", 30)+"b")
	d, _ = d.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	testkit.AssertEqual(t, "wheel down increases offset", d.offset, 1)
	d, _ = d.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	testkit.AssertEqual(t, "wheel up decreases offset", d.offset, 0)
}

func TestDiffView_UpdateInvisibleIsNoop(t *testing.T) {
	t.Parallel()
	d := NewDiffView()
	d.SetSize(80, 24)
	d2, cmd := d.Update(keytest.Press("enter"))
	testkit.AssertEqual(t, "still invisible", d2.IsVisible(), false)
	testkit.AssertEqual(t, "no command", cmd == nil, true)
}

func TestDiffView_InterceptConsumesKeyWhenVisible(t *testing.T) {
	t.Parallel()
	d := NewDiffView()
	d.SetSize(80, 24)
	d.Show(testTitle, "old", "new")
	_, consumed := d.Intercept(keytest.Press("j"))
	testkit.AssertEqual(t, "key consumed", consumed, true)
}

func TestDiffView_InterceptPassesThroughWhenHidden(t *testing.T) {
	t.Parallel()
	d := NewDiffView()
	_, consumed := d.Intercept(keytest.Press("j"))
	testkit.AssertEqual(t, "not consumed when hidden", consumed, false)
}

func TestDiffView_RenderDrawsOnBackground(t *testing.T) {
	t.Parallel()
	d := NewDiffView()
	d.SetSize(80, 24)
	d.Show(testTitle, "old text", "new text")
	bg := testkit.BlankCanvas(80, 24)
	out := d.Render(bg, 80, 24)
	if !strings.Contains(stripANSI(out), testTitle) {
		t.Errorf("expected title in rendered output, got %q", stripANSI(out))
	}
}

func TestDiffView_RenderInvisibleReturnsBackground(t *testing.T) {
	t.Parallel()
	d := NewDiffView()
	bg := "bg content"
	out := d.Render(bg, 80, 24)
	testkit.AssertEqual(t, "bg passthrough", out, bg)
}

func TestDiffView_VisibleH(t *testing.T) {
	t.Parallel()
	d := NewDiffView()
	d.SetSize(80, 24)
	testkit.AssertEqual(t, "visibleH calculation", d.visibleH(), 24-4)
}
