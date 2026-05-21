# Plan Q — Lesson 14 (Time, strings, bytes, regex) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author lesson 14 of Phase 2 — students learn the `time`, `strings`, `bytes`, and `regexp` standard library packages as a coherent toolkit, plus the Go-culture attitude that "regex is a tool of last resort." The warmup implements a round-trip date formatter/parser using the famous reference-time idiom. The main exercise builds a tiny log-line parser: regex extracts timestamp + level + message; a `CountByLevel` helper tallies entries per level. **Standalone** — no tracker integration (per Phase 2 spec; tracker resumes in L15 capstone).

**Architecture:** Same per-lesson pattern as Plans D-P. Six tasks. Skeleton tests in warmup + main subpackages (Phase 2 default). Heavy-explanatory slide deck with **five** concept blocks (time formatting → strings → bytes → regexp basics → when to reach for what).

**Tech Stack:** Go 1.23 stdlib only (`bufio`, `bytes`, `errors`, `fmt`, `io`, `regexp`, `strings`, `time`, `testing`). Reveal.js 5.1.0.

---

## Scope

After Plan Q: lesson 14 is complete; `make test` green; both subpackages exercised through their public APIs in solution tests; slides build and `14-time-strings-regex` lands in the index.

### Design decisions (2 user-approved + 7 plan-recommended)

**User-approved via brainstorming:**

1. **`logparse` has TWO functions**: `Parse(r io.Reader) ([]LogEntry, error)` returning structured entries, plus `CountByLevel(entries []LogEntry) map[string]int` as a pure tally. The compositional split lets tests assert on EXACT parsed timestamps/levels/messages (more rigorous than aggregate counts), gives the regex's capture groups a properly-typed home (`Time time.Time`, not string), and mirrors L13's `csvimport.Parse` shape for visual continuity.

