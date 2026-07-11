## wootctl platform agent-bots

Manage global agent bots (platform app token)

### Options

```
  -h, --help   help for agent-bots
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

* [wootctl platform](wootctl_platform.md)	 - Platform API (accounts, users, agent bots) — needs a platform app token
* [wootctl platform agent-bots create](wootctl_platform_agent-bots_create.md)	 - Create a agent-bot
* [wootctl platform agent-bots delete](wootctl_platform_agent-bots_delete.md)	 - Delete a agent-bot
* [wootctl platform agent-bots get](wootctl_platform_agent-bots_get.md)	 - Get a single agent-bot
* [wootctl platform agent-bots list](wootctl_platform_agent-bots_list.md)	 - List agent-bots
* [wootctl platform agent-bots update](wootctl_platform_agent-bots_update.md)	 - Update a agent-bot

