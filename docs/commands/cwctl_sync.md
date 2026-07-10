## cwctl sync

Copy account config from one instance to another (the multi-instance payoff)

### Synopsis

Reconcile the active profile's account config INTO another profile's account:
create missing, update changed, skip unchanged; --prune removes resources on the target that
the source lacks. This is what the single-instance official CLI can't do — promote labels,
canned responses, and automation from staging to production, or keep two support instances
aligned. Matching and safety are identical to restore.

Always dry-run first.

```
cwctl sync --to <profile> [flags]
```

### Examples

```
  cwctl sync --to acue --dry-run
  cwctl sync --to acue --only canned-responses,labels
  cwctl --profile staging sync --to prod --prune
```

### Options

```
      --from string    source profile (default: the active profile)
  -h, --help           help for sync
      --only strings   restrict to these resource kinds (comma-separated)
      --prune          delete target resources not present on the source
      --to string      destination profile (required)
```

### Options inherited from parent commands

```
      --account-id string   override the profile's account id for this invocation
      --all                 fetch all pages (list commands)
      --base-url string     override the instance base URL
      --columns strings     comma-separated columns to show
      --dry-run             print the equivalent curl and make no request
      --filter strings      client-side field=value filters (list commands)
      --jq string           gojq expression applied to the response before rendering
      --limit int           max items to output, applied client-side (list commands)
      --no-color            disable colored output
  -o, --output string       output format: table|json|yaml|csv|id
      --page int            page number to fetch (list commands; Chatwoot pages are server-sized)
      --profile string      named profile to use (instance + account + token)
      --quiet               suppress non-essential chatter
      --rps rps             max requests per second (default 5; also per-profile rps in config)
      --show-token          reveal the API token in dry-run output
      --sort string         sort field, prefix with - for descending (where the API supports it)
  -v, --verbose             verbose request logging (stderr)
```

### SEE ALSO

* [cwctl](cwctl.md)	 - A fast, scriptable CLI for the full Chatwoot API

