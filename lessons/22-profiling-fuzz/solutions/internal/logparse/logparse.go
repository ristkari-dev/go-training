// Package logparse is the lesson 22 reference implementation.
//
// parseLineRegex (carried from L14/L20) is kept as the fuzz/benchmark
// oracle; Parse now uses the allocation-free parseLineFast.
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
		e, err := parseLineFast(text)
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

// parseLineRegex is the carried-forward reference parser, retained as
// the fuzz/benchmark oracle.
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

// parseLineFast is the allocation-free hand-written parser. It matches
// parseLineRegex whenever the regex accepts; it never panics.
func parseLineFast(s string) (LogEntry, error) {
	const tsLen = 19 // 2006-01-02T15:04:05
	if len(s) < tsLen {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}
	tsStr := s[:tsLen]
	rest := s[tsLen:]

	// \s+ between timestamp and level.
	rest, ok := cutLeadingSpace(rest)
	if !ok {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}

	// level runs up to the next \s char.
	sp := -1
	for i := 0; i < len(rest); i++ {
		if isLogSpace(rest[i]) {
			sp = i
			break
		}
	}
	if sp < 0 {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}
	level := rest[:sp]
	if level != "INFO" && level != "WARN" && level != "ERROR" {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}

	// \s+ between level and message; message non-empty, no '\n'.
	msg, ok := cutLeadingSpace(rest[sp:])
	if !ok || msg == "" {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}
	if strings.IndexByte(msg, '\n') >= 0 {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}

	t, err := time.Parse(timeLayout, tsStr)
	if err != nil {
		return LogEntry{}, fmt.Errorf("invalid timestamp %q: %w", tsStr, err)
	}
	return LogEntry{Time: t, Level: level, Message: msg}, nil
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
