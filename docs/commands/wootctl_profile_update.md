## wootctl profile update

Update your own profile (name, signature, password, …)

```
wootctl profile update [flags]
```

### Examples

```
  wootctl profile update --display-name "Juan R."
  wootctl profile update --message-signature "— Juan, Soporte"
```

### Options

```
      --current-password string        current password (required to change the password)
  -d, --data string                    JSON body: inline, @file, or - for stdin
      --display-name string            name agents see
      --email string                   login email
  -h, --help                           help for update
      --message-signature string       signature appended to outgoing replies
      --name string                    full name
      --password string                new password
      --password-confirmation string   new password again
      --phone-number string            phone number
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

* [wootctl profile](wootctl_profile.md)	 - Read and update your own user profile

