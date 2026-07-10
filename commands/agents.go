package commands

import "github.com/jjuanrivvera/cwctl/internal/api"

func init() {
	registerResource("", resourceSpec[api.Rec]{
		Use:     "agents",
		Aliases: []string{"agent"},
		Short:   "Manage the account's agents",
		New:     func(c *api.Client) *api.Resource[api.Rec] { return c.Agents() },
		Columns: []string{"id", "name", "email", "role", "availability_status"},
		NoGet:   true, // the API exposes no single-agent endpoint
		CreateFields: []field{
			{Flag: "name", Usage: "agent's full name", Required: true},
			{Flag: "email", Usage: "agent's email (an invite is sent)", Required: true},
			{Flag: "role", Usage: "agent | administrator", Required: true},
			{Flag: "availability", Usage: "available | busy | offline"},
			{Flag: "auto-offline", Kind: fieldBool, Usage: "mark offline automatically when away"},
		},
		UpdateFields: []field{
			{Flag: "role", Usage: "agent | administrator", Required: true},
			{Flag: "availability", Usage: "available | busy | offline"},
			{Flag: "auto-offline", Kind: fieldBool, Usage: "mark offline automatically when away"},
		},
	})
}
