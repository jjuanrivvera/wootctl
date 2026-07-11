package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestHookScript_BashExecution exercises the generated hook script with real bash to verify
// the adversarial cases: obfuscation, path-invoked binaries, alias paths, the raw api
// escape hatch, and the benign lookalikes that must stay allowed. Gated on a POSIX shell
// being available so it is safe in the regular suite.
func TestHookScript_BashExecution(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bash hook tests require a POSIX shell; skipping on windows")
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not found in PATH; skipping hook execution tests")
	}

	// Generate the hook from the real classification so blocked_cmds/blocked_tools are
	// fully populated (canonical + alias paths).
	hookContent := hookScript(classifyAPICommands(false))
	tmpDir := t.TempDir()
	hookFile := filepath.Join(tmpDir, "wootctl-guard.sh")
	if err := os.WriteFile(hookFile, []byte(hookContent), 0o755); err != nil { // #nosec G306 -- hook must be executable
		t.Fatalf("write hook: %v", err)
	}

	bashPayload := func(command string) string {
		b, _ := json.Marshal(map[string]any{
			"tool_name":  "Bash",
			"tool_input": map[string]any{"command": command},
		})
		return string(b)
	}
	mcpPayload := func(toolName string) string {
		b, _ := json.Marshal(map[string]any{
			"tool_name":  toolName,
			"tool_input": map[string]any{},
		})
		return string(b)
	}

	runHook := func(t *testing.T, payload string) string {
		t.Helper()
		cmd := exec.Command(bash, hookFile)
		cmd.Stdin = strings.NewReader(payload)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		// The hook always exits 0; the decision is in the JSON output.
		if err := cmd.Run(); err != nil {
			t.Logf("hook output: %s", out.String())
			t.Fatalf("hook script exited non-zero: %v", err)
		}
		return out.String()
	}

	isDenied := func(output string) bool {
		return strings.Contains(output, `"permissionDecision":"deny"`)
	}

	cases := []struct {
		name       string
		payload    string
		wantDenied bool
	}{
		// --- direct blocked commands ---
		{"labels_delete_denied", bashPayload("wootctl labels delete 5"), true},
		{"contacts_delete_denied", bashPayload("wootctl contacts delete 12"), true},
		{"contacts_merge_denied", bashPayload("wootctl contacts merge --base 1 --mergee 2"), true},
		{"messages_delete_denied", bashPayload("wootctl messages delete 42 105"), true},
		{"teams_remove_members_denied", bashPayload("wootctl teams remove-members 3 --user-ids 1"), true},
		{"inboxes_remove_members_denied", bashPayload("wootctl inboxes remove-members 3 --user-ids 1"), true},
		{"integrations_delete_hook_denied", bashPayload("wootctl integrations delete-hook 5"), true},
		{"platform_users_delete_denied", bashPayload("wootctl platform users delete 7"), true},
		{"platform_accounts_delete_denied", bashPayload("wootctl platform accounts delete 2"), true},
		// --- cobra alias paths (bypass without enumeration) ---
		{"label_alias_delete_denied", bashPayload("wootctl label delete 5"), true},
		{"msg_alias_delete_denied", bashPayload("wootctl msg delete 42 105"), true},
		{"msgs_alias_delete_denied", bashPayload("wootctl msgs delete 42 105"), true},
		{"contact_alias_merge_denied", bashPayload("wootctl contact merge --base 1 --mergee 2"), true},
		{"canned_alias_delete_denied", bashPayload("wootctl canned delete 9"), true},
		{"filters_alias_delete_denied", bashPayload("wootctl filters delete 4"), true},
		// --- alias minting ---
		{"alias_set_denied", bashPayload(`wootctl alias set kill "labels delete"`), true},
		// --- obfuscation ---
		{"quote_split_denied", bashPayload(`wootctl labels de""lete 5`), true},
		{"single_quote_split_denied", bashPayload(`wootctl contacts me''rge --base 1 --mergee 2`), true},
		{"backslash_denied", bashPayload(`wootctl labels de\lete 5`), true},
		{"newline_continuation_denied", bashPayload("wootctl labels \\\ndelete 5"), true},
		// --- command position after separators ---
		{"after_semicolon_denied", bashPayload("true; wootctl labels delete 5"), true},
		{"after_pipe_denied", bashPayload("echo hi | wootctl labels delete 5"), true},
		{"after_and_denied", bashPayload("true && wootctl contacts delete 12"), true},
		{"trailing_separator_denied", bashPayload("wootctl labels delete 5;true"), true},
		{"env_prefix_denied", bashPayload("env WOOTCTL_API_KEY=x wootctl labels delete 5"), true},
		// --- path-invoked binaries ---
		{"relative_path_binary_denied", bashPayload("./bin/wootctl labels delete 5"), true},
		{"absolute_path_binary_denied", bashPayload("/usr/local/bin/wootctl labels delete 5"), true},
		{"absolute_path_api_denied", bashPayload("/usr/local/bin/wootctl api DELETE api/v1/accounts/1/labels/5"), true},
		// --- raw api escape hatch (METHOD position; only GET/HEAD/OPTIONS pass) ---
		{"api_delete_denied", bashPayload("wootctl api DELETE api/v1/accounts/1/labels/5"), true},
		{"api_lowercase_delete_denied", bashPayload("wootctl api delete api/v1/accounts/1/labels/5"), true},
		{"api_post_denied", bashPayload("wootctl api POST api/v1/accounts/1/labels -d '{}'"), true},
		{"api_patch_denied", bashPayload("wootctl api PATCH api/v1/accounts/1/labels/5 -d '{}'"), true},
		{"api_put_denied", bashPayload("wootctl api PUT api/v1/profile -d '{}'"), true},
		{"api_flag_before_method_denied", bashPayload("wootctl api -q x=1 DELETE api/v1/accounts/1/labels/5"), true},
		{"api_compound_get_then_delete_denied", bashPayload("wootctl api GET api/v1/profile;wootctl api DELETE api/v1/accounts/1/labels/5"), true},
		// --- raw api reads stay allowed ---
		{"api_get_allowed", bashPayload("wootctl api GET api/v1/profile"), false},
		{"api_get_lowercase_allowed", bashPayload("wootctl api get api/v1/profile"), false},
		{"api_head_allowed", bashPayload("wootctl api HEAD api/v1/profile"), false},
		{"api_get_delete_in_path_allowed", bashPayload("wootctl api GET api/v1/accounts/1/labels/delete_club"), false},
		// --- benign lookalikes that must stay allowed ---
		{"conversations_list_allowed", bashPayload("wootctl conversations list --status open"), false},
		{"create_with_delete_in_arg_allowed", bashPayload(`wootctl messages create 42 --content "how to delete a label"`), false},
		{"cat_file_allowed", bashPayload("cat labels_delete.go"), false},
		{"other_binary_allowed", bashPayload("mywootctl labels delete 5"), false},
		{"other_binary_api_allowed", bashPayload("mywootctl api DELETE api/v1/accounts/1/labels/5"), false},
		{"toggle_status_is_write_not_blocked", bashPayload("wootctl conversations toggle-status 42 --status resolved"), false},
		{"conv_alias_list_allowed", bashPayload("wootctl conv list"), false},
		// --- MCP branch ---
		{"mcp_labels_delete_denied", mcpPayload("mcp__wootctl__cw_labels_delete"), true},
		{"mcp_contacts_merge_denied", mcpPayload("mcp__wootctl__cw_contacts_merge"), true},
		{"mcp_audit_logs_hyphen_delete_denied", mcpPayload("mcp__wootctl__cw_agent-bots_delete"), true},
		{"mcp_platform_users_delete_denied", mcpPayload("mcp__wootctl__cw_platform_users_delete"), true},
		{"mcp_conversations_list_allowed", mcpPayload("mcp__wootctl__cw_conversations_list"), false},
		{"mcp_near_miss_allowed", mcpPayload("mcp__wootctl__cw_labels_delete2"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := runHook(t, tc.payload)
			if denied := isDenied(output); denied != tc.wantDenied {
				t.Errorf("want denied=%v, got denied=%v\noutput: %s", tc.wantDenied, denied, output)
			}
		})
	}
}

