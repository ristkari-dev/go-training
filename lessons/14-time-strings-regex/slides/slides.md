<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">14</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 2 — Idiomatic Go</div>
<h1>Time, strings, bytes, regex</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Learn the <code>time</code>, <code>strings</code>, <code>bytes</code>, and <code>regexp</code> packages as a coherent std-lib toolkit. Master Go's quirky reference-time format <code>2006-01-02 15:04:05</code>, the common string/byte operations, the basics of regular expressions, and the Go-culture attitude that "regex is a tool of last resort."</p>
</div>
</div>
</div>

---

## What we'll cover

- **Time formatting & parsing** — the famous Go reference time `2006-01-02 15:04:05`.
- **`strings` package** — `Split`, `Join`, `Contains`, `TrimSpace`, `ToLower`/`ToUpper`, `Replace`.
- **`bytes` package** — same operations on `[]byte`; the `==` on slices gotcha.
- **`regexp` basics** — `MustCompile`, `FindStringSubmatch`, capture groups.
- **When to reach for what** — string ops cover 90%; regex is a power tool, easy to overuse.

---

## Concept 1: Time formatting & parsing

### Motivation

Every language has a way to format dates. Most use cryptic single-letter directives — `%Y-%m-%d`, `YYYY-MM-DD`, etc. Go does it differently: you write the LITERAL reference time `2006-01-02 15:04:05` in the shape you want, and the formatter pattern-matches.

The reference moment, in full:

```
Mon Jan 2 15:04:05 MST 2006
```

Numerically: 01/02 03:04:05PM '06 MST. Each component is a different digit, which makes the pattern unambiguous.

---

### The basics

Layout string = the OUTPUT SHAPE expressed in terms of the reference moment:

```go
t := time.Date(2026, 5, 21, 14, 30, 0, 0, time.UTC)

t.Format("2006-01-02")           // "2026-05-21"
t.Format("15:04:05")             // "14:30:00"
t.Format("2006-01-02 15:04:05")  // "2026-05-21 14:30:00"
t.Format("Mon Jan 2")            // "Thu May 21"
```

The same string works for parsing:

```go
got, err := time.Parse("2006-01-02 15:04:05", "2026-05-21 14:30:00")
// got == time.Date(2026, 5, 21, 14, 30, 0, 0, time.UTC)
```

