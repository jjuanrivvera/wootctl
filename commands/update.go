package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/wootctl/internal/update"
	"github.com/jjuanrivvera/wootctl/internal/version"
)

func init() {
	// The updater is self-contained; deps (client/config) are unused here.
	metaRegistrars = append(metaRegistrars, func(_ *deps) *cobra.Command { return newUpdateCmd() })
}

// newUpdater is the seam the command's own tests replace. The real one talks to GitHub, which
// a test must not do — and must not be allowed to replace the test binary either — so a test
// swaps in an updater pointed at an httptest server with an ExecutablePath it owns.
var newUpdater = func(currentVersion string) *update.Updater {
	return update.NewUpdater(currentVersion)
}

// newUpdateCmd builds `wootctl update` (self-update) + `update check`.
func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update wootctl to the latest release",
		Long: `Check GitHub for a newer release and, if one exists, download it, verify it
against the release checksums, and replace the running binary in place.`,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
			defer cancel()
			res := newUpdater(version.Get().Version).CheckAndUpdate(ctx)
			if res.Error != nil {
				return res.Error
			}
			out := cmd.OutOrStdout()
			if res.Updated {
				fmt.Fprintf(out, "Updated %s → %s. Restart to use the new version.\n", res.FromVersion, res.ToVersion)
			} else {
				fmt.Fprintln(out, "Already on the latest version.")
			}
			return nil
		},
	}
	cmd.AddCommand(&cobra.Command{
		Use:          "check",
		Short:        "Check for a newer release without installing it",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()
			cur := version.Get().Version
			rel, err := newUpdater(cur).GetLatestRelease(ctx)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Current: %s\nLatest:  %s\n", cur, rel.TagName)
			if update.IsNewer(rel.TagName, cur) {
				fmt.Fprintln(out, "An update is available. Run `wootctl update` to install it.")
			} else {
				fmt.Fprintln(out, "You are on the latest version.")
			}
			return nil
		},
	})
	return cmd
}
