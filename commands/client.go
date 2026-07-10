package commands

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/cwctl/internal/api"
)

// The public client API drives contact-facing chat flows (build your own widget/importer).
// It is UNAUTHENTICATED by design: routes are scoped by inbox identifier + contact source
// id (DECISIONS.md #6). Every command here passes NoAuth so no token is required.

func init() {
	registerResource("client", resourceSpec[api.Rec]{
		Use:   "inbox",
		Short: "Public inbox details",
		New: func(c *api.Client) *api.Resource[api.Rec] {
			return api.NewResource[api.Rec](c, "public/api/v1/inboxes")
		},
		Columns: []string{"identifier", "name"},
		NoAuth:  true,
		NoList:  true, NoCreate: true, NoUpdate: true, NoDelete: true,
		GetByArg: "inbox-identifier",
	})

	registerResource("client", resourceSpec[api.Rec]{
		Use:   "contacts",
		Short: "Public (contact-facing) contact endpoints",
		New: func(c *api.Client) *api.Resource[api.Rec] {
			return api.NewResource[api.Rec](c, "public/api/v1/inboxes")
		},
		Columns: []string{"id", "source_id", "name", "email"},
		NoAuth:  true,
		NoList:  true, NoGet: true, NoCreate: true, NoUpdate: true, NoDelete: true,
		Extra: []extraCommand{
			{Kind: kindWrite, Build: clientContactCreateCmd},
			{Kind: kindRead, Build: clientContactGetCmd},
			{Kind: kindWrite, Build: clientContactUpdateCmd},
		},
	})

	registerResource("client", resourceSpec[api.Rec]{
		Use:   "conversations",
		Short: "Public (contact-facing) conversation endpoints",
		New: func(c *api.Client) *api.Resource[api.Rec] {
			return api.NewResource[api.Rec](c, "public/api/v1/inboxes")
		},
		Columns: []string{"id", "inbox_id", "status"},
		NoAuth:  true,
		NoList:  true, NoGet: true, NoCreate: true, NoUpdate: true, NoDelete: true,
		Extra: []extraCommand{
			{Kind: kindRead, Build: clientConvListCmd},
			{Kind: kindWrite, Build: clientConvCreateCmd},
			{Kind: kindRead, Build: clientConvGetCmd},
			{Kind: kindWrite, Build: clientConvResolveCmd},
			{Kind: kindWrite, Build: clientConvToggleTypingCmd},
			{Kind: kindWrite, Build: clientConvUpdateLastSeenCmd},
		},
	})

	registerResource("client", resourceSpec[api.Rec]{
		Use:   "messages",
		Short: "Public (contact-facing) message endpoints",
		New: func(c *api.Client) *api.Resource[api.Rec] {
			return api.NewResource[api.Rec](c, "public/api/v1/inboxes")
		},
		Columns: []string{"id", "content", "message_type"},
		NoAuth:  true,
		NoList:  true, NoGet: true, NoCreate: true, NoUpdate: true, NoDelete: true,
		Extra: []extraCommand{
			{Kind: kindRead, Build: clientMsgListCmd},
			{Kind: kindWrite, Build: clientMsgCreateCmd},
			{Kind: kindWrite, Build: clientMsgUpdateCmd},
		},
	})
}

func publicInboxPath(inbox string) string {
	return "public/api/v1/inboxes/" + url.PathEscape(inbox)
}

func publicContactPath(inbox, contact string) string {
	return publicInboxPath(inbox) + "/contacts/" + url.PathEscape(contact)
}

func publicConvPath(inbox, contact, conv string) string {
	return publicContactPath(inbox, contact) + "/conversations/" + url.PathEscape(conv)
}

