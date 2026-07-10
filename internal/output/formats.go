package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

func renderJSON(v any, opts Options) error {
	enc := json.NewEncoder(opts.Out)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

func renderYAML(v any, opts Options) error {
	// json.Number is not directly YAML-friendly; normalize it to plain scalars first.
	enc := yaml.NewEncoder(opts.Out)
	enc.SetIndent(2)
	defer func() { _ = enc.Close() }()
	return enc.Encode(yamlNormalize(v))
}

// yamlNormalize converts json.Number to int64/float64 and walks maps/slices so YAML emits
// clean scalars instead of a tagged number type.
func yamlNormalize(v any) any {
	switch t := v.(type) {
	case json.Number:
		if i, err := t.Int64(); err == nil {
			return i
		}
		if f, err := t.Float64(); err == nil {
			return f
		}
		return t.String()
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[k] = yamlNormalize(val)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, val := range t {
			out[i] = yamlNormalize(val)
		}
		return out
	default:
		return v
	}
}

func renderID(v any, opts Options) error {
	rows := toRows(v)
	if len(rows) == 0 {
		return nil
	}
	key := idColumn(rows[0], opts.Columns)
	for _, r := range rows {
		if val, ok := r[key]; ok && val != "" {
			if _, err := fmt.Fprintln(opts.Out, val); err != nil {
				return err
			}
		}
	}
	return nil
}

// idColumn picks the id-bearing column for `-o id`: an explicit first --columns entry, else
// the first present of a small id-like preference list, else the first column.
func idColumn(row map[string]string, columns []string) string {
	if len(columns) > 0 {
		return columns[0]
	}
	for _, k := range []string{"id", "identifier", "source_id", "conversation_id", "message_id"} {
		if _, ok := row[k]; ok {
			return k
		}
	}
	cols, _ := columnOrder([]map[string]string{row}, Options{MaxCols: 1})
	if len(cols) > 0 {
		return cols[0]
	}
	return "value"
}

func renderCSV(v any, opts Options) error {
	rows := toRows(v)
	if len(rows) == 0 {
		return nil
	}
	cols, capped := columnOrder(rows, opts)
	if capped && !opts.Quiet {
		_, _ = fmt.Fprintf(opts.Err, "note: showing %d columns; use -o json for the full record\n", len(cols))
	}
	w := csv.NewWriter(opts.Out)
	if err := w.Write(cols); err != nil {
		return err
	}
	for _, r := range rows {
		rec := make([]string, len(cols))
		for i, c := range cols {
			rec[i] = sanitizeCSV(r[c])
		}
		if err := w.Write(rec); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

// sanitizeCSV neutralizes spreadsheet formula injection (CWE-1236): a leading = + @ — or a
// leading - that is not a real negative number — is prefixed with a single quote so the cell
// can never execute as a formula in Excel/Sheets.
func sanitizeCSV(s string) string {
	if s == "" {
		return s
	}
	switch s[0] {
	case '=', '+', '@':
		return "'" + s
	case '-':
		if !isNumeric(s) {
			return "'" + s
		}
	}
	return s
}

func isNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	dot := false
	for i, r := range s {
		switch {
		case r == '-' && i == 0:
		case r == '.' && !dot:
			dot = true
		case r >= '0' && r <= '9':
		default:
			return false
		}
	}
	return true
}

// truncCell rune-aware truncates a wide cell with an ellipsis so tables stay readable.
func truncCell(s string, max int) (string, bool) {
	if max <= 0 || utf8.RuneCountInString(s) <= max {
		return s, false
	}
	runes := []rune(s)
	if max < 1 {
		max = 1
	}
	return string(runes[:max-1]) + "…", true
}
