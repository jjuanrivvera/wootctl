## wootctl contacts update

Update a contact

```
wootctl contacts update <id> [flags]
```

### Examples

```
  wootctl contacts update 42 --data '{...}'
```

### Options

```
      --additional-attributes string   additional attributes object (JSON)
      --avatar-url string              URL of an avatar image
      --blocked                        block/unblock the contact
      --custom-attributes string       custom attributes object (JSON)
  -d, --data string                    JSON body: inline, @file, or - for stdin
      --email string                   email address
  -h, --help                           help for update
      --identifier string              external unique identifier
      --name string                    contact name
      --phone-number string            phone in E.164
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

* [wootctl contacts](wootctl_contacts.md)	 - Manage contacts

