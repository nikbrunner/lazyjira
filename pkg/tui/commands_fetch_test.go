package tui

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nikbrunner/lazyjira/v2/pkg/jira"
	"github.com/nikbrunner/lazyjira/v2/pkg/jira/jiratest"
)

func assertErrorMsg(t *testing.T, msg tea.Msg) {
	t.Helper()
	if _, ok := msg.(errorMsg); !ok {
		t.Fatalf("msg = %T, want errorMsg", msg)
	}
}

func newFakeClient(t *testing.T) *jiratest.FakeClient {
	t.Helper()
	return &jiratest.FakeClient{T: t}
}

func TestFetchIssuesByJQL(t *testing.T) {
	t.Parallel()

	t.Run("success tags the tab index", func(t *testing.T) {
		t.Parallel()
		fake := &jiratest.FakeClient{T: t}
		fake.SearchIssuesFunc = func(_ context.Context, _ string, _, _ int) (*jira.SearchResult, error) {
			return &jira.SearchResult{Issues: []jira.Issue{{Key: testKey}}}, nil
		}

		msg := fetchIssuesByJQL(fake, "project = PLAT", 2, 50)()

		loaded, ok := msg.(issuesLoadedMsg)
		if !ok {
			t.Fatalf("msg = %T, want issuesLoadedMsg", msg)
		}
		if loaded.tab != 2 || len(loaded.issues) != 1 {
			t.Errorf("loaded = %+v, want tab=2 with one issue", loaded)
		}
	})

	t.Run("error returns errorMsg", func(t *testing.T) {
		t.Parallel()
		fake := newFakeClient(t)
		fake.SearchIssuesFunc = func(_ context.Context, _ string, _, _ int) (*jira.SearchResult, error) {
			return nil, errors.New("boom")
		}
		assertErrorMsg(t, fetchIssuesByJQL(fake, "x", 0, 10)())
	})
}

func TestFetchJQLSearch(t *testing.T) {
	t.Parallel()

	t.Run("success carries jql and issues", func(t *testing.T) {
		t.Parallel()
		fake := &jiratest.FakeClient{T: t}
		fake.SearchIssuesFunc = func(_ context.Context, _ string, _, _ int) (*jira.SearchResult, error) {
			return &jira.SearchResult{Issues: []jira.Issue{{Key: testKey}}}, nil
		}

		msg := fetchJQLSearch(fake, "project = PLAT", 25)()

		result, ok := msg.(jqlSearchResultMsg)
		if !ok {
			t.Fatalf("msg = %T, want jqlSearchResultMsg", msg)
		}
		if result.jql != "project = PLAT" || len(result.issues) != 1 {
			t.Errorf("result = %+v", result)
		}
	})

	t.Run("error returns jqlSearchErrorMsg", func(t *testing.T) {
		t.Parallel()
		fake := newFakeClient(t)
		fake.SearchIssuesFunc = func(_ context.Context, _ string, _, _ int) (*jira.SearchResult, error) {
			return nil, errors.New("bad jql")
		}
		if _, ok := fetchJQLSearch(fake, "x", 10)().(jqlSearchErrorMsg); !ok {
			t.Error("want jqlSearchErrorMsg")
		}
	})
}

func TestFetchChildren_TagsKeyAndEpoch(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetChildrenFunc = func(_ context.Context, _ string) ([]jira.Issue, error) {
		return []jira.Issue{{Key: "SUB-1"}}, nil
	}

	msg := fetchChildren(fake, testKey, 7)()

	loaded, ok := msg.(childrenLoadedMsg)
	if !ok {
		t.Fatalf("msg = %T, want childrenLoadedMsg", msg)
	}
	if loaded.key != testKey || loaded.epoch != 7 || loaded.err != nil || len(loaded.issues) != 1 {
		t.Errorf("loaded = %+v", loaded)
	}
}

