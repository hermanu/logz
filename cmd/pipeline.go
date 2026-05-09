package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/hermanu/logz/internal/config"
	"github.com/hermanu/logz/internal/output"
	"github.com/hermanu/logz/internal/parser"
)

// resolveSources returns one io.ReadCloser per input. An empty paths list
// yields a single source reading from stdin.
func resolveSources(paths []string, stdin io.Reader) ([]namedSource, error) {
	if len(paths) == 0 {
		return []namedSource{{name: "stdin", rc: io.NopCloser(stdin)}}, nil
	}
	sources := make([]namedSource, 0, len(paths))
	for _, p := range paths {
		f, err := os.Open(p) //nolint:gosec // G304 is expected for log file handling
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", p, err)
		}
		sources = append(sources, namedSource{name: p, rc: f})
	}
	return sources, nil
}

type namedSource struct {
	name string
	rc   io.ReadCloser
}

// pickParser chooses a parser based on (in order): explicit format flag,
// sniffed sample, fall-through to text.
func pickParser(formatOverride string, sample []byte, opts parser.Options) (parser.Parser, error) {
	switch formatOverride {
	case "":
		return parser.Detect(sample, opts), nil
	case "json":
		return parser.NewJSONParser(opts), nil
	case "text":
		return parser.NewTextParser(opts), nil
	default:
		return nil, fmt.Errorf("unknown --format %q (want: json, text)", formatOverride)
	}
}

// pickOutput resolves output mode, honoring config + flag override.
func pickOutput(cfg config.Config, flag string, w io.Writer, noColor bool) (output.Writer, error) {
	mode := cfg.Output
	if flag != "" {
		mode = flag
	}
	switch mode {
	case "", "pretty":
		return output.New(output.ModePretty, w, noColor), nil
	case "json":
		return output.New(output.ModeJSON, w, noColor), nil
	default:
		return nil, fmt.Errorf("unknown --output %q (want: pretty, json)", mode)
	}
}

// streamLines reads from rc and yields one line at a time via cb. Lines beyond
// the bufio default are handled by raising the buffer ceiling to 10MB.
func streamLines(rc io.Reader, cb func(line string) error) error {
	br := bufio.NewReaderSize(rc, 64*1024)
	for {
		line, err := br.ReadString('\n')
		if line != "" {
			// Strip trailing newline (and CR on Windows).
			n := len(line)
			if n > 0 && line[n-1] == '\n' {
				n--
				if n > 0 && line[n-1] == '\r' {
					n--
				}
			}
			if cbErr := cb(line[:n]); cbErr != nil {
				return cbErr
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}
