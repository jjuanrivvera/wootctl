## wootctl auth login

Store a Chatwoot token and verify it

### Synopsis

Capture your instance URL and api_access_token, verify them against GET /api/v1/profile,
pick the account to scope commands to, and save everything under the active profile.

Find the token in Chatwoot: Profile Settings → Access Token.

```
wootctl auth login [flags]
```

### Examples

```
  wootctl auth login                                  # interactive (hidden token input)
  wootctl auth login --base-url https://app.chatwoot.com --api-key <token>
  wootctl --profile staging auth login                # save a second instance
  wootctl auth login --platform-token <token>         # also store a platform app token
```

### Options

```
      --account string                        account id to scope commands to (alias of --account-id)
      --api-key string                        api_access_token (omit to be prompted with hidden input)
  -h, --help                                  help for login
      --no-verify                             skip the /api/v1/profile verification call
      --platform-token wootctl platform …   platform app token for wootctl platform … (optional)
      --url string                            instance base URL (alias of --base-url)
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

* [wootctl auth](wootctl_auth.md)	 - Manage Chatwoot tokens and verify authentication

