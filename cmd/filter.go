package cmd

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/spf13/cobra"

	"github.com/hermanu/logz/internal/filter"
	"github.com/hermanu/logz/internal/parser"
)

// filterFlags is the bag of values populated by cobra for `logz filter`.
type filterFlags struct {
	level  string
	since  string
	until  string
	match  string
	fields []string
	limit  int
	invert bool
}

func newFilterCmd() *cobra.Command {
	flags := &filterFlags{}

	cmd := &cobra.Command{
		Use:   "filter [flags] [file...]",
		Short: "Filter and display log lines",
		Long:  "Filter log lines by level, time range, keyword, or field value. Reads from stdin if no file is given.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFilter(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.level, "level", "l", "", "minimum level: trace, debug, info, warn, error, fatal")
	cmd.Flags().StringVar(&flags.since, "since", "", "show logs after this time (1h, 30m, or RFC3339)")
	cmd.Flags().StringVar(&flags.until, "until", "", "show logs before this time")
	cmd.Flags().StringVarP(&flags.match, "match", "m", "", "regex pattern to match against the line or message")
	cmd.Flags().StringSliceVar(&flags.fields, "field", nil, "filter by field, e.g. --field service=auth (repeatable)")
	cmd.Flags().IntVarP(&flags.limit, "limit", "n", 0, "max number of matching lines to output (0 = no limit)")
	cmd.Flags().BoolVarP(&flags.invert, "invert", "v", false, "invert the match (like grep -v)")

	return cmd
}

func runFilter(cmd *cobra.Command, args []string, flags *filterFlags) error {
	g := getGlobals(cmd.Context())
	cfg, err := loadConfig(g.cfgPath)
	if err != nil {
		return err
	}
	opts, err := cfg.ParserOptions()
	if err != nil {
		return err
	}

	f, err := buildFilter(flags)
	if err != nil {
		return err
	}

	out, err := pickOutput(cfg, g.output, cmd.OutOrStdout(), g.noColor)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	sources, err := resolveSources(args, cmd.InOrStdin())
	if err != nil {
		return err
	}

	var matched int
	for _, src := range sources {
		remaining := 0
		if flags.limit > 0 {
			remaining = flags.limit - matched
			if remaining <= 0 {
				break
			}
		}
		n, err := runFilterPipeline(src, g.format, opts, f, out, remaining, cmd.ErrOrStderr())
		if closeErr := src.rc.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		matched += n
		if err != nil {
			return err
		}
	}

	if matched == 0 {
		return errNoMatches
	}
	return nil
}

// buildFilter materializes a [filter.Filter] from CLI flag values.
func buildFilter(flags *filterFlags) (*filter.Filter, error) {
	f := &filter.Filter{Invert: flags.invert}
	if flags.level != "" {
		lvl, ok := parser.ParseLevel(flags.level)
		if !ok {
			return nil, fmt.Errorf("unknown --level %q", flags.level)
		}
		f.MinLevel = lvl
	}
	now := time.Now()
	since, err := filter.ParseSince(flags.since, now)
	if err != nil {
		return nil, err
	}
	f.Since = since
	until, err := filter.ParseSince(flags.until, now)
	if err != nil {
		return nil, err
	}
	f.Until = until
	if flags.match != "" {
		re, reErr := regexp.Compile(flags.match)
		if reErr != nil {
			return nil, fmt.Errorf("--match: %w", reErr)
		}
		f.Match = re
	}
	if len(flags.fields) > 0 {
		f.FieldEquals = make(map[string]string, len(flags.fields))
		for _, spec := range flags.fields {
			k, v, specErr := filter.ParseFieldSpec(spec)
			if specErr != nil {
				return nil, specErr
			}
			f.FieldEquals[k] = v
		}
	}
	return f, nil
}

// runFilterPipeline reads from src, parses, filters, and writes matches.
// limit ≤ 0 means unlimited.
func runFilterPipeline(src namedSource, formatOverride string, opts parser.Options, f *filter.Filter, out interface {
	Write(parser.Entry) error
}, limit int, stderr io.Writer,
) (int, error) {
	const sniffSize = 16 * 1024
	sample, body, err := parser.SniffSample(src.rc, sniffSize)
	if err != nil {
		return 0, fmt.Errorf("sniff %s: %w", src.name, err)
	}
	p, err := pickParser(formatOverride, sample, opts)
	if err != nil {
		return 0, err
	}

	var matched, skipped int
	streamErr := streamLines(body, func(line string) error {
		entry, perr := p.Parse(line)
		if perr != nil {
			if errors.Is(perr, parser.ErrSkip) {
				return nil
			}
			skipped++
			return nil
		}
		if !f.Allow(entry) {
			return nil
		}
		matched++
		if writeErr := out.Write(entry); writeErr != nil {
			return writeErr
		}
		if limit > 0 && matched >= limit {
			return io.EOF // sentinel to stop early
		}
		return nil
	})
	if streamErr != nil && !errors.Is(streamErr, io.EOF) {
		return matched, fmt.Errorf("read %s: %w", src.name, streamErr)
	}
	if skipped > 0 {
		fmt.Fprintf(stderr, "logz: %d unparseable line(s) skipped in %s\n", skipped, src.name)
	}
	return matched, nil
}
