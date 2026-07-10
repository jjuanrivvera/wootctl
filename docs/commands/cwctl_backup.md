## cwctl backup

Back up account config (labels, canned responses, automation, teams, …) to a git-friendly dir

### Synopsis

Write the account's portable CONFIG to a directory of YAML files (one per resource
kind), keeping only the writable fields so the output is stable and diffable in git.
Conversations, contacts, and messages are live data, not config, and are never backed up.

```
cwctl backup --dir <dir> [flags]
```

### Examples

```
  cwctl backup --dir ./chatwoot-config
  cwctl backup --dir ./cfg --only labels,canned-responses
```

### Options

```
      --dir string     directory to write the backup into (required)
  -h, --help           help for backup
      --only strings   restrict to these resource kinds (comma-separated)
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

