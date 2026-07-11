package commands

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jjuanrivvera/wootctl/internal/api"
	"github.com/jjuanrivvera/wootctl/internal/auth"
	"github.com/jjuanrivvera/wootctl/internal/config"
)

// labelsKind is the registry entry used across these tests.
func labelsKind(t *testing.T) portableKind {
	t.Helper()
	k, ok := portableKindByName("labels")
	require.True(t, ok)
	return k
}

func clientFor(t *testing.T, h http.HandlerFunc) *api.Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c := api.New(srv.URL, "t", api.WithRateLimit(1000))
	c.AccountID = "1"
	return c
}

func TestReconcile_CreateUpdateSkip(t *testing.T) {
	c := clientFor(t, jsonHandler(`{"payload":[{"id":1,"title":"keep","color":"#111"},{"id":2,"title":"change","color":"#222"}]}`))
	desired := []map[string]any{
		{"title": "keep", "color": "#111"},   // identical → skip
		{"title": "change", "color": "#999"}, // color differs → update id 2
		{"title": "brand-new", "color": "#0f0"},
	}
	plan, err := reconcile(t.Context(), c, labelsKind(t), desired, false, func(string) {})
	require.NoError(t, err)

	byKey := map[string]planItem{}
	for _, it := range plan {
		byKey[it.Key] = it
	}
	assert.Equal(t, actSkip, byKey["keep"].Action)
	assert.Equal(t, actUpdate, byKey["change"].Action)
	assert.Equal(t, "2", byKey["change"].id)
	assert.Equal(t, actCreate, byKey["brand-new"].Action)
	// No prune requested → no prune items even though nothing else matches.
	for _, it := range plan {
		assert.NotEqual(t, actPrune, it.Action)
	}
}

func TestReconcile_PruneRemovesDrift(t *testing.T) {
	c := clientFor(t, jsonHandler(`{"payload":[{"id":1,"title":"keep"},{"id":9,"title":"orphan"}]}`))
	desired := []map[string]any{{"title": "keep"}}
	plan, err := reconcile(t.Context(), c, labelsKind(t), desired, true, func(string) {})
	require.NoError(t, err)
	var pruned []string
	for _, it := range plan {
		if it.Action == actPrune {
			pruned = append(pruned, it.Key+"="+it.id)
		}
	}
	assert.Equal(t, []string{"orphan=9"}, pruned)
}

func TestReconcile_DuplicateKeySkippedNeverPruned(t *testing.T) {
	// Two live labels share the title "dup" — ambiguous. Must warn, and NEVER prune either
	// (the --prune-deletes-the-wrong-thing bug, GOAL.md §3c).
	c := clientFor(t, jsonHandler(`{"payload":[{"id":1,"title":"dup"},{"id":2,"title":"dup"},{"id":3,"title":"solo"}]}`))
	desired := []map[string]any{{"title": "solo"}}
	var warnings []string
	plan, err := reconcile(t.Context(), c, labelsKind(t), desired, true, func(s string) { warnings = append(warnings, s) })
	require.NoError(t, err)
	for _, it := range plan {
		assert.NotEqual(t, "dup", it.Key, "ambiguous key must never be acted on")
	}
	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "ambiguous")
}

func TestReconcile_DesiredMatchingDuplicateLiveIsNotCreated(t *testing.T) {
	// Desired wants "dup" too, but it's ambiguous live → skip entirely (don't create a third).
	c := clientFor(t, jsonHandler(`{"payload":[{"id":1,"title":"dup"},{"id":2,"title":"dup"}]}`))
	desired := []map[string]any{{"title": "dup", "color": "#000"}}
	plan, err := reconcile(t.Context(), c, labelsKind(t), desired, false, func(string) {})
	require.NoError(t, err)
	assert.Empty(t, plan, "ambiguous desired match must be a no-op, not a create")
}

func TestWritableEqual_IgnoresServerFieldsAndAbsentDesired(t *testing.T) {
	k := labelsKind(t)
	live := map[string]any{"id": 5.0, "title": "x", "color": "#abc", "updated_at": "2026", "created_at": "2025"}
	// Desired omits color entirely → treated as leave-as-is, so still equal.
	assert.True(t, writableEqual(live, map[string]any{"title": "x"}, k.Writable))
	// Desired changes color → not equal.
	assert.False(t, writableEqual(live, map[string]any{"title": "x", "color": "#zzz"}, k.Writable))
	// Server-only fields (updated_at) never enter the comparison.
	assert.True(t, writableEqual(live, map[string]any{"title": "x", "color": "#abc"}, k.Writable))
}

func TestWritableEqual_ArrayFieldCanonicalized(t *testing.T) {
	k, _ := portableKindByName("webhooks")
	live := map[string]any{"url": "https://x", "subscriptions": []any{"message_created", "conversation_created"}}
	same := map[string]any{"url": "https://x", "subscriptions": []any{"message_created", "conversation_created"}}
	diff := map[string]any{"url": "https://x", "subscriptions": []any{"message_created"}}
	assert.True(t, writableEqual(live, same, k.Writable))
	assert.False(t, writableEqual(live, diff, k.Writable))
}

func TestSelectKinds(t *testing.T) {
	all, err := selectKinds(nil)
	require.NoError(t, err)
	assert.Len(t, all, len(portableKinds))

	sub, err := selectKinds([]string{"labels", "webhooks"})
	require.NoError(t, err)
	require.Len(t, sub, 2)
	assert.Equal(t, "labels", sub[0].Name) // registry order preserved

	_, err = selectKinds([]string{"conversations"})
	require.Error(t, err, "non-config kinds are not portable")
}