func TestFetchJQLAutocompleteData(t *testing.T) {
	t.Parallel()

	t.Run("success returns fields", func(t *testing.T) {
		t.Parallel()
		fake := &jiratest.FakeClient{T: t}
		fake.GetJQLAutocompleteDataFunc = func(context.Context) ([]jira.AutocompleteField, error) {
			return []jira.AutocompleteField{{Value: "status"}}, nil
		}
		if _, ok := fetchJQLAutocompleteData(fake)().(jqlFieldsLoadedMsg); !ok {
			t.Error("want jqlFieldsLoadedMsg")
		}
	})

	t.Run("error is swallowed to nil", func(t *testing.T) {
		t.Parallel()
		fake := newFakeClient(t)
		fake.GetJQLAutocompleteDataFunc = func(context.Context) ([]jira.AutocompleteField, error) {
			return nil, errors.New("nope")
		}
		if msg := fetchJQLAutocompleteData(fake)(); msg != nil {
			t.Errorf("msg = %T, want nil", msg)
		}
	})
}

func TestFetchJQLSuggestions(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetJQLAutocompleteSuggestionsFunc = func(_ context.Context, _, _ string) ([]jira.AutocompleteSuggestion, error) {
		return []jira.AutocompleteSuggestion{{Value: "Done"}}, nil
	}

	if _, ok := fetchJQLSuggestions(fake, "status", "Do")().(jqlSuggestionsMsg); !ok {
		t.Error("want jqlSuggestionsMsg")
	}
}

func TestUpdateIssueField(t *testing.T) {
	t.Parallel()

	t.Run("success returns issueUpdatedMsg with field", func(t *testing.T) {
		t.Parallel()
		fake := &jiratest.FakeClient{T: t}
		fake.UpdateIssueFunc = func(_ context.Context, _ string, _ map[string]any) error { return nil }

		msg := updateIssueField(fake, testKey, testSummary, "new")()

		updated, ok := msg.(issueUpdatedMsg)
		if !ok {
			t.Fatalf("msg = %T, want issueUpdatedMsg", msg)
		}
		if updated.issueKey != testKey || updated.field != testSummary {
			t.Errorf("updated = %+v", updated)
		}
	})

	t.Run("error returns errorMsg", func(t *testing.T) {
		t.Parallel()
		fake := newFakeClient(t)
		fake.UpdateIssueFunc = func(_ context.Context, _ string, _ map[string]any) error { return errors.New("x") }
		assertErrorMsg(t, updateIssueField(fake, testKey, testSummary, "new")())
	})
}

func TestRemoveIssueParent(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.RemoveIssueParentFunc = func(_ context.Context, _ string) error { return nil }

	msg := removeIssueParent(fake, testKey)()

	updated, ok := msg.(issueUpdatedMsg)
	if !ok {
		t.Fatalf("msg = %T, want issueUpdatedMsg", msg)
	}
	if updated.field != "parent" {
		t.Errorf("field = %q, want parent", updated.field)
	}
}

func TestAddComment(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.AddCommentFunc = func(_ context.Context, _ string, _ any) (*jira.Comment, error) {
		return &jira.Comment{ID: "1"}, nil
	}

	if _, ok := addComment(fake, testKey, "hi")().(commentAddedMsg); !ok {
		t.Error("want commentAddedMsg")
	}
}

func TestUpdateComment(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.UpdateCommentFunc = func(_ context.Context, _, _ string, _ any) error { return nil }

	if _, ok := updateComment(fake, testKey, "10", "edit")().(commentUpdatedMsg); !ok {
		t.Error("want commentUpdatedMsg")
	}
}

func TestFetchCreateMeta(t *testing.T) {
	t.Parallel()

	t.Run("success returns createMetaLoadedMsg", func(t *testing.T) {
		t.Parallel()
		fake := &jiratest.FakeClient{T: t}
		fake.GetCreateMetaFunc = func(_ context.Context, _, _ string) ([]jira.CreateMetaField, error) {
			return []jira.CreateMetaField{{FieldID: testSummary}}, nil
		}
		if _, ok := fetchCreateMeta(fake, testProject, "10001", 0)().(createMetaLoadedMsg); !ok {
			t.Error("want createMetaLoadedMsg")
		}
	})

	t.Run("error returns createPreFormErrorMsg", func(t *testing.T) {
		t.Parallel()
		fake := newFakeClient(t)
		fake.GetCreateMetaFunc = func(_ context.Context, _, _ string) ([]jira.CreateMetaField, error) {
			return nil, errors.New("x")
		}
		if _, ok := fetchCreateMeta(fake, testProject, "10001", 0)().(createPreFormErrorMsg); !ok {
			t.Error("want createPreFormErrorMsg")
		}
	})
}

