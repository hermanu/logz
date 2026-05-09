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
// disabled regardless of TTY detection. Colors are configured per-instance
// (no mutation of fatih/color's package-level state), so concurrent writers
// in tests don't race.
func NewPretty(w io.Writer, noColor bool) *PrettyWriter {
	mk := func(attrs ...color.Attribute) *color.Color {
		c := color.New(attrs...)
		if noColor {
			c.DisableColor()
		}
		return c
	}
	return &PrettyWriter{
		w:       w,
		noColor: noColor,
		dim:     mk(color.Faint),
		colors: map[parser.Level]*color.Color{
			parser.LevelTrace: mk(color.FgHiBlack),
			parser.LevelDebug: mk(color.FgBlue),
			parser.LevelInfo:  mk(color.FgGreen),
			parser.LevelWarn:  mk(color.FgYellow),
			parser.LevelError: mk(color.FgRed),
			parser.LevelFatal: mk(color.FgRed, color.Bold),
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
