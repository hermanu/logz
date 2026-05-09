package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/hermanu/logz/internal/filter"
	"github.com/hermanu/logz/internal/parser"
	"github.com/hermanu/logz/internal/summary"
)

func newSummaryCmd() *cobra.Command {
	var (
		since string
		until string
		topN  int
	)

	cmd := &cobra.Command{
		Use:   "summary [flags] [file...]",
		Short: "Summarize a log file",
		Long:  "Print a summary report: counts by level, time range, top errors, top fields, and events/sec.",
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

			f := &filter.Filter{}
			now := time.Now()
			if f.Since, err = filter.ParseSince(since, now); err != nil {
				return err
			}
			if f.Until, err = filter.ParseSince(until, now); err != nil {
				return err
			}

			if topN <= 0 {
				topN = cfg.Top
			}
			s := summary.New(topN)

			sources, err := resolveSources(args, cmd.InOrStdin())
			if err != nil {
				return err
			}

			for _, src := range sources {
				if err := summarizeSource(src, g.format, opts, f, s); err != nil {
					_ = src.rc.Close()
					return err
				}
				_ = src.rc.Close()
			}

			return s.Render(cmd.OutOrStdout(), topN)
		},
	}

	cmd.Flags().StringVar(&since, "since", "", "only include logs after this time")
	cmd.Flags().StringVar(&until, "until", "", "only include logs before this time")
	cmd.Flags().IntVar(&topN, "top", 0, "number of top entries to show (default from config)")

	return cmd
}

func summarizeSource(src namedSource, formatOverride string, opts parser.Options, f *filter.Filter, s *summary.Summary) error {
	const sniffSize = 16 * 1024
	sample, body, err := parser.SniffSample(src.rc, sniffSize)
	if err != nil {
		return fmt.Errorf("sniff %s: %w", src.name, err)
	}
	p, err := pickParser(formatOverride, sample, opts)
	if err != nil {
		return err
	}
	return streamLines(body, func(line string) error {
		entry, perr := p.Parse(line)
		if perr != nil {
			if errors.Is(perr, parser.ErrSkip) {
				return nil
			}
			s.ObserveSkip()
			return nil
		}
		if !f.Allow(entry) {
			return nil
		}
		s.Observe(entry)
		return nil
	})
}
