# wootctl demo — rendered with VHS (https://github.com/charmbracelet/vhs).
#
# NO ACCOUNT, NO SIGNUP, NO KEYRING, NO NETWORK beyond loopback. All data is
# invented by tools/demoserver; tables, JSON, --jq and command execution are real.
# Configuration is isolated from the recorder's own profiles and credentials.
#
# Regenerate:
#   go build -o bin/wootctl ./cmd/wootctl && go build -o bin/demoserver ./tools/demoserver && vhs .demo/demo.tape

Output assets/demo.gif

Set Shell bash
Set FontSize 15
Set Width 1200
Set Height 760
Set Padding 18
Set Theme "Dracula"
Set TypingSpeed 28ms

Hide
Type "export PATH=$PWD/bin:$PATH XDG_CONFIG_HOME=$(mktemp -d) WOOTCTL_API_KEY=demo-token-no-account WOOTCTL_ACCOUNT_ID=999"
Enter
Type "demoserver 127.0.0.1:8646 >/dev/null 2>&1 & demo_pid=$!; trap 'kill $demo_pid 2>/dev/null; rm -rf $XDG_CONFIG_HOME' EXIT; sleep 1"
Enter
Type "alias wootctl='wootctl --base-url http://127.0.0.1:8646 --profile demo'; clear"
Enter
Show

Sleep 500ms

Type "wootctl conversations list --status open"
Sleep 250ms
Enter
Sleep 2400ms

Type "wootctl messages list 9001"
Sleep 250ms
Enter
Sleep 2400ms

Type "clear"
Enter
Sleep 300ms

Type "wootctl messages create 9001 --content 'Use Reports > Export.' -o json"
Sleep 250ms
Enter
Sleep 2400ms

Type "clear"
Enter
Sleep 300ms

Type "wootctl messages list 9001 -o json --jq '.[].content'"
Sleep 250ms
Enter
Sleep 2400ms

Type "clear"
Enter
Sleep 300ms

Type "wootctl conversations toggle-status 9001 --status resolved -o json"
Sleep 250ms
Enter
Sleep 2400ms
