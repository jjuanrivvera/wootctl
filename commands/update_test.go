package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jjuanrivvera/wootctl/internal/update"
	"github.com/jjuanrivvera/wootctl/internal/version"
)

// releaseServer answers the one endpoint the updater asks for, with the tag it is given.
// An empty tag makes it fail, for the error paths.
func releaseServer(t *testing.T, tag string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tag == "" {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		if !strings.HasSuffix(r.URL.Path, "/releases/latest") {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"tag_name": tag, "assets": []any{}})
	}))
	t.Cleanup(srv.Close)
	return srv
}

// withUpdater points the command's updater at a test server for the duration of one test,
// and gives the build a real version number: a "dev" build deliberately never updates
// itself, so with the default the interesting branches are unreachable.
func withUpdater(t *testing.T, srv *httptest.Server) {
	t.Helper()
	originalVersion := version.Version
	version.Version = "0.5.0"
	t.Cleanup(func() { version.Version = originalVersion })

	original := newUpdater
	newUpdater = func(currentVersion string) *update.Updater {
		u := update.NewUpdaterWithBaseURL(currentVersion, srv.URL)
		u.HTTPClient = srv.Client()
		// Never let a test replace the binary that is running it.
		u.ExecutablePath = t.TempDir() + "/wootctl"
		return u
	}
	t.Cleanup(func() { newUpdater = original })
}

func runUpdateCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newUpdateCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(t.Context())
	return out.String(), err
}

func TestUpdateCheck_ReportsAnAvailableUpdate(t *testing.T) {
	withUpdater(t, releaseServer(t, "v99.0.0"))

	out, err := runUpdateCmd(t, "check")
	require.NoError(t, err)
	assert.Contains(t, out, "Latest:  v99.0.0")
	assert.Contains(t, out, "An update is available")
}

func TestUpdateCheck_ReportsUpToDate(t *testing.T) {
	withUpdater(t, releaseServer(t, "v0.0.1"))

	out, err := runUpdateCmd(t, "check")
	require.NoError(t, err)
	assert.Contains(t, out, "You are on the latest version")
	assert.NotContains(t, out, "An update is available")
}

func TestUpdateCheck_SurfacesTheAPIError(t *testing.T) {
	withUpdater(t, releaseServer(t, ""))

	_, err := runUpdateCmd(t, "check")
	require.Error(t, err)
}

// `update` with nothing newer must say so and leave the binary alone.
func TestUpdate_NothingNewer(t *testing.T) {
	withUpdater(t, releaseServer(t, "v0.0.1"))

	out, err := runUpdateCmd(t)
	require.NoError(t, err)
	assert.Contains(t, out, "Already on the latest version")
}

func TestUpdate_SurfacesTheAPIError(t *testing.T) {
	withUpdater(t, releaseServer(t, ""))

	_, err := runUpdateCmd(t)
	require.Error(t, err)
}

// A dev build must never update itself, however new the release on the other end is.
func TestUpdate_DevBuildStaysPut(t *testing.T) {
	withUpdater(t, releaseServer(t, "v99.0.0"))
	version.Version = "dev"

	out, err := runUpdateCmd(t)
	require.NoError(t, err)
	assert.Contains(t, out, "Already on the latest version")
}
