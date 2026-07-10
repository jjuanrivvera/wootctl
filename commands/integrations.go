package commands

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/cwctl/internal/api"
)

// integrations pair a read-only app catalog (GET integrations/apps) with hook CRUD under
// integrations/hooks — neither is collection-CRUD, so all verbs are Extra.
func init() {
	registerResource("", resourceSpec[api.Rec]{
		Use:   "integrations",
		Short: "List integration apps and manage integration hooks",
		New: func(c *api.Client) *api.Resource[api.Rec] {
			return api.NewResource[api.Rec](c, c.AccountPath("integrations"))
		},
		Columns: []string{"id", "name", "enabled"},
		NoList:  true, NoGet: true, NoCreate: true, NoUpdate: true, NoDelete: true,
		Extra: []extraCommand{
			{Kind: kindRead, Build: integrationsAppsCmd},
			{Kind: kindWrite, Build: integrationsCreateHookCmd},
			{Kind: kindWrite, Build: integrationsUpdateHookCmd},
			{Kind: kindDestructive, Build: integrationsDeleteHookCmd},
		},
	})
}

func integrationsAppsCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "apps",
		Short:   "List available integration apps and their status",
		Example: "  cwctl integrations apps -o json",
		Args:    cobra.NoArgs,
		RunE: runE(d, false, []string{"id", "name", "enabled"}, func(cmd *cobra.Command, c *api.Client, _ []string) (json.RawMessage, error) {
			var out json.RawMessage
			err := c.Send(cmd.Context(), http.MethodGet, c.AccountPath("integrations/apps"), nil, nil, &out)
			return out, err
		}),
	}
}

func integrationsCreateHookCmd(d *deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-hook",
		Short:   "Create an integration hook",
		Example: `  cwctl integrations create-hook --app-id dialogflow --settings '{"project_id":"x","credentials":{}}' --inbox-id 3`,
		Args:    cobra.NoArgs,
	}
	collect := registerBodyFlags(cmd, []field{
		{Flag: "app-id", Usage: "integration app id (e.g. slack, dialogflow, dyte)", Required: true},
		{Flag: "inbox-id", Kind: fieldInt, Usage: "inbox to attach the hook to (inbox-scoped apps)"},
		{Flag: "settings", Kind: fieldJSON, Usage: "app-specific settings object"},
	})
	cmd.RunE = runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, _ []string) (json.RawMessage, error) {
		body, err := collect(cmd)
		if err != nil {
			return nil, err
		}
		var out json.RawMessage
		err = c.Send(cmd.Context(), http.MethodPost, c.AccountPath("integrations/hooks"), nil, body, &out)
		return out, err
	})
	return cmd
}

func integrationsUpdateHookCmd(d *deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-hook <hook-id>",
		Short:   "Update an integration hook's settings",
		Example: `  cwctl integrations update-hook 5 --settings '{"project_id":"y"}'`,
		Args:    cobra.ExactArgs(1),
	}
	collect := registerBodyFlags(cmd, []field{
		{Flag: "settings", Kind: fieldJSON, Usage: "app-specific settings object"},
		{Flag: "status", Usage: "enabled | disabled"},
	})
	cmd.RunE = runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
		body, err := collect(cmd)
		if err != nil {
			return nil, err
		}
		var out json.RawMessage
		err = c.Send(cmd.Context(), http.MethodPatch, c.AccountPath("integrations/hooks/"+url.PathEscape(args[0])), nil, body, &out)
		return out, err
	})
	return cmd
}

func integrationsDeleteHookCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "delete-hook <hook-id>",
		Short:   "Delete an integration hook",
		Example: "  cwctl integrations delete-hook 5",
		Args:    cobra.ExactArgs(1),
		RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
			if err := c.Send(cmd.Context(), http.MethodDelete, c.AccountPath("integrations/hooks/"+url.PathEscape(args[0])), nil, nil, nil); err != nil {
				return nil, err
			}
			if !d.gf.quiet && !c.DryRun {
				cmd.Printf("deleted integration hook %s\n", args[0])
			}
			return nil, nil
		}),
	}
}