func TestFetchCustomFieldOptions(t *testing.T) {
	t.Parallel()

	t.Run("found field carries options", func(t *testing.T) {
		t.Parallel()
		fake := &jiratest.FakeClient{T: t}
		fake.GetCreateMetaFunc = func(_ context.Context, _, _ string) ([]jira.CreateMetaField, error) {
			return []jira.CreateMetaField{{
				FieldID:       "customfield_1",
				Schema:        jira.CreateMetaSchema{Type: "option"},
				AllowedValues: []jira.CreateMetaValue{{ID: "1", Name: "A"}},
			}}, nil
		}

		msg := fetchCustomFieldOptions(fake, testProject, "10001", customFieldOptionsMsg{fieldID: "customfield_1"})()

		info, ok := msg.(customFieldOptionsMsg)
		if !ok {
			t.Fatalf("msg = %T, want customFieldOptionsMsg", msg)
		}
		if info.fieldNotFound || len(info.options) != 1 || info.schemaType != "option" {
			t.Errorf("info = %+v", info)
		}
	})

	t.Run("missing field sets fieldNotFound", func(t *testing.T) {
		t.Parallel()
		fake := &jiratest.FakeClient{T: t}
		fake.GetCreateMetaFunc = func(_ context.Context, _, _ string) ([]jira.CreateMetaField, error) {
			return []jira.CreateMetaField{{FieldID: "other"}}, nil
		}

		msg := fetchCustomFieldOptions(fake, testProject, "10001", customFieldOptionsMsg{fieldID: "customfield_1"})()

		info, ok := msg.(customFieldOptionsMsg)
		if !ok || !info.fieldNotFound {
			t.Errorf("msg = %#v, want fieldNotFound", msg)
		}
	})
}

func TestCreateIssue(t *testing.T) {
	t.Parallel()

	t.Run("success returns issueCreatedMsg", func(t *testing.T) {
		t.Parallel()
		fake := &jiratest.FakeClient{T: t}
		fake.CreateIssueFunc = func(_ context.Context, _ map[string]any) (*jira.Issue, error) {
			return &jira.Issue{Key: "PLAT-99"}, nil
		}

		msg := createIssue(fake, map[string]any{testSummary: "x"})()

		created, ok := msg.(issueCreatedMsg)
		if !ok {
			t.Fatalf("msg = %T, want issueCreatedMsg", msg)
		}
		if created.issue == nil || created.issue.Key != "PLAT-99" {
			t.Errorf("created = %+v", created)
		}
	})

	t.Run("error returns createErrorMsg", func(t *testing.T) {
		t.Parallel()
		fake := newFakeClient(t)
		fake.CreateIssueFunc = func(_ context.Context, _ map[string]any) (*jira.Issue, error) {
			return nil, errors.New("x")
		}
		if _, ok := createIssue(fake, nil)().(createErrorMsg); !ok {
			t.Error("want createErrorMsg")
		}
	})
}

func TestFetchMyself(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetMyselfFunc = func(context.Context) (*jira.User, error) {
		return &jira.User{AccountID: "me"}, nil
	}

	msg := fetchMyself(fake)()
	loaded, ok := msg.(myselfLoadedMsg)
	if !ok || loaded.user == nil || loaded.user.AccountID != "me" {
		t.Errorf("msg = %#v, want myselfLoadedMsg with user", msg)
	}
}

func TestFetchFieldDiscovery_WrapsError(t *testing.T) {
	t.Parallel()

	t.Run("nil error", func(t *testing.T) {
		t.Parallel()
		fake := &jiratest.FakeClient{T: t}
		fake.DiscoverFieldsFunc = func(context.Context) error { return nil }

		msg := fetchFieldDiscovery(fake)()
		discovered, ok := msg.(fieldsDiscoveredMsg)
		if !ok || discovered.err != nil {
			t.Errorf("msg = %#v, want fieldsDiscoveredMsg with nil err", msg)
		}
	})

	t.Run("propagates error", func(t *testing.T) {
		t.Parallel()
		fake := &jiratest.FakeClient{T: t}
		fake.DiscoverFieldsFunc = func(context.Context) error { return errors.New("no field") }

		msg := fetchFieldDiscovery(fake)()
		discovered, ok := msg.(fieldsDiscoveredMsg)
		if !ok || discovered.err == nil {
			t.Errorf("msg = %#v, want fieldsDiscoveredMsg with err", msg)
		}
	})
}

