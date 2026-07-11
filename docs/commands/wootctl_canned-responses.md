## wootctl canned-responses

Manage canned responses (saved reply snippets)

### Options

```
  -h, --help   help for canned-responses
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
* [wootctl canned-responses create](wootctl_canned-responses_create.md)	 - Create a canned-response
* [wootctl canned-responses delete](wootctl_canned-responses_delete.md)	 - Delete a canned-response
* [wootctl canned-responses list](wootctl_canned-responses_list.md)	 - List canned-responses
* [wootctl canned-responses update](wootctl_canned-responses_update.md)	 - Update a canned-response

