package output

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// jsonWith builds valid JSON (ESC encoded as ) for a one-field record.
func jsonWith(name string) string {
	b, _ := json.Marshal([]map[string]string{{"name": name}})
	return string(b)
}

func TestSanitizeTerminal(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"plain text untouched", "Ana María", "Ana María"},
		{"strips CSI color", "\x1b[31mred\x1b[0m", "red"},
		{"strips OSC title-set (BEL)", "\x1b]0;pwned\x07name", "name"},
		{"strips OSC title-set (ST)", "\x1b]0;pwned\x1b\\name", "name"},
		{"drops control chars", "a\x00b\x08c", "abc"},
		{"keeps tab and newline", "a\tb\nc", "a\tb\nc"},
		{"lone escape keeps next char", "a\x1bb", "ab"},
		{"emoji/unicode preserved", "hola 👋 niño", "hola 👋 niño"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, SanitizeTerminal(tc.in))
		})
	}
}

func TestSanitizeTerminal_FastPath(t *testing.T) {
	// A clean string must be returned as-is, exercising the no-alloc fast path.
	s := "just normal text with an = sign"
	assert.Equal(t, s, SanitizeTerminal(s))
}

func TestCellOneLine(t *testing.T) {
	assert.Equal(t, "a b c", cellOneLine("a\tb\nc"))
	assert.Equal(t, "one two", cellOneLine("one\r\ntwo"))
	assert.Equal(t, "plain", cellOneLine("plain"))
}

func TestRender_TableSanitizesEscapes(t *testing.T) {
	// A malicious API value reaching the human table must be neutralized. The OSC title-set
	// sequence (ESC ] 0;… BEL) would rewrite the terminal title; the table must strip it.
	out, _ := render(t, jsonWith("\x1b]0;pwned\x07evil"), Options{Format: FormatTable, NoColor: true})
	assert.NotContains(t, out, "\x1b")
	assert.NotContains(t, out, "pwned", "OSC payload stripped")
	assert.Contains(t, out, "evil")
}

func TestRender_JSONStaysFaithful(t *testing.T) {
	// Machine JSON output must NOT be altered — the consumer owns rendering, and wootctl must
	// not corrupt data on the wire. The ESC survives as an escaped .
	out, _ := render(t, jsonWith("\x1b[31mred"), Options{Format: FormatJSON})
	assert.Contains(t, out, "\\u001b")
}
