# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.2] - 2026-07-12

### Security
- Bump the Go toolchain to **go1.25.12**, clearing the reachable standard-library advisories
  (crypto/tls GO-2026-5856, crypto/x509, net/http, net/textproto) that govulncheck flagged.

## [0.3.1] - 2026-07-12

### Fixed
- The hidden secret prompt (`init` / `auth`) now reads in **raw mode** instead of
  `term.ReadPassword`'s canonical mode (capped at MAX_CANON, 1024 bytes on macOS), so a long
  pasted token no longer hangs the prompt until Ctrl-C. Bracketed-paste markers are still
  stripped as a defensive guard.

## [0.3.0] - 2026-07-11

### Changed

- **Renamed the tool from `cwctl` to `wootctl`.** `cwctl` collided with Chatwoot's own
  official server-management CLI (ships with Chatwoot v2.7.0+, installs to
  `/usr/local/bin/cwctl`). The binary, Go module path, config dir (`~/.wootctl-cli`), env
  vars (`WOOTCTL_*`), keyring service, Homebrew cask/tap, Scoop bucket, and Docker image all
  move to `wootctl`. Re-run `wootctl init` to re-store your token under the new name.

### Added

- **`chatwoot` drop-in step in `init`.** wootctl's command surface is a superset of the
  official `chatwoot` CLI, so `init` offers to symlink `chatwoot` → `wootctl` — every official
  command works here, plus multi-profile, keyring fallback, and backup/sync.
- **One-line install script** (`install.sh`) for macOS/Linux, checksum-verified:
  `curl -fsSL https://raw.githubusercontent.com/jjuanrivvera/wootctl/main/install.sh | sh`.
- **OpenAPI contract testing** (`commands/contract_test.go`): every request the CLI builds
  for the application API is validated against Chatwoot's own OpenAPI spec (method, path, and
  request-body schema), vendored at `internal/api/testdata/application_swagger.json`. A
  self-check test proves the harness can fail, so the gate can't silently green-light drift.

### Security

- Strip ANSI escape sequences and control characters from API-returned text (contact names,
  labels, error bodies) before printing to the human table and error output — closes a
  terminal-escape-injection vector where a crafted value could rewrite the terminal title or
  move the cursor. Machine formats (json/yaml/csv) stay faithful.

## [0.2.0] - 2026-07-10

### Added

- **Beyond-the-API layer** (config-as-code + multi-instance):
  - `wootctl backup --dir <dir>` dumps account config (labels, canned-responses,
    custom-attributes, custom-filters, automation-rules, teams, webhooks, agent-bots) to
    a git-friendly directory of YAML files, keeping only writable fields.
  - `wootctl restore --dir <dir>` reconciles a backup into the account (create/update/skip;
    `--prune` removes drift), matching by natural key with duplicate-detect-and-skip.
  - `wootctl sync --to <profile>` promotes config between instances — the multi-instance
    payoff the single-instance official CLI can't offer.
  - `restore`/`sync` are classified destructive (they can `--prune`), so `agent guard`
    hard-blocks them; `backup` stays allowed. All three preview with `--dry-run`.

## [0.1.2] - 2026-07-10

### Fixed

- `update` no longer requires create-only fields: a partial PATCH (e.g. `labels update
  <id> --description X`) no longer fails demanding `--title`. Resources with explicit
  `UpdateFields` keep their own required flags. Found during live testing.
- Decode Chatwoot's double-nested envelopes: single records wrapped as
  `{"payload":{"<resource>":{…}}}` (webhook create) and lists as
  `{"payload":{"webhooks":[…]}}` are now unwrapped, so `webhooks create`/`list` render
  the record(s) instead of `null`/an error. Found during live testing.

## [0.1.1] - 2026-07-10

### Security

- Upgrade `modelcontextprotocol/go-sdk` to v1.6.1, past GO-2026-4773 (cross-site tool
  execution in the SDK's HTTP transport, reachable via `wootctl mcp stream`) and
  GO-2026-5771 (DNS-rebinding protection default). stdio MCP (`wootctl mcp start`) was
  unaffected.

### Fixed

- The pre-commit hook no longer aborts silently on commits that stage only
  `go.mod`/`go.sum`.

## [0.1.0] - 2026-07-10

### Added

- Full Chatwoot API surface: 144/144 documented operations across the application
  API (`/api/v1`, `/api/v2` reports), the platform API (`/platform/api/v1`), and the
  public client API (`/public/api/v1`), enforced by spec-check + spec-completeness
  gates against a manifest derived from Chatwoot's own OpenAPI spec.
- Named profiles (instance + account + tokens) with `--profile` / `WOOTCTL_PROFILE`;
  tokens in the OS keyring with an AES-256-GCM encrypted-file fallback for headless
  hosts (`WOOTCTL_KEYRING_PASSWORD`).
- Output formats: table (TTY-colored), json, yaml, csv (formula-injection-safe), id;
  `--columns`, `--filter`, `--limit`, `--all` pagination, `--jq` (gojq).
- `--dry-run` prints the equivalent curl with the token redacted; multipart uploads
  (`messages create --attachment`) included.
- Resilience: full-jitter retries on idempotent methods only, `Retry-After` honored,
  fixed-RPS rate limiting with halve-on-429 and gradual restore, Ctrl-C cancels
  in-flight work.
- Meta commands: `auth login/logout/status`, `config`, `init`, `doctor`, `completion`,
  `alias`, `api` (raw escape hatch), `version --check`.
- AI agent surface: `wootctl mcp` (ophis MCP server, 125 annotated tools, secret flags
  and setup commands excluded) and `wootctl agent guard` for claude-code / codex /
  opencode with a hardened PreToolUse hook (alias-path enumeration, de-obfuscation,
  METHOD-gated raw api, strict no-jq fallback).

[Unreleased]: https://github.com/jjuanrivvera/wootctl/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/jjuanrivvera/wootctl/compare/v0.1.2...v0.2.0
[0.1.2]: https://github.com/jjuanrivvera/wootctl/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/jjuanrivvera/wootctl/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/jjuanrivvera/wootctl/releases/tag/v0.1.0
