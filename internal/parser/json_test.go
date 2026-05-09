package parser_test

import (
	"errors"
	"testing"
	"time"

	"github.com/hermanu/logz/internal/parser"
)

func TestJSONParser_Parse(t *testing.T) {
	t.Parallel()
	p := parser.NewJSONParser(parser.DefaultOptions())

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		e, err := p.Parse(`{"level":"error","msg":"boom","ts":"2024-01-01T12:00:00Z","service":"api"}`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.Level != parser.LevelError {
			t.Errorf("level: got %v, want error", e.Level)
		}
		if e.Message != "boom" {
			t.Errorf("message: got %q, want boom", e.Message)
		}
		if e.Timestamp.IsZero() {
			t.Error("timestamp not parsed")
		} else if !e.Timestamp.Equal(time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)) {
			t.Errorf("timestamp: got %v", e.Timestamp)
		}
		if e.Fields["service"] != "api" {
			t.Errorf("fields[service]: got %q", e.Fields["service"])
		}
	})

	t.Run("blank line skipped", func(t *testing.T) {
		t.Parallel()
		_, err := p.Parse("   ")
		if !errors.Is(err, parser.ErrSkip) {
			t.Errorf("got %v, want ErrSkip", err)
		}
	})

	t.Run("malformed surfaces error", func(t *testing.T) {
		t.Parallel()
		_, err := p.Parse(`{not json`)
		if err == nil || errors.Is(err, parser.ErrSkip) {
			t.Errorf("got %v, want parse error", err)
		}
	})

	t.Run("numeric values stringified", func(t *testing.T) {
		t.Parallel()
		e, err := p.Parse(`{"level":"info","msg":"x","retry":3,"factor":1.5}`)
		if err != nil {
			t.Fatal(err)
		}
		if e.Fields["retry"] != "3" {
			t.Errorf("retry: got %q, want 3", e.Fields["retry"])
		}
		if e.Fields["factor"] != "1.5" {
			t.Errorf("factor: got %q, want 1.5", e.Fields["factor"])
		}
	})

	t.Run("level alias", func(t *testing.T) {
		t.Parallel()
		opts := parser.DefaultOptions()
		opts.LevelAliases = map[string]parser.Level{"30": parser.LevelInfo}
		ap := parser.NewJSONParser(opts)
		e, err := ap.Parse(`{"level":"30","msg":"hi"}`)
		if err != nil {
			t.Fatal(err)
		}
		if e.Level != parser.LevelInfo {
			t.Errorf("alias not applied: got %v", e.Level)
		}
	})

	t.Run("custom keys", func(t *testing.T) {
		t.Parallel()
		opts := parser.DefaultOptions()
		opts.LevelKeys = []string{"severity"}
		opts.MessageKeys = []string{"log"}
		opts.TimestampKeys = []string{"@timestamp"}
		cp := parser.NewJSONParser(opts)
		e, err := cp.Parse(`{"severity":"warn","log":"slow","@timestamp":"2024-01-01T00:00:00Z"}`)
		if err != nil {
			t.Fatal(err)
		}
		if e.Level != parser.LevelWarn || e.Message != "slow" || e.Timestamp.IsZero() {
			t.Errorf("custom keys not honored: %+v", e)
		}
	})
}
