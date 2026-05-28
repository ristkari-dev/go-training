package logparse

import "testing"

// FuzzParseLineFast is a SKELETON. The shipped target asserts:
//
//	(1) parseLineFast never panics, and
//	(2) whenever parseLineRegex ACCEPTS, parseLineFast agrees identically.
//
// Run with: go test -fuzz=FuzzParseLineFast
//
// TODO: seed with well-formed lines AND the no-space crashers ("0", "",
// "INFO"); in f.Fuzz compare parseLineFast vs parseLineRegex one-way.
func FuzzParseLineFast(f *testing.F) {
	f.Add("2026-01-02T15:04:05 INFO ok")
	f.Fuzz(func(t *testing.T, s string) {
		// TODO: remove this Skip once parseLineFast is implemented. Until
		// then it panics, so the seed corpus must not invoke it.
		t.Skip("TODO: implement FuzzParseLineFast")
		_, _ = parseLineFast(s)
		_ = parseLineRegex
	})
}
