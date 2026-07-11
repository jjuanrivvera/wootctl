package commands

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/wootctl/internal/api"
)

// account is Extra-only: its endpoint IS the account-scope path itself
// (GET/PATCH /api/v1/accounts/{account_id}), so the generic collection builder can't apply.
func init() {
	registerResource("", resourceSpec[api.Rec]{
		Use:     "account",
		Aliases: []string{"acct"},
		Short:   "Read and update the current account",
		New:     func(c *api.Client) *api.Resource[api.Rec] { return api.NewResource[api.Rec](c, accountSelfPath(c)) },
		Columns: []string{"id", "name", "locale", "domain"},
		NoList:  true, NoGet: true, NoCreate: true, NoUpdate: true, NoDelete: true,
		Extra: []extraCommand{
			{Kind: kindRead, Build: accountGetCmd},
			{Kind: kindWrite, Build: accountUpdateCmd},
		},
	})
}

func accountSelfPath(c *api.Client) string {
	return strings.TrimRight(c.AccountPath(""), "/")
}

func accountGetCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "get",
		Short:   "Get the current account's details",
		Example: "  wootctl account get -o yaml",
		Args:    cobra.NoArgs,
		RunE: runE(d, false, []string{"id", "name", "locale", "domain"}, func(cmd *cobra.Command, c *api.Client, _ []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Send(cmd.Context(), http.MethodGet, accountSelfPath(c), nil, nil, &out)
			return out, err
		}),
	}
}

func accountUpdateCmd(d *deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the current account",
		Example: `  wootctl account update --name "Soporte Invitas"
  wootctl account update -d '{"locale":"es","timezone":"America/Bogota"}'`,
		Args: cobra.NoArgs,
	}
	collect := registerBodyFlags(cmd, []field{
		{Flag: "name", Usage: "account name"},
		{Flag: "locale", Usage: "default language, e.g. en, es"},
		{Flag: "domain", Usage: "account domain"},
		{Flag: "support-email", Usage: "support email address"},
		{Flag: "timezone", Usage: "IANA timezone, e.g. America/Bogota"},
		{Flag: "industry", Usage: "industry label"},
		{Flag: "company-size", Usage: "company size bucket"},
		{Flag: "auto-resolve-after", Kind: fieldInt, Usage: "minutes of inactivity before auto-resolve"},
	})
	cmd.RunE = runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, _ []string) (json.RawMessage, error) {
		body, err := collect(cmd)
		if err != nil {
			return nil, err
		}
		var out json.RawMessage
		err = c.Send(cmd.Context(), http.MethodPatch, accountSelfPath(c), nil, body, &out)
		return out, err
	})
	return cmd
}
