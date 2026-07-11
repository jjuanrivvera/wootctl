## wootctl custom-attributes

Manage custom attribute definitions

### Options

```
  -h, --help   help for custom-attributes
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
* [wootctl custom-attributes create](wootctl_custom-attributes_create.md)	 - Create a custom-attribute
* [wootctl custom-attributes delete](wootctl_custom-attributes_delete.md)	 - Delete a custom-attribute
* [wootctl custom-attributes get](wootctl_custom-attributes_get.md)	 - Get a single custom-attribute
* [wootctl custom-attributes list](wootctl_custom-attributes_list.md)	 - List custom-attributes
* [wootctl custom-attributes update](wootctl_custom-attributes_update.md)	 - Update a custom-attribute

