package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// DefaultTextPattern matches the common
// "<timestamp> <LEVEL> <message>" shape. Named groups: timestamp, level,
// message. Any non-matching line is preserved as Raw with no extracted fields.
//
// Examples it handles:
//
//	2024-01-01 12:00:00 ERROR timeout
//	2024-01-01T12:00:00Z WARN  slow query
const DefaultTextPattern = `^(?P<timestamp>\d{4}[-/]\d{2}[-/]\d{2}[T ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?)\s+(?P<level>[A-Za-z]+)\s+(?P<message>.*)$`

// TextParser extracts structure from plain-text logs using a configurable
// regular expression. The expression must use named capture groups; recognized
// names are "timestamp", "level", and "message". Additional named groups are
// stored in Fields.
type TextParser struct {
	opts    Options
	pattern *regexp.Regexp
	names   []string
}

// NewTextParser builds a TextParser using the default pattern.
func NewTextParser(opts Options) *TextParser {
	p, err := NewTextParserWithPattern(DefaultTextPattern, opts)
	if err != nil {
		// DefaultTextPattern is a constant; a compile error would be a bug.
		panic(fmt.Sprintf("parser: invalid DefaultTextPattern: %v", err))
	}
	return p
}

// NewTextParserWithPattern compiles pattern as the line regex. Returns an
// error if the pattern is invalid.
func NewTextParserWithPattern(pattern string, opts Options) (*TextParser, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("text parser: compile pattern: %w", err)
	}
	return &TextParser{
		opts:    opts,
		pattern: re,
		names:   re.SubexpNames(),
	}, nil
}

// Name implements [Parser].
func (*TextParser) Name() string { return "text" }

// Parse implements [Parser]. Lines that don't match the pattern are returned
// with the message set to the raw line and level unknown — they're still
// usable for `--match` keyword filtering.
func (p *TextParser) Parse(line string) (Entry, error) {
	line = strings.TrimRight(line, "\r\n")
	if strings.TrimSpace(line) == "" {
		return Entry{}, ErrSkip
	}

	e := Entry{Raw: line, Message: line, Fields: map[string]string{}}

	match := p.pattern.FindStringSubmatch(line)
	if match == nil {
		return e, nil
	}

	for i, name := range p.names {
		if i == 0 || name == "" {
			continue
		}
		val := match[i]
		switch name {
		case "timestamp":
			if ts := parseTime(val, p.opts.TimestampLayouts); !ts.IsZero() {
				e.Timestamp = ts
			}
			e.Fields[name] = val
		case "level":
			if lvl, ok := p.resolveLevel(val); ok {
				e.Level = lvl
			}
			e.Fields[name] = val
		case "message":
			e.Message = val
		default:
			e.Fields[name] = val
		}
	}
	return e, nil
}

func (p *TextParser) resolveLevel(s string) (Level, bool) {
	if alias, ok := p.opts.LevelAliases[strings.ToLower(strings.TrimSpace(s))]; ok {
		return alias, true
	}
	return ParseLevel(s)
}
