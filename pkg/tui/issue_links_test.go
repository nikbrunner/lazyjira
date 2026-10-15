package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nikbrunner/lazyjira/pkg/internal/testkit"
	"github.com/nikbrunner/lazyjira/pkg/jira"
	"github.com/nikbrunner/lazyjira/pkg/jira/jiratest"
	"github.com/nikbrunner/lazyjira/pkg/tui/components"
	"github.com/nikbrunner/lazyjira/pkg/tui/views"
)

var testBlocks = jira.IssueLinkType{Name: "Blocks", Inward: "is blocked by", Outward: "blocks"}

func TestLinkEnds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		outward     bool
		wantInward  string
		wantOutward string
	}{
		{"this blocks target: this is the outward issue", true, "PLAT-12", testKey},
		{"this is blocked by target: this is the inward issue", false, testKey, "PLAT-12"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			inward, outward := linkEnds(testKey, "PLAT-12", tt.outward)
			testkit.AssertEqual(t, "inward", inward, tt.wantInward)
			testkit.AssertEqual(t, "outward", outward, tt.wantOutward)
		})
	}
}

func TestLinkDirections_ListsBothDirectionsOnce(t *testing.T) {
	t.Parallel()
	dirs := linkDirections([]jira.IssueLinkType{testBlocks, {Name: "Relates", Inward: "relates to", Outward: "relates to"}})

	phrases := make([]string, 0, len(dirs))
	for _, d := range dirs {
		phrases = append(phrases, d.phrase)
	}
	testkit.AssertSliceEqual(t, "phrases", phrases, []string{"blocks", "is blocked by", "relates to"})
	testkit.AssertEqual(t, "blocks is outward", dirs[0].outward, true)
	testkit.AssertEqual(t, "is blocked by is inward", dirs[1].outward, false)
	testkit.AssertEqual(t, "type name", dirs[1].typeName, "Blocks")
}

func linksApp(t *testing.T, fake *jiratest.FakeClient, links ...jira.IssueLink) *App {
	t.Helper()
	app := editFlowApp(t, fake)
	app.overlays = components.OverlayStack{&app.confirmModal, &app.issueLookup, &app.modal}
	app.modal.SetSize(app.width, app.height)
	app.confirmModal.SetSize(app.width, app.height)
	app.projectKey = testProject
	app.linkTypesCache = newTTLCache[[]jira.IssueLinkType](true, cacheTTL("5m"))
	issue := &jira.Issue{Key: testKey, Summary: testSummary, IssueLinks: links}
	app.issueCache[testKey] = issue
	app.issuesList.SetIssues([]jira.Issue{*issue})
	app.infoPanel.SetIssue(issue)
	app.infoPanel.SetActiveTab(views.InfoTabLinks)
	app.infoPanel.Cursor = 0
	app.focusPane(focusInfo)
	return app
}

// send delivers msg and returns the message its command produces.
func send(app *App, msg tea.Msg) tea.Msg {
	_, cmd := app.Update(msg)
	if cmd == nil {
		return nil
	}
	return cmd()
}

func stubLinkTypes(fake *jiratest.FakeClient) {
	fake.GetIssueLinkTypesFunc = func(context.Context) ([]jira.IssueLinkType, error) {
		return []jira.IssueLinkType{testBlocks, {Name: "Relates", Inward: "relates to", Outward: "relates to"}}, nil
	}
}

