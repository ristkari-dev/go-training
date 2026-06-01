// Package logparse parses application log lines into structured entries.
// Carried forward from L14; L22's optimized hand-written parser is now
// simply the parser (the regex reference + fuzz/bench scaffolding from
// L22 are dropped). ParseLine is exported so callers can parse and
// count lines individually.
package logparse

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"
)

type LogEntry struct {
	Time    time.Time
	Level   string
	Message string
}

const timeLayout = "2006-01-02T15:04:05"

// Parse reads log lines from r and returns parsed entries. Stops at the
// first malformed line, returning the entries parsed so far plus an error.
func Parse(r io.Reader) ([]LogEntry, error) {
	s := bufio.NewScanner(r)
	out := []LogEntry{}
	for line := 1; s.Scan(); line++ {
		text := strings.TrimSpace(s.Text())
		if text == "" {
			continue
		}
		e, err := ParseLine(text)
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

// ParseLine parses a single log line. The allocation-free hand-written
// parser from L22, now exported. Never panics; returns an error for
// malformed input.
func ParseLine(s string) (LogEntry, error) {
	const tsLen = 19 // 2006-01-02T15:04:05
	if len(s) < tsLen {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}
	tsStr := s[:tsLen]
	rest := s[tsLen:]

	rest, ok := cutLeadingSpace(rest)
	if !ok {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}

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

// cutLeadingSpace strips a run of >=1 \s chars from the front.
func cutLeadingSpace(s string) (string, bool) {
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
