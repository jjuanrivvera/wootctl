## wootctl inboxes

Manage inboxes and their members

### Options

```
  -h, --help   help for inboxes
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
* [wootctl inboxes add-members](wootctl_inboxes_add-members.md)	 - Add agents to an inbox
* [wootctl inboxes agent-bot](wootctl_inboxes_agent-bot.md)	 - Show the agent bot attached to an inbox
* [wootctl inboxes create](wootctl_inboxes_create.md)	 - Create a inboxe
* [wootctl inboxes get](wootctl_inboxes_get.md)	 - Get a single inboxe
* [wootctl inboxes list](wootctl_inboxes_list.md)	 - List inboxes
* [wootctl inboxes members](wootctl_inboxes_members.md)	 - List the agents in an inbox
* [wootctl inboxes remove-members](wootctl_inboxes_remove-members.md)	 - Remove agents from an inbox
* [wootctl inboxes set-agent-bot](wootctl_inboxes_set-agent-bot.md)	 - Attach an agent bot to an inbox (0 detaches)
* [wootctl inboxes update](wootctl_inboxes_update.md)	 - Update a inboxe
* [wootctl inboxes update-members](wootctl_inboxes_update-members.md)	 - Replace an inbox's agents

