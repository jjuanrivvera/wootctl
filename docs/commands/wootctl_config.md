## wootctl config

Inspect and edit wootctl configuration

### Synopsis

View the config file, switch profiles, and set per-profile options. Secrets live in the keyring and are never shown here.

### Options

```
  -h, --help   help for config
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
* [wootctl config list-profiles](wootctl_config_list-profiles.md)	 - List configured profiles
* [wootctl config path](wootctl_config_path.md)	 - Print the config file path
* [wootctl config set](wootctl_config_set.md)	 - Set a per-profile option (keys: base_url, account_id, rps)
* [wootctl config use](wootctl_config_use.md)	 - Switch the active profile
* [wootctl config view](wootctl_config_view.md)	 - Show the current configuration (secrets redacted)

