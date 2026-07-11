package commands

import "github.com/jjuanrivvera/wootctl/internal/api"

func init() {
	registerResource("", resourceSpec[api.Rec]{
		Use:     "webhooks",
		Aliases: []string{"webhook"},
		Short:   "Manage account webhook subscriptions",
		New:     func(c *api.Client) *api.Resource[api.Rec] { return c.Webhooks() },
		Columns: []string{"id", "url", "subscriptions"},
		NoGet:   true, // the API exposes no single-webhook endpoint
		CreateFields: []field{
			{Flag: "url", Usage: "endpoint to POST events to", Required: true},
			{Flag: "subscriptions", Kind: fieldStringSlice, Usage: "events: conversation_created,conversation_status_changed,message_created,…"},
			{Flag: "name", Usage: "webhook name"},
		},
		UpdateFields: []field{
			{Flag: "url", Usage: "endpoint to POST events to"},
			{Flag: "subscriptions", Kind: fieldStringSlice, Usage: "events to subscribe to"},
			{Flag: "name", Usage: "webhook name"},
		},
	})
}
