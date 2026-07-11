## wootctl contacts merge

Merge two contacts (the mergee is deleted)

### Synopsis

Merge the mergee contact into the base contact. The base keeps all conversations and data; the mergee is deleted. This cannot be undone.

```
wootctl contacts merge [flags]
```

### Examples

```
  wootctl contacts merge --base 1 --mergee 2
  wootctl contacts merge --base 1 --mergee 2 --dry-run
```

### Options

```
      --base int     contact id that remains after the merge
  -h, --help         help for merge
      --mergee int   contact id merged into the base and deleted
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

* [wootctl contacts](wootctl_contacts.md)	 - Manage contacts

