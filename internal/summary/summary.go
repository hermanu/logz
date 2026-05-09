// Package summary aggregates [parser.Entry] streams into a printable report.
package summary

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/hermanu/logz/internal/parser"
)

// Summary accumulates statistics over a stream of entries.
type Summary struct {
	Total         int
	Skipped       int
	ByLevel       map[parser.Level]int
	TopMessages   *topK
	TopFields     map[string]int
	First, Last   time.Time
	timestamps    []time.Time // for peak EPS computation
	maxTimestamps int
}

// New returns a Summary configured with the given top-N capacity.
func New(topN int) *Summary {
	if topN <= 0 {
		topN = 10
	}
	return &Summary{
		ByLevel:       map[parser.Level]int{},
		TopMessages:   newTopK(topN),
		TopFields:     map[string]int{},
		maxTimestamps: 1_000_000, // cap memory at ~16MB of timestamps
	}
}

// Observe records one parsed entry.
func (s *Summary) Observe(e parser.Entry) {
	s.Total++
	s.ByLevel[e.Level]++

	if e.Level >= parser.LevelError && e.Message != "" {
		s.TopMessages.add(e.Message)
	}
	for k := range e.Fields {
		s.TopFields[k]++
	}
	if !e.Timestamp.IsZero() {
		if s.First.IsZero() || e.Timestamp.Before(s.First) {
			s.First = e.Timestamp
		}
		if e.Timestamp.After(s.Last) {
			s.Last = e.Timestamp
		}
		if len(s.timestamps) < s.maxTimestamps {
			s.timestamps = append(s.timestamps, e.Timestamp)
		}
	}
}

// ObserveSkip records an unparseable line.
func (s *Summary) ObserveSkip() { s.Skipped++ }

// Render writes a human-readable summary to w. topN limits the number of rows
// in each top-N table.
func (s *Summary) Render(w io.Writer, topN int) error {
	if topN <= 0 {
		topN = 10
	}

	if _, err := fmt.Fprintf(w, "Total lines:   %d\n", s.Total); err != nil {
		return err
	}
	if s.Skipped > 0 {
		fmt.Fprintf(w, "Unparseable:   %d\n", s.Skipped)
	}
	if !s.First.IsZero() {
		fmt.Fprintf(w, "Time range:    %s → %s\n", s.First.Format(time.RFC3339), s.Last.Format(time.RFC3339))
	}

	fmt.Fprintln(w, "\nBy level:")
	levels := []parser.Level{
		parser.LevelTrace, parser.LevelDebug, parser.LevelInfo,
		parser.LevelWarn, parser.LevelError, parser.LevelFatal, parser.LevelUnknown,
	}
	for _, l := range levels {
		if n := s.ByLevel[l]; n > 0 {
			fmt.Fprintf(w, "  %-7s %d\n", l, n)
		}
	}

	if avg, peak := s.eventsPerSecond(); peak > 0 {
		fmt.Fprintf(w, "\nEvents/sec:    avg %.2f, peak %d\n", avg, peak)
	}

	if msgs := s.TopMessages.top(topN); len(msgs) > 0 {
		fmt.Fprintf(w, "\nTop %d error messages:\n", len(msgs))
		for _, m := range msgs {
			fmt.Fprintf(w, "  %5d  %s\n", m.count, truncate(m.key, 100))
		}
	}

	if len(s.TopFields) > 0 {
		fields := topFromMap(s.TopFields, topN)
		fmt.Fprintf(w, "\nTop %d fields:\n", len(fields))
		for _, f := range fields {
			fmt.Fprintf(w, "  %5d  %s\n", f.count, f.key)
		}
	}

	return nil
}

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
	span := s.Last.Sub(s.First).Seconds()
	if span <= 0 {
		span = 1
	}
	return float64(len(s.timestamps)) / span, peak
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

type kv struct {
	key   string
	count int
}

func topFromMap(m map[string]int, n int) []kv {
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

// topK keeps a tally and returns the most-frequent keys. Implementation is a
// plain map; sufficient for typical log volumes since we only care about
// distinct messages, which are bounded in practice.
type topK struct {
	capacity int
	counts   map[string]int
}

func newTopK(n int) *topK {
	return &topK{capacity: n, counts: map[string]int{}}
}

func (t *topK) add(key string) { t.counts[key]++ }

func (t *topK) top(n int) []kv { return topFromMap(t.counts, n) }
