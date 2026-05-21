# Lesson 14: Time, strings, bytes, regex

## What you'll learn

By the end of this lesson you can:

- Format and parse dates using Go's reference time `2006-01-02 15:04:05`.
- Reach for the right `strings` function (`Split`, `Join`, `Contains`, `TrimSpace`, `ToLower/ToUpper`, `Replace`) and chain them for common parsing.
- Use the `bytes` package as the byte-equivalent of `strings`, and know why `[]byte == []byte` is a compile error.
- Write Go regular expressions with `regexp.MustCompile` at package init, extract captures with `FindStringSubmatch`, and use named captures.
- Choose between string ops and regex deliberately — and recognise the Go-culture default of "regex is a tool of last resort."

## What's different from L13

L13 brought the tracker back and refactored `JSONStore` to use streaming Encoder/Decoder. Lesson 14 is **standalone again** — no tracker integration. The lesson exists in service of std-lib literacy: students see `time`, `strings`, `bytes`, `regexp` as a coherent toolkit rather than scattered packages used ad-hoc through earlier lessons.

The tracker stays where L13 left it. **L15 capstone** reorganises everything (cmd/ + internal/, golden file tests, benchmarks) — that's where the tracker resumes.

## The package layout

```
lessons/14-time-strings-regex/
├── exercises/
│   ├── warmup/dates/    ← you implement: FormatDate + ParseDate
│   └── logparse/         ← you implement: LogEntry, Parse, CountByLevel
└── solutions/            ← reference implementations
```

12 files total. Same footprint as L12 (also a standalone Phase 2 lesson).

---

## Concept 1 — Time formatting & parsing

Go's date formatting is famously different from every other language. Instead of cryptic directives (`%Y-%m-%d`, `YYYY-MM-DD`), you write the LITERAL reference time `2006-01-02 15:04:05` in the shape you want, and the formatter pattern-matches.

The reference moment, in full: `Mon Jan 2 15:04:05 MST 2006`. Numerically: 01/02 03:04:05PM '06 MST. Each component is a different digit, which makes the pattern unambiguous.

```go
const layout = "2006-01-02 15:04:05"

t.Format(layout)                                 // "2026-05-21 14:30:00"
got, err := time.Parse(layout, "2026-05-21 14:30:00")  // parses back
```

