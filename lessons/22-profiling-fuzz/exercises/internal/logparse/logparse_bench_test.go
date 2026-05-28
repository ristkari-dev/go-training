package logparse

import "testing"

// sink prevents the compiler from eliminating the parse result as dead
// code (a classic benchmarking mistake).
var sink LogEntry

// BenchmarkParseLineRegex / BenchmarkParseLineFast compare the two
// parsers. Run: go test -bench=ParseLine -benchmem
//
// TODO: build a fixed line, loop b.N times calling each parser, assign
// to sink, ReportAllocs. See the solution for the shape.
func BenchmarkParseLineRegex(b *testing.B) {
	_ = sink
	b.Skip("TODO: implement BenchmarkParseLineRegex")
}

func BenchmarkParseLineFast(b *testing.B) {
	b.Skip("TODO: implement BenchmarkParseLineFast")
}
