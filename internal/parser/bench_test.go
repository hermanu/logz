package parser_test

import (
	"testing"

	"github.com/hermanu/logz/internal/parser"
)

const (
	benchJSONLine = `{"level":"error","msg":"timeout connecting to database","ts":"2024-01-01T12:00:00Z","service":"api","retry":3,"request_id":"abc-123-def"}`
	benchTextLine = `2024-01-01T12:00:00Z ERROR timeout connecting to database`
)

func BenchmarkJSONParser_Parse(b *testing.B) {
	p := parser.NewJSONParser(parser.DefaultOptions())
	b.ReportAllocs()
	b.SetBytes(int64(len(benchJSONLine)))
	for i := 0; i < b.N; i++ {
		if _, err := p.Parse(benchJSONLine); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTextParser_Parse(b *testing.B) {
	p := parser.NewTextParser(parser.DefaultOptions())
	b.ReportAllocs()
	b.SetBytes(int64(len(benchTextLine)))
	for i := 0; i < b.N; i++ {
		if _, err := p.Parse(benchTextLine); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDetect(b *testing.B) {
	sample := []byte(benchJSONLine + "\n")
	opts := parser.DefaultOptions()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = parser.Detect(sample, opts)
	}
}
