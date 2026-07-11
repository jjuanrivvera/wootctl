package commands

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jjuanrivvera/wootctl/internal/auth"
	"github.com/jjuanrivvera/wootctl/internal/config"
)

const profileBody = `{"id":7,"name":"Juan","email":"juan@example.com","accounts":[{"id":1,"name":"Soporte","role":"administrator"}]}`

const twoAccountsBody = `{"id":7,"name":"Juan","email":"juan@example.com","accounts":[{"id":1,"name":"A","role":"administrator"},{"id":2,"name":"B","role":"agent"}]}`

func TestAuthLogin_StoresTokenAndProfile(t *testing.T) {
	e := newEnv(t, jsonHandler(profileBody))
	t.Setenv("WOOTCTL_API_KEY", "") // login must not shortcut through env
	t.Setenv("WOOTCTL_ACCOUNT_ID", "")
	t.Setenv("WOOTCTL_BASE_URL", "")

	out, errOut, err := e.run("auth", "login", "--url", e.srv.URL, "--api-key", "secret-token")
	require.NoError(t, err)
	assert.Contains(t, errOut, "verified as Juan")
	assert.Contains(t, out, `profile "default" ready`)
	assert.Contains(t, out, "account 1")

	tok, err := e.store.Get("default")
	require.NoError(t, err)
	assert.Equal(t, "secret-token", tok)

	cfg, err := config.Load()
	require.NoError(t, err)
	prof, ok := cfg.Profile("default")
	require.True(t, ok)
	assert.Equal(t, e.srv.URL, prof.BaseURL)
	assert.Equal(t, "1", prof.AccountID)
	assert.Equal(t, "juan@example.com", prof.Email)
}

func TestAuthLogin_MultipleAccountsNeedsChoice(t *testing.T) {
	e := newEnv(t, jsonHandler(twoAccountsBody))
	t.Setenv("WOOTCTL_ACCOUNT_ID", "")

	// Non-interactive with several accounts and no --account: fails with the list.
	_, _, err := e.run("auth", "login", "--url", e.srv.URL, "--api-key", "tok")
	require.Error(t, err)

	// Explicit --account selects without prompting.
	out, _, err := e.run("auth", "login", "--url", e.srv.URL, "--api-key", "tok", "--account", "2")
	require.NoError(t, err)
	assert.Contains(t, out, "account 2")
}

func TestAuthLogin_RejectsBadToken(t *testing.T) {
	e := newEnv(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errors":["Invalid token"]}`))
	})
	_, _, err := e.run("auth", "login", "--url", e.srv.URL, "--api-key", "bad")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "verification failed")
}

func TestAuthLogin_PlatformTokenOnly(t *testing.T) {
	e := newEnv(t, jsonHandler(profileBody))
	_, errOut, err := e.run("auth", "login", "--url", e.srv.URL, "--platform-token", "plat-tok")
	require.NoError(t, err)
	assert.Contains(t, errOut, "stored platform app token")
	tok, err := e.store.Get(auth.PlatformKey("default"))
	require.NoError(t, err)
	assert.Equal(t, "plat-tok", tok)
	_, err = e.store.Get("default")
	assert.Error(t, err, "user token untouched in platform-only mode")
}

func TestAuthStatus_ReportsIdentity(t *testing.T) {
	e := newEnv(t, jsonHandler(profileBody))
	out, _, err := e.run("auth", "status", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, `"valid": true`)
	assert.Contains(t, out, "juan@example.com")

	// whoami alias reaches the same command.
	out, _, err = e.run("auth", "whoami", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, `"valid": true`)
}

