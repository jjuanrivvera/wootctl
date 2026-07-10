package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/cwctl/internal/api"
)

// cmdKind classifies a command for MCP/agent-guard annotations.
type cmdKind int

const (
	kindRead        cmdKind = iota // read-only: MCP readOnlyHint (+idempotentHint)
	kindWrite                      // creates/changes remote state: MCP openWorldHint
	kindDestructive                // irreversible (delete): MCP destructiveHint
)

// MCP tool annotation keys (the singular MCP hint keys; ophis reads these from
// cmd.Annotations). Stamped once here in the generic builder — never per command.
const (
	annReadOnly    = "readOnlyHint"
	annDestructive = "destructiveHint"
	annOpenWorld   = "openWorldHint"
	annIdempotent  = "idempotentHint"
)

// annotate stamps the MCP classification for kind onto cmd.
func annotate(cmd *cobra.Command, kind cmdKind) *cobra.Command {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	switch kind {
	case kindRead:
		cmd.Annotations[annReadOnly] = "true"
		cmd.Annotations[annIdempotent] = "true"
	case kindWrite:
		cmd.Annotations[annOpenWorld] = "true"
	case kindDestructive:
		cmd.Annotations[annOpenWorld] = "true"
		cmd.Annotations[annDestructive] = "true"
	}
	return cmd
}

type fieldKind int

const (
	fieldString fieldKind = iota
	fieldInt
	fieldBool
	fieldFloat
	fieldStringSlice
	fieldJSON // value parsed as JSON and embedded as-is (objects/arrays)
)

// field declares one convenience flag for create/update and the JSON key it maps to.
// Anything not covered by fields still reaches the API via --data (full-payload JSON).
type field struct {
	Flag     string
	Key      string // defaults to Flag with '-' → '_'
	Kind     fieldKind
	Required bool
	Usage    string
}

func (f field) key() string {
	if f.Key != "" {
		return f.Key
	}
	return strings.ReplaceAll(f.Flag, "-", "_")
}

// listFilter maps a CLI flag to a server-side query param on list commands.
type listFilter struct {
	Flag  string
	Query string // defaults to Flag with '-' → '_'
	Usage string
}

func (f listFilter) query() string {
	if f.Query != "" {
		return f.Query
	}
	return strings.ReplaceAll(f.Flag, "-", "_")
}

// extraCommand is a custom verb (toggle-status, search, …) contributed by a resource file.
// The generic builder stamps its MCP annotation from Kind so nothing ships unclassified.
type extraCommand struct {
	Kind  cmdKind
	Build func(d *deps) *cobra.Command
}

// resourceSpec declares a resource's CLI surface. A new resource = a type + a Client
// accessor + one registerResource call in init(). No shared code changes.
type resourceSpec[T any] struct {
	Use     string
	Aliases []string
	Short   string
	New     func(*api.Client) *api.Resource[T]
	Columns []string

	ListFilters []listFilter
	ListShort   string // override the default "List <use>" help line

	CreateFields []field
	UpdateFields []field
	CreateShort  string
	UpdateShort  string

	NoList, NoGet, NoCreate, NoUpdate, NoDelete bool

	// NoAuth marks resources on the public client API: commands run without a stored token.
	NoAuth bool

	// GetByArg overrides the positional arg name in help for get/update/delete (default "id").
	GetByArg string

	Extra []extraCommand
}

// registerResource queues a resource for the given group ("" application, "platform",
// "client"); NewRootCmd applies the queue when assembling a tree.
func registerResource[T any](group string, spec resourceSpec[T]) {
	registrars = append(registrars, registrar{group: group, build: func(d *deps) *cobra.Command {
		return buildResourceCmd(d, spec)
	}})
}

func buildResourceCmd[T any](d *deps, spec resourceSpec[T]) *cobra.Command {
	parent := &cobra.Command{
		Use:     spec.Use,
		Aliases: spec.Aliases,
		Short:   spec.Short,
	}
	if !spec.NoList {
		parent.AddCommand(buildListCmd(d, spec))
	}
	if !spec.NoGet {
		parent.AddCommand(buildGetCmd(d, spec))
	}
	if !spec.NoCreate {
		parent.AddCommand(buildCreateCmd(d, spec))
	}
	if !spec.NoUpdate {
		parent.AddCommand(buildUpdateCmd(d, spec))
	}
	if !spec.NoDelete {
		parent.AddCommand(buildDeleteCmd(d, spec))
	}
	for _, ex := range spec.Extra {
		parent.AddCommand(annotate(ex.Build(d), ex.Kind))
	}
	return parent
}

