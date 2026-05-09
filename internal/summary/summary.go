// Package summary aggregates a stream of [parser.Entry] values into a
// printable report. Aggregation is single-goroutine and streaming: callers
// drive it via [Summary.Observe] and read the result with [Summary.Render].
package summary

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/hermanu/logz/internal/parser"
)

// Summary accumulates statistics over a stream of entries. The zero value is
// not usable; construct instances with [New].
type Summary struct {
	topN int

	total      int
	skipped    int
	byLevel    map[parser.Level]int
	messages   map[string]int
	fieldHits  map[string]int
	timestamps []time.Time
	first      time.Time
	last       time.Time

	// timestampCap caps the number of timestamps retained for events/sec
	// computation. Without it, summarizing a multi-GB log file would allocate
	// proportionally; the cap trades a little accuracy on very large files for
	// bounded memory.
	timestampCap int
}

// New returns a Summary that retains up to topN entries in each top-N table.
// A non-positive topN falls back to the package default of 10.
func New(topN int) *Summary {
	if topN <= 0 {
		topN = 10
	}
	return &Summary{
		topN:         topN,
		byLevel:      make(map[parser.Level]int),
		messages:     make(map[string]int),
		fieldHits:    make(map[string]int),
		timestampCap: 1_000_000,
	}
}

// Observe records one parsed entry. Safe to call repeatedly; Summary is
// designed to be driven by a single goroutine.
func (s *Summary) Observe(e parser.Entry) {
	s.total++
	s.byLevel[e.Level]++

	if e.Level >= parser.LevelError && e.Message != "" {
		s.messages[e.Message]++
	}
	for k := range e.Fields {
		s.fieldHits[k]++
	}
	if !e.Timestamp.IsZero() {
		if s.first.IsZero() || e.Timestamp.Before(s.first) {
			s.first = e.Timestamp
		}
		if e.Timestamp.After(s.last) {
			s.last = e.Timestamp
		}
		if len(s.timestamps) < s.timestampCap {
			s.timestamps = append(s.timestamps, e.Timestamp)
		}
	}
}

// ObserveSkip records one line that the parser could not handle. These are
// surfaced in the rendered report under "Unparseable".
func (s *Summary) ObserveSkip() { s.skipped++ }

// Total returns the number of entries observed.
func (s *Summary) Total() int { return s.total }

// Skipped returns the number of unparseable lines observed.
func (s *Summary) Skipped() int { return s.skipped }

// LevelCount returns the number of entries observed at the given level.
func (s *Summary) LevelCount(l parser.Level) int { return s.byLevel[l] }

// TimeRange returns the earliest and latest timestamps observed. If no entry
// carried a timestamp both return values are the zero [time.Time].
func (s *Summary) TimeRange() (first, last time.Time) { return s.first, s.last }

// Render writes a human-readable report to w. topN limits the number of rows
// in each top-N table; a non-positive value falls back to the value passed
// to [New].
func (s *Summary) Render(w io.Writer, topN int) error {
	if topN <= 0 {
		topN = s.topN
	}
	ew := &errWriter{w: w}

	s.renderHeader(ew)
	s.renderLevels(ew)
	s.renderRate(ew)
	s.renderTopMessages(ew, topN)
	s.renderTopFields(ew, topN)

	return ew.err
}

func (s *Summary) renderHeader(w *errWriter) {
	w.printf("Total lines:   %d\n", s.total)
	if s.skipped > 0 {
		w.printf("Unparseable:   %d\n", s.skipped)
	}
	if !s.first.IsZero() {
		w.printf("Time range:    %s → %s\n",
			s.first.Format(time.RFC3339), s.last.Format(time.RFC3339))
	}
}

func (s *Summary) renderLevels(w *errWriter) {
	w.printf("\nBy level:\n")
	for _, l := range []parser.Level{
		parser.LevelTrace, parser.LevelDebug, parser.LevelInfo,
		parser.LevelWarn, parser.LevelError, parser.LevelFatal, parser.LevelUnknown,
	} {
		if n := s.byLevel[l]; n > 0 {
			w.printf("  %-7s %d\n", l, n)
		}
	}
}

func (s *Summary) renderRate(w *errWriter) {
	avg, peak := s.eventsPerSecond()
	if peak > 0 {
		w.printf("\nEvents/sec:    avg %.2f, peak %d\n", avg, peak)
	}
}

func (s *Summary) renderTopMessages(w *errWriter, topN int) {
	msgs := topFromMap(s.messages, topN)
	if len(msgs) == 0 {
		return
	}
	w.printf("\nTop %d error messages:\n", len(msgs))
	for _, m := range msgs {
		w.printf("  %5d  %s\n", m.count, truncate(m.key, 100))
	}
}

func (s *Summary) renderTopFields(w *errWriter, topN int) {
	if len(s.fieldHits) == 0 {
		return
	}
	fields := topFromMap(s.fieldHits, topN)
	w.printf("\nTop %d fields:\n", len(fields))
	for _, f := range fields {
		w.printf("  %5d  %s\n", f.count, f.key)
	}
}

// errWriter wraps an io.Writer and stickily records the first error so
// callers can check once at the end instead of after every write. Adopted
// from the pattern in [Errors are values] (https://go.dev/blog/errors-are-values).
type errWriter struct {
	w   io.Writer
	err error
}

func (e *errWriter) printf(format string, args ...any) {
	if e.err != nil {
		return
	}
	_, e.err = fmt.Fprintf(e.w, format, args...)
}

// eventsPerSecond returns the mean and peak rate observed across the retained
// timestamps. peak is 0 when no timestamps are available.
func (s *Summary) eventsPerSecond() (avg float64, peak int) {
	if len(s.timestamps) == 0 {
		return 0, 0
	}
	buckets := make(map[int64]int, len(s.timestamps))
	for _, t := range s.timestamps {
		buckets[t.Unix()]++
	}
	for _, n := range buckets {
		if n > peak {
			peak = n
		}
	}
	span := s.last.Sub(s.first).Seconds()
	if span <= 0 {
		span = 1
	}
	return float64(len(s.timestamps)) / span, peak
}

// truncate shortens s to at most n runes plus an ellipsis.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// kv pairs a string key with a count. Used internally for top-N tables.
type kv struct {
	key   string
	count int
}

// topFromMap returns the top-n highest-count entries from m, sorted by count
// descending and key ascending for ties.
func topFromMap(m map[string]int, n int) []kv {
	if len(m) == 0 {
		return nil
	}
	out := make([]kv, 0, len(m))
	for k, v := range m {
		out = append(out, kv{key: k, count: v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].count != out[j].count {
			return out[i].count > out[j].count
		}
		return out[i].key < out[j].key
	})
	if len(out) > n {
		out = out[:n]
	}
	return out
}
