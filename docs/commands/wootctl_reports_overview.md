## wootctl reports overview

Time-series statistics for a metric (account/agent/inbox/label/team)

```
wootctl reports overview [flags]
```

### Examples

```
  wootctl reports overview --metric conversations_count --type account --since 2026-06-01 --until 2026-07-01
  wootctl reports overview --metric avg_first_response_time --type inbox --id 3 --since 2026-06-01
```

### Options

```
  -h, --help            help for overview
      --id string       object id when type != account
      --metric string   conversations_count | incoming_messages_count | outgoing_messages_count | avg_first_response_time | avg_resolution_time | resolutions_count | bot_resolutions_count | bot_handoffs_count | reply_time
      --since string    range start: unix seconds or YYYY-MM-DD
      --type string     account | agent | inbox | label | team (default "account")
      --until string    range end: unix seconds or YYYY-MM-DD
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