func (s resourceSpec[T]) idArg() string {
	if s.GetByArg != "" {
		return s.GetByArg
	}
	return "id"
}

func buildListCmd[T any](d *deps, spec resourceSpec[T]) *cobra.Command {
	short := spec.ListShort
	if short == "" {
		short = "List " + spec.Use
	}
	var filterVals map[string]*string
	cmd := &cobra.Command{
		Use:   "list",
		Short: short,
		Example: fmt.Sprintf("  cwctl %s list\n  cwctl %s list --all -o json\n  cwctl %s list --filter status=active",
			spec.Use, spec.Use, spec.Use),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, _, err := d.getAPIClient(!spec.NoAuth)
			if err != nil {
				return err
			}
			res := spec.New(c)
			p := api.ListParams{Page: d.gf.page, Sort: d.gf.sort, Extra: map[string]string{}}
			for q, v := range filterVals {
				if v != nil && *v != "" {
					p.Extra[q] = *v
				}
			}
			ctx := cmd.Context()
			var items []T
			if d.gf.all {
				items, err = res.ListAll(ctx, p)
			} else {
				items, err = res.List(ctx, p)
			}
			if err != nil {
				return err
			}
			if c.DryRun {
				return nil
			}
			filtered, err := applyClientFilters(items, d.gf.filter)
			if err != nil {
				return err
			}
			if d.gf.limit > 0 && len(filtered) > d.gf.limit {
				filtered = filtered[:d.gf.limit]
			}
			return d.render(cmd, filtered, spec.Columns)
		},
	}
	filterVals = map[string]*string{}
	for _, f := range spec.ListFilters {
		v := new(string)
		cmd.Flags().StringVar(v, f.Flag, "", f.Usage)
		filterVals[f.query()] = v
	}
	return annotate(cmd, kindRead)
}

func buildGetCmd[T any](d *deps, spec resourceSpec[T]) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get <" + spec.idArg() + ">",
		Short:   "Get a single " + singular(spec.Use),
		Example: fmt.Sprintf("  cwctl %s get 42 -o yaml", spec.Use),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, _, err := d.getAPIClient(!spec.NoAuth)
			if err != nil {
				return err
			}
			item, err := spec.New(c).Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if c.DryRun {
				return nil
			}
			return d.render(cmd, item, spec.Columns)
		},
	}
	return annotate(cmd, kindRead)
}

func buildCreateCmd[T any](d *deps, spec resourceSpec[T]) *cobra.Command {
	short := spec.CreateShort
	if short == "" {
		short = "Create a " + singular(spec.Use)
	}
	cmd := &cobra.Command{
		Use:     "create",
		Short:   short,
		Example: fmt.Sprintf("  cwctl %s create --data '{...}'\n  cwctl %s create -d @payload.json", spec.Use, spec.Use),
		Args:    cobra.NoArgs,
	}
	collect := registerBodyFlags(cmd, spec.CreateFields)
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		c, _, err := d.getAPIClient(!spec.NoAuth)
		if err != nil {
			return err
		}
		body, err := collect(cmd)
		if err != nil {
			return err
		}
		var out json.RawMessage
		if err := spec.New(c).Create(cmd.Context(), body, &out); err != nil {
			return err
		}
		if c.DryRun || d.gf.quiet {
			return nil
		}
		return d.render(cmd, out, spec.Columns)
	}
	return annotate(cmd, kindWrite)
}

