package tui

import (
	"context"
	"testing"
	"time"

	"github.com/nikbrunner/lazyjira/v2/pkg/config"
	"github.com/nikbrunner/lazyjira/v2/pkg/jira"
	"github.com/nikbrunner/lazyjira/v2/pkg/jira/jiratest"
	"github.com/nikbrunner/lazyjira/v2/pkg/tui/components"
)

func TestTTLCacheHitDoesNotExtendExpiry(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 9, 25, 10, 0, 0, 0, time.UTC)
	cache := newTTLCache[string](true, 5*time.Minute)
	cache.now = func() time.Time { return start }
	cache.set("key", "value")

	start = start.Add(4 * time.Minute)
	if got, ok := cache.get("key"); !ok || got != "value" {
		t.Fatalf("cache hit = (%q, %v), want (%q, true)", got, ok, "value")
	}

	start = start.Add(61 * time.Second)
	if got, ok := cache.get("key"); ok || got != "" {
		t.Fatalf("expired cache hit = (%q, %v), want empty miss", got, ok)
	}
}

func TestNewAppUsesCacheConfig(t *testing.T) {
	t.Parallel()
	cfg := config.DefaultConfig()
	cfg.Cache.TTL = "42s"
	app := NewAppWithAuth(cfg, &jiratest.FakeClient{T: t, SetOnRequestFunc: func(func(jira.RequestLog)) {}}, AuthDemo)
	if !app.sprintsCache.enabled || app.sprintsCache.ttl != 42*time.Second {
		t.Fatalf("sprint cache = enabled:%v ttl:%s, want enabled 42s", app.sprintsCache.enabled, app.sprintsCache.ttl)
	}

	cfg.Cache.Enabled = false
	disabled := NewAppWithAuth(cfg, &jiratest.FakeClient{T: t, SetOnRequestFunc: func(func(jira.RequestLog)) {}}, AuthDemo)
	if disabled.boardsCache.enabled || disabled.sprintsCache.enabled || disabled.usersCache.enabled || disabled.createMetaCache.enabled {
		t.Fatal("cache.enabled=false did not disable all reference caches")
	}
}

func TestTTLCacheDisabledAndClear(t *testing.T) {
	t.Parallel()
	cache := newTTLCache[string](false, time.Minute)
	cache.set("key", "value")
	if got, ok := cache.get("key"); ok || got != "" {
		t.Fatalf("disabled cache hit = (%q, %v), want empty miss", got, ok)
	}

	cache.enabled = true
	cache.set("key", "value")
	cache.clear()
	if got, ok := cache.get("key"); ok || got != "" {
		t.Fatalf("cleared cache hit = (%q, %v), want empty miss", got, ok)
	}
}

func TestReferenceRefreshRejectsPreRefreshUserAndMetadataResponses(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.projectKey = testProject
	app.createForm = components.NewCreateForm(nil)
	app.createCtx = createCtx{projectKey: testProject, issueTypeID: "1", issueTypeName: "Task"}
	oldVersion := app.referenceCacheVersion
	app.invalidateReferenceCaches()

	if model, cmd := app.handleUsersLoaded(usersLoadedMsg{users: []jira.User{{AccountID: "u1"}}, projectKey: testProject, cacheVersion: oldVersion}); model != app || cmd != nil {
		t.Fatal("stale user response returned an unexpected update")
	}
	if _, ok := app.usersCache.get(testProject); ok {
		t.Fatal("pre-refresh response repopulated the user cache")
	}
	app.usersCache.set(testProject, nil)
	if model, cmd := app.handleCreateMetaLoaded(createMetaLoadedMsg{
		fields:       []jira.CreateMetaField{{FieldID: "summary"}},
		projectKey:   testProject,
		issueTypeID:  "1",
		cacheVersion: oldVersion,
	}); model != app || cmd != nil {
		t.Fatal("stale create metadata response returned an unexpected update")
	}

	if _, ok := app.createMetaCache.get(testProject + ":1"); ok {
		t.Fatal("pre-refresh response repopulated the create metadata cache")
	}
}

