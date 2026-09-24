package tui

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/tui/views"
)

func newTestApp() *App {
	return &App{
		cfg:         &config.Config{Jira: config.JiraConfig{Host: "example.atlassian.net"}},
		issuesList:  views.NewIssuesList(),
		projectList: views.NewProjectList(),
		detailView:  views.NewDetailView(views.BuiltinRenderer{}),
		side:        sideLeft,
		leftFocus:   focusIssues,
	}
}

func TestActiveContexts(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		setup func(app *App)
		want  []config.Context
	}{
		{
			"left issues",
			func(app *App) { app.side = sideLeft; app.leftFocus = focusIssues },
			[]config.Context{config.CtxIssues},
		},
		{
			"left info",
			func(app *App) { app.side = sideLeft; app.leftFocus = focusInfo },
			[]config.Context{config.CtxInfo},
		},
		{
			"left projects",
			func(app *App) { app.side = sideLeft; app.leftFocus = focusProjects },
			[]config.Context{config.CtxProjects},
		},
		{
			"right detail details",
			func(app *App) {
				app.side = sideRight
				app.detailView.SetIssue(&jira.Issue{Key: "X-1"})
				app.detailView.SetActiveTab(views.TabDetails)
			},
			[]config.Context{config.CtxDetail},
		},
		{
			"right detail comments",
			func(app *App) {
				app.side = sideRight
				app.detailView.SetIssue(&jira.Issue{Key: "X-1"})
				app.detailView.SetActiveTab(views.TabComments)
			},
			[]config.Context{config.CtxDetailComments, config.CtxDetail},
		},
		{
			"right project mode",
			func(app *App) {
				app.side = sideRight
				app.detailView.SetProject(&jira.Project{Key: "P"})
			},
			[]config.Context{config.CtxProjects},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			app := newTestApp()
			testCase.setup(app)
			got := app.activeContexts()
			if len(got) != len(testCase.want) {
				t.Fatalf("len = %d (%v), want %d (%v)", len(got), got, len(testCase.want), testCase.want)
			}
			for i := range got {
				if got[i] != testCase.want[i] {
					t.Errorf("[%d] = %q, want %q", i, got[i], testCase.want[i])
				}
			}
		})
	}
}

func resolvedCommand(t *testing.T, key, name, command string, contexts ...config.Context) config.ResolvedCustomCommand {
	t.Helper()
	rawContexts := make([]string, len(contexts))
	for i, ctx := range contexts {
		rawContexts[i] = string(ctx)
	}
	cfg := &config.Config{CustomCommands: []config.CustomCommandConfig{{
		Key: key, Name: name, Command: command, Contexts: rawContexts,
	}}}
	commands, err := cfg.ResolveCustomCommands()
	if err != nil {
		t.Fatalf("resolve command: %v", err)
	}
	return commands[0]
}

func TestBuildCommandData_SingleScopeFlat(t *testing.T) {
	t.Parallel()
	app := newTestApp()
	app.issuesList.SetIssues([]jira.Issue{{Key: "ABC-1", Summary: "hi; $(touch /tmp/pwned)"}})
	rc := resolvedCommand(t, "y", "test", "printf '%s|%s' {{.Key}} {{.Summary}}", config.CtxIssues)
	data, ok := app.buildCommandData(rc)
	if !ok {
		t.Fatal("expected ok")
	}
	rendered, err := rc.Render(data)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if got := rendered; got != "printf '%s|%s' 'ABC-1' 'hi; $(touch /tmp/pwned)'" {
		t.Errorf("got %q", got)
	}
}

func TestBuildCommandData_DetailCommentsFlat(t *testing.T) {
	t.Parallel()
	app := newTestApp()
	issue := jira.Issue{Key: "ABC-1", Comments: []jira.Comment{{ID: "10", Body: "hello"}}}
	app.issuesList.SetIssues([]jira.Issue{issue})
	app.detailView.SetIssue(&issue)
	app.detailView.SetActiveTab(views.TabComments)

	rc := resolvedCommand(t, "c", "test", "printf '%s-%s' {{.Key}} {{.CommentID}}", config.CtxDetailComments)
	data, ok := app.buildCommandData(rc)
	if !ok {
		t.Fatal("expected ok")
	}
	rendered, err := rc.Render(data)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if got := rendered; got != "printf '%s-%s' 'ABC-1' '10'" {
		t.Errorf("got %q", got)
	}
}

