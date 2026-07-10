// Package commands wires the cobra command tree. root.go owns the global flags, the shared
// API client factory, and the single render() path used by every resource command. The tree
// is built fresh per NewRootCmd() call so tests never leak flag state across cases.
package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/cwctl/internal/api"
	"github.com/jjuanrivvera/cwctl/internal/auth"
	"github.com/jjuanrivvera/cwctl/internal/config"
	"github.com/jjuanrivvera/cwctl/internal/output"
)

// globalFlags holds the persistent flag values for one command tree.
type globalFlags struct {
	outputFormat string
	profile      string
	accountID    string
	baseURL      string
	dryRun       bool
	showToken    bool
	verbose      bool
	noColor      bool
	columns      []string
	quiet        bool
	jq           string
	rps          float64

	// list flags (registered globally so the generic builder can read them)
	all    bool
	page   int
	limit  int
	sort   string
	filter []string
}

// deps carries the per-tree state into every command builder.
type deps struct {
	gf *globalFlags

	// overridable in tests
	loadConfig func() (*config.Config, error)
	store      func() auth.Store
	out        *os.File
}

func newDeps() *deps {
	return &deps{
		gf:         &globalFlags{},
		loadConfig: config.Load,
		store: func() auth.Store {
			dir, err := config.Dir()
			if err != nil {
				dir = "."
			}
			return auth.New(dir)
		},
	}
}

// registrar builds one top-level resource command for a given group ("" = application,
// "platform", "client"). Resource files append to registrars from init().
type registrar struct {
	group string
	build func(d *deps) *cobra.Command
}

var registrars []registrar

// NewRootCmd assembles the full command tree. main.go calls
// NewRootCmd().ExecuteContext(ctx) with a signal.NotifyContext so Ctrl-C cancels
// in-flight work.
func NewRootCmd() *cobra.Command { return newRootCmd(newDeps()) }

// newRootCmd is the deps-injected assembly used by tests (fake store, temp config).
func newRootCmd(d *deps) *cobra.Command {
	root := &cobra.Command{
		Use:   "cwctl",
		Short: "A fast, scriptable CLI for the full Chatwoot API",
		Long: `cwctl drives Chatwoot from the terminal: conversations, messages, contacts,
agents, teams, inboxes, reports, the platform API, and the public client API — 144/144
documented operations, with named profiles for working across several instances.

Examples:
  cwctl auth login
  cwctl conversations list --status open
  cwctl messages create 123 --content "On it!"
  cwctl contacts search --q ana -o json
  cwctl reports summary --since 2026-06-01 --until 2026-07-01
  cwctl --profile staging conversations list --all`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if d.gf.outputFormat != "" && !output.Format(d.gf.outputFormat).Valid() {
				return fmt.Errorf("unknown output format %q (want table|json|yaml|csv|id)", d.gf.outputFormat)
			}
			return nil
		},
	}
	registerGlobalFlags(root, d.gf)

	// Nested API-group parents (DECISIONS.md #15): platform (platform app token) and
	// client (public API, unauthenticated).
	platform := &cobra.Command{Use: "platform", Short: "Platform API (accounts, users, agent bots) — needs a platform app token"}
	client := &cobra.Command{Use: "client", Aliases: []string{"public"}, Short: "Public client API (inbox/contact/conversation flows) — no token required"}

	names := map[string]*cobra.Command{"": root, "platform": platform, "client": client}
	for _, r := range registrars {
		parent, ok := names[r.group]
		if !ok {
			panic("unknown resource group " + r.group)
		}
		parent.AddCommand(r.build(d))
	}
	root.AddCommand(platform, client)

	for _, b := range metaRegistrars {
		root.AddCommand(b(d))
	}
	return root
}

// metaRegistrars register the non-resource commands (auth, config, doctor, …).
var metaRegistrars []func(d *deps) *cobra.Command

func registerGlobalFlags(root *cobra.Command, gf *globalFlags) {
	pf := root.PersistentFlags()
	pf.StringVarP(&gf.outputFormat, "output", "o", "", "output format: table|json|yaml|csv|id")
	pf.StringVar(&gf.profile, "profile", "", "named profile to use (instance + account + token)")
	pf.StringVar(&gf.accountID, "account-id", "", "override the profile's account id for this invocation")
	pf.StringVar(&gf.baseURL, "base-url", "", "override the instance base URL")
	pf.BoolVar(&gf.dryRun, "dry-run", false, "print the equivalent curl and make no request")
	pf.BoolVar(&gf.showToken, "show-token", false, "reveal the API token in dry-run output")
	pf.BoolVarP(&gf.verbose, "verbose", "v", false, "verbose request logging (stderr)")
	pf.BoolVar(&gf.noColor, "no-color", false, "disable colored output")
	pf.StringSliceVar(&gf.columns, "columns", nil, "comma-separated columns to show")
	pf.BoolVar(&gf.quiet, "quiet", false, "suppress non-essential chatter")
	pf.StringVar(&gf.jq, "jq", "", "gojq expression applied to the response before rendering")
	pf.Float64Var(&gf.rps, "rps", 0, "max requests per second (default 5; also per-profile `rps` in config)")

	// List flags (read by the generic builder's list command).
	pf.BoolVar(&gf.all, "all", false, "fetch all pages (list commands)")
	pf.IntVar(&gf.page, "page", 0, "page number to fetch (list commands; Chatwoot pages are server-sized)")
	pf.IntVar(&gf.limit, "limit", 0, "max items to output, applied client-side (list commands)")
	pf.StringVar(&gf.sort, "sort", "", "sort field, prefix with - for descending (where the API supports it)")
	pf.StringSliceVar(&gf.filter, "filter", nil, "client-side field=value filters (list commands)")
}

