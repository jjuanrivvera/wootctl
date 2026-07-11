## wootctl

A fast, scriptable CLI for the full Chatwoot API

### Synopsis

wootctl drives Chatwoot from the terminal: conversations, messages, contacts,
agents, teams, inboxes, reports, the platform API, and the public client API — 144/144
documented operations, with named profiles for working across several instances.

Examples:
  wootctl auth login
  wootctl conversations list --status open
  wootctl messages create 123 --content "On it!"
  wootctl contacts search --q ana -o json
  wootctl reports summary --since 2026-06-01 --until 2026-07-01
  wootctl --profile staging conversations list --all

### Options

```
      --account-id string   override the profile's account id for this invocation
      --all                 fetch all pages (list commands)
      --base-url string     override the instance base URL
      --columns strings     comma-separated columns to show
      --dry-run             print the equivalent curl and make no request
      --filter strings      client-side field=value filters (list commands)
  -h, --help                help for wootctl
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

* [wootctl account](wootctl_account.md)	 - Read and update the current account
* [wootctl agent](wootctl_agent.md)	 - AI-agent integration helpers
* [wootctl agent-bots](wootctl_agent-bots.md)	 - Manage account agent bots
* [wootctl agents](wootctl_agents.md)	 - Manage the account's agents
* [wootctl alias](wootctl_alias.md)	 - Manage user-defined command aliases
* [wootctl api](wootctl_api.md)	 - Send a raw authenticated request (escape hatch)
* [wootctl audit-logs](wootctl_audit-logs.md)	 - Read the account audit log (Enterprise)
* [wootctl auth](wootctl_auth.md)	 - Manage Chatwoot tokens and verify authentication
* [wootctl automation-rules](wootctl_automation-rules.md)	 - Manage automation rules
* [wootctl backup](wootctl_backup.md)	 - Back up account config (labels, canned responses, automation, teams, …) to a git-friendly dir
* [wootctl canned-responses](wootctl_canned-responses.md)	 - Manage canned responses (saved reply snippets)
* [wootctl client](wootctl_client.md)	 - Public client API (inbox/contact/conversation flows) — no token required
* [wootctl completion](wootctl_completion.md)	 - Generate shell completion scripts
* [wootctl config](wootctl_config.md)	 - Inspect and edit wootctl configuration
* [wootctl contacts](wootctl_contacts.md)	 - Manage contacts
* [wootctl conversations](wootctl_conversations.md)	 - Manage conversations
* [wootctl csat](wootctl_csat.md)	 - CSAT survey page for a conversation
* [wootctl custom-attributes](wootctl_custom-attributes.md)	 - Manage custom attribute definitions
* [wootctl custom-filters](wootctl_custom-filters.md)	 - Manage saved custom filters
* [wootctl doctor](wootctl_doctor.md)	 - Diagnose configuration, credentials, and connectivity
* [wootctl inboxes](wootctl_inboxes.md)	 - Manage inboxes and their members
* [wootctl init](wootctl_init.md)	 - First-run setup wizard
* [wootctl integrations](wootctl_integrations.md)	 - List integration apps and manage integration hooks
* [wootctl labels](wootctl_labels.md)	 - Manage labels
* [wootctl mcp](wootctl_mcp.md)	 - MCP server management
* [wootctl messages](wootctl_messages.md)	 - Read and send messages in a conversation
* [wootctl platform](wootctl_platform.md)	 - Platform API (accounts, users, agent bots) — needs a platform app token
* [wootctl portals](wootctl_portals.md)	 - Manage help-center portals, articles, and categories
* [wootctl profile](wootctl_profile.md)	 - Read and update your own user profile
* [wootctl reports](wootctl_reports.md)	 - Account analytics (v2 reports, summaries, reporting events)
* [wootctl restore](wootctl_restore.md)	 - Reconcile a backup dir into the account (create/update/skip; --prune removes drift)
* [wootctl sync](wootctl_sync.md)	 - Copy account config from one instance to another (the multi-instance payoff)
* [wootctl teams](wootctl_teams.md)	 - Manage teams and their members
* [wootctl version](wootctl_version.md)	 - Print version, commit, and build date
* [wootctl webhooks](wootctl_webhooks.md)	 - Manage account webhook subscriptions

