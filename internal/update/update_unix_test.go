//go:build !windows

package update

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A directory the process cannot write to (an install under /usr/local owned by root, the
// common case on the VPS) must fail at the staging step with a message that says so, not
// with a rename error further down.
func TestApplyUpdate_UnwritableDirectory(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root writes anywhere, so there is no failure to observe")
	}
	dir := t.TempDir()
	exe := filepath.Join(dir, "binary")
	if err := os.WriteFile(exe, []byte("OLD"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o500); err != nil { // r-x: readable, not writable
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

	u := &Updater{CurrentVersion: "0.5.0", ExecutablePath: exe}
	err := u.applyUpdate([]byte("NEW"))
	if err == nil {
		t.Fatal("expected an error when the install directory is not writable")
	}
	if !strings.Contains(err.Error(), "temp file") {
		t.Errorf("error = %v, want it to name the staging step", err)
	}
	if got, _ := os.ReadFile(exe); string(got) != "OLD" {
		t.Errorf("the installed binary was touched: %q", got)
	}
}
