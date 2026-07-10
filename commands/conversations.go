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
		Use:      "conversations",
		Aliases:  []string{"conv", "convs"},
		Short:    "Manage conversations",
		New:      func(c *api.Client) *api.Resource[api.Rec] { return c.Conversations() },
		Columns:  []string{"id", "inbox_id", "status", "priority"},
		NoDelete: true, // the API exposes no conversation delete
		ListFilters: []listFilter{
			{Flag: "status", Usage: "open | resolved | pending | snoozed"},
			{Flag: "assignee-type", Usage: "me | unassigned | all | assigned"},
			{Flag: "inbox-id", Usage: "only this inbox"},
			{Flag: "team-id", Usage: "only this team"},
			{Flag: "labels", Usage: "comma-separated labels"},
			{Flag: "q", Usage: "text search in messages"},
		},
		CreateFields: []field{
			{Flag: "source-id", Usage: "contact-inbox source id", Required: true},
			{Flag: "inbox-id", Kind: fieldInt, Usage: "inbox for the conversation"},
			{Flag: "contact-id", Kind: fieldInt, Usage: "contact for the conversation"},
			{Flag: "status", Usage: "open | resolved | pending"},
			{Flag: "assignee-id", Kind: fieldInt, Usage: "agent to assign"},
			{Flag: "team-id", Kind: fieldInt, Usage: "team to assign"},
			{Flag: "message", Kind: fieldJSON, Usage: `first message, e.g. '{"content":"hola"}'`},
			{Flag: "custom-attributes", Kind: fieldJSON, Usage: "custom attributes object"},
			{Flag: "additional-attributes", Kind: fieldJSON, Usage: "additional attributes object"},
		},
		UpdateFields: []field{
			{Flag: "priority", Usage: "urgent | high | medium | low | none"},
			{Flag: "sla-policy-id", Kind: fieldInt, Usage: "SLA policy id (Enterprise)"},
		},
		Extra: []extraCommand{
			{Kind: kindRead, Build: convMetaCmd},
			{Kind: kindRead, Build: convFilterCmd},
			{Kind: kindRead, Build: convLabelsCmd},
			{Kind: kindWrite, Build: convAddLabelsCmd},
			{Kind: kindWrite, Build: convAssignCmd},
			{Kind: kindWrite, Build: convToggleStatusCmd},
			{Kind: kindWrite, Build: convTogglePriorityCmd},
			{Kind: kindWrite, Build: convToggleTypingCmd},
			{Kind: kindWrite, Build: convSetCustomAttributesCmd},
			{Kind: kindRead, Build: convReportingEventsCmd},
		},
	})
}

func convMetaCmd(d *deps) *cobra.Command {
	var status, q, inboxID, teamID, labels string
	cmd := &cobra.Command{
		Use:     "meta",
		Short:   "Conversation counts (mine, unassigned, assigned, all)",
		Example: "  cwctl conversations meta --status open",
		Args:    cobra.NoArgs,
		RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, _ []string) (json.RawMessage, error) {
			query := url.Values{}
			for k, v := range map[string]string{"status": status, "q": q, "inbox_id": inboxID, "team_id": teamID, "labels": labels} {
				if v != "" {
					query.Set(k, v)
				}
			}
			var out json.RawMessage
			err := c.Conversations().Action(cmd.Context(), http.MethodGet, "meta", query, nil, &out)
			return out, err
		}),
	}
	cmd.Flags().StringVar(&status, "status", "", "open | resolved | pending | snoozed")
	cmd.Flags().StringVar(&q, "q", "", "text search")
	cmd.Flags().StringVar(&inboxID, "inbox-id", "", "only this inbox")
	cmd.Flags().StringVar(&teamID, "team-id", "", "only this team")
	cmd.Flags().StringVar(&labels, "labels", "", "comma-separated labels")
	return cmd
}

