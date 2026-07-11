// Package auth stores bot tokens out of plaintext. The OS keyring (macOS Keychain, Linux
// Secret Service, Windows Credential Manager) is primary; an encrypted file is the fallback
// for headless hosts where no keyring is available (GOAL.md §1).
package auth

import (
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

// Service is the keyring service name under which tokens are stored, keyed by profile.
const Service = "wootctl"

// ErrNotFound is returned when no token is stored for a profile.
var ErrNotFound = errors.New("no token stored for this profile")

// Store persists and retrieves a per-profile secret token.
type Store interface {
	Set(profile, token string) error
	Get(profile string) (string, error)
	Delete(profile string) error
	// Backend names where the token currently lives ("keyring" or "file"), for doctor output.
	Backend() string
}

// store tries the keyring first and transparently falls back to an encrypted file when the
// keyring is unavailable (no Secret Service on a headless Linux box, for example).
type store struct {
	service string
	fb      *fileStore
	backend string
}

// New returns the default token store. dir is where the encrypted fallback file lives
// (the config dir); it is only touched if the keyring is unreachable.
func New(dir string) Store {
	return &store{service: Service, fb: newFileStore(dir), backend: "keyring"}
}

func (s *store) Backend() string { return s.backend }

func (s *store) Set(profile, token string) error {
	if err := keyring.Set(s.service, profile, token); err != nil {
		s.backend = "file"
		return s.fb.Set(profile, token)
	}
	return nil
}

func (s *store) Get(profile string) (string, error) {
	tok, err := keyring.Get(s.service, profile)
	if err == nil {
		return tok, nil
	}
	if errors.Is(err, keyring.ErrNotFound) {
		// Keyring works but has nothing — check the fallback file before giving up, in case
		// the token was written on a host without a keyring.
		if tok, ferr := s.fb.Get(profile); ferr == nil {
			s.backend = "file"
			return tok, nil
		}
		return "", ErrNotFound
	}
	// Keyring is unavailable entirely → use the fallback file.
	s.backend = "file"
	tok, ferr := s.fb.Get(profile)
	if ferr != nil {
		return "", ErrNotFound
	}
	return tok, nil
}

func (s *store) Delete(profile string) error {
	// Remove from both backends. The keyring delete is best-effort: it may be entirely
	// unavailable on this host, in which case the encrypted file is the real store. We only
	// surface an error if the file backend itself fails for a reason other than "not found".
	_ = keyring.Delete(s.service, profile)
	ferr := s.fb.Delete(profile)
	if ferr == nil || errors.Is(ferr, ErrNotFound) {
		return nil
	}
	return fmt.Errorf("delete token: %w", ferr)
}

// PlatformKey namespaces the platform app token keyring entry for a profile. The "/" is
// safe: profile names reject it (config.ValidateProfileName), so a user profile can never
// collide with a platform entry.
func PlatformKey(profile string) string { return profile + "/platform" }
