## wootctl teams

Manage teams and their members

### Options

```
  -h, --help   help for teams
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
* [wootctl teams add-members](wootctl_teams_add-members.md)	 - Add agents to a team
* [wootctl teams create](wootctl_teams_create.md)	 - Create a team
* [wootctl teams delete](wootctl_teams_delete.md)	 - Delete a team
* [wootctl teams get](wootctl_teams_get.md)	 - Get a single team
* [wootctl teams list](wootctl_teams_list.md)	 - List teams
* [wootctl teams members](wootctl_teams_members.md)	 - List the agents in a team
* [wootctl teams remove-members](wootctl_teams_remove-members.md)	 - Remove agents from a team
* [wootctl teams update](wootctl_teams_update.md)	 - Update a team
* [wootctl teams update-members](wootctl_teams_update-members.md)	 - Replace a team's agents

