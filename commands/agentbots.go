package commands

import "github.com/jjuanrivvera/wootctl/internal/api"

func init() {
	registerResource("", resourceSpec[api.Rec]{
		Use:     "agent-bots",
		Aliases: []string{"bots"},
		Short:   "Manage account agent bots",
		New:     func(c *api.Client) *api.Resource[api.Rec] { return c.AgentBots() },
		Columns: []string{"id", "name", "description"},
		CreateFields: []field{
			{Flag: "name", Usage: "bot name", Required: true},
			{Flag: "description", Usage: "bot description"},
			{Flag: "outgoing-url", Usage: "webhook URL the bot receives events on"},
			{Flag: "bot-type", Usage: "bot type (e.g. webhook)"},
			{Flag: "bot-config", Kind: fieldJSON, Usage: "bot configuration object"},
		},
	})
}
