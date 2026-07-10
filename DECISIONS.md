# DECISIONS.md — pinned assumptions (cliwright GOAL.md §11)

One line each: question → decision → why. The build loop reads this back every pass and
never silently re-decides.

1. **Enumeration source** → `chatwoot/chatwoot@33dea837168b swagger/swagger.json` (2026-06-02),
   144 operations over 89 paths → it is the machine spec AND the exact source that renders
   developers.chatwoot.com (chatwoot/docs `docs.json` embeds `swagger/tag_groups/*` from that repo),
   so docs and enumeration cannot diverge.
2. **Swagger duplicate** → `GET /accounts/{account_id}/conversations/{conversation_id}/messages`
   (opId `getConversationMessages`, tag "Conversation") is the same Rails endpoint as
   `messages list` listed a second time without the `/api/v1` prefix → counted as covered via the
   manifest's flat `methods[]` entry, not a second command.
3. **Resource pattern** → Pattern A (generic-core) → the API is uniform CRUD for the majority of
   resources (labels, teams, webhooks, canned-responses, custom-attributes, custom-filters,
   automation-rules, agent-bots, portals…); none of §11's Pattern-B triggers (per-resource
   include params, masquerade) exist. Non-CRUD verbs (toggles, filter, merge, members, reports)
   ride the generic `Action`/`Extra` hooks.
4. **profile_flag** → `profile` (playbook default) → Chatwoot's natural nouns collide with API
   concepts: an "account" is a path axis (`account_id`) and a resource; "instance" would read as
   base-url-only. A profile bundles base_url + account_id + tokens.
5. **Auth** → `api_access_token` header; per-profile **user token** in the OS keyring
   (service `cwctl`, user `<profile>`), optional **platform app token** as a second keyring entry
   (`<profile>/platform`) required only by `platform *` commands (403-style hint when absent).
   Agent-bot tokens are accepted via `CWCTL_API_KEY` for ad-hoc use but not stored as a class.
6. **Public (client) API** → unauthenticated by design (swagger `security: []`); `client *`
   commands take `--inbox <identifier>` / `--contact <source-id>` and skip the token entirely.
7. **Pagination** → `page=` query param; Chatwoot serves fixed 25/page on paginated lists
   (conversations, contacts, audit logs…); `--all` walks pages until a short/empty page;
   conversations/contacts `meta` counts are informative only (not used as a stop condition,
   filtered lists report totals inconsistently). Many list endpoints are unpaginated — the walker
   stops after one page identical-response guard.
8. **List envelopes** → `decodeList` normalizes `{data:{meta,payload}}` (conversations),
   `{payload:[…]}` (contacts, inboxes, labels…), `{data:[…]}`, and bare arrays (agents, teams…);
   single objects unwrap `{payload:{…}}`/`{data:{…}}` one level.
9. **Rate limiting** → Chatwoot exposes **no quota headers** (Rack::Attack server-side) → fixed
   RPS limiter (default 5 rps, `--rps`/config override) + halve-on-429 with gradual restore;
   honor `Retry-After` when present.
10. **Update verbs** → PATCH is the API default; `contacts update` and `profile update` are PUT
    (per spec) via the generic `UpdateMethod` knob.
11. **IDs** → numeric in Chatwoot JSON → flexible `ID` type (string|number in, string out) keeps
    table rendering and >2^53 safety.
12. **Reports** → v2 endpoints are read-only, query-param driven; `reports agent-conversations`
    maps to the swagger's trailing-slash sibling of `/reports/conversations` (Rails routes them
    identically) with `type=agent&user_id=…`.
13. **CSAT page** → `GET /survey/responses/{uuid}` is an unauthenticated HTML page (swagger:
    "redirect the client to this URL") → `csat page <uuid>` prints the URL by default and fetches
    the HTML with `--fetch`; documented as the one non-JSON endpoint.
14. **account_id resolution** → from the active profile (captured at `auth login`), overridable
    per-invocation with the global `--account-id`; excluded from the MCP tool surface alongside
    `--profile` (context switching stays operator-only).
15. **Command grouping** → resources ordered application → platform → client, alphabetical
    within group; `platform`/`client` are nested command groups (manifest names carry the space;
    spec-check word-splits resource names).
16. **Identity/distribution** → binary `cwctl`, module `github.com/jjuanrivvera/cwctl`, MIT,
    tap `jjuanrivvera/homebrew-cwctl`, scoop bucket `jjuanrivvera/scoop-cwctl`;
    `distribution_scope=+release` authorized by Juan 2026-07-10 (build → repo → push → v0.1.0).
17. **Why a new CLI over the official `chatwoot`** → official CLI lacks multi-profile (Juan's PR
    open 3+ weeks with no review while newer PRs merged), lacks keyring fallback for headless
    Linux, and wraps a fraction of the API; this build wraps 144/144 enumerated operations.
    Recorded here, kept out of public docs beyond an honest comparison.
18. **Keyring fallback** → `zalando/go-keyring` first; when no OS keyring is available (headless
    Linux/VPS), fall back to an AES-256-GCM encrypted file under the config dir keyed by
    `CWCTL_KEYRING_PASSWORD` (scrypt-derived), mirroring the fleet's file-keyring pattern;
    `CWCTL_API_KEY` env always wins for CI.
19. **Records decode as raw JSON** (`api.Rec = json.RawMessage`), not typed structs → Chatwoot
    adds response fields every release; raw decode means `-o json`/`-o yaml` never silently drop
    fields, and the renderer's JSON normalization drives tables identically. Typed structs exist
    only where cwctl itself consumes fields (whoami, list meta). Flexible types (ID/FlexInt/
    FlexTime/FlexBool) guard those typed paths.
20. **--limit is client-side** → Chatwoot fixes page size server-side (no per_page param), so
    --limit truncates after fetch; --page selects the server page. Documented in the flag help.
21. **Beyond-the-API (§3c) is intentionally OUTSIDE the manifest** → `backup`/`restore`/`sync`
    are value-adds, not API endpoints, so they don't appear in `api-manifest.json` and don't
    count toward completeness. spec-check only enforces manifest ⊆ CLI, so extra commands are
    fine. Portable kinds = account CONFIG only (labels, canned-responses, custom-attributes,
    custom-filters, automation-rules, teams, webhooks, agent-bots) — never conversations/
    contacts/messages (live data). Chatwoot exposes no stable cross-instance handle, so matching
    is by natural key (title/short_code/name/url/attribute_key) with mandatory duplicate-detect-
    and-skip; "unchanged" compares only the writable-field whitelist (auto-drops id/timestamps).
    `restore`/`sync` are annotated destructive (they can `--prune`), so the agent guard
    hard-blocks them. Inboxes/portals excluded from portability (channel credentials / no delete).
