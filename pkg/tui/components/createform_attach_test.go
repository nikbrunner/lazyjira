package components

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/nikbrunner/lazyjira/pkg/internal/testkit"
)

func writeFile(t *testing.T, dir, name string, size int) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, make([]byte, size), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestPastedImages_ParsesDroppedPaths(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	spaced := writeFile(t, dir, "a b.png", 10)
	jpg := writeFile(t, dir, "c.JPG", 20)
	txt := writeFile(t, dir, "notes.txt", 5)
	escaped := strings.ReplaceAll(spaced, " ", `\ `)

	cases := []struct {
		name, text string
		want       []string
	}{
		{"escaped", escaped, []string{spaced}},
		{"single quoted", "'" + spaced + "'", []string{spaced}},
		{"double quoted", `"` + spaced + `"`, []string{spaced}},
		{"file url", "file://" + strings.ReplaceAll(spaced, " ", "%20"), []string{spaced}},
		{"several", escaped + " " + jpg + "\n", []string{spaced, jpg}},
		{"mixed with non-image", escaped + " " + txt, nil},
		{"missing file", filepath.Join(dir, "gone.png"), nil},
		{"plain text", "hello world", nil},
		{"blank", "  ", nil},
	}
	for _, tc := range cases {
		var got []string
		for _, img := range pastedImages(tc.text) {
			got = append(got, img.Path)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: pastedImages(%q) = %v, want %v", tc.name, tc.text, got, tc.want)
		}
	}
	imgs := pastedImages(jpg)
	testkit.AssertEqual(t, "name", imgs[0].Name, "c.JPG")
	testkit.AssertEqual(t, "size", imgs[0].Size, int64(20))
	testkit.AssertEqual(t, "user file is not temp", imgs[0].Temp, false)
}

func TestCreateForm_PasteDroppedImagesInsertsTokens(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := writeFile(t, dir, "a.png", 10)
	b := writeFile(t, dir, "b.webp", 10)
	form := showDescFocused(t, "")
	form.AttachImage(CreateAttachment{Path: writeFile(t, dir, "clip.png", 1), Size: 1, Temp: true})
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" "), Paste: false})

	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(a + " " + b), Paste: true})

	testkit.AssertEqual(t, "tokens numbered by attachment count", form.DescriptionText(), "[Image #1] [Image #2] [Image #3]")
	names := make([]string, 0, len(form.attachments))
	for _, att := range form.attachments {
		names = append(names, att.Name)
	}
	testkit.AssertSliceEqual(t, "names", names, []string{"image-1.png", "a.png", "b.webp"})
}

func TestCreateForm_PasteTextWithoutImagesStaysText(t *testing.T) {
	t.Parallel()
	form := showDescFocused(t, "")
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/no/such.png and text"), Paste: true})
	testkit.AssertEqual(t, "text pasted", form.DescriptionText(), "/no/such.png and text")
	testkit.AssertEqual(t, "no attachments", len(form.attachments), 0)
}

func TestCreateForm_CtrlVRequestsClipboardImage(t *testing.T) {
	t.Parallel()
	form := showDescFocused(t, "")
	cmd, _ := form.Intercept(tea.KeyMsg{Type: tea.KeyCtrlV})
	if cmd == nil {
		t.Fatal("expected a paste request")
	}
	if _, ok := cmd().(CreateFormPasteImageMsg); !ok {
		t.Errorf("ctrl+v = %T, want CreateFormPasteImageMsg", cmd())
	}
}

func TestCreateForm_SubmitCarriesAttachments(t *testing.T) {
	t.Parallel()
	form := showDescFocused(t, "")
	att := CreateAttachment{Path: "/tmp/x.png", Name: "x.png", Size: 3}
	form.AttachImage(att)
	cmd, _ := form.Intercept(tea.KeyMsg{Type: tea.KeyCtrlS})
	msg := cmd().(CreateFormSubmitMsg)
	testkit.AssertSliceEqual(t, "attachments", msg.Attachments, []CreateAttachment{att})
	testkit.AssertEqual(t, "token in description", msg.Fields["description"], any("[Image #1]"))
}

func TestCreateForm_CancelDeletesOnlyTempFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	temp := writeFile(t, dir, "clip.png", 1)
	own := writeFile(t, dir, "own.png", 1)
	form := showDescFocused(t, "")
	form.AttachImage(CreateAttachment{Path: temp, Temp: true})
	form.AttachImage(CreateAttachment{Path: own, Name: "own.png"})

	form.Intercept(tea.KeyMsg{Type: tea.KeyEsc})

	if _, err := os.Stat(temp); !os.IsNotExist(err) {
		t.Error("cancel should delete the clipboard temp file")
	}
	if _, err := os.Stat(own); err != nil {
		t.Error("cancel must keep the user's own file")
	}
}

func TestCreateForm_RendersAttachmentLine(t *testing.T) {
	t.Parallel()
	for _, size := range []struct{ w, h int }{{80, 21}, {60, 21}, {200, 50}} {
		form := NewCreateForm()
		form.SetSize(size.w, size.h)
		form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
		form.AttachImage(CreateAttachment{Name: "screenshot.png", Size: 214 * 1024})
		form.AttachImage(CreateAttachment{Name: "日本語.png", Size: 3 * 1024 * 1024 / 2})
		out := form.Render(testkit.BlankCanvas(size.w, size.h), size.w, size.h)
		lines := strings.Split(out, "\n")
		testkit.AssertEqual(t, "line count", len(lines), size.h)
		for i, line := range lines {
			if got := lipgloss.Width(line); got != size.w {
				t.Errorf("%dx%d line %d: width %d, want %d", size.w, size.h, i, got, size.w)
			}
		}
		plain := stripANSI(out)
		if size.w == 200 && !strings.Contains(plain, "Attachments  [Image #1] screenshot.png  214 KB · [Image #2] 日本語.png  1.5 MB") {
			t.Errorf("attachment line missing:\n%s", plain)
		}
		if !strings.Contains(plain, "Attachments  [Image #1]") {
			t.Errorf("%dx%d: attachment line missing", size.w, size.h)
		}
	}
}
