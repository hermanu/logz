package parser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// JSONParser parses one JSON object per line. Field discovery is configurable
// via [Options].
type JSONParser struct {
	opts Options
}

// NewJSONParser returns a JSONParser configured with opts. The opts struct is
// copied; defaults are filled in for empty slice fields.
func NewJSONParser(opts Options) *JSONParser {
	d := DefaultOptions()
	if len(opts.LevelKeys) == 0 {
		opts.LevelKeys = d.LevelKeys
	}
	if len(opts.MessageKeys) == 0 {
		opts.MessageKeys = d.MessageKeys
	}
	if len(opts.TimestampKeys) == 0 {
		opts.TimestampKeys = d.TimestampKeys
	}
	return &JSONParser{opts: opts}
}

// Name implements [Parser].
func (*JSONParser) Name() string { return "json" }

// Parse implements [Parser]. Returns [ErrSkip] for blank lines.
func (p *JSONParser) Parse(line string) (Entry, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return Entry{}, ErrSkip
	}

	var raw map[string]any
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return Entry{Raw: line}, fmt.Errorf("json: %w", err)
	}

	e := Entry{
		Raw:    line,
		Fields: make(map[string]string, len(raw)),
	}
	for k, v := range raw {
		e.Fields[k] = stringify(v)
	}

	for _, k := range p.opts.LevelKeys {
		if v, ok := e.Fields[k]; ok {
			if lvl, ok := p.resolveLevel(v); ok {
				e.Level = lvl
				break
			}
		}
	}
	for _, k := range p.opts.MessageKeys {
		if v, ok := e.Fields[k]; ok {
			e.Message = v
			break
		}
	}
	for _, k := range p.opts.TimestampKeys {
		if v, ok := e.Fields[k]; ok {
			if ts := parseTime(v, p.opts.TimestampLayouts); !ts.IsZero() {
				e.Timestamp = ts
				break
			}
		}
	}
	return e, nil
}

func (p *JSONParser) resolveLevel(s string) (Level, bool) {
	if alias, ok := p.opts.LevelAliases[strings.ToLower(strings.TrimSpace(s))]; ok {
		return alias, true
	}
	return ParseLevel(s)
}

// stringify normalizes any JSON-decoded value into a string for the Fields map.
func stringify(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case float64:
		// JSON numbers are float64 by default; trim ".0" for integers.
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", t)
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return fmt.Sprintf("%v", t)
		}
		return string(b)
	}
}
