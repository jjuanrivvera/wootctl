package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/wootctl/internal/api"
	"github.com/jjuanrivvera/wootctl/internal/auth"
	"github.com/jjuanrivvera/wootctl/internal/config"
)

// chatwootProfile is the GET /api/v1/profile response — the whoami endpoint every login
// verifies against. Accounts carries every account the token can access.
type chatwootProfile struct {
	ID      api.ID `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	ATCount int    `json:"-"`
	Acc     []struct {
		ID   api.ID `json:"id"`
		Name string `json:"name"`
		Role string `json:"role"`
	} `json:"accounts"`
}

func init() {
	metaRegistrars = append(metaRegistrars, func(d *deps) *cobra.Command {
		authCmd := &cobra.Command{
			Use:   "auth",
			Short: "Manage Chatwoot tokens and verify authentication",
			Long: `Capture, verify, and remove the tokens for a profile. Tokens are stored in your OS
keyring (with an encrypted-file fallback on headless hosts), never in the config file.

A profile can hold two token kinds:
  user      the api_access_token from your Chatwoot profile page — drives the
            application API (conversations, contacts, reports, …)
  platform  a platform app token (self-hosted; from a super-admin platform app) —
            drives the 'wootctl platform …' commands`,
		}
		authCmd.AddCommand(authLoginCmd(d), authLogoutCmd(d), authStatusCmd(d))
		return authCmd
	})
}

func authLoginCmd(d *deps) *cobra.Command {
	var apiKey, platformToken, baseURL, accountID string
	var noVerify bool
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Store a Chatwoot token and verify it",
		Long: `Capture your instance URL and api_access_token, verify them against GET /api/v1/profile,
pick the account to scope commands to, and save everything under the active profile.

Find the token in Chatwoot: Profile Settings → Access Token.`,
		Example: `  wootctl auth login                                  # interactive (hidden token input)
  wootctl auth login --base-url https://app.chatwoot.com --api-key <token>
  wootctl --profile staging auth login                # save a second instance
  wootctl auth login --platform-token <token>         # also store a platform app token`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := d.loadConfig()
			if err != nil {
				return err
			}
			profileName := cfg.ResolveProfileName(d.gf.profile)
			prof, _ := cfg.Profile(profileName)

			base := config.FirstNonEmpty(baseURL, d.gf.baseURL, prof.BaseURL)
			if base == "" {
				base, err = promptLine(cmd, "Chatwoot base URL (e.g. https://app.chatwoot.com): ")
				if err != nil {
					return err
				}
			}
			base = strings.TrimRight(base, "/")
			if err := config.ValidateBaseURL(base); err != nil {
				return err
			}
			prof.BaseURL = base

			// Prompt for the user token unless it was given — or the invocation is purely
			// "store a platform token" (--platform-token with no --api-key).
			if apiKey == "" && platformToken == "" {
				apiKey, err = promptSecret(cmd, "api_access_token: ")
				if err != nil {
					return err
				}
			}

			if apiKey != "" {
				if !noVerify {
					me, err := verifyToken(cmd.Context(), base, apiKey)
					if err != nil {
						return fmt.Errorf("token verification failed: %w", err)
					}
					fmt.Fprintf(cmd.ErrOrStderr(), "verified as %s <%s>\n", me.Name, me.Email)
					prof.UserID = me.ID.String()
					prof.Email = me.Email
					acct, err := chooseAccount(cmd, me, config.FirstNonEmpty(accountID, d.gf.accountID, prof.AccountID))
					if err != nil {
						return err
					}
					prof.AccountID = acct
				} else if v := config.FirstNonEmpty(accountID, d.gf.accountID, prof.AccountID); v != "" {
					prof.AccountID = v
				}
				if err := d.store().Set(profileName, apiKey); err != nil {
					return fmt.Errorf("store token: %w", err)
				}
			}

			if platformToken != "" {
				if err := d.store().Set(auth.PlatformKey(profileName), platformToken); err != nil {
					return fmt.Errorf("store platform token: %w", err)
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "stored platform app token for profile %q\n", profileName)
			}

			if err := cfg.SetProfile(profileName, prof); err != nil {
				return err
			}
			if cfg.CurrentProfile == "" {
				cfg.CurrentProfile = profileName
			}
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "profile %q ready: %s (account %s)\n", profileName, prof.BaseURL, prof.AccountID)
			return nil
		},
	}
	cmd.Flags().StringVar(&apiKey, "api-key", "", "api_access_token (omit to be prompted with hidden input)")
	cmd.Flags().StringVar(&platformToken, "platform-token", "", "platform app token for `wootctl platform …` (optional)")
	cmd.Flags().StringVar(&baseURL, "url", "", "instance base URL (alias of --base-url)")
	cmd.Flags().StringVar(&accountID, "account", "", "account id to scope commands to (alias of --account-id)")
	cmd.Flags().BoolVar(&noVerify, "no-verify", false, "skip the /api/v1/profile verification call")
	return cmd
}

