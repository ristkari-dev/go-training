// Package logparse is the lesson 22 reference implementation.
//
// L22 optimization arc: parseLineRegex (carried from L14/L20) is the
// reference. You implement parseLineFast — a hand-written parser that
// avoids the per-line regexp allocation. Profile first, then optimize.
package logparse

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

type LogEntry struct {
	Time    time.Time
	Level   string
	Message string
}

var logLineRE = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})\s+(INFO|WARN|ERROR)\s+(.+)$`)

const timeLayout = "2006-01-02T15:04:05"

// Parse reads log lines from r and returns parsed entries.
func Parse(r io.Reader) ([]LogEntry, error) {
	s := bufio.NewScanner(r)
	out := []LogEntry{}
	for line := 1; s.Scan(); line++ {
		text := strings.TrimSpace(s.Text())
		if text == "" {
			continue
		}
		// TODO (L22): once parseLineFast is implemented and benchmarked,
		// switch this call to parseLineFast.
		e, err := parseLineRegex(text)
		if err != nil {
			return out, fmt.Errorf("logparse: line %d: %w", line, err)
		}
		out = append(out, e)
	}
	if err := s.Err(); err != nil {
		return out, fmt.Errorf("logparse: %w", err)
	}
	return out, nil
}

// parseLineRegex is the carried-forward reference parser. Kept as the
// fuzz/benchmark oracle even after Parse switches to parseLineFast.
func parseLineRegex(s string) (LogEntry, error) {
	m := logLineRE.FindStringSubmatch(s)
	if m == nil {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}
	t, err := time.Parse(timeLayout, m[1])
	if err != nil {
		return LogEntry{}, fmt.Errorf("invalid timestamp %q: %w", m[1], err)
	}
	return LogEntry{Time: t, Level: m[2], Message: m[3]}, nil
}

// parseLineFast is the hand-written parser you implement. No regexp,
// no per-call slice allocation.
//
// The format is fixed: "2006-01-02T15:04:05 LEVEL message".
//   - timestamp is exactly 19 chars
//   - one or more whitespace (\s = [\t\n\f\r ])
//   - level is INFO|WARN|ERROR
//   - one or more whitespace
//   - non-empty message containing no '\n' (regex '.' excludes newline)
//
// CRITICAL: guard every index. strings.IndexByte returns -1 when the
// byte is absent; slicing s[:-1] PANICS. Fuzzing will find this if you
// forget (try seed "0"). Return an error, never panic.
//
// Hint:
//
//	const tsLen = 19
//	if len(s) < tsLen { return LogEntry{}, errMalformed }
//	tsStr, rest := s[:tsLen], s[tsLen:]
//	rest, ok := cutLeadingSpace(rest); if !ok { return ... }
//	find next \s in rest -> sp; if sp < 0 { return ... }
//	level := rest[:sp]; validate INFO|WARN|ERROR
//	msg, ok := cutLeadingSpace(rest[sp:]); if !ok || msg == "" { return ... }
//	if strings.IndexByte(msg, '\n') >= 0 { return ... }   // '.' excludes '\n'
//	t, err := time.Parse(timeLayout, tsStr); ...
func parseLineFast(s string) (LogEntry, error) {
	_ = cutLeadingSpace
	_ = isLogSpace
	panic("TODO: hand-written parser; guard every index; match parseLineRegex when it accepts")
}

// isLogSpace matches RE2's \s class: [\t\n\f\r ].
func isLogSpace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\f', '\r':
		return true
	}
	return false
}

// cutLeadingSpace strips a run of >=1 \s chars from the front. ok is
// false when there was no leading \s char at all (mirrors regex \s+).
func cutLeadingSpace(s string) (rest string, ok bool) {
	i := 0
	for i < len(s) && isLogSpace(s[i]) {
		i++
	}
	return s[i:], i > 0
}

// CountByLevel tallies entries per level.
func CountByLevel(entries []LogEntry) map[string]int {
	out := map[string]int{}
	for _, e := range entries {
		out[e.Level]++
	}
	return out
}
