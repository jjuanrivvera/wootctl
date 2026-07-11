## wootctl portals

Manage help-center portals, articles, and categories

### Options

```
  -h, --help   help for portals
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
* [wootctl portals create](wootctl_portals_create.md)	 - Create a portal
* [wootctl portals create-article](wootctl_portals_create-article.md)	 - Add an article to a portal
* [wootctl portals create-category](wootctl_portals_create-category.md)	 - Add a category to a portal
* [wootctl portals list](wootctl_portals_list.md)	 - List portals
* [wootctl portals update](wootctl_portals_update.md)	 - Update a portal