2. **Five concepts in the slide deck**: time formatting → strings → **bytes (its own slot)** → regexp basics → "when to reach for what." The bytes split (vs bundling with strings) earns its keep with a real Common-mistake slot: `==` on `[]byte` is a compile error (slices aren't comparable), and `bytes.Equal`/`bytes.Buffer` mutation semantics are pedagogically distinct from strings' immutability.

**Plan-recommended:**

3. **Standalone — no cmd binaries.** Per Phase 2 spec. Library-level lesson; the deliverables are the `dates` and `logparse` packages + slides + README. No `cmd/logparser` or `cmd/dateformat`. Adding a binary would dilute the focus.

4. **`logparse.Parse` error semantics match L13's `csvimport.Parse`**: returns partial entries on first malformed line + a wrapped error `fmt.Errorf("logparse: line %d: malformed log line %q", n, line)`. Visual parallel — students recognize the shape from L13. Pedagogically aligned: error handling and partial-result patterns carry forward, not get reinvented.

5. **Log format**: `<RFC3339-ish timestamp> <LEVEL> <message>`. Timestamp uses `2006-01-02T15:04:05` (no timezone offset — keeps the format flag literal and recognizable). Levels are uppercase `INFO`, `WARN`, `ERROR` only. Free-form message after.

6. **Single package-level regex compiled with `MustCompile`**. `MustCompile` panics at package init if the regex is invalid — students see why "compile-once, use-many" is idiomatic. Per-call `regexp.Compile` would be a Common-mistake example in slides.

7. **`dates.ParseDate` wraps errors**: `fmt.Errorf("dates: invalid date %q: %w", s, err)`. Same pattern as L11's `parseage`. Tests assert the wrapped error preserves the underlying `time.Parse` failure.

8. **No new public APIs in `expense`/`store`/`csvimport`**. L14 is standalone; previous tracker code stays untouched.

9. **bytes is taught but not exercised in code.** The slide deck has a worked example showing `bytes.Equal` and `bytes.Contains`, but the test suite doesn't require students to write any `[]byte` parsing. Reason: the log-line parsing naturally lives in string-space (regexp on `[]byte` exists but is uncommon in real Go). Keeps the exercise focused on regex over strings.

---

## Plans F-P lessons-learned applied here

1. Subpackages have their own namespace — no `Warmup*` prefix.
2. Lint exclusion (`.golangci.yml` line 24, `lessons/.*/exercises/`) covers nested subpackages — no config changes.
3. Common-mistake content in README (5 per lesson, one per concept).
4. Slides + README written inline by controller.
5. `gofmt -w .` and `go vet ./...` mentions in README.
6. Skeleton tests pass vacuously.
7. Scaffold + restructure (delete 8 flat scaffolder files; create subpackage tree).
8. gofmt 1.19+ normalises godoc list indentation — if hint blocks complain, run `gofmt -w`.

---

## File Structure

After Plan Q (12 files total — same shape as L12; standalone footprint):

```
lessons/14-time-strings-regex/
├── README.md                                       (Task 5)
├── slides/
│   ├── index.html, slides.md, assets/.gitkeep      (Task 1 + Task 4)
├── exercises/
│   ├── warmup/dates/
│   │   ├── dates.go                                (Task 2)
│   │   └── dates_test.go                           (Task 2 — SKELETON)
│   └── logparse/
│       ├── logparse.go                             (Task 3)
│       └── logparse_test.go                        (Task 3 — SKELETON)
└── solutions/
    └── (mirrored structure with full implementations)
```

---

## Conventions

- **Branch:** `feature/plan-q-lesson-14-time-strings-regex`
- **Commit messages:** Conventional Commits

---

## Task 1: Scaffold + restructure

Same dance as Plans J/K/L/M/N/O/P.

- [ ] **Step 1:** `make new-lesson NAME=14-time-strings-regex`
- [ ] **Step 2:** Delete 8 unwanted flat scaffolder files:

```bash
rm lessons/14-time-strings-regex/exercises/warmup.go
rm lessons/14-time-strings-regex/exercises/warmup_test.go
rm lessons/14-time-strings-regex/exercises/main.go
rm lessons/14-time-strings-regex/exercises/main_test.go
rm lessons/14-time-strings-regex/solutions/warmup.go
rm lessons/14-time-strings-regex/solutions/warmup_test.go
rm lessons/14-time-strings-regex/solutions/main.go
rm lessons/14-time-strings-regex/solutions/main_test.go
```

- [ ] **Step 3:** Verify 4-file scaffolded tree.

- [ ] **Step 4:** Commit:

```bash
git add lessons/14-time-strings-regex/
git commit -m "feat(lessons): scaffold lesson 14-time-strings-regex with empty subpackage layout"
```

---

## Task 2: Author the warm-up — `dates` subpackage

Two functions: `FormatDate(t time.Time) string` and `ParseDate(s string) (time.Time, error)`. Together they're a round-trip — `ParseDate(FormatDate(t)) == t.Truncate(time.Second)`. Demonstrates Go's quirky reference-time format and the wrap-with-input error idiom.

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/14-time-strings-regex/{exercises,solutions}/warmup/dates`

- [ ] **Step 2:** Create `lessons/14-time-strings-regex/exercises/warmup/dates/dates.go`:

```go
// Package dates is the lesson 14 warm-up: format and parse dates using
// the famous Go reference-time idiom.
//
// Go's date formatting uses the literal reference time
//
//	Mon Jan 2 15:04:05 MST 2006
//
// (which numerically is 01/02 03:04:05PM '06 MST). Instead of cryptic
// %Y-%m-%d directives, you write what the OUTPUT should look like in
// terms of that reference moment. To format "year-month-day" you write
// "2006-01-02"; to format "hour:minute:second" you write "15:04:05".
//
// This package teaches the round trip — format a time, parse it back,
// get the same time (truncated to seconds).
package dates

import (
	"strconv"
	"time"
)

// FormatDate returns t formatted as "YYYY-MM-DD HH:MM:SS" (e.g.,
// "2026-05-21 14:30:00"). Uses the reference-time idiom: the layout
// string "2006-01-02 15:04:05" describes the OUTPUT shape using the
// reference moment's components.
//
// Hint: return t.Format("2006-01-02 15:04:05")
func FormatDate(t time.Time) string {
	_ = time.Now
	_ = strconv.Itoa
	panic("TODO: t.Format with the reference time layout")
}

// ParseDate parses s as a "YYYY-MM-DD HH:MM:SS" date.
//
// On success returns (parsed time, nil). On failure returns
// (zero time, wrapped error mentioning the offending input).
//
// Hint:
//   1. t, err := time.Parse("2006-01-02 15:04:05", s)
//   2. if err != nil → return time.Time{}, fmt.Errorf("dates: invalid date %q: %w", s, err)
//   3. return t, nil
//
// time.Parse uses the SAME reference-time layout string as Format. The
// trick: same string for both directions.
func ParseDate(s string) (time.Time, error) {
	panic("TODO: time.Parse with the reference time layout; wrap errors with %w")
}
```

- [ ] **Step 3:** Create `lessons/14-time-strings-regex/exercises/warmup/dates/dates_test.go` (SKELETON):

```go
package dates

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// TestFormatDate is a SKELETON. Build a few time.Time values via
// time.Date(year, month, day, hour, min, sec, nsec, time.UTC) and
// assert FormatDate produces the expected "YYYY-MM-DD HH:MM:SS" string.
func TestFormatDate(t *testing.T) {
	cases := []struct {
		name string
		in   time.Time
		want string
	}{
		// TODO: at least 3 cases. E.g.:
		// {"epoch", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), "1970-01-01 00:00:00"},
		// {"sample", time.Date(2026, 5, 21, 14, 30, 0, 0, time.UTC), "2026-05-21 14:30:00"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := FormatDate(tc.in); if got != tc.want → t.Errorf("...")
			_ = tc
		})
	}
}

