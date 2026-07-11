## wootctl agents create

Create a agent

```
wootctl agents create [flags]
```

### Examples

```
  wootctl agents create --data '{...}'
  wootctl agents create -d @payload.json
```

### Options

```
      --auto-offline          mark offline automatically when away
      --availability string   available | busy | offline
  -d, --data string           JSON body: inline, @file, or - for stdin
      --email string          agent's email (an invite is sent)
  -h, --help                  help for create
      --name string           agent's full name
      --role string           agent | administrator
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

* [wootctl agents](wootctl_agents.md)	 - Manage the account's agents

