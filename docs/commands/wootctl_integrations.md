## wootctl integrations

List integration apps and manage integration hooks

### Options

```
  -h, --help   help for integrations
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
* [wootctl integrations apps](wootctl_integrations_apps.md)	 - List available integration apps and their status
* [wootctl integrations create-hook](wootctl_integrations_create-hook.md)	 - Create an integration hook
* [wootctl integrations delete-hook](wootctl_integrations_delete-hook.md)	 - Delete an integration hook
* [wootctl integrations update-hook](wootctl_integrations_update-hook.md)	 - Update an integration hook's settings

