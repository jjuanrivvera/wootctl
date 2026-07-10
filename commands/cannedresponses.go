package commands

import "github.com/jjuanrivvera/cwctl/internal/api"

func init() {
	registerResource("", resourceSpec[api.Rec]{
		Use:     "canned-responses",
		Aliases: []string{"canned"},
		Short:   "Manage canned responses (saved reply snippets)",
		New:     func(c *api.Client) *api.Resource[api.Rec] { return c.CannedResponses() },
		Columns: []string{"id", "short_code", "content"},
		NoGet:   true, // the API exposes no single-canned-response endpoint
		CreateFields: []field{
			{Flag: "content", Usage: "the reply text", Required: true},
			{Flag: "short-code", Usage: "shortcut typed after / in the reply box", Required: true},
		},
		UpdateFields: []field{
			{Flag: "content", Usage: "the reply text"},
			{Flag: "short-code", Usage: "shortcut typed after / in the reply box"},
		},
	})
}
