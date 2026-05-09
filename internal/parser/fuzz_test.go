package parser_test

import (
	"errors"
	"testing"

	"github.com/hermanu/logz/internal/parser"
)

// FuzzJSONParser_Parse asserts the parser never panics, regardless of input.
// It accepts both well-formed and malformed JSON; the only contract is that
// either an Entry comes back with Raw populated, or a non-panicking error.
func FuzzJSONParser_Parse(f *testing.F) {
	seeds := []string{
		`{"level":"info","msg":"hi"}`,
		`{}`,
		``,
		`   `,
		`{not json`,
		`{"level":42}`,
		`{"level":"info","msg":null}`,
		`{"nested":{"a":1}}`,
	}
	for _, s := range seeds {
		f.Add(s)
	}

	p := parser.NewJSONParser(parser.DefaultOptions())
	f.Fuzz(func(t *testing.T, line string) {
		entry, err := p.Parse(line)
		switch {
		case err == nil, errors.Is(err, parser.ErrSkip):
			// ok
		default:
			// Any error must carry the original line in Raw so the caller can
			// surface it in diagnostics.
			if entry.Raw != line && line != "" {
				t.Errorf("error path lost Raw: got %q, want %q", entry.Raw, line)
			}
		}
	})
}

// FuzzTextParser_Parse exercises the text parser. Non-matching lines should
// still come back as a usable Entry with the line preserved as the message.
func FuzzTextParser_Parse(f *testing.F) {
	seeds := []string{
		`2024-01-01 12:00:00 INFO hello`,
		``,
		`unstructured`,
		"line with\ttab",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	p := parser.NewTextParser(parser.DefaultOptions())
	f.Fuzz(func(t *testing.T, line string) {
		entry, err := p.Parse(line)
		if err != nil && !errors.Is(err, parser.ErrSkip) {
			t.Errorf("text parser unexpectedly errored on %q: %v", line, err)
		}
		_ = entry
	})
}