Duration arithmetic uses `time.Duration` (a typedef'd int64 of nanoseconds):

```go
t.Add(24 * time.Hour)            // tomorrow
later.Sub(earlier)               // returns time.Duration
```

---

### A worked example

The L14 warm-up:

```go
const layout = "2006-01-02 15:04:05"

func FormatDate(t time.Time) string {
    return t.Format(layout)
}

func ParseDate(s string) (time.Time, error) {
    t, err := time.Parse(layout, s)
    if err != nil {
        return time.Time{}, fmt.Errorf("dates: invalid date %q: %w", s, err)
    }
    return t, nil
}
```

Same layout string, both directions. The reference-time idiom collapses what would be two separate format mini-languages in other ecosystems.

---

### Common mistake

Writing the layout as `"YYYY-MM-DD HH:MM:SS"`:

```go
// the wrong way — everyone tries this on day one
t.Format("YYYY-MM-DD")  // returns "YYYY-MM-DD" literally
```

Go doesn't treat `YYYY` etc. as placeholders — it treats them as LITERAL characters. The reference-time digits ARE the placeholders. You have to use `2006-01-02`.

The second classic mistake: getting the reference numbers slightly wrong. `2006-01-02 15:04:05` works; `2006-1-2 15:4:5` does not (Go uses the EXACT reference digits, not a sloppy version).

---

### Recap

- Go's date formatting uses the literal reference time `2006-01-02 15:04:05`, not `%Y-%m-%d`.
- The layout string describes the OUTPUT SHAPE in reference-moment components.
- Same string works for `Format` (time → string) and `Parse` (string → time).
- `time.Duration` for arithmetic; `t.Add(24*time.Hour)`, `t2.Sub(t1)`.

---

## Concept 2: `strings` package

### Motivation

90% of string manipulation needs are covered by the `strings` package. Splitting on a delimiter, joining a slice, checking for substrings, trimming whitespace, changing case — all there, all efficient, all return new strings (Go strings are immutable).

---

### The basics

The dozen or so functions you'll reach for most:

```go
strings.Split("a,b,c", ",")            // ["a", "b", "c"]
strings.Join([]string{"a","b"}, "-")   // "a-b"
strings.Contains("hello world", "world")   // true
strings.HasPrefix("https://x", "https://") // true
strings.HasSuffix("file.txt", ".txt")      // true
strings.TrimSpace("  hi  ")            // "hi"
strings.ToLower("HELLO")               // "hello"
strings.ToUpper("hello")               // "HELLO"
strings.Replace("hi hi", "hi", "bye", 1)   // "bye hi" (n=1)
strings.ReplaceAll("hi hi", "hi", "bye")   // "bye bye"
strings.Index("hello", "ll")           // 2
strings.Count("hello", "l")            // 2
```

All return new strings. Strings in Go are read-only byte sequences — you can't modify them in place.

---

### A worked example

L13's `csvimport.parseLine` chained several `strings` operations:

```go
func parseLine(s string) (expense.Expense, error) {
    fields := strings.Split(s, ",")              // split on comma
    if len(fields) != 3 { ... }
    date     := strings.TrimSpace(fields[0])      // trim whitespace
    amountStr := strings.TrimSpace(fields[1])
    category := strings.TrimSpace(fields[2])
    amount, err := strconv.ParseFloat(amountStr, 64)
    ...
}
```

L14's `logparse.Parse` does similar trimming:

```go
text := strings.TrimSpace(s.Text())
if text == "" { continue }  // skip blank lines after trim
```

Pattern: `Split` for known delimiters; `TrimSpace` before further parsing; `Contains` / `HasPrefix` for membership checks.

---

### Common mistake

Trying to mutate a string:

```go
// the wrong way
s := "hello"
s[0] = 'H'   // ❌ compile error: cannot assign to s[0]
```

Strings are immutable in Go. Every `strings` function returns a NEW string. If you find yourself building up a string with many concatenations, use `strings.Builder` (or `bytes.Buffer` — concept 3):

```go
var b strings.Builder
for _, word := range []string{"hello", " ", "world"} {
    b.WriteString(word)
}
result := b.String()  // "hello world"
```

`strings.Builder` amortises the underlying allocation across many small writes.

---

### Recap

- `strings` has the everyday string operations. Memorise the names.
- All functions return new strings — Go strings are immutable.
- For many-small-writes, use `strings.Builder`.

---

## Concept 3: `bytes` package

### Motivation

`[]byte` is the byte-slice equivalent of `string` — same data, mutable in place, slightly different ergonomics. The `bytes` package mirrors the `strings` API: same function names, same semantics, just `[]byte` instead of `string`.

You'll reach for `bytes` when you're processing binary data, reading from `io.Reader` directly into a buffer, or building up output efficiently with `bytes.Buffer`.

---

### The basics

The symmetric API:

```go
bytes.Split([]byte("a,b,c"), []byte(","))              // [][]byte
bytes.Contains([]byte("hello"), []byte("ll"))          // true
bytes.HasPrefix(buf, []byte("HTTP/"))                  // true
bytes.TrimSpace([]byte("  hi  "))                      // []byte("hi")
bytes.Equal([]byte("a"), []byte("a"))                  // true
```

And `bytes.Buffer` — a growable byte buffer that implements BOTH `io.Reader` AND `io.Writer`:

```go
var buf bytes.Buffer
buf.WriteString("hello ")
buf.WriteString("world")
fmt.Println(buf.String())   // "hello world"
```

`bytes.Buffer` is what `strings.Builder` would be if it also satisfied `io.Reader`. Useful when you need to feed accumulated bytes to a function that takes a Reader.

---

### A worked example

A simple HTTP-ish header parser using `bytes.Buffer` and `bytes.HasPrefix`:

```go
var buf bytes.Buffer
buf.WriteString("HTTP/1.1 200 OK\r\n")
buf.WriteString("Content-Type: application/json\r\n")

if bytes.HasPrefix(buf.Bytes(), []byte("HTTP/1.1 200")) {
    // OK response
}
```

Or building up output efficiently when writing to an `io.Writer`:

```go
var buf bytes.Buffer
for _, e := range entries {
    fmt.Fprintln(&buf, e.Format())  // buf is an io.Writer
}
io.Copy(os.Stdout, &buf)            // buf is also an io.Reader
```

---

### Common mistake

Comparing `[]byte` with `==`:

```go
// the wrong way
a := []byte("hello")
b := []byte("hello")
if a == b { ... }   // ❌ compile error: invalid operation: a == b (slice can only be compared to nil)
```

Slices (including `[]byte`) are NOT comparable with `==` in Go. You have to use `bytes.Equal`:

```go
if bytes.Equal(a, b) { ... }   // ✓
```

The same gotcha applies to `[]int`, `[]string`, and any other slice type — but `[]byte` is where developers hit it first because the muscle-memory of `==` on strings is so strong.

---

### Recap

- `bytes` mirrors `strings` for `[]byte` data.
- `bytes.Buffer` is a growable buffer that's both an `io.Reader` and an `io.Writer`.
- Use `bytes.Equal` for `[]byte` equality — `==` is a compile error on slices.
- Strings are immutable; `[]byte` is mutable in place.

---

## Concept 4: `regexp` basics

### Motivation

When string operations aren't enough — when you need to match a PATTERN (multiple possible shapes of input) or extract structured pieces — regular expressions are the right tool. Go's `regexp` package uses RE2 syntax (no backreferences, but linear-time guaranteed — no catastrophic backtracking).

---

### The basics

Compile once at package init via `MustCompile`:

```go
var emailRE = regexp.MustCompile(`^[a-z0-9._]+@[a-z0-9.-]+\.[a-z]+$`)
```

`MustCompile` panics at package init if the pattern is invalid — which is what you want for known-good patterns. Use `regexp.Compile` (returns error) only for user-supplied patterns.

The match/find API:

```go
re.MatchString(s)                  // bool — does s match the pattern?
re.FindString(s)                   // string — first match (or "" if none)
re.FindStringSubmatch(s)           // []string — [full match, capture1, capture2, ...] or nil
re.FindAllString(s, -1)            // []string — all matches
re.FindAllStringSubmatch(s, -1)    // [][]string — all matches with captures
```

Capture groups are parenthesised parts of the pattern:

```go
re := regexp.MustCompile(`(\d+)-(\d+)`)
m := re.FindStringSubmatch("score: 42-13")
// m == ["42-13", "42", "13"]
//      ↑ full   ↑ cap1 ↑ cap2
```

Named captures `(?P<name>...)` are clearer for complex patterns:

```go
re := regexp.MustCompile(`(?P<year>\d{4})-(?P<month>\d{2})`)
m := re.FindStringSubmatch("2026-05")
// Access by index (m[1], m[2]) or by name via re.SubexpNames()
```

---

### A worked example

L14's log-line regex with three capture groups (timestamp, level, message):

```go
var logLineRE = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})\s+(INFO|WARN|ERROR)\s+(.+)$`)

