package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// offerChatwootDropIn optionally installs a `chatwoot` symlink pointing at the wootctl binary,
// so the official Chatwoot CLI's commands (chatwoot agents, chatwoot contacts, …) run through
// wootctl — a drop-in replacement. wootctl's surface is a superset of the official CLI's, so
// this is safe. Interactive-only: on a pipe/CI it returns immediately so scripts never block.
func offerChatwootDropIn(cmd *cobra.Command) {
	f, ok := cmd.InOrStdin().(*os.File)
	if !ok || !term.IsTerminal(int(f.Fd())) {
		return
	}

	exe, err := os.Executable()
	if err != nil {
		return
	}
	link := filepath.Join(filepath.Dir(exe), "chatwoot")
	out := cmd.ErrOrStderr()

	// Already a wootctl-owned drop-in? Nothing to do.
	if target, err := os.Readlink(link); err == nil && target == exe {
		return
	}

	fmt.Fprintln(out, "\nwootctl is a drop-in superset of the official `chatwoot` CLI — the same commands,")
	fmt.Fprintln(out, "plus multi-profile, keyring fallback, and backup/sync.")
	if p, err := exec.LookPath("chatwoot"); err == nil {
		fmt.Fprintf(out, "(heads up: a `chatwoot` is already on your PATH at %s — probably the official CLI, which this would shadow.)\n", p)
	}
	ans, err := promptLine(cmd, "Install a `chatwoot` drop-in so `chatwoot …` runs wootctl? [y/N]: ")
	yes := strings.EqualFold(ans, "y") || strings.EqualFold(ans, "yes")
	if err != nil || !yes {
		return
	}

	manual := fmt.Sprintf("  ln -sf %q %q", exe, link)
	// Homebrew keeps binaries in a managed dir (Cellar/Caskroom); symlinking there is fragile
	// and off-PATH. Point brew users at a PATH dir instead of guessing.
	if strings.Contains(exe, "/Caskroom/") || strings.Contains(exe, "/Cellar/") {
		fmt.Fprintf(out, "wootctl is Homebrew-managed here. Create the drop-in on your PATH, e.g.:\n  ln -sf %q $(brew --prefix)/bin/chatwoot\n", exe)
		return
	}
	// Replace only our own/stale symlink; never clobber a real file (e.g. the official binary).
	if fi, err := os.Lstat(link); err == nil {
		if fi.Mode()&os.ModeSymlink == 0 {
			fmt.Fprintf(out, "a non-symlink `chatwoot` already exists at %s — leaving it. To force it:\n%s\n", link, manual)
			return
		}
		_ = os.Remove(link)
	}
	if err := os.Symlink(exe, link); err != nil {
		fmt.Fprintf(out, "couldn't create it (%v). Run this yourself:\n%s\n", err, manual)
		return
	}
	fmt.Fprintf(out, "done — `chatwoot` now runs wootctl (%s).\n", link)
}
