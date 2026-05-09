package parser_test

import (
	"testing"

	"github.com/hermanu/logz/internal/parser"
)

func TestParseLevel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want parser.Level
		ok   bool
	}{
		{"INFO", parser.LevelInfo, true},
		{"info", parser.LevelInfo, true},
		{" Warning ", parser.LevelWarn, true},
		{"err", parser.LevelError, true},
		{"panic", parser.LevelFatal, true},
		{"weird", parser.LevelUnknown, false},
		{"", parser.LevelUnknown, false},
	}
	for _, tc := range cases {
		got, ok := parser.ParseLevel(tc.in)
		if got != tc.want || ok != tc.ok {
			t.Errorf("ParseLevel(%q) = (%v,%v), want (%v,%v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestLevelOrdering(t *testing.T) {
	t.Parallel()
	if parser.LevelDebug >= parser.LevelInfo {
		t.Error("LevelDebug should be < LevelInfo")
	}
	if parser.LevelError <= parser.LevelWarn {
		t.Error("LevelError should be > LevelWarn")
	}
}

func TestDetect(t *testing.T) {
	t.Parallel()
	opts := parser.DefaultOptions()

	tests := map[string]struct {
		sample   string
		wantName string
	}{
		"json":         {`{"level":"info","msg":"hi"}`, "json"},
		"text":         {`2024-01-01 12:00:00 INFO hello`, "text"},
		"empty_to_text": {``, "text"},
		"unknown":      {`???`, "text"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			p := parser.Detect([]byte(tc.sample), opts)
			if p.Name() != tc.wantName {
				t.Errorf("Detect: got %s, want %s", p.Name(), tc.wantName)
			}
		})
	}
}
