package commands

import "github.com/jjuanrivvera/wootctl/internal/api"

func init() {
	registerResource("", resourceSpec[api.Rec]{
		Use:     "custom-filters",
		Aliases: []string{"filters"},
		Short:   "Manage saved custom filters",
		New:     func(c *api.Client) *api.Resource[api.Rec] { return c.CustomFilters() },
		Columns: []string{"id", "name", "filter_type"},
		ListFilters: []listFilter{
			{Flag: "filter-type", Usage: "conversation | contact | report"},
		},
		CreateFields: []field{
			{Flag: "name", Usage: "filter name", Required: true},
			{Flag: "type", Usage: "conversation | contact | report", Required: true},
			{Flag: "query", Kind: fieldJSON, Usage: `filter query, e.g. '{"payload":[{"attribute_key":"status","filter_operator":"equal_to","values":["open"]}]}'`},
		},
	})
}
