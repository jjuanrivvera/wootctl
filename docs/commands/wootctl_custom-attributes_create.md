## wootctl custom-attributes create

Create a custom-attribute

```
wootctl custom-attributes create [flags]
```

### Examples

```
  wootctl custom-attributes create --data '{...}'
  wootctl custom-attributes create -d @payload.json
```

### Options

```
      --attribute-description string    description
      --attribute-display-name string   human name
      --attribute-display-type int      0 text, 1 number, 2 currency, 3 percent, 4 link, 5 date, 6 list, 7 checkbox
      --attribute-key string            machine key
      --attribute-model int             0 = conversation, 1 = contact
      --attribute-values strings        allowed values (list type)
  -d, --data string                     JSON body: inline, @file, or - for stdin
  -h, --help                            help for create
      --regex-cue string                hint shown when the regex rejects a value
      --regex-pattern string            validation regex (text type)
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

* [wootctl custom-attributes](wootctl_custom-attributes.md)	 - Manage custom attribute definitions

