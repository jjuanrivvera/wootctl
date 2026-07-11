## wootctl auth

Manage Chatwoot tokens and verify authentication

### Synopsis

Capture, verify, and remove the tokens for a profile. Tokens are stored in your OS
keyring (with an encrypted-file fallback on headless hosts), never in the config file.

A profile can hold two token kinds:
  user      the api_access_token from your Chatwoot profile page — drives the
            application API (conversations, contacts, reports, …)
  platform  a platform app token (self-hosted; from a super-admin platform app) —
            drives the 'wootctl platform …' commands

### Options

```
  -h, --help   help for auth
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
* [wootctl auth login](wootctl_auth_login.md)	 - Store a Chatwoot token and verify it
* [wootctl auth logout](wootctl_auth_logout.md)	 - Remove stored tokens for the active profile
* [wootctl auth status](wootctl_auth_status.md)	 - Show the active profile, base URL, identity, and token validity

