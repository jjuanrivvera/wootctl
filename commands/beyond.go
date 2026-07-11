package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/jjuanrivvera/wootctl/internal/api"
	"github.com/jjuanrivvera/wootctl/internal/config"
)

// Beyond-the-API layer (cliwright GOAL.md §3c): git-friendly backup/restore of an account's
// CONFIG and cross-instance sync. These are the multi-instance payoff the single-instance
// official CLI can't offer — they are NOT raw API endpoints, so they live outside
// api-manifest.json by design (DECISIONS.md #21).

// portableKind describes a config resource that can be backed up, restored, and synced.
// Only account CONFIG is portable — never conversations/contacts/messages (that's live
// data, not config). Matching is by a natural Key (Chatwoot exposes no stable cross-instance
// handle), and "unchanged" is decided by comparing ONLY the Writable fields, which auto-drops
// id/created_at/server-managed noise (GOAL.md §3c).
type portableKind struct {
	Name     string
	Key      string   // natural-key field, e.g. "title"
	Writable []string // fields compared and sent on create/update (includes Key)
	New      func(*api.Client) *api.Resource[api.Rec]
}

// portableKinds is the ordered registry. Order matters for restore/backup determinism.
var portableKinds = []portableKind{
	{Name: "labels", Key: "title",
		Writable: []string{"title", "description", "color", "show_on_sidebar"},
		New:      func(c *api.Client) *api.Resource[api.Rec] { return c.Labels() }},
	{Name: "canned-responses", Key: "short_code",
		Writable: []string{"short_code", "content"},
		New:      func(c *api.Client) *api.Resource[api.Rec] { return c.CannedResponses() }},
	{Name: "custom-attributes", Key: "attribute_key",
		Writable: []string{"attribute_display_name", "attribute_key", "attribute_model", "attribute_display_type", "attribute_description", "attribute_values", "regex_pattern", "regex_cue"},
		New:      func(c *api.Client) *api.Resource[api.Rec] { return c.CustomAttributes() }},
	{Name: "custom-filters", Key: "name",
		Writable: []string{"name", "type", "query"},
		New:      func(c *api.Client) *api.Resource[api.Rec] { return c.CustomFilters() }},
	{Name: "automation-rules", Key: "name",
		Writable: []string{"name", "description", "event_name", "active", "conditions", "actions"},
		New:      func(c *api.Client) *api.Resource[api.Rec] { return c.AutomationRules() }},
	{Name: "teams", Key: "name",
		Writable: []string{"name", "description", "allow_auto_assign"},
		New:      func(c *api.Client) *api.Resource[api.Rec] { return c.Teams() }},
	{Name: "webhooks", Key: "url",
		Writable: []string{"url", "subscriptions"},
		New:      func(c *api.Client) *api.Resource[api.Rec] { return c.Webhooks() }},
	{Name: "agent-bots", Key: "name",
		Writable: []string{"name", "description", "outgoing_url"},
		New:      func(c *api.Client) *api.Resource[api.Rec] { return c.AgentBots() }},
}

func portableKindByName(name string) (portableKind, bool) {
	for _, k := range portableKinds {
		if k.Name == name {
			return k, true
		}
	}
	return portableKind{}, false
}

// selectKinds resolves an --only filter to the ordered subset (all kinds when empty),
// erroring on an unknown name so a typo never silently backs up nothing.
func selectKinds(only []string) ([]portableKind, error) {
	if len(only) == 0 {
		return portableKinds, nil
	}
	want := map[string]bool{}
	for _, n := range only {
		if _, ok := portableKindByName(n); !ok {
			return nil, fmt.Errorf("unknown resource %q (portable kinds: %s)", n, strings.Join(portableKindNames(), ", "))
		}
		want[n] = true
	}
	var out []portableKind
	for _, k := range portableKinds {
		if want[k.Name] {
			out = append(out, k)
		}
	}
	return out, nil
}

