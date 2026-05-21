// Package logparse parses simple structured log lines.
//
// Expected log line format:
//
//	2026-05-21T14:30:00 INFO connection accepted from 192.168.1.5
//	|------ timestamp ------|--lvl|---------- message ---------|
//
// Three space-separated parts:
//
//  1. Timestamp in RFC3339-ish "2006-01-02T15:04:05" format (no timezone)
//  2. Level — one of INFO, WARN, ERROR (uppercase)
//  3. Free-form message text
//
// The package exports:
//
//   - LogEntry struct with parsed Time / Level / Message
//   - Parse(r io.Reader) ([]LogEntry, error) — bufio.Scanner-based reader
//   - CountByLevel(entries []LogEntry) map[string]int — tally helper
//
// Parse returns partial entries plus a wrapped error if any line is
// malformed, mirroring lesson 13's csvimport.Parse shape.
package logparse

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

// LogEntry is one parsed log line.
type LogEntry struct {
	Time    time.Time
	Level   string
	Message string
}

// logLineRE matches:
//   - capture 1: timestamp in "2006-01-02T15:04:05" shape
//   - capture 2: level INFO|WARN|ERROR
//   - capture 3: free-form message (everything after the level)
//
// Compiled ONCE at package init via MustCompile. If the regex string
// were invalid, the program would panic immediately instead of failing
// at first call — that's the idiomatic Go choice for package-level
// regexes you know are correct.
var logLineRE = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})\s+(INFO|WARN|ERROR)\s+(.+)$`)

const timeLayout = "2006-01-02T15:04:05"

// Parse reads log lines from r via bufio.Scanner. Returns the parsed
// entries plus a wrapped error if any line is malformed.
//
// Blank lines are skipped. On a malformed line, Parse returns the
// partial list parsed so far plus the error.
//
// Hint:
//  1. s := bufio.NewScanner(r)
//  2. out := []LogEntry{}
//  3. for line := 1; s.Scan(); line++ {
//     text := strings.TrimSpace(s.Text())
//     if text == "" { continue }
//     e, err := parseLine(text)
//     if err != nil → return out, fmt.Errorf("logparse: line %d: %w", line, err)
//     out = append(out, e)
//     }
//  4. if err := s.Err(); err != nil → return out, fmt.Errorf("logparse: %w", err)
//  5. return out, nil
//
// parseLine is a helper: apply the regex, expect 4 elements
// ([full match, ts, level, msg]), parse the timestamp via time.Parse.
func Parse(r io.Reader) ([]LogEntry, error) {
	_ = bufio.NewScanner
	_ = strings.TrimSpace
	_ = fmt.Errorf
	_ = time.Parse
	_ = logLineRE.FindStringSubmatch
	_ = timeLayout
	panic("TODO: scan loop with trim/skip-blank + parseLine + line-numbered wrapping")
}

// CountByLevel tallies entries per level. Returns a map from level → count.
//
// Hint: out := map[string]int{}; for _, e := range entries { out[e.Level]++ }; return out.
func CountByLevel(entries []LogEntry) map[string]int {
	panic("TODO: range over entries; map[level]++")
}
