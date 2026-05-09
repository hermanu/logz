// Package logz holds top-level metadata (version, commit, build date) shared
// across the binary. Values are injected via -ldflags at build time.
package logz

// Build metadata. Overridden via -ldflags by the Makefile and GoReleaser.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Version returns the semantic version string (e.g. "v1.0.0" or "dev").
func Version() string { return version }

// Commit returns the short git commit hash this binary was built from.
func Commit() string { return commit }

// Date returns the build date in RFC3339 format.
func Date() string { return date }
