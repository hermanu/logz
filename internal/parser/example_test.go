package parser_test

import (
	"errors"
	"fmt"

	"github.com/hermanu/logz/internal/parser"
)

func ExampleParseLevel() {
	for _, s := range []string{"INFO", "Warning", "err", "panic", "?"} {
		lvl, ok := parser.ParseLevel(s)
		fmt.Printf("%-7s -> %s (ok=%t)\n", s, lvl, ok)
	}
	// Output:
	// INFO    -> info (ok=true)
	// Warning -> warn (ok=true)
	// err     -> error (ok=true)
	// panic   -> fatal (ok=true)
	// ?       -> unknown (ok=false)
}

func ExampleJSONParser_Parse() {
	p := parser.NewJSONParser(parser.DefaultOptions())
	e, err := p.Parse(`{"level":"error","msg":"timeout","service":"api"}`)
	if err != nil {
		fmt.Println("parse failed:", err)
		return
	}
	fmt.Printf("level=%s msg=%q service=%s\n", e.Level, e.Message, e.Fields["service"])
	// Output: level=error msg="timeout" service=api
}

func ExampleDetect() {
	opts := parser.DefaultOptions()
	for _, sample := range []string{
		`{"level":"info"}`,
		`2024-01-01 12:00:00 INFO hello`,
	} {
		fmt.Println(parser.Detect([]byte(sample), opts).Name())
	}
	// Output:
	// json
	// text
}

func ExampleErrSkip() {
	p := parser.NewJSONParser(parser.DefaultOptions())
	_, err := p.Parse("   ") // blank line
	fmt.Println(errors.Is(err, parser.ErrSkip))
	// Output: true
}
