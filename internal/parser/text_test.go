package parser_test

import (
	"testing"

	"github.com/hermanu/logz/internal/parser"
)

func TestTextParser_DefaultPattern(t *testing.T) {
	t.Parallel()
	p := parser.NewTextParser(parser.DefaultOptions())

	t.Run("matches", func(t *testing.T) {
		t.Parallel()
		e, err := p.Parse("2024-01-01 12:00:00 ERROR timeout occurred")
		if err != nil {
			t.Fatal(err)
		}
		if e.Level != parser.LevelError {
			t.Errorf("level: got %v", e.Level)
		}
		if e.Message != "timeout occurred" {
			t.Errorf("message: got %q", e.Message)
		}
		if e.Timestamp.IsZero() {
			t.Error("timestamp not parsed")
		}
	})

	t.Run("non-matching keeps raw as message", func(t *testing.T) {
		t.Parallel()
		e, err := p.Parse("just a random line")
		if err != nil {
			t.Fatal(err)
		}
		if e.Message != "just a random line" {
			t.Errorf("message: got %q", e.Message)
		}
		if e.Level != parser.LevelUnknown {
			t.Errorf("expected unknown level, got %v", e.Level)
		}
	})
}

func TestTextParser_CustomPattern(t *testing.T) {
	t.Parallel()
	pattern := `^\[(?P<level>[A-Z]+)\]\s+(?P<service>\S+)\s+(?P<message>.*)$`
	p, err := parser.NewTextParserWithPattern(pattern, parser.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	e, perr := p.Parse("[WARN] payments slow query")
	if perr != nil {
		t.Fatal(perr)
	}
	if e.Level != parser.LevelWarn {
		t.Errorf("level: got %v", e.Level)
	}
	if e.Fields["service"] != "payments" {
		t.Errorf("service: got %q", e.Fields["service"])
	}
	if e.Message != "slow query" {
		t.Errorf("message: got %q", e.Message)
	}
}

func TestTextParser_InvalidPattern(t *testing.T) {
	t.Parallel()
	if _, err := parser.NewTextParserWithPattern("(", parser.DefaultOptions()); err == nil {
		t.Error("expected compile error")
	}
}