func TestAuthStatus_FailsNonZeroOnBadToken(t *testing.T) {
	e := newEnv(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"Unauthorized"}`))
	})
	_, _, err := e.run("auth", "status")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token invalid")
}

func TestAuthLogout_RemovesTokens(t *testing.T) {
	e := newEnv(t, jsonHandler(profileBody))
	require.NoError(t, e.store.Set("default", "u"))
	require.NoError(t, e.store.Set(auth.PlatformKey("default"), "p"))

	out, _, err := e.run("auth", "logout")
	require.NoError(t, err)
	assert.Contains(t, out, `logged out`)
	_, err = e.store.Get("default")
	assert.Error(t, err)
	_, err = e.store.Get(auth.PlatformKey("default"))
	assert.Error(t, err)
}

func TestAuthLogout_PlatformOnly(t *testing.T) {
	e := newEnv(t, jsonHandler(profileBody))
	require.NoError(t, e.store.Set("default", "u"))
	require.NoError(t, e.store.Set(auth.PlatformKey("default"), "p"))
	_, _, err := e.run("auth", "logout", "--platform-only")
	require.NoError(t, err)
	_, err = e.store.Get("default")
	assert.NoError(t, err, "user token kept")
	_, err = e.store.Get(auth.PlatformKey("default"))
	assert.Error(t, err)
}

func TestConfig_SetViewUseListProfiles(t *testing.T) {
	e := newEnv(t, jsonHandler(profileBody))

	_, _, err := e.run("config", "set", "base_url", e.srv.URL)
	require.NoError(t, err)
	_, _, err = e.run("config", "set", "account_id", "3")
	require.NoError(t, err)
	_, _, err = e.run("config", "set", "rps", "2.5")
	require.NoError(t, err)
	_, _, err = e.run("config", "set", "rps", "-1")
	require.Error(t, err)
	_, _, err = e.run("config", "set", "bogus", "x")
	require.Error(t, err)

	out, _, err := e.run("config", "view", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, `"account_id": "3"`)
	assert.NotContains(t, out, "secret", "no token material in config view")

	out, _, err = e.run("config", "list-profiles", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, `"profile": "default"`)

	_, _, err = e.run("config", "use", "missing")
	require.Error(t, err)
	_, _, err = e.run("config", "use", "default")
	require.NoError(t, err)

	out, _, err = e.run("config", "path")
	require.NoError(t, err)
	assert.Contains(t, out, "config.yaml")
}

func TestAlias_SetListRemoveAndExpansion(t *testing.T) {
	e := newEnv(t, jsonHandler(`[]`))

	out, _, err := e.run("alias", "set", "open", "conversations list --status open")
	require.NoError(t, err)
	assert.Contains(t, out, `alias "open"`)

	// Built-ins can never be shadowed.
	_, _, err = e.run("alias", "set", "auth", "labels list")
	require.Error(t, err)

	out, _, err = e.run("alias", "list")
	require.NoError(t, err)
	assert.Contains(t, out, "open = conversations list --status open")

	got := ExpandAliases([]string{"open", "--all"})
	assert.Equal(t, []string{"conversations", "list", "--status", "open", "--all"}, got)

	assert.Equal(t, []string{"labels", "list"}, ExpandAliases([]string{"labels", "list"}), "built-ins pass through")
	assert.Equal(t, []string{"unknown"}, ExpandAliases([]string{"unknown"}), "unknown names pass through")

	_, _, err = e.run("alias", "remove", "open")
	require.NoError(t, err)
	_, _, err = e.run("alias", "remove", "open")
	require.Error(t, err)
}

func TestDoctor_HappyAndJSON(t *testing.T) {
	e := newEnv(t, jsonHandler(profileBody))
	out, _, err := e.run("doctor")
	require.NoError(t, err)
	assert.Contains(t, out, "API reachable + token valid")
	assert.Contains(t, out, "✓")

	out, _, err = e.run("doctor", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, `"ok": true`)
}

func TestDoctor_FailsWithoutBaseURL(t *testing.T) {
	e := newEnv(t, jsonHandler(profileBody))
	t.Setenv("WOOTCTL_BASE_URL", "")
	_, _, err := e.run("doctor")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "check(s) failed")
}

func TestVersionCmd_JSON(t *testing.T) {
	e := newEnv(t, jsonHandler(`{}`))
	out, _, err := e.run("version", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, `"version"`)
}

func TestCompletion_GeneratesScripts(t *testing.T) {
	e := newEnv(t, jsonHandler(`{}`))
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		out, _, err := e.run("completion", shell)
		require.NoError(t, err, shell)
		assert.NotEmpty(t, out)
	}
	_, _, err := e.run("completion", "tcsh")
	require.Error(t, err)
}

func TestHelp_ShowsGroups(t *testing.T) {
	e := newEnv(t, jsonHandler(`{}`))
	out, _, err := e.run("--help")
	require.NoError(t, err)
	for _, want := range []string{"conversations", "platform", "client", "reports", "auth"} {
		assert.Contains(t, out, want)
	}
}

func TestGuardWriteFiles_CreatesUnderDotClaude(t *testing.T) {
	e := newEnv(t, jsonHandler(`{}`))
	dir := t.TempDir()
	t.Chdir(dir)

	_, errOut, err := e.run("agent", "guard", "--host", "claude-code", "--write")
	require.NoError(t, err)
	assert.Contains(t, errOut, "wrote .claude/hooks/wootctl-guard.sh")

	// Never overwrites.
	_, _, err = e.run("agent", "guard", "--host", "claude-code", "--write")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestGuard_UnknownHost(t *testing.T) {
	e := newEnv(t, jsonHandler(`{}`))
	_, _, err := e.run("agent", "guard", "--host", "vim")
	require.Error(t, err)
}

func TestGuard_OutFlagWritesFile(t *testing.T) {
	e := newEnv(t, jsonHandler(`{}`))
	dest := t.TempDir() + "/codex.toml"
	_, errOut, err := e.run("agent", "guard", "--host", "codex", "--out", dest)
	require.NoError(t, err)
	assert.Contains(t, errOut, "wrote codex safety config")
	b, err := os.ReadFile(dest) // #nosec G304 -- test-owned temp path
	require.NoError(t, err)
	assert.Contains(t, string(b), "sandbox_mode")
}
