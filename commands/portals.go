package commands

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/cwctl/internal/api"
)

func init() {
	registerResource("", resourceSpec[api.Rec]{
		Use:     "portals",
		Aliases: []string{"portal", "help-center"},
		Short:   "Manage help-center portals, articles, and categories",
		New:     func(c *api.Client) *api.Resource[api.Rec] { return c.Portals() },
		Columns: []string{"id", "name", "slug", "archived"},
		// The spec exposes list/create/update only (no single-get, no delete).
		NoGet: true, NoDelete: true,
		CreateFields: []field{
			{Flag: "name", Usage: "portal name", Required: true},
			{Flag: "slug", Usage: "portal slug (URL segment)"},
			{Flag: "custom-domain", Usage: "custom domain for the portal"},
			{Flag: "color", Usage: "accent color, e.g. #0055ff"},
			{Flag: "page-title", Usage: "HTML page title"},
			{Flag: "header-text", Usage: "header text"},
			{Flag: "homepage-link", Usage: "link back to your site"},
			{Flag: "archived", Kind: fieldBool, Usage: "archive the portal"},
			{Flag: "config", Kind: fieldJSON, Usage: `portal config, e.g. '{"allowed_locales":["en","es"]}'`},
		},
		Extra: []extraCommand{
			{Kind: kindWrite, Build: portalCreateArticleCmd},
			{Kind: kindWrite, Build: portalCreateCategoryCmd},
		},
	})
}

func portalCreateArticleCmd(d *deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-article <portal-id>",
		Short:   "Add an article to a portal",
		Example: `  cwctl portals create-article 1 --title "Cómo empezar" --content "..." --author-id 1 --category-id 2`,
		Args:    cobra.ExactArgs(1),
	}
	collect := registerBodyFlags(cmd, []field{
		{Flag: "title", Usage: "article title", Required: true},
		{Flag: "content", Usage: "article body (markdown)", Required: true},
		{Flag: "author-id", Kind: fieldInt, Usage: "author user id"},
		{Flag: "category-id", Kind: fieldInt, Usage: "category id"},
		{Flag: "description", Usage: "meta description"},
		{Flag: "slug", Usage: "article slug"},
		{Flag: "status", Usage: "draft | published | archived"},
		{Flag: "locale", Usage: "article locale, e.g. en, es"},
		{Flag: "position", Kind: fieldInt, Usage: "sort position"},
		{Flag: "meta", Kind: fieldJSON, Usage: "SEO meta object"},
	})
	cmd.RunE = runE(d, false, []string{"id", "title", "status"}, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
		body, err := collect(cmd)
		if err != nil {
			return nil, err
		}
		var out json.RawMessage
		err = c.Portals().Action(cmd.Context(), http.MethodPost, url.PathEscape(args[0])+"/articles", nil, body, &out)
		return out, err
	})
	return cmd
}

func portalCreateCategoryCmd(d *deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-category <portal-id>",
		Short:   "Add a category to a portal",
		Example: `  cwctl portals create-category 1 --name "Facturación" --locale es`,
		Args:    cobra.ExactArgs(1),
	}
	collect := registerBodyFlags(cmd, []field{
		{Flag: "name", Usage: "category name", Required: true},
		{Flag: "description", Usage: "category description"},
		{Flag: "slug", Usage: "category slug"},
		{Flag: "locale", Usage: "category locale, e.g. en, es"},
		{Flag: "position", Kind: fieldInt, Usage: "sort position"},
		{Flag: "icon", Usage: "category icon"},
		{Flag: "parent-category-id", Kind: fieldInt, Usage: "parent category id"},
		{Flag: "associated-category-id", Kind: fieldInt, Usage: "linked category id (another locale)"},
	})
	cmd.RunE = runE(d, false, []string{"id", "name", "slug"}, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
		body, err := collect(cmd)
		if err != nil {
			return nil, err
		}
		var out json.RawMessage
		err = c.Portals().Action(cmd.Context(), http.MethodPost, url.PathEscape(args[0])+"/categories", nil, body, &out)
		return out, err
	})
	return cmd
}
