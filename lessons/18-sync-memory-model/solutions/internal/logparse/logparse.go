// Package logparse is the lesson 14 main reference implementation.
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
		e, err := parseLine(text)
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

func parseLine(s string) (LogEntry, error) {
	m := logLineRE.FindStringSubmatch(s)
	if m == nil {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}
	// m[0] is the full match; m[1..3] are the captures.
	t, err := time.Parse(timeLayout, m[1])
	if err != nil {
		return LogEntry{}, fmt.Errorf("invalid timestamp %q: %w", m[1], err)
	}
	return LogEntry{Time: t, Level: m[2], Message: m[3]}, nil
}

// CountByLevel tallies entries per level.
func CountByLevel(entries []LogEntry) map[string]int {
	out := map[string]int{}
	for _, e := range entries {
		out[e.Level]++
	}
	return out
}
