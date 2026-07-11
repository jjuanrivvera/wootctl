## wootctl integrations create-hook

Create an integration hook

```
wootctl integrations create-hook [flags]
```

### Examples

```
  wootctl integrations create-hook --app-id dialogflow --settings '{"project_id":"x","credentials":{}}' --inbox-id 3
```

### Options

```
      --app-id string     integration app id (e.g. slack, dialogflow, dyte)
  -d, --data string       JSON body: inline, @file, or - for stdin
  -h, --help              help for create-hook
      --inbox-id int      inbox to attach the hook to (inbox-scoped apps)
      --settings string   app-specific settings object (JSON)
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

* [wootctl integrations](wootctl_integrations.md)	 - List integration apps and manage integration hooks