func TestAddLink_PicksDirectionThenTargetAndRefetches(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		downs       int
		wantInward  string
		wantOutward string
	}{
		{"blocks", 0, "PLAT-12", testKey},
		{"is blocked by", 1, testKey, "PLAT-12"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fake := &jiratest.FakeClient{T: t}
			stubLinkTypes(fake)
			fake.SuggestIssuesFunc = func(context.Context, string) ([]jira.IssueSuggestion, error) { return nil, nil }
			fake.CreateIssueLinkFunc = func(context.Context, string, string, string) error { return nil }
			linked := &jira.Issue{Key: testKey, IssueLinks: []jira.IssueLink{{ID: "1", Type: &testBlocks, OutwardIssue: &jira.Issue{Key: "PLAT-12"}}}}
			stubFullIssueFetch(fake, linked)
			app := linksApp(t, fake)

			app.Update(send(app, keyRunes("n")))
			if !app.modal.IsVisible() {
				t.Fatal("link type picker not shown")
			}
			view := app.modal.View()
			for _, phrase := range []string{"blocks", "is blocked by", "relates to"} {
				if !strings.Contains(view, phrase) {
					t.Errorf("picker misses %q", phrase)
				}
			}
			testkit.AssertEqual(t, "relates to listed once", strings.Count(view, "relates to"), 1)

			for range tt.downs {
				send(app, tea.KeyMsg{Type: tea.KeyDown})
			}
			send(app, send(app, tea.KeyMsg{Type: tea.KeyEnter}))
			if !app.issueLookup.IsVisible() {
				t.Fatal("issue lookup not shown after picking a direction")
			}
			send(app, keyRunes("12"))
			created := send(app, send(app, tea.KeyMsg{Type: tea.KeyEnter}))

			if len(fake.CreateIssueLinkCalls) != 1 {
				t.Fatalf("CreateIssueLink calls = %d, want 1", len(fake.CreateIssueLinkCalls))
			}
			call := fake.CreateIssueLinkCalls[0]
			testkit.AssertEqual(t, "type", call.TypeName, "Blocks")
			testkit.AssertEqual(t, "inward", call.InwardKey, tt.wantInward)
			testkit.AssertEqual(t, "outward", call.OutwardKey, tt.wantOutward)

			app.Update(send(app, created))
			if len(fake.GetIssueCalls) == 0 || fake.GetIssueCalls[0].Key != testKey {
				t.Fatalf("GetIssue calls = %+v, want a refetch of %s", fake.GetIssueCalls, testKey)
			}
			testkit.AssertEqual(t, "listed link", app.infoPanel.SelectedLinkKey(), "PLAT-12")
		})
	}
}

func TestAddLink_CachesLinkTypes(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	stubLinkTypes(fake)
	app := linksApp(t, fake)

	app.Update(send(app, keyRunes("n")))
	send(app, tea.KeyMsg{Type: tea.KeyEsc})
	send(app, keyRunes("n"))

	testkit.AssertEqual(t, "link type requests", len(fake.GetIssueLinkTypesCalls), 1)
	testkit.AssertEqual(t, "picker visible", app.modal.IsVisible(), true)
}

func TestAddLink_RejectsSelfLink(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := linksApp(t, fake)
	app.linkLookup = &pendingLink{issueKey: testKey, dir: linkDirection{typeName: "Blocks", phrase: "blocks", outward: true}}

	msg := app.handleIssueLookupSelected(strings.ToLower(testKey))()

	if _, ok := msg.(errorMsg); !ok {
		t.Fatalf("msg = %T, want errorMsg", msg)
	}
	testkit.AssertEqual(t, "create calls", len(fake.CreateIssueLinkCalls), 0)
}

func TestAddLink_LookupCancelDropsPendingLink(t *testing.T) {
	t.Parallel()
	app := linksApp(t, &jiratest.FakeClient{T: t})
	app.linkLookup = &pendingLink{issueKey: testKey}

	app.Update(components.IssueLookupCancelledMsg{})

	if app.linkLookup != nil {
		t.Error("pending link kept after cancel")
	}
}

func TestIssueLinkErrors_Surface(t *testing.T) {
	t.Parallel()
	boom := errors.New("boom")
	tests := []struct {
		name string
		run  func(app *App, fake *jiratest.FakeClient) tea.Msg
	}{
		{"link types", func(app *App, fake *jiratest.FakeClient) tea.Msg {
			fake.GetIssueLinkTypesFunc = func(context.Context) ([]jira.IssueLinkType, error) { return nil, boom }
			return send(app, keyRunes("n"))
		}},
		{"create", func(app *App, fake *jiratest.FakeClient) tea.Msg {
			fake.CreateIssueLinkFunc = func(context.Context, string, string, string) error { return boom }
			app.linkLookup = &pendingLink{issueKey: testKey, dir: linkDirection{typeName: "Blocks", outward: true}}
			return app.handleIssueLookupSelected("PLAT-12")()
		}},
		{"delete", func(app *App, fake *jiratest.FakeClient) tea.Msg {
			fake.DeleteIssueLinkFunc = func(context.Context, string) error { return boom }
			send(app, keyRunes("d"))
			return send(app, send(app, keyRunes("y")))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fake := &jiratest.FakeClient{T: t}
			app := linksApp(t, fake, jira.IssueLink{ID: "7", Type: &testBlocks, OutwardIssue: &jira.Issue{Key: "PLAT-12"}})

			msg := tt.run(app, fake)

			emsg, ok := msg.(errorMsg)
			if !ok || !errors.Is(emsg.err, boom) {
				t.Fatalf("msg = %#v, want errorMsg wrapping boom", msg)
			}
			app.Update(msg)
			testkit.AssertEqual(t, "error modal", app.modal.IsVisible(), true)
		})
	}
}

