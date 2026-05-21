<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">15</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 2 — Idiomatic Go · CAPSTONE</div>
<h1>Project structure &amp; testing patterns</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Reorganise the tracker into the idiomatic <code>cmd/</code> + <code>internal/</code> layout. Learn Go's testing toolkit beyond table tests: <code>t.Helper</code>, <code>t.TempDir</code>, golden file tests with <code>-update</code>, benchmarks (<code>testing.B</code>), and a fuzz preview (<code>testing.F</code>). Wraps up Phase 2.</p>
</div>
</div>
</div>

---

## What we'll cover

- **`cmd/` + `internal/`** — project layout. `internal/` is compiler-enforced visibility.
- **Test helpers** — `t.Run` subtests, `t.Helper()`, `t.TempDir()`.
- **Golden file tests** — `testdata/<name>.golden` + `-update` flag.
- **Benchmarks** — `testing.B`, the `b.N` loop, `b.ResetTimer()`, `b.ReportAllocs()`.
- **Fuzz tests** — `testing.F`, `f.Add` seeds, `f.Fuzz` body, `go test -fuzz=`.

---

## Concept 1: `cmd/` and `internal/`

### Motivation

Through Phase 2 the tracker grew into a sprawling pile of subpackages — `expense`, `store`, `csvimport`, all imported by `cmd/expenses` and `cmd/expenses-import`. Without any organising principle, anyone reading the repo has to figure out which packages are "public API" and which are "implementation detail" by reading every doc comment.

Go has TWO directory conventions that solve this with no runtime cost: `cmd/` (binaries) and `internal/` (private packages, compiler-enforced).

---

### The basics

`cmd/<binary>/main.go` — each subdirectory is its own `package main`. Idiomatic since L07-L08.

`internal/<pkg>/` — packages here are importable ONLY from code in the parent directory or its descendants. A sibling module trying to import them fails at compile time:

```
use of internal package github.com/your/module/internal/store not allowed
```

No opt-out. No annotation. The compiler enforces it just from the directory name.

For a repository at `lessons/15-structure/`, `internal/store/` is visible to everything under `lessons/15-structure/`. Code outside that subtree — say, `lessons/16-other/` or `tools/`, or some downstream module pulling us in via `go get` — cannot import it.

---

### A worked example

The L15 v2 tracker layout:

```
lessons/15-structure/exercises/
├── cmd/
│   ├── expenses/main.go              ← imports internal/expense, store, summary
│   └── expenses-import/main.go       ← imports internal/csvimport, store
└── internal/
    ├── expense/                       ← Expense struct + methods
    ├── store/                         ← Store interface + JSONStore + MemoryStore
    ├── csvimport/                     ← CSV → []Expense parser
    ├── summary/                       ← TotalsByCategory + BiggestCategory + Bar
    └── testutil/                      ← AssertGolden helper for golden file tests
```

