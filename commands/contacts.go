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
		Use:     "contacts",
		Aliases: []string{"contact"},
		Short:   "Manage contacts",
		New:     func(c *api.Client) *api.Resource[api.Rec] { return c.Contacts() },
		Columns: []string{"id", "name", "email", "phone_number"},
		CreateFields: []field{
			{Flag: "inbox-id", Kind: fieldInt, Usage: "inbox to attach the contact to", Required: true},
			{Flag: "name", Usage: "contact name"},
			{Flag: "email", Usage: "email address"},
			{Flag: "phone-number", Usage: "phone in E.164, e.g. +573001112233"},
			{Flag: "identifier", Usage: "external unique identifier"},
			{Flag: "avatar-url", Usage: "URL of an avatar image"},
			{Flag: "blocked", Kind: fieldBool, Usage: "block the contact"},
			{Flag: "custom-attributes", Kind: fieldJSON, Usage: "custom attributes object"},
			{Flag: "additional-attributes", Kind: fieldJSON, Usage: "additional attributes object"},
		},
		UpdateFields: []field{
			{Flag: "name", Usage: "contact name"},
			{Flag: "email", Usage: "email address"},
			{Flag: "phone-number", Usage: "phone in E.164"},
			{Flag: "identifier", Usage: "external unique identifier"},
			{Flag: "avatar-url", Usage: "URL of an avatar image"},
			{Flag: "blocked", Kind: fieldBool, Usage: "block/unblock the contact"},
			{Flag: "custom-attributes", Kind: fieldJSON, Usage: "custom attributes object"},
			{Flag: "additional-attributes", Kind: fieldJSON, Usage: "additional attributes object"},
		},
		ListFilters: []listFilter{},
		Extra: []extraCommand{
			{Kind: kindRead, Build: contactSearchCmd},
			{Kind: kindRead, Build: contactFilterCmd},
			{Kind: kindWrite, Build: contactMergeCmd},
			{Kind: kindRead, Build: contactConversationsCmd},
			{Kind: kindRead, Build: contactContactableInboxesCmd},
			{Kind: kindWrite, Build: contactCreateContactInboxCmd},
			{Kind: kindRead, Build: contactLabelsCmd},
			{Kind: kindWrite, Build: contactAddLabelsCmd},
		},
	})
}

func contactSearchCmd(d *deps) *cobra.Command {
	var q string
	cmd := &cobra.Command{
		Use:     "search",
		Short:   "Search contacts by name, identifier, email, or phone",
		Example: `  cwctl contacts search --q ana\n  cwctl contacts search --q "+57300" -o json`,
		Args:    cobra.NoArgs,
		RunE: runE(d, false, []string{"id", "name", "email", "phone_number"}, func(cmd *cobra.Command, c *api.Client, _ []string) (json.RawMessage, error) {
			var out json.RawMessage
			query := url.Values{"q": {q}}
			if d.gf.page > 0 {
				query.Set("page", itoa(d.gf.page))
			}
			if d.gf.sort != "" {
				query.Set("sort", d.gf.sort)
			}
			err := c.Contacts().Action(cmd.Context(), http.MethodGet, "search", query, nil, &out)
			return out, err
		}),
	}
	cmd.Flags().StringVar(&q, "q", "", "search term (name, email, phone, identifier)")
	_ = cmd.MarkFlagRequired("q")
	return cmd
}

func contactFilterCmd(d *deps) *cobra.Command {
	var payload string
	cmd := &cobra.Command{
		Use:   "filter",
		Short: "Filter contacts with the query DSL",
		Long:  "POST a filter payload: an array of {attribute_key, filter_operator, values, query_operator} objects.",
		Example: `  cwctl contacts filter --payload '[{"attribute_key":"country_code","filter_operator":"equal_to","values":["CO"]}]'
  cwctl contacts filter --payload @filter.json`,
		Args: cobra.NoArgs,
		RunE: runE(d, false, []string{"id", "name", "email", "phone_number"}, func(cmd *cobra.Command, c *api.Client, _ []string) (json.RawMessage, error) {
			raw, err := readDataArg(cmd, payload)
			if err != nil {
				return nil, err
			}
			var arr []any
			if err := json.Unmarshal(raw, &arr); err != nil {
				return nil, err
			}
			var out json.RawMessage
			query := url.Values{}
			if d.gf.page > 0 {
				query.Set("page", itoa(d.gf.page))
			}
			err = c.Contacts().Action(cmd.Context(), http.MethodPost, "filter", query, map[string]any{"payload": arr}, &out)
			return out, err
		}),
	}
	cmd.Flags().StringVar(&payload, "payload", "", "filter conditions JSON array: inline, @file, or - for stdin")
	_ = cmd.MarkFlagRequired("payload")
	return cmd
}

