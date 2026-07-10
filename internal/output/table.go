package output

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

// maxCellWidth bounds a single table cell so one long text field can't blow out the layout.
const maxCellWidth = 48

func renderTable(v any, opts Options) error {
	switch v.(type) {
	case nil:
		return nil
	case map[string]any, []any:
		// fall through to the tabular path below
	default:
		// A scalar result (e.g. `true` from deleteMessage, a count from getChatMemberCount)
		// renders as a bare value, not a one-cell VALUE table.
		if opts.Quiet {
			return nil
		}
		_, err := fmt.Fprintln(opts.Out, scalar(v))
		return err
	}

	rows := toRows(v)
	if len(rows) == 0 {
		return nil
	}

	cols, capped := columnOrder(rows, opts)
	if capped && !opts.Quiet {
		_, _ = fmt.Fprintf(opts.Err, "note: showing %d of more columns; use -o json for the full record\n", len(cols))
	}

	// Compute display cells (truncated) and per-column widths.
	truncated := false
	display := make([][]string, len(rows))
	widths := make([]int, len(cols))
	for i, c := range cols {
		widths[i] = utf8.RuneCountInString(c)
	}
	for ri, r := range rows {
		display[ri] = make([]string, len(cols))
		for ci, c := range cols {
			cell, cut := truncCell(r[c], maxCellWidth)
			truncated = truncated || cut
			display[ri][ci] = cell
			if w := utf8.RuneCountInString(cell); w > widths[ci] {
				widths[ci] = w
			}
		}
	}

	color := useColor(opts)
	var b strings.Builder
	// Header
	for i, c := range cols {
		h := pad(strings.ToUpper(c), widths[i])
		if color {
			h = bold(h)
		}
		b.WriteString(h)
		if i < len(cols)-1 {
			b.WriteString("  ")
		}
	}
	b.WriteByte('\n')
	// Rows
	for _, r := range display {
		for i := range cols {
			b.WriteString(pad(r[i], widths[i]))
			if i < len(cols)-1 {
				b.WriteString("  ")
			}
		}
		b.WriteByte('\n')
	}

	if _, err := fmt.Fprint(opts.Out, b.String()); err != nil {
		return err
	}
	if truncated && !opts.Quiet {
		_, _ = fmt.Fprintln(opts.Err, "note: some cells were truncated; use -o json for full values")
	}
	return nil
}

// pad right-pads s to width display columns (rune-aware).
func pad(s string, width int) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return s
	}
	return s + strings.Repeat(" ", width-n)
}

// useColor decides whether to emit ANSI color: never when --no-color or NO_COLOR is set, and
// only when the output is an interactive terminal (honored in the renderer, not just claimed).
func useColor(opts Options) bool {
	if opts.NoColor {
		return false
	}
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	return isTerminal(opts.Out)
}

func isTerminal(w any) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

func bold(s string) string { return "\x1b[1m" + s + "\x1b[0m" }
