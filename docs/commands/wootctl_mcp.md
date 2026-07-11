## wootctl mcp

MCP server management

### Synopsis

Manage MCP servers for AI assistants and code editors

### Options

```
  -h, --help   help for mcp
```

### Options inherited from parent commands

```
      --account-id string   override the profile's account id for this invocation
      --all                 fetch all pages (list commands)
      --base-url string     override the instance base URL
      --columns strings     comma-separated columns to show
      --dry-run             print the equivalent curl and make no request
      --filter strings      client-side field=value filters (list commands)
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

* [wootctl](wootctl.md)	 - A fast, scriptable CLI for the full Chatwoot API
* [wootctl mcp claude](wootctl_mcp_claude.md)	 - Manage Claude Desktop MCP servers
* [wootctl mcp cursor](wootctl_mcp_cursor.md)	 - Manage Cursor MCP servers
* [wootctl mcp start](wootctl_mcp_start.md)	 - Start the MCP server
* [wootctl mcp stream](wootctl_mcp_stream.md)	 - Stream the MCP server over HTTP
* [wootctl mcp tools](wootctl_mcp_tools.md)	 - Export tools as JSON
* [wootctl mcp vscode](wootctl_mcp_vscode.md)	 - Manage VSCode MCP servers

