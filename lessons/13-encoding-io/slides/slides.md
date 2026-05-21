<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">13</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 2 — Idiomatic Go</div>
<h1>Encoding &amp; I/O</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Learn the building blocks of Go I/O: the <code>io.Reader</code>/<code>io.Writer</code> interfaces formally, <code>bufio.Scanner</code> for line-based streaming, the struct-tag gotchas in <code>encoding/json</code>, the <code>Marshal</code>/<code>Unmarshal</code>-vs-<code>Encoder</code>/<code>Decoder</code> distinction, and read-transform-write streaming patterns with <code>bufio.Writer</code>.</p>
</div>
</div>
</div>

---

## What we'll cover

- **`io.Reader` and `io.Writer`** — the two prototype interfaces of Go I/O. Two methods each.
- **`bufio.Scanner`** — line-by-line reading from any `io.Reader`. The scan loop pattern.
- **Struct tags revisited** — `json:"name,omitempty"`; the unexported-field gotcha; the `omitempty` gotcha.
- **`encoding/json` — Marshal/Unmarshal vs Encoder/Decoder** — whole-blob vs streaming.
- **Streaming patterns + `bufio.Writer`** — the read-transform-write idiom; the importer binary in action.

---

## Concept 1: `io.Reader` and `io.Writer`

### Motivation

In lesson 10 we learned that **small interfaces are powerful** — `error` has one method, `Store` has two. The two MOST important interfaces in the Go standard library each have **one method** (`io.Reader` has `Read`; `io.Writer` has `Write`). Almost every piece of code that touches bytes implements one or both. Files. Network sockets. Buffers. Compressors. Cryptographic streams. They all satisfy these two interfaces, which means functions can accept "any source of bytes" or "any destination for bytes" without caring what's underneath.

---

### The basics

The whole interfaces, in full:

```go
package io

type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}
```

That's it. Read fills `p` with up to `len(p)` bytes; returns how many it actually read plus an error (most importantly, `io.EOF` on clean end-of-stream). Write copies bytes from `p` to the destination; returns how many it actually wrote plus any error.

Standard library types that satisfy these:

| Type | Reader | Writer |
|---|---|---|
| `*os.File` | yes | yes |
| `*strings.Reader` | yes | — |
| `*bytes.Buffer` | yes | yes |
| `*bufio.Reader` / `*bufio.Writer` | yes / — | — / yes |
| `*net.TCPConn` | yes | yes |

---

### A worked example

The L13 warm-up, `CountLines`, accepts an `io.Reader`:

```go
func CountLines(r io.Reader) (int, error) {
    s := bufio.NewScanner(r)
    var n int
    for s.Scan() { n++ }
    if err := s.Err(); err != nil { return n, err }
    return n, nil
}
```

Same function, three different callers:

```go
// (a) Counting lines in a real file.
f, _ := os.Open("README.md")
defer f.Close()
n, _ := CountLines(f)

// (b) Counting lines in a string literal (no temp file needed in tests).
n, _ := CountLines(strings.NewReader("a\nb\nc\n"))

// (c) Counting lines from stdin.
n, _ := CountLines(os.Stdin)
```

CountLines doesn't know or care which of these it's reading from. **Accept `io.Reader`** — the consumer gets to pick the source.

---

### Common mistake

Over-constraining a function signature by taking `*os.File` instead of `io.Reader`:

```go
// the wrong way
func CountLines(f *os.File) (int, error) { ... }
```

This works for case (a) above, but breaks (b) and (c) — `strings.NewReader` doesn't return `*os.File`. **Accept the smallest interface that does the job.** This is the same "accept interfaces, return structs" rule from lesson 10.

---

### Recap

- `io.Reader` and `io.Writer` are 1-method interfaces. Most Go I/O fits one or both.
- Functions that touch bytes should accept `io.Reader` or `io.Writer`, not concrete types.
- `*os.File` implements both. `strings.NewReader` is a Reader. `bytes.Buffer` is both. `os.Stdin` and `os.Stdout` are real `*os.File` values.

