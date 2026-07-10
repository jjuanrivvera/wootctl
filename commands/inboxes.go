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
		Use:      "inboxes",
		Aliases:  []string{"inbox"},
		Short:    "Manage inboxes and their members",
		New:      func(c *api.Client) *api.Resource[api.Rec] { return c.Inboxes() },
		Columns:  []string{"id", "name", "channel_type"},
		NoDelete: true, // the API exposes no inbox delete
		CreateFields: []field{
			{Flag: "name", Usage: "inbox name", Required: true},
			{Flag: "channel", Kind: fieldJSON, Usage: `channel config, e.g. '{"type":"web_widget","website_url":"https://invitas.co"}' or '{"type":"api","webhook_url":"https://…"}'`},
			{Flag: "greeting-enabled", Kind: fieldBool, Usage: "send a greeting on first message"},
			{Flag: "greeting-message", Usage: "greeting text"},
			{Flag: "enable-auto-assignment", Kind: fieldBool, Usage: "auto-assign conversations"},
			{Flag: "working-hours-enabled", Kind: fieldBool, Usage: "enable working hours"},
			{Flag: "out-of-office-message", Usage: "out-of-office reply"},
			{Flag: "timezone", Usage: "IANA timezone"},
			{Flag: "csat-survey-enabled", Kind: fieldBool, Usage: "send CSAT surveys on resolve"},
			{Flag: "lock-to-single-conversation", Kind: fieldBool, Usage: "one conversation per contact"},
			{Flag: "portal-id", Kind: fieldInt, Usage: "help-center portal to link"},
		},
		Extra: []extraCommand{
			{Kind: kindRead, Build: inboxAgentBotCmd},
			{Kind: kindWrite, Build: inboxSetAgentBotCmd},
			{Kind: kindRead, Build: inboxMembersCmd},
			{Kind: kindWrite, Build: inboxMembersEditCmd("add-members", http.MethodPost, "Add agents to an inbox")},
			{Kind: kindWrite, Build: inboxMembersEditCmd("update-members", http.MethodPatch, "Replace an inbox's agents")},
			{Kind: kindDestructive, Build: inboxMembersEditCmd("remove-members", http.MethodDelete, "Remove agents from an inbox")},
		},
	})
}

func inboxAgentBotCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "agent-bot <inbox-id>",
		Short:   "Show the agent bot attached to an inbox",
		Example: "  cwctl inboxes agent-bot 3",
		Args:    cobra.ExactArgs(1),
		RunE: runE(d, false, []string{"id", "name", "description"}, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Inboxes().Action(cmd.Context(), http.MethodGet, url.PathEscape(args[0])+"/agent_bot", nil, nil, &out)
			return out, err
		}),
	}
}

func inboxSetAgentBotCmd(d *deps) *cobra.Command {
	var botID int
	cmd := &cobra.Command{
		Use:     "set-agent-bot <inbox-id>",
		Short:   "Attach an agent bot to an inbox (0 detaches)",
		Example: "  cwctl inboxes set-agent-bot 3 --agent-bot 1",
		Args:    cobra.ExactArgs(1),
		RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Inboxes().Action(cmd.Context(), http.MethodPost, url.PathEscape(args[0])+"/set_agent_bot", nil,
				map[string]any{"agent_bot": botID}, &out)
			return out, err
		}),
	}
	cmd.Flags().IntVar(&botID, "agent-bot", 0, "agent bot id")
	_ = cmd.MarkFlagRequired("agent-bot")
	return cmd
}

// inboxMembersPath is the odd one out: members live under /inbox_members, not /inboxes.
func inboxMembersPath(c *api.Client) string { return c.AccountPath("inbox_members") }

func inboxMembersCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "members <inbox-id>",
		Short:   "List the agents in an inbox",
		Example: "  cwctl inboxes members 3",
		Args:    cobra.ExactArgs(1),
		RunE: runE(d, false, []string{"id", "name", "email", "role"}, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Send(cmd.Context(), http.MethodGet, inboxMembersPath(c)+"/"+url.PathEscape(args[0]), nil, nil, &out)
			return out, err
		}),
	}
}

// inboxMembersEditCmd builds the three write verbs sharing one body: {inbox_id, user_ids}.
// The inbox id rides the BODY (not the path) on these endpoints, DELETE included.
func inboxMembersEditCmd(use, method, short string) func(d *deps) *cobra.Command {
	return func(d *deps) *cobra.Command {
		var userIDs []int
		cmd := &cobra.Command{
			Use:     use + " <inbox-id>",
			Short:   short,
			Example: "  cwctl inboxes " + use + " 3 --user-ids 1,2",
			Args:    cobra.ExactArgs(1),
			RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
				var out json.RawMessage
				err := c.Send(cmd.Context(), method, inboxMembersPath(c), nil,
					map[string]any{"inbox_id": args[0], "user_ids": userIDs}, &out)
				return out, err
			}),
		}
		cmd.Flags().IntSliceVar(&userIDs, "user-ids", nil, "agent user ids (comma-separated)")
		_ = cmd.MarkFlagRequired("user-ids")
		return cmd
	}
}
