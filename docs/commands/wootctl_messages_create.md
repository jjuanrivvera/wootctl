## wootctl messages create

Send a message (reply) in a conversation

### Synopsis

Send an outgoing message, a private note (--private), or an incoming message
(--message-type incoming, API channels). Attach files with repeatable --attachment
(switches to a multipart upload).

```
wootctl messages create <conversation-id> [flags]
```

### Examples

```
  wootctl messages create 42 --content "On it — checking now."
  wootctl messages create 42 --content "internal note" --private
  wootctl messages create 42 --content "see attached" --attachment ./invoice.pdf
```

### Options

```
      --attachment strings          file to attach (repeatable; forces multipart)
      --campaign-id int             campaign id
      --content string              message text
      --content-attributes string   content attributes object (interactive messages) (JSON)
      --content-type string         text | input_email | cards | input_select | form | article
  -d, --data string                 JSON body: inline, @file, or - for stdin
  -h, --help                        help for create
      --message-type string         outgoing | incoming (default outgoing)
      --private                     private note (not visible to the contact)
      --template-params string      WhatsApp template params object (JSON)
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

* [wootctl messages](wootctl_messages.md)	 - Read and send messages in a conversation

