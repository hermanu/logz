package filter_test

import (
	"fmt"
	"regexp"
	"time"

	"github.com/hermanu/logz/internal/filter"
	"github.com/hermanu/logz/internal/parser"
)

func ExampleFilter_Allow() {
	f := &filter.Filter{
		MinLevel: parser.LevelWarn,
		Match:    regexp.MustCompile("timeout"),
	}

	cases := []parser.Entry{
		{Level: parser.LevelInfo, Message: "request timeout"},
		{Level: parser.LevelError, Message: "request timeout"},
		{Level: parser.LevelError, Message: "request succeeded"},
	}
	for _, e := range cases {
		fmt.Printf("%-5s %-20q -> %t\n", e.Level, e.Message, f.Allow(e))
	}
	// Output:
	// info  "request timeout"    -> false
	// error "request timeout"    -> true
	// error "request succeeded"  -> false
}

func ExampleParseSince() {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	rel, _ := filter.ParseSince("1h", now)
	abs, _ := filter.ParseSince("2024-01-01T11:30:00Z", now)

	fmt.Println(rel.Format(time.RFC3339))
	fmt.Println(abs.Format(time.RFC3339))
	// Output:
	// 2024-01-01T11:00:00Z
	// 2024-01-01T11:30:00Z
}