func parseLine(s string) (LogEntry, error) {
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
```

Three captures: timestamp goes through `time.Parse`; level is one of INFO/WARN/ERROR; message is everything else.

---

### Common mistake

Recompiling the regex inside a hot loop:

```go
// the wrong way — recompiles every call
func parseLine(s string) (LogEntry, error) {
    re := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})\s+...`)
    m := re.FindStringSubmatch(s)
    ...
}
```

`MustCompile` is expensive — it parses the pattern, builds the automaton, etc. Doing it on every call wastes work. The idiomatic Go pattern is **package-level `var = MustCompile(...)`** — compile once, reuse forever.

---

### Recap

- Use `regexp.MustCompile` at package init; never inside a loop.
- `FindStringSubmatch` returns `[full, cap1, cap2, ...]` or `nil` — the `nil` check is the "no match" test.
- Named captures with `(?P<name>...)` for clarity in complex patterns.
- Go's RE2 has no backreferences but guarantees linear-time matching.

---

## Concept 5: When to reach for what

### Motivation

Regular expressions are powerful — they let you parse patterns that would take pages of hand-written code. They're also **easy to overuse**. A regex is harder to read than `strings.Contains`. A regex with multiple captures is harder to maintain than a `strings.Split` + assertion. The Go culture leans heavily on strings ops first, regex only when needed.

---

### The basics

A rule of thumb:

| Need | Reach for |
|---|---|
| "Does the string contain X?" | `strings.Contains` |
| "Does it start/end with X?" | `strings.HasPrefix` / `strings.HasSuffix` |
| "Split on a known delimiter" | `strings.Split` |
| "Trim whitespace / specific chars" | `strings.TrimSpace` / `strings.Trim` |
| "Match a fixed list of options" | `switch` on `strings.ToLower(s)` |
| **"Match a variable PATTERN"** | `regexp` |
| **"Extract structured pieces"** | `regexp` with capture groups |

