// Command cwctl is a command-line tool for the full Chatwoot API.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jjuanrivvera/cwctl/commands"
	"github.com/jjuanrivvera/cwctl/internal/version"
)

func main() {
	// signal.NotifyContext makes Ctrl-C (SIGINT/SIGTERM) cancel in-flight work: --all
	// pagination, retry backoff, and rate-limit waits all observe this context.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	root := commands.NewRootCmd()
	root.Version = version.Get().Version
	root.SetVersionTemplate(version.String() + "\n")

	// Expand user-defined aliases BEFORE cobra parses, so an alias can map to any command
	// without shadowing a built-in.
	root.SetArgs(commands.ExpandAliases(os.Args[1:]))

	if err := root.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
