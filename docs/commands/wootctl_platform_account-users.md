## wootctl platform account-users

Manage the users of an account (platform app token)

### Options

```
  -h, --help   help for account-users
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

* [wootctl platform](wootctl_platform.md)	 - Platform API (accounts, users, agent bots) — needs a platform app token
* [wootctl platform account-users create](wootctl_platform_account-users_create.md)	 - Add a user to an account
* [wootctl platform account-users delete](wootctl_platform_account-users_delete.md)	 - Remove a user from an account
* [wootctl platform account-users list](wootctl_platform_account-users_list.md)	 - List an account's users