Regex earns its place when there's PATTERN VARIATION — the timestamp might have varying digit counts, the level is one-of-N, the message is variable-length text. Fixed strings → `strings.Contains` wins on clarity AND speed.

---

### A worked example

Parsing `"10:30am"` two ways:

```go
// (a) strings.Split — clear, fast, no regex compile cost
parts := strings.Split(s[:len(s)-2], ":")
hour, _ := strconv.Atoi(parts[0])
minute, _ := strconv.Atoi(parts[1])
ampm := s[len(s)-2:]

// (b) regex — works, but heavier
re := regexp.MustCompile(`^(\d+):(\d+)(am|pm)$`)
m := re.FindStringSubmatch(s)
hour, _ := strconv.Atoi(m[1])
minute, _ := strconv.Atoi(m[2])
ampm := m[3]
```

Both work. The Split version is shorter to read, doesn't require thinking about regex syntax, and avoids the `MustCompile` cost. For a fixed shape like `"H:Mam"`, prefer Split.

When the pattern HAS variation — say, optionally-prefixed log levels, or timestamps in multiple formats, or extractable IP addresses inside free-form messages — regex starts to earn its keep.

---

### Common mistake

Reaching for regex by reflex when string ops would do:

```go
// the wrong way — overkill
if regexp.MustCompile(`^https://`).MatchString(url) { ... }

// the right way
if strings.HasPrefix(url, "https://") { ... }
```

The second version is shorter, faster (no regex compile, no automaton walk), and obvious at a glance. Don't summon the regex hammer when a screwdriver will do.

---

### Recap

- String ops cover ~90% of text needs. Try them first.
- Regex earns its place for VARIABLE patterns and structured extraction.
- Don't use regex for fixed strings — `Contains` / `HasPrefix` are faster and clearer.
- The Go community's reflex: **prefer string ops; reach for regex only when justified.**

---

## Practice

### Warm-up

Implement `FormatDate(t time.Time) string` returning `"YYYY-MM-DD HH:MM:SS"` and `ParseDate(s string) (time.Time, error)` as its inverse. The round-trip property — `ParseDate(FormatDate(t)) == t` (truncated to seconds) — is tested.

```bash
cd lessons/14-time-strings-regex/exercises/warmup/dates
go test -v
```

---

### Main

Build a small log-line parser in `logparse`:

1. **`LogEntry` struct** with `Time`, `Level`, `Message` fields.
2. **`Parse(r io.Reader) ([]LogEntry, error)`** using a package-level `MustCompile`'d regex over a `bufio.Scanner`. Capture timestamp + level + message; parse timestamp with `time.Parse`.
3. **`CountByLevel(entries []LogEntry) map[string]int`** — pure tally helper.

```bash
cd lessons/14-time-strings-regex/exercises/logparse
go test -v
```

Note:
Same shape as L13's `csvimport.Parse` — `(r io.Reader) → ([]Entry, error)`. Recognising this is part of the lesson: library functions consuming an `io.Reader` return a typed slice plus an error.

---

## Closing thought

Go's standard library has the same character as the language: small, composable pieces with names that mean what they say. `strings.Split`, `bytes.Equal`, `time.Parse`, `regexp.MustCompile` — none of them are surprising once you know the rules.

The reference-time format is the most-likely-to-be-surprising thing in this lesson, and even that's just "write the output shape in literal reference-time digits." Once seen, never forgotten.

---

## What we learned

- Go's reference-time format: `2006-01-02 15:04:05`. Layout = output shape.
- `strings` package covers the common string operations. Functions return new strings.
- `bytes` mirrors `strings` for `[]byte`. Use `bytes.Equal`, not `==`, for comparison.
- `regexp.MustCompile` at package init; `FindStringSubmatch` returns `[full, cap1, ...]`.
- String ops first; regex only when there's real pattern variation.

---

## Up next

Lesson 15 — **Idiomatic project structure & testing patterns (CAPSTONE)**. The tracker that's evolved through L09/L10/L11/L13 gets reorganised into a polished v2: `cmd/` + `internal/`, subtests, `t.Helper()`, golden file tests, benchmarks, fuzz syntax. Phase 2 wraps up.
