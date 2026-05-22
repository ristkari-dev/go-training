# Phase 3 — Concurrency & Systems (lessons 16-22) Design

**Status:** Approved (brainstorming complete, awaiting per-lesson implementation plans)
**Date:** 2026-05-22
**Owner:** Aki Ristkari
**Master design:** [`2026-05-05-go-course-design.md`](2026-05-05-go-course-design.md)
**Phase 2 design (predecessor):** [`2026-05-18-phase-2-idiomatic-go-design.md`](2026-05-18-phase-2-idiomatic-go-design.md)

## Summary

Phase 3 takes students from "I can read and write idiomatic Go" (end of Phase 2) to "I can read, write, and reason about concurrent Go." Seven lessons (16-22) covering goroutines + channels, `select`, `sync` + the memory model, `context`, concurrency patterns, networking, and profiling/benchmarking/fuzzing.

The **log aggregator** (extends L14's `logparse` package) is the running example for L16-L20: a CLI that walks a directory of log files in parallel, accumulates level counts, exposes per-file timeouts, shared-state coordination, context-based cancellation, and a worker pool. L21 (networking) and L22 (profiling/fuzz) are standalone — networking uses a small TCP echo server/client; profiling/fuzz wraps Phase 3 by measuring the aggregator built across L16-L20.

The expense tracker from Phase 1/2 stays where L15 left it. The log aggregator is a NEW running example because concurrency doesn't fit the tracker naturally — a single-user file-backed CLI has nothing to parallelize meaningfully.

## Audience and pacing

- Phase 3 students: software engineering students who have completed all 15 Phase 1+2 lessons. They can read and write idiomatic Go: pointers, interfaces, errors, generics, encoding/IO, time/strings/regex, project structure, golden file tests, benchmarks, and basic fuzz syntax.
- Each lesson: ~90 minutes live + ~90 minutes self-study (same as Phase 2).
- Concurrency is the hardest part of Go for newcomers. Expect L18 (sync & memory model) and L19 (context) to take longer than 90 min for many students.
- 7 lessons total. Roughly half a semester of weekly sessions.

## Core decisions (locked in during brainstorming)

| Decision | Choice |
|---|---|
| Running example | Log aggregator (extends L14 `logparse`); L21 standalone; L22 profiles the aggregator built across L16-L20 |
| `errgroup` policy | Lesson-local `errgroupx` (~30 lines) — preserves stdlib-only invariant |
| Phase 3 capstone | None. 7 lessons (16-22). L22 = profile/benchmark/fuzz the aggregator |
| Race detector | New `make test-race` target; Phase 3 lessons feature it in daily habits |
| Tracker continuity | Tracker stays where L15 left it; log aggregator is the new running example |
| Test scaffolding | Skeleton tests in both warmup and main (continued from Phase 2 default) |
| Slide concept rhythm | 4-5 concepts per lesson (continued from Phase 2) |
| Std-lib only | Zero third-party deps preserved — no `golang.org/x/sync`, no anything |

### Why each decision

**Running example: log aggregator (vs URL fetcher vs all-standalone)**
The aggregator extends L14's `logparse` package — students see existing code grow rather than start from zero each lesson. It maps cleanly to the concurrency progression: L16 spawns one goroutine per file, L17 adds timeouts, L18 introduces a mutex on shared state, L19 wires context cancellation, L20 turns it into a worker pool. URL fetcher would be more "tutorial-recognisable" but requires more test fixturing (httptest servers, mocks). All-standalone would lose the narrative arc that worked in Phase 2.

**`errgroup` policy: lesson-local `errgroupx`**
`golang.org/x/sync/errgroup` is the de facto standard but lives outside the stdlib. Phase 1 and Phase 2 made "Go 1.23 stdlib only" a core promise; breaking it in Phase 3 for one package opens death by a thousand exceptions (the same logic could apply to `golang.org/x/oauth2`, `golang.org/x/exp/...`, etc., in Phase 4). Implementing a lesson-local `internal/errgroupx/` (~30 lines wrapping `sync.WaitGroup` + `context.CancelFunc` + a once-protected first-error) is pedagogically valuable — students see HOW the canonical "first error wins, cancel the rest" pattern is built. The README name-drops `golang.org/x/sync/errgroup` as "the package you'll actually use in production; same shape."

**No Phase 3 capstone**
Phase 1 ended with L08 capstone (polished tracker CLI), Phase 2 with L15 capstone (cmd/+internal/ reorganization). Phase 3's running example IS the integrative thread — by L20 the aggregator is already a polished worker pool with context cancellation and first-error-cancels-rest. L22 (profile/benchmark/fuzz the aggregator) gives the same integrative moment as a capstone would, without a separate framing. Phase 4 starts at L23 (HTTP servers) per original course design; adding a Phase 3 capstone would break the 4-phase rhythm (7 + 7 + 7 + 7 = 28; with capstone, 7 + 7 + 8 + 6 = 28 but Phase 4 would shrink to 6 lessons).

**`make test-race` as a new target (vs `-race` in `make test` globally)**
Phase 1-2 had zero goroutines; running them under `-race` has no pedagogical payoff, only 5-20× overhead. `make test` should keep meaning "run the standard test suite quickly." Phase 3 introduces `make test-race` as the explicit concurrency-verification target — students learn the convention "race detector is its own tool, used deliberately" rather than discovering their `make test` has been running `-race` invisibly. Each phase's CI is then right-sized: Phase 1-2 CI runs `make test`; Phase 3+ CI runs both.

**Std-lib only**
Phase 1 + Phase 2 ran for 15 lessons with zero third-party imports. This is a Go-community-genuine choice (the stdlib is large enough for most engineering work) and a teaching choice (no dependency drama, no version pinning, no supply-chain considerations). Preserving the invariant through Phase 3 reinforces the lesson.

## Per-lesson breakdown

### Lesson 16 — Goroutines & channels (basics)

- **Concepts:** the `go` keyword and the goroutine model (no return values; communicate via channels); unbuffered channels (synchronous handoff); buffered channels (capacity for in-flight messages); channel direction syntax (`chan<- T`, `<-chan T`); `range` over channel; `close` and the zero-value-on-closed semantics; common-mistake: leaked goroutines (sender blocked forever because nobody reads).
- **Per-lesson example:** Walk a directory of log files in parallel — spawn one goroutine per file, each parses with `logparse.Parse`, sends a per-file `levelCounts map[string]int` through a results channel; main collects N results and reduces.
- **Warm-up:** Implement `Echo(in <-chan string) <-chan string` — read everything from `in`, send each item through to a new output channel, close output when input closes.
- **Main:** Add `internal/aggregator/Walk(dir string) (map[string]int, error)` that uses `os.ReadDir` to list files, spawns one goroutine per file, collects results via channel. Carries forward `internal/logparse/` verbatim from L14.
- **New imports introduced:** none directly (goroutines + channels are language primitives); deeper use of `os` (ReadDir).
- **Running-example contribution:** the aggregator's bones — parallel walk + channel-based result collection.

### Lesson 17 — Select & timers

- **Concepts:** the `select` statement (multiplex receives/sends across N channels); `time.After(d)` returns a `<-chan Time`; `time.NewTicker(d)` for periodic events; the `default` branch (non-blocking select); cancellation via close-of-done-channel; common-mistake: tickers leak goroutines if `Stop()` isn't called.
- **Per-lesson example:** Add per-file timeout to the aggregator — a slow file (artificial sleep) should NOT block forever. `select { case res := <-fileCh: …; case <-time.After(timeout): logSlow(); skip(); }`.
- **Warm-up:** Implement `WaitWithTimeout[T any](ch <-chan T, d time.Duration) (T, error)` — block on ch or return error after d.
- **Main:** `aggregator.Walk(dir, timeout)` skips slow files and logs them. Updates `cmd/aggregator/main.go` to take a `-timeout=<dur>` flag.
- **New imports introduced:** `time` (revisit from L14).
- **Running-example contribution:** per-file timeout; one slow file no longer blocks the rest.

### Lesson 18 — `sync` & memory model

- **Concepts:** `sync.Mutex` (mutual exclusion), `sync.RWMutex` (many readers / one writer), `sync.WaitGroup` (count-down barrier), `sync.Once` (run-exactly-once), `sync/atomic` (lock-free single-word ops); the Go memory model intuition ("happens-before" relations created by channels, mutexes, sync primitives); the **race detector** introduced formally; the "share by communicating" vs "share by locking" tradeoff.
- **Per-lesson example:** Refactor the aggregator's coordination from "send maps through channels" (L16-17) to "shared map protected by `sync.Mutex` + `sync.WaitGroup`." Compare the two styles in slides; demonstrate the buggy non-mutex version failing under `-race`.
- **Warm-up:** Implement a `Counter` struct with `Inc()` and `Value() int`. Buggy version (no mutex) demonstrably fails under `go test -race`; mutex version passes. Both versions in the lesson; students see the failure first.
- **Main:** Aggregator gets a `sharedCounts map[string]int` guarded by `sync.Mutex`; workers `wg.Add(1)` then `defer wg.Done()`; main `wg.Wait()` before reading the map. README introduces `make test-race` as a daily habit from this lesson onward.
- **New imports introduced:** `sync`, `sync/atomic` (briefly).
- **Running-example contribution:** demonstrates the "share by locking" alternative; aggregator now has both implementations side-by-side for comparison.

### Lesson 19 — `context`

- **Concepts:** `context.Context` interface; `context.Background()` and `context.TODO()`; `context.WithCancel`, `WithTimeout`, `WithDeadline`; context propagation as the function's FIRST parameter; the `ctx.Done()` + `ctx.Err()` pattern; common-mistake: storing `ctx` in a struct; common-mistake: ignoring `ctx.Done()` in long-running loops.
- **Per-lesson example:** Add `ctx context.Context` as the first parameter of `aggregator.Walk`. Workers check `ctx.Done()` between files and stop early. `cmd/aggregator/main.go` uses `os/signal.NotifyContext(ctx, os.Interrupt)` so SIGINT cancels the walk gracefully.
- **Warm-up:** Implement `SleepWithCtx(ctx context.Context, d time.Duration) error` — sleep d unless ctx cancelled first; returns `ctx.Err()` on cancellation.
- **Main:** `Walk(ctx, dir, timeout)`; workers check `ctx.Done()`; CLI binary wires signal handling.
- **New imports introduced:** `context`, `os/signal`.
- **Running-example contribution:** the aggregator is now interruptible — Ctrl-C during a long walk stops cleanly with a partial result.

### Lesson 20 — Concurrency patterns

- **Concepts:** worker pool (N workers consume from a shared channel); fan-in (multiplex N channels into one); fan-out (split work across N consumers); pipeline (chain stages with channels between); **lesson-local `errgroupx`** for "first error cancels the rest." Slides + README name-drop `golang.org/x/sync/errgroup` as the production equivalent.
- **Per-lesson example:** Aggregator becomes a proper worker pool — `cmd/aggregator -workers=N`. Producer sends file paths into a `pathsCh`; N workers read from it and send results into a `resultsCh`; main reduces. If any worker errors, the `errgroupx.Group` cancels the rest via the shared ctx.
- **Warm-up:** Implement `FanIn[T any](chans ...<-chan T) <-chan T` (generic from L12) — multiplex N input channels into one output.
- **Main:** Aggregator worker pool + new `internal/errgroupx/` package (Group with `Go(func() error)` + `Wait() error`; uses `sync.Once` to capture first error + `context.CancelFunc` to cancel siblings).
- **New imports introduced:** none net-new; introduces lesson-local `errgroupx` subpackage.
- **Running-example contribution:** the aggregator is now a proper concurrent program — bounded parallelism via `-workers=N` and first-error-cancels-rest.

### Lesson 21 — Networking & syscalls (standalone)

- **Concepts:** the `net` package; TCP basics (`net.Listen("tcp", addr)` + accept loop + per-connection goroutine); `net.Dial` for clients; UDP basics (briefly); signals (`os/signal`, `signal.NotifyContext`); file descriptors as a concept; the `syscall` package overview (rarely used directly in app code).
- **Per-lesson example:** A tiny TCP "echo" service — server in `cmd/echo-server/main.go`, client in `cmd/echo-client/main.go`. Server's accept loop spawns one goroutine per connection; each connection reads lines and writes them back uppercased. **Standalone** — does NOT extend the aggregator (network and aggregator don't naturally compose without bigger scope; Phase 4's HTTP work bridges them).
- **Warm-up:** Implement `LineEcho(conn net.Conn) error` — read lines from conn (bufio.Scanner from L13), write each back uppercased (bufio.Writer or fmt.Fprintln).
- **Main:** Echo server + client. Tests use `net.Listen("tcp", "127.0.0.1:0")` to bind to an ephemeral port, then dial it for in-process integration testing.
- **New imports introduced:** `net`, `os/signal` (revisit).
- **Running-example contribution:** none (standalone).

### Lesson 22 — Profiling, benchmarking, fuzzing — Phase 3 wrap-up

- **Concepts:** `runtime/pprof` for CPU + heap profiles; `net/http/pprof` (import-side-effect) for live introspection; `go test -bench` (deeper than L15 preview); escape analysis intuition; fuzz tests (deeper than L15 preview — actually run with `-fuzz` and find a bug).
- **Per-lesson example:** Profile the L20 aggregator under load — generate a synthetic 1000-file dataset, run `cmd/aggregator-profile/main.go` with pprof CPU + heap collection, identify an allocation hot path (e.g., repeated regex compilation or string concatenation), fix it, demonstrate the speedup via benchmark. Add fuzz tests for `logparse.parseLine` that run for 30 seconds and find at least one input that exposes a recovery path.
- **Warm-up:** Add `BenchmarkParse` for `logparse.Parse` over 10k synthetic log lines (carry-forward measurement before optimization).
- **Main:** Profile + benchmark + fuzz the aggregator. Includes `cmd/aggregator-profile/main.go` that wraps `Walk` with pprof CPU + heap dumps written to `cpu.prof` and `heap.prof`. README documents `go tool pprof -http=:8080 cpu.prof` for interactive analysis.
- **New imports introduced:** `runtime/pprof`, `net/http/pprof`.
- **Running-example contribution:** the aggregator is profiled, optimized, and fuzz-tested — Phase 3 closes with a measured, hardened version of what we built.

## File layout per lesson (Phase 2 pattern continues)

```
lessons/NN-name/
├── README.md
├── slides/
│   ├── index.html, slides.md, assets/.gitkeep
├── exercises/
│   ├── warmup/<topic>/
│   │   ├── <topic>.go
│   │   └── <topic>_test.go
│   ├── cmd/<binary>/main.go             (where natural)
│   └── internal/
│       ├── logparse/                     (carry-forward from L14)
│       ├── aggregator/                   (the running example, L16-L20)
│       └── <other-subpkgs>/
└── solutions/  (mirrored)
```

### Running-example carry-forward

Each lesson L17-L20 carries forward the PREVIOUS lesson's `internal/aggregator/` and modifies/extends it. Same pattern as Phase 2's tracker evolution (L09 → L10 → L11 → L13 → L15). Each lesson's exercises/ tree has the prior implementation pre-installed; students extend it.

L21 and L22 carry forward selectively:
- L21 (networking, standalone): no aggregator code; uses `internal/logparse` only as backdrop reference if at all
- L22 (profiling, wraps Phase 3): carries forward L20's full aggregator + `errgroupx`

## Lesson naming (working slugs)

| # | Topic | Slug |
|---|---|---|
| 16 | Goroutines & channels | `16-goroutines-channels` |
| 17 | Select & timers | `17-select-timers` |
| 18 | sync & memory model | `18-sync-memory-model` |
| 19 | context | `19-context` |
| 20 | Concurrency patterns | `20-concurrency-patterns` |
| 21 | Networking & syscalls | `21-networking` |
| 22 | Profiling, bench, fuzz | `22-profiling-fuzz` |

The `tools/build-index/main.go` master list needs updating to match — each lesson's plan can confirm slug accuracy (precedent: Plan Q fixed L14's slug from `stdlib` to `time-strings-regex`).

