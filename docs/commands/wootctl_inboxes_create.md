## wootctl inboxes create

Create a inboxe

```
wootctl inboxes create [flags]
```

### Examples

```
  wootctl inboxes create --data '{...}'
  wootctl inboxes create -d @payload.json
```

### Options

```
      --channel string                 channel config, e.g. '{"type":"web_widget","website_url":"https://invitas.co"}' or '{"type":"api","webhook_url":"https://…"}' (JSON)
      --csat-survey-enabled            send CSAT surveys on resolve
  -d, --data string                    JSON body: inline, @file, or - for stdin
      --enable-auto-assignment         auto-assign conversations
      --greeting-enabled               send a greeting on first message
      --greeting-message string        greeting text
  -h, --help                           help for create
      --lock-to-single-conversation    one conversation per contact
      --name string                    inbox name
      --out-of-office-message string   out-of-office reply
      --portal-id int                  help-center portal to link
      --timezone string                IANA timezone
      --working-hours-enabled          enable working hours
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

* [wootctl inboxes](wootctl_inboxes.md)	 - Manage inboxes and their members

