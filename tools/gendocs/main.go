// Command gendocs generates the Markdown command reference under docs/commands from the
// live cobra tree, so the published docs never drift from the actual CLI surface.
package main

import (
	"log"
	"os"

	"github.com/spf13/cobra/doc"

	"github.com/jjuanrivvera/wootctl/commands"
)

func main() {
	const out = "docs/commands"
	if err := os.MkdirAll(out, 0o750); err != nil {
		log.Fatalf("mkdir %s: %v", out, err)
	}
	root := commands.NewRootCmd()
	root.DisableAutoGenTag = true // omit the timestamp so output is reproducible (no CI drift)
	if err := doc.GenMarkdownTree(root, out); err != nil {
		log.Fatalf("generate docs: %v", err)
	}
}
