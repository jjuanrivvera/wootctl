// Package config resolves cwctl configuration with a manual flag > env > file > default
// precedence (no Viper, per the cliwright house pattern). Profiles let one user drive
// several Chatwoot instances/accounts; the secret tokens never live here — only in the OS
// keyring (see internal/auth).
package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// DefaultProfile is used when no profile is selected anywhere.
const DefaultProfile = "default"

// Profile holds the non-secret settings for one Chatwoot instance + account. A profile
// bundles base_url + account_id + (keyring-held) tokens, so switching profile switches
// everything at once (DECISIONS.md #4).
type Profile struct {
	BaseURL   string  `yaml:"base_url,omitempty"`
	AccountID string  `yaml:"account_id,omitempty"`
	UserID    string  `yaml:"user_id,omitempty"` // from GET /api/v1/profile at auth login
	Email     string  `yaml:"email,omitempty"`   // identity display for auth status
	Rps       float64 `yaml:"rps,omitempty"`     // per-profile rate limit override
}

// Config is the on-disk configuration.
type Config struct {
	CurrentProfile string             `yaml:"current_profile,omitempty"`
	Profiles       map[string]Profile `yaml:"profiles,omitempty"`
	Aliases        map[string]string  `yaml:"aliases,omitempty"` // name -> expansion (e.g. "open" -> "conversations list --status open")

	path string `yaml:"-"`
}

// Dir returns the configuration directory: $XDG_CONFIG_HOME/cwctl if set, else ~/.cwctl-cli.
func Dir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "cwctl"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cwctl-cli"), nil
}

// Path returns the config file path.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Load reads the config file, returning an empty (but valid) Config if it does not exist.
func Load() (*Config, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	return LoadFrom(p)
}

// LoadFrom reads a config from an explicit path (used in tests).
func LoadFrom(p string) (*Config, error) {
	c := &Config{Profiles: map[string]Profile{}, path: p}
	data, err := os.ReadFile(p) //nolint:gosec // G304: p is the user's own config path
	if os.IsNotExist(err) {
		return c, nil
	}
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", p, err)
	}
	if c.Profiles == nil {
		c.Profiles = map[string]Profile{}
	}
	c.path = p
	return c, nil
}

// Save writes the config atomically: a temp file in the same directory then a rename, with
// dir 0700 and file 0600, so a token-adjacent config is never world-readable or torn.
func (c *Config) Save() error {
	if c.path == "" {
		p, err := Path()
		if err != nil {
			return err
		}
		c.path = p
	}
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".config-*.yaml.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op once renamed
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, c.path)
}

// FilePath returns where this config reads/writes.
func (c *Config) FilePath() string { return c.path }

// Profile returns the named profile (and whether it exists).
func (c *Config) Profile(name string) (Profile, bool) {
	p, ok := c.Profiles[name]
	return p, ok
}

// SetProfile creates or replaces a profile after validating its name.
func (c *Config) SetProfile(name string, p Profile) error {
	if err := ValidateProfileName(name); err != nil {
		return err
	}
	if p.BaseURL != "" {
		if err := ValidateBaseURL(p.BaseURL); err != nil {
			return err
		}
	}
	if c.Profiles == nil {
		c.Profiles = map[string]Profile{}
	}
	c.Profiles[name] = p
	return nil
}

// ProfileNames returns the profile names (unsorted; callers sort for display).
func (c *Config) ProfileNames() []string {
	names := make([]string, 0, len(c.Profiles))
	for n := range c.Profiles {
		names = append(names, n)
	}
	return names
}

// ResolveProfileName applies precedence: explicit flag > CWCTL_PROFILE > current_profile >
// default.
func (c *Config) ResolveProfileName(flag string) string {
	return FirstNonEmpty(flag, os.Getenv("CWCTL_PROFILE"), c.CurrentProfile, DefaultProfile)
}

// FirstNonEmpty returns the first non-empty string — the manual precedence helper used
// across cwctl instead of a config framework.
func FirstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// invalidProfileChars are path/traversal-dangerous characters disallowed in a profile name.
const invalidProfileChars = `/\:*?"<>|#`

// ValidateProfileName rejects empty names and traversal-dangerous characters, so a profile
// name can never be used to escape the config/keyring namespace.
func ValidateProfileName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	if name == "." || name == ".." {
		return fmt.Errorf("invalid profile name %q", name)
	}
	if strings.ContainsAny(name, invalidProfileChars) {
		return fmt.Errorf("profile name %q contains an invalid character (one of %s)", name, invalidProfileChars)
	}
	return nil
}

// ValidateBaseURL requires an http/https URL with a host and rejects cleartext http:// for a
// non-loopback host, so a token is never sent in the clear over the network.
func ValidateBaseURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid base URL %q: %w", raw, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("base URL must be http or https, got %q", u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("base URL %q has no host", raw)
	}
	if u.Scheme == "http" && !isLoopback(u.Hostname()) {
		return fmt.Errorf("refusing cleartext http:// for non-loopback host %q (the token would leak); use https", u.Hostname())
	}
	return nil
}

func isLoopback(host string) bool {
	switch host {
	case "localhost", "127.0.0.1", "::1":
		return true
	}
	return strings.HasPrefix(host, "127.")
}
