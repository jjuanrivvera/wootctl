# wootctl

A fast, scriptable CLI for the full [Chatwoot](https://www.chatwoot.com) API.

`wootctl` wraps every operation Chatwoot documents — 144/144 across the application,
platform, and public client APIs — with named profiles for working across several
instances, OS-keyring token storage (with an encrypted-file fallback for headless
hosts), table/json/yaml/csv output, and an MCP server so AI agents can drive it safely.

```console
$ wootctl conversations list --status open
ID  INBOX_ID  STATUS  PRIORITY
42  3         open    high
57  1         open

$ wootctl messages create 42 --content "On it — checking now."
$ wootctl conversations toggle-status 42 --status resolved
```

## Install

```bash
# Install script (macOS / Linux) — downloads the release binary, verifies its checksum
curl -fsSL https://raw.githubusercontent.com/jjuanrivvera/wootctl/main/install.sh | sh

# Homebrew (macOS / Linux)
brew install jjuanrivvera/wootctl/wootctl-cli

# Scoop (Windows)
scoop bucket add wootctl https://github.com/jjuanrivvera/scoop-wootctl
scoop install wootctl

# Go
go install github.com/jjuanrivvera/wootctl/cmd/wootctl@latest
```

deb/rpm/apk packages and prebuilt archives are on the
[releases page](https://github.com/jjuanrivvera/wootctl/releases). Releases are
cosign-signed and ship SBOMs.

## Setup

```bash
wootctl auth login
```

You are prompted for your instance URL and `api_access_token` (Chatwoot → Profile
Settings → Access Token; input is hidden). The token is verified live, the account is
selected, and everything is saved: the token in your OS keyring, the rest in
`~/.wootctl-cli/config.yaml`.

Headless host with no keyring? Set `WOOTCTL_KEYRING_PASSWORD` and tokens go to an
AES-256-GCM encrypted file instead. CI? `WOOTCTL_API_KEY` overrides everything.

### Profiles

Work across several Chatwoot instances (or accounts) by saving each as a profile:

```bash
wootctl --profile staging auth login     # save a second instance
wootctl config list-profiles             # see them
wootctl config use staging               # switch the default
wootctl --profile staging convs list     # one-off override
```

A profile bundles base URL + account id + tokens, so switching profiles switches
everything at once. `WOOTCTL_PROFILE` selects one per shell.

## The surface

| Group | Commands |
|---|---|
| Conversations | `list · meta · create · get · update · filter · labels · add-labels · assign · toggle-status · toggle-priority · toggle-typing · set-custom-attributes · reporting-events` |
| Messages | `list · create (with --attachment multipart) · delete` |
| Contacts | `list · create · get · update · delete · search · filter · merge · conversations · contactable-inboxes · create-contact-inbox · labels · add-labels` |
| Inboxes | `list · create · get · update · agent-bot · set-agent-bot · members · add/update/remove-members` |
| Teams | `list · create · get · update · delete · members · add/update/remove-members` |
| Reports | `overview · summary · account/agent-conversations · first-response-time-distribution · inbox-label-matrix · outgoing-messages-count · agent/channel/inbox/team-summary · events` |
| Also | `account · agents · agent-bots · audit-logs · automation-rules · canned-responses · custom-attributes · custom-filters · integrations · labels · portals · profile · webhooks · csat` |
| `platform …` | accounts, account-users, agent-bots, users (+ sso-link) — platform app token |
| `client …` | the public, unauthenticated contact-facing API (inbox, contacts, conversations, messages) |

Anything not wrapped (there is nothing today) is reachable via the escape hatch:

```bash
wootctl api GET api/v2/accounts/1/reports/summary -q since=1780272000
```

The full command reference lives in [docs/commands](docs/commands/wootctl.md).

## Output and scripting

```bash
wootctl contacts search --q ana -o json | jq '.[0].email'
wootctl labels list -o id | xargs -n1 wootctl labels get
wootctl conversations list --all --filter status=open -o csv > open.csv
wootctl agents list --columns name,email
wootctl conversations list --jq '.[3].meta.sender.name'
```

`table` (default, TTY-colored, `NO_COLOR` honored), `json`, `yaml`, `csv`
(formula-injection-sanitized), and `id` (one per line, pipe-friendly). `--all` walks
every page. `--dry-run` prints the exact curl (token redacted) instead of calling:

```console
$ wootctl labels create --title vip --dry-run
curl -X POST 'https://…/api/v1/accounts/1/labels' -H 'api_access_token: REDACTED' \
  -H 'Content-Type: application/json' -d '{"title":"vip"}'
```

## AI agents

`wootctl` is agent-ready in both directions:

```bash
wootctl mcp claude enable    # expose 125 annotated MCP tools (reads/writes/destructive)
wootctl agent guard --host claude-code --write   # install guardrails for agents using Bash
```

The MCP surface excludes auth/config/profile switching and every secret flag. The
guard hard-blocks irreversible operations (deletes, `contacts merge`) — including
every cobra alias path and the raw `api` hatch for write methods — with a PreToolUse
hook that survives quoting tricks and path-prefixed binaries. `codex` and `opencode`
targets are included.

## Backup, restore & sync (beyond the API)

Past the raw endpoints, wootctl treats your account **config** as something you can version and
move between instances — labels, canned responses, custom attributes, custom filters,
automation rules, teams, webhooks, and agent bots (never conversations/contacts/messages,
which are live data).

```bash
# dump config to a git-friendly directory (one YAML file per kind, writable fields only)
wootctl backup --dir ./chatwoot-config
git -C ./chatwoot-config add -A && git -C ./chatwoot-config commit -m "chatwoot config"

# reconcile a backup back into the account: create missing, update changed, skip unchanged
wootctl restore --dir ./chatwoot-config --dry-run     # always preview first
wootctl restore --dir ./chatwoot-config
wootctl restore --dir ./chatwoot-config --only labels --prune   # also delete drift

# the multi-instance payoff the official CLI can't do: promote config between instances
wootctl sync --to production --dry-run
wootctl sync --to production --only canned-responses,labels
wootctl --profile staging sync --to production --prune
```

Matching is by natural key (title, short_code, name, url, attribute_key); "unchanged" compares
only writable fields, so `id`/timestamps never cause phantom updates. A key that appears twice
in an account is skipped and never pruned — wootctl won't act on an ambiguous match. `restore`
and `sync` are classified destructive, so `wootctl agent guard` hard-blocks them for AI agents.

## vs the official `chatwoot` CLI

The official [Chatwoot CLI](https://developers.chatwoot.com/cli) is good at what it
covers, and if you only drive one instance's conversations it may be all you need.
`wootctl` exists because we needed more:

- **Coverage**: wootctl wraps all 144 documented operations (application + platform +
  public client APIs), enforced by a spec-completeness gate in CI. The official CLI
  covers the core conversation workflow.
- **Multi-profile**: first-class named profiles across instances/accounts.
- **Headless hosts**: encrypted-file keyring fallback when no OS keyring exists (VPS,
  containers).
- **Agent surface**: MCP server with safety annotations + generated guardrails.
- **Beyond the API**: git-friendly `backup`/`restore` of account config and cross-instance
  `sync` — promote labels/canned-responses/automation between instances.
- **Scripting**: json/yaml/csv/id output, `--jq`, `--dry-run` curls, `--all` pagination.

Where the official CLI is genuinely better: it is first-party (tracks new endpoints the
moment they are documented by the same team), has upstream support, and its TUI-style
conversation flow is more polished for interactive triage. Both can coexist — the
binaries do not collide.

## Development

```bash
make build      # bin/wootctl
make check      # fmt + vet + lint + tests
make verify     # the full deterministic gate (spec-check, completeness, coverage, DoD)
```

The CLI surface is pinned to [api-manifest.json](api-manifest.json), derived from
Chatwoot's own OpenAPI spec; `make spec-check` fails if they diverge, and
`make spec-completeness` fails if the manifest covers less than the enumerated API.
Design decisions are recorded in [DECISIONS.md](DECISIONS.md).

## License

MIT