func TestBuildCommandData_SharedFieldsProjectScope(t *testing.T) {
	t.Parallel()
	app := newTestApp()
	app.gitBranch = "feature/x"
	app.gitRepoPath = "/tmp/repo"
	app.projectList.SetProjects([]jira.Project{{Key: "P", Name: "Proj"}})
	app.side = sideLeft
	app.leftFocus = focusProjects

	rc := resolvedCommand(t, "p", "test", "printf '%s|%s|%s|%s' {{.ProjectKey}} {{.JiraHost}} {{.GitBranch}} {{.GitRepoPath}}", config.CtxProjects)
	data, ok := app.buildCommandData(rc)
	if !ok {
		t.Fatal("expected ok")
	}
	rendered, err := rc.Render(data)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if got, want := rendered, "printf '%s|%s|%s|%s' 'P' 'example.atlassian.net' 'feature/x' '/tmp/repo'"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuildCommandData_SharedFieldsDetailComments(t *testing.T) {
	t.Parallel()
	app := newTestApp()
	app.gitBranch = "feature/y"
	app.gitRepoPath = "/tmp/repo2"
	issue := jira.Issue{Key: "ABC-1", Comments: []jira.Comment{{ID: "10", Body: "hello"}}}
	app.issuesList.SetIssues([]jira.Issue{issue})
	app.detailView.SetIssue(&issue)
	app.detailView.SetActiveTab(views.TabComments)

	rc := resolvedCommand(t, "c", "test", "printf '%s|%s|%s|%s|%s' {{.Key}} {{.CommentID}} {{.JiraHost}} {{.GitBranch}} {{.GitRepoPath}}", config.CtxDetailComments)
	data, ok := app.buildCommandData(rc)
	if !ok {
		t.Fatal("expected ok")
	}
	rendered, err := rc.Render(data)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if got, want := rendered, "printf '%s|%s|%s|%s|%s' 'ABC-1' '10' 'example.atlassian.net' 'feature/y' '/tmp/repo2'"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuildCommandData_MissingSelectionSwallows(t *testing.T) {
	t.Parallel()
	app := newTestApp()
	app.side = sideLeft
	app.leftFocus = focusProjects
	rc := resolvedCommand(t, "n", "test", "echo {{.ProjectKey}}", config.CtxProjects)
	if _, ok := app.buildCommandData(rc); ok {
		t.Error("expected ok=false with no selected project")
	}
}

func TestHandleCustomCommand_SpecificityDispatch(t *testing.T) {
	t.Parallel()
	app := newTestApp()
	issue := jira.Issue{Key: "ABC-1", Comments: []jira.Comment{{ID: "9", Body: "b"}}}
	app.issuesList.SetIssues([]jira.Issue{issue})
	app.detailView.SetIssue(&issue)
	app.detailView.SetActiveTab(views.TabComments)
	app.side = sideRight

	detailCmd := resolvedCommand(t, "x", "detail-one", "echo detail", config.CtxDetail)
	commentsCmd := resolvedCommand(t, "x", "comments-one", "echo comments", config.CtxDetailComments)
	app.customCmds = []config.ResolvedCustomCommand{detailCmd, commentsCmd}

	ctxs := app.activeContexts()
	if len(ctxs) == 0 || ctxs[0] != config.CtxDetailComments {
		t.Fatalf("activeContexts = %v, want detail.comments first", ctxs)
	}
	var chosen string
	for _, ctx := range ctxs {
		for _, rc := range app.customCmds {
			if rc.Key == "x" && rc.HasContext(ctx) {
				chosen = rc.Name
				goto done
			}
		}
	}
done:
	if chosen != "comments-one" {
		t.Errorf("chose %q, want comments-one", chosen)
	}
}

func TestQuitReachableWarning(t *testing.T) {
	t.Parallel()
	allCtxs := []config.Context{
		config.CtxIssues,
		config.CtxInfo,
		config.CtxProjects,
		config.CtxDetail,
		config.CtxDetailComments,
	}
	defaultKm := Keymap{ActQuit: {"q", "ctrl+c"}}

	cases := []struct {
		name     string
		keymap   Keymap
		cmds     []config.ResolvedCustomCommand
		wantWarn bool
	}{
		{
			name:     "no custom commands",
			keymap:   defaultKm,
			cmds:     nil,
			wantWarn: false,
		},
		{
			name:   "custom command on unrelated key",
			keymap: defaultKm,
			cmds: []config.ResolvedCustomCommand{
				{Key: "y", Contexts: allCtxs},
			},
			wantWarn: false,
		},
		{
			name:   "shadows q in issues only",
			keymap: defaultKm,
			cmds: []config.ResolvedCustomCommand{
				{Key: "q", Contexts: []config.Context{config.CtxIssues}},
			},
			wantWarn: false,
		},
		{
			name:   "shadows q and ctrl+c everywhere",
			keymap: defaultKm,
			cmds: []config.ResolvedCustomCommand{
				{Key: "q", Contexts: allCtxs},
				{Key: "ctrl+c", Contexts: allCtxs},
			},
			wantWarn: true,
		},
		{
			name:   "user remapped quit to alt+q, q shadowed everywhere",
			keymap: Keymap{ActQuit: {"alt+q"}},
			cmds: []config.ResolvedCustomCommand{
				{Key: "q", Contexts: allCtxs},
			},
			wantWarn: false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			warning := quitReachableWarning(testCase.keymap, testCase.cmds)
			got := warning != ""
			if got != testCase.wantWarn {
				t.Errorf("got warning=%q, want warn=%v", warning, testCase.wantWarn)
			}
		})
	}
}
