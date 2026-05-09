package filter_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/hermanu/logz/internal/filter"
	"github.com/hermanu/logz/internal/parser"
)

func entry(level parser.Level, msg string, ts time.Time, fields map[string]string) parser.Entry {
	return parser.Entry{
		Timestamp: ts,
		Level:     level,
		Message:   msg,
		Fields:    fields,
		Raw:       msg,
	}
}

func TestFilter_Allow(t *testing.T) {
	t.Parallel()
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := map[string]struct {
		f    filter.Filter
		e    parser.Entry
		want bool
	}{
		"empty filter allows everything": {
			f:    filter.Filter{},
			e:    entry(parser.LevelDebug, "x", now, nil),
			want: true,
		},
		"min level drops below": {
			f:    filter.Filter{MinLevel: parser.LevelWarn},
			e:    entry(parser.LevelInfo, "x", now, nil),
			want: false,
		},
		"min level allows equal": {
			f:    filter.Filter{MinLevel: parser.LevelWarn},
			e:    entry(parser.LevelWarn, "x", now, nil),
			want: true,
		},
		"since drops earlier": {
			f:    filter.Filter{Since: now},
			e:    entry(parser.LevelInfo, "x", now.Add(-time.Hour), nil),
			want: false,
		},
		"until drops later": {
			f:    filter.Filter{Until: now},
			e:    entry(parser.LevelInfo, "x", now.Add(time.Hour), nil),
			want: false,
		},
		"match keyword in message": {
			f:    filter.Filter{Match: regexp.MustCompile("timeout")},
			e:    entry(parser.LevelError, "request timeout after 5s", now, nil),
			want: true,
		},
		"match misses": {
			f:    filter.Filter{Match: regexp.MustCompile("nonexistent")},
			e:    entry(parser.LevelError, "all good", now, nil),
			want: false,
		},
		"field equals": {
			f:    filter.Filter{FieldEquals: map[string]string{"service": "api"}},
			e:    entry(parser.LevelInfo, "x", now, map[string]string{"service": "api"}),
			want: true,
		},
		"field mismatch": {
			f:    filter.Filter{FieldEquals: map[string]string{"service": "api"}},
			e:    entry(parser.LevelInfo, "x", now, map[string]string{"service": "auth"}),
			want: false,
		},
		"invert flips": {
			f:    filter.Filter{MinLevel: parser.LevelError, Invert: true},
			e:    entry(parser.LevelInfo, "x", now, nil),
			want: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := tc.f.Allow(tc.e)
			if got != tc.want {
				t.Errorf("Allow: got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseFieldSpec(t *testing.T) {
	t.Parallel()
	k, v, err := filter.ParseFieldSpec("service=api")
	if err != nil || k != "service" || v != "api" {
		t.Errorf("got (%q,%q,%v)", k, v, err)
	}

	if _, _, err := filter.ParseFieldSpec("noequals"); err == nil {
		t.Error("expected error on missing =")
	}
}

func TestParseSince(t *testing.T) {
	t.Parallel()
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("relative", func(t *testing.T) {
		t.Parallel()
		got, err := filter.ParseSince("1h", now)
		if err != nil {
			t.Fatal(err)
		}
		want := now.Add(-time.Hour)
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("absolute RFC3339", func(t *testing.T) {
		t.Parallel()
		got, err := filter.ParseSince("2024-01-01T00:00:00Z", now)
		if err != nil {
			t.Fatal(err)
		}
		if got.IsZero() {
			t.Error("not parsed")
		}
	})

	t.Run("empty returns zero", func(t *testing.T) {
		t.Parallel()
		got, err := filter.ParseSince("", now)
		if err != nil || !got.IsZero() {
			t.Errorf("got (%v,%v)", got, err)
		}
	})

	t.Run("garbage errors", func(t *testing.T) {
		t.Parallel()
		if _, err := filter.ParseSince("not a time", now); err == nil {
			t.Error("expected error")
		}
	})
}
