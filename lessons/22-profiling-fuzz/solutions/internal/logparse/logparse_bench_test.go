package logparse

import "testing"

// sink prevents dead-code elimination of the parse result.
var sink LogEntry

const benchLine = "2026-01-02T15:04:05 INFO message number 42 here"

func BenchmarkParseLineRegex(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e, _ := parseLineRegex(benchLine)
		sink = e
	}
}

func BenchmarkParseLineFast(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e, _ := parseLineFast(benchLine)
		sink = e
	}
}
