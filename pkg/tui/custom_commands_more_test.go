package tui

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
)

var errCommandFailed = errors.New("command failed")

func TestInitCustomCommands_ValidConfig(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.cfg.CustomCommands = []config.CustomCommandConfig{
		{
			Key:      "x",
			Name:     "test-cmd",
			Command:  "echo hello",
			Contexts: []string{"issues"},
		},
	}

	app.initCustomCommands()

	if len(app.customCmds) != 1 {
		t.Errorf("customCmds len = %d, want 1", len(app.customCmds))
	}
}

func TestInitCustomCommands_InvalidConfig(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.cfg.CustomCommands = []config.CustomCommandConfig{
		{
			Key:      "x",
			Name:     "bad-cmd",
			Command:  "{{.Broken",
			Contexts: []string{"issues"},
		},
	}

	app.initCustomCommands()

	if len(app.customCmds) != 0 {
		t.Error("customCmds should be empty after invalid config")
	}
}

func TestScopeNoun_AllBranches(t *testing.T) {
	t.Parallel()

	cases := []struct {
		scope config.ScopeMask
		want  string
	}{
		{config.ScopeIssue, "issue"},
		{config.ScopeProject, "project"},
		{config.ScopeIssue | config.ScopeComment, "comment"},
		{config.ScopeComment, "selection"},
	}

	for _, testCase := range cases {
		t.Run(testCase.want, func(t *testing.T) {
			t.Parallel()
			testkit.AssertEqual(t, "noun", scopeNoun(testCase.scope), testCase.want)
		})
	}
}

func TestLastNonEmptyLine_ReturnsLastLine(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"single line", "hello world", "hello world"},
		{"trailing newline", "line one\nline two\n", "line two"},
		{"empty input", "", ""},
		{"only newlines", "\n\n\n", ""},
		{"multiple lines", "line1\nline2\nline3", "line3"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertEqual(t, "last line", lastNonEmptyLine(testCase.input), testCase.want)
		})
	}
}

func TestCustomCommandBindings_ReturnsMatchingContext(t *testing.T) {
	t.Parallel()
	app := newTestApp()
	app.customCmds = []config.ResolvedCustomCommand{
		resolvedCommand(t, "y", "my-cmd", "echo hello", config.CtxIssues),
		resolvedCommand(t, "z", "other-cmd", "echo hello", config.CtxProjects),
	}

	bindings := app.customCommandBindings(config.CtxIssues)

	if len(bindings) != 1 {
		t.Errorf("bindings len = %d, want 1", len(bindings))
	}
	testkit.AssertEqual(t, "binding key", bindings[0].Key, "y")
	testkit.AssertEqual(t, "binding description", bindings[0].Description, "my-cmd")
}

func TestHandleCustomCommandFinished_ErrorShowsInStatusPanel(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()

	_, _ = app.handleCustomCommandFinished(customCommandFinishedMsg{
		err:    errCommandFailed,
		output: "something went wrong\ndetailed error",
	})

	if app.statusPanel.ErrorMessage() == "" {
		t.Error("status panel should show error message after command failure")
	}
}

func TestHandleCustomCommandFinished_SuccessWithOutput(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()

	_, _ = app.handleCustomCommandFinished(customCommandFinishedMsg{
		output: "operation succeeded",
	})

	if app.statusPanel.ErrorMessage() != "" {
		t.Errorf("status panel should not show error on success, got: %q", app.statusPanel.ErrorMessage())
	}
}

func TestHandleCustomCommandFinished_RefreshFetchesIssue(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	stubFullIssueFetch(fake, &jira.Issue{Key: testKey, Summary: "fresh"})
	app := newAppWithFake(t, fake)
	app.keymap = DefaultKeymap()
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})
	app.previewKey = testKey
	app.issueCache[testKey] = &jira.Issue{Key: testKey, Summary: "stale"}

	_, cmd := app.handleCustomCommandFinished(customCommandFinishedMsg{
		refresh: true,
	})

	if cmd == nil {
		t.Error("expected refresh cmd when refresh=true")
	}
}

func TestHandleCustomCommandRejectsRawContextInjection(t *testing.T) {
	t.Parallel()
	app := newTestApp()
	marker := filepath.Join(t.TempDir(), "executed")
	suspendFalse := false
	app.cfg.CustomCommands = []config.CustomCommandConfig{{
		Key: "x", Name: "mixed", Command: `printf '%s\\n' {{.Summary | shellraw}} {{.Key}}`,
		Suspend: &suspendFalse, Contexts: []string{"issues"},
	}}
	app.initCustomCommands()
	app.issuesList.SetIssues([]jira.Issue{{
		Key: "ABC-1", Summary: "$(touch " + marker + ") #",
	}})

	_, cmd, handled := app.handleCustomCommand("x")
	if !handled || cmd == nil {
		t.Fatalf("handled, cmd = %v, %v; want true and a command", handled, cmd)
	}
	rawMsg := cmd()
	msg, ok := rawMsg.(customCommandFinishedMsg)
	if !ok {
		t.Fatalf("message = %T, want customCommandFinishedMsg", rawMsg)
	}
	if msg.err == nil {
		t.Fatal("expected rendered command to be rejected")
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("marker exists after render rejection: %v", err)
	}
}

func TestExecuteCustomCommandUsesResolvedShell(t *testing.T) {
	shell, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh is not installed")
	}
	t.Setenv("SHELL", shell)
	app := newTestApp()
	app.ctx = t.Context()
	suspendFalse := false
	app.cfg.CustomCommands = []config.CustomCommandConfig{{
		Key: "x", Name: "uses resolved shell", Command: `printf '%s' {{.Key}}`,
		Suspend: &suspendFalse, Contexts: []string{"issues"},
	}}
	app.initCustomCommands()
	if len(app.customCmds) != 1 {
		t.Fatal("expected one resolved custom command")
	}

	t.Setenv("SHELL", filepath.Join(t.TempDir(), "missing-shell"))
	msg, ok := app.executeCustomCommand(app.customCmds[0], map[string]string{"Key": "shell-check"})().(customCommandFinishedMsg)
	if !ok {
		t.Fatalf("message = %T, want customCommandFinishedMsg", msg)
	}
	if msg.err != nil {
		t.Fatalf("command failed: %v", msg.err)
	}
	if msg.output != "shell-check" {
		t.Fatalf("output = %q, want shell-check", msg.output)
	}
}

func TestExecuteCustomCommand_BackgroundCapture(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.cfg.CustomCommands = nil
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	app.ctx = ctx
	suspendFalse := false
	rc := resolvedCommand(t, "x", "bg-cmd", "echo test-output", config.CtxIssues)
	rc.Suspend = &suspendFalse

	cmd := app.executeCustomCommand(rc, map[string]string{"Key": testKey})
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}

	msg := cmd()
	result, ok := msg.(customCommandFinishedMsg)
	if !ok {
		t.Fatalf("expected customCommandFinishedMsg, got %T", msg)
	}
	if result.err != nil {
		t.Errorf("unexpected error: %v", result.err)
	}
	if !strings.Contains(result.output, "test-output") {
		t.Errorf("output = %q, want to contain test-output", result.output)
	}
}
