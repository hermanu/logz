package summary_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/hermanu/logz/internal/parser"
	"github.com/hermanu/logz/internal/summary"
)

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