## Cross-lesson invariants

- **Skeleton tests in BOTH warmup and main** — Phase 2 default continues.
- **Subpackage layout** — no flat warmup.go/main.go at the lesson root.
- **No `Warmup*` prefix** — subpackages provide namespacing.
- **Slides + README written inline by controller** — Plan K-R pattern; subagents handle code-bearing tasks.
- **`make test-race`** featured in Phase 3 daily-habits READMEs starting L18.
- **5-concept slide deck** where natural (4 acceptable for smaller-scope lessons).
- **Lesson-local subpackages over carry-forward import** — each lesson has its own copy of `logparse`, `aggregator`, etc., with import paths rewritten. (Same as Phase 2.)
- **Conventional Commits.**
- **gofmt 1.19+ godoc list normalization** acceptable for skeleton files.

## Phase 3 non-scope

- Distributed systems (Phase 4 — message queues, gRPC, etc.).
- HTTP servers (Phase 4 lesson 23).
- Database connections / connection pooling (could appear in Phase 4 deployment chapters).
- Goroutine schedulers internals (advanced topic; mention in passing only).
- Custom allocators / GC tuning (advanced; mention in passing only).
- Generics beyond what L12 already taught (Phase 3 uses them in `FanIn[T]`, `WaitWithTimeout[T]`, and the `errgroupx` `Go` method type — that's it).
- HTTP-based aggregator (Phase 4 territory — would extend the aggregator with a `/stats` endpoint).
- TLS / cryptography (Phase 4 deployment lessons).

## Open issues (to revisit per-lesson)

- **L21 scope**: TCP echo is a tight standalone example. If it feels too light, consider adding a small UDP example or a simple line-based protocol (e.g., a "type a number, get the square back" service). Per-lesson brainstorming locks the exact scope.
- **L22 fuzz target**: `logparse.parseLine` is the obvious target. If the regex is bulletproof and fuzz finds nothing in 30 seconds, that's a feature not a bug — the slides reframe "no crash found" as a positive signal. Per-lesson brainstorming confirms.
- **Aggregator + `cmd/expenses` coexistence in lessons/16+**: every Phase 3 lesson's `internal/` will contain `logparse` + `aggregator`; no `expense`/`store`/`csvimport`/`summary` from L15. Students who want the tracker work back can `cd lessons/15-structure/` — Phase 3 lessons don't re-host it.

## Execution after approval

Mirror the Phase 2 design PR (PR #21) process:

1. User reviews this design doc → approves or requests changes
2. PR for the design doc only (not the implementation)
3. Once merged, per-lesson brainstorming begins for L16 (Plan S)
4. Each lesson follows the established subagent-driven development flow: brainstorm → plan → execute → review → PR → merge
