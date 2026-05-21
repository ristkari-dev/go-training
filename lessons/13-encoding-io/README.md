# Lesson 13: Encoding & I/O

## What you'll learn

By the end of this lesson you can:

- Read from and write to anything that satisfies `io.Reader` / `io.Writer` — files, network sockets, in-memory buffers, stdin/stdout.
- Use `bufio.Scanner` for line-based input from any `io.Reader`, and know to always end with `s.Err()`.
- Recognise the three classic struct-tag gotchas in `encoding/json`: unexported fields, `omitempty` zero-values, silent typos.
- Choose between `Marshal`/`Unmarshal` (whole-blob) and `Encoder`/`Decoder` (streaming) for JSON I/O.
- Build read-transform-write pipelines that compose Scanner/Decoder + transform + Encoder/Writer.

## What's different from L12

L12 was standalone (generics in isolation). L13 brings back the **tracker** — and finally opens up the `storage` internals that have been a black box since L08. `JSONStore` is refactored from whole-blob `Marshal`/`Unmarshal` to streaming `Encoder`/`Decoder` (same public API, same observable behavior, observably-tested by the same test suite).

A new binary joins the tracker: `cmd/expenses-import` reads CSV from stdin (or a file) and writes JSON via the existing `store.JSONStore`. It's a streaming pipeline from end to end and the worked example for concept 5.

## The package layout

```
lessons/13-encoding-io/
├── exercises/
│   ├── warmup/countlines/       ← you implement: CountLines(io.Reader)
│   ├── expense/                  ← verbatim from L11 (no changes)
│   ├── store/                    ← you REFACTOR Load + Save to Encoder/Decoder
│   ├── csvimport/                ← you implement: Parse(io.Reader)
│   ├── cmd/expenses/             ← verbatim from L11 (the tracker CLI still works)
│   └── cmd/expenses-import/      ← you wire: stdin → csvimport.Parse → store.Save
└── solutions/                    ← reference implementations + integration tests
```

The `store` refactor is the most interesting bit. The public API and test suite stay identical; only the internals change. The point: streaming JSON is a drop-in replacement for whole-blob JSON when the destination/source is a file or stream.

---

## Concept 1 — `io.Reader` and `io.Writer`

The two MOST important interfaces in the Go standard library — each with **one method**:

```go
type Reader interface { Read(p []byte) (n int, err error) }
type Writer interface { Write(p []byte) (n int, err error) }
```

Almost every type that touches bytes implements one or both: `*os.File`, `*strings.Reader`, `*bytes.Buffer`, `*net.TCPConn`, compressors, cryptographic streams. Functions that accept `io.Reader` or `io.Writer` work with all of them — no per-source code needed.

A function that takes `io.Reader` is testable without writing real files: `strings.NewReader("hello\nworld")` is a Reader over a string literal. The L13 warmup CountLines uses exactly this pattern in its test suite.

### Common mistake

Over-constraining a signature with `*os.File` when `io.Reader` would do. `*os.File` is concrete — your function won't work with `strings.NewReader` (in tests), `bytes.Buffer` (for in-memory pipelines), or `os.Stdin` (well — actually that one's a `*os.File`, but you get the idea). Lean on the interface; the L10 rule "accept interfaces" applies here at its sharpest.

---

## Concept 2 — `bufio.Scanner`

`io.Reader.Read` is byte-oriented. Most code wants lines. `bufio.NewScanner(r)` wraps any Reader and gives you the canonical scan loop:

```go
s := bufio.NewScanner(r)
for s.Scan() {
    line := s.Text()  // current line, no trailing \n
    // ... process line
}
if err := s.Err(); err != nil {
    // mid-stream error, not EOF
}
```

Two things to internalise:

- **`Scan()` returns `false` on either EOF or error.** You can't tell the difference until you check `Err()` after the loop.
- **`Text()` strips the trailing newline.** `"foo\n"` and `"foo"` both produce one Scan iteration with `Text() == "foo"`.

### Common mistake