func portableKindNames() []string {
	names := make([]string, len(portableKinds))
	for i, k := range portableKinds {
		names[i] = k.Name
	}
	return names
}

// --- reconcile engine ---

type action string

const (
	actCreate action = "create"
	actUpdate action = "update"
	actSkip   action = "skip"
	actPrune  action = "prune"
)

type planItem struct {
	Kind   string
	Key    string
	Action action
	id     string         // live id, for update/prune
	body   map[string]any // desired writable subset, for create/update
}

// reconcile diffs desired records against the live target set for one kind. Matching is by
// natural Key. A key that is DUPLICATED in the live set is skipped with a warning and never
// pruned — acting on an arbitrary one of two same-named resources is the
// "--prune deletes the wrong thing" bug (GOAL.md §3c).
func reconcile(ctx context.Context, target *api.Client, kind portableKind, desired []map[string]any, prune bool, warn func(string)) ([]planItem, error) {
	liveRecs, err := kind.New(target).ListAll(ctx, api.ListParams{})
	if err != nil {
		return nil, fmt.Errorf("list live %s: %w", kind.Name, err)
	}
	live, err := recsToMaps(liveRecs)
	if err != nil {
		return nil, err
	}
	liveByKey, liveDup := indexByKey(live, kind.Key)
	for k := range liveDup {
		warn(fmt.Sprintf("%s: %d live items share %s=%q — skipped (ambiguous match)", kind.Name, liveDup[k], kind.Key, k))
	}

	desiredByKey := map[string]map[string]any{}
	var desiredOrder []string
	for _, d := range desired {
		key, ok := recordKey(d, kind.Key)
		if !ok || key == "" {
			warn(fmt.Sprintf("%s: a desired record has no %s — skipped", kind.Name, kind.Key))
			continue
		}
		if _, dup := desiredByKey[key]; dup {
			warn(fmt.Sprintf("%s: desired set has two records with %s=%q — using the first", kind.Name, kind.Key, key))
			continue
		}
		desiredByKey[key] = d
		desiredOrder = append(desiredOrder, key)
	}

	var plan []planItem
	for _, key := range desiredOrder {
		d := desiredByKey[key]
		body := project(d, kind.allFields())
		if liveDup[key] > 0 {
			continue // ambiguous — already warned; never touch
		}
		liveRec, ok := liveByKey[key]
		if !ok {
			plan = append(plan, planItem{Kind: kind.Name, Key: key, Action: actCreate, body: body})
			continue
		}
		if writableEqual(liveRec, d, kind.Writable) {
			plan = append(plan, planItem{Kind: kind.Name, Key: key, Action: actSkip, id: idOf(liveRec)})
		} else {
			plan = append(plan, planItem{Kind: kind.Name, Key: key, Action: actUpdate, id: idOf(liveRec), body: body})
		}
	}
	if prune {
		var pruneKeys []string
		for key := range liveByKey {
			if _, want := desiredByKey[key]; !want && liveDup[key] == 0 {
				pruneKeys = append(pruneKeys, key)
			}
		}
		sort.Strings(pruneKeys)
		for _, key := range pruneKeys {
			plan = append(plan, planItem{Kind: kind.Name, Key: key, Action: actPrune, id: idOf(liveByKey[key])})
		}
	}
	return plan, nil
}

func (k portableKind) allFields() []string {
	if contains(k.Writable, k.Key) {
		return k.Writable
	}
	return append([]string{k.Key}, k.Writable...)
}

