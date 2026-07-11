// Package version holds build metadata, injected at link time via -ldflags.
package version

import "fmt"

// These are set by the Makefile / GoReleaser ldflags. Defaults make `go run` work.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// String renders a one-line human summary, e.g. "wootctl v1.2.3 (abc1234, 2026-06-29)".
func String() string {
	return fmt.Sprintf("wootctl %s (%s, %s)", Version, Commit, Date)
}

// Info is the machine-readable form returned by `version --json`.
type Info struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

// Get returns the current build metadata.
func Get() Info { return Info{Version: Version, Commit: Commit, Date: Date} }
