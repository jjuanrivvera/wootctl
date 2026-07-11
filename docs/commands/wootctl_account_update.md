## wootctl account update

Update the current account

```
wootctl account update [flags]
```

### Examples

```
  wootctl account update --name "Soporte Invitas"
  wootctl account update -d '{"locale":"es","timezone":"America/Bogota"}'
```

### Options

```
      --auto-resolve-after int   minutes of inactivity before auto-resolve
      --company-size string      company size bucket
  -d, --data string              JSON body: inline, @file, or - for stdin
      --domain string            account domain
  -h, --help                     help for update
      --industry string          industry label
      --locale string            default language, e.g. en, es
      --name string              account name
      --support-email string     support email address
      --timezone string          IANA timezone, e.g. America/Bogota
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

* [wootctl account](wootctl_account.md)	 - Read and update the current account