// execute applies a plan against the target client. dryRun makes it print only.
func execute(ctx context.Context, target *api.Client, kind portableKind, plan []planItem, dryRun bool, out func(string)) error {
	res := kind.New(target)
	for _, it := range plan {
		switch it.Action {
		case actSkip:
			continue
		case actCreate:
			out(fmt.Sprintf("  + create %s %q", kind.Name, it.Key))
			if !dryRun {
				if err := res.Create(ctx, it.body, nil); err != nil {
					return fmt.Errorf("create %s %q: %w", kind.Name, it.Key, err)
				}
			}
		case actUpdate:
			out(fmt.Sprintf("  ~ update %s %q", kind.Name, it.Key))
			if !dryRun {
				if err := res.Update(ctx, it.id, it.body, nil); err != nil {
					return fmt.Errorf("update %s %q: %w", kind.Name, it.Key, err)
				}
			}
		case actPrune:
			out(fmt.Sprintf("  - prune %s %q", kind.Name, it.Key))
			if !dryRun {
				if err := res.Delete(ctx, it.id); err != nil {
					return fmt.Errorf("prune %s %q: %w", kind.Name, it.Key, err)
				}
			}
		}
	}
	return nil
}

// planCounts summarizes a plan for the run header.
func planCounts(plan []planItem) (create, update, skip, prune int) {
	for _, it := range plan {
		switch it.Action {
		case actCreate:
			create++
		case actUpdate:
			update++
		case actSkip:
			skip++
		case actPrune:
			prune++
		}
	}
	return
}

// --- record helpers ---

func recsToMaps(recs []api.Rec) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(recs))
	for _, r := range recs {
		var m map[string]any
		if err := json.Unmarshal(r, &m); err != nil {
			return nil, fmt.Errorf("decode record: %w", err)
		}
		out = append(out, m)
	}
	return out, nil
}

func recordKey(rec map[string]any, keyField string) (string, bool) {
	v, ok := rec[keyField]
	if !ok || v == nil {
		return "", false
	}
	return stringifyField(v), true
}

func idOf(rec map[string]any) string {
	if v, ok := rec["id"]; ok && v != nil {
		return stringifyField(v)
	}
	return ""
}

// indexByKey maps records by their natural key and reports how many share each key when >1.
func indexByKey(records []map[string]any, keyField string) (map[string]map[string]any, map[string]int) {
	byKey := map[string]map[string]any{}
	count := map[string]int{}
	for _, rec := range records {
		key, ok := recordKey(rec, keyField)
		if !ok || key == "" {
			continue
		}
		count[key]++
		if _, seen := byKey[key]; !seen {
			byKey[key] = rec
		}
	}
	dup := map[string]int{}
	for k, n := range count {
		if n > 1 {
			dup[k] = n
		}
	}
	return byKey, dup
}

// project keeps only the named fields that are present (used for backup output and bodies).
func project(rec map[string]any, fields []string) map[string]any {
	out := map[string]any{}
	for _, f := range fields {
		if v, ok := rec[f]; ok {
			out[f] = v
		}
	}
	return out
}

// writableEqual reports whether every writable field matches, comparing canonical JSON so
// field order and equivalent number/array encodings don't create phantom diffs. A field
// absent on the desired side is treated as "leave as-is" (not a diff) so a partial backup
// doesn't force spurious updates.
func writableEqual(live, desired map[string]any, writable []string) bool {
	for _, f := range writable {
		dv, has := desired[f]
		if !has {
			continue
		}
		if canonicalJSON(dv) != canonicalJSON(live[f]) {
			return false
		}
	}
	return true
}

func canonicalJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// --- backup / restore file format ---

// backupFile is the on-disk shape per kind: a plain list of writable-projected records.
func backupPath(dir, kind string) string { return filepath.Join(dir, kind+".yaml") }

func writeBackup(dir string, kind portableKind, records []map[string]any) error {
	projected := make([]map[string]any, 0, len(records))
	for _, r := range records {
		projected = append(projected, project(r, kind.allFields()))
	}
	data, err := yaml.Marshal(projected)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	return os.WriteFile(backupPath(dir, kind.Name), data, 0o600)
}

func readBackup(dir string, kind portableKind) ([]map[string]any, error) {
	data, err := os.ReadFile(backupPath(dir, kind.Name)) // #nosec G304 -- user's own --dir
	if os.IsNotExist(err) {
		return nil, nil // a kind absent from the dir is simply not reconciled
	}
	if err != nil {
		return nil, err
	}
	var records []map[string]any
	if err := yaml.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("parse %s: %w", backupPath(dir, kind.Name), err)
	}
	return records, nil
}

