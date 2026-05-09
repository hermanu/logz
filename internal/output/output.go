// Package output renders [parser.Entry] values to a writer in pretty or
// JSON form.
package output

import (
	"io"

	"github.com/hermanu/logz/internal/parser"
)

// Mode selects an output format.
type Mode uint8

// Output modes.
const (
	ModePretty Mode = iota
	ModeJSON
)

// Writer renders entries to an underlying io.Writer. Implementations must be
// safe for concurrent calls only if the caller serializes them — the CLI
// pipeline writes from a single goroutine.
type Writer interface {
	Write(e parser.Entry) error
	Close() error
}

// New constructs a Writer for the given mode. NoColor only affects [ModePretty].
func New(mode Mode, w io.Writer, noColor bool) Writer {
	switch mode {
	case ModeJSON:
		return NewJSON(w)
	case ModePretty:
		return NewPretty(w, noColor)
	default:
		return NewPretty(w, noColor)
	}
}
