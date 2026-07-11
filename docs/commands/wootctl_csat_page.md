## wootctl csat page

Print (or fetch with --fetch) the CSAT survey page URL for a conversation

```
wootctl csat page <conversation-uuid> [flags]
```

### Examples

```
  wootctl csat page 8f286537-3216-4d47-a869-6a08128d9dc9
  wootctl csat page 8f286537-3216-4d47-a869-6a08128d9dc9 --fetch > survey.html
```

### Options

```
      --fetch   fetch the page instead of printing its URL
  -h, --help    help for page
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

* [wootctl csat](wootctl_csat.md)	 - CSAT survey page for a conversation