---

## Concept 2: `bufio.Scanner`

### Motivation

`io.Reader.Read` is byte-oriented — it fills a `[]byte`. Most real-world code wants to work line by line: log files, CSV rows, configuration files, stdin input. Without a helper, you'd write a byte-by-byte loop that buffers up bytes until you see `\n`. `bufio.Scanner` does that for you.

---

### The basics

```go
import "bufio"

s := bufio.NewScanner(r)  // r is any io.Reader

for s.Scan() {            // returns false on EOF OR error
    line := s.Text()      // current line, no trailing \n
    // ... do something with line
}

if err := s.Err(); err != nil {
    // Scan returned false because of an error, not clean EOF
}
```

Three pieces to internalise:

- **`Scan()` returns a bool.** False means "no more lines available" — but the reason could be EOF (good) or a real error (bad). You don't find out until you check `Err()` at the end.
- **`Text()` returns the current line as a string.** No trailing `\n` — Scanner's default `SplitFunc` (`bufio.ScanLines`) strips it. Empty lines are returned as `""` (a Scan iteration happens; `Text()` is empty).
- **`s.Err()` after the loop.** Without this check, a reader that fails mid-stream is indistinguishable from clean EOF.

---

### A worked example

The warm-up's `CountLines` is the textbook Scanner usage:

```go
func CountLines(r io.Reader) (int, error) {
    s := bufio.NewScanner(r)
    var n int
    for s.Scan() {
        n++
    }
    if err := s.Err(); err != nil {
        return n, err
    }
    return n, nil
}
```

The CSV importer's `Parse` uses the same shape with one extra step (parse each line):

```go
func Parse(r io.Reader) ([]expense.Expense, error) {
    s := bufio.NewScanner(r)
    out := []expense.Expense{}
    for line := 1; s.Scan(); line++ {
        text := strings.TrimSpace(s.Text())
        if text == "" { continue }       // skip blanks
        e, err := parseLine(text)
        if err != nil {
            return out, fmt.Errorf("csvimport: line %d: %w", line, err)
        }
        out = append(out, e)
    }
    if err := s.Err(); err != nil {
        return out, fmt.Errorf("csvimport: %w", err)
    }
    return out, nil
}
```

Scanner is doing all the line-buffering work. Your code just decides what to do with each line.

---

### Common mistake

Forgetting the `s.Err()` check:

```go
// the wrong way — silent failure
for s.Scan() {
    process(s.Text())
}
// returns happily even if the reader failed at byte 4093
```

If the underlying reader (a flaky network connection, a half-corrupt file, a stdin pipe that was closed early) errored, Scan returns false — same as clean EOF. Without the `Err()` check, you can't tell the difference, and you'll silently process a partial input as if it were complete. **Always end the scan loop with `Err()`.**

---

### Recap

- `bufio.NewScanner(r)` wraps any `io.Reader` for line-by-line reading.
- `for s.Scan() { line := s.Text() }` is the canonical pattern.
- **Always** call `s.Err()` after the loop to detect mid-stream errors.

---

## Concept 3: Struct tags revisited

### Motivation

You've been using `` `json:"date"` `` tags on `expense.Expense` since lesson 08. They're how `encoding/json` knows which JSON key maps to which Go field. This lesson is a deep dive into the corners — three gotchas that bite even experienced Go developers.

---

### The basics

A struct tag is a string literal in backticks that lives between the field name and the field type:

```go
type Expense struct {
    Date     string  `json:"date"`
    Amount   float64 `json:"amount,omitempty"`
    Category string  `json:"category"`
}
```

