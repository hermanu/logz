package summary_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/hermanu/logz/internal/parser"
	"github.com/hermanu/logz/internal/summary"
)

func TestSummary_Counts(t *testing.T) {
	t.Parallel()
	s := summary.New(3)

	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		s.Observe(parser.Entry{
			Timestamp: base.Add(time.Duration(i) * time.Second),
			Level:     parser.LevelInfo,
			Message:   "ok",
		})
	}
	for i := 0; i < 3; i++ {
		s.Observe(parser.Entry{
			Timestamp: base.Add(time.Duration(10+i) * time.Second),
			Level:     parser.LevelError,
			Message:   "boom",
			Fields:    map[string]string{"service": "api"},
		})
	}
	s.ObserveSkip()

	if got, want := s.Total(), 8; got != want {
		t.Errorf("Total: got %d, want %d", got, want)
	}
	if got, want := s.Skipped(), 1; got != want {
		t.Errorf("Skipped: got %d, want %d", got, want)
	}
	if got, want := s.LevelCount(parser.LevelInfo), 5; got != want {
		t.Errorf("LevelCount(info): got %d, want %d", got, want)
	}
	if got, want := s.LevelCount(parser.LevelError), 3; got != want {
		t.Errorf("LevelCount(error): got %d, want %d", got, want)
	}

	first, last := s.TimeRange()
	if !first.Equal(base) {
		t.Errorf("First: got %v, want %v", first, base)
	}
	if !last.Equal(base.Add(12 * time.Second)) {
		t.Errorf("Last: got %v, want %v", last, base.Add(12*time.Second))
	}
}

func TestSummary_Render(t *testing.T) {
	t.Parallel()
	s := summary.New(3)

	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		s.Observe(parser.Entry{
			Timestamp: base.Add(time.Duration(i) * time.Second),
			Level:     parser.LevelInfo,
			Message:   "ok",
		})
	}
	for i := 0; i < 3; i++ {
		s.Observe(parser.Entry{
			Timestamp: base.Add(time.Duration(10+i) * time.Second),
			Level:     parser.LevelError,
			Message:   "boom",
			Fields:    map[string]string{"service": "api"},
		})
	}
	s.ObserveSkip()

	var buf bytes.Buffer
	if err := s.Render(&buf, 3); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	for _, want := range []string{"Total lines:   8", "Unparseable:   1", "info    5", "error   3", "boom"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestSummary_EmptyRender(t *testing.T) {
	t.Parallel()
	s := summary.New(0) // zero falls through to default 10
	var buf bytes.Buffer
	if err := s.Render(&buf, 0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Total lines:   0") {
		t.Errorf("empty summary: got %q", buf.String())
	}
}

func BenchmarkSummary_Observe(b *testing.B) {
	s := summary.New(10)
	e := parser.Entry{
		Timestamp: time.Now(),
		Level:     parser.LevelError,
		Message:   "boom",
		Fields:    map[string]string{"service": "api", "request_id": "abc123"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Observe(e)
	}
}
