# AGENTS.md — working in the wootctl repo

`wootctl` is a command-line tool for the **Chatwoot API** (application + platform + public
client APIs), built to the cliwright standard (Go + Cobra + GoReleaser). This file orients
an AI agent (or human) contributing.

## The one rule that matters
**`make verify` is the gate.** A change is done only when `make verify` exits `0`. It runs
`make check` (fmt, vet, golangci-lint, tests) + `spec-check` (the built surface matches
`api-manifest.json`) + `spec-completeness` (the manifest wraps the enumerated API — 144/144
operations from Chatwoot's own OpenAPI spec) + `cover-check` (≥80% coverage) +
`dod-check.sh`. Run the full `make verify` for any change that touches the command surface
or a documented behavior — not just `make check`.

## Architecture (where things live)
- `internal/api/` — the generic client core: `api_access_token` auth, retry with full
  jitter, fixed-RPS rate limit with halve-on-429, dry-run curl, the generic `Resource[T]`
  (List/ListAll/Get/Create/Update/Delete/Action), page-based pagination, list-envelope
  normalization (`{data:{meta,payload}}` / `{payload:[…]}` / bare arrays), flexible JSON
  types, `APIError` with actionable hints. Written once; never copy-paste per resource.
- `commands/` — thin, declarative resource files. Adding a resource is a type + a `Client`
  accessor + one `registerResource(...)` in `init()` — **zero edits to shared code**. The
  generic builder stamps MCP read-only/write/destructive annotations.
- `internal/{config,auth,output,version}` — named profiles + manual precedence (no Viper),
  keyring token storage (user + optional platform token, encrypted-file fallback for
  headless hosts), the table/json/yaml/csv renderer, build metadata.
- `cmd/wootctl/main.go` — entry point: `signal.NotifyContext` (Ctrl-C cancels in-flight
  work) + alias expansion before cobra parses.

## Chatwoot specifics you must not re-derive
- Three API groups, one header: application (`/api/v1` + `/api/v2` reports, user token),
  platform (`/platform/api/v1`, platform app token — a separate keyring entry), and client
  (`/public/api/v1`, **unauthenticated**). All auth is the `api_access_token` header.
- Account-scoped paths get `/api/v1/accounts/{account_id}` prefixed from the profile;
  `--account-id` overrides per invocation.
- Pagination is `page=` with a server-fixed 25/page; no quota headers exist, so the limiter
  is fixed-RPS + halve-on-429 (`Retry-After` honored).
- `contacts update` and `profile update` are PUT; everything else updates via PATCH.

## House rules
- Comments explain **WHY**, not WHAT.
- Thread `cmd.Context()` everywhere; never `context.Background()` (it breaks Ctrl-C). Tests
  use `t.Context()`.
- Secrets live in the OS keyring — never in config-in-repo, code, or commit messages.
- Pin every ambiguous API assumption in `DECISIONS.md`; read it back, never silently
  re-decide.
- The resource set derives from the enumerated spec (`api-manifest.json`, 144 ops @
  `chatwoot/chatwoot@33dea837168b`), not hand-picking. Surface changes require updating the
  manifest AND passing both spec gates.