// verifyToken calls the whoami endpoint with a throwaway client (the profile isn't saved
// yet, so getAPIClient can't be used).
func verifyToken(ctx context.Context, base, token string) (*chatwootProfile, error) {
	c := api.New(base, token)
	var me chatwootProfile
	if err := c.GetJSON(ctx, "api/v1/profile", nil, &me); err != nil {
		return nil, err
	}
	return &me, nil
}

// chooseAccount picks the account id: an explicit value wins, a single-account token
// auto-selects, several accounts prompt (or fail with the list when non-interactive).
func chooseAccount(cmd *cobra.Command, me *chatwootProfile, explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	switch len(me.Acc) {
	case 0:
		return "", fmt.Errorf("this token has no account memberships; pass --account <id>")
	case 1:
		fmt.Fprintf(cmd.ErrOrStderr(), "account %s (%s)\n", me.Acc[0].ID, me.Acc[0].Name)
		return me.Acc[0].ID.String(), nil
	}
	fmt.Fprintln(cmd.ErrOrStderr(), "accounts available to this token:")
	for _, a := range me.Acc {
		fmt.Fprintf(cmd.ErrOrStderr(), "  %s  %s (%s)\n", a.ID, a.Name, a.Role)
	}
	choice, err := promptLine(cmd, "account id to use: ")
	if err != nil {
		return "", fmt.Errorf("several accounts available — pass --account <id>: %w", err)
	}
	for _, a := range me.Acc {
		if a.ID.String() == choice {
			return choice, nil
		}
	}
	return "", fmt.Errorf("account %q is not among this token's accounts", choice)
}

func authLogoutCmd(d *deps) *cobra.Command {
	var platformOnly bool
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove stored tokens for the active profile",
		Example: `  wootctl auth logout                  # remove user + platform tokens
  wootctl auth logout --platform-only  # keep the user token`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := d.loadConfig()
			if err != nil {
				return err
			}
			profileName := cfg.ResolveProfileName(d.gf.profile)
			store := d.store()
			if !platformOnly {
				_ = store.Delete(profileName)
			}
			_ = store.Delete(auth.PlatformKey(profileName))
			fmt.Fprintf(cmd.OutOrStdout(), "logged out of profile %q\n", profileName)
			return nil
		},
	}
	cmd.Flags().BoolVar(&platformOnly, "platform-only", false, "remove only the platform app token")
	return cmd
}

func authStatusCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Aliases: []string{"whoami"},
		Short:   "Show the active profile, base URL, identity, and token validity",
		Example: `  wootctl auth status
  wootctl auth whoami -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// A real check: missing/invalid token exits non-zero so `wootctl auth status && …`
			// gates correctly, while still printing why.
			c, cfg, err := d.getAPIClient(true)
			if err != nil {
				return err
			}
			profileName := cfg.ResolveProfileName(d.gf.profile)
			var me chatwootProfile
			if err := c.GetJSON(cmd.Context(), "api/v1/profile", nil, &me); err != nil {
				return fmt.Errorf("token invalid for profile %q: %w", profileName, err)
			}
			if c.DryRun {
				return nil
			}
			accounts := make([]string, 0, len(me.Acc))
			for _, a := range me.Acc {
				accounts = append(accounts, fmt.Sprintf("%s (%s, %s)", a.ID, a.Name, a.Role))
			}
			sort.Strings(accounts)
			_, hasPlatform := firstStoreHit(d, auth.PlatformKey(profileName))
			status := map[string]any{
				"profile":        profileName,
				"base_url":       c.BaseURL,
				"account_id":     c.AccountID,
				"valid":          true,
				"user":           me.Name,
				"email":          me.Email,
				"user_id":        me.ID.String(),
				"accounts":       accounts,
				"token_backend":  d.store().Backend(),
				"platform_token": hasPlatform,
			}
			return d.render(cmd, mustJSON(status), []string{"profile", "base_url", "account_id", "valid", "user", "email"})
		},
	}
}

// firstStoreHit reports whether a keyring entry exists without surfacing its value.
func firstStoreHit(d *deps, key string) (string, bool) {
	v, err := d.store().Get(key)
	return "", err == nil && v != ""
}

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
