## wootctl platform accounts create

Create a account

```
wootctl platform accounts create [flags]
```

### Examples

```
  wootctl accounts create --data '{...}'
  wootctl accounts create -d @payload.json
```

### Options

```
      --custom-attributes string   custom attributes object (JSON)
  -d, --data string                JSON body: inline, @file, or - for stdin
      --domain string              account domain
  -h, --help                       help for create
      --limits string              usage limits, e.g. '{"agents":10,"inboxes":5}' (JSON)
      --locale string              default language, e.g. en, es
      --name string                account name
      --status string              active | suspended
      --support-email string       support email
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

* [wootctl platform accounts](wootctl_platform_accounts.md)	 - Provision accounts (platform app token)