The backtick value is just a string — Go itself doesn't parse it. Packages like `encoding/json` read the tags via reflection at run time. The conventional format is `name,opt1,opt2`. Other packages (`xml`, `yaml`, struct validators, ORMs) use the same syntax with their own keys: `` `xml:"name,attr" yaml:"name,omitempty"` ``.

Default behavior with **no** tag: encoding/json uses the field name as the JSON key (so `Date` → `"Date"`, capitalised).

---

### A worked example

The Expense JSON before and after the tags:

```go
// With tags:
type Expense struct {
    Date string `json:"date"`
}
// → {"date":"2026-05-21"}

// Without tags:
type Expense struct {
    Date string
}
// → {"Date":"2026-05-21"}     ← capital D, often wrong
```

The tag changes the key in BOTH directions — Marshal uses it to emit, Unmarshal uses it to parse. The expense file format you've been using since L08 is downstream of these tags.

---

### Common mistake (three of them)

**(a) Unexported fields are invisible to encoding/json.** Lowercase first letter means no serialization, regardless of tags:

```go
type expenseBad struct {
    date string `json:"date"`   // unexported → not serialized
}
```

The compiler doesn't warn you. The output JSON just silently doesn't have the field. Fix: capitalise the field name.

**(b) `omitempty` omits zero values — including `0`, `false`, and `""`.** Not what you usually want:

```go
type Activity struct {
    Steps    int  `json:"steps,omitempty"`
    Achieved bool `json:"achieved,omitempty"`
}

a := Activity{Steps: 0, Achieved: false}
// → {}                          ← both omitted because 0 and false ARE zero values
```

Fix if you genuinely need to distinguish "absent" from "zero": use a pointer (`*int`, `*bool`). Then `nil` means absent, `&zero` means "explicitly zero".

**(c) Typos in tags compile silently.** Tags are strings — Go doesn't validate them:

```go
type Expense struct {
    Date string `josn:"date"`   // "josn" not "json" → tag ignored, key becomes "Date"
}
```

There's no compile error, no runtime warning. The tag is silently ignored and you get the default capitalised key. `go vet ./...` catches some struct-tag typos — run it.

---

### Recap

- Struct tags are strings; the syntax is convention, parsed at run time via reflection.
- Unexported fields are invisible to `encoding/json` regardless of tags — capitalise to expose.
- `omitempty` omits zero values — surprising for `bool`/`int`. Use pointers if you need to distinguish.
- `go vet` catches some tag typos. Run it.

---

## Concept 4: `encoding/json` — Marshal/Unmarshal vs Encoder/Decoder

### Motivation

There are **two** ways to do JSON in Go, and they're symmetric. The whole-blob version (`Marshal`/`Unmarshal`) operates on `[]byte` — convenient when you already have or want bytes in memory. The streaming version (`Encoder`/`Decoder`) operates on `io.Writer`/`io.Reader` — better when the source/sink is a file, a network connection, or anything else you'd rather not slurp into RAM all at once.

---

### The basics

Whole-blob:

```go
data, err := json.Marshal(v)              // v → []byte
err := json.Unmarshal(data, &v)           // []byte → v
```

Streaming:

```go
enc := json.NewEncoder(w)                 // wraps any io.Writer
enc.SetIndent("", "  ")                   // optional: pretty-print
err := enc.Encode(v)                      // v → w (writes one JSON value + \n)

dec := json.NewDecoder(r)                 // wraps any io.Reader
err := dec.Decode(&v)                     // reads one JSON value from r
```

Same data; different plumbing. Encoder/Decoder don't allocate a copy of your bytes — they pipe directly between the reader/writer and the Go value.

---

### A worked example

The L11 `JSONStore.Load` was whole-blob:

```go
// LESSON 11 — whole-blob version
func (j *JSONStore) Load() ([]expense.Expense, error) {
    data, err := os.ReadFile(j.Path)   // file → []byte (whole file in memory)
    if err != nil { ... }
    var es []expense.Expense
    if err := json.Unmarshal(data, &es); err != nil { ... }
    return es, nil
}
```