func buildUpdateCmd[T any](d *deps, spec resourceSpec[T]) *cobra.Command {
	short := spec.UpdateShort
	if short == "" {
		short = "Update a " + singular(spec.Use)
	}
	fields := spec.UpdateFields
	if fields == nil {
		// Updates are partial edits: when reusing the create fields, required-ness must
		// not carry over (a PATCH without --title is valid; the create POST is not).
		fields = make([]field, len(spec.CreateFields))
		for i, f := range spec.CreateFields {
			f.Required = false
			fields[i] = f
		}
	}
	cmd := &cobra.Command{
		Use:     "update <" + spec.idArg() + ">",
		Short:   short,
		Example: fmt.Sprintf("  cwctl %s update 42 --data '{...}'", spec.Use),
		Args:    cobra.ExactArgs(1),
	}
	collect := registerBodyFlags(cmd, fields)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		c, _, err := d.getAPIClient(!spec.NoAuth)
		if err != nil {
			return err
		}
		body, err := collect(cmd)
		if err != nil {
			return err
		}
		var out json.RawMessage
		if err := spec.New(c).Update(cmd.Context(), args[0], body, &out); err != nil {
			return err
		}
		if c.DryRun || d.gf.quiet {
			return nil
		}
		return d.render(cmd, out, spec.Columns)
	}
	return annotate(cmd, kindWrite)
}

func buildDeleteCmd[T any](d *deps, spec resourceSpec[T]) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <" + spec.idArg() + ">",
		Short:   "Delete a " + singular(spec.Use),
		Example: fmt.Sprintf("  cwctl %s delete 42 --dry-run\n  cwctl %s delete 42", spec.Use, spec.Use),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, _, err := d.getAPIClient(!spec.NoAuth)
			if err != nil {
				return err
			}
			if err := spec.New(c).Delete(cmd.Context(), args[0]); err != nil {
				return err
			}
			if c.DryRun {
				return nil
			}
			if !d.gf.quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "deleted %s %s\n", singular(spec.Use), args[0])
			}
			return nil
		},
	}
	return annotate(cmd, kindDestructive)
}

// registerBodyFlags wires the universal --data flag plus the resource's convenience fields
// and returns the body collector. --data accepts inline JSON, @file, or - (stdin);
// convenience flags override --data keys so a partial payload can be patched inline.
func registerBodyFlags(cmd *cobra.Command, fields []field) func(*cobra.Command) (map[string]any, error) {
	var data string
	cmd.Flags().StringVarP(&data, "data", "d", "", "JSON body: inline, @file, or - for stdin")

	strVals := map[string]*string{}
	intVals := map[string]*int{}
	boolVals := map[string]*bool{}
	floatVals := map[string]*float64{}
	sliceVals := map[string]*[]string{}
	jsonVals := map[string]*string{}

	for _, f := range fields {
		switch f.Kind {
		case fieldString:
			strVals[f.key()] = cmd.Flags().String(f.Flag, "", f.Usage)
		case fieldInt:
			intVals[f.key()] = cmd.Flags().Int(f.Flag, 0, f.Usage)
		case fieldBool:
			boolVals[f.key()] = cmd.Flags().Bool(f.Flag, false, f.Usage)
		case fieldFloat:
			floatVals[f.key()] = cmd.Flags().Float64(f.Flag, 0, f.Usage)
		case fieldStringSlice:
			sliceVals[f.key()] = cmd.Flags().StringSlice(f.Flag, nil, f.Usage)
		case fieldJSON:
			jsonVals[f.key()] = cmd.Flags().String(f.Flag, "", f.Usage+" (JSON)")
		}
		if f.Required {
			// Required-ness is enforced in the collector, not cobra: --data may supply the
			// key, which cobra's MarkFlagRequired could not see.
			continue
		}
	}

	flagToKey := map[string]string{}
	requiredKeys := map[string]string{}
	for _, f := range fields {
		flagToKey[f.Flag] = f.key()
		if f.Required {
			requiredKeys[f.key()] = f.Flag
		}
	}

	return func(cmd *cobra.Command) (map[string]any, error) {
		body := map[string]any{}
		if data != "" {
			raw, err := readDataArg(cmd, data)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(raw, &body); err != nil {
				return nil, fmt.Errorf("--data must be a JSON object: %w", err)
			}
		}
		for _, f := range fields {
			if !cmd.Flags().Changed(f.Flag) {
				continue
			}
			key := f.key()
			switch f.Kind {
			case fieldString:
				body[key] = *strVals[key]
			case fieldInt:
				body[key] = *intVals[key]
			case fieldBool:
				body[key] = *boolVals[key]
			case fieldFloat:
				body[key] = *floatVals[key]
			case fieldStringSlice:
				body[key] = *sliceVals[key]
			case fieldJSON:
				var v any
				if err := json.Unmarshal([]byte(*jsonVals[key]), &v); err != nil {
					return nil, fmt.Errorf("--%s: invalid JSON: %w", f.Flag, err)
				}
				body[key] = v
			}
		}
		for key, flag := range requiredKeys {
			if _, ok := body[key]; !ok {
				return nil, fmt.Errorf("missing required --%s (or a %q key in --data)", flag, key)
			}
		}
		return body, nil
	}
}

