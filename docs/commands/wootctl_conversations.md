## wootctl conversations

Manage conversations

### Options

```
  -h, --help   help for conversations
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
* [wootctl conversations add-labels](wootctl_conversations_add-labels.md)	 - Add labels to a conversation (replaces the label set)
* [wootctl conversations assign](wootctl_conversations_assign.md)	 - Assign a conversation to an agent or a team
* [wootctl conversations create](wootctl_conversations_create.md)	 - Create a conversation
* [wootctl conversations filter](wootctl_conversations_filter.md)	 - Filter conversations with the query DSL
* [wootctl conversations get](wootctl_conversations_get.md)	 - Get a single conversation
* [wootctl conversations labels](wootctl_conversations_labels.md)	 - List a conversation's labels
* [wootctl conversations list](wootctl_conversations_list.md)	 - List conversations
* [wootctl conversations meta](wootctl_conversations_meta.md)	 - Conversation counts (mine, unassigned, assigned, all)
* [wootctl conversations reporting-events](wootctl_conversations_reporting-events.md)	 - List a conversation's reporting events (first response, resolved, …)
* [wootctl conversations set-custom-attributes](wootctl_conversations_set-custom-attributes.md)	 - Set custom attributes on a conversation
* [wootctl conversations toggle-priority](wootctl_conversations_toggle-priority.md)	 - Change a conversation's priority
* [wootctl conversations toggle-status](wootctl_conversations_toggle-status.md)	 - Change a conversation's status (open/resolved/pending/snoozed)
* [wootctl conversations toggle-typing](wootctl_conversations_toggle-typing.md)	 - Flip the typing indicator on or off
* [wootctl conversations update](wootctl_conversations_update.md)	 - Update a conversation

