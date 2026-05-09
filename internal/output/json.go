package output

import (
	"encoding/json"
	"io"
	"time"

	"github.com/hermanu/logz/internal/parser"
)

// JSONWriter emits one JSON object per entry (NDJSON). Suitable for piping
// into jq.
type JSONWriter struct {
	enc *json.Encoder
}

// NewJSON constructs a JSONWriter writing to w.
func NewJSON(w io.Writer) *JSONWriter {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return &JSONWriter{enc: enc}
}

// jsonEntry is the wire shape — keeps zero-valued fields out of the output.
type jsonEntry struct {
	Timestamp *time.Time        `json:"timestamp,omitempty"`
	Level     string            `json:"level,omitempty"`
	Message   string            `json:"message,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
	Raw       string            `json:"raw,omitempty"`
}

// Write renders one entry as a JSON object on its own line.
func (j *JSONWriter) Write(e parser.Entry) error {
	out := jsonEntry{
		Level:   e.Level.String(),
		Message: e.Message,
		Fields:  e.Fields,
		Raw:     e.Raw,
	}
	if !e.Timestamp.IsZero() {
		ts := e.Timestamp
		out.Timestamp = &ts
	}
	if e.Level == parser.LevelUnknown {
		out.Level = ""
	}
	return j.enc.Encode(out)
}

// Close is a no-op; included to satisfy [Writer].
func (*JSONWriter) Close() error { return nil }
