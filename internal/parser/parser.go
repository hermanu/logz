// Package parser converts raw log lines into structured [Entry] values.
//
// Implementations are stateless and safe for concurrent use. The package also
// exposes [Detect], which sniffs a sample byte slice to pick a parser.
package parser

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"time"
)

// Level is a log severity. Zero value is [LevelUnknown].
type Level uint8

// Severity levels in ascending order.
const (
	LevelUnknown Level = iota
	LevelTrace
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String returns the canonical lowercase name for the level.
func (l Level) String() string {
	switch l {
	case LevelUnknown:
		return "unknown"
	case LevelTrace:
		return "trace"
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelFatal:
		return "fatal"
	}
	return "unknown"
}

// ParseLevel maps a string (case-insensitive) to a [Level].
//
// Recognized aliases: "warning"→warn, "err"→error, "panic"/"crit"/"critical"→fatal.
// Unknown strings return [LevelUnknown] and false.
func ParseLevel(s string) (Level, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "trace":
		return LevelTrace, true
	case "debug", "dbg":
		return LevelDebug, true
	case "info", "information", "notice":
		return LevelInfo, true
	case "warn", "warning":
		return LevelWarn, true
	case "error", "err":
		return LevelError, true
	case "fatal", "panic", "crit", "critical", "emergency", "alert":
		return LevelFatal, true
	default:
		return LevelUnknown, false
	}
}

// Entry is a parsed log line. Fields preserves the original key/value pairs
// from structured formats; Raw is always populated with the original line so
// callers can fall back to it for output.
type Entry struct {
	Timestamp time.Time
	Level     Level
	Message   string
	Fields    map[string]string
	Raw       string
}

// ErrSkip signals the parser explicitly chose to drop a line (e.g. a comment
// or blank line). Callers should treat it as a soft skip, not an error.
var ErrSkip = errors.New("parser: skip line")

// Parser turns a single log line (without trailing newline) into an [Entry].
//
// Implementations must be safe for concurrent use across goroutines.
type Parser interface {
	// Name returns a stable identifier (e.g. "json", "text").
	Name() string
	// Parse converts a single line into an Entry. Returning [ErrSkip] tells the
	// caller to drop the line silently; any other error is a parse failure that
	// should be surfaced to the user (typically as a warning count).
	Parse(line string) (Entry, error)
}

// Options configures parser behavior. The zero value is valid and yields
// sensible defaults; users can override via the config system.
type Options struct {
	// LevelKeys are the structured-field keys to inspect for log level, in
	// priority order. Defaults to {"level", "severity", "lvl"}.
	LevelKeys []string

	// MessageKeys are the structured-field keys to inspect for the human
	// message. Defaults to {"msg", "message", "log"}.
	MessageKeys []string

	// TimestampKeys are the structured-field keys to inspect for timestamps.
	// Defaults to {"ts", "time", "timestamp", "@timestamp"}.
	TimestampKeys []string

	// TimestampLayouts are extra time.Parse layouts to attempt, in addition to
	// the built-in defaults (RFC3339, common variants, Unix epoch).
	TimestampLayouts []string

	// LevelAliases maps custom level strings (lowercased) to canonical levels.
	// E.g. {"30": "info"} for numeric Bunyan-style levels.
	LevelAliases map[string]Level
}

// DefaultOptions returns a copy of the built-in defaults.
func DefaultOptions() Options {
	return Options{
		LevelKeys:     []string{"level", "severity", "lvl"},
		MessageKeys:   []string{"msg", "message", "log"},
		TimestampKeys: []string{"ts", "time", "timestamp", "@timestamp"},
	}
}

// Detect sniffs sample bytes and returns the most likely parser.
//
// Detection is best-effort: it reads up to maxLines from sample, looking for
// JSON object syntax. Anything else falls through to a [TextParser].
func Detect(sample []byte, opts Options) Parser {
	const maxLines = 8
	scanner := bufio.NewScanner(bytes.NewReader(sample))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for i := 0; i < maxLines && scanner.Scan(); i++ {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		if line[0] == '{' && line[len(line)-1] == '}' {
			return NewJSONParser(opts)
		}
		// First non-empty line wasn't JSON — fall back to text.
		break
	}
	return NewTextParser(opts)
}

// SniffSample reads up to n bytes from r without consuming them, returning a
// new io.Reader that replays the sniffed prefix followed by the rest. Used by
// callers that want to peek at a stream before constructing the parser.
func SniffSample(r io.Reader, n int) ([]byte, io.Reader, error) {
	br := bufio.NewReaderSize(r, n)
	sample, err := br.Peek(n)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, bufio.ErrBufferFull) {
		return nil, nil, err
	}
	return sample, br, nil
}

// commonLayouts is consulted when parsing timestamp strings; ordered so that
// stricter / more common layouts win.
var commonLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05.000Z07:00",
	"2006-01-02 15:04:05.000",
	"2006-01-02 15:04:05",
	"2006/01/02 15:04:05",
	"02/Jan/2006:15:04:05 -0700",
	time.RFC1123,
}

// parseTime tries the configured plus default layouts. Returns zero time if
// none match.
func parseTime(s string, extra []string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	for _, layout := range extra {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	for _, layout := range commonLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
