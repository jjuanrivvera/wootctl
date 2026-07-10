// Package output renders Chatwoot API JSON in table/json/yaml/csv/id. One renderer serves
// every command, driven by JSON normalization: an object becomes one row, an array of objects
// becomes many, nested fields flatten to dotted keys. Output is deterministic (preferred-key
// order then alphabetical) and pipe-clean (notes go to stderr) per the cliwright standard.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/itchyny/gojq"
)

// Format is an output format.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
	FormatCSV   Format = "csv"
	FormatID    Format = "id"
)

// Valid reports whether f is a known format.
func (f Format) Valid() bool {
	switch f {
	case FormatTable, FormatJSON, FormatYAML, FormatCSV, FormatID:
		return true
	}
	return false
}

// Options configure a single Render call.
type Options struct {
	Format  Format
	Columns []string // explicit column selection (overrides the preferred order)
	NoColor bool
	Quiet   bool
	JQ      string // optional gojq expression applied before rendering
	MaxCols int    // cap auto-detected columns (default 10)
	Out     io.Writer
	Err     io.Writer
}

func (o *Options) defaults() {
	if o.Out == nil {
		o.Out = os.Stdout
	}
	if o.Err == nil {
		o.Err = os.Stderr
	}
	if o.MaxCols == 0 {
		o.MaxCols = 10
	}
	if o.Format == "" {
		o.Format = FormatTable
	}
}

// preferredOrder lists Chatwoot fields that should lead a table when present, before the
// rest fall back to alphabetical. Deterministic ordering is required by GOAL.md §11.
var preferredOrder = []string{
	"id", "identifier", "source_id", "name", "title", "email", "phone_number",
	"status", "priority", "role", "availability_status", "content", "message_type",
	"inbox_id", "channel_type", "short_code", "event_name", "active", "url",
	"description", "created_at", "updated_at",
}

// Render writes data in the requested format. data is the raw API response JSON.
func Render(data json.RawMessage, opts Options) error {
	opts.defaults()

	var v any
	if len(data) == 0 {
		v = nil
	} else {
		dec := json.NewDecoder(strings.NewReader(string(data)))
		dec.UseNumber() // preserve large int64 ids without float rounding
		if err := dec.Decode(&v); err != nil {
			return fmt.Errorf("decode result: %w", err)
		}
	}

	if opts.JQ != "" {
		out, err := applyJQ(opts.JQ, v)
		if err != nil {
			return err
		}
		v = out
	}

	switch opts.Format {
	case FormatJSON:
		return renderJSON(v, opts)
	case FormatYAML:
		return renderYAML(v, opts)
	case FormatCSV:
		return renderCSV(v, opts)
	case FormatID:
		return renderID(v, opts)
	case FormatTable:
		return renderTable(v, opts)
	default:
		return fmt.Errorf("unknown output format %q (want table|json|yaml|csv|id)", opts.Format)
	}
}

// applyJQ runs a gojq program over v and returns the single combined result (scalar when one
// output, slice when many) so downstream rendering stays uniform.
func applyJQ(program string, v any) (any, error) {
	q, err := gojq.Parse(program)
	if err != nil {
		return nil, fmt.Errorf("invalid --jq expression: %w", err)
	}
	iter := q.Run(v)
	var results []any
	for {
		out, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := out.(error); ok {
			return nil, fmt.Errorf("--jq: %w", err)
		}
		results = append(results, out)
	}
	switch len(results) {
	case 0:
		return nil, nil
	case 1:
		return results[0], nil
	default:
		return results, nil
	}
}

// toRows normalizes any decoded value into table rows plus the union of their keys.
func toRows(v any) []map[string]string {
	switch t := v.(type) {
	case nil:
		return nil
	case []any:
		rows := make([]map[string]string, 0, len(t))
		for _, item := range t {
			rows = append(rows, flatten(item))
		}
		return rows
	default:
		return []map[string]string{flatten(v)}
	}
}

// flatten reduces a value to a single row of dotted-key → string-cell. Objects recurse;
// arrays and leftover values are rendered as compact JSON so a cell is always printable.
func flatten(v any) map[string]string {
	row := map[string]string{}
	switch t := v.(type) {
	case map[string]any:
		flattenInto(row, "", t)
	default:
		row["value"] = scalar(v)
	}
	return row
}

func flattenInto(row map[string]string, prefix string, m map[string]any) {
	for k, val := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch child := val.(type) {
		case map[string]any:
			flattenInto(row, key, child)
		default:
			row[key] = scalar(val)
		}
	}
}

func scalar(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case json.Number:
		return t.String()
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

// columnOrder picks and orders the columns for a row set: an explicit --columns list wins;
// otherwise preferred keys lead, then the rest alphabetically, capped at MaxCols.
func columnOrder(rows []map[string]string, opts Options) ([]string, bool) {
	if len(opts.Columns) > 0 {
		return opts.Columns, false
	}
	seen := map[string]bool{}
	for _, r := range rows {
		for k := range r {
			seen[k] = true
		}
	}
	var ordered []string
	for _, k := range preferredOrder {
		if seen[k] {
			ordered = append(ordered, k)
			delete(seen, k)
		}
	}
	rest := make([]string, 0, len(seen))
	for k := range seen {
		rest = append(rest, k)
	}
	sort.Strings(rest)
	ordered = append(ordered, rest...)

	capped := false
	if len(ordered) > opts.MaxCols {
		ordered = ordered[:opts.MaxCols]
		capped = true
	}
	return ordered, capped
}
