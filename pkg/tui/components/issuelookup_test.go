package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func lookupRunes(l IssueLookup, s string) (IssueLookup, tea.Cmd) {
	return l.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)})
}

func lookupMsg(t *testing.T, cmd tea.Cmd) tea.Msg {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected a command")
	}
	return cmd()
}

var lookupProjects = []ProjectChoice{
	{Key: "PLAT", Name: "Platform Services"},
	{Key: "WEBSDK", Name: "Web SDK"},
	{Key: "SHOP", Name: "Online Shop for web"},
}

func TestIssueLookupPrefillAsksForIssues(t *testing.T) {
	t.Parallel()
	l := NewIssueLookup()
	msg := lookupMsg(t, l.Show("PLAT-", lookupProjects))
	if got, ok := msg.(IssueLookupQueryMsg); !ok || got.Query != "PLAT-" {
		t.Fatalf("msg = %#v, want IssueLookupQueryMsg for PLAT-", msg)
	}
}

func TestIssueLookupCompletesProjectKey(t *testing.T) {
	t.Parallel()
	l := NewIssueLookup()
	l.Show("", lookupProjects)

	l, _ = lookupRunes(l, "web")
	if len(l.suggestions) != 2 || l.suggestions[0].Key != "WEBSDK" || l.suggestions[1].Key != "SHOP" {
		t.Fatalf("suggestions = %#v, want key match WEBSDK before name match SHOP", l.suggestions)
	}

	l, cmd := l.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if l.Query() != "WEBSDK-" || !l.IsVisible() {
		t.Fatalf("query = %q, visible = %v, want WEBSDK- and still open", l.Query(), l.IsVisible())
	}
	if msg, ok := lookupMsg(t, cmd).(IssueLookupQueryMsg); !ok || msg.Query != "WEBSDK-" {
		t.Fatalf("completion should ask for issues, got %#v", msg)
	}
}

func TestIssueLookupEnterOpensTypedKey(t *testing.T) {
	t.Parallel()
	l := NewIssueLookup()
	l.Show("plat-", lookupProjects)
	l, _ = lookupRunes(l, "12")
	l.SetIssueSuggestions("plat-12", []LookupSuggestion{{Key: "PLAT-120"}, {Key: "PLAT-12"}})

	l, cmd := l.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if msg, ok := lookupMsg(t, cmd).(IssueLookupSelectedMsg); !ok || msg.Key != "PLAT-12" {
		t.Fatalf("msg = %#v, want the typed key PLAT-12", msg)
	}
	if l.IsVisible() {
		t.Error("lookup should close after selecting")
	}
}

func TestIssueLookupEnterOpensNavigatedSuggestion(t *testing.T) {
	t.Parallel()
	l := NewIssueLookup()
	l.Show("PLAT-1", lookupProjects)
	l.SetIssueSuggestions("PLAT-1", []LookupSuggestion{{Key: "PLAT-1"}, {Key: "PLAT-12"}})

	l, _ = l.Update(tea.KeyMsg{Type: tea.KeyDown})
	_, cmd := l.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if msg, ok := lookupMsg(t, cmd).(IssueLookupSelectedMsg); !ok || msg.Key != "PLAT-12" {
		t.Fatalf("msg = %#v, want the highlighted PLAT-12", msg)
	}
}

func TestIssueLookupIgnoresSuggestionsForOldQuery(t *testing.T) {
	t.Parallel()
	l := NewIssueLookup()
	l.Show("PLAT-1", lookupProjects)
	l, _ = lookupRunes(l, "2")

	l.SetIssueSuggestions("PLAT-1", []LookupSuggestion{{Key: "PLAT-1"}})
	if len(l.suggestions) != 0 {
		t.Fatalf("suggestions = %#v, want none for the outdated query", l.suggestions)
	}
}

func TestIssueLookupEscCancels(t *testing.T) {
	t.Parallel()
	l := NewIssueLookup()
	l.Show("PLAT-", lookupProjects)

	l, cmd := l.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if _, ok := lookupMsg(t, cmd).(IssueLookupCancelledMsg); !ok || l.IsVisible() {
		t.Fatal("esc should close the lookup and send IssueLookupCancelledMsg")
	}
}

func TestIssueLookupTruncatesLongSummaries(t *testing.T) {
	t.Parallel()
	l := NewIssueLookup()
	l.SetSize(100, 40)
	l.Show("PLAT-1", lookupProjects)
	long := strings.Repeat("very long summary ", 20)
	l.SetIssueSuggestions("PLAT-1", []LookupSuggestion{{Key: "PLAT-1", Label: long}, {Key: "PLAT-12", Label: "short"}})

	lines := strings.Split(l.View(), "\n")
	if got, want := len(lines), 5; got != want {
		t.Fatalf("view has %d lines, want %d (border, prompt, two suggestions, border):\n%s", got, want, l.View())
	}
	for _, line := range lines {
		if w := ansi.StringWidth(line); w > 85 {
			t.Errorf("line is %d cells wide, want at most 85: %q", w, ansi.Strip(line))
		}
	}
}