// TestHookScript_BashExecutionNoJq exercises the no-jq fallback path with a STRICT PATH: a
// bin dir holding only the POSIX tools the hook needs, so jq is genuinely unreachable
// (merely prepending an empty dir leaves jq resolvable — the test flaw that masked a
// fail-open bug in two audited repos, GOAL.md §3b #3).
func TestHookScript_BashExecutionNoJq(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bash hook tests require a POSIX shell; skipping on windows")
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not found in PATH; skipping hook execution tests")
	}

	hookContent := hookScript(classifyAPICommands(false))
	tmpDir := t.TempDir()
	hookFile := filepath.Join(tmpDir, "wootctl-guard.sh")
	if err := os.WriteFile(hookFile, []byte(hookContent), 0o755); err != nil { // #nosec G306 -- hook must be executable
		t.Fatalf("write hook: %v", err)
	}

	binDir := filepath.Join(tmpDir, "nojq-bin")
	if err := os.Mkdir(binDir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for _, tool := range []string{"cat", "tr", "grep", "sed", "printf", "env"} {
		p, lerr := exec.LookPath(tool)
		if lerr != nil {
			continue // shell builtins (printf) need no symlink
		}
		if serr := os.Symlink(p, filepath.Join(binDir, tool)); serr != nil {
			t.Fatalf("symlink %s: %v", tool, serr)
		}
	}

	bashPayload := func(command string) string {
		b, _ := json.Marshal(map[string]any{
			"tool_name":  "Bash",
			"tool_input": map[string]any{"command": command},
		})
		return string(b)
	}

	runHookNoJq := func(t *testing.T, payload string) string {
		t.Helper()
		cmd := exec.Command(bash, hookFile)
		cmd.Stdin = strings.NewReader(payload)
		env := make([]string, 0, len(os.Environ()))
		for _, e := range os.Environ() {
			if !strings.HasPrefix(e, "PATH=") {
				env = append(env, e)
			}
		}
		cmd.Env = append(env, "PATH="+binDir)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			t.Logf("hook output: %s", out.String())
			t.Fatalf("hook script exited non-zero: %v", err)
		}
		return out.String()
	}

	isDenied := func(output string) bool {
		return strings.Contains(output, `"permissionDecision":"deny"`)
	}

	cases := []struct {
		name       string
		payload    string
		wantDenied bool
	}{
		{"nojq_labels_delete_denied", bashPayload("wootctl labels delete 5"), true},
		{"nojq_obfuscated_delete_denied", bashPayload(`wootctl labels de""lete 5`), true},
		{"nojq_path_binary_denied", bashPayload("./bin/wootctl labels delete 5"), true},
		{"nojq_api_delete_denied", bashPayload("wootctl api DELETE api/v1/accounts/1/labels/5"), true},
		{"nojq_cat_file_allowed", bashPayload("cat labels_delete.go"), false},
		{"nojq_create_allowed", bashPayload(`wootctl messages create 42 --content "delete this later"`), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := runHookNoJq(t, tc.payload)
			if denied := isDenied(output); denied != tc.wantDenied {
				t.Errorf("want denied=%v, got denied=%v\noutput: %s", tc.wantDenied, denied, output)
			}
		})
	}
}