func TestUserAndCreateMetadataCacheCallersHonorTTL(t *testing.T) {
	t.Parallel()
	userCalls, metaCalls := 0, 0
	fake := &jiratest.FakeClient{T: t}
	fake.GetUsersFunc = func(context.Context, string) ([]jira.User, error) {
		userCalls++
		return []jira.User{{AccountID: "u1"}}, nil
	}
	fake.GetCreateMetaFunc = func(context.Context, string, string) ([]jira.CreateMetaField, error) {
		metaCalls++
		return []jira.CreateMetaField{{FieldID: "summary"}}, nil
	}
	app := newAppWithFake(t, fake)
	app.projectKey = testProject
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})
	app.createForm = components.NewCreateForm(nil)

	now := time.Date(2026, 9, 25, 10, 0, 0, 0, time.UTC)
	app.usersCache.now = func() time.Time { return now }
	app.createMetaCache.now = func() time.Time { return now }
	app.usersCache.set(testProject, []jira.User{{AccountID: "u1"}})
	app.createMetaCache.set(testProject+":1", []jira.CreateMetaField{{FieldID: "summary"}})
	if users, ok := app.usersCache.get(testProject); !ok || len(users) != 1 {
		t.Fatal("seeded user cache entry is not readable")
	}
	if fields, ok := app.createMetaCache.get(testProject + ":1"); !ok || len(fields) != 1 {
		t.Fatal("seeded create metadata entry is not readable")
	}
	userExpiry := app.usersCache.entries[testProject].expiresAt
	metaExpiry := app.createMetaCache.entries[testProject+":1"].expiresAt
	now = now.Add(app.usersCache.ttl - time.Second)

	if _, _, handled := app.handleIssueAction(ActAssignee); !handled {
		t.Fatal("assignee action was not handled")
	}
	if app.usersCache.entries[testProject].expiresAt != userExpiry || userCalls != 0 {
		t.Fatal("user cache hit renewed the TTL or fetched users")
	}
	app.modal.Hide()
	app.handleModalCancelled()
	app.createCtx = createCtx{projectKey: testProject}

	_, cmd := app.handleCreateFormTypeSelected(components.CreateFormTypeSelectedMsg{TypeID: "1", TypeName: "Task"})
	if cmd != nil {
		t.Fatalf("create metadata cache lookup returned command %T", cmd)
	}
	if app.createMetaCache.entries[testProject+":1"].expiresAt != metaExpiry || metaCalls != 0 {
		t.Fatal("create metadata cache hit renewed the TTL or fetched metadata")
	}

	now = now.Add(time.Second)
	_, userCmd, handled := app.handleIssueAction(ActAssignee)
	if !handled {
		t.Fatal("expired assignee action was not handled")
	}
	if userCmd == nil {
		t.Fatal("expired user cache did not trigger a fetch")
	}
	userCmd()
	if userCalls != 1 {
		t.Fatalf("GetUsers calls = %d after expiry, want 1", userCalls)
	}

	_, metaCmd := app.handleCreateFormTypeSelected(components.CreateFormTypeSelectedMsg{TypeID: "1", TypeName: "Task"})
	if metaCmd == nil {
		t.Fatal("expired create metadata cache did not trigger a fetch")
	}
	metaCmd()
	if metaCalls != 1 {
		t.Fatalf("GetCreateMeta calls = %d after expiry, want 1", metaCalls)
	}

	app.usersCache.set(testProject, []jira.User{{AccountID: "u1"}})
	app.createMetaCache.set(testProject+":1", []jira.CreateMetaField{{FieldID: "summary"}})
	app.usersCache.enabled = false
	if _, userCmd, _ := app.handleIssueAction(ActAssignee); userCmd == nil {
		t.Fatal("disabled user cache did not trigger a fetch")
	} else {
		userCmd()
	}
	if userCalls != 2 {
		t.Fatalf("GetUsers calls with caching disabled = %d, want 2", userCalls)
	}
	app.createMetaCache.enabled = false
	if _, metaCmd := app.handleCreateFormTypeSelected(components.CreateFormTypeSelectedMsg{TypeID: "1", TypeName: "Task"}); metaCmd == nil {
		t.Fatal("disabled create metadata cache did not trigger a fetch")
	} else {
		metaCmd()
	}
	if metaCalls != 2 {
		t.Fatalf("GetCreateMeta calls with caching disabled = %d, want 2", metaCalls)
	}
}
