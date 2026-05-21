package csvimport

import (
	"testing"
)

// FuzzParseLine exercises parseLine against arbitrary string input.
// Seeded with three valid CSV lines; the fuzzer mutates from there
// looking for inputs that panic.
//
// Run normally (just the seeds, ~milliseconds):
//
//	go test ./lessons/15-structure/solutions/internal/csvimport/
//
// Run as a fuzzer (until interrupted, explores the input space):
//
//	go test -fuzz=FuzzParseLine ./lessons/15-structure/solutions/internal/csvimport/
//
// parseLine returns an error for malformed input — that's fine. We
// assert it never PANICS. Any panic is a real bug.
func FuzzParseLine(f *testing.F) {
	f.Add("2026-05-21,4.50,coffee")
	f.Add("")
	f.Add("a,b,c")

	f.Fuzz(func(t *testing.T, line string) {
		// parseLine is package-private; this fuzz test lives in the
		// same package so it can call it directly.
		_, _ = parseLine(line) // any panic fails the test
	})
}