func TestFetchUsers_TagsIssueKey(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetUsersFunc = func(_ context.Context, _ string) ([]jira.User, error) {
		return []jira.User{{AccountID: "u1"}}, nil
	}

	msg := fetchUsers(fake, testProject, testKey, 0)()
	loaded, ok := msg.(usersLoadedMsg)
	if !ok || loaded.issueKey != testKey || len(loaded.users) != 1 {
		t.Errorf("msg = %#v", msg)
	}
}

func TestFetchComponents(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetComponentsFunc = func(_ context.Context, _ string) ([]jira.Component, error) {
		return []jira.Component{{Name: "backend"}}, nil
	}

	if _, ok := fetchComponents(fake, testProject)().(componentsLoadedMsg); !ok {
		t.Error("want componentsLoadedMsg")
	}
}

func TestFetchSprints(t *testing.T) {
	t.Parallel()

	t.Run("loads active and future sprints from Scrum boards only", func(t *testing.T) {
		t.Parallel()
		fake := &jiratest.FakeClient{T: t}
		fake.GetBoardsFunc = func(context.Context) ([]jira.Board, error) {
			return []jira.Board{
				{ID: 1, Name: "Kanban", Type: "kanban", ProjectKey: "WEBSDK"},
				{ID: 2, Name: "Cloud Platform Sprints", Type: "scrum", ProjectKey: "CP"},
				{ID: 3, Name: "Cloud Sprints", Type: "scrum", ProjectKey: "CP"},
			}, nil
		}
		fake.GetSprintsFunc = func(_ context.Context, boardID int) ([]jira.Sprint, error) {
			switch boardID {
			case 2:
				return []jira.Sprint{
					{ID: 10, Name: "Current", State: "active"},
					{ID: 11, Name: "Next", State: "future"},
					{ID: 12, Name: "Closed", State: "closed"},
				}, nil
			case 3:
				return []jira.Sprint{{ID: 10, Name: "Current", State: "active"}}, nil
			default:
				t.Fatalf("unexpected board %d", boardID)
				return nil, nil
			}
		}

		loaded, ok := fetchSprints(fake, sprintPickerTarget{issueKey: testKey})().(sprintsLoadedMsg)
		if !ok {
			t.Fatalf("message = %#v, want sprintsLoadedMsg", loaded)
		}
		if loaded.err != nil {
			t.Fatalf("loaded.err = %v", loaded.err)
		}
		if len(loaded.options) != 2 {
			t.Fatalf("len(options) = %d, want 2", len(loaded.options))
		}
		if got := loaded.options[0].boardNames; len(got) != 2 {
			t.Errorf("duplicate sprint board names = %v, want both source boards", got)
		}
		if got := fake.GetSprintsCalls; len(got) != 2 || got[0].BoardID != 2 || got[1].BoardID != 3 {
			t.Errorf("GetSprints calls = %#v, want only Scrum boards 2 and 3", got)
		}
	})

	t.Run("board discovery error is returned", func(t *testing.T) {
		t.Parallel()
		fake := newFakeClient(t)
		fake.GetBoardsFunc = func(context.Context) ([]jira.Board, error) {
			return nil, errors.New("boards unavailable")
		}

		loaded, ok := fetchSprints(fake, sprintPickerTarget{issueKey: testKey})().(sprintsLoadedMsg)
		if !ok || loaded.err == nil {
			t.Fatalf("message = %#v, want sprintsLoadedMsg with error", loaded)
		}
	})

	t.Run("sprint fetch error is returned", func(t *testing.T) {
		t.Parallel()
		fake := newFakeClient(t)
		fake.GetBoardsFunc = func(context.Context) ([]jira.Board, error) {
			return []jira.Board{{ID: 2, Type: "scrum"}}, nil
		}
		fake.GetSprintsFunc = func(context.Context, int) ([]jira.Sprint, error) {
			return nil, errors.New("sprints unavailable")
		}

		loaded, ok := fetchSprints(fake, sprintPickerTarget{issueKey: testKey})().(sprintsLoadedMsg)
		if !ok || loaded.err == nil {
			t.Fatalf("message = %#v, want sprintsLoadedMsg with error", loaded)
		}
	})
}