// getAPIClient builds an authenticated client for the ACTIVE profile, honoring
// flag > env > config precedence. requireAuth=false lets public/client and dry-run-ish
// commands proceed tokenless.
func (d *deps) getAPIClient(requireAuth bool) (*api.Client, *config.Config, error) {
	cfg, err := d.loadConfig()
	if err != nil {
		return nil, nil, err
	}
	profileName := cfg.ResolveProfileName(d.gf.profile)
	c, err := d.clientForProfile(cfg, profileName, requireAuth, true)
	return c, cfg, err
}

// clientForProfile builds a client for an explicitly named profile. allowGlobals wires the
// global --base-url/--account-id flags and the CWCTL_* env overrides; it MUST be false for a
// secondary profile (e.g. `sync --to <profile>`), whose base/account/token come only from its
// own stored config + keyring so the active profile's flags don't leak across instances.
func (d *deps) clientForProfile(cfg *config.Config, profileName string, requireAuth, allowGlobals bool) (*api.Client, error) {
	prof, _ := cfg.Profile(profileName)

	baseURL := prof.BaseURL
	if allowGlobals {
		baseURL = config.FirstNonEmpty(d.gf.baseURL, os.Getenv("CWCTL_BASE_URL"), prof.BaseURL)
	}
	if baseURL == "" {
		return nil, fmt.Errorf("no base URL for profile %q — run `cwctl --profile %s auth login`", profileName, profileName)
	}

	var token string
	if allowGlobals {
		token = os.Getenv("CWCTL_API_KEY")
	}
	if token == "" {
		if t, err := d.store().Get(profileName); err == nil {
			token = t
		}
	}
	if requireAuth && token == "" {
		return nil, fmt.Errorf("no API token for profile %q — run `cwctl --profile %s auth login`", profileName, profileName)
	}

	var platformToken string
	if allowGlobals {
		platformToken = os.Getenv("CWCTL_PLATFORM_TOKEN")
	}
	if platformToken == "" {
		if t, err := d.store().Get(auth.PlatformKey(profileName)); err == nil {
			platformToken = t
		}
	}

	rps := d.gf.rps
	if rps == 0 {
		rps = prof.Rps
	}
	if rps == 0 {
		rps = 5
	}

	c := api.New(baseURL, token,
		api.WithDryRun(d.gf.dryRun, d.stdout()),
		api.WithPlatformToken(platformToken),
		api.WithRateLimit(rps),
	)
	c.AccountID = prof.AccountID
	if allowGlobals {
		c.AccountID = config.FirstNonEmpty(d.gf.accountID, os.Getenv("CWCTL_ACCOUNT_ID"), prof.AccountID)
	}
	c.ShowToken = d.gf.showToken
	c.Verbose = d.gf.verbose
	c.VerboseOut = os.Stderr
	return c, nil
}

func (d *deps) stdout() *os.File {
	if d.out != nil {
		return d.out
	}
	return os.Stdout
}

// render is the single output path for every command: normalize v to JSON, then hand it to
// the shared renderer with the resolved global flags.
func (d *deps) render(cmd *cobra.Command, v any, defaultColumns []string) error {
	raw, ok := v.(json.RawMessage)
	if !ok {
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		raw = b
	}
	cols := normalizeColumns(d.gf.columns)
	if len(cols) == 0 {
		cols = defaultColumns
	}
	return output.Render(raw, output.Options{
		Format:  output.Format(config.FirstNonEmpty(d.gf.outputFormat, string(output.FormatTable))),
		Columns: cols,
		NoColor: d.gf.noColor,
		Quiet:   d.gf.quiet,
		JQ:      d.gf.jq,
		Out:     cmd.OutOrStdout(),
		Err:     cmd.ErrOrStderr(),
	})
}

func normalizeColumns(cols []string) []string {
	var out []string
	for _, c := range cols {
		if c = strings.TrimSpace(c); c != "" {
			out = append(out, c)
		}
	}
	return out
}
