package commands

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneric_ListRendersTable(t *testing.T) {
	e := newEnv(t, jsonHandler(`{"payload":[{"id":1,"title":"vip","color":"#00f","show_on_sidebar":true}]}`))
	out, _, err := e.run("labels", "list")
	require.NoError(t, err)
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "TITLE")
	assert.Contains(t, out, "vip")
}

func TestGeneric_ListJSONKeepsUnknownFields(t *testing.T) {
	// A field wootctl knows nothing about must survive to -o json (DECISIONS #19).
	e := newEnv(t, jsonHandler(`{"payload":[{"id":1,"title":"vip","brand_new_field":"kept"}]}`))
	out, _, err := e.run("labels", "list", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "brand_new_field")
	assert.Contains(t, out, "kept")
}

func TestGeneric_ListServerFilterAndPage(t *testing.T) {
	h, log := recordingHandler(t, func(r *http.Request) (int, string) { return 200, `[]` })
	e := newEnv(t, h)
	_, _, err := e.run("conversations", "list", "--status", "open", "--assignee-type", "me", "--page", "3")
	require.NoError(t, err)
	require.Len(t, *log, 1)
	q := (*log)[0].Query
	assert.Contains(t, q, "status=open")
	assert.Contains(t, q, "assignee_type=me")
	assert.Contains(t, q, "page=3")
	assert.Equal(t, "/api/v1/accounts/1/conversations", (*log)[0].Path)
}

func TestGeneric_ListClientSideFilterAndLimit(t *testing.T) {
	e := newEnv(t, jsonHandler(`[{"id":1,"name":"a","role":"agent"},{"id":2,"name":"b","role":"administrator"},{"id":3,"name":"c","role":"agent"}]`))
	out, _, err := e.run("agents", "list", "--filter", "role=agent", "-o", "id")
	require.NoError(t, err)
	assert.Equal(t, "1\n3\n", out)

	out, _, err = e.run("agents", "list", "--filter", "role=agent", "--limit", "1", "-o", "id")
	require.NoError(t, err)
	assert.Equal(t, "1\n", out)

	_, _, err = e.run("agents", "list", "--filter", "not-a-pair")
	require.Error(t, err)
}

func TestGeneric_ListAllWalksPages(t *testing.T) {
	h, log := recordingHandler(t, func(r *http.Request) (int, string) {
		switch r.URL.Query().Get("page") {
		case "1":
			return 200, `{"payload":[{"id":1},{"id":2}]}`
		case "2":
			return 200, `{"payload":[{"id":3}]}`
		default:
			return 200, `{"payload":[]}`
		}
	})
	e := newEnv(t, h)
	out, _, err := e.run("labels", "list", "--all", "-o", "id")
	require.NoError(t, err)
	assert.Equal(t, "1\n2\n3\n", out)
	assert.GreaterOrEqual(t, len(*log), 3)
}

func TestGeneric_GetUnwrapsEnvelope(t *testing.T) {
	e := newEnv(t, jsonHandler(`{"payload":{"id":7,"title":"vip"}}`))
	out, _, err := e.run("labels", "get", "7", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, `"id": 7`)
}

func TestGeneric_CreateMergesDataAndFlags(t *testing.T) {
	h, log := recordingHandler(t, func(r *http.Request) (int, string) { return 200, `{"id":9}` })
	e := newEnv(t, h)
	_, _, err := e.run("labels", "create", "-d", `{"description":"from data","title":"overridden"}`, "--title", "vip")
	require.NoError(t, err)
	require.Len(t, *log, 1)
	body := (*log)[0].Body
	assert.Contains(t, body, `"title":"vip"`, "flags override --data keys")
	assert.Contains(t, body, `"description":"from data"`)
	assert.Equal(t, http.MethodPost, (*log)[0].Method)
}

func TestGeneric_CreateRequiredFieldEnforced(t *testing.T) {
	e := newEnv(t, jsonHandler(`{}`))
	_, _, err := e.run("labels", "create", "--color", "#00f")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--title")

	// The same key supplied via --data satisfies the requirement.
	_, _, err = e.run("labels", "create", "-d", `{"title":"via-data"}`)
	require.NoError(t, err)
}

func TestGeneric_CreateDataFromFileAndStdin(t *testing.T) {
	h, log := recordingHandler(t, func(r *http.Request) (int, string) { return 200, `{}` })
	e := newEnv(t, h)

	payload := filepath.Join(t.TempDir(), "payload.json")
	require.NoError(t, os.WriteFile(payload, []byte(`{"title":"from-file"}`), 0o600))
	_, _, err := e.run("labels", "create", "-d", "@"+payload)
	require.NoError(t, err)
	assert.Contains(t, (*log)[0].Body, "from-file")

	// Invalid JSON fails with a pointed error.
	_, _, err = e.run("labels", "create", "-d", `not-json`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JSON")
}

func TestGeneric_UpdateUsesConfiguredMethod(t *testing.T) {
	h, log := recordingHandler(t, func(r *http.Request) (int, string) { return 200, `{}` })
	e := newEnv(t, h)
	_, _, err := e.run("labels", "update", "5", "--title", "renamed")
	require.NoError(t, err)
	assert.Equal(t, http.MethodPatch, (*log)[0].Method)

	_, _, err = e.run("contacts", "update", "5", "--name", "Ana")
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, (*log)[1].Method, "contacts update is PUT per the spec")
}

