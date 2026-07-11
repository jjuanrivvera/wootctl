## wootctl client

Public client API (inbox/contact/conversation flows) — no token required

### Options

```
  -h, --help   help for client
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
* [wootctl client contacts](wootctl_client_contacts.md)	 - Public (contact-facing) contact endpoints
* [wootctl client conversations](wootctl_client_conversations.md)	 - Public (contact-facing) conversation endpoints
* [wootctl client inbox](wootctl_client_inbox.md)	 - Public inbox details
* [wootctl client messages](wootctl_client_messages.md)	 - Public (contact-facing) message endpoints