// --- commands ---

func init() {
	metaRegistrars = append(metaRegistrars, backupCmd, restoreCmd, syncCmd)
}

func backupCmd(d *deps) *cobra.Command {
	var dir string
	var only []string
	cmd := &cobra.Command{
		Use:   "backup --dir <dir>",
		Short: "Back up account config (labels, canned responses, automation, teams, …) to a git-friendly dir",
		Long: `Write the account's portable CONFIG to a directory of YAML files (one per resource
kind), keeping only the writable fields so the output is stable and diffable in git.
Conversations, contacts, and messages are live data, not config, and are never backed up.`,
		Example: `  wootctl backup --dir ./chatwoot-config
  wootctl backup --dir ./cfg --only labels,canned-responses`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			kinds, err := selectKinds(only)
			if err != nil {
				return err
			}
			if dir == "" {
				return fmt.Errorf("--dir is required")
			}
			c, _, err := d.getAPIClient(true)
			if err != nil {
				return err
			}
			c.DryRun = false // backup only reads the API and writes local files; never a curl no-op
			for _, k := range kinds {
				recs, err := k.New(c).ListAll(cmd.Context(), api.ListParams{})
				if err != nil {
					return fmt.Errorf("list %s: %w", k.Name, err)
				}
				maps, err := recsToMaps(recs)
				if err != nil {
					return err
				}
				if err := writeBackup(dir, k, maps); err != nil {
					return err
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "backed up %d %s → %s\n", len(maps), k.Name, backupPath(dir, k.Name))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "directory to write the backup into (required)")
	cmd.Flags().StringSliceVar(&only, "only", nil, "restrict to these resource kinds (comma-separated)")
	return annotate(cmd, kindRead)
}

func restoreCmd(d *deps) *cobra.Command {
	var dir string
	var only []string
	var prune bool
	cmd := &cobra.Command{
		Use:   "restore --dir <dir>",
		Short: "Reconcile a backup dir into the account (create/update/skip; --prune removes drift)",
		Long: `Apply a backup directory to the active profile's account: create missing resources,
update changed ones (comparing only writable fields), skip unchanged. With --prune, live
resources absent from the backup are deleted. Matching is by natural key (title, short_code,
name, url, attribute_key); a key duplicated in the account is skipped, never pruned.

Always dry-run first — restore mutates real config.`,
		Example: `  wootctl restore --dir ./chatwoot-config --dry-run
  wootctl restore --dir ./chatwoot-config
  wootctl restore --dir ./chatwoot-config --only labels --prune`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			kinds, err := selectKinds(only)
			if err != nil {
				return err
			}
			if dir == "" {
				return fmt.Errorf("--dir is required")
			}
			c, _, err := d.getAPIClient(true)
			if err != nil {
				return err
			}
			// The plan needs live reads even under --dry-run; --dry-run only suppresses WRITES.
			c.DryRun = false
			desiredFor := func(k portableKind) ([]map[string]any, error) { return readBackup(dir, k) }
			return runReconcile(cmd, c, kinds, desiredFor, prune, d.gf.dryRun)
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "backup directory to apply (required)")
	cmd.Flags().StringSliceVar(&only, "only", nil, "restrict to these resource kinds (comma-separated)")
	cmd.Flags().BoolVar(&prune, "prune", false, "delete live resources not present in the backup")
	// Destructive: with --prune it deletes config. The guard hard-blocks it for agents.
	return annotate(cmd, kindDestructive)
}

func syncCmd(d *deps) *cobra.Command {
	var toProfile, fromProfile string
	var only []string
	var prune bool
	cmd := &cobra.Command{
		Use:   "sync --to <profile>",
		Short: "Copy account config from one instance to another (the multi-instance payoff)",
		Long: `Reconcile the active profile's account config INTO another profile's account:
create missing, update changed, skip unchanged; --prune removes resources on the target that
the source lacks. This is what the single-instance official CLI can't do — promote labels,
canned responses, and automation from staging to production, or keep two support instances
aligned. Matching and safety are identical to restore.

Always dry-run first.`,
		Example: `  wootctl sync --to acue --dry-run
  wootctl sync --to acue --only canned-responses,labels
  wootctl --profile staging sync --to prod --prune`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			kinds, err := selectKinds(only)
			if err != nil {
				return err
			}
			if toProfile == "" {
				return fmt.Errorf("--to <profile> is required")
			}
			cfg, err := d.loadConfig()
			if err != nil {
				return err
			}
			sourceName := cfg.ResolveProfileName(config.FirstNonEmpty(fromProfile, d.gf.profile))
			if toProfile == sourceName {
				return fmt.Errorf("--to and the source profile are both %q; choose different instances", toProfile)
			}
			if _, ok := cfg.Profile(toProfile); !ok {
				return fmt.Errorf("no such profile %q — create it with `wootctl --profile %s auth login`", toProfile, toProfile)
			}
			source, err := d.clientForProfile(cfg, sourceName, true, true)
			if err != nil {
				return fmt.Errorf("source profile %q: %w", sourceName, err)
			}
			target, err := d.clientForProfile(cfg, toProfile, true, false)
			if err != nil {
				return fmt.Errorf("target profile %q: %w", toProfile, err)
			}
			// Reads (source + target live state) must run to compute the plan even under
			// --dry-run; only the target WRITES are gated by dryRun in execute().
			source.DryRun, target.DryRun = false, false
			fmt.Fprintf(cmd.ErrOrStderr(), "sync %s → %s\n", sourceName, toProfile)
			desiredFor := func(k portableKind) ([]map[string]any, error) {
				recs, err := k.New(source).ListAll(cmd.Context(), api.ListParams{})
				if err != nil {
					return nil, fmt.Errorf("read source %s: %w", k.Name, err)
				}
				return recsToMaps(recs)
			}
			return runReconcile(cmd, target, kinds, desiredFor, prune, d.gf.dryRun)
		},
	}
	cmd.Flags().StringVar(&toProfile, "to", "", "destination profile (required)")
	cmd.Flags().StringVar(&fromProfile, "from", "", "source profile (default: the active profile)")
	cmd.Flags().StringSliceVar(&only, "only", nil, "restrict to these resource kinds (comma-separated)")
	cmd.Flags().BoolVar(&prune, "prune", false, "delete target resources not present on the source")
	return annotate(cmd, kindDestructive)
}

