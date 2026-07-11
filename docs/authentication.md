# Authentication & profiles

## Token classes

Chatwoot authenticates everything with one header (`api_access_token`) but three
token classes exist:

| Class | Used by | Stored as |
|---|---|---|
| user token | application API (`wootctl <resource> …`) | keyring key `<profile>` |
| platform app token | `wootctl platform …` (self-hosted provisioning) | keyring key `<profile>/platform` |
| none | `wootctl client …` (public contact-facing API), `csat` | — |

wootctl picks the class from the path automatically — including for the raw
`wootctl api METHOD PATH` escape hatch.

```bash
wootctl auth login                                   # user token (verified live)
wootctl auth login --platform-token <token>          # add a platform token
wootctl auth status                                  # identity + validity + backend
wootctl auth logout                                  # remove both tokens
```

## Where tokens live

1. **OS keyring** — macOS Keychain, Linux Secret Service, Windows Credential Manager.
2. **Encrypted file fallback** — when no keyring is reachable (VPS, container), an
   AES-256-GCM `credentials.enc` under the config dir. Set `WOOTCTL_KEYRING_PASSWORD`
   to key it (scrypt-derived); without it the key is host-bound, which resists casual
   copying but is not a hard boundary.
3. **Env override** — `WOOTCTL_API_KEY` beats everything (CI).

Non-secret settings live in `~/.wootctl-cli/config.yaml`
(or `$XDG_CONFIG_HOME/wootctl/config.yaml`), written 0600.

## Profiles

A profile = base URL + account id + tokens. Use them for staging vs production, or
two companies on one Cloud login:

```bash
wootctl --profile staging auth login
wootctl config list-profiles
wootctl config use staging
wootctl --profile prod conversations list     # one-off override
WOOTCTL_PROFILE=staging wootctl doctor          # per-shell
```

`wootctl config set base_url|account_id|rps <value>` edits the active profile;
`--account-id` overrides the account for a single invocation.

## Diagnosing

```bash
wootctl doctor          # config, profile, base URL, token presence + validity, account
wootctl doctor --json   # for scripts; exits non-zero on any failure
```
