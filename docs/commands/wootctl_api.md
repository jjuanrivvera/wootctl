## wootctl api

Send a raw authenticated request (escape hatch)

### Synopsis

Call any Chatwoot endpoint directly. The path is relative to the instance root
(e.g. api/v1/accounts/1/labels, platform/api/v1/users, public/api/v1/inboxes/…);
the credential class is chosen from the path exactly like first-class commands
(platform/* uses the platform token; public/* sends none).

This is the documented escape hatch for anything wootctl does not wrap as a
first-class command. It honors --dry-run, -o/--output, and --jq like every other
command. Non-GET methods are never auto-retried.

```
wootctl api <METHOD> <PATH> [-d body] [-q key=value ...] [flags]
```

### Examples

```
  wootctl api GET api/v1/accounts/1/conversations/42
  wootctl api GET api/v2/accounts/1/reports/summary -q since=1719800000 -q until=1722400000
  wootctl api POST api/v1/accounts/1/labels -d '{"title":"vip","color":"#0055ff"}'
  wootctl api DELETE api/v1/accounts/1/labels/9 --dry-run
```

### Options

```
  -d, --data string         JSON body: inline, @file, or - for stdin
  -h, --help                help for api
  -q, --query stringArray   query parameter key=value (repeatable)
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

