## wootctl portals create-category

Add a category to a portal

```
wootctl portals create-category <portal-id> [flags]
```

### Examples

```
  wootctl portals create-category 1 --name "Facturación" --locale es
```

### Options

```
      --associated-category-id int   linked category id (another locale)
  -d, --data string                  JSON body: inline, @file, or - for stdin
      --description string           category description
  -h, --help                         help for create-category
      --icon string                  category icon
      --locale string                category locale, e.g. en, es
      --name string                  category name
      --parent-category-id int       parent category id
      --position int                 sort position
      --slug string                  category slug
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

* [wootctl portals](wootctl_portals.md)	 - Manage help-center portals, articles, and categories