// clientContactCreateCmd: POST /public/api/v1/inboxes/{inbox}/contacts
func clientContactCreateCmd(d *deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <inbox-identifier>",
		Short:   "Create a contact in a public inbox",
		Example: `  cwctl client contacts create Fbd1h… --name Ana --email ana@example.com`,
		Args:    cobra.ExactArgs(1),
	}
	collect := registerBodyFlags(cmd, []field{
		{Flag: "identifier", Usage: "external unique identifier"},
		{Flag: "identifier-hash", Usage: "HMAC of identifier (identity validation)"},
		{Flag: "name", Usage: "contact name"},
		{Flag: "email", Usage: "email address"},
		{Flag: "phone-number", Usage: "phone in E.164"},
		{Flag: "custom-attributes", Kind: fieldJSON, Usage: "custom attributes object"},
	})
	cmd.RunE = runE(d, true, []string{"id", "source_id", "name", "email"}, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
		body, err := collect(cmd)
		if err != nil {
			return nil, err
		}
		var out json.RawMessage
		err = c.Send(cmd.Context(), http.MethodPost, publicInboxPath(args[0])+"/contacts", nil, body, &out)
		return out, err
	})
	return cmd
}

func clientContactGetCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "get <inbox-identifier> <contact-identifier>",
		Short:   "Get a public contact (by its source id)",
		Example: "  cwctl client contacts get Fbd1h… c7f3…",
		Args:    cobra.ExactArgs(2),
		RunE: runE(d, true, []string{"id", "source_id", "name", "email"}, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Send(cmd.Context(), http.MethodGet, publicContactPath(args[0], args[1]), nil, nil, &out)
			return out, err
		}),
	}
}

func clientContactUpdateCmd(d *deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update <inbox-identifier> <contact-identifier>",
		Short:   "Update a public contact",
		Example: `  cwctl client contacts update Fbd1h… c7f3… --name "Ana María"`,
		Args:    cobra.ExactArgs(2),
	}
	collect := registerBodyFlags(cmd, []field{
		{Flag: "name", Usage: "contact name"},
		{Flag: "email", Usage: "email address"},
		{Flag: "phone-number", Usage: "phone in E.164"},
		{Flag: "custom-attributes", Kind: fieldJSON, Usage: "custom attributes object"},
	})
	cmd.RunE = runE(d, true, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
		body, err := collect(cmd)
		if err != nil {
			return nil, err
		}
		var out json.RawMessage
		err = c.Send(cmd.Context(), http.MethodPatch, publicContactPath(args[0], args[1]), nil, body, &out)
		return out, err
	})
	return cmd
}

func clientConvListCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "list <inbox-identifier> <contact-identifier>",
		Short:   "List a contact's conversations in a public inbox",
		Example: "  cwctl client conversations list Fbd1h… c7f3…",
		Args:    cobra.ExactArgs(2),
		RunE: runE(d, true, []string{"id", "inbox_id", "status"}, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Send(cmd.Context(), http.MethodGet, publicContactPath(args[0], args[1])+"/conversations", nil, nil, &out)
			return out, err
		}),
	}
}

func clientConvCreateCmd(d *deps) *cobra.Command {
	var attrs string
	cmd := &cobra.Command{
		Use:     "create <inbox-identifier> <contact-identifier>",
		Short:   "Start a conversation as a contact",
		Example: "  cwctl client conversations create Fbd1h… c7f3…",
		Args:    cobra.ExactArgs(2),
		RunE: runE(d, true, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var body map[string]any
			if attrs != "" {
				var obj map[string]any
				if err := json.Unmarshal([]byte(attrs), &obj); err != nil {
					return nil, err
				}
				body = map[string]any{"custom_attributes": obj}
			}
			var out json.RawMessage
			err := c.Send(cmd.Context(), http.MethodPost, publicContactPath(args[0], args[1])+"/conversations", nil, body, &out)
			return out, err
		}),
	}
	cmd.Flags().StringVar(&attrs, "custom-attributes", "", "custom attributes JSON object")
	return cmd
}

func clientConvGetCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "get <inbox-identifier> <contact-identifier> <conversation-id>",
		Short:   "Get one of the contact's conversations",
		Example: "  cwctl client conversations get Fbd1h… c7f3… 42",
		Args:    cobra.ExactArgs(3),
		RunE: runE(d, true, []string{"id", "inbox_id", "status"}, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Send(cmd.Context(), http.MethodGet, publicConvPath(args[0], args[1], args[2]), nil, nil, &out)
			return out, err
		}),
	}
}

func clientConvResolveCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "resolve <inbox-identifier> <contact-identifier> <conversation-id>",
		Short:   "Resolve a conversation as the contact (toggle_status)",
		Example: "  cwctl client conversations resolve Fbd1h… c7f3… 42",
		Args:    cobra.ExactArgs(3),
		RunE: runE(d, true, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Send(cmd.Context(), http.MethodPost, publicConvPath(args[0], args[1], args[2])+"/toggle_status", nil, nil, &out)
			return out, err
		}),
	}
}

