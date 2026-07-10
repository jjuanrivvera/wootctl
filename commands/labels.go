package commands

import "github.com/jjuanrivvera/cwctl/internal/api"

func init() {
	registerResource("", resourceSpec[api.Rec]{
		Use:     "labels",
		Aliases: []string{"label"},
		Short:   "Manage labels",
		New:     func(c *api.Client) *api.Resource[api.Rec] { return c.Labels() },
		Columns: []string{"id", "title", "color", "show_on_sidebar"},
		CreateFields: []field{
			{Flag: "title", Usage: "label name", Required: true},
			{Flag: "description", Usage: "label description"},
			{Flag: "color", Usage: "hex color, e.g. #0055ff"},
			{Flag: "show-on-sidebar", Kind: fieldBool, Usage: "pin the label to the sidebar"},
		},
	})
}