func TestRemoveLink_AsksFirst(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		answer      tea.KeyMsg
		wantDeletes int
	}{
		{"accept", keyRunes("y"), 1},
		{"cancel", keyRunes("n"), 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fake := &jiratest.FakeClient{T: t}
			fake.DeleteIssueLinkFunc = func(context.Context, string) error { return nil }
			stubFullIssueFetch(fake, &jira.Issue{Key: testKey})
			app := linksApp(t, fake,
				jira.IssueLink{ID: "7", Type: &testBlocks, OutwardIssue: &jira.Issue{Key: "PLAT-12"}},
				jira.IssueLink{ID: "8", Type: &testBlocks, InwardIssue: &jira.Issue{Key: "PLAT-3"}},
			)
			app.infoPanel.Cursor = 1

			send(app, keyRunes("d"))
			if !app.confirmModal.IsVisible() {
				t.Fatal("no confirmation before removing a link")
			}
			if view := app.confirmModal.View(); !strings.Contains(view, "Remove link: is blocked by PLAT-3?") {
				t.Errorf("confirmation = %q", view)
			}
			testkit.AssertEqual(t, "deletes before answer", len(fake.DeleteIssueLinkCalls), 0)

			if updated := send(app, send(app, tt.answer)); updated != nil {
				app.Update(send(app, updated))
			}

			testkit.AssertEqual(t, "deletes", len(fake.DeleteIssueLinkCalls), tt.wantDeletes)
			if tt.wantDeletes == 1 {
				testkit.AssertEqual(t, "deleted id", fake.DeleteIssueLinkCalls[0], "8")
				testkit.AssertEqual(t, "refetched", fake.GetIssueCalls[0].Key, testKey)
				testkit.AssertEqual(t, "links after refetch", app.infoPanel.SelectedLinkKey(), "")
			}
		})
	}
}

func TestLinkKeys_OnlyOnLinksTab(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		tab  views.InfoPanelTab
	}{
		{"info", views.InfoTabFields},
		{"sub", views.InfoTabSubtasks},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fake := &jiratest.FakeClient{T: t}
			app := linksApp(t, fake, jira.IssueLink{ID: "7", Type: &testBlocks, OutwardIssue: &jira.Issue{Key: "PLAT-12"}})
			app.infoPanel.SetActiveTab(tt.tab)

			if msg := send(app, keyRunes("n")); msg != nil {
				t.Errorf("n produced %T", msg)
			}
			send(app, keyRunes("d"))

			testkit.AssertEqual(t, "picker", app.modal.IsVisible(), false)
			testkit.AssertEqual(t, "confirmation", app.confirmModal.IsVisible(), false)
			testkit.AssertEqual(t, "link type requests", len(fake.GetIssueLinkTypesCalls), 0)
		})
	}
}

func TestRemoveLink_NoLinksDoesNothing(t *testing.T) {
	t.Parallel()
	app := linksApp(t, &jiratest.FakeClient{T: t})

	send(app, keyRunes("d"))

	testkit.AssertEqual(t, "confirmation", app.confirmModal.IsVisible(), false)
}

func TestLinkHelp_OnlyOnLinksTab(t *testing.T) {
	t.Parallel()
	app := linksApp(t, &jiratest.FakeClient{T: t})

	hasLinkHelp := func() bool {
		for _, b := range app.ContextBindings() {
			if b.Description == "remove link" {
				return true
			}
		}
		return false
	}

	testkit.AssertEqual(t, "help on Lnk tab", hasLinkHelp(), true)
	app.infoPanel.SetActiveTab(views.InfoTabFields)
	testkit.AssertEqual(t, "help on Info tab", hasLinkHelp(), false)
}