// runReconcile is the shared plan-and-apply loop for restore and sync.
func runReconcile(cmd *cobra.Command, target *api.Client, kinds []portableKind, desiredFor func(portableKind) ([]map[string]any, error), prune, dryRun bool) error {
	warn := func(s string) { fmt.Fprintln(cmd.ErrOrStderr(), "warning:", s) }
	out := func(s string) { fmt.Fprintln(cmd.OutOrStdout(), s) }
	var tc, tu, ts, tp int
	for _, k := range kinds {
		desired, err := desiredFor(k)
		if err != nil {
			return err
		}
		if desired == nil {
			continue
		}
		plan, err := reconcile(cmd.Context(), target, k, desired, prune, warn)
		if err != nil {
			return err
		}
		c, u, s, p := planCounts(plan)
		tc, tu, ts, tp = tc+c, tu+u, ts+s, tp+p
		if c+u+p > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "%s: +%d ~%d -%d (=%d unchanged)\n", k.Name, c, u, p, s)
		}
		if err := execute(cmd.Context(), target, k, plan, dryRun, out); err != nil {
			return err
		}
	}
	verb := "applied"
	if dryRun {
		verb = "planned (dry-run)"
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "%s: %d created, %d updated, %d pruned, %d unchanged\n", verb, tc, tu, tp, ts)
	return nil
}