// readDataArg resolves a --data value: '-' reads stdin, '@path' reads a file (the user's
// own path — not data-derived, so not confined), anything else is inline JSON.
func readDataArg(cmd *cobra.Command, data string) ([]byte, error) {
	switch {
	case data == "-":
		return io.ReadAll(cmd.InOrStdin())
	case strings.HasPrefix(data, "@"):
		return os.ReadFile(data[1:]) // #nosec G304 -- the user's own --data @file argument
	default:
		return []byte(data), nil
	}
}

// applyClientFilters keeps only items whose JSON fields match every field=value pair.
// Filtering is client-side (post-fetch) so it works uniformly for any resource; dotted
// keys traverse nested objects.
func applyClientFilters[T any](items []T, filters []string) ([]T, error) {
	if len(filters) == 0 {
		return items, nil
	}
	want := map[string]string{}
	for _, f := range filters {
		k, v, ok := strings.Cut(f, "=")
		if !ok {
			return nil, fmt.Errorf("invalid --filter %q (want field=value)", f)
		}
		want[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	var out []T
	for _, it := range items {
		b, err := json.Marshal(it)
		if err != nil {
			return nil, err
		}
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		match := true
		for k, v := range want {
			if stringifyField(lookupPath(m, k)) != v {
				match = false
				break
			}
		}
		if match {
			out = append(out, it)
		}
	}
	return out, nil
}

// lookupPath walks dotted keys ("meta.sender.name") through nested maps.
func lookupPath(m map[string]any, path string) any {
	parts := strings.Split(path, ".")
	var cur any = m
	for _, p := range parts {
		mm, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = mm[p]
	}
	return cur
}

func stringifyField(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'g', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}

// itoa keeps query-building call sites terse.
func itoa(n int) string { return strconv.Itoa(n) }

// errMissingOneOf standardizes the "pick at least one flag" failure.
func errMissingOneOf(flags ...string) error {
	return fmt.Errorf("provide at least one of %s", strings.Join(flags, ", "))
}

// singular is a crude depluralizer good enough for our resource names.
func singular(s string) string {
	if len(s) > 1 && strings.HasSuffix(s, "s") {
		return s[:len(s)-1]
	}
	return s
}

// runE wraps a custom verb's body with the shared client/dry-run/render plumbing so Extra
// commands stay one closure each. fn returns nil output for render-nothing commands.
func runE(d *deps, noAuth bool, cols []string, fn func(*cobra.Command, *api.Client, []string) (json.RawMessage, error)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		c, _, err := d.getAPIClient(!noAuth)
		if err != nil {
			return err
		}
		out, err := fn(cmd, c, args)
		if err != nil {
			return err
		}
		if c.DryRun || out == nil {
			return nil
		}
		return d.render(cmd, out, cols)
	}
}

// runListE is runE for custom verbs whose response is a LIST in one of Chatwoot's
// envelopes: it normalizes {data:{meta,payload}} / {payload:[…]} / bare arrays to rows so
// tables and `-o id` work, instead of rendering the wrapper object.
func runListE(d *deps, noAuth bool, cols []string, fn func(*cobra.Command, *api.Client, []string) (json.RawMessage, error)) func(*cobra.Command, []string) error {
	return runE(d, noAuth, cols, func(cmd *cobra.Command, c *api.Client, args []string) (json.RawMessage, error) {
		out, err := fn(cmd, c, args)
		if err != nil || out == nil {
			return nil, err
		}
		items, err := api.NormalizeList(out)
		if err != nil {
			// Not list-shaped after all (some deployments return a bare object) — render as-is.
			return out, nil //nolint:nilerr // fallback by design
		}
		normalized, err := json.Marshal(items)
		if err != nil {
			return nil, err
		}
		return normalized, nil
	})
}
