package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/wootctl/internal/version"
)

func init() {
	metaRegistrars = append(metaRegistrars, func(_ *deps) *cobra.Command {
		var jsonOut, check bool
		cmd := &cobra.Command{
			Use:   "version",
			Short: "Print version, commit, and build date",
			Long:  "Print build metadata. With --check, compare against the latest GitHub release.",
			Example: `  wootctl version
  wootctl version --json
  wootctl version --check`,
			Args: cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				info := version.Get()
				if jsonOut {
					b, _ := json.MarshalIndent(info, "", "  ")
					fmt.Fprintln(cmd.OutOrStdout(), string(b))
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), version.String())
				}
				if check {
					return reportLatest(cmd, info.Version)
				}
				return nil
			},
		}
		cmd.Flags().BoolVar(&jsonOut, "json", false, "output as JSON")
		cmd.Flags().BoolVar(&check, "check", false, "check for a newer release on GitHub")
		return cmd
	})
}

const latestReleaseURL = "https://api.github.com/repos/jjuanrivvera/wootctl/releases/latest"

// reportLatest fetches the newest published release tag and tells the user whether they are
// up to date. It is best-effort: a network failure is reported, not fatal.
func reportLatest(cmd *cobra.Command, current string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestReleaseURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "could not check for updates: %v\n", err)
		return nil
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(cmd.ErrOrStderr(), "could not check for updates: GitHub returned %d\n", resp.StatusCode)
		return nil
	}
	var rel struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil
	}
	latest := strings.TrimPrefix(rel.TagName, "v")
	switch {
	case latest == "":
		fmt.Fprintln(cmd.ErrOrStderr(), "no published release found")
	case strings.TrimPrefix(current, "v") == latest:
		fmt.Fprintln(cmd.OutOrStdout(), "up to date")
	default:
		fmt.Fprintf(cmd.OutOrStdout(), "a newer release is available: %s (you have %s)\n", latest, current)
	}
	return nil
}
