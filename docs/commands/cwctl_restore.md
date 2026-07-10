## cwctl restore

Reconcile a backup dir into the account (create/update/skip; --prune removes drift)

### Synopsis

Apply a backup directory to the active profile's account: create missing resources,
update changed ones (comparing only writable fields), skip unchanged. With --prune, live
resources absent from the backup are deleted. Matching is by natural key (title, short_code,
name, url, attribute_key); a key duplicated in the account is skipped, never pruned.

Always dry-run first — restore mutates real config.

```
cwctl restore --dir <dir> [flags]
```

### Examples

```
  cwctl restore --dir ./chatwoot-config --dry-run
  cwctl restore --dir ./chatwoot-config
  cwctl restore --dir ./chatwoot-config --only labels --prune
```

### Options

```
      --dir string     backup directory to apply (required)
  -h, --help           help for restore
      --only strings   restrict to these resource kinds (comma-separated)
      --prune          delete live resources not present in the backup
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

