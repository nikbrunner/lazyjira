package components

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/nikbrunner/lazyjira/pkg/internal/testkit"
	"github.com/nikbrunner/lazyjira/pkg/tui/keyname/keytest"
)

func TestTextInput_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		initial     string
		cursor      int
		key         string
		wantValue   string
		wantCursor  int
		wantChanged bool
	}{
		{name: "left moves cursor", initial: "hello", cursor: 3, key: "left", wantValue: "hello", wantCursor: 2},
		{name: "right moves cursor", initial: "hello", cursor: 3, key: "right", wantValue: "hello", wantCursor: 4},
		{name: "right stops at end", initial: "hi", cursor: 2, key: "right", wantValue: "hi", wantCursor: 2},
		{name: "home jumps to start", initial: "hello", cursor: 3, key: "home", wantValue: "hello", wantCursor: 0},
		{name: "end jumps to end", initial: "hello", cursor: 1, key: "end", wantValue: "hello", wantCursor: 5},
		{name: "backspace deletes before cursor", initial: "hello", cursor: 3, key: "backspace", wantValue: "helo", wantCursor: 2, wantChanged: true},
		{name: "backspace at start is noop", initial: "hello", cursor: 0, key: "backspace", wantValue: "hello", wantCursor: 0},
		{name: "delete removes under cursor", initial: "hello", cursor: 1, key: "delete", wantValue: "hllo", wantCursor: 1, wantChanged: true},
		{name: "delete at end is noop", initial: "hi", cursor: 2, key: "delete", wantValue: "hi", wantCursor: 2},
		{name: "ctrl+a jumps to start", initial: "hello", cursor: 3, key: "ctrl+a", wantValue: "hello", wantCursor: 0},
		{name: "ctrl+e jumps to end", initial: "hello", cursor: 1, key: "ctrl+e", wantValue: "hello", wantCursor: 5},
		{name: "ctrl+w deletes word", initial: "foo bar", cursor: 7, key: "ctrl+w", wantValue: "foo ", wantCursor: 4, wantChanged: true},
		{name: "ctrl+k kills to end", initial: "hello", cursor: 2, key: "ctrl+k", wantValue: "he", wantCursor: 2, wantChanged: true},
		{name: "ctrl+u kills to start", initial: "hello", cursor: 2, key: "ctrl+u", wantValue: "llo", wantCursor: 0, wantChanged: true},
		{name: "space inserts", initial: "ab", cursor: 1, key: " ", wantValue: "a b", wantCursor: 2, wantChanged: true},
		{name: "rune inserts at cursor", initial: "ab", cursor: 1, key: "X", wantValue: "aXb", wantCursor: 2, wantChanged: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := NewTextInput()
			input.SetValue(tt.initial)
			input.setCursor(tt.cursor)

			got, changed := input.Update(keytest.Press(tt.key))

			testkit.AssertEqual(t, "value", got.Value(), tt.wantValue)
			testkit.AssertEqual(t, "cursor", got.CursorPos(), tt.wantCursor)
			testkit.AssertEqual(t, "changed", changed, tt.wantChanged)
		})
	}
}

func TestTextInput_PasteInsertsTextAtCursor(t *testing.T) {
	t.Parallel()
	input := NewTextInput()
	input.SetValue("ab")
	input.setCursor(1)

	got, changed := input.Update(tea.PasteMsg{Content: "X Y"})

	testkit.AssertEqual(t, "value", got.Value(), "aX Yb")
	testkit.AssertEqual(t, "cursor", got.CursorPos(), 4)
	testkit.AssertEqual(t, "changed", changed, true)
}

func TestTextInput_InsertAtCursorAndSetValue(t *testing.T) {
	t.Parallel()

	input := NewTextInput()
	input.SetValue("ac")
	testkit.AssertEqual(t, "cursor at end after SetValue", input.CursorPos(), 2)

	input.setCursor(1)
	input.InsertAtCursor("b")
	testkit.AssertEqual(t, "value after insert", input.Value(), "abc")
	testkit.AssertEqual(t, "cursor after insert", input.CursorPos(), 2)
}