func clientConvToggleTypingCmd(d *deps) *cobra.Command {
	var typing string
	cmd := &cobra.Command{
		Use:     "toggle-typing <inbox-identifier> <contact-identifier> <conversation-id>",
		Short:   "Flip the contact-side typing indicator",
		Example: "  cwctl client conversations toggle-typing Fbd1h… c7f3… 42 --typing-status on",
		Args:    cobra.ExactArgs(3),
		RunE: runE(d, true, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Send(cmd.Context(), http.MethodPost, publicConvPath(args[0], args[1], args[2])+"/toggle_typing", nil,
				map[string]any{"typing_status": typing}, &out)
			return out, err
		}),
	}
	cmd.Flags().StringVar(&typing, "typing-status", "", "on | off")
	_ = cmd.MarkFlagRequired("typing-status")
	return cmd
}

func clientConvUpdateLastSeenCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "update-last-seen <inbox-identifier> <contact-identifier> <conversation-id>",
		Short:   "Mark the conversation read up to now (contact side)",
		Example: "  cwctl client conversations update-last-seen Fbd1h… c7f3… 42",
		Args:    cobra.ExactArgs(3),
		RunE: runE(d, true, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Send(cmd.Context(), http.MethodPost, publicConvPath(args[0], args[1], args[2])+"/update_last_seen", nil, nil, &out)
			return out, err
		}),
	}
}

func clientMsgListCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "list <inbox-identifier> <contact-identifier> <conversation-id>",
		Short:   "List messages in a public conversation",
		Example: "  cwctl client messages list Fbd1h… c7f3… 42",
		Args:    cobra.ExactArgs(3),
		RunE: runE(d, true, []string{"id", "content", "message_type"}, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Send(cmd.Context(), http.MethodGet, publicConvPath(args[0], args[1], args[2])+"/messages", nil, nil, &out)
			return out, err
		}),
	}
}

func clientMsgCreateCmd(d *deps) *cobra.Command {
	var content, echoID string
	cmd := &cobra.Command{
		Use:     "create <inbox-identifier> <contact-identifier> <conversation-id>",
		Short:   "Send a message as the contact",
		Example: `  cwctl client messages create Fbd1h… c7f3… 42 --content "gracias!"`,
		Args:    cobra.ExactArgs(3),
		RunE: runE(d, true, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			body := map[string]any{"content": content}
			if echoID != "" {
				body["echo_id"] = echoID
			}
			var out json.RawMessage
			err := c.Send(cmd.Context(), http.MethodPost, publicConvPath(args[0], args[1], args[2])+"/messages", nil, body, &out)
			return out, err
		}),
	}
	cmd.Flags().StringVar(&content, "content", "", "message text")
	cmd.Flags().StringVar(&echoID, "echo-id", "", "client-side temporary id echoed back")
	_ = cmd.MarkFlagRequired("content")
	return cmd
}

func clientMsgUpdateCmd(d *deps) *cobra.Command {
	var submittedValues string
	cmd := &cobra.Command{
		Use:     "update <inbox-identifier> <contact-identifier> <conversation-id> <message-id>",
		Short:   "Update a message (submit interactive form/select values)",
		Example: `  cwctl client messages update Fbd1h… c7f3… 42 9001 --submitted-values '[{"name":"size","value":"M"}]'`,
		Args:    cobra.ExactArgs(4),
		RunE: runE(d, true, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			var vals any
			if err := json.Unmarshal([]byte(submittedValues), &vals); err != nil {
				return nil, err
			}
			var out json.RawMessage
			err := c.Send(cmd.Context(), http.MethodPatch, publicConvPath(args[0], args[1], args[2])+"/messages/"+url.PathEscape(args[3]), nil,
				map[string]any{"submitted_values": vals}, &out)
			return out, err
		}),
	}
	cmd.Flags().StringVar(&submittedValues, "submitted-values", "", "submitted values JSON array")
	_ = cmd.MarkFlagRequired("submitted-values")
	return cmd
}
