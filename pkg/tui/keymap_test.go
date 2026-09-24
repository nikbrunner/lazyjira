package tui

import (
	"slices"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/tui/components"
)

func TestKeymapFromConfig_OverridesAndMatches(t *testing.T) {
	t.Parallel()

	var keybindingConfig config.KeybindingConfig
	keybindingConfig.Universal.Quit = "Q"
	keybindingConfig.Navigation.Down = "n"

	keymap := KeymapFromConfig(keybindingConfig)

	testkit.AssertSliceEqual(t, "quit binding overridden", keymap[ActQuit], []string{"Q"})
	testkit.AssertEqual(t, "Match resolves override", keymap.Match("Q"), ActQuit)
	testkit.AssertEqual(t, "MatchNav resolves override", keymap.MatchNav("n"), components.NavDown)
}

func TestKeymapFromConfig_EmptyKeepsDefaults(t *testing.T) {
	t.Parallel()

	defaults := DefaultKeymap()
	keymap := KeymapFromConfig(config.KeybindingConfig{})

	testkit.AssertSliceEqual(t, "quit default preserved", keymap[ActQuit], defaults[ActQuit])
}

func TestKeymapFromConfig_ExplicitBindingsDisplaceDefaults(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		issues config.IssueKeys
		key    string
		want   Action
	}{
		{name: "legacy create binding", issues: config.IssueKeys{CreateBranch: "b"}, key: "b", want: ActCreateBranch},
		{name: "existing worktree key override", issues: config.IssueKeys{Browser: "w"}, key: "w", want: ActBrowser},
		{name: "swapped branch bindings", issues: config.IssueKeys{CreateBranch: "b", CopyBranchName: "B"}, key: "B", want: ActCopyBranchName},
		{name: "one default alias displaced", issues: config.IssueKeys{CreateBranch: "ctrl+c"}, key: "ctrl+c", want: ActCreateBranch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			km := KeymapFromConfig(config.KeybindingConfig{Issues: tt.issues})
			for action, keys := range km {
				if action != tt.want && slices.Contains(keys, tt.key) {
					t.Errorf("%q also bound to %s; should belong only to %s", tt.key, action, tt.want)
				}
			}
			testkit.AssertEqual(t, "configured key", km.Match(tt.key), tt.want)
			testkit.AssertEqual(t, "unaffected quit alias", km.Match("q"), ActQuit)
		})
	}
}

func TestKeymapFromConfig_FocusPaneDefaultsAndOverrides(t *testing.T) {
	t.Parallel()
	defaults := DefaultKeymap()
	for _, tc := range []struct {
		key    string
		action Action
	}{
		{"0", ActFocusProj},
		{"1", ActFocusIssueTabs},
		{"2", ActFocusIssues},
		{"3", ActFocusInfo},
		{"4", ActFocusDetail},
	} {
		if got := defaults.Match(tc.key); got != tc.action {
			t.Errorf("default key %q = %q, want %q", tc.key, got, tc.action)
		}
	}
	config := config.KeybindingConfig{}
	config.Universal.FocusProj = "p"
	config.Universal.FocusIssueTabs = "t"
	config.Universal.FocusIssues = "i"
	config.Universal.FocusInfo = "f"
	config.Universal.FocusDetail = "d"
	km := KeymapFromConfig(config)
	for _, tc := range []struct {
		key    string
		action Action
	}{
		{"p", ActFocusProj}, {"t", ActFocusIssueTabs}, {"i", ActFocusIssues}, {"f", ActFocusInfo}, {"d", ActFocusDetail},
	} {
		if got := km.Match(tc.key); got != tc.action {
			t.Errorf("configured key %q = %q, want %q", tc.key, got, tc.action)
		}
	}
}

func TestKeymap_MatchUnknownReturnsEmpty(t *testing.T) {
	t.Parallel()

	keymap := DefaultKeymap()

	testkit.AssertEqual(t, "unknown key", keymap.Match("this-key-is-unbound"), Action(""))
	testkit.AssertEqual(t, "unknown nav key", keymap.MatchNav("this-key-is-unbound"), components.NavNone)
}