// TestParseDateValid is a SKELETON. Pass valid date strings; assert
// ParseDate returns the expected time.Time + nil error.
func TestParseDateValid(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want time.Time
	}{
		// TODO: at least 2 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   got, err := ParseDate(tc.in)
			//   if err != nil → t.Fatalf("unexpected error: %v", err)
			//   if !got.Equal(tc.want) → t.Errorf("...")
			_ = tc
		})
	}
}

// TestParseDateInvalidWraps is a SKELETON. For invalid inputs, assert:
//   - err is non-nil
//   - err message contains the offending input
//   - errors.Unwrap returns a non-nil underlying time.Parse error
func TestParseDateInvalidWraps(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		// TODO: at least 2 cases ("not-a-date", "2026-13-01 99:99:99", etc.)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   _, err := ParseDate(tc.in)
			//   if err == nil → t.Fatalf("expected error for %q", tc.in)
			//   if !strings.Contains(err.Error(), tc.in) → t.Errorf("error message missing input")
			//   if errors.Unwrap(err) == nil → t.Errorf("expected wrapped error")
			_ = tc
			_ = strings.Contains
			_ = errors.Unwrap
		})
	}
}

// TestRoundTrip is a SKELETON. ParseDate(FormatDate(t)) should equal t
// (truncated to seconds — Format drops sub-second precision).
func TestRoundTrip(t *testing.T) {
	// TODO:
	//   original := time.Date(2026, 5, 21, 14, 30, 45, 0, time.UTC)
	//   s := FormatDate(original)
	//   parsed, err := ParseDate(s)
	//   if err != nil → t.Fatalf("ParseDate: %v", err)
	//   if !parsed.Equal(original) → t.Errorf("roundtrip: %v != %v", parsed, original)
}
```

- [ ] **Step 4:** Create `lessons/14-time-strings-regex/solutions/warmup/dates/dates.go`:

```go
// Package dates is the lesson 14 warm-up reference implementation.
package dates

import (
	"fmt"
	"time"
)

const layout = "2006-01-02 15:04:05"

// FormatDate returns t formatted with the lesson's "YYYY-MM-DD HH:MM:SS"
// layout, using Go's reference-time idiom.
func FormatDate(t time.Time) string {
	return t.Format(layout)
}

