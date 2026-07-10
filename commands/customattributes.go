package commands

import "github.com/jjuanrivvera/cwctl/internal/api"

func init() {
	registerResource("", resourceSpec[api.Rec]{
		Use:     "custom-attributes",
		Aliases: []string{"attributes"},
		Short:   "Manage custom attribute definitions",
		New:     func(c *api.Client) *api.Resource[api.Rec] { return c.CustomAttributes() },
		Columns: []string{"id", "attribute_display_name", "attribute_key", "attribute_model", "attribute_display_type"},
		ListFilters: []listFilter{
			{Flag: "attribute-model", Usage: "0 = conversation attributes, 1 = contact attributes"},
		},
		CreateFields: []field{
			{Flag: "attribute-display-name", Usage: "human name", Required: true},
			{Flag: "attribute-key", Usage: "machine key", Required: true},
			{Flag: "attribute-model", Kind: fieldInt, Usage: "0 = conversation, 1 = contact", Required: true},
			{Flag: "attribute-display-type", Kind: fieldInt, Usage: "0 text, 1 number, 2 currency, 3 percent, 4 link, 5 date, 6 list, 7 checkbox"},
			{Flag: "attribute-description", Usage: "description"},
			{Flag: "attribute-values", Kind: fieldStringSlice, Usage: "allowed values (list type)"},
			{Flag: "regex-pattern", Usage: "validation regex (text type)"},
			{Flag: "regex-cue", Usage: "hint shown when the regex rejects a value"},
		},
	})
}
