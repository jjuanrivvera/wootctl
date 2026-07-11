## wootctl agent guard

Generate agent-safety config that blocks destructive wootctl operations

### Synopsis

Classify every API command (read / write / irreversible) from the live command tree
and emit host safety config: irreversible operations (deletes, contacts merge,
remove-members, delete-hook) are hard-blocked, ordinary writes require approval, and
reads are allowed. Cobra alias paths are covered too — "conv delete" hits the same
rails as "conversations delete".

For claude-code the output also includes a PreToolUse hook script
(.claude/hooks/wootctl-guard.sh): it strips quote/backslash obfuscation, matches blocked
subcommand paths at the command position even for path-invoked binaries (./bin/wootctl,
/usr/local/bin/wootctl), and gates the raw "wootctl api <METHOD> <PATH>" escape hatch at
the METHOD position — only GET/HEAD/OPTIONS pass; POST/PUT/PATCH/DELETE are denied
case-insensitively, while a GET whose path merely contains "delete" stays allowed.
"wootctl alias set" is denied so an agent cannot mint a new shorthand for a blocked
command.

MCP-only operation is the hard guarantee; the Bash rails are best-effort — the hook
defeats quoting tricks and path prefixes, but not variable indirection
(a=delete; wootctl labels $a 1) or shell aliases. Conservative false positives are
accepted: a line that merely QUOTES a blocked command (rg "wootctl labels delete") is
denied.

```
wootctl agent guard --host <claude-code|codex|opencode> [flags]
```

### Examples

```
  wootctl agent guard --host claude-code
  wootctl agent guard --host claude-code --write          # write the files into .claude/
  wootctl agent guard --host codex --out ~/.codex/config.toml
  wootctl agent guard --host opencode --all-writes
```

### Options

```
      --all-writes    also hard-block ordinary writes, not just irreversible ops
  -h, --help          help for guard
      --host string   target agent host: claude-code|codex|opencode (required)
      --out string    write to this file instead of stdout
      --write         claude-code only: write hook + settings fragment under .claude/ (never overwrites)
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

* [wootctl agent](wootctl_agent.md)	 - AI-agent integration helpers

