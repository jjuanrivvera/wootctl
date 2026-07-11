## wootctl reports events

List account reporting events (first response, resolutions, …)

```
wootctl reports events [flags]
```

### Examples

```
  wootctl reports events --name first_response --since 2026-06-01
```

### Options

```
  -h, --help              help for events
      --inbox-id string   only this inbox
      --name string       event name: conversation_creation | first_response | conversation_resolved | …
      --since string      range start: unix seconds or YYYY-MM-DD
      --until string      range end: unix seconds or YYYY-MM-DD
      --user-id string    only this agent
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

* [wootctl reports](wootctl_reports.md)	 - Account analytics (v2 reports, summaries, reporting events)

