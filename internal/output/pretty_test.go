package output_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/hermanu/logz/internal/output"
	"github.com/hermanu/logz/internal/parser"
)

func TestPrettyWriter(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	w := output.NewPretty(&buf, true) // no color, deterministic output

	e := parser.Entry{
		Timestamp: time.Date(2024, 1, 1, 12, 0, 1, 0, time.UTC),
		Level:     parser.LevelError,
		Message:   "timeout",
		Fields:    map[string]string{"service": "payments", "retry": "3", "msg": "timeout"},
	}
	if err := w.Write(e); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	for _, want := range []string{"2024-01-01 12:00:01", "ERROR", "timeout", "service=payments", "retry=3"} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q\nfull: %s", want, got)
		}
	}
	// "msg" is hidden because it's already rendered as the message column.
	if strings.Contains(got, "msg=timeout") {
		t.Errorf("msg= should be hidden:\n%s", got)
	}
}
