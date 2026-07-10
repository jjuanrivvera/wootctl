package commands

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/cwctl/internal/api"
)

func init() {
	registerResource("", resourceSpec[api.Rec]{
		Use:     "teams",
		Aliases: []string{"team"},
		Short:   "Manage teams and their members",
		New:     func(c *api.Client) *api.Resource[api.Rec] { return c.Teams() },
		Columns: []string{"id", "name", "allow_auto_assign"},
		CreateFields: []field{
			{Flag: "name", Usage: "team name", Required: true},
			{Flag: "description", Usage: "team description"},
			{Flag: "allow-auto-assign", Kind: fieldBool, Usage: "auto-assign conversations to team members"},
		},
		Extra: []extraCommand{
			{Kind: kindRead, Build: teamMembersCmd},
			{Kind: kindWrite, Build: teamMembersEditCmd("add-members", http.MethodPost, "Add agents to a team")},
			{Kind: kindWrite, Build: teamMembersEditCmd("update-members", http.MethodPatch, "Replace a team's agents")},
			{Kind: kindDestructive, Build: teamMembersEditCmd("remove-members", http.MethodDelete, "Remove agents from a team")},
		},
	})
}

func teamMembersCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "members <team-id>",
		Short:   "List the agents in a team",
		Example: "  cwctl teams members 3",
		Args:    cobra.ExactArgs(1),
		RunE: runE(d, false, []string{"id", "name", "email", "role"}, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Teams().Action(cmd.Context(), http.MethodGet, url.PathEscape(args[0])+"/team_members", nil, nil, &out)
			return out, err
		}),
	}
}

// teamMembersEditCmd builds the three write verbs that share one body shape: {user_ids}.
func teamMembersEditCmd(use, method, short string) func(d *deps) *cobra.Command {
	return func(d *deps) *cobra.Command {
		var userIDs []int
		cmd := &cobra.Command{
			Use:     use + " <team-id>",
			Short:   short,
			Example: "  cwctl teams " + use + " 3 --user-ids 1,2",
			Args:    cobra.ExactArgs(1),
			RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
				var out json.RawMessage
				err := c.Teams().Action(cmd.Context(), method, url.PathEscape(args[0])+"/team_members", nil,
					map[string]any{"user_ids": userIDs}, &out)
				return out, err
			}),
		}
		cmd.Flags().IntSliceVar(&userIDs, "user-ids", nil, "agent user ids (comma-separated)")
		_ = cmd.MarkFlagRequired("user-ids")
		return cmd
	}
}
