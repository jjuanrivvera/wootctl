package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/wootctl/internal/api"
)

// csat wraps the unauthenticated CSAT survey page endpoint. The swagger's own guidance is
// "redirect the client to this URL", so the default is to print the URL; --fetch retrieves
// the page itself (the one non-JSON endpoint, DECISIONS.md #13).
func init() {
	registerResource("", resourceSpec[api.Rec]{
		Use:   "csat",
		Short: "CSAT survey page for a conversation",
		New: func(c *api.Client) *api.Resource[api.Rec] {
			return api.NewResource[api.Rec](c, "survey/responses")
		},
		NoList: true, NoGet: true, NoCreate: true, NoUpdate: true, NoDelete: true,
		Extra: []extraCommand{
			{Kind: kindRead, Build: csatPageCmd},
		},
	})
}

func csatPageCmd(d *deps) *cobra.Command {
	var fetch bool
	cmd := &cobra.Command{
		Use:   "page <conversation-uuid>",
		Short: "Print (or fetch with --fetch) the CSAT survey page URL for a conversation",
		Example: `  wootctl csat page 8f286537-3216-4d47-a869-6a08128d9dc9
  wootctl csat page 8f286537-3216-4d47-a869-6a08128d9dc9 --fetch > survey.html`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, _, err := d.getAPIClient(false)
			if err != nil {
				return err
			}
			path := "survey/responses/" + url.PathEscape(args[0])
			if !fetch {
				fmt.Fprintln(cmd.OutOrStdout(), c.BaseURL+"/"+path)
				return nil
			}
			status, _, body, err := c.Raw(cmd.Context(), http.MethodGet, path, nil, nil)
			if err != nil {
				return err
			}
			if status == 0 { // dry-run
				return nil
			}
			if json.Valid(body) {
				return d.render(cmd, json.RawMessage(body), nil)
			}
			_, err = cmd.OutOrStdout().Write(body)
			return err
		},
	}
	cmd.Flags().BoolVar(&fetch, "fetch", false, "fetch the page instead of printing its URL")
	return cmd
}
