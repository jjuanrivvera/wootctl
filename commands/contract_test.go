package commands

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/stretchr/testify/require"

	"github.com/jjuanrivvera/wootctl/internal/auth"
)

// Contract test: every request wootctl builds for the application API must validate against
// Chatwoot's own OpenAPI spec (method, path, and request-body schema). This is the wire-level
// check the official CLI has and wootctl lacked — spec-check proves a command EXISTS; this proves
// the bytes it sends match the schema. The spec is vendored at internal/api/testdata (pinned to
// the manifest's swagger commit).

var (
	contractRouterOnce sync.Once
	contractRouter     routers.Router
	contractRouterErr  error
)

func loadContractRouter() (routers.Router, error) {
	contractRouterOnce.Do(func() {
		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromFile("../internal/api/testdata/application_swagger.json")
		if err != nil {
			contractRouterErr = err
			return
		}
		// Match on PATH only — the spec's server is app.chatwoot.com, but requests hit the
		// test server's host. A single "/" server makes routing host-agnostic.
		doc.Servers = openapi3.Servers{{URL: "/"}}
		contractRouter, contractRouterErr = gorillamux.NewRouter(doc)
	})
	return contractRouter, contractRouterErr
}

// contractServer is a test server that validates each incoming request against the spec,
// recording any violation, then replies with a benign enveloped record so the command finishes.
func contractServer(t *testing.T) (*httptest.Server, *[]string) {
	t.Helper()
	router, err := loadContractRouter()
	require.NoError(t, err, "load OpenAPI spec")

	var mu sync.Mutex
	var violations []string
	record := func(s string) { mu.Lock(); violations = append(violations, s); mu.Unlock() }

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(body))

		route, pathParams, ferr := router.FindRoute(r)
		if ferr != nil {
			// A path the application spec doesn't define. Reads of some nested sub-resources
			// (e.g. team_members) aren't in the application tag-group; only flag writes, which
			// are the schema-bearing calls this test targets.
			if r.Method != http.MethodGet {
				record("no spec route: " + r.Method + " " + r.URL.Path)
			}
		} else if r.Method != http.MethodGet {
			input := &openapi3filter.RequestValidationInput{
				Request:    r,
				PathParams: pathParams,
				Route:      route,
				Options:    &openapi3filter.Options{AuthenticationFunc: openapi3filter.NoopAuthenticationFunc},
			}
			if verr := openapi3filter.ValidateRequest(r.Context(), input); verr != nil {
				record(r.Method + " " + r.URL.Path + ": " + verr.Error())
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"payload":{"id":1}}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &violations
}

func TestContract_RequestsMatchOpenAPISpec(t *testing.T) {
	srv, violations := contractServer(t)

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("WOOTCTL_BASE_URL", srv.URL)
	t.Setenv("WOOTCTL_ACCOUNT_ID", "1")
	t.Setenv("WOOTCTL_API_KEY", "test-token")
	t.Setenv("WOOTCTL_PROFILE", "")
	t.Setenv("NO_COLOR", "1")

	store := newFakeStore()
	run := func(args ...string) {
		t.Helper()
		d := newDeps()
		d.store = func() auth.Store { return store }
		_, _, err := runWithDeps(t, d, args...)
		require.NoError(t, err, "%v", args)
	}

	// Battery of schema-bearing (write) requests across the application API.
	run("labels", "create", "--title", "vip", "--color", "#0055ff", "--show-on-sidebar")
	run("labels", "update", "1", "--description", "changed")
	run("canned-responses", "create", "--content", "Hello there", "--short-code", "greet")
	run("custom-attributes", "create", "--attribute-display-name", "Order", "--attribute-key", "order_id", "--attribute-model", "1", "--attribute-display-type", "0")
	run("custom-filters", "create", "--name", "open-urgent", "--type", "conversation", "--query", `{"payload":[]}`)
	run("automation-rules", "create", "--name", "auto", "--event-name", "conversation_created", "--active",
		"--conditions", `[{"attribute_key":"status","filter_operator":"equal_to","values":["open"],"query_operator":null}]`,
		"--actions", `[{"action_name":"assign_team","action_params":[1]}]`)
	run("teams", "create", "--name", "support", "--allow-auto-assign")
	run("teams", "update", "1", "--description", "front line")
	run("webhooks", "create", "--url", "https://example.com/hook", "--subscriptions", "message_created,conversation_created")
	run("contacts", "create", "--inbox-id", "1", "--name", "Ana", "--email", "ana@example.com")
	run("contacts", "update", "1", "--name", "Ana Maria")
	run("conversations", "toggle-status", "5", "--status", "resolved")
	run("conversations", "toggle-priority", "5", "--priority", "urgent")
	run("conversations", "assign", "5", "--assignee-id", "7")
	run("conversations", "add-labels", "5", "--labels", "vip,billing")
	run("messages", "create", "5", "--content", "On it")

	require.Empty(t, *violations, "OpenAPI contract violations:\n%s", joinLines(*violations))
}

// TestContract_HarnessActuallyValidates proves the contract server can FAIL — a
// deliberately schema-violating request must be recorded. A contract test that can't detect
// a violation is worthless (it would green-light real drift).
func TestContract_HarnessActuallyValidates(t *testing.T) {
	srv, violations := contractServer(t)
	client := srv.Client()

	// 1. color must be a string; send a number → schema violation.
	badType := `{"title":"x","color":12345}`
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/accounts/1/labels", bytes.NewBufferString(badType))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()

	// 2. a write to a path the spec doesn't define → no-route violation.
	req2, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/accounts/1/not_a_resource", bytes.NewBufferString(`{}`))
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := client.Do(req2)
	require.NoError(t, err)
	_ = resp2.Body.Close()

	require.GreaterOrEqual(t, len(*violations), 2, "harness failed to flag a bad body and an unknown path — it is not really validating")
}

func joinLines(ss []string) string {
	var b bytes.Buffer
	for _, s := range ss {
		b.WriteString("  - ")
		b.WriteString(s)
		b.WriteByte('\n')
	}
	return b.String()
}