Forgetting `s.Err()` at the end of the loop. Without it, a reader that fails mid-stream (a flaky network connection, a half-corrupt file, a closed pipe) looks identical to clean EOF. Your code silently processes a partial input as complete. Bake the `Err()` check into your fingers.

---

## Concept 3 — Struct tags revisited

Tags are how `encoding/json` knows the JSON key for each field:

```go
type Expense struct {
    Date     string  `json:"date"`
    Amount   float64 `json:"amount"`
    Category string  `json:"category"`
}
```

The backtick string is just a string — Go doesn't parse it. `encoding/json` reads tags via reflection at run time. Other packages reuse the convention with their own keys (`xml:`, `yaml:`, validator libraries, ORM tags).

Three gotchas:

1. **Unexported fields are invisible.** `date string` (lowercase d) — encoding/json can't see it via reflection regardless of tags. Output JSON silently omits the field. Capitalise to expose.
2. **`omitempty` omits zero values.** Surprising for `bool` (false IS zero), `int` (0 IS zero), and `string` (empty IS zero). If you need to distinguish "absent" from "explicitly zero", use a pointer (`*int`, `*bool`).
3. **Typos compile silently.** `josn:"date"` — Go doesn't validate; the typo gets ignored, the JSON key defaults to "Date" (capitalised field name). `go vet ./...` catches some struct-tag typos. Run it.

### Common mistake

Putting `omitempty` on a `bool` field meaning "user-controlled toggle". When the user sets it to `false` (which is meaningful!), it gets serialized as ABSENT. The next reader can't distinguish "user said false" from "never set". Drop `omitempty` for booleans you care about.

---

## Concept 4 — `encoding/json`: Marshal/Unmarshal vs Encoder/Decoder

Go has two parallel JSON APIs. They're symmetric:

```go
// Whole-blob — convenient when bytes are already in memory
data, err := json.Marshal(v)
err := json.Unmarshal(data, &v)

// Streaming — better when source/sink is a file or network connection
enc := json.NewEncoder(w); enc.Encode(v)
dec := json.NewDecoder(r); dec.Decode(&v)
```

Same struct tags apply to both. Same behavior for typical inputs. The streaming version doesn't allocate a `[]byte` copy of the whole document — for a 100MB file, that's the difference between OK and OOM.

The L13 `store.JSONStore` refactor IS this swap: same `JSONStore.Load`/`JSONStore.Save` public API, but the internals went from `os.ReadFile` + `json.Unmarshal` to `os.Open` + `json.NewDecoder(f).Decode`. The test suite is the same. The behavior is the same. The memory profile is better.

### Common mistake

Reaching for `Marshal`/`Unmarshal` on a large file when streaming would do. The whole-blob version reads the entire file into memory (`os.ReadFile`) before parsing — for a 500MB JSON file you've now allocated 500MB just to start. The streaming version reads as needed. Default to streaming for any file-backed JSON; reach for whole-blob only for small fixed payloads (config files, request bodies, etc.).

---

## Concept 5 — Streaming patterns + `bufio.Writer`

A lot of real-world I/O is a **read-transform-write** pipeline:

```
io.Reader → [Scanner | Decoder] → transform → [Writer | Encoder] → io.Writer
```

Each stage is a small interface with a small set of operations. The L13 `cmd/expenses-import` binary is exactly this shape:

```
stdin → bufio.Scanner → csvimport.Parse → []expense.Expense → json.Encoder → file
```

`bufio.Writer` is the symmetric counterpart of `bufio.Scanner`: many small writes get amortised into fewer system calls. Useful for any "many small writes" pattern (log lines, CSV rows, anything line-by-line outgoing).

### Common mistake

Forgetting `w.Flush()` on a `bufio.Writer`. The buffer holds pending writes; if your program exits before the flush, the data sits in memory and is lost. Always `defer w.Flush()` immediately after `bufio.NewWriter(...)`. The bug is silent (program runs, no error) which makes it harder to spot.

`json.Encoder` doesn't have this gotcha — `Encode` flushes after each value via the wrapped Writer. But if you wrapped that Writer in a `bufio.Writer` first, the bufio buffer still needs its own flush.

---

## Exercise: warm-up — `countlines`