func contactMergeCmd(d *deps) *cobra.Command {
	var base, mergee int
	cmd := &cobra.Command{
		Use:   "merge",
		Short: "Merge two contacts (the mergee is deleted)",
		Long:  "Merge the mergee contact into the base contact. The base keeps all conversations and data; the mergee is deleted. This cannot be undone.",
		Example: `  cwctl contacts merge --base 1 --mergee 2
  cwctl contacts merge --base 1 --mergee 2 --dry-run`,
		Args: cobra.NoArgs,
		RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, _ []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Send(cmd.Context(), http.MethodPost, c.AccountPath("actions/contact_merge"), nil,
				map[string]any{"base_contact_id": base, "mergee_contact_id": mergee}, &out)
			return out, err
		}),
	}
	cmd.Flags().IntVar(&base, "base", 0, "contact id that remains after the merge")
	cmd.Flags().IntVar(&mergee, "mergee", 0, "contact id merged into the base and deleted")
	_ = cmd.MarkFlagRequired("base")
	_ = cmd.MarkFlagRequired("mergee")
	return cmd
}

func contactConversationsCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "conversations <id>",
		Short:   "List a contact's conversations",
		Example: "  cwctl contacts conversations 12",
		Args:    cobra.ExactArgs(1),
		RunE: runE(d, false, []string{"id", "inbox_id", "status"}, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Contacts().Action(cmd.Context(), http.MethodGet, url.PathEscape(args[0])+"/conversations", nil, nil, &out)
			return out, err
		}),
	}
}

func contactContactableInboxesCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "contactable-inboxes <id>",
		Short:   "List the inboxes a contact can be reached through",
		Example: "  cwctl contacts contactable-inboxes 12",
		Args:    cobra.ExactArgs(1),
		RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Contacts().Action(cmd.Context(), http.MethodGet, url.PathEscape(args[0])+"/contactable_inboxes", nil, nil, &out)
			return out, err
		}),
	}
}

func contactCreateContactInboxCmd(d *deps) *cobra.Command {
	var inboxID int
	var sourceID string
	cmd := &cobra.Command{
		Use:     "create-contact-inbox <id>",
		Short:   "Attach a contact to an inbox (creates a contact-inbox link)",
		Example: "  cwctl contacts create-contact-inbox 12 --inbox-id 3",
		Args:    cobra.ExactArgs(1),
		RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			body := map[string]any{"inbox_id": inboxID}
			if sourceID != "" {
				body["source_id"] = sourceID
			}
			var out json.RawMessage
			err := c.Contacts().Action(cmd.Context(), http.MethodPost, url.PathEscape(args[0])+"/contact_inboxes", nil, body, &out)
			return out, err
		}),
	}
	cmd.Flags().IntVar(&inboxID, "inbox-id", 0, "inbox id (API-channel inboxes)")
	cmd.Flags().StringVar(&sourceID, "source-id", "", "contact-inbox source id (autogenerated when omitted)")
	_ = cmd.MarkFlagRequired("inbox-id")
	return cmd
}

func contactLabelsCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "labels <id>",
		Short:   "List a contact's labels",
		Example: "  cwctl contacts labels 12",
		Args:    cobra.ExactArgs(1),
		RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Contacts().Action(cmd.Context(), http.MethodGet, url.PathEscape(args[0])+"/labels", nil, nil, &out)
			return out, err
		}),
	}
}

func contactAddLabelsCmd(d *deps) *cobra.Command {
	var labels []string
	cmd := &cobra.Command{
		Use:     "add-labels <id>",
		Short:   "Add labels to a contact (replaces the label set)",
		Example: `  cwctl contacts add-labels 12 --labels vip,billing`,
		Args:    cobra.ExactArgs(1),
		RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Contacts().Action(cmd.Context(), http.MethodPost, url.PathEscape(args[0])+"/labels", nil,
				map[string]any{"labels": labels}, &out)
			return out, err
		}),
	}
	cmd.Flags().StringSliceVar(&labels, "labels", nil, "labels to set (comma-separated)")
	_ = cmd.MarkFlagRequired("labels")
	return cmd
}
