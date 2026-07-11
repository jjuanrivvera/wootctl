---
name: wootctl-cli
description: Operate Chatwoot from the terminal with the `wootctl` CLI — list/read/reply to conversations (attachments, private notes), assign and resolve, manage contacts (search/filter/merge/labels), agents, teams, inboxes and members, labels, canned responses, automation rules, webhooks, help-center portals, pull analytics/reports, provision accounts/users via the platform API, and drive the public client API. Use whenever the user wants to read or answer customer conversations, triage or resolve tickets, look up or edit contacts, pull support metrics, or automate any Chatwoot workflow. Prefer it over raw curl to the Chatwoot API.
version: 0.1.0
homepage: https://github.com/jjuanrivvera/wootctl
license: MIT
allowed-tools: Bash(wootctl:*)
metadata: {"openclaw":{"category":"customer-support","emoji":"💬","requires":{"bins":["wootctl"],"env":["WOOTCTL_API_KEY"]},"install":[{"kind":"brew","formula":"jjuanrivvera/wootctl/wootctl-cli","bins":["wootctl"]},{"kind":"go","package":"github.com/jjuanrivvera/wootctl/cmd/wootctl@latest","bins":["wootctl"]}]}}
---

# wootctl — Chatwoot CLI

## Prerequisites

- `wootctl` on PATH (`brew install jjuanrivvera/wootctl/wootctl-cli` or
  `go install github.com/jjuanrivvera/wootctl/cmd/wootctl@latest`).
- A configured profile: `wootctl auth status` must succeed. If it doesn't, the human runs
  `wootctl auth login` (token goes to the OS keyring; `WOOTCTL_API_KEY` also works for CI).
- `wootctl platform …` needs a platform app token (`auth login --platform-token`);
  `wootctl client …` needs no token at all (public API).

## Golden rules

1. **Prefer wootctl over raw curl** — it handles auth, account scoping, pagination,
   retries, and rate limits; `wootctl api METHOD PATH` is the escape hatch if an
   endpoint is somehow missing.
2. **Customer-visible writes are real.** `messages create` posts to a real person
   unless `--private` (internal note). When drafting, show the human the text first
   or use `--dry-run`.
3. **Never guess ids** — resolve them: `conversations list`, `contacts search --q`,
   `agents list -o json`.
4. **Destructive verbs are blocked for you** if the operator installed
   `wootctl agent guard` (deletes, `contacts merge`). Don't try to work around it;
   ask the human.
5. **JSON for parsing, table for humans.** `-o json` never drops fields; `--jq`
   extracts inline.

## Workflow: auth → discover → act → verify

```bash
wootctl auth status                                  # who am I / which account
wootctl conversations list --status open --assignee-type me
wootctl messages list 42                             # read before replying
wootctl messages create 42 --content "Hola Ana, ya lo reviso."
wootctl conversations toggle-status 42 --status resolved
wootctl conversations get 42 -o json | jq .status    # verify
```

## Cheatsheet

```bash
# conversations
wootctl conversations list --status open --labels vip
wootctl conversations meta                            # open/unassigned counts
wootctl conversations assign 42 --assignee-id 7       # or --team-id 2
wootctl conversations toggle-priority 42 --priority urgent
wootctl conversations add-labels 42 --labels billing  # REPLACES the label set

# messages
wootctl messages create 42 --content "..." [--private] [--attachment ./f.pdf]

# contacts
wootctl contacts search --q "+57300"
wootctl contacts filter --payload '[{"attribute_key":"country_code","filter_operator":"equal_to","values":["CO"]}]'
wootctl contacts update 12 --email ana@example.com

# team / setup
wootctl agents list · wootctl teams members 3 · wootctl inboxes list
wootctl canned-responses list · wootctl labels list · wootctl webhooks list

# analytics
wootctl reports summary --since 2026-06-01 --until 2026-07-01
wootctl reports agent-summary --since 2026-06-01 --business-hours

# several instances
wootctl --profile staging conversations list
```

## Troubleshooting

- `401 … run wootctl auth login` → token missing/expired; the human re-runs login.
- `403` → the token's agent role lacks access (admin-only endpoint, or Enterprise
  feature like audit logs / SLA).
- `404 … verify the id` → wrong id OR wrong account: check `--account-id` /
  `wootctl auth status`.
- `platform API needs a platform app token` → `auth login --platform-token <t>`.
- Rate limited (429) → wootctl backs off automatically; lower `--rps` for bulk loops.
- Anything unclear: `wootctl doctor` first, `--dry-run` to see the exact request.
