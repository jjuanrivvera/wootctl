## cwctl

A fast, scriptable CLI for the full Chatwoot API

### Synopsis

cwctl drives Chatwoot from the terminal: conversations, messages, contacts,
agents, teams, inboxes, reports, the platform API, and the public client API — 144/144
documented operations, with named profiles for working across several instances.

Examples:
  cwctl auth login
  cwctl conversations list --status open
  cwctl messages create 123 --content "On it!"
  cwctl contacts search --q ana -o json
  cwctl reports summary --since 2026-06-01 --until 2026-07-01
  cwctl --profile staging conversations list --all

### Options

```
      --account-id string   override the profile's account id for this invocation
      --all                 fetch all pages (list commands)
      --base-url string     override the instance base URL
      --columns strings     comma-separated columns to show
      --dry-run             print the equivalent curl and make no request
      --filter strings      client-side field=value filters (list commands)
  -h, --help                help for cwctl
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

* [cwctl account](cwctl_account.md)	 - Read and update the current account
* [cwctl agent](cwctl_agent.md)	 - AI-agent integration helpers
* [cwctl agent-bots](cwctl_agent-bots.md)	 - Manage account agent bots
* [cwctl agents](cwctl_agents.md)	 - Manage the account's agents
* [cwctl alias](cwctl_alias.md)	 - Manage user-defined command aliases
* [cwctl api](cwctl_api.md)	 - Send a raw authenticated request (escape hatch)
* [cwctl audit-logs](cwctl_audit-logs.md)	 - Read the account audit log (Enterprise)
* [cwctl auth](cwctl_auth.md)	 - Manage Chatwoot tokens and verify authentication
* [cwctl automation-rules](cwctl_automation-rules.md)	 - Manage automation rules
* [cwctl backup](cwctl_backup.md)	 - Back up account config (labels, canned responses, automation, teams, …) to a git-friendly dir
* [cwctl canned-responses](cwctl_canned-responses.md)	 - Manage canned responses (saved reply snippets)
* [cwctl client](cwctl_client.md)	 - Public client API (inbox/contact/conversation flows) — no token required
* [cwctl completion](cwctl_completion.md)	 - Generate shell completion scripts
* [cwctl config](cwctl_config.md)	 - Inspect and edit cwctl configuration
* [cwctl contacts](cwctl_contacts.md)	 - Manage contacts
* [cwctl conversations](cwctl_conversations.md)	 - Manage conversations
* [cwctl csat](cwctl_csat.md)	 - CSAT survey page for a conversation
* [cwctl custom-attributes](cwctl_custom-attributes.md)	 - Manage custom attribute definitions
* [cwctl custom-filters](cwctl_custom-filters.md)	 - Manage saved custom filters
* [cwctl doctor](cwctl_doctor.md)	 - Diagnose configuration, credentials, and connectivity
* [cwctl inboxes](cwctl_inboxes.md)	 - Manage inboxes and their members
* [cwctl init](cwctl_init.md)	 - First-run setup wizard
* [cwctl integrations](cwctl_integrations.md)	 - List integration apps and manage integration hooks
* [cwctl labels](cwctl_labels.md)	 - Manage labels
* [cwctl mcp](cwctl_mcp.md)	 - MCP server management
* [cwctl messages](cwctl_messages.md)	 - Read and send messages in a conversation
* [cwctl platform](cwctl_platform.md)	 - Platform API (accounts, users, agent bots) — needs a platform app token
* [cwctl portals](cwctl_portals.md)	 - Manage help-center portals, articles, and categories
* [cwctl profile](cwctl_profile.md)	 - Read and update your own user profile
* [cwctl reports](cwctl_reports.md)	 - Account analytics (v2 reports, summaries, reporting events)
* [cwctl restore](cwctl_restore.md)	 - Reconcile a backup dir into the account (create/update/skip; --prune removes drift)
* [cwctl sync](cwctl_sync.md)	 - Copy account config from one instance to another (the multi-instance payoff)
* [cwctl teams](cwctl_teams.md)	 - Manage teams and their members
* [cwctl version](cwctl_version.md)	 - Print version, commit, and build date
* [cwctl webhooks](cwctl_webhooks.md)	 - Manage account webhook subscriptions