func convFilterCmd(d *deps) *cobra.Command {
	var payload string
	cmd := &cobra.Command{
		Use:   "filter",
		Short: "Filter conversations with the query DSL",
		Example: `  cwctl conversations filter --payload '[{"attribute_key":"status","filter_operator":"equal_to","values":["pending"]}]'
  cwctl conversations filter --payload @filter.json`,
		Args: cobra.NoArgs,
		RunE: runE(d, false, []string{"id", "inbox_id", "status", "priority"}, func(cmd *cobra.Command, c *api.Client, _ []string) (json.RawMessage, error) {
			raw, err := readDataArg(cmd, payload)
			if err != nil {
				return nil, err
			}
			var arr []any
			if err := json.Unmarshal(raw, &arr); err != nil {
				return nil, err
			}
			query := url.Values{}
			if d.gf.page > 0 {
				query.Set("page", itoa(d.gf.page))
			}
			var out json.RawMessage
			err = c.Conversations().Action(cmd.Context(), http.MethodPost, "filter", query, map[string]any{"payload": arr}, &out)
			return out, err
		}),
	}
	cmd.Flags().StringVar(&payload, "payload", "", "filter conditions JSON array: inline, @file, or - for stdin")
	_ = cmd.MarkFlagRequired("payload")
	return cmd
}

func convLabelsCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "labels <conversation-id>",
		Short:   "List a conversation's labels",
		Example: "  cwctl conversations labels 42",
		Args:    cobra.ExactArgs(1),
		RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Conversations().Action(cmd.Context(), http.MethodGet, url.PathEscape(args[0])+"/labels", nil, nil, &out)
			return out, err
		}),
	}
}

func convAddLabelsCmd(d *deps) *cobra.Command {
	var labels []string
	cmd := &cobra.Command{
		Use:     "add-labels <conversation-id>",
		Short:   "Add labels to a conversation (replaces the label set)",
		Example: "  cwctl conversations add-labels 42 --labels billing,urgente",
		Args:    cobra.ExactArgs(1),
		RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Conversations().Action(cmd.Context(), http.MethodPost, url.PathEscape(args[0])+"/labels", nil,
				map[string]any{"labels": labels}, &out)
			return out, err
		}),
	}
	cmd.Flags().StringSliceVar(&labels, "labels", nil, "labels to set (comma-separated)")
	_ = cmd.MarkFlagRequired("labels")
	return cmd
}

func convAssignCmd(d *deps) *cobra.Command {
	var assigneeID, teamID int
	cmd := &cobra.Command{
		Use:   "assign <conversation-id>",
		Short: "Assign a conversation to an agent or a team",
		Example: `  cwctl conversations assign 42 --assignee-id 7
  cwctl conversations assign 42 --team-id 2`,
		Args: cobra.ExactArgs(1),
		RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			body := map[string]any{}
			if cmd.Flags().Changed("assignee-id") {
				body["assignee_id"] = assigneeID
			}
			if cmd.Flags().Changed("team-id") {
				body["team_id"] = teamID
			}
			if len(body) == 0 {
				return nil, errMissingOneOf("--assignee-id", "--team-id")
			}
			var out json.RawMessage
			err := c.Conversations().Action(cmd.Context(), http.MethodPost, url.PathEscape(args[0])+"/assignments", nil, body, &out)
			return out, err
		}),
	}
	cmd.Flags().IntVar(&assigneeID, "assignee-id", 0, "agent user id (set 0 to unassign)")
	cmd.Flags().IntVar(&teamID, "team-id", 0, "team id (ignored when --assignee-id is present)")
	return cmd
}

