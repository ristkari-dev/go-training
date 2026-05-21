# Lesson 15: Project Structure & Testing Patterns

> **Phase 2 capstone.**

## What you'll learn

By the end of this lesson you can:

- Organise a Go project with the `cmd/` + `internal/` layout and understand the compiler-enforced visibility rule.
- Use `t.Run` sub-tests, `t.Helper()` (so failures report the CALLER), and `t.TempDir()` (auto-cleaned per-test directories).
- Write golden file tests: expected output in `testdata/<name>.golden`, comparison with `-update` flag for regeneration.
- Write benchmarks with `testing.B` — the `b.N` loop, `b.ResetTimer()`, `b.ReportAllocs()`.
- Recognise fuzz test syntax — `testing.F`, `f.Add`, `f.Fuzz` — and know to run with `-fuzz=` for actual exploration.

## What's different from L14

L14 was standalone (time/strings/bytes/regex toolkit). **L15 is the Phase 2 CAPSTONE** — the tracker that's evolved through L08/L09/L10/L11/L13 gets reorganised into a polished v2 with the idiomatic `cmd/` + `internal/` layout. Plus four new testing patterns: golden files, benchmarks, fuzz preview, and the `t.Helper`/`t.TempDir` toolkit.

Most code in this lesson is mechanical CARRY-FORWARD from prior lessons (with import-path rewrites only). The substantive new code is:

- `internal/testutil/golden.go` — provided AssertGolden helper (~25 lines)
- `cmd/expenses/main_test.go` + `testdata/summary.golden` — the golden file test
- `internal/summary/summary_bench_test.go` — the BenchmarkTotalsByCategory benchmark
- `internal/csvimport/csvimport_fuzz_test.go` — the FuzzParseLine preview

## The v2 layout

```
lessons/15-structure/exercises/
├── warmup/
│   ├── main.go                          ← demonstrates internal/ access
│   └── internal/util/format.go          ← the imported package
├── cmd/
│   ├── expenses/
│   │   ├── main.go                      ← the CLI (now uses internal/summary)
│   │   ├── main_test.go                 ← golden file test
│   │   └── testdata/summary.golden      ← expected output
│   └── expenses-import/
│       ├── main.go                      ← CSV importer (carry-forward from L13)
│       └── main_test.go                 ← os/exec integration tests
└── internal/
    ├── expense/                          ← Expense + methods (from L09/L11)
    ├── store/                            ← Store + JSONStore + MemoryStore (from L13)
    ├── csvimport/                        ← Parse(io.Reader) (from L13)
    │   └── csvimport_fuzz_test.go        ← NEW: FuzzParseLine preview
    ├── summary/                          ← TotalsByCategory + BiggestCategory + Bar (from L08)
    │   └── summary_bench_test.go         ← NEW: BenchmarkTotalsByCategory
    └── testutil/golden.go                ← NEW: AssertGolden helper (provided)
```

Nothing outside `lessons/15-structure/` can import any package under `internal/`. The compiler enforces it; no annotation needed.

---

## Concept 1 — `cmd/` + `internal/`

Two directory conventions, both compiler-supported:

- **`cmd/<binary>/main.go`** — one `package main` per subdirectory. You met this in L07 and used it through L08-L14.
- **`internal/<pkg>/`** — packages here can ONLY be imported by code in the parent of `internal/` or its descendants. A sibling module trying to import an `internal/` package fails at compile time:

```
use of internal package github.com/your/module/internal/store not allowed
```

No annotation, no opt-out. The directory NAME is the gatekeeper.

For a repo at `lessons/15-structure/`, code under `lessons/15-structure/internal/store` is visible to anything inside `lessons/15-structure/`. Anything outside that subtree — another lesson, a downstream module pulling us in via `go get`, the build tool — can't import it. You can refactor `internal/store` freely; the compiler proves there are no consumers you don't know about.

### Common mistake

Putting `cmd/` inside `internal/` (or vice versa). `cmd/` signals "public-facing binary"; `internal/` signals "private package". Putting one inside the other contradicts both. Keep them as siblings at the project root.

---

## Concept 2 — Test helpers

Three `testing.T` methods you'll reach for often:

- **`t.Run(name, fn)`** — runs `fn` as a sub-test. Standard for table tests; the sub-test gets its own `t`, fails independently, shows up in `go test -v` output as `TestX/name`.
- **`t.Helper()`** — marks a function as a test helper. Failures inside the helper report the CALLER's line, not the helper's. Without this, every assertion failure points to the same line inside your helper; useless.
- **`t.TempDir()`** — returns a fresh empty directory for this test. Auto-deleted on test end (pass, fail, OR panic). No `defer os.RemoveAll` needed.

```go
func assertEqual(t *testing.T, got, want int) {
    t.Helper()    // ← without this, failures report assertEqual's line
    if got != want {
        t.Errorf("got %d, want %d", got, want)
    }
}

func TestX(t *testing.T) {
    dir := t.TempDir()             // ← auto-cleaned
    // ... write a fixture, test, return; no cleanup
}
```

### Common mistake

Forgetting `t.Helper()` in a helper. Every test failure points to line 42 of `helpers.go` instead of line 17 of `my_test.go`. The fix is one line. Bake it into your fingers: every helper that calls `t.Errorf` / `t.Fatal` starts with `t.Helper()`.

---

## Concept 3 — Golden file tests

For commands producing structured output (multi-line reports, JSON snippets, HTML), inline `want := "<huge string>"` is unmaintainable. Golden files solve it:

1. Expected output lives in `testdata/<name>.golden` (Go's build system reserves `testdata/`).
2. The test reads `got`, compares to the file, fails with diff on mismatch.
3. A `-update` flag regenerates the file when output legitimately changes.

The lesson's `internal/testutil/golden.go` provides `AssertGolden(t, got, path)`:

```go
output := runMyCommand(...)
testutil.AssertGolden(t, output, "testdata/expected.golden")
```

Running:

```bash
go test ./cmd/expenses/                    # compares
go test -update ./cmd/expenses/            # regenerates after intentional changes
```

The golden file is SOURCE CODE — commit it, review changes in PRs.

### Common mistake

Forgetting to run `-update` after intentional output changes. Test fails with a wall of diff; the reflex is to revert the code change or edit the golden by hand. Both wrong. Once you've verified the new output is correct, just `go test -update ./...` and commit the regenerated files.

The corollary mistake: not committing the golden file at all. An empty `testdata/` is the same as no golden file — the first `go test` then fails with "no such file."

---

## Concept 4 — Benchmarks

A benchmark looks like a test but takes `*testing.B`:

```go
func BenchmarkX(b *testing.B) {
    data := makeFixture()                 // setup (once)

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        _ = DoSomething(data)
    }
}
```

Go runs the function with increasing `b.N` until it's confident in the measurement, then reports `ns/op` (nanoseconds per iteration) plus `B/op` and `allocs/op` if you called `b.ReportAllocs()`.

Two critical bits:

- **`b.ResetTimer()`** after setup — otherwise the fixture-building time gets counted against your benchmark.
- **`b.ReportAllocs()`** — opt-in allocation reporting. Catches accidental boxing and intermediate slices.

Run with:

```bash
go test -bench=BenchmarkX -benchmem ./pkg/
```

The L15 benchmark generates 10,000 random expenses (fixed seed) and benchmarks `summary.TotalsByCategory`. Output looks like:

```
BenchmarkTotalsByCategory-10    5000    212875 ns/op    8192 B/op    1 allocs/op
```

### Common mistake

Doing setup INSIDE the `b.N` loop. You're now benchmarking `setup + DoSomething`. Hoist setup out and call `b.ResetTimer()` to discard the setup time.

Second mistake: comparing benchmark numbers across different machines. The numbers are RELATIVE — useful for "is my new version faster than my old version?" on the same machine. Cross-machine comparisons are noise.

---

## Concept 5 — Fuzz tests (preview)

Go 1.18 added fuzzing as first-class. A fuzz function takes `*testing.F`, registers seeds, and calls `f.Fuzz` with a function the framework invokes on mutated inputs:

```go
func FuzzX(f *testing.F) {
    f.Add("valid seed 1")
    f.Add("")

    f.Fuzz(func(t *testing.T, input string) {
        _, _ = ParseSomething(input)   // any panic fails the test
    })
}
```

Two ways to run:

```bash
go test ./pkg/                              # runs the seeds only (fast)
go test -fuzz=FuzzX -fuzztime=30s ./pkg/    # explores for 30 seconds
```

Without `-fuzz=`, the fuzz function acts like an awkward table test (just the seeds). The VALUE is in the exploration phase — the framework mutates the seeds and finds inputs you didn't think of. Crashes get saved to `testdata/fuzz/FuzzX/<hash>` as automatic regression tests.

L15 includes one fuzz function — `FuzzParseLine` on the CSV parser's `parseLine` helper. It seeds with three inputs (valid, empty, malformed) and asserts no panics. Phase 3 lesson 22 dives deeper.

### Common mistake

Writing the fuzz function and never running with `-fuzz=`. Without that flag, only the seeds run — a few hand-written inputs, no exploration. Add `-fuzz=FuzzX -fuzztime=30s` to your release process to actually catch bugs.

---

## Exercise: warm-up — internal/util demo

`warmup/internal/util/format.go` exposes `Money(amount float64) string` returning `"€<2-decimal>"`. `warmup/main.go` imports it. The lesson is the COMPILATION SUCCESS: same module, internal/ access from a sibling is allowed.

**Time:** 5 minutes.

## Exercise: main — v2 reorg

Three parts:

1. **Move existing code into `internal/<pkg>/`** — expense, store, csvimport, summary. Mostly mechanical: copy from L11/L13/L08, rewrite import paths.
2. **Convert `cmd/expenses summary` to a golden file test** using `testutil.AssertGolden`. Commit `testdata/summary.golden`.
3. **Add `BenchmarkTotalsByCategory`** in `internal/summary/summary_bench_test.go` and **`FuzzParseLine`** in `internal/csvimport/csvimport_fuzz_test.go`.

**Time:** 60-90 minutes. Mostly mechanical; the new testing patterns are the substantive parts.

---

## Daily habits

After every change:

```bash
gofmt -w ./...
go vet ./...
go test ./...
```

For the new patterns:

```bash
go test -update ./cmd/expenses/                          # regenerate golden after changes
go test -bench=. -benchmem ./internal/summary/           # the benchmark
go test -fuzz=FuzzParseLine -fuzztime=5s ./internal/csvimport/   # fuzz exploration
```

## How to run

```bash
# Skeleton (passes vacuously until you implement)
cd lessons/15-structure/exercises
go test ./...

# Reference solution
go test -v ./lessons/15-structure/solutions/...

# The binaries
go run ./lessons/15-structure/solutions/cmd/expenses -file=/tmp/t.json summary
echo "2026-05-21,4.50,coffee" | go run ./lessons/15-structure/solutions/cmd/expenses-import -file=/tmp/i.json

# The warmup
go run ./lessons/15-structure/solutions/warmup
```

## Going further

### Read

- **Effective Go — "Package names"** (still useful in 2026): <https://go.dev/doc/effective_go#names>
- **`testing` package docs** — Benchmark + Fuzz sections: <https://pkg.go.dev/testing>
- **Russ Cox — "internal" packages release notes (Go 1.4)**: <https://go.dev/doc/go1.4#internalpackages>

### Try

- **Add a golden file test for `cmd/expenses list`.** Pattern is the same as `summary`: seed a MemoryStore, call cmdList capturing into bytes.Buffer, AssertGolden. Decide whether listing with zero entries should produce an empty golden (just newlines) or skip the test.
- **Benchmark `csvimport.Parse` on a 100k-line input.** Generate the input with `strings.Repeat`. Compare ns/op with various line counts to see if Parse scales linearly.
- **Run the fuzzer for an hour.** `go test -fuzz=FuzzParseLine -fuzztime=1h ./internal/csvimport/`. If you find a crash, file an issue describing the offending input. (Likely outcome: no crash — parseLine has good defensive checks.)
- **Move the warmup's `main.go` up one directory** (to `exercises/`) and re-run `go build ./...`. Observe the compile error. Then move it back. This is the demo Phase 2 spec described.

---

> **Phase 2 complete.** Next stop: Phase 3 — concurrency. Goroutines, channels, the context package, networking. The single hardest part of Go for newcomers, and where you'll learn what makes Go's runtime distinctive.