Duration arithmetic uses `time.Duration` (a typedef'd int64 of nanoseconds):

```go
t.Add(24 * time.Hour)
later.Sub(earlier)  // returns time.Duration
```

### Common mistake

Writing the layout with placeholder letters:

```go
t.Format("YYYY-MM-DD")  // returns "YYYY-MM-DD" literally
```

Go doesn't treat `YYYY` as a placeholder — it treats it as LITERAL characters. The reference-time digits ARE the placeholders. Use `2006-01-02`. Once seen, never forgotten.

The second classic mistake: getting the digits slightly wrong. `2006-1-2 15:4:5` doesn't work — Go uses the EXACT reference digits, not a sloppy version.

---

## Concept 2 — `strings` package

90% of string manipulation needs are covered here:

```go
strings.Split("a,b,c", ",")                // ["a", "b", "c"]
strings.Join([]string{"a","b"}, "-")       // "a-b"
strings.Contains("hello", "ell")           // true
strings.HasPrefix("https://x", "https://") // true
strings.TrimSpace("  hi  ")                // "hi"
strings.ToLower("HELLO")                   // "hello"
strings.Replace("hi hi", "hi", "bye", 1)   // "bye hi"
strings.ReplaceAll("hi hi", "hi", "bye")   // "bye bye"
```

All functions return NEW strings. Go strings are immutable — you can't modify them in place.

For many small concatenations, use `strings.Builder`:

```go
var b strings.Builder
for _, w := range words {
    b.WriteString(w)
}
result := b.String()
```

### Common mistake

Trying to mutate a string with `s[0] = 'H'` — Go gives a compile error. Strings are read-only byte sequences. Every operation on a string returns a NEW string. If you're building up output one piece at a time, use `strings.Builder` so allocations amortise; otherwise repeated `+` concatenation in a loop allocates a fresh string each time.

---

## Concept 3 — `bytes` package

`[]byte` is the byte-slice equivalent of `string` — same data, mutable in place, slightly different ergonomics. The `bytes` package mirrors the `strings` API:

```go
bytes.Split([]byte("a,b,c"), []byte(","))   // [][]byte
bytes.Contains(buf, []byte("HTTP/"))         // bool
bytes.HasPrefix(buf, []byte("GET "))         // bool
bytes.TrimSpace(buf)                         // []byte
bytes.Equal(a, b)                            // bool
```

And `bytes.Buffer` — a growable buffer that implements BOTH `io.Reader` AND `io.Writer`:

```go
var buf bytes.Buffer
buf.WriteString("hello ")
buf.WriteString("world")
io.Copy(os.Stdout, &buf)  // buf is also an io.Reader
```

Useful when you need to feed accumulated bytes to a function that takes a Reader.

### Common mistake

Comparing `[]byte` with `==`:

```go
a := []byte("hello")
b := []byte("hello")
if a == b { ... }   // compile error: slice can only be compared to nil
```

Slices (including `[]byte`) are NOT comparable with `==`. Use `bytes.Equal`. The same restriction applies to `[]int`, `[]string`, any other slice — but `[]byte` is where developers usually hit it first because the muscle-memory of `==` on strings is so strong.

---

## Concept 4 — `regexp` basics

When string operations aren't enough — when you need to match a PATTERN (multiple possible shapes) or extract structured pieces — regular expressions are the right tool. Go's `regexp` uses RE2 syntax: no backreferences, but linear-time guaranteed.

Compile once at package init:

```go
var emailRE = regexp.MustCompile(`^[a-z0-9._]+@[a-z0-9.-]+\.[a-z]+$`)
```

`MustCompile` panics at init if the pattern is invalid — exactly what you want for known-good patterns. Use `regexp.Compile` (returns error) only for user-supplied patterns.

The API:

```go
re.MatchString(s)                  // bool
re.FindString(s)                   // first match, or ""
re.FindStringSubmatch(s)           // [full, cap1, cap2, ...] or nil
re.FindAllString(s, -1)            // all matches
```

Capture groups are parenthesised parts:

```go
re := regexp.MustCompile(`(\d+)-(\d+)`)
m := re.FindStringSubmatch("score: 42-13")
// m == ["42-13", "42", "13"]
```

Named captures for clarity:

```go
re := regexp.MustCompile(`(?P<year>\d{4})-(?P<month>\d{2})`)
```

### Common mistake

Recompiling the regex inside a hot loop:

```go
func parseLine(s string) (LogEntry, error) {
    re := regexp.MustCompile(`...`)  // recompiles every call
    ...
}
```

`MustCompile` is expensive — it parses the pattern, builds the automaton. The idiomatic Go pattern is **package-level `var = MustCompile(...)`** — compile once, reuse forever. The Go compiler doesn't memoize for you.

---

## Concept 5 — When to reach for what

Regular expressions are powerful and easy to overuse. A regex is harder to read than `strings.Contains`. A regex with multiple captures is harder to maintain than a `strings.Split` + assertion. The Go culture leans heavily on strings ops first.

Rule of thumb:

| Need | Reach for |
|---|---|
| "Does the string contain X?" | `strings.Contains` |
| "Does it start/end with X?" | `strings.HasPrefix` / `strings.HasSuffix` |
| "Split on a known delimiter" | `strings.Split` |
| "Trim whitespace" | `strings.TrimSpace` |
| "Match a fixed list" | `switch` on `strings.ToLower(s)` |
| "Match a variable PATTERN" | `regexp` |
| "Extract structured pieces" | `regexp` with capture groups |

Regex earns its place when there's PATTERN VARIATION — the timestamp might have different digit counts, the level is one-of-N, the message is variable-length text. For FIXED strings, `strings.Contains` wins on clarity and speed.

### Common mistake

Reaching for regex by reflex:

```go
// the wrong way
if regexp.MustCompile(`^https://`).MatchString(url) { ... }

// the right way
if strings.HasPrefix(url, "https://") { ... }
```

The second is shorter, faster (no compile, no automaton walk), and obvious at a glance. Don't summon the regex hammer when a screwdriver will do.

---

## Exercise: warm-up — `dates`

Implement `FormatDate(t time.Time) string` and `ParseDate(s string) (time.Time, error)` in `exercises/warmup/dates/dates.go`. Tests cover format/parse correctness across multiple dates plus the round-trip property `ParseDate(FormatDate(t)) == t`.

**Time:** 5-10 minutes.

## Exercise: main — `logparse`

Build a small log-line parser in `exercises/logparse/logparse.go`:

1. **`LogEntry` struct** with `Time time.Time`, `Level string`, `Message string`.
2. **`Parse(r io.Reader) ([]LogEntry, error)`** using a package-level `MustCompile`'d regex over a `bufio.Scanner`. Three capture groups (timestamp, level, message). Parse timestamp with `time.Parse`. Return partial entries + line-numbered wrapped error on first malformed line.
3. **`CountByLevel(entries []LogEntry) map[string]int`** — pure tally helper.

This is the same shape as L13's `csvimport.Parse` (`(r io.Reader) → ([]Entry, error)`) — recognising the parallel is part of the lesson.

**Time:** 15-25 minutes.

---

## Daily habits

After every change:

```bash
gofmt -w ./...           # auto-format
go vet ./...             # catches struct-tag typos, shadowed vars, format-verb mismatches
go test ./...            # run the suite
```

Particularly useful in this lesson: `go vet` flags regex strings that don't compile.

## How to run

```bash
# Skeleton (passes vacuously until you implement)
cd lessons/14-time-strings-regex/exercises
go test ./...

# Or run a specific subpackage
go test ./warmup/dates -v
go test ./logparse -v

# Reference solution
go test ./lessons/14-time-strings-regex/solutions/... -v
```

## Going further

### Read

- **Go blog — "The Go Memory Model" (not strictly relevant, but linked from the time package docs):** <https://go.dev/ref/mem>
- **`time` package docs** — read the package overview; the reference-time discussion is at the top: <https://pkg.go.dev/time>
- **`regexp/syntax` package** — the RE2 syntax reference Go's regexp accepts: <https://pkg.go.dev/regexp/syntax>
- **"Regular expression matching can be simple and fast" (Russ Cox, 2007)** — why Go uses RE2 instead of PCRE-style backtracking: <https://swtch.com/~rsc/regexp/regexp1.html>

### Try

- **Extend `logparse` to extract IP addresses.** Add a second regex that finds IPv4 addresses inside log messages. Add a method `LogEntry.IPs() []string` that returns all IP-looking substrings. Test with messages like `"connection from 10.0.0.5 to 192.168.1.10"`.
- **Benchmark regex vs strings.** Write two functions that count occurrences of `"INFO"` at the start of lines — one using `regexp.MustCompile(`^INFO`)`, one using `strings.HasPrefix`. Benchmark with `func BenchmarkX(b *testing.B)`. Compare ns/op.
- **Parse multiple timestamp formats.** Some logs use `2006-01-02T15:04:05`, others `2006-01-02 15:04:05`. Add `parseFlexibleTime(s string) (time.Time, error)` that tries both layouts. (Hint: `time.Parse` errors are cheap; nested try blocks are fine here.)
