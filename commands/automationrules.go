package commands

import "github.com/jjuanrivvera/cwctl/internal/api"

func init() {
	registerResource("", resourceSpec[api.Rec]{
		Use:     "automation-rules",
		Aliases: []string{"automations"},
		Short:   "Manage automation rules",
		New:     func(c *api.Client) *api.Resource[api.Rec] { return c.AutomationRules() },
		Columns: []string{"id", "name", "event_name", "active"},
		CreateFields: []field{
			{Flag: "name", Usage: "rule name", Required: true},
			{Flag: "event-name", Usage: "trigger: conversation_created | conversation_updated | message_created", Required: true},
			{Flag: "description", Usage: "rule description"},
			{Flag: "active", Kind: fieldBool, Usage: "enable the rule"},
			{Flag: "conditions", Kind: fieldJSON, Usage: `conditions array, e.g. '[{"attribute_key":"status","filter_operator":"equal_to","values":["open"],"query_operator":"AND"}]'`},
			{Flag: "actions", Kind: fieldJSON, Usage: `actions array, e.g. '[{"action_name":"assign_team","action_params":[1]}]'`},
		},
	})
}