func convToggleStatusCmd(d *deps) *cobra.Command {
	var status string
	var snoozedUntil int
	cmd := &cobra.Command{
		Use:   "toggle-status <conversation-id>",
		Short: "Change a conversation's status (open/resolved/pending/snoozed)",
		Example: `  cwctl conversations toggle-status 42 --status resolved
  cwctl conversations toggle-status 42 --status snoozed --snoozed-until 1757506877`,
		Args: cobra.ExactArgs(1),
		RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			body := map[string]any{"status": status}
			if cmd.Flags().Changed("snoozed-until") {
				body["snoozed_until"] = snoozedUntil
			}
			var out json.RawMessage
			err := c.Conversations().Action(cmd.Context(), http.MethodPost, url.PathEscape(args[0])+"/toggle_status", nil, body, &out)
			return out, err
		}),
	}
	cmd.Flags().StringVar(&status, "status", "", "open | resolved | pending | snoozed")
	cmd.Flags().IntVar(&snoozedUntil, "snoozed-until", 0, "unix seconds to reopen a snoozed conversation")
	_ = cmd.MarkFlagRequired("status")
	return cmd
}

func convTogglePriorityCmd(d *deps) *cobra.Command {
	var priority string
	cmd := &cobra.Command{
		Use:     "toggle-priority <conversation-id>",
		Short:   "Change a conversation's priority",
		Example: "  cwctl conversations toggle-priority 42 --priority urgent",
		Args:    cobra.ExactArgs(1),
		RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Conversations().Action(cmd.Context(), http.MethodPost, url.PathEscape(args[0])+"/toggle_priority", nil,
				map[string]any{"priority": priority}, &out)
			return out, err
		}),
	}
	cmd.Flags().StringVar(&priority, "priority", "", "urgent | high | medium | low | none")
	_ = cmd.MarkFlagRequired("priority")
	return cmd
}

func convToggleTypingCmd(d *deps) *cobra.Command {
	var typing string
	var private bool
	cmd := &cobra.Command{
		Use:     "toggle-typing <conversation-id>",
		Short:   "Flip the typing indicator on or off",
		Example: "  cwctl conversations toggle-typing 42 --typing-status on",
		Args:    cobra.ExactArgs(1),
		RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			body := map[string]any{"typing_status": typing}
			if private {
				body["is_private"] = true
			}
			var out json.RawMessage
			err := c.Conversations().Action(cmd.Context(), http.MethodPost, url.PathEscape(args[0])+"/toggle_typing_status", nil, body, &out)
			return out, err
		}),
	}
	cmd.Flags().StringVar(&typing, "typing-status", "", "on | off")
	cmd.Flags().BoolVar(&private, "private", false, "typing indicator for private notes")
	_ = cmd.MarkFlagRequired("typing-status")
	return cmd
}

func convSetCustomAttributesCmd(d *deps) *cobra.Command {
	var attrs string
	cmd := &cobra.Command{
		Use:     "set-custom-attributes <conversation-id>",
		Short:   "Set custom attributes on a conversation",
		Example: `  cwctl conversations set-custom-attributes 42 --attributes '{"order_id":"12345"}'`,
		Args:    cobra.ExactArgs(1),
		RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var obj map[string]any
			if err := json.Unmarshal([]byte(attrs), &obj); err != nil {
				return nil, err
			}
			var out json.RawMessage
			err := c.Conversations().Action(cmd.Context(), http.MethodPost, url.PathEscape(args[0])+"/custom_attributes", nil,
				map[string]any{"custom_attributes": obj}, &out)
			return out, err
		}),
	}
	cmd.Flags().StringVar(&attrs, "attributes", "", "custom attributes JSON object")
	_ = cmd.MarkFlagRequired("attributes")
	return cmd
}

func convReportingEventsCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "reporting-events <conversation-id>",
		Short:   "List a conversation's reporting events (first response, resolved, …)",
		Example: "  cwctl conversations reporting-events 42",
		Args:    cobra.ExactArgs(1),
		RunE: runE(d, false, []string{"id", "name", "value", "created_at"}, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Conversations().Action(cmd.Context(), http.MethodGet, url.PathEscape(args[0])+"/reporting_events", nil, nil, &out)
			return out, err
		}),
	}
}
