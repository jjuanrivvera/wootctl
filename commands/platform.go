package commands

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/cwctl/internal/api"
)

// The platform API provisions accounts, users, and agent bots on a self-hosted install.
// Every command here authenticates with the profile's PLATFORM app token (selected by the
// /platform path prefix) — `cwctl auth login --platform-token <t>` stores it.

func init() {
	registerResource("platform", resourceSpec[api.Rec]{
		Use:      "accounts",
		Short:    "Provision accounts (platform app token)",
		New:      func(c *api.Client) *api.Resource[api.Rec] { return c.PlatformAccounts() },
		Columns:  []string{"id", "name", "locale"},
		NoList:   true, // the platform API exposes no account list
		GetByArg: "account-id",
		CreateFields: []field{
			{Flag: "name", Usage: "account name", Required: true},
			{Flag: "locale", Usage: "default language, e.g. en, es"},
			{Flag: "domain", Usage: "account domain"},
			{Flag: "support-email", Usage: "support email"},
			{Flag: "status", Usage: "active | suspended"},
			{Flag: "limits", Kind: fieldJSON, Usage: `usage limits, e.g. '{"agents":10,"inboxes":5}'`},
			{Flag: "custom-attributes", Kind: fieldJSON, Usage: "custom attributes object"},
		},
	})

	registerResource("platform", resourceSpec[api.Rec]{
		Use:     "account-users",
		Short:   "Manage the users of an account (platform app token)",
		New:     func(c *api.Client) *api.Resource[api.Rec] { return c.PlatformAccounts() },
		Columns: []string{"account_id", "user_id", "role"},
		NoList:  true, NoGet: true, NoCreate: true, NoUpdate: true, NoDelete: true,
		Extra: []extraCommand{
			{Kind: kindRead, Build: platformAccountUsersListCmd},
			{Kind: kindWrite, Build: platformAccountUsersCreateCmd},
			{Kind: kindDestructive, Build: platformAccountUsersDeleteCmd},
		},
	})

	registerResource("platform", resourceSpec[api.Rec]{
		Use:     "agent-bots",
		Short:   "Manage global agent bots (platform app token)",
		New:     func(c *api.Client) *api.Resource[api.Rec] { return c.PlatformAgentBots() },
		Columns: []string{"id", "name", "description"},
		CreateFields: []field{
			{Flag: "name", Usage: "bot name", Required: true},
			{Flag: "description", Usage: "bot description"},
			{Flag: "outgoing-url", Usage: "webhook URL the bot receives events on"},
			{Flag: "account-id", Kind: fieldInt, Usage: "restrict the bot to one account"},
		},
	})

	registerResource("platform", resourceSpec[api.Rec]{
		Use:     "users",
		Short:   "Provision users (platform app token)",
		New:     func(c *api.Client) *api.Resource[api.Rec] { return c.PlatformUsers() },
		Columns: []string{"id", "name", "email"},
		NoList:  true, // the platform API exposes no user list
		CreateFields: []field{
			{Flag: "name", Usage: "user's full name", Required: true},
			{Flag: "email", Usage: "login email", Required: true},
			{Flag: "password", Usage: "initial password"},
			{Flag: "display-name", Usage: "display name"},
			{Flag: "custom-attributes", Kind: fieldJSON, Usage: "custom attributes object"},
		},
		Extra: []extraCommand{
			{Kind: kindRead, Build: platformUserSSOLinkCmd},
		},
	})
}

func platformAccountUsersPath(accountID string) string {
	return "platform/api/v1/accounts/" + url.PathEscape(accountID) + "/account_users"
}

func platformAccountUsersListCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "list <account-id>",
		Short:   "List an account's users",
		Example: "  cwctl platform account-users list 2",
		Args:    cobra.ExactArgs(1),
		RunE: runE(d, false, []string{"account_id", "user_id", "role"}, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Send(cmd.Context(), http.MethodGet, platformAccountUsersPath(args[0]), nil, nil, &out)
			return out, err
		}),
	}
}

func platformAccountUsersCreateCmd(d *deps) *cobra.Command {
	var userID int
	var role string
	cmd := &cobra.Command{
		Use:     "create <account-id>",
		Short:   "Add a user to an account",
		Example: "  cwctl platform account-users create 2 --user-id 7 --role agent",
		Args:    cobra.ExactArgs(1),
		RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Send(cmd.Context(), http.MethodPost, platformAccountUsersPath(args[0]), nil,
				map[string]any{"user_id": userID, "role": role}, &out)
			return out, err
		}),
	}
	cmd.Flags().IntVar(&userID, "user-id", 0, "user id to add")
	cmd.Flags().StringVar(&role, "role", "", "agent | administrator")
	_ = cmd.MarkFlagRequired("user-id")
	_ = cmd.MarkFlagRequired("role")
	return cmd
}

func platformAccountUsersDeleteCmd(d *deps) *cobra.Command {
	var userID int
	cmd := &cobra.Command{
		Use:     "delete <account-id>",
		Short:   "Remove a user from an account",
		Example: "  cwctl platform account-users delete 2 --user-id 7",
		Args:    cobra.ExactArgs(1),
		RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			// The user id rides the DELETE body on this endpoint.
			err := c.Send(cmd.Context(), http.MethodDelete, platformAccountUsersPath(args[0]), nil,
				map[string]any{"user_id": userID}, nil)
			if err != nil {
				return nil, err
			}
			if !d.gf.quiet && !c.DryRun {
				cmd.Printf("removed user %d from account %s\n", userID, args[0])
			}
			return nil, nil
		}),
	}
	cmd.Flags().IntVar(&userID, "user-id", 0, "user id to remove")
	_ = cmd.MarkFlagRequired("user-id")
	return cmd
}

func platformUserSSOLinkCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "sso-link <id>",
		Short:   "Get a user's one-time SSO login URL",
		Example: "  cwctl platform users sso-link 7",
		Args:    cobra.ExactArgs(1),
		RunE: runE(d, false, []string{"url"}, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.PlatformUsers().Action(cmd.Context(), http.MethodGet, url.PathEscape(args[0])+"/login", nil, nil, &out)
			return out, err
		}),
	}
}
