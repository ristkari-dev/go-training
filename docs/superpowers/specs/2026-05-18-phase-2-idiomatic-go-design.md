# Phase 2 — Idiomatic Go (lessons 9-15) Design

**Status:** Approved (brainstorming complete, awaiting per-lesson implementation plans)
**Date:** 2026-05-18
**Owner:** Aki Ristkari
**Master design:** [`2026-05-05-go-course-design.md`](2026-05-05-go-course-design.md)
**Phase 1 design (predecessor):** [`2026-05-07-plan-c-phase-1-foundations-design.md`](2026-05-07-plan-c-phase-1-foundations-design.md)

## Summary

Phase 2 takes students from "I can write small Go programs" (where Phase 1 ended) to "I can read and write idiomatic Go" — pointers, interfaces, generic-with-restraint, mature error design, encoding & I/O, std-lib literacy, and project structure. Seven lessons (9-15). The expense tracker from Phase 1 evolves through most lessons; lesson 15 reorganises it into a polished v2 with `cmd/` + `internal/` + golden file tests. Two lessons (L12 generics, L14 std-lib) use standalone examples because their concepts don't fit the tracker naturally.

## Audience and pacing

- Phase 2 students: software engineering students who have completed all 8 Phase 1 lessons. They can read and write basic Go; structs/methods/packages/errors/tests are second nature.
- Each lesson: ~90 minutes live + ~90 minutes self-study (slightly bumped from Phase 1's 60-90 self-study because concepts are richer and exercises are more open-ended).
- 7 lessons total. Roughly half a semester of weekly sessions.

## Core decisions (locked in during brainstorming)

| Decision | Choice |
|---|---|
| Phase 2 capstone | Evolve the expense tracker (lesson 15 polish + reorganisation) |
| Cross-lesson continuity | Most lessons advance the tracker; L12 (generics) and L14 (std-lib) use standalone examples |
| Test scaffolding | Skeleton tests in both warm-up and main exercises (same shape as Phase 1's L05-L08 ending) |
| Layout default | Subpackages + `cmd/<binaryname>/main.go` is the default from L09 onward (mirrors Phase 1's L07-L08 mature shape) |
| Slide concept count | Default 4 per lesson; some lessons may need 3 or 5 |
| Per-lesson example | Each lesson's main exercise touches the tracker (where natural) |
| New imports per lesson | One or two on average; L13 finally opens up `encoding/json` (used opaquely since L08) |
| `internal/` introduced | Lesson 15 (project structure) — first time students see `internal/` formally |
| Pointer receivers default | From L09 onward, methods that mutate use pointer receivers; methods that read use value receivers (formalising the lesson 06 "light touch") |

## Curriculum

Each lesson below has the same shape: concepts, per-lesson example (tracker-evolution or standalone), warm-up exercise, main exercise.

### Lesson 9 — Pointers

- **Concepts:** value vs reference semantics, `&` (address-of) and `*` (dereference), pointer receivers (deep dive after L06's light touch), the "value receiver copies; pointer receiver shares" rule, escape-analysis intuition (no internals), `nil` pointer dereferences as the most common runtime panic.
- **Per-lesson example:** Convert `Expense` to support mutation methods with pointer receivers. Add `(*Expense) ApplyDiscount(rate float64)` (mutates in place) alongside the existing value-receiver `Format()` / `IsHigh()`. Demonstrates "be consistent across methods of a type" by either making all methods pointer receivers OR documenting why mixing is OK here.
- **Warm-up:** A small `Counter` type with `(c *Counter) Inc()` and `(c Counter) Value() int` — the smallest possible "pointer receiver mutates; value receiver doesn't" demo.
- **Main:** Add `ApplyDiscount(rate float64)` as a pointer receiver on `Expense`; add `(*Expense) Bump(amount float64)` as another mutating method. Test that both stick on the caller's value.
- **New imports introduced:** none.
- **Tracker contribution:** mutation methods on Expense.

### Lesson 10 — Interfaces

- **Concepts:** implicit satisfaction (no `implements` keyword), small interfaces (`io.Reader`, `io.Writer`, `error`), `any` (formerly `interface{}`), type assertions (`v, ok := x.(T)`), the "accept interfaces, return structs" rule, the empty interface as a code-smell signal.
- **Per-lesson example:** Introduce a `Store` interface in the tracker — abstracts persistence. Two implementations: the existing `JSONStore` (from lesson 08's storage package) and a new `MemoryStore` (in-memory map; useful for tests). The CLI can be configured to use either via a `-store=mem|json` flag.
- **Warm-up:** Define a `Greeter` interface with one method `Greet(name string) string`; two implementations (an English greeter and a Finnish greeter). Show how `Greet` callers don't care which.
- **Main:** Define `Store` interface in `internal/store` (forward reference — `internal/` arrives formally L15; here we just put it in a subpackage). Two implementations. Refactor cmd/expenses to depend on the interface, not the concrete type. Tests use `MemoryStore` for speed.
- **New imports introduced:** none (interfaces are a language feature).
- **Tracker contribution:** `Store` interface + `MemoryStore` impl.

### Lesson 11 — Errors

- **Concepts:** sentinel errors (`var ErrNotFound = errors.New(...)`), wrapping with `fmt.Errorf("...: %w", err)`, `errors.Is` (check sentinel identity) vs `errors.Wrap` (older idiom — historical only), `errors.As` (extract typed error), custom error types via struct + `Error() string`, error-design philosophy ("does the caller need to discriminate? add a sentinel or type. If not, just wrap.").
- **Per-lesson example:** Storage errors get rich context. `storage.ErrNotFound` (sentinel). Wrapped errors for malformed JSON include the file path and line number. Custom `ParseError` type with `Line` and `Cause` fields.
- **Warm-up:** A `parseAge(s string) (int, error)` function that wraps the underlying `strconv.Atoi` error with the offending input string.
- **Main:** Add sentinel + custom error types to the tracker's storage package. Update cmd/expenses to use `errors.Is` for the "file not found" case (currently silenced by returning empty slice) and present a better error to the user for malformed JSON.
- **New imports introduced:** none (already use `errors`); deeper use of `fmt.Errorf` with `%w`.
- **Tracker contribution:** richer error returns + matching error handling at the call site.

### Lesson 12 — Generics (standalone)

- **Concepts:** type parameters (`func F[T any](x T) T`), constraints (`any`, `comparable`, `constraints.Ordered` from `golang.org/x/exp/constraints` — note: forward-reference to Go 1.21's `cmp.Ordered`, which we'll use), when NOT to use generics ("if a function is called once with one type, don't bother"), the cost (cognitive complexity; longer compile times in extreme cases).
- **Per-lesson example:** A small `slices`-style helper library. `Filter[T any](xs []T, pred func(T) bool) []T`, `Map[T, U any](xs []T, f func(T) U) []U`, `Max[T cmp.Ordered](xs []T) (T, error)`. Standalone — not part of the tracker (forcing generics into the tracker is the "premature abstraction" anti-pattern Go itself warns against).
- **Warm-up:** Implement `Max[T cmp.Ordered](xs []T) (T, error)` — generic version of lesson 04's `WarmupMinMax`. Mirrors familiar shape.
- **Main:** Implement `Filter` + `Map`. Use them on small `[]int` / `[]string` slices in the tests. One example briefly shows them on `[]Expense` (loose tracker connection: filter Expenses where `e.IsHigh()`) but that's a one-liner in the test, not a refactor.
- **New imports introduced:** `cmp` (for `cmp.Ordered`).
- **Tracker contribution:** none structurally; one one-liner usage in the slides as motivation.

### Lesson 13 — Encoding & I/O

- **Concepts:** `encoding/json` — `Marshal`/`Unmarshal` (whole-blob), `Encoder`/`Decoder` (streaming), struct tags revisited; `os.Open` / `os.Create` returning `*os.File` (which implements `io.Reader` + `io.Writer`); the `io.Reader` / `io.Writer` interfaces formally; `bufio.Scanner` for line-based reading; `bufio.Writer` for buffered output; streaming patterns (read line by line, transform, write line by line).
- **Per-lesson example:** Open up the storage package's internals — students now understand and modify what was a black box in L08. Add a `cmd/expenses-import/main.go` binary that reads CSV from stdin and uses `storage.SaveExpenses` to persist the result.
- **Warm-up:** Implement `CountLines(r io.Reader) (int, error)` using `bufio.Scanner`. Demonstrates reading from any `io.Reader`.
- **Main:** Write a CSV-to-expenses importer. Reads `date,amount,category` rows; parses into `[]expense.Expense`; calls `storage.SaveExpenses`. Skeleton tests use `strings.NewReader` (an `io.Reader` over a string literal — students learn the "test with any Reader" trick).
- **New imports introduced:** `bufio`, `io`. `encoding/json` opened up.
- **Tracker contribution:** CSV importer binary + storage internals demystified.

### Lesson 14 — Time, strings, bytes, regex (standalone)

- **Concepts:** `time.Time`, `time.Duration`, time arithmetic, time formatting and parsing with the reference time `2006-01-02 15:04:05`; the `strings` package (`Split`, `Join`, `Contains`, `TrimSpace`, `ToLower`); the `bytes` package as the byte-equivalent of `strings`; `regexp` basics (`MustCompile`, `FindAll`, capture groups), regex as a tool of last resort.
- **Per-lesson example:** A log-line parser. Input: `2026-05-18T14:30:00 INFO connection accepted from 192.168.1.5`. Parses out timestamp + level + message via regex; counts entries per level. Standalone — the tracker doesn't naturally need regex/time/string manipulation at this depth.
- **Warm-up:** Parse and format dates. Implement `FormatDate(t time.Time) string` returning `"YYYY-MM-DD HH:MM:SS"` and `ParseDate(s string) (time.Time, error)` doing the inverse.
- **Main:** A `LogParser` that takes an `io.Reader` of log lines, returns `map[string]int` of level counts (`INFO`, `WARN`, `ERROR`). Uses regex + bufio.Scanner. Skeleton tests use `strings.NewReader` again.
- **New imports introduced:** `time`, `strings`, `bytes`, `regexp`.
- **Tracker contribution:** none structurally. (The lesson exists in service of std-lib literacy; the tracker's date field stays `string` for backwards compat — converting to `time.Time` would be an L15 polish.)

### Lesson 15 — Idiomatic project structure & testing patterns (CAPSTONE)

- **Concepts:** the `cmd/` convention (already met in L07-L08), the `internal/` convention (formal introduction — packages under `internal/` are visible only to siblings + parent, enforced by the compiler), subtests with `t.Run` for non-table cases, `t.Helper()` for cleaner failure messages, `t.TempDir()` (already met in L08 integration tests), golden file tests (compare actual output against a `.golden` file; `-update` flag regenerates), benchmark functions (`func BenchmarkX(b *testing.B)`), fuzz tests (`func FuzzX(f *testing.F)` — Phase 3 lesson 22 dives deep; L15 only mentions the syntax).
- **Per-lesson example:** Reorganise the tracker that's evolved through L09/L10/L11/L13 into a polished v2:
  - `cmd/expenses/` (existing — main CLI)
  - `cmd/expenses-import/` (CSV importer from L13)
  - `internal/expense/` (Expense + value methods + pointer-receiver methods from L09)
  - `internal/store/` (Store interface from L10 + JSONStore + MemoryStore)
  - `internal/summary/` (TotalsByCategory + BiggestCategory + Bar)
  Add golden file tests for the summary output. Add a benchmark for `summary.TotalsByCategory` on a 10,000-entry slice.
- **Warm-up:** Move one helper into `internal/util/` — show the compile error you get when you try to import it from outside the module (well, demonstrate the access rule).
- **Main:** Full reorg per the bullet list above. Migrate the L08 integration tests to use golden files. Add one benchmark (lesson previews L22's deeper coverage).
- **New imports introduced:** none net-new (testing.B / testing.F are part of `testing`).
- **Tracker contribution:** the polished v2 — the capstone.

## Capstone shape (lesson 15)

### What the v2 tracker looks like

Directory layout after L15:

```
lessons/15-structure/exercises/
├── cmd/
│   ├── expenses/
│   │   ├── main.go              (the CLI)
│   │   └── main_test.go         (integration tests via os/exec + golden files)
│   │   └── testdata/
│   │       └── summary.golden   (the golden expected output)
│   └── expenses-import/
│       └── main.go              (the CSV importer from L13)
└── internal/
    ├── expense/
    │   ├── expense.go           (struct + value methods + pointer-receiver methods from L09)
    │   └── expense_test.go
    ├── store/
    │   ├── store.go             (Store interface from L10 + JSONStore + MemoryStore)
    │   └── store_test.go        (uses MemoryStore for fast tests)
    └── summary/
        ├── summary.go           (TotalsByCategory + BiggestCategory + Bar)
        ├── summary_test.go
        └── summary_bench_test.go (benchmark)
```

`solutions/` mirrors the same layout exactly.

### What the student writes

1. **Move existing code** from L13/L11/L10/L09 into `internal/<subpkg>/`. Mostly mechanical.
2. **Convert L08 integration tests** to use golden files. `testdata/summary.golden` holds the expected stdout; the test calls the binary, compares stdout to the golden file, and runs `t.Errorf` with a diff on mismatch. A `-update` flag (`go test -update`) regenerates the golden file when intentional output changes happen.
3. **Add one benchmark** in `summary_bench_test.go`. Generates 10,000 random Expenses; benchmarks `TotalsByCategory`. Shows ns/op + B/op. Output is informational — no assertions.

### What's provided

- **The L08-L14 work students did in previous lessons** is what they reorganise. L15 doesn't provide new starter code; it provides a `LAYOUT.md` instructing students how to move the existing files into the v2 shape.
- **A golden-files helper** — a tiny `internal/testutil/golden.go` package providing `AssertGolden(t, got, path)`. Implementation is provided (~20 lines); students just use it.

### Capstone non-scope

- Sorting, edit/delete (still — same as L08's non-scope).
- Multi-user / multi-file expense tracking.
- Web UI (Phase 4).
- Concurrency (Phase 3).
- Proper config file (just CLI flags + `-store=mem|json` from L10).

These appear in lesson 15's "Going further" section as Phase 3-4 suggestions.

## Lesson anatomy

### Folder layout default (lessons 9-14)

Subpackages + `cmd/` is the default (per the L07-L08 mature shape):

```
lessons/NN-name/
├── README.md
├── slides/
│   ├── index.html
│   ├── slides.md
│   └── assets/
├── exercises/
│   ├── cmd/                   (only when the lesson has a binary)
│   │   └── <binaryname>/
│   │       └── main.go
│   ├── <library-subpackage>/
│   │   ├── <file>.go
│   │   └── <file>_test.go     (SKELETON tests — students fill in)
│   └── (warmup follows the same shape)
└── solutions/
    └── (mirrored — full reference implementations + tests)
```

Lessons without a binary (e.g., L11 errors might just update existing packages) may omit the `cmd/` subtree.

### Folder layout for lesson 15 (capstone)

Adds `internal/` and `testdata/`. See "Capstone shape" above.

### `README.md` structure (every lesson)

Same as Phase 1's:

1. **Learning goals** — 3-5 bullets.
2. **Prerequisites** — links to earlier lessons.
3. **Concepts** — full prose mirroring the deck.
4. **Exercise: warm-up** — what to build, expected `go test` output.
5. **Exercise: main** — same shape, larger.
6. **How to run** — exact `go test` commands.
7. **Going further** — 2-3 links + 1-2 stretch problems.

### Slide structure (continues Phase 1's heavy-explanatory pattern)

Same per-concept rhythm (Motivation / The basics / A worked example / Common mistake / Recap). Default 4 concepts per lesson; some may justify 3 (L11 errors, L15 structure) or 5 (L10 interfaces, L13 encoding/IO).

## Platform updates (Phase 2 tasks 1-2, before lesson plans are authored)

### Update the lesson scaffolder

`tools/new-lesson/template/` currently produces a flat layout (`exercises/{warmup,main}.go`). Phase 2's default is subpackages + `cmd/`. Two options:

- **A. Keep the flat scaffolder.** Each Phase 2 lesson's plan deletes the flat files + creates the subfolder tree (same approach as Plan J's Task 1 — restructure). 7 lessons × that same deletion dance.
- **B. Add a `--layout=subpackages` flag to the scaffolder.** Generates the subpackage tree directly. Saves the per-lesson restructure step.

Recommendation: **A** (status quo). The restructure step has worked twice (Plans J, K); adding a layout flag is more code than the savings justify. Defer the flag to Phase 3 if every lesson there also uses subpackages.

### `internal/testutil/golden.go`

The L15 capstone needs a golden-files helper. Ship it in the lesson 15 starter (`exercises/internal/testutil/golden.go` and `solutions/internal/testutil/golden.go`). Identical content in both — students don't write it.

Implementation sketch (~20 lines): `AssertGolden(t, got string, path string)` — reads `path`, compares to `got`, on mismatch either prints a diff (if `-update` is NOT set) or rewrites the file (if `-update` IS set).

### What's NOT changing

- `Makefile` — `go test ./...` already covers nested subpackages; no new targets.
- `tools/build-index` — walks `lessons/*/slides/`; subpackages are invisible to it.
- `tools/slides-dev` — unchanged.
- The deploy pipeline.
- `go.mod` — Phase 2 stays stdlib-only (no third-party deps); `cmp` is stdlib as of Go 1.21+.

## Cross-cutting threads

### Testing thread

- **Lessons 09-14:** Skeleton tests in both warm-up and main, continuing Phase 1 L05-L08's pattern. Students fill in cases + assertion bodies.
- **Lesson 15:** Testing patterns are formally introduced:
  - Subtests with `t.Run` for non-table cases.
  - `t.Helper()` for cleaner failure traces in helper functions.
  - `t.TempDir()` — already met in L08; revisited.
  - Golden file tests via the provided `testutil.AssertGolden`.
  - Benchmarks (`testing.B`) — basic shape only; L22 covers deeply.
  - Fuzz tests (`testing.F`) — one-line forward reference; L22 covers.

### Tooling thread

- `gofmt` + `go vet` — routine from Phase 1 onward.
- `go doc` / `go list` / `go mod tidy` — established in L07.
- `go test -bench` — introduced in L15.
- `go test -fuzz` — mentioned in L15 as a forward reference; not used in exercises.

### Error handling thread

- **L09-L10:** Errors continue from Phase 1's `if err != nil { return err }` pattern. No deeper machinery yet.
- **L11:** Formal treatment — sentinel, wrapping, `errors.Is`/`As`, custom error types.
- **L13-L14:** Students use wrapped errors when their I/O code can fail in multiple ways (file not found vs malformed input vs validation).
- **L15:** Errors are part of the "polished v2" — every error returned from `internal/` packages should either be a sentinel (so callers can `errors.Is`), a wrapped error with context, or a custom type with structured fields.

### Imports thread

| Lesson | New imports introduced |
|---|---|
| L09 Pointers | none |
| L10 Interfaces | none |
| L11 Errors | none (deeper use of `errors` + `fmt.Errorf` with `%w`) |
| L12 Generics | `cmp` (for `cmp.Ordered`) |
| L13 Encoding & I/O | `bufio`, `io`, `encoding/json` (opened up) |
| L14 Std-lib | `time`, `strings`, `bytes`, `regexp` |
| L15 Structure | none (testing.B / testing.F are in `testing`) |

Phase 2 introduces 7-8 new stdlib packages total. Still no third-party deps.

### "Going further" thread

Per-lesson README has the same `## Going further` section as Phase 1: 1-3 read links + 1-2 stretch problems with no reference solutions.

### Identifier-naming thread

- Same rules as Phase 1 — full-word names, idiomatic short names where Go itself uses them, `Test<Symbol>` for tests, kebab-case sub-test names.
- New: error sentinel naming convention — `ErrXxx` for exported sentinels, `errXxx` for unexported. Lesson 11 introduces this.

## Out of scope (Phase 2 boundaries)

### Deferred to Phase 3 (concurrency)

- Goroutines, channels, `select`, `sync`, `context`. Phase 2 does the synchronous equivalent everywhere.
- Networking (`net`, TCP/UDP), syscalls beyond `os.Open`/`os.ReadFile`.
- Profiling, benchmarking deep dive (L22). L15 mentions benchmarks only.
- Fuzz testing deep dive (L22). L15 mentions only.

### Deferred to Phase 4 (production)

- HTTP servers / clients. gRPC. Config / secrets / signal handling. Observability. Containerisation.

### Hard non-goals (not arriving later either, in Phase 2 framing)

- No third-party dependencies — `go.mod`'s `require` block stays empty through all of Phase 2.
- No web framework. No databases. No GUI.
- No discussion of "clean architecture" / "hexagonal architecture" / DDD as named patterns. The polished v2 in L15 demonstrates *some* of those ideas (separation of cmd/ vs internal/, dependency inversion via the Store interface) but doesn't lecture about them as named methodologies.
- No reflection (`reflect` package). Phase 3+ if ever — most Go code never needs it.
- No `unsafe`. Same reason.
- No `cgo`. Same reason.

## Open items deferred to per-lesson implementation planning

- Exact wording for the `cmp.Ordered` forward reference in L12 (do we mention `golang.org/x/exp/constraints` historically, or just use `cmp.Ordered` directly and not look back?).
- The `Store` interface's exact method set (L10). Plausible options: `Load() ([]Expense, error)` + `Save([]Expense) error`, OR per-item methods `Add(Expense) error` / `List() ([]Expense, error)` / `Get(id string) (Expense, error)`. Plan L10 (lesson 10) decides.
- L11 custom error type's exact field list (just `Cause error` + `Line int`, or also `Path string`, or also `Offset int`?). Plan L11 decides.
- L13's CSV format details — header row required? Tab-separated also accepted? Quoting rules? Plan L13 decides.
- L14 log-line regex's exact format — Plan L14 decides.
- L15's golden-file diff format — line-by-line, unified diff, or just "got vs want" blocks? Plan L15 decides.
- Whether to retire the `WarmupX` naming prefix once subpackages are universal (subpackage namespacing already solves the collision risk the prefix was for). Plan L09 likely retires it; flagged here for the first L09 author to consider.