func TestMoveToSprint(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.MoveToSprintFunc = func(_ context.Context, _ int, _ string) error { return nil }

	if _, ok := moveToSprint(fake, 9, testKey)().(issueUpdatedMsg); !ok {
		t.Error("want issueUpdatedMsg")
	}
}

func TestFetchIssueTypes(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetIssueTypesFunc = func(_ context.Context, _ string) ([]jira.IssueType, error) {
		return []jira.IssueType{{Name: "Story"}}, nil
	}

	if _, ok := fetchIssueTypes(fake, "10000")().(issueTypesLoadedMsg); !ok {
		t.Error("want issueTypesLoadedMsg")
	}
}

func TestFetchIssueDetail(t *testing.T) {
	t.Parallel()

	t.Run("success enriches with comments and changelog", func(t *testing.T) {
		t.Parallel()
		fake := &jiratest.FakeClient{T: t}
		fake.GetIssueFunc = func(_ context.Context, _ string) (*jira.Issue, error) {
			return &jira.Issue{Key: testKey}, nil
		}
		fake.GetCommentsFunc = func(_ context.Context, _ string) ([]jira.Comment, error) {
			return []jira.Comment{{ID: "c1"}}, nil
		}
		fake.GetChangelogFunc = func(_ context.Context, _ string) ([]jira.ChangelogEntry, error) {
			return []jira.ChangelogEntry{{}}, nil
		}

		msg := fetchIssueDetail(fake, testKey)()
		loaded, ok := msg.(issueDetailLoadedMsg)
		if !ok {
			t.Fatalf("msg = %T, want issueDetailLoadedMsg", msg)
		}
		if len(loaded.issue.Comments) != 1 || len(loaded.issue.Changelog) != 1 {
			t.Errorf("issue not enriched: %+v", loaded.issue)
		}
	})

	t.Run("get issue error returns errorMsg", func(t *testing.T) {
		t.Parallel()
		fake := newFakeClient(t)
		fake.GetIssueFunc = func(_ context.Context, _ string) (*jira.Issue, error) {
			return nil, errors.New("404")
		}
		assertErrorMsg(t, fetchIssueDetail(fake, testKey)())
	})
}

func TestBatchPrefetch(t *testing.T) {
	t.Parallel()

	t.Run("success returns batchPrefetchedMsg", func(t *testing.T) {
		t.Parallel()
		var gotJQL string
		fake := &jiratest.FakeClient{T: t}
		fake.SearchIssuesFunc = func(_ context.Context, jql string, _, _ int) (*jira.SearchResult, error) {
			gotJQL = jql
			return &jira.SearchResult{Issues: []jira.Issue{{Key: "A-1"}, {Key: "B-2"}}}, nil
		}

		msg := batchPrefetch(fake, []string{"A-1", "B-2"})()
		if _, ok := msg.(batchPrefetchedMsg); !ok {
			t.Fatalf("msg = %T, want batchPrefetchedMsg", msg)
		}
		if gotJQL != "key in (A-1,B-2)" {
			t.Errorf("jql = %q, want key in (A-1,B-2)", gotJQL)
		}
	})

	t.Run("error is silent nil", func(t *testing.T) {
		t.Parallel()
		fake := newFakeClient(t)
		fake.SearchIssuesFunc = func(_ context.Context, _ string, _, _ int) (*jira.SearchResult, error) {
			return nil, errors.New("x")
		}
		if msg := batchPrefetch(fake, []string{"A-1"})(); msg != nil {
			t.Errorf("msg = %T, want nil", msg)
		}
	})
}

func TestPrefetchIssue_NilOnFailure(t *testing.T) {
	t.Parallel()
	fake := newFakeClient(t)
	fake.GetIssueFunc = func(_ context.Context, _ string) (*jira.Issue, error) {
		return nil, errors.New("gone")
	}

	if msg := prefetchIssue(fake, testKey)(); msg != nil {
		t.Errorf("msg = %T, want nil for failed prefetch", msg)
	}
}
