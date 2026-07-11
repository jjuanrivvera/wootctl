package commands

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/wootctl/internal/api"
)

// messages are conversation-scoped (…/conversations/{id}/messages), so every verb takes
// the conversation id positionally and the generic collection builder doesn't apply.
func init() {
	registerResource("", resourceSpec[api.Rec]{
		Use:     "messages",
		Aliases: []string{"msg", "msgs"},
		Short:   "Read and send messages in a conversation",
		New: func(c *api.Client) *api.Resource[api.Rec] {
			return api.NewResource[api.Rec](c, c.AccountPath("conversations"))
		},
		Columns: []string{"id", "content", "message_type", "created_at"},
		NoList:  true, NoGet: true, NoCreate: true, NoUpdate: true, NoDelete: true,
		Extra: []extraCommand{
			{Kind: kindRead, Build: messagesListCmd},
			{Kind: kindWrite, Build: messagesCreateCmd},
			{Kind: kindDestructive, Build: messagesDeleteCmd},
		},
	})
}

func messagesPath(c *api.Client, conversationID string) string {
	return c.AccountPath("conversations/" + url.PathEscape(conversationID) + "/messages")
}

func messagesListCmd(d *deps) *cobra.Command {
	var before, after string
	cmd := &cobra.Command{
		Use:   "list <conversation-id>",
		Short: "List messages in a conversation",
		Example: `  wootctl messages list 42
  wootctl messages list 42 --before 105023 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: runListE(d, false, []string{"id", "content", "message_type", "created_at"}, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			query := url.Values{}
			if before != "" {
				query.Set("before", before)
			}
			if after != "" {
				query.Set("after", after)
			}
			var out json.RawMessage
			err := c.Send(cmd.Context(), http.MethodGet, messagesPath(c, args[0]), query, nil, &out)
			return out, err
		}),
	}
	cmd.Flags().StringVar(&before, "before", "", "return messages before this message id")
	cmd.Flags().StringVar(&after, "after", "", "return messages after this message id")
	return cmd
}

func messagesCreateCmd(d *deps) *cobra.Command {
	var attachments []string
	cmd := &cobra.Command{
		Use:   "create <conversation-id>",
		Short: "Send a message (reply) in a conversation",
		Long: `Send an outgoing message, a private note (--private), or an incoming message
(--message-type incoming, API channels). Attach files with repeatable --attachment
(switches to a multipart upload).`,
		Example: `  wootctl messages create 42 --content "On it — checking now."
  wootctl messages create 42 --content "internal note" --private
  wootctl messages create 42 --content "see attached" --attachment ./invoice.pdf`,
		Args: cobra.ExactArgs(1),
	}
	collect := registerBodyFlags(cmd, []field{
		{Flag: "content", Usage: "message text", Required: true},
		{Flag: "message-type", Usage: "outgoing | incoming (default outgoing)"},
		{Flag: "private", Kind: fieldBool, Usage: "private note (not visible to the contact)"},
		{Flag: "content-type", Usage: "text | input_email | cards | input_select | form | article"},
		{Flag: "content-attributes", Kind: fieldJSON, Usage: "content attributes object (interactive messages)"},
		{Flag: "template-params", Kind: fieldJSON, Usage: "WhatsApp template params object"},
		{Flag: "campaign-id", Kind: fieldInt, Usage: "campaign id"},
	})
	cmd.Flags().StringSliceVar(&attachments, "attachment", nil, "file to attach (repeatable; forces multipart)")
	cmd.RunE = runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
		body, err := collect(cmd)
		if err != nil {
			return nil, err
		}
		var out json.RawMessage
		if len(attachments) > 0 {
			// Attachments require multipart/form-data; scalar fields ride as form values.
			fields := url.Values{}
			for k, v := range body {
				switch t := v.(type) {
				case string:
					fields.Set(k, t)
				case bool:
					fields.Set(k, boolStr(t))
				case float64, int:
					fields.Set(k, stringifyField(v))
				default: // nested objects (content_attributes…) as JSON strings
					b, _ := json.Marshal(v)
					fields.Set(k, string(b))
				}
			}
			files := make([]api.UploadFile, 0, len(attachments))
			for _, p := range attachments {
				files = append(files, api.UploadFile{Field: "attachments[]", Path: p})
			}
			err = c.DoMultipart(cmd.Context(), http.MethodPost, messagesPath(c, args[0]), fields, files, &out)
			return out, err
		}
		err = c.Send(cmd.Context(), http.MethodPost, messagesPath(c, args[0]), nil, body, &out)
		return out, err
	})
	return cmd
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func messagesDeleteCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "delete <conversation-id> <message-id>",
		Short:   "Delete a message from a conversation",
		Example: "  wootctl messages delete 42 105023",
		Args:    cobra.ExactArgs(2),
		RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			err := c.Send(cmd.Context(), http.MethodDelete, messagesPath(c, args[0])+"/"+url.PathEscape(args[1]), nil, nil, nil)
			if err != nil {
				return nil, err
			}
			if !d.gf.quiet && !c.DryRun {
				cmd.Printf("deleted message %s from conversation %s\n", args[1], args[0])
			}
			return nil, nil
		}),
	}
}
