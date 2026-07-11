package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_SaveLoadRoundTrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config.yaml")
	c, err := LoadFrom(p)
	require.NoError(t, err)
	require.NoError(t, c.SetProfile("prod", Profile{BaseURL: "https://soporte.example.com", AccountID: "1", Email: "me@example.com", Rps: 2.5}))
	c.CurrentProfile = "prod"
	require.NoError(t, c.Save())

	// File must be 0600 (Unix only — Windows has no POSIX mode bits).
	info, err := os.Stat(p)
	require.NoError(t, err)
	if runtime.GOOS != "windows" {
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	}

	got, err := LoadFrom(p)
	require.NoError(t, err)
	assert.Equal(t, "prod", got.CurrentProfile)
	pr, ok := got.Profile("prod")
	require.True(t, ok)
	assert.Equal(t, "1", pr.AccountID)
	assert.Equal(t, 2.5, pr.Rps)
	assert.Equal(t, []string{"prod"}, got.ProfileNames())
}

func TestConfig_LoadMissingIsEmpty(t *testing.T) {
	c, err := LoadFrom(filepath.Join(t.TempDir(), "nope.yaml"))
	require.NoError(t, err)
	assert.Empty(t, c.Profiles)
}

func TestResolveProfileName_Precedence(t *testing.T) {
	c := &Config{CurrentProfile: "fromfile"}
	t.Setenv("WOOTCTL_PROFILE", "")
	assert.Equal(t, "flag", c.ResolveProfileName("flag"))
	assert.Equal(t, "fromfile", c.ResolveProfileName(""))

	t.Setenv("WOOTCTL_PROFILE", "fromenv")
	assert.Equal(t, "fromenv", c.ResolveProfileName(""))
	assert.Equal(t, "flag", c.ResolveProfileName("flag"), "flag still wins over env")

	empty := &Config{}
	t.Setenv("WOOTCTL_PROFILE", "")
	assert.Equal(t, DefaultProfile, empty.ResolveProfileName(""))
}

func TestValidateProfileName(t *testing.T) {
	for _, ok := range []string{"default", "prod", "my-bot_2"} {
		assert.NoError(t, ValidateProfileName(ok), ok)
	}
	for _, bad := range []string{"", "  ", ".", "..", "a/b", `a\b`, "a:b", "a*b", "a?b"} {
		assert.Error(t, ValidateProfileName(bad), bad)
	}
}

func TestValidateBaseURL(t *testing.T) {
	assert.NoError(t, ValidateBaseURL("https://app.chatwoot.com"))
	assert.NoError(t, ValidateBaseURL("http://localhost:8081"))
	assert.NoError(t, ValidateBaseURL("http://127.0.0.1:8081"))
	assert.Error(t, ValidateBaseURL("ftp://example.com"))
	assert.Error(t, ValidateBaseURL("https://"))
	assert.Error(t, ValidateBaseURL("http://example.com"), "cleartext http to a public host must be rejected")
}

func TestFirstNonEmpty(t *testing.T) {
	assert.Equal(t, "a", FirstNonEmpty("", "", "a", "b"))
	assert.Equal(t, "", FirstNonEmpty("", ""))
}

func TestSetProfile_RejectsBadName(t *testing.T) {
	c := &Config{}
	assert.Error(t, c.SetProfile("a/b", Profile{}))
	assert.Error(t, c.SetProfile("ok", Profile{BaseURL: "http://example.com"}))
}

func TestDirAndPath_XDG(t *testing.T) {
	base := t.TempDir() // platform-appropriate absolute path
	t.Setenv("XDG_CONFIG_HOME", base)
	dir, err := Dir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(base, "wootctl"), dir)
	p, err := Path()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(base, "wootctl", "config.yaml"), p)
}

func TestDir_HomeFallback(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	dir, err := Dir()
	require.NoError(t, err)
	assert.Contains(t, dir, ".wootctl-cli")
}

func TestLoad_UsesXDG(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	c, err := Load()
	require.NoError(t, err)
	assert.NotEmpty(t, c.FilePath())
	// Save then reload through the public Load path.
	require.NoError(t, c.SetProfile("p", Profile{AccountID: "9"}))
	require.NoError(t, c.Save())
	again, err := Load()
	require.NoError(t, err)
	_, ok := again.Profile("p")
	assert.True(t, ok)
}
