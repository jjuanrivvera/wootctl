package commands

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/jjuanrivvera/wootctl/internal/auth"
)

// fakeStore is an in-memory auth.Store so tests never touch a real OS keyring.
type fakeStore struct {
	mu   sync.Mutex
	data map[string]string
}

func newFakeStore() *fakeStore { return &fakeStore{data: map[string]string{}} }

func (f *fakeStore) Set(profile, token string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.data[profile] = token
	return nil
}

func (f *fakeStore) Get(profile string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if t, ok := f.data[profile]; ok && t != "" {
		return t, nil
	}
	return "", auth.ErrNotFound
}

func (f *fakeStore) Delete(profile string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.data, profile)
	return nil
}

func (f *fakeStore) Backend() string { return "fake" }

// env wires one test invocation: an httptest Chatwoot, an isolated config dir, a fake
// keyring, and env-provided credentials so getAPIClient resolves without real state.
type env struct {
	t     *testing.T
	srv   *httptest.Server
	store *fakeStore
}

// newEnv starts a mock Chatwoot server and isolates all state under t.TempDir().
func newEnv(t *testing.T, handler http.HandlerFunc) *env {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("WOOTCTL_BASE_URL", srv.URL)
	t.Setenv("WOOTCTL_ACCOUNT_ID", "1")
	t.Setenv("WOOTCTL_API_KEY", "test-token")
	t.Setenv("WOOTCTL_PLATFORM_TOKEN", "platform-token")
	t.Setenv("WOOTCTL_PROFILE", "")
	t.Setenv("NO_COLOR", "1")

	return &env{t: t, srv: srv, store: newFakeStore()}
}

// run executes the real command tree with captured output.
func (e *env) run(args ...string) (string, string, error) {
	e.t.Helper()
	d := newDeps()
	d.store = func() auth.Store { return e.store }
	return runWithDeps(e.t, d, args...)
}

// runWithDeps builds a fresh tree from d and runs it with captured output — used by tests
// that need custom deps (e.g. a two-profile sync).
func runWithDeps(t *testing.T, d *deps, args ...string) (string, string, error) {
	t.Helper()
	root := newRootCmd(d)
	root.SetArgs(args)
	var out, errB bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errB)
	err := root.ExecuteContext(t.Context())
	return out.String(), errB.String(), err
}

// jsonHandler answers every request with one body (helper for single-shot tests).
func jsonHandler(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}
}

// recordingHandler captures requests and replies per-route.
type recorded struct {
	Method string
	Path   string
	Query  string
	Body   string
}

func recordingHandler(t *testing.T, respond func(r *http.Request) (int, string)) (http.HandlerFunc, *[]recorded) {
	t.Helper()
	var mu sync.Mutex
	var log []recorded
	h := func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r.Body)
		mu.Lock()
		log = append(log, recorded{Method: r.Method, Path: r.URL.Path, Query: r.URL.RawQuery, Body: buf.String()})
		mu.Unlock()
		status, body := respond(r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
	return h, &log
}
