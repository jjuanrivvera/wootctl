# Backup, restore & sync

Beyond wrapping the API, cwctl treats your account **config** as something you can version in
git and move between instances. This is the multi-instance payoff a single-instance tool
can't offer.

## What's portable

Account **config** — never live data:

| Portable (config) | Not portable (data) |
|---|---|
| labels, canned-responses, custom-attributes, custom-filters, automation-rules, teams, webhooks, agent-bots | conversations, contacts, messages |

Inboxes and help-center portals are intentionally excluded (channel credentials / no delete
endpoint).

## backup

```bash
cwctl backup --dir ./chatwoot-config
cwctl backup --dir ./cfg --only labels,canned-responses
```

Writes one YAML file per kind, keeping **only the writable fields** — `id`, timestamps, and
other server-managed noise are dropped, so the output is stable and diffable. Commit the
directory to git and you have versioned, reviewable Chatwoot config.

## restore

```bash
cwctl restore --dir ./chatwoot-config --dry-run     # preview — always do this first
cwctl restore --dir ./chatwoot-config               # create missing, update changed, skip unchanged
cwctl restore --dir ./chatwoot-config --only labels --prune   # also delete drift
```

Reconciles the directory into the active profile's account:

- **create** resources present in the backup but missing live,
- **update** resources whose writable fields differ,
- **skip** unchanged ones,
- with `--prune`, **delete** live resources absent from the backup.

## sync

```bash
cwctl sync --to production --dry-run
cwctl sync --to production --only canned-responses,labels
cwctl --profile staging sync --to production --prune
cwctl sync --from staging --to production          # explicit source
```

Same reconcile, but the desired state is another profile's **live** account instead of a
directory. Promote canned responses and automation from staging to production, or keep two
support instances aligned. The target's own base URL, account id, and token are used (the
active profile's `--base-url`/`--account-id` flags never leak across instances).

## Matching & safety

- **Natural-key matching.** Chatwoot exposes no stable cross-instance id, so resources match
  by a natural key: `title` (labels), `short_code` (canned responses), `name` (teams,
  filters, rules, bots), `url` (webhooks), `attribute_key` (custom attributes).
- **Unchanged = writable fields equal.** Only writable fields are compared, canonicalized as
  JSON, so field order and equivalent encodings never create phantom updates.
- **Ambiguous keys are never touched.** If two resources in an account share a key, cwctl
  warns and skips them — it will never update or prune an arbitrary one of two same-named
  resources.
- **Destructive by classification.** `restore` and `sync` can delete with `--prune`, so
  `cwctl agent guard` hard-blocks them for AI agents. `backup` is read-only and stays allowed.
- **Always `--dry-run` first.** Both print a per-resource plan (`+ create`, `~ update`,
  `- prune`) and a summary before you commit to the real thing.
