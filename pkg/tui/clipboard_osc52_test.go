package tui

import (
	"reflect"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/nikbrunner/lazyjira/pkg/tui/keyname/keytest"
)

func sendsOSC52(cmd tea.Cmd, text string) bool {
	return cmd != nil && reflect.DeepEqual(cmd(), tea.SetClipboard(text)())
}

//nolint:paralleltest // Replaces the process-wide clipboard command runner.
func TestCopy_SendsOSC52WithExternalCommand(t *testing.T) {
	calls := stubExternal(t)
	app := copyApp(t)

	_, cmd := app.handleKeyMsg(keytest.Press("y"))

	if !sendsOSC52(cmd, "A-205 web Release notes") {
		t.Fatal("no OSC 52 command with the copied summary")
	}
	if got := calls.lastCopied(t); got != "A-205 web Release notes" {
		t.Fatalf("external command got %q", got)
	}
}

//nolint:paralleltest // Replaces the process-wide clipboard command runner.
func TestCopy_SendsOSC52WithoutExternalCommand(t *testing.T) {
	original := runExternalCommand
	t.Cleanup(func() { runExternalCommand = original })
	runExternalCommand = func(string, bool, string, ...string) {}
	app := copyApp(t)

	_, cmd := app.handleKeyMsg(keytest.Press("Y"))

	if !sendsOSC52(cmd, app.issueURL("A-205")) {
		t.Fatal("no OSC 52 command without an external clipboard tool")
	}
}
