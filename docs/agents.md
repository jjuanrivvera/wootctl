# AI agents

wootctl is built to be driven by agents — safely.

## MCP server

```bash
wootctl mcp claude enable     # Claude Desktop
wootctl mcp cursor enable     # Cursor
wootctl mcp vscode enable     # VS Code
wootctl mcp start             # stdio server for anything else
```

Every resource command becomes an MCP tool (`cw_conversations_list`,
`cw_messages_create`, …) with truthful annotations: reads carry `readOnlyHint`,
writes `openWorldHint`, deletes `destructiveHint` — so hosts can gate them.

The tool surface **excludes**: `auth`, `config`, `alias`, `init`, `doctor`, `agent`,
the raw `api` hatch, and the secret/instance flags (`--show-token`, `--profile`,
`--base-url`, `--account-id`). An agent can neither read your token, switch
instances, nor disable its own guardrails. The server operates on whatever profile
was active when it started.

## Agent guard

For agents that run shell commands (Claude Code, Codex, OpenCode), generate
guardrails from the live command tree:

```bash
wootctl agent guard --host claude-code --write   # .claude/hooks + settings fragment
wootctl agent guard --host codex --out ~/.codex/config.toml
wootctl agent guard --host opencode
wootctl agent guard --host claude-code --all-writes   # block writes too, not just deletes
```

What the claude-code guard enforces:

- **Hard-blocks irreversible operations** — every `delete`, `contacts merge`,
  `remove-members`, `delete-hook` — under their canonical paths AND every cobra alias
  path (`label delete`, `msg delete`, …).
- **Gates the raw hatch by METHOD**: `wootctl api GET …` passes, `POST/PUT/PATCH/DELETE`
  are denied, case-insensitively, even path-invoked (`./bin/wootctl`).
- **Defeats obfuscation**: quote-splitting (`de""lete`), backslashes, newline
  continuations, command chaining (`;`, `|`, `&&`), `env`-prefixed invocations.
- **Denies `alias set`** so an agent cannot mint a shorthand for a blocked command.

Known limits (by design, documented): variable indirection (`a=delete; wootctl labels $a`)
and shell aliases are not defeated — MCP-only operation is the hard guarantee, the
Bash hook is defense in depth. A line that merely quotes a blocked command
(`rg "wootctl labels delete"`) is denied; that false positive is the safe direction.

Regenerate the guard after upgrading wootctl so new commands are covered.