func TestBackup_RestoreRoundTrip(t *testing.T) {
	e := newEnv(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/labels") {
			_, _ = w.Write([]byte(`{"payload":[{"id":1,"title":"vip","color":"#00f","show_on_sidebar":true,"updated_at":"ignore-me"}]}`))
			return
		}
		_, _ = w.Write([]byte(`[]`))
	})
	dir := t.TempDir()
	_, errOut, err := e.run("backup", "--dir", dir, "--only", "labels")
	require.NoError(t, err)
	assert.Contains(t, errOut, "backed up 1 labels")

	// The file holds only writable fields — server noise dropped.
	data, err := os.ReadFile(filepath.Join(dir, "labels.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "vip")
	assert.NotContains(t, string(data), "updated_at")
	assert.NotContains(t, string(data), "id:")

	// Restoring the same dir against the same (identical) live state → all unchanged.
	_, errOut, err = e.run("restore", "--dir", dir, "--only", "labels")
	require.NoError(t, err)
	assert.Contains(t, errOut, "0 created, 0 updated, 0 pruned, 1 unchanged")
}

func TestRestore_CreatesAndDryRunMakesNoWrites(t *testing.T) {
	var writes int
	h, log := recordingHandler(t, func(r *http.Request) (int, string) {
		if r.Method != http.MethodGet {
			writes++
		}
		if r.Method == http.MethodGet {
			return 200, `{"payload":[]}` // empty account → everything is a create
		}
		return 200, `{"payload":{"id":1}}`
	})
	e := newEnv(t, h)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "labels.yaml"),
		[]byte("- title: vip\n  color: \"#00f\"\n- title: urgent\n"), 0o600))

	// Dry-run: plan shows 2 creates, but no write requests hit the server.
	out, _, err := e.run("restore", "--dir", dir, "--only", "labels", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "labels: +2 ~0 -0")
	assert.Contains(t, out, "+ create labels \"vip\"")
	assert.Zero(t, writes, "dry-run must not write")

	// Real run: two creates happen.
	_, _, err = e.run("restore", "--dir", dir, "--only", "labels")
	require.NoError(t, err)
	var posts int
	for _, rec := range *log {
		if rec.Method == http.MethodPost {
			posts++
		}
	}
	assert.Equal(t, 2, posts)
}

func TestRestore_RequiresDir(t *testing.T) {
	e := newEnv(t, jsonHandler(`[]`))
	_, _, err := e.run("restore")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--dir")
}

// TestSync_BetweenProfiles exercises the multi-instance payoff: two live instances, config
// copied source→target, honoring the target's own base/account (allowGlobals=false).
func TestSync_BetweenProfiles(t *testing.T) {
	// Source instance: one label "shared".
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"payload":[{"id":1,"title":"shared","color":"#123"}]}`))
	}))
	defer source.Close()

	// Target instance: empty → expect a create of "shared".
	var mu sync.Mutex
	var targetWrites []recorded
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"payload":[]}`))
			return
		}
		raw, _ := io.ReadAll(r.Body)
		mu.Lock()
		targetWrites = append(targetWrites, recorded{Method: r.Method, Path: r.URL.Path, Body: string(raw)})
		mu.Unlock()
		_, _ = w.Write([]byte(`{"payload":{"id":5}}`))
	}))
	defer target.Close()

	// Two profiles in an isolated config; env points the ACTIVE profile at source.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("WOOTCTL_BASE_URL", "")
	t.Setenv("WOOTCTL_ACCOUNT_ID", "")
	t.Setenv("WOOTCTL_API_KEY", "")
	t.Setenv("WOOTCTL_PROFILE", "")
	t.Setenv("NO_COLOR", "1")

	cfg, err := config.Load()
	require.NoError(t, err)
	require.NoError(t, cfg.SetProfile("default", config.Profile{BaseURL: source.URL, AccountID: "1"}))
	require.NoError(t, cfg.SetProfile("acue", config.Profile{BaseURL: target.URL, AccountID: "1"}))
	cfg.CurrentProfile = "default"
	require.NoError(t, cfg.Save())

	store := newFakeStore()
	require.NoError(t, store.Set("default", "src-token"))
	require.NoError(t, store.Set("acue", "dst-token"))

	run := func(args ...string) (string, string, error) {
		d := newDeps()
		d.store = func() auth.Store { return store }
		return runWithDeps(t, d, args...)
	}

	// Dry-run first: plan a create on the target, no writes.
	out, _, err := run("sync", "--to", "acue", "--only", "labels", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "labels: +1 ~0 -0")
	mu.Lock()
	assert.Empty(t, targetWrites, "dry-run must not write to the target")
	mu.Unlock()

	// Real sync: the label is created on the TARGET (its base/account), not the source.
	_, errOut, err := run("sync", "--to", "acue", "--only", "labels")
	require.NoError(t, err)
	assert.Contains(t, errOut, "1 created")
	mu.Lock()
	defer mu.Unlock()
	require.Len(t, targetWrites, 1)
	assert.Equal(t, http.MethodPost, targetWrites[0].Method)
	assert.Contains(t, targetWrites[0].Body, `"title":"shared"`)
	assert.Contains(t, targetWrites[0].Body, `"color":"#123"`)
}

func TestSync_RejectsSameProfile(t *testing.T) {
	e := newEnv(t, jsonHandler(`[]`))
	_, _, err := e.run("sync", "--to", "default")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "different instances")
}

func TestSync_UnknownTargetProfile(t *testing.T) {
	e := newEnv(t, jsonHandler(`[]`))
	_, _, err := e.run("sync", "--to", "ghost")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no such profile")
}
