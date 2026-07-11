# Getting started

## 1. Connect

```bash
wootctl auth login
```

You are asked for:

- **Base URL** — your instance root (`https://app.chatwoot.com`, or your self-hosted
  domain).
- **api_access_token** — from Chatwoot: Profile Settings → Access Token. Input is
  hidden and the token goes to your OS keyring, never to a file in plaintext.

The token is verified against `GET /api/v1/profile` and the account is captured (a
multi-account token prompts you to pick one). `wootctl init` runs the same flow plus a
smoke check.

## 2. Look around

```bash
wootctl auth status            # who am I, which account, is the token valid
wootctl inboxes list
wootctl conversations meta     # open/unassigned counts
wootctl conversations list --status open --assignee-type me
```

## 3. Work a conversation

```bash
wootctl messages list 42
wootctl messages create 42 --content "On it."
wootctl messages create 42 --content "internal note" --private
wootctl messages create 42 --content "see attached" --attachment ./invoice.pdf
wootctl conversations assign 42 --assignee-id 7
wootctl conversations toggle-status 42 --status resolved
```

## 4. Script it

```bash
wootctl contacts search --q ana -o json | jq '.[0].id'
wootctl conversations list --all -o csv > backlog.csv
wootctl labels list -o id | xargs -n1 -I{} wootctl labels get {}
```

Every command honors `--dry-run` (prints the exact curl, token redacted), so you can
always see what would happen before it does.

## Environment variables

| Variable | Effect |
|---|---|
| `WOOTCTL_API_KEY` | token override (CI) — beats the keyring |
| `WOOTCTL_PLATFORM_TOKEN` | platform app token override |
| `WOOTCTL_BASE_URL` / `WOOTCTL_ACCOUNT_ID` | instance/account override |
| `WOOTCTL_PROFILE` | profile selection per shell |
| `WOOTCTL_KEYRING_PASSWORD` | key for the encrypted-file fallback on headless hosts |
| `NO_COLOR` | disable table colors |
