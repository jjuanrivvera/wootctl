package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	metaRegistrars = append(metaRegistrars, func(d *deps) *cobra.Command {
		return &cobra.Command{
			Use:   "init",
			Short: "First-run setup wizard",
			Long: `Walk through connecting cwctl to a Chatwoot instance: base URL, api_access_token
(hidden input), account selection, and a live verification — then a smoke check.
init is auth login plus a doctor pass; re-running it is safe.`,
			Example: `  cwctl init
  cwctl --profile staging init`,
			Args: cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				fmt.Fprintln(cmd.ErrOrStderr(), "cwctl setup — connect to your Chatwoot instance")
				login := authLoginCmd(d)
				login.SetIn(cmd.InOrStdin())
				login.SetOut(cmd.OutOrStdout())
				login.SetErr(cmd.ErrOrStderr())
				login.SetContext(cmd.Context())
				if err := login.RunE(login, nil); err != nil {
					return err
				}
				fmt.Fprintln(cmd.ErrOrStderr(), "\nsmoke check:")
				c, _, err := d.getAPIClient(true)
				if err != nil {
					return err
				}
				var me chatwootProfile
				if err := c.GetJSON(cmd.Context(), "api/v1/profile", nil, &me); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "all set — authenticated as %s <%s> on %s (account %s)\n",
					me.Name, me.Email, c.BaseURL, c.AccountID)
				fmt.Fprintln(cmd.ErrOrStderr(), "\ntry:  cwctl conversations list --status open")
				return nil
			},
		}
	})
}