func TestGeneric_DeleteConfirmsOnStdout(t *testing.T) {
	h, log := recordingHandler(t, func(r *http.Request) (int, string) { return 200, `` })
	e := newEnv(t, h)
	out, _, err := e.run("labels", "delete", "5")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, (*log)[0].Method)
	assert.Equal(t, "/api/v1/accounts/1/labels/5", (*log)[0].Path)
	assert.Contains(t, out, "deleted label 5")

	out, _, err = e.run("labels", "delete", "6", "--quiet")
	require.NoError(t, err)
	assert.NotContains(t, out, "deleted")
}

func TestGeneric_DryRunSkipsServer(t *testing.T) {
	hits := 0
	e := newEnv(t, func(w http.ResponseWriter, r *http.Request) { hits++ })
	_, _, err := e.run("labels", "delete", "5", "--dry-run")
	require.NoError(t, err)
	assert.Zero(t, hits, "dry-run must not touch the server")
}

func TestGeneric_ColumnsFlag(t *testing.T) {
	e := newEnv(t, jsonHandler(`[{"id":1,"title":"vip","color":"#00f"}]`))
	out, _, err := e.run("labels", "list", "--columns", "title")
	require.NoError(t, err)
	assert.Contains(t, out, "TITLE")
	assert.NotContains(t, out, "COLOR")
}

func TestGeneric_JQFlag(t *testing.T) {
	e := newEnv(t, jsonHandler(`{"payload":[{"id":1,"title":"vip"}]}`))
	out, _, err := e.run("labels", "list", "-o", "json", "--jq", ".[0].title")
	require.NoError(t, err)
	assert.Contains(t, out, `"vip"`)
}

func TestRoot_RejectsUnknownOutputFormat(t *testing.T) {
	e := newEnv(t, jsonHandler(`[]`))
	_, _, err := e.run("labels", "list", "-o", "xml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown output format")
}

func TestRoot_APIErrorSurfacesHint(t *testing.T) {
	e := newEnv(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"Invalid token"}`))
	})
	_, _, err := e.run("labels", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auth login")
}

func TestGeneric_ManyResourcesListSmoke(t *testing.T) {
	// One httptest sweep across every generic list to catch path typos.
	var paths []string
	e := newEnv(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		_, _ = fmt.Fprint(w, `[]`)
	})
	for i, res := range []string{"agent-bots", "agents", "audit-logs", "automation-rules", "canned-responses", "contacts", "conversations", "custom-attributes", "custom-filters", "inboxes", "labels", "portals", "teams", "webhooks"} {
		_, _, err := e.run(res, "list")
		require.NoError(t, err, res)
		require.Len(t, paths, i+1)
		assert.True(t, strings.HasPrefix(paths[i], "/api/v1/accounts/1/"), "%s hit %s", res, paths[i])
	}
}

func TestGeneric_PlatformCRUDSmoke(t *testing.T) {
	h, log := recordingHandler(t, func(r *http.Request) (int, string) {
		if r.Method == http.MethodGet && r.URL.Path == "/platform/api/v1/agent_bots" {
			return 200, `[]`
		}
		return 200, `{"id":2}`
	})
	e := newEnv(t, h)

	_, _, err := e.run("platform", "accounts", "create", "--name", "acme")
	require.NoError(t, err)
	assert.Equal(t, "/platform/api/v1/accounts", (*log)[0].Path)

	_, _, err = e.run("platform", "agent-bots", "list")
	require.NoError(t, err)
	assert.Equal(t, "/platform/api/v1/agent_bots", (*log)[1].Path)

	_, _, err = e.run("platform", "users", "sso-link", "7")
	require.NoError(t, err)
	assert.Equal(t, "/platform/api/v1/users/7/login", (*log)[2].Path)

	_, _, err = e.run("platform", "account-users", "create", "2", "--user-id", "7", "--role", "agent")
	require.NoError(t, err)
	assert.Equal(t, "/platform/api/v1/accounts/2/account_users", (*log)[3].Path)
	assert.Contains(t, (*log)[3].Body, `"role":"agent"`)

	_, _, err = e.run("platform", "account-users", "delete", "2", "--user-id", "7")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, (*log)[4].Method)
	assert.Contains(t, (*log)[4].Body, `"user_id":7`, "user id rides the DELETE body")
}

func TestGeneric_AllPagesRespectsCount(t *testing.T) {
	calls := 0
	e := newEnv(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = fmt.Fprint(w, `{"meta":{"count":2,"current_page":"`+strconv.Itoa(calls)+`"},"payload":[{"id":`+strconv.Itoa(calls)+`}]}`)
	})
	out, _, err := e.run("contacts", "list", "--all", "-o", "id")
	require.NoError(t, err)
	assert.Equal(t, "1\n2\n", out)
	assert.Equal(t, 2, calls)
}

func TestGeneric_UpdateInheritedFieldsDropRequired(t *testing.T) {
	h, log := recordingHandler(t, func(r *http.Request) (int, string) { return 200, `{}` })
	e := newEnv(t, h)
	// labels create requires --title; update inherits the fields but must NOT require it.
	_, _, err := e.run("labels", "update", "1", "--description", "partial patch")
	require.NoError(t, err)
	assert.Contains(t, (*log)[0].Body, `"description":"partial patch"`)

	// Explicit UpdateFields keep their spec-mandated required flags (agents update needs --role).
	_, _, err = e.run("agents", "update", "1", "--availability", "busy")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--role")
}

func TestGeneric_CreateUnwrapsDoubleNestedEnvelope(t *testing.T) {
	// Chatwoot wraps webhook create as {"payload":{"webhook":{…}}} — the rendered record
	// must be the webhook itself, not the wrapper.
	e := newEnv(t, jsonHandler(`{"payload":{"webhook":{"id":3,"url":"https://x"}}}`))
	out, _, err := e.run("webhooks", "create", "--url", "https://x", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, `"id": 3`)
	assert.NotContains(t, out, `"payload"`)
}