// ParseDate parses s with the lesson's layout. Wraps time.Parse errors
// with the offending input string.
func ParseDate(s string) (time.Time, error) {
	t, err := time.Parse(layout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("dates: invalid date %q: %w", s, err)
	}
	return t, nil
}
```

- [ ] **Step 5:** Create `lessons/14-time-strings-regex/solutions/warmup/dates/dates_test.go`:

```go
package dates

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestFormatDate(t *testing.T) {
	cases := []struct {
		name string
		in   time.Time
		want string
	}{
		{"epoch", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), "1970-01-01 00:00:00"},
		{"sample", time.Date(2026, 5, 21, 14, 30, 0, 0, time.UTC), "2026-05-21 14:30:00"},
		{"end-of-year", time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC), "2026-12-31 23:59:59"},
		{"single-digit-padding", time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), "2026-01-02 03:04:05"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatDate(tc.in)
			if got != tc.want {
				t.Errorf("FormatDate(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseDateValid(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want time.Time
	}{
		{"epoch", "1970-01-01 00:00:00", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"sample", "2026-05-21 14:30:00", time.Date(2026, 5, 21, 14, 30, 0, 0, time.UTC)},
		{"end-of-year", "2026-12-31 23:59:59", time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseDate(tc.in)
			if err != nil {
				t.Fatalf("ParseDate(%q) unexpected error: %v", tc.in, err)
			}
			if !got.Equal(tc.want) {
				t.Errorf("ParseDate(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseDateInvalidWraps(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"not-a-date", "not-a-date"},
		{"wrong-shape", "2026/05/21 14:30:00"},
		{"out-of-range", "2026-13-01 12:00:00"},
		{"empty", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseDate(tc.in)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
			if !strings.Contains(err.Error(), tc.in) && tc.in != "" {
				t.Errorf("error message %q doesn't mention input %q", err.Error(), tc.in)
			}
			if errors.Unwrap(err) == nil {
				t.Errorf("expected wrapped error, got %v", err)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	original := time.Date(2026, 5, 21, 14, 30, 45, 0, time.UTC)
	s := FormatDate(original)
	parsed, err := ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	if !parsed.Equal(original) {
		t.Errorf("round-trip: parsed=%v, original=%v", parsed, original)
	}
}
```

> Note on `t.Equal(other)`: comparing two `time.Time` values with `==` checks the wall-clock representation including monotonic clock reading and location. `Equal()` does a proper semantic comparison ignoring those, which is what you want in tests.

> Note on the empty-string test: `time.Parse("", "")` returns a meaningful error, but our wrapped message becomes `dates: invalid date "": parsing time "" as "2006-01-02 15:04:05": cannot parse "" as "2006"`. The `strings.Contains` check is conditional on non-empty input because otherwise we'd be asserting `contains("...","")` which is trivially true and doesn't catch the regression.

- [ ] **Step 6:** Verify and commit:

```bash
gofmt -l lessons/14-time-strings-regex/
go test ./lessons/14-time-strings-regex/exercises/warmup/dates/... -v 2>&1 | tail -10
go test ./lessons/14-time-strings-regex/solutions/warmup/dates/... -v 2>&1 | tail -25
make test
golangci-lint run ./...
go vet ./...

git add lessons/14-time-strings-regex/exercises/warmup/ lessons/14-time-strings-regex/solutions/warmup/
git commit -m "feat(lesson-14): warmup — dates (Format/Parse round-trip with reference time)"
```

Expected: gofmt empty; exercises pass vacuously; solutions all sub-tests PASS; make test green; lint 0 issues; vet clean.

---

## Task 3: Main — `logparse` subpackage

LogEntry struct + Parse function (regex + bufio.Scanner) + CountByLevel tally.

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/14-time-strings-regex/{exercises,solutions}/logparse`

- [ ] **Step 2:** Create `lessons/14-time-strings-regex/exercises/logparse/logparse.go`:

```go
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
//   1. s := bufio.NewScanner(r)
//   2. out := []LogEntry{}
//   3. for line := 1; s.Scan(); line++ {
//        text := strings.TrimSpace(s.Text())
//        if text == "" { continue }
//        e, err := parseLine(text)
//        if err != nil → return out, fmt.Errorf("logparse: line %d: %w", line, err)
//        out = append(out, e)
//      }
//   4. if err := s.Err(); err != nil → return out, fmt.Errorf("logparse: %w", err)
//   5. return out, nil
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
```

- [ ] **Step 3:** Create `lessons/14-time-strings-regex/exercises/logparse/logparse_test.go` (SKELETON):

```go
package logparse

import (
	"strings"
	"testing"
)

// TestParse is a SKELETON. Cover at least:
//   - Single well-formed line → one entry, no error
//   - Multiple well-formed lines → three entries
//   - Blank lines interspersed → skipped, count matches non-blank lines
//   - One malformed line in the middle → partial result + wrapped error mentioning line number
func TestParse(t *testing.T) {
	// TODO: build inputs with strings.NewReader; call Parse; assert structure.
	_ = strings.NewReader
}

// TestCountByLevel is a SKELETON. Cover at least:
//   - Mixed levels → correct counts per level
//   - All same level → single map entry with full count
//   - Empty input → empty map (or nil — either acceptable)
func TestCountByLevel(t *testing.T) {
	// TODO: build []LogEntry directly; call CountByLevel; assert map content.
}
```

- [ ] **Step 4:** Create `lessons/14-time-strings-regex/solutions/logparse/logparse.go`:

```go
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
```

- [ ] **Step 5:** Create `lessons/14-time-strings-regex/solutions/logparse/logparse_test.go`:

```go
package logparse

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func mkTime(s string) time.Time {
	t, err := time.Parse(timeLayout, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestParse(t *testing.T) {
	t.Run("single-entry", func(t *testing.T) {
		in := "2026-05-21T14:30:00 INFO connection accepted from 192.168.1.5"
		got, err := Parse(strings.NewReader(in))
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		want := []LogEntry{
			{Time: mkTime("2026-05-21T14:30:00"), Level: "INFO", Message: "connection accepted from 192.168.1.5"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Parse = %+v, want %+v", got, want)
		}
	})

	t.Run("multiple-entries", func(t *testing.T) {
		in := strings.Join([]string{
			"2026-05-21T14:30:00 INFO server started",
			"2026-05-21T14:31:00 WARN high memory usage",
			"2026-05-21T14:32:00 ERROR connection refused",
		}, "\n")
		got, err := Parse(strings.NewReader(in))
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if len(got) != 3 {
			t.Fatalf("got %d entries, want 3", len(got))
		}
		levels := []string{got[0].Level, got[1].Level, got[2].Level}
		want := []string{"INFO", "WARN", "ERROR"}
		if !reflect.DeepEqual(levels, want) {
			t.Errorf("levels = %v, want %v", levels, want)
		}
	})

	t.Run("blank-lines-skipped", func(t *testing.T) {
		in := "\n2026-05-21T14:30:00 INFO a\n\n2026-05-21T14:31:00 WARN b\n"
		got, err := Parse(strings.NewReader(in))
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("got %d entries (blank lines should be skipped), want 2", len(got))
		}
	})

	t.Run("malformed-line-returns-partial-and-error", func(t *testing.T) {
		in := strings.Join([]string{
			"2026-05-21T14:30:00 INFO ok",
			"garbage not a log line",
			"2026-05-21T14:32:00 WARN also ok",
		}, "\n")
		out, err := Parse(strings.NewReader(in))
		if err == nil {
			t.Fatal("expected error from malformed line")
		}
		if !strings.Contains(err.Error(), "line 2") {
			t.Errorf("error should mention line 2, got %v", err)
		}
		if len(out) != 1 {
			t.Errorf("expected 1 partial entry from line 1, got %d", len(out))
		}
	})
}

func TestCountByLevel(t *testing.T) {
	t.Run("mixed", func(t *testing.T) {
		entries := []LogEntry{
			{Level: "INFO"},
			{Level: "WARN"},
			{Level: "INFO"},
			{Level: "ERROR"},
			{Level: "INFO"},
		}
		got := CountByLevel(entries)
		want := map[string]int{"INFO": 3, "WARN": 1, "ERROR": 1}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("CountByLevel = %v, want %v", got, want)
		}
	})

	t.Run("all-same", func(t *testing.T) {
		entries := []LogEntry{
			{Level: "INFO"}, {Level: "INFO"}, {Level: "INFO"},
		}
		got := CountByLevel(entries)
		want := map[string]int{"INFO": 3}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("CountByLevel = %v, want %v", got, want)
		}
	})

	t.Run("empty", func(t *testing.T) {
		got := CountByLevel(nil)
		if len(got) != 0 {
			t.Errorf("CountByLevel(nil) = %v, want empty", got)
		}
	})
}
```

> Note on the `mkTime` helper: it's a test-local helper that panics on parse failure, which is the right call for a test fixture (a malformed literal in a test is a bug in the test, not a runtime condition). Same pattern as `regexp.MustCompile` at package level.

> Note on `FindStringSubmatch`: returns a `[]string` of `[full match, capture 1, capture 2, ...]` on a match, or `nil` if no match. So checking `m == nil` is the idiomatic "no match" test. Don't check `len(m) == 0`.

- [ ] **Step 6:** Verify and commit:

```bash
gofmt -l lessons/14-time-strings-regex/
go test ./lessons/14-time-strings-regex/exercises/logparse/... -v 2>&1 | tail -10
go test ./lessons/14-time-strings-regex/solutions/logparse/... -v 2>&1 | tail -25
make test
golangci-lint run ./...
go vet ./...

git add lessons/14-time-strings-regex/exercises/logparse/ lessons/14-time-strings-regex/solutions/logparse/
git commit -m "feat(lesson-14): main — logparse (regex + bufio.Scanner; CountByLevel)"
```

Expected: gofmt empty; exercises pass vacuously; solutions all sub-tests PASS; make test green.

---

## Task 4: Slide deck — 5 concepts

Heavy-explanatory pattern (Motivation/Basics/Worked-example/Common-mistake/Recap) per concept. Plus cover + roadmap + Practice + Closing thought + What-we-learned + Up-next.

**File:** `lessons/14-time-strings-regex/slides/slides.md`

The deck has ~30 slides (5 concepts × ~5 + ~5 framing).

**Concept order and structure:**

1. **Time formatting & parsing** — Motivation: every language has cryptic date directives; Go uses the literal reference time `2006-01-02 15:04:05`. Basics: layout string is the OUTPUT SHAPE in reference-time components. Worked: `dates.FormatDate` + `dates.ParseDate`. Common-mistake: writing `"YYYY-MM-DD"` (everyone tries this; it doesn't work). Brief duration mention: `t.Add(24*time.Hour)`, `t2.Sub(t1)` returns `time.Duration`. Recap.
2. **`strings` package** — Motivation: 90% of string manipulation needs are covered here. Basics: `Split`, `Join`, `Contains`, `TrimSpace`, `HasPrefix`/`HasSuffix`, `ToLower`/`ToUpper`, `Replace`. Worked: CSV-ish parsing chain (from L13 conceptually). Common-mistake: assuming strings are mutable; all `strings` functions return new strings. Recap.
3. **`bytes` package** — Motivation: same operations on `[]byte` instead of `string`; symmetric API surface. Basics: `bytes.Equal`, `bytes.Contains`, `bytes.Buffer` as efficient string-builder. Worked: building up a multi-line message with `bytes.Buffer`. Common-mistake: trying `[]byte == []byte` (compile error — slices aren't comparable). Recap.
4. **`regexp` basics** — Motivation: when string ops aren't enough, regex extracts structured pieces from variable shapes. Basics: `MustCompile` at package init (panic on bad regex), `FindStringSubmatch` returns full + captures, named captures with `(?P<name>...)`. Worked: `logparse`'s log-line regex with timestamp/level/message captures. Common-mistake: recompiling the regex inside the loop (use package-level `var = MustCompile`). Recap.
5. **"When to reach for what"** — Motivation: regex is a power tool that's easy to overuse. Basics: rule of thumb — `Contains`/`HasPrefix` for fixed strings, `Split` for known delimiters, regex for variable patterns or structured extraction. Worked: parsing "10:30am" with `Split` vs regex (Split wins). Common-mistake: reaching for regex by reflex when `strings.Contains` would do. Recap.

- [ ] **Step 1-7:** Author the deck section by section (cover → roadmap → 5 concepts → Practice → Closing thought → What-we-learned → Up-next). Concept content follows the structure above.

- [ ] **Step 8:** Verify slides build.

```bash
make slides-build
grep -q "14-time-strings-regex" dist/index.html && echo "✓ 14-time-strings-regex in index"
rm -rf dist
```

- [ ] **Step 9:** Commit:

```bash
git add lessons/14-time-strings-regex/slides/
git commit -m "feat(lesson-14): slides — Time, strings, bytes, regex (5 concepts)"
```

---

## Task 5: README

Mirrors slide concepts (5 sections); one Common-mistake paragraph per concept; gofmt/vet daily-habits block; "What's different" section noting L14 is standalone (the tracker returns in the L15 capstone).

**File:** `lessons/14-time-strings-regex/README.md`

- [ ] **Step 1:** Author the README (~220 lines):
  - "What you'll learn" — 5 bullets mirroring concepts.
  - "What's different from L13" — L13 brought the tracker back; L14 is standalone again (per Phase 2 spec). The tracker stays where L13 left it; L15 capstone reorganizes everything.
  - "The package layout" — annotated tree.
  - Per-concept sections (5) — each with Motivation/Key idea/Common mistake.
  - "Exercise: warm-up — dates" (~3 lines)
  - "Exercise: main — logparse" (~5 lines)
  - "Daily habits" — gofmt -w, go vet, go test.
  - "How to run" — both subpackage test commands.
  - "Going further" — Read (Go blog on time package; regexp docs); Try (extend logparse to extract IP addresses from messages via second regex; benchmark regex vs strings.Contains).

- [ ] **Step 2:** Commit:

```bash
git add lessons/14-time-strings-regex/README.md
git commit -m "docs(lesson-14): README — Time, strings, bytes, regex self-study"
```

---

## Task 6: End-to-end verification

- [ ] **Step 1:** Full test sweep.

```bash
make test
```

Expected: all lessons (01-14) + tools pass.

- [ ] **Step 2:** Solution-only verbose tests.

```bash
go test -v ./lessons/14-time-strings-regex/solutions/...
```

Expected (~15 sub-tests total):
- `dates.TestFormatDate` — 4 sub-tests PASS
- `dates.TestParseDateValid` — 3 sub-tests PASS
- `dates.TestParseDateInvalidWraps` — 4 sub-tests PASS
- `dates.TestRoundTrip` PASS
- `logparse.TestParse` — 4 sub-tests PASS
- `logparse.TestCountByLevel` — 3 sub-tests PASS

- [ ] **Step 3:** Exercise tests pass vacuously.

```bash
go test ./lessons/14-time-strings-regex/exercises/...
```

Expected: ok (no sub-tests run).

- [ ] **Step 4:** Static analysis.

```bash
go vet ./...
golangci-lint run ./...
gofmt -l lessons/14-time-strings-regex/
```

Expected: zero gofmt; 0 lint issues; vet clean.

- [ ] **Step 5:** Slides build smoke.

```bash
make slides-build
grep -q "14-time-strings-regex" dist/index.html
rm -rf dist
```

- [ ] **Step 6:** (No binary smoke test — L14 has no cmd binary by design.)

- [ ] **Step 7:** Final code review subagent → push branch → create PR.

```bash
git push -u origin feature/plan-q-lesson-14-time-strings-regex
gh pr create --title "feat(lesson-14): Time, strings, bytes, regex — log-line parser" --body "..."
```

---

## Verification

After all 6 tasks:

```bash
make test                                                                             # green
go test -v ./lessons/14-time-strings-regex/solutions/... 2>&1 | grep -E "PASS|FAIL"  # ~15 sub-tests PASS
go vet ./...                                                                          # clean
golangci-lint run ./...                                                               # 0 issues
gofmt -l lessons/14-time-strings-regex/                                               # empty
make slides-build && grep -q "14-time-strings-regex" dist/index.html && rm -rf dist  # green
```

## Critical file paths

To be created:

- `lessons/14-time-strings-regex/` (directory)
- `lessons/14-time-strings-regex/README.md`
- `lessons/14-time-strings-regex/slides/{index.html, slides.md, assets/.gitkeep}`
- `lessons/14-time-strings-regex/exercises/warmup/dates/{dates.go, dates_test.go}`
- `lessons/14-time-strings-regex/exercises/logparse/{logparse.go, logparse_test.go}`
- `lessons/14-time-strings-regex/solutions/...` (mirrored)

To be referenced (not modified):

- `docs/superpowers/specs/2026-05-18-phase-2-idiomatic-go-design.md` (Lesson 14 section)
- `lessons/13-encoding-io/...` (lesson 13 stays untouched; L14 introduces no carry-forward)
- `.golangci.yml` (no changes; existing exclusion `lessons/.*/exercises/` covers nested paths)
