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

func newFilterCmd() *cobra.Command {
	var (
		levelStr string
		since    string
		until    string
		match    string
		fields   []string
		limit    int
		invert   bool
	)

	cmd := &cobra.Command{
		Use:   "filter [flags] [file...]",
		Short: "Filter and display log lines",
		Long:  "Filter log lines by level, time range, keyword, or field value. Reads from stdin if no file is given.",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := getGlobals(cmd.Context())
			cfg, err := loadConfig(g.cfgPath)
			if err != nil {
				return err
			}
			opts, err := cfg.ParserOptions()
			if err != nil {
				return err
			}

			f := &filter.Filter{Invert: invert}
			if levelStr != "" {
				lvl, ok := parser.ParseLevel(levelStr)
				if !ok {
					return fmt.Errorf("unknown --level %q", levelStr)
				}
				f.MinLevel = lvl
			}
			now := time.Now()
			if f.Since, err = filter.ParseSince(since, now); err != nil {
				return err
			}
			if f.Until, err = filter.ParseSince(until, now); err != nil {
				return err
			}
			if match != "" {
				re, err := regexp.Compile(match)
				if err != nil {
					return fmt.Errorf("--match: %w", err)
				}
				f.Match = re
			}
			if len(fields) > 0 {
				f.FieldEquals = make(map[string]string, len(fields))
				for _, spec := range fields {
					k, v, err := filter.ParseFieldSpec(spec)
					if err != nil {
						return err
					}
					f.FieldEquals[k] = v
				}
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

			var matched, written int
			for _, src := range sources {
				n, err := runFilterPipeline(src, g.format, opts, f, out, limit-written, cmd.ErrOrStderr())
				if closeErr := src.rc.Close(); closeErr != nil && err == nil {
					err = closeErr
				}
				written += n
				matched += n
				if err != nil {
					return err
				}
				if limit > 0 && written >= limit {
					break
				}
			}

			if matched == 0 {
				return errNoMatches
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&levelStr, "level", "l", "", "minimum level: trace, debug, info, warn, error, fatal")
	cmd.Flags().StringVar(&since, "since", "", "show logs after this time (1h, 30m, or RFC3339)")
	cmd.Flags().StringVar(&until, "until", "", "show logs before this time")
	cmd.Flags().StringVarP(&match, "match", "m", "", "regex pattern to match against the line or message")
	cmd.Flags().StringSliceVar(&fields, "field", nil, "filter by field, e.g. --field service=auth (repeatable)")
	cmd.Flags().IntVarP(&limit, "limit", "n", 0, "max number of matching lines to output (0 = no limit)")
	cmd.Flags().BoolVarP(&invert, "invert", "v", false, "invert the match (like grep -v)")

	return cmd
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
		if err := out.Write(entry); err != nil {
			return err
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

