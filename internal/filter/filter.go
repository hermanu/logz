// Package filter applies user-supplied predicates to parsed log entries.
package filter

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hermanu/logz/internal/parser"
)

// Filter is a set of predicates evaluated in order against each [parser.Entry].
// The zero value matches every entry.
type Filter struct {
	// MinLevel, when set, drops entries whose Level is below this threshold.
	// [parser.LevelUnknown] is treated as below all real levels.
	MinLevel parser.Level

	// Since drops entries with timestamps before this point. Entries without a
	// timestamp are kept (we can't prove they're outside the window).
	Since time.Time

	// Until drops entries with timestamps after this point.
	Until time.Time

	// Match, if non-nil, requires Raw or Message to contain a match.
	Match *regexp.Regexp

	// FieldEquals requires entry.Fields[k] == v for every (k,v) in this map.
	FieldEquals map[string]string

	// Invert flips the final decision (analogous to grep -v).
	Invert bool
}

// Allow reports whether the entry passes the filter.
func (f *Filter) Allow(e parser.Entry) bool {
	allowed := f.evaluate(e)
	if f.Invert {
		return !allowed
	}
	return allowed
}

func (f *Filter) evaluate(e parser.Entry) bool {
	if f.MinLevel != parser.LevelUnknown && e.Level < f.MinLevel {
		return false
	}
	if !f.Since.IsZero() && !e.Timestamp.IsZero() && e.Timestamp.Before(f.Since) {
		return false
	}
	if !f.Until.IsZero() && !e.Timestamp.IsZero() && e.Timestamp.After(f.Until) {
		return false
	}
	if f.Match != nil && !f.Match.MatchString(e.Raw) && !f.Match.MatchString(e.Message) {
		return false
	}
	for k, want := range f.FieldEquals {
		if got, ok := e.Fields[k]; !ok || got != want {
			return false
		}
	}
	return true
}

// ParseFieldSpec turns "key=value" into its components. Returns an error if
// the input lacks "=".
func ParseFieldSpec(s string) (key, value string, err error) {
	idx := strings.IndexByte(s, '=')
	if idx <= 0 {
		return "", "", fmt.Errorf("filter: expected key=value, got %q", s)
	}
	return s[:idx], s[idx+1:], nil
}

// ParseSince parses a "--since" / "--until" value.
//
// Accepts:
//   - Go duration syntax: "1h", "30m", "2h45m" — interpreted as relative to now.
//   - Absolute timestamps in any of [parser]'s built-in layouts.
//
// now is injected for testability; callers typically pass [time.Now]().
func ParseSince(s string, now time.Time) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, nil
	}
	if d, err := time.ParseDuration(s); err == nil {
		return now.Add(-d), nil
	}
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("filter: unrecognized time %q", s)
}
