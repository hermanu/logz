package parser_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/hermanu/logz/internal/parser"
)

func TestJSONParser_Parse(t *testing.T) {
	t.Parallel()

	type want struct {
		entry parser.Entry
		err   error // matched with errors.Is; nil means no error
	}

	tests := []struct {
		name string
		in   string
		opts parser.Options
		want want
	}{
		{
			name: "happy path",
			in:   `{"level":"error","msg":"boom","ts":"2024-01-01T12:00:00Z","service":"api"}`,
			opts: parser.DefaultOptions(),
			want: want{entry: parser.Entry{
				Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				Level:     parser.LevelError,
				Message:   "boom",
				Fields: map[string]string{
					"level": "error", "msg": "boom",
					"ts": "2024-01-01T12:00:00Z", "service": "api",
				},
			}},
		},
		{
			name: "blank line skipped",
			in:   "   ",
			opts: parser.DefaultOptions(),
			want: want{err: parser.ErrSkip},
		},
		{
			name: "numeric values stringified",
			in:   `{"level":"info","msg":"x","retry":3,"factor":1.5}`,
			opts: parser.DefaultOptions(),
			want: want{entry: parser.Entry{
				Level:   parser.LevelInfo,
				Message: "x",
				Fields: map[string]string{
					"level": "info", "msg": "x",
					"retry": "3", "factor": "1.5",
				},
			}},
		},
		{
			name: "level alias",
			in:   `{"level":"30","msg":"hi"}`,
			opts: func() parser.Options {
				o := parser.DefaultOptions()
				o.LevelAliases = map[string]parser.Level{"30": parser.LevelInfo}
				return o
			}(),
			want: want{entry: parser.Entry{
				Level:   parser.LevelInfo,
				Message: "hi",
				Fields:  map[string]string{"level": "30", "msg": "hi"},
			}},
		},
		{
			name: "custom keys",
			in:   `{"severity":"warn","log":"slow","@timestamp":"2024-01-01T00:00:00Z"}`,
			opts: parser.Options{
				LevelKeys:     []string{"severity"},
				MessageKeys:   []string{"log"},
				TimestampKeys: []string{"@timestamp"},
			},
			want: want{entry: parser.Entry{
				Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Level:     parser.LevelWarn,
				Message:   "slow",
				Fields: map[string]string{
					"severity": "warn", "log": "slow",
					"@timestamp": "2024-01-01T00:00:00Z",
				},
			}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := parser.NewJSONParser(tc.opts)
			got, err := p.Parse(tc.in)

			switch {
			case tc.want.err != nil:
				if !errors.Is(err, tc.want.err) {
					t.Fatalf("err: got %v, want errors.Is(_, %v)", err, tc.want.err)
				}
				return
			case err != nil:
				t.Fatalf("unexpected error: %v", err)
			}

			// Raw is the input verbatim and not interesting for the diff.
			diff := cmp.Diff(tc.want.entry, got,
				cmpopts.IgnoreFields(parser.Entry{}, "Raw"))
			if diff != "" {
				t.Errorf("Entry mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestJSONParser_Malformed(t *testing.T) {
	t.Parallel()
	p := parser.NewJSONParser(parser.DefaultOptions())

	in := `{not json`
	_, err := p.Parse(in)
	if err == nil || errors.Is(err, parser.ErrSkip) {
		t.Errorf("got %v, want a parse error", err)
	}
}