The L13 version is streaming:

```go
// LESSON 13 — streaming version
func (j *JSONStore) Load() ([]expense.Expense, error) {
    f, err := os.Open(j.Path)          // open the file
    if err != nil { ... }
    defer f.Close()
    var es []expense.Expense
    dec := json.NewDecoder(f)          // wrap the *os.File
    if err := dec.Decode(&es); err != nil { ... }
    return es, nil
}
```

Same public API. Same observable behavior. But the streaming version never materialises the whole file as a `[]byte` — it reads from `f` as needed. For a 100MB JSON file, that's the difference between OK and OOM.

The Save side mirrors:

```go
// LESSON 13 — streaming Save
f, _ := os.Create(j.Path)
defer f.Close()
enc := json.NewEncoder(f)
enc.SetIndent("", "  ")
return enc.Encode(es)
```

---

### Common mistake

Using whole-blob on a large file when streaming would do:

```go
// the wrong way — for a 500MB log file
data, _ := os.ReadFile("big.json")    // 500MB allocated
var rows []Row
json.Unmarshal(data, &rows)           // another copy while parsing
```

vs.

```go
// the right way
f, _ := os.Open("big.json")
defer f.Close()
var rows []Row
json.NewDecoder(f).Decode(&rows)      // streams; minimal extra allocation
```

For small fixed payloads (config files, request bodies of a few KB) `Marshal`/`Unmarshal` is fine and slightly more ergonomic. For anything that grows with input size, prefer the streaming version.

---

### Recap

- Two JSON APIs: `Marshal`/`Unmarshal` (whole-blob, `[]byte`) and `Encoder`/`Decoder` (streaming, `io.Reader`/`io.Writer`).
- Streaming wraps any Reader/Writer; no extra `[]byte` allocation.
- For large inputs, streaming. For small fixed payloads, whole-blob is fine.
- Both APIs respect struct tags identically — you can switch without changing the structs.

---

## Concept 5: Streaming patterns + `bufio.Writer`

### Motivation

A lot of real-world I/O code is a **read-transform-write** pipeline. Read CSV → parse → write JSON. Read log lines → filter → write to another file. Read gzipped bytes → decompress → write plain text. The shape is always the same; the building blocks are the ones you've now seen.

---

### The basics

The pipeline:

```
io.Reader → [bufio.Scanner | json.Decoder] → transform → [bufio.Writer | json.Encoder] → io.Writer
```

Each stage:

- **Input side** — `bufio.Scanner` for line-based; `json.Decoder` for JSON; `csv.Reader` for CSV (we don't use this, but it exists).
- **Transform** — your application code. Parse, validate, filter, project.
- **Output side** — `bufio.Writer` for line-based; `json.Encoder` for JSON; `csv.Writer` for CSV.

`bufio.Writer` is the symmetric counterpart of `bufio.Scanner`: many small writes get amortised into fewer system calls. Always call `w.Flush()` before close.

---

### A worked example

The L13 importer binary, `cmd/expenses-import`, IS this pipeline:

```go
func run(args []string, inputReader io.Reader) error {
    filePath, inputPath, err := parseFlags(args)
    if err != nil { return err }

    r := inputReader
    if inputPath != "" {
        f, _ := os.Open(inputPath)
        defer f.Close()
        r = f
    }

    // STAGE 1: read CSV from stdin or file (bufio.Scanner inside Parse)
    es, err := csvimport.Parse(r)
    if err != nil { return err }

    // STAGE 2: write JSON to file (json.Encoder inside Save)
    return store.NewJSONStore(filePath).Save(es)
}
```

Bytes flow: stdin → bufio.Scanner → Go structs → json.Encoder → file. No stage holds the whole input in memory unnecessarily.

If we wanted the importer to also write one-expense-per-line to stdout while it worked (a "tee" pattern), `bufio.Writer` would be exactly the right tool:

```go
w := bufio.NewWriter(os.Stdout)
defer w.Flush()                                // critical
for _, e := range es {
    fmt.Fprintln(w, e.Format())                // many small writes; bufio batches them
}
```

---

### Common mistake

Forgetting `bufio.Writer.Flush()`:

```go
// the wrong way
w := bufio.NewWriter(os.Stdout)
fmt.Fprintln(w, "hello")
// program exits; output buffer never flushed; nothing on stdout
```

`bufio.Writer` holds writes in a buffer until either the buffer fills (default 4KB) or you call `Flush()`. If your program ends before the flush happens (and you didn't `defer w.Flush()`), the data sits in the buffer and is lost. **Always `defer w.Flush()` right after creation.**

The same gotcha doesn't apply to `json.Encoder` — `Encode` flushes after each value via the underlying Writer. But if you wrapped the Writer in a `bufio.Writer` first, the bufio buffer still needs flushing.

---

### Recap

- Read-transform-write is the canonical streaming pipeline.
- Input side: `bufio.Scanner` or `json.Decoder`. Output side: `bufio.Writer` or `json.Encoder`.
- `bufio.Writer` batches small writes for performance.
- **Always** `defer w.Flush()` immediately after `bufio.NewWriter(...)`.

---

## Practice

### Warm-up

Implement `CountLines(r io.Reader) (int, error)` using `bufio.Scanner`. Two tests: golden path (empty, single line, multiple lines, trailing newline) and error propagation via `io.MultiReader`.

```bash
cd lessons/13-encoding-io/exercises/warmup/countlines
go test -v
```

---

### Main

Three pieces:

1. **Refactor `store.JSONStore`** from `Marshal`/`Unmarshal` to `Encoder`/`Decoder`. Same public API; same tests must pass.
2. **Implement `csvimport.Parse(r io.Reader)`** using `bufio.Scanner` + `strings.Split(line, ",")` + `strconv.ParseFloat`. Wrap errors with line numbers.
3. **Wire up `cmd/expenses-import`** — reads CSV from stdin (or `-input=<path>`), writes JSON to `-file=<path>`.

```bash
cd lessons/13-encoding-io/exercises
go test ./...
```

Manual smoke test:

```bash
echo "2026-05-21,4.50,coffee" | go run ./cmd/expenses-import -file=/tmp/expenses.json
cat /tmp/expenses.json
```

Note:
The `-file` flag is required (no default). Forgetting it should print a helpful error and exit 1, not crash. The integration tests cover this.

---

## Closing thought

A theme of Go: **lots of small interfaces compose into surprisingly powerful systems**. `io.Reader` and `io.Writer` are each one method; `bufio.Scanner` and `json.Decoder` wrap them; `*os.File` satisfies them. By the time you've written `csvimport.Parse(os.Stdin)`, you're using a stack of four or five packages, each of which only knows about the one or two interfaces it consumes.

This is what people mean when they say Go has small standard library APIs: not that the library is small (it isn't), but that the APIs themselves are small enough to fit in your head.

---

## What we learned

- `io.Reader` and `io.Writer` are 1-method interfaces; most Go I/O fits them.
- `bufio.Scanner` reads lines from any Reader; always end with `s.Err()`.
- Struct tags have three classic gotchas: unexported fields, `omitempty` zero-values, silent typos.
- `encoding/json` has two APIs: whole-blob (`Marshal`/`Unmarshal`) and streaming (`Encoder`/`Decoder`).
- Read-transform-write streaming pipelines compose Scanner/Decoder + transform + Encoder/Writer.

---

## Up next

Lesson 14 — **Time, strings, bytes, regex**. Standalone again: `time.Time` and Go's quirky reference-time format `2006-01-02 15:04:05`, the `strings` and `bytes` packages, and `regexp` basics. We build a small log-line parser. The tracker stays where L13 left it.
