// Package logz holds top-level metadata (version, commit, build date) shared
// across the binary. Values are injected via -ldflags at build time.
package logz

// Build metadata. Overridden via -ldflags by the Makefile and GoReleaser.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Version returns the semantic version string set at build time, or "dev"
// when the binary was not built through the project's release tooling.
func Version() string { return version }

// Commit returns the short git commit hash this binary was built from, or
// "none" when commit information was not injected at build time.
func Commit() string { return commit }

// Date returns the build date in RFC3339 format, or "unknown" when the
// build date was not injected at build time.
func Date() string { return date }
