package output_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/hermanu/logz/internal/output"
	"github.com/hermanu/logz/internal/parser"
)

func TestJSONWriter(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	w := output.NewJSON(&buf)

	e := parser.Entry{
		Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Level:     parser.LevelError,
		Message:   "boom",
		Fields:    map[string]string{"service": "api"},
		Raw:       `{"level":"error","msg":"boom"}`,
	}
	if err := w.Write(e); err != nil {
		t.Fatal(err)
	}

	out := strings.TrimSpace(buf.String())
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	if got["level"] != "error" {
		t.Errorf("level: got %v", got["level"])
	}
	if got["message"] != "boom" {
		t.Errorf("message: got %v", got["message"])
	}
}

func TestJSONWriter_OmitsZeroLevel(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	w := output.NewJSON(&buf)
	if err := w.Write(parser.Entry{Message: "hi", Raw: "hi"}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), `"level"`) {
		t.Errorf("expected no level key for unknown level: %s", buf.String())
	}
}
