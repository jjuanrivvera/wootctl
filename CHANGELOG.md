# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
  execution in the SDK's HTTP transport, reachable via `cwctl mcp stream`) and
  GO-2026-5771 (DNS-rebinding protection default). stdio MCP (`cwctl mcp start`) was
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
- Named profiles (instance + account + tokens) with `--profile` / `CWCTL_PROFILE`;
  tokens in the OS keyring with an AES-256-GCM encrypted-file fallback for headless
  hosts (`CWCTL_KEYRING_PASSWORD`).
- Output formats: table (TTY-colored), json, yaml, csv (formula-injection-safe), id;
  `--columns`, `--filter`, `--limit`, `--all` pagination, `--jq` (gojq).
- `--dry-run` prints the equivalent curl with the token redacted; multipart uploads
  (`messages create --attachment`) included.
- Resilience: full-jitter retries on idempotent methods only, `Retry-After` honored,
  fixed-RPS rate limiting with halve-on-429 and gradual restore, Ctrl-C cancels
  in-flight work.
- Meta commands: `auth login/logout/status`, `config`, `init`, `doctor`, `completion`,
  `alias`, `api` (raw escape hatch), `version --check`.
- AI agent surface: `cwctl mcp` (ophis MCP server, 125 annotated tools, secret flags
  and setup commands excluded) and `cwctl agent guard` for claude-code / codex /
  opencode with a hardened PreToolUse hook (alias-path enumeration, de-obfuscation,
  METHOD-gated raw api, strict no-jq fallback).

[Unreleased]: https://github.com/jjuanrivvera/cwctl/compare/v0.1.2...HEAD
[0.1.2]: https://github.com/jjuanrivvera/cwctl/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/jjuanrivvera/cwctl/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/jjuanrivvera/cwctl/releases/tag/v0.1.0