Both binaries can import any internal package (they're descendants of `lessons/15-structure/exercises/`). Nothing outside that subtree can.

Refactoring `internal/store` to add a new method? Safe. The Go compiler verifies that the only callers are inside the same subtree — you literally cannot break a downstream consumer because there are no downstream consumers.

---

### Common mistake

Nesting `cmd/` inside `internal/`:

```
internal/
└── cmd/
    └── tool/main.go   ← weird; the binary is inside internal/
```

This compiles, but it's confusing: `cmd/` signals "public-facing binary"; `internal/` signals "private". Putting one inside the other contradicts both. Keep them as siblings:

```
cmd/
└── tool/main.go
internal/
└── pkg/...
```

---

### Recap

- `cmd/<binary>/main.go` — one main per subdirectory.
- `internal/<pkg>/` — compiler-enforced "private" packages.
- A package under `internal/` is visible only to code in the parent of `internal/` and its descendants.
- Refactor internal/ freely; you can't break unknown consumers.

---

## Concept 2: Test helpers

### Motivation

`testing.T` has a small but useful set of methods beyond `t.Fatal` and `t.Errorf`. Three you'll reach for often: `t.Run` for sub-tests, `t.Helper()` for cleaner failure messages, and `t.TempDir()` for auto-cleaned per-test workspaces.

You've seen all three in earlier lessons. This concept names them and teaches the gotchas.

---

### The basics

**`t.Run(name, func(t *testing.T) { ... })`** — runs the body as a sub-test. The sub-test gets its own t, can fail independently, and shows up in `go test -v` output as `TestX/name`. Standard for table tests:

```go
for _, tc := range cases {
    t.Run(tc.name, func(t *testing.T) {
        got := DoSomething(tc.in)
        if got != tc.want { t.Errorf("...") }
    })
}
```

**`t.Helper()`** — marks the current function as a test helper. When the helper fails, Go reports the line in the CALLER, not inside the helper:

```go
func assertEqual(t *testing.T, got, want int) {
    t.Helper()    // ← without this, failures report assertEqual's line
    if got != want {
        t.Errorf("got %d, want %d", got, want)
    }
}
```

Without `t.Helper()`, every failure points to the same line inside `assertEqual` — useless. With it, the failure points to YOUR code where you called `assertEqual` — useful.

**`t.TempDir()`** — returns a fresh empty directory for this test. Auto-deleted when the test finishes (whether it passes, fails, or panics):

```go
dir := t.TempDir()
path := filepath.Join(dir, "expenses.json")
// ...write a fixture, run a test, no cleanup needed
```

---

### A worked example

The lesson's `testutil.AssertGolden` uses two of these:

```go
func AssertGolden(t *testing.T, got []byte, path string) {
    t.Helper()    // ← failures point to the test, not this line
    if *Update {
        // ... write file, return
    }
    want, _ := os.ReadFile(path)
    if !bytes.Equal(got, want) {
        t.Errorf("output mismatch with %s: ...", path)
    }
}
```

And the integration test for `cmd/expenses-import` uses `t.TempDir()`:

```go
dir := t.TempDir()
bin := filepath.Join(dir, "expenses-import")
cmd := exec.Command("go", "build", "-o", bin, ".")
// ... no cleanup of dir; t.TempDir handles it
```

---

### Common mistake

Forgetting `t.Helper()` in a helper function. The test reports failure at line 42 of `assertEqual.go` instead of line 17 of `my_test.go`. You stare at the unhelpful traceback, eventually realise you forgot the call. Bake it into your fingers: every helper that calls `t.Errorf` / `t.Fatal` starts with `t.Helper()`.

---

### Recap

- `t.Run(name, fn)` for sub-tests — required for table tests, useful in any non-trivial test.
- `t.Helper()` in helpers so failures report the CALLER's line.
- `t.TempDir()` for per-test temp directories. Auto-cleanup.

---

## Concept 3: Golden file tests

### Motivation

For commands that produce structured output (a JSON file, a multi-line report, an HTML snippet), inline `want := "<huge string>"` in the test source is unmaintainable. You can't see whitespace differences. Diffs in code review become noise. Pasting a 50-line expected output into a Go string literal is painful.

Golden files solve this: put the expected output in a separate file, compare against it. When the output legitimately changes, regenerate with a flag.

---

### The basics

**Layout** — Go reserves `testdata/` as a directory that the build system ignores. Put your golden files there:

```
cmd/expenses/
├── main.go
├── main_test.go
└── testdata/
    └── summary.golden
```

**The helper** — `AssertGolden(t, got, path)` reads the file, compares to `got`, fails with diff on mismatch. With a `-update` flag, writes `got` to the file instead.

```go
var Update = flag.Bool("update", false, "regenerate golden files")

func AssertGolden(t *testing.T, got []byte, path string) {
    t.Helper()
    if *Update {
        _ = os.WriteFile(path, got, 0o644)
        t.Logf("updated %s", path)
        return
    }
    want, _ := os.ReadFile(path)
    if !bytes.Equal(got, want) {
        t.Errorf("output mismatch with %s:\n--- got ---\n%s\n--- want ---\n%s",
            path, got, want)
    }
}
```

**Usage**:

```go
go test ./cmd/expenses/                  # normal run; compares
go test -update ./cmd/expenses/          # regenerate golden after intentional changes
```

---

### A worked example

The L15 `cmd/expenses` summary command:

```go
func TestSummaryGolden(t *testing.T) {
    s := store.NewMemoryStore()
    _ = s.Save([]expense.Expense{...})           // seed
    var buf bytes.Buffer
    _ = cmdSummary(s, &buf)                       // capture output
    testutil.AssertGolden(t, buf.Bytes(),
        filepath.Join("testdata", "summary.golden"))
}
```

The golden file:

```
4 expenses, total €100.00
by category:
  coffee     €8.00   
  groceries  €80.00  
  lunch      €12.00  
biggest: groceries €80.00
  ████████████████████
```

When `cmdSummary`'s format string changes, run `go test -update ./cmd/expenses/` once and commit the regenerated `summary.golden`. The PR diff makes the output change visible to reviewers.

---

### Common mistake

Forgetting to run `-update` after intentional changes. Test fails with a huge multi-line diff; the dev panics and either (a) reverts the legitimate change or (b) edits the golden file by hand. Both are wrong. Once you've verified the new output is correct, just `go test -update ./...` and commit the regenerated files alongside your code change. The golden files ARE source code — version them, review them.

The other common mistake: not committing the golden file at all. Empty `testdata/` is the same as no golden file. The first `go test` then fails with "no such file" — confusing if the dev forgot they generated it locally and never committed.

---

### Recap

- `testdata/<name>.golden` for expected output. Go ignores `testdata/` automatically.
- `AssertGolden(t, got, path)` compares; `-update` regenerates.
- Commit the golden file — it's source code.
- Regenerate intentionally; never edit by hand.

---

## Concept 4: Benchmarks

### Motivation

How fast IS your code? You can guess based on the algorithm. You can measure with `time` once. Or you can write a benchmark function that runs hundreds or thousands of iterations, reports ns/op, and rerun whenever you make changes — catching regressions before they ship.

Go's `testing.B` makes benchmarks first-class.

---

### The basics

A benchmark looks like a test function with `B` instead of `T`:

```go
func BenchmarkX(b *testing.B) {
    // setup (runs once)
    data := makeFixture()

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        _ = DoSomething(data)
    }
}
```

Go calls the function with successively larger `b.N` until it's confident in the measurement, then reports `ns/op` (nanoseconds per iteration) plus `B/op` and `allocs/op` if you call `b.ReportAllocs()`.

Two critical bits:

- **`b.ResetTimer()`** after setup. Without it, the fixture-building time counts against the benchmark — your "DoSomething benchmark" is actually mostly "makeFixture benchmark."
- **`b.ReportAllocs()`** — opt-in allocation reporting. Catches accidental boxing, intermediate slices, etc.

Run with:

```bash
go test -bench=BenchmarkX -benchmem ./pkg/
```

---

### A worked example

The L15 `internal/summary` benchmark:

```go
func BenchmarkTotalsByCategory(b *testing.B) {
    r := rand.New(rand.NewSource(42))           // fixed seed = reproducible
    categories := []string{"coffee", "lunch", "groceries", ...}
    es := make([]expense.Expense, 10000)
    for i := range es {
        es[i] = expense.Expense{
            Date:     fmt.Sprintf("2026-%02d-%02d", 1+r.Intn(12), 1+r.Intn(28)),
            Amount:   r.Float64() * 100,
            Category: categories[r.Intn(len(categories))],
        }
    }

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        _ = TotalsByCategory(es)
    }
}
```

Typical output:

```
BenchmarkTotalsByCategory-10    5000    212875 ns/op    8192 B/op    1 allocs/op
```

`-10` is the GOMAXPROCS suffix; `5000` is the chosen N; `212875 ns/op` says one TotalsByCategory call on 10k entries takes ~213 microseconds. The map allocation (`1 allocs/op`) is the result map itself.

---

### Common mistake

Doing setup inside the `b.N` loop:

```go
// the wrong way
for i := 0; i < b.N; i++ {
    es := makeBigSlice()          // setup INSIDE the loop
    _ = TotalsByCategory(es)
}
```

You're now benchmarking `makeBigSlice + TotalsByCategory`. Hoist setup out of the loop AND call `b.ResetTimer()` to discard the setup time.

The other classic mistake: comparing benchmarks across different machines or with different background load. Benchmark numbers are RELATIVE — useful for "is my new version faster than my old version?" Don't compare your laptop number to a CI runner number to a colleague's machine.

---

### Recap

- `func BenchmarkX(b *testing.B)` + the `for i := 0; i < b.N; i++` loop.
- `b.ResetTimer()` after setup; `b.ReportAllocs()` for allocation info.
- Run with `go test -bench=. -benchmem ./pkg/`.
- Benchmarks measure RELATIVE performance. Compare runs on the same machine, not across.

---

## Concept 5: Fuzz tests (preview)

### Motivation

Tests check that your code works for the inputs you THOUGHT of. Fuzz tests check that your code doesn't CRASH on inputs you didn't think of. Go 1.18 added fuzzing as a first-class feature; `testing.F` lets you write fuzz functions that the framework explores with millions of mutated inputs.

L22 (Phase 3) goes deep. This lesson previews the syntax so you've seen it once.

---

### The basics

A fuzz function takes `*testing.F`, registers seeds with `f.Add`, and calls `f.Fuzz` with a function the framework will invoke with mutated inputs:

```go
func FuzzX(f *testing.F) {
    f.Add("valid seed 1")
    f.Add("valid seed 2")

    f.Fuzz(func(t *testing.T, input string) {
        _, _ = ParseSomething(input)   // any panic fails the test
    })
}
```

Run two ways:

```bash
go test ./pkg/                              # just runs the seeds (fast)
go test -fuzz=FuzzX ./pkg/                  # explores; runs until interrupted
```

Without `-fuzz=`, `go test` runs the seeds only — your fuzz function acts like a regular table test. With `-fuzz=`, the framework starts MUTATING the seeds and trying random inputs, looking for any input that panics or fails an assertion.

When the fuzzer finds a crash, it saves the offending input to `testdata/fuzz/FuzzX/<hash>`. The next `go test` run includes that input automatically as a regression test.

---

### A worked example

The L15 `internal/csvimport` fuzz preview:

```go
func FuzzParseLine(f *testing.F) {
    f.Add("2026-05-21,4.50,coffee")
    f.Add("")
    f.Add("a,b,c")

    f.Fuzz(func(t *testing.T, line string) {
        // parseLine is package-private; this fuzz test lives in the
        // same package so it can call it directly.
        _, _ = parseLine(line)
    })
}
```

Three seeds (valid, empty, malformed). The fuzz body just calls `parseLine` and ignores its return values. Any panic — any unhandled error path inside `parseLine` — is a real bug. Running with `-fuzz=FuzzParseLine` for a few seconds explores thousands of mutated inputs.

Notice the test file's package is `csvimport` (not `csvimport_test`) — that lets the fuzz function call the unexported `parseLine` directly.

---

### Common mistake

Writing the fuzz function and never running with `-fuzz=`. Without that flag, `go test` runs only the SEEDS — three or four inputs you wrote by hand, which is just an awkward table test. The whole VALUE of fuzzing is in the exploration phase. Add a manual fuzz run to your release process: `go test -fuzz=FuzzX -fuzztime=30s ./pkg/` runs for 30 seconds and catches most low-hanging bugs.

---

### Recap

- `func FuzzX(f *testing.F)` + `f.Add(seed)` + `f.Fuzz(func(t, input) { ... })`.
- `go test` runs seeds only; `go test -fuzz=FuzzX` explores.
- Found crashes are saved to `testdata/fuzz/` as automatic regression tests.
- Phase 3 lesson 22 dives deeper.

---

## Practice

### Warm-up

Move a helper into `warmup/internal/util/format.go` and import it from `warmup/main.go`. Compiles cleanly because they share the parent `warmup/` directory. Try moving `main.go` one level UP (to `exercises/`) and watch the build break.

```bash
cd lessons/15-structure/exercises/warmup
go run .   # prints "€4.50"
```

---

### Main

The v2 reorg has three parts:

1. **Move existing code into `internal/<pkg>/`** — expense, store, csvimport, summary. Mostly mechanical: copy from L11/L13/L08, update import paths.
2. **Add the golden file test** — convert `cmd/expenses summary` to a golden file test using `internal/testutil.AssertGolden`. Commit `testdata/summary.golden`.
3. **Add a benchmark and a fuzz preview** — `BenchmarkTotalsByCategory` on 10k expenses; `FuzzParseLine` on csvimport's parseLine.

```bash
cd lessons/15-structure/exercises
go test ./...                                            # full suite
go test -bench=. ./internal/summary/                     # the benchmark
go test -fuzz=FuzzParseLine -fuzztime=5s ./internal/csvimport/  # fuzz (5 sec)
go test -update ./cmd/expenses/                          # regenerate golden after changes
```

Note:
This is a capstone. Most code is carry-forward; the lesson is in the REORGANIZATION + the new testing patterns. Take time to read the v2 layout and notice the relationships between `cmd/` and `internal/`.

---

## Closing thought — Phase 2 wrap-up

Seven lessons. We started L09 with a string-typed `Expense.Date` and ended L15 with the same field — but the surrounding code is unrecognisable.

**Pointers (L09)** gave methods that mutate state in place.
**Interfaces (L10)** let us swap `JSONStore` for `MemoryStore` without changing callers.
**Errors (L11)** made the tracker's failure modes legible — sentinel for missing files, custom type for parse errors.
**Generics (L12)** added a side library; we learned restraint.
**Encoding & I/O (L13)** brought streaming JSON and a CSV importer.
**Time/strings/regex (L14)** added the std-lib literacy for parsing real-world data.
**Project structure (L15)** organised everything into the layout you'd ship.

Phase 2's overarching theme: **the standard library is enough**. Through 7 lessons we've added zero third-party dependencies. Go's stdlib is large enough for almost everything a backend engineer needs day-to-day.

---

## Up next

Lesson 16 — Phase 3 begins. **Concurrency**: goroutines and channels. The single hardest concept in Go for newcomers; we'll spend three lessons (16, 17, 18) on it. Lesson 19 covers context and cancellation; 20-21 cover networking and HTTP servers; 22 dives deep into fuzz tests; 23 wraps Phase 3 with a Phase-3 capstone.

Lessons 24-29 (Phase 4) cover real-world deployment: Docker, observability, a polished HTTP service, and a multi-week capstone project.

You're a third of the way through. The hard part is behind you.
