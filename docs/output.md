# Output & filtering

One renderer serves every command, driven by JSON normalization — so everything below
works identically on conversations, contacts, reports, and the raw `api` hatch.

## Formats

```bash
wootctl labels list                 # table (default; colored on a TTY, NO_COLOR honored)
wootctl labels list -o json         # full fidelity — unknown API fields are never dropped
wootctl labels list -o yaml
wootctl labels list -o csv          # cells sanitized against =SUM() formula injection
wootctl labels list -o id           # one id per line, xargs-ready
```

Tables cap at ~10 auto-detected columns and truncate wide cells with `…` (a stderr
note tells you when; `-o json` always has everything). Notes go to stderr so stdout
stays pipe-clean; `--quiet` silences them.

## Choosing fields

```bash
wootctl agents list --columns name,email          # exact columns, in that order
wootctl conversations list --jq '.[0].meta.sender.name'   # gojq over the response
```

## Filtering & pagination

```bash
wootctl conversations list --status open --labels vip     # server-side params
wootctl agents list --filter role=agent                   # client-side field=value
wootctl contacts list --page 3                            # one server page
wootctl contacts list --all                               # walk every page
wootctl contacts list --all --limit 50                    # stop the output at 50
```

Chatwoot fixes the page size server-side (25 on paginated lists), so `--limit` trims
client-side and `--page` selects a server page. `--all` stops on an empty page, an
identical-page echo (endpoints that ignore `page`), or the advertised total.

## Dry-run

```bash
$ wootctl webhooks create --url https://example.com/hook --dry-run
curl -X POST 'https://…/api/v1/accounts/1/webhooks' -H 'api_access_token: REDACTED' \
  -H 'Content-Type: application/json' -d '{"url":"https://example.com/hook"}'
```

Copy-paste runnable (add your token), and nothing is sent. `--show-token` reveals the
real token when you actually want it.
