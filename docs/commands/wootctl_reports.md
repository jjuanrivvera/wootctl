## wootctl reports

Account analytics (v2 reports, summaries, reporting events)

### Options

```
  -h, --help   help for reports
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

* [wootctl](wootctl.md)	 - A fast, scriptable CLI for the full Chatwoot API
* [wootctl reports account-conversations](wootctl_reports_account-conversations.md)	 - Account-level open/unattended conversation metrics
* [wootctl reports agent-conversations](wootctl_reports_agent-conversations.md)	 - Per-agent conversation metrics (all agents, or one with --user-id)
* [wootctl reports agent-summary](wootctl_reports_agent-summary.md)	 - Summary report grouped by agent
* [wootctl reports channel-summary](wootctl_reports_channel-summary.md)	 - Summary report grouped by channel
* [wootctl reports events](wootctl_reports_events.md)	 - List account reporting events (first response, resolutions, …)
* [wootctl reports first-response-time-distribution](wootctl_reports_first-response-time-distribution.md)	 - First-response-time distribution buckets
* [wootctl reports inbox-label-matrix](wootctl_reports_inbox-label-matrix.md)	 - Conversation counts as an inbox × label matrix
* [wootctl reports inbox-summary](wootctl_reports_inbox-summary.md)	 - Summary report grouped by inbox
* [wootctl reports outgoing-messages-count](wootctl_reports_outgoing-messages-count.md)	 - Outgoing message counts grouped by day/week/month/year
* [wootctl reports overview](wootctl_reports_overview.md)	 - Time-series statistics for a metric (account/agent/inbox/label/team)
* [wootctl reports summary](wootctl_reports_summary.md)	 - Aggregate statistics for a range (conversations, response times, resolutions)
* [wootctl reports team-summary](wootctl_reports_team-summary.md)	 - Summary report grouped by team

