package commands

import (
	"encoding/json"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/wootctl/internal/api"
)

// profile wraps the account-independent /api/v1/profile pair (whoami + self-update).
func init() {
	registerResource("", resourceSpec[api.Rec]{
		Use:     "profile",
		Aliases: []string{"me"},
		Short:   "Read and update your own user profile",
		New:     func(c *api.Client) *api.Resource[api.Rec] { return api.NewResource[api.Rec](c, "api/v1/profile") },
		Columns: []string{"id", "name", "email", "role"},
		NoList:  true, NoGet: true, NoCreate: true, NoUpdate: true, NoDelete: true,
		Extra: []extraCommand{
			{Kind: kindRead, Build: profileGetCmd},
			{Kind: kindWrite, Build: profileUpdateCmd},
		},
	})
}

func profileGetCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "get",
		Short:   "Get your own profile (the token's identity)",
		Example: "  wootctl profile get -o json",
		Args:    cobra.NoArgs,
		RunE: runE(d, false, []string{"id", "name", "email", "role"}, func(cmd *cobra.Command, c *api.Client, _ []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Send(cmd.Context(), http.MethodGet, "api/v1/profile", nil, nil, &out)
			return out, err
		}),
	}
}

func profileUpdateCmd(d *deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update your own profile (name, signature, password, …)",
		Example: `  wootctl profile update --display-name "Juan R."
  wootctl profile update --message-signature "— Juan, Soporte"`,
		Args: cobra.NoArgs,
	}
	collect := registerBodyFlags(cmd, []field{
		{Flag: "name", Usage: "full name"},
		{Flag: "display-name", Usage: "name agents see"},
		{Flag: "email", Usage: "login email"},
		{Flag: "message-signature", Usage: "signature appended to outgoing replies"},
		{Flag: "phone-number", Usage: "phone number"},
		{Flag: "current-password", Usage: "current password (required to change the password)"},
		{Flag: "password", Usage: "new password"},
		{Flag: "password-confirmation", Usage: "new password again"},
	})
	cmd.RunE = runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, _ []string) (json.RawMessage, error) {
		fieldsBody, err := collect(cmd)
		if err != nil {
			return nil, err
		}
		// The API nests the update under a "profile" key (PUT /api/v1/profile).
		var out json.RawMessage
		err = c.Send(cmd.Context(), http.MethodPut, "api/v1/profile", nil, map[string]any{"profile": fieldsBody}, &out)
		return out, err
	})
	return cmd
}
