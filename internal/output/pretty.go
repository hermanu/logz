package output

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/fatih/color"

	"github.com/hermanu/logz/internal/parser"
)

// PrettyWriter emits human-readable, colored output.
type PrettyWriter struct {
	w        io.Writer
	noColor  bool
	colors   map[parser.Level]*color.Color
	dim      *color.Color
	hideKeys map[string]struct{}
}

// NewPretty constructs a PrettyWriter. If noColor is true, all colors are
// disabled regardless of TTY detection.
func NewPretty(w io.Writer, noColor bool) *PrettyWriter {
	if noColor {
		color.NoColor = true
	}
	return &PrettyWriter{
		w:       w,
		noColor: noColor,
		dim:     color.New(color.Faint),
		colors: map[parser.Level]*color.Color{
			parser.LevelTrace: color.New(color.FgHiBlack),
			parser.LevelDebug: color.New(color.FgBlue),
			parser.LevelInfo:  color.New(color.FgGreen),
			parser.LevelWarn:  color.New(color.FgYellow),
			parser.LevelError: color.New(color.FgRed),
			parser.LevelFatal: color.New(color.FgRed, color.Bold),
		},
		// Fields already rendered as separate columns shouldn't appear twice.
		hideKeys: map[string]struct{}{
			"level": {}, "lvl": {}, "severity": {},
			"msg": {}, "message": {}, "log": {},
			"ts": {}, "time": {}, "timestamp": {}, "@timestamp": {},
		},
	}
}

// Write renders one entry.
func (p *PrettyWriter) Write(e parser.Entry) error {
	var b strings.Builder

	if !e.Timestamp.IsZero() {
		_, _ = p.dim.Fprintf(&b, "%s  ", e.Timestamp.Format("2006-01-02 15:04:05"))
	}

	level := e.Level
	if level == parser.LevelUnknown {
		level = parser.LevelInfo
	}
	if c, ok := p.colors[level]; ok {
		_, _ = c.Fprintf(&b, "%-5s ", strings.ToUpper(e.Level.String()))
	} else {
		fmt.Fprintf(&b, "%-5s ", strings.ToUpper(e.Level.String()))
	}

	if e.Message != "" {
		b.WriteString(e.Message)
	} else {
		b.WriteString(e.Raw)
	}

	keys := make([]string, 0, len(e.Fields))
	for k := range e.Fields {
		if _, hide := p.hideKeys[k]; hide {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) > 0 {
		b.WriteString("  ")
		for i, k := range keys {
			if i > 0 {
				b.WriteByte(' ')
			}
			_, _ = p.dim.Fprintf(&b, "%s=%s", k, e.Fields[k])
		}
	}

	b.WriteByte('\n')
	_, err := io.WriteString(p.w, b.String())
	return err
}

// Close is a no-op; included to satisfy [Writer].
func (*PrettyWriter) Close() error { return nil }
