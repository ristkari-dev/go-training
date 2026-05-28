package logparse

import (
	"strings"
	"testing"
)

// TestParseLineFastMatchesRegex checks the fast parser agrees with the
// regex reference across well-formed and malformed inputs. Inputs are
// trimmed first, exactly as Parse feeds them to the line parser.
func TestParseLineFastMatchesRegex(t *testing.T) {
	cases := []string{
		"2026-01-02T15:04:05 INFO server started",
		"2026-01-02T15:04:05 WARN slow query",
		"2026-01-02T15:04:05 ERROR boom",
		"2026-01-02T15:04:05   INFO   extra spaces",
		"2026-01-02T15:04:05\tINFO\ttabs",
		"2026-01-02T15:04:05 INFO trailing words here",
		"2026-13-02T15:04:05 INFO bad month",
		"not a log line",
		"",
		"INFO",
		"2026-01-02T15:04:05 DEBUG unknown level",
		"2026-01-02T15:04:05 INFO ",
	}
	for _, raw := range cases {
		s := strings.TrimSpace(raw)
		want, errRe := parseLineRegex(s)
		got, errFast := parseLineFast(s)
		if errRe == nil {
			if errFast != nil {
				t.Errorf("regex accepted %q but fast rejected: %v", s, errFast)
				continue
			}
			if got != want {
				t.Errorf("mismatch for %q: fast=%+v regex=%+v", s, got, want)
			}
		}
	}
}

// FuzzParseLineFast asserts two contracts:
//
//	(1) parseLineFast never panics on ANY input — even raw fuzz bytes.
//	(2) On the inputs Parse actually hands it (trimmed), whenever the
//	    regex reference accepts, parseLineFast accepts identically.
//
// Why trim for contract (2): Parse calls strings.TrimSpace before the
// line parser, so the parser's real contract is "trimmed, non-empty
// line". The regex's greedy \s+(.+)$ backtracks and accepts a
// whitespace-only message on RAW input (e.g. "TS INFO  " → message
// " "); the hand-written parser strips all leading space and rejects
// it. That divergence only exists on untrimmed input Parse never
// produces, so contract (2) compares on the trimmed form. Contract (1)
// still guards every raw byte — that is what catches the classic
// no-space panic (try the "0" seed against a naive parser).
//
// The seeds include the no-space inputs ("0", "", "INFO") that crash a
// naive parser — permanent regression cases. Run the random fuzzer with:
//
//	go test -run='^$' -fuzz=FuzzParseLineFast -fuzztime=30s
func FuzzParseLineFast(f *testing.F) {
	seeds := []string{
		"2026-01-02T15:04:05 INFO ok",
		"2026-01-02T15:04:05 WARN w",
		"2026-01-02T15:04:05 ERROR e",
		"2026-01-02T15:04:05   INFO   spaces",
		"",
		"0",
		"INFO",
		"x",
		"2026-01-02T15:04:05",
		"2026-01-02T15:04:05 ",
		"2026-01-02T15:04:05 INFO ",
		"2026-01-02T15:04:05 INFO  ",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		// Contract (1): never panic on raw input.
		_, _ = parseLineFast(s)

		// Contract (2): agree with the oracle on trimmed input.
		trimmed := strings.TrimSpace(s)
		got, errFast := parseLineFast(trimmed)
		want, errRe := parseLineRegex(trimmed)
		if errRe == nil {
			if errFast != nil {
				t.Errorf("regex accepted %q but fast rejected: %v", trimmed, errFast)
			} else if got != want {
				t.Errorf("mismatch for %q: fast=%+v regex=%+v", trimmed, got, want)
			}
		}
	})
}
