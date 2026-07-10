package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/cwctl/internal/auth"
	"github.com/jjuanrivvera/cwctl/internal/config"
	"github.com/jjuanrivvera/cwctl/internal/version"
)

type doctorCheck struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
}

func init() {
	metaRegistrars = append(metaRegistrars, func(d *deps) *cobra.Command {
		var jsonOut bool
		cmd := &cobra.Command{
			Use:   "doctor",
			Short: "Diagnose configuration, credentials, and connectivity",
			Long:  "Run end-to-end checks: config file, active profile, base URL, token presence and validity, account scope, platform token. Exits non-zero when anything fails.",
			Example: `  cwctl doctor
  cwctl doctor --json`,
			Args: cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				var checks []doctorCheck
				add := func(name string, ok bool, detail string) {
					checks = append(checks, doctorCheck{Name: name, OK: ok, Detail: detail})
				}

				cfgPath, _ := config.Path()
				cfg, err := d.loadConfig()
				add("config readable", err == nil, cfgPath)
				var profileName string
				if err == nil {
					profileName = cfg.ResolveProfileName(d.gf.profile)
					_, exists := cfg.Profile(profileName)
					add("profile resolved", true, profileName)
					if !exists && profileName != config.DefaultProfile {
						add("profile exists", false, fmt.Sprintf("profile %q not in config — run `cwctl --profile %s auth login`", profileName, profileName))
					}
				}

				c, _, cErr := d.getAPIClient(false)
				if cErr != nil {
					add("base URL", false, cErr.Error())
				} else {
					add("base URL", true, c.BaseURL)
					add("account id", c.AccountID != "", orMsg(c.AccountID, "unset — run `cwctl auth login` or `cwctl config set account_id <id>`"))

					token, tErr := d.store().Get(profileName)
					hasToken := tErr == nil && token != ""
					detail := "backend: " + d.store().Backend()
					if !hasToken && os.Getenv("CWCTL_API_KEY") != "" {
						hasToken, detail = true, "from CWCTL_API_KEY env"
					}
					add("user token stored", hasToken, detail)

					if hasToken {
						var me chatwootProfile
						if err := c.GetJSON(cmd.Context(), "api/v1/profile", nil, &me); err != nil {
							add("API reachable + token valid", false, err.Error())
						} else {
							add("API reachable + token valid", true, fmt.Sprintf("%s <%s>", me.Name, me.Email))
						}
					}

					if _, ok := firstStoreHit(d, auth.PlatformKey(profileName)); ok {
						add("platform token stored", true, "cwctl platform … available")
					} else {
						add("platform token stored", true, "absent (only needed for `cwctl platform …`)")
					}
				}
				add("version", true, version.String())

				failed := 0
				for _, ch := range checks {
					if !ch.OK {
						failed++
					}
				}
				if jsonOut {
					b, _ := json.MarshalIndent(map[string]any{"checks": checks, "ok": failed == 0}, "", "  ")
					fmt.Fprintln(cmd.OutOrStdout(), string(b))
				} else {
					for _, ch := range checks {
						mark := "✓"
						if !ch.OK {
							mark = "✗"
						}
						fmt.Fprintf(cmd.OutOrStdout(), "%s %-30s %s\n", mark, ch.Name, ch.Detail)
					}
				}
				if failed > 0 {
					return fmt.Errorf("%d check(s) failed", failed)
				}
				return nil
			},
		}
		cmd.Flags().BoolVar(&jsonOut, "json", false, "output as JSON")
		return cmd
	})
}

func orMsg(v, fallback string) string {
	if v != "" {
		return v
	}
	return fallback
}
