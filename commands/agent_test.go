package commands

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spf13/cobra"
)

// TestEveryAPICommandIsAnnotated locks §3b hardening #4: every runnable command outside
// the local/meta groups must carry an MCP classification annotation. A hand-added command
// without one shows up here (it would classify destructive at guard time — safe — but the
// build should fail loudly instead of shipping an unclassified verb).
func TestEveryAPICommandIsAnnotated(t *testing.T) {
	root := NewRootCmd()
	var offenders []string
	var walk func(cmd *cobra.Command, path string)
	walk = func(cmd *cobra.Command, path string) {
		for _, child := range cmd.Commands() {
			p := strings.TrimSpace(path + " " + child.Name())
			if path == "" && (slices.Contains(localGroups, child.Name()) || child.Name() == "api") {
				continue
			}
			if child.Runnable() {
				if kindAnnotated(child.Annotations) == "" {
					offenders = append(offenders, p)
				}
			}
			walk(child, p)
		}
	}
	walk(root, "")
	assert.Empty(t, offenders, "unannotated API commands (annotate in the builder or add to localGroups)")
}

func kindAnnotated(ann map[string]string) string {
	for _, k := range []string{annReadOnly, annOpenWorld, annDestructive} {
		if ann[k] == "true" {
			return k
		}
	}
	return ""
}

// TestClassifyAPICommands locks the read/write/destructive split for the load-bearing
// cases so a refactor can't silently downgrade a destructive verb.
func TestClassifyAPICommands(t *testing.T) {
	cls := classifyAPICommands(false)

	paths := func(cmds []apiCmdInfo) []string {
		var out []string
		for _, c := range cmds {
			out = append(out, c.Path)
		}
		return out
	}
	reads, writes, destr := paths(cls.Read), paths(cls.Write), paths(cls.Destructive)

	for _, want := range []string{
		"labels delete", "contacts delete", "contacts merge", "messages delete",
		"teams remove-members", "inboxes remove-members", "integrations delete-hook",
		"platform users delete", "platform accounts delete", "platform account-users delete",
		"webhooks delete", "agents delete", "automation-rules delete",
	} {
		assert.Contains(t, destr, want, "must be destructive")
	}
	for _, want := range []string{
		"conversations list", "conversations get", "conversations meta", "contacts search",
		"reports summary", "messages list", "audit-logs list", "platform users sso-link",
		"client inbox get", "profile get", "csat page",
	} {
		assert.Contains(t, reads, want, "must be read-only")
	}
	for _, want := range []string{
		"messages create", "conversations toggle-status", "conversations assign",
		"contacts add-labels", "labels create", "teams update", "platform users create",
		"client messages create",
	} {
		assert.Contains(t, writes, want, "must be write (ask), not read or destructive")
	}

	// A destructive path must never also classify as read (verb-name collision guard).
	for _, r := range reads {
		assert.NotContains(t, destr, r, "path classified both read and destructive")
	}
}

// TestClassify_AllWritesPromotes verifies --all-writes hard-blocks ordinary writes.
func TestClassify_AllWritesPromotes(t *testing.T) {
	cls := classifyAPICommands(true)
	assert.Empty(t, cls.Write)
	var destr []string
	for _, c := range cls.Destructive {
		destr = append(destr, c.Path)
	}
	assert.Contains(t, destr, "messages create")
}

// TestAliasCrossProduct locks §3b hardening #5: alias paths are enumerated.
func TestAliasCrossProduct(t *testing.T) {
	var conv apiCmdInfo
	for _, c := range classifyTree() {
		if c.Path == "conversations get" {
			conv = c
			break
		}
	}
	require.NotEmpty(t, conv.Path, "conversations get not found in classification")
	all := conv.AllPaths()
	assert.Contains(t, all, "conv get")
	assert.Contains(t, all, "convs get")

	// Nested group aliases expand too: client → public.
	var clientList apiCmdInfo
	for _, c := range classifyTree() {
		if c.Path == "client conversations list" {
			clientList = c
			break
		}
	}
	require.NotEmpty(t, clientList.Path)
	assert.Contains(t, clientList.AllPaths(), "public conversations list")
}

// TestMCPExcludesSetupCommands locks the MCP tool surface: no setup/meta/secret command
// may be exposed as a tool, and no secret flag may reach a tool schema.
func TestMCPExcludesSetupCommands(t *testing.T) {
	for _, name := range []string{"agent", "auth", "config", "alias", "init", "doctor", "completion", "version", "api"} {
		assert.Contains(t, excludedFromMCP, name)
	}
	for _, flag := range []string{"show-token", "profile", "base-url", "account-id"} {
		assert.Contains(t, secretFlags, flag)
	}
}

// TestGuardRenderers smoke-checks each host output for the load-bearing content.
func TestGuardRenderers(t *testing.T) {
	cls := classifyAPICommands(false)

	claude, err := renderClaudeCode(cls)
	require.NoError(t, err)
	assert.Contains(t, claude, `Bash(wootctl labels delete:*)`)
	assert.Contains(t, claude, `Bash(wootctl label delete:*)`, "alias path present in deny rules")
	assert.Contains(t, claude, `Bash(wootctl api DELETE:*)`)
	assert.Contains(t, claude, `Bash(wootctl alias set:*)`)
	assert.Contains(t, claude, "mcp__wootctl__cw_labels_delete")
	assert.Contains(t, claude, "mcp__wootctl__cw_agent-bots_delete", "hyphen preserved in exact tool name")
	assert.Contains(t, claude, "PreToolUse")
	assert.Contains(t, claude, "blocked_cmds=(")

	codex, err := renderCodex(cls)
	require.NoError(t, err)
	assert.Contains(t, codex, `approval_policy = "on-request"`)
	assert.Contains(t, codex, `sandbox_mode = "read-only"`)

	oc, err := renderOpenCode(cls)
	require.NoError(t, err)
	assert.Contains(t, oc, `"permission"`)
	assert.Contains(t, oc, `"bash"`)
	assert.Contains(t, oc, `"wootctl labels delete*": "deny"`)
	assert.Contains(t, oc, `"wootctl conversations list*": "allow"`)
}