Implement `CountLines(r io.Reader) (int, error)` in `exercises/warmup/countlines/countlines.go` using `bufio.Scanner`. Two test functions exercise the golden path (empty, single line, multiple lines, trailing newline) and the error-propagation path (via `io.MultiReader` glueing a clean reader to an errReader stub).

**Time:** 5-10 minutes.

## Exercise: main — store refactor + csvimport + cmd/expenses-import

Three pieces, all in `exercises/`:

1. **Refactor `store.JSONStore`** to use `Encoder`/`Decoder` instead of `Marshal`/`Unmarshal`. The public API and tests are unchanged — your refactor must keep the same observable behavior. Hints in the file: open with `os.Open`/`os.Create`, defer close, use `json.NewDecoder(f).Decode(&es)` and `enc := json.NewEncoder(f); enc.SetIndent("", "  "); enc.Encode(es)`.
2. **Implement `csvimport.Parse(r io.Reader)`** using `bufio.Scanner` + `strings.Split(line, ",")` + `strconv.ParseFloat`. Wrap errors with line numbers: `fmt.Errorf("csvimport: line %d: %w", n, err)`. Skip blank lines.
3. **Wire `cmd/expenses-import`** — parse `-file=<path>` (required) and `-input=<path>` (optional, defaults to stdin). Call `csvimport.Parse`, then `store.JSONStore.Save`. Exit 1 on errors with stderr message.

The integration tests in `exercises/cmd/expenses-import/main_test.go` build the binary into `t.TempDir()` and exec it — they're skeletons in exercises, full tests in solutions.

**Time:** 30-45 minutes.

---

## Daily habits

After every change:

```bash
gofmt -w ./...           # auto-format
go vet ./...             # catches struct-tag typos, shadowed vars, format-verb mismatches
go test ./...            # run the suite
```

`go vet` is especially valuable in this lesson — it catches `json:"data"` when you meant `json:"date"` (typos that compile silently otherwise).

## How to run

```bash
# Skeleton (passes vacuously until you implement)
cd lessons/13-encoding-io/exercises
go test ./...

# Reference solution
go test -v ./lessons/13-encoding-io/solutions/...

# Smoke test the importer binary
TMP=$(mktemp -d)
echo "2026-05-21,4.50,coffee" | go run ./lessons/13-encoding-io/solutions/cmd/expenses-import -file="$TMP/expenses.json"
cat "$TMP/expenses.json"
rm -rf "$TMP"

# Use the existing tracker CLI (carried forward from L11)
go run ./lessons/13-encoding-io/solutions/cmd/expenses -file=/tmp/tracker.json add 2026-05-21 4.50 coffee
go run ./lessons/13-encoding-io/solutions/cmd/expenses -file=/tmp/tracker.json list
```

## Going further

### Read

- **Go blog — "JSON and Go" (2011, still authoritative)** — the canonical walkthrough: <https://go.dev/blog/json>
- **`bufio` package docs** — Scanner's split functions (`ScanWords`, `ScanRunes`, custom): <https://pkg.go.dev/bufio>
- **`io` package docs** — beyond Reader/Writer: `io.Copy`, `io.Pipe`, `io.TeeReader`: <https://pkg.go.dev/io>

### Try

- **Extend the importer to support TSV.** Add a `-format=csv|tsv` flag (default `csv`). For TSV, split on `\t` instead of `,`. The change should fit inside `csvimport` with a parameterised `Parse(r io.Reader, sep rune)` — or as a separate `ParseTSV` if you prefer.
- **Export the other direction.** Add `cmd/expenses-export` that reads the JSON file via `store.JSONStore.Load` and writes CSV to stdout using `bufio.Writer`. Don't forget the `defer w.Flush()`.
- **Stream a million rows.** Generate a synthetic 1M-row CSV (`for i in {1..1000000}; do echo "2026-05-21,$i,gen"; done > big.csv`) and import it. Compare memory usage between the streaming `Encoder` version (current L13) and a whole-blob version (would need to rewrite Save with `MarshalIndent`). `/usr/bin/time -v go run ...` on Linux; `gtime -v` (Homebrew) on macOS.
