## wootctl conversations create

Create a conversation

```
wootctl conversations create [flags]
```

### Examples

```
  wootctl conversations create --data '{...}'
  wootctl conversations create -d @payload.json
```

### Options

```
      --additional-attributes string   additional attributes object (JSON)
      --assignee-id int                agent to assign
      --contact-id int                 contact for the conversation
      --custom-attributes string       custom attributes object (JSON)
  -d, --data string                    JSON body: inline, @file, or - for stdin
  -h, --help                           help for create
      --inbox-id int                   inbox for the conversation
      --message string                 first message, e.g. '{"content":"hola"}' (JSON)
      --source-id string               contact-inbox source id
      --status string                  open | resolved | pending
      --team-id int                    team to assign
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

* [wootctl conversations](wootctl_conversations.md)	 - Manage conversations

