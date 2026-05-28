# Lesson 22: Profiling, benchmarking & fuzzing

## What you'll learn

By the end of this lesson you can:

- Write **benchmarks** with `go test -bench` and read `ns/op` · `B/op` · `allocs/op`.
- Collect and read **CPU profiles** with `runtime/pprof` + `go tool pprof` — and trust the profile over your hunch.
- Collect **heap/allocation profiles**, tell `alloc_space` from `inuse_space`, and use **escape analysis** (`go build -gcflags=-m`) to understand *why* a value lands on the heap.
- Expose live profiles from a running process with **`net/http/pprof`** — safely.
- Write **fuzz tests** (`go test -fuzz`), seed a corpus, use an **oracle**, and catch the bug an optimization slips in.

This is the Phase 3 finale. We measure and harden the log aggregator built across L16-L20 — the capstone by another name.

## What's different from L21

L21 was a standalone networking detour (TCP echo). **L22 returns to the aggregator** (`internal/{logparse,aggregator,errgroupx}` + `cmd/aggregator` are carried forward verbatim from L20) and does the thing every engineer eventually must: prove a change is faster *and* still correct. The new surface is the measurement layer — a `bench` warm-up, a hand-written `parseLineFast`, benchmark + fuzz tests, and a `cmd/aggregator-profile` binary.

## The package layout

```
lessons/22-profiling-fuzz/exercises/
├── warmup/bench/                       ← you implement: GenLines + BenchmarkParse
├── cmd/
│   ├── aggregator/                     carried forward from L20 (the CLI)
│   └── aggregator-profile/             ← provided: pprof CPU+heap + live endpoint
└── internal/
    ├── logparse/                       ← you implement parseLineFast; + bench + fuzz
    ├── aggregator/                     carried forward from L20 (Walk/WalkLocked/WalkPool)
    └── errgroupx/                      carried forward from L20
```

---

## Concept 1 — Benchmarking

`go test` has a benchmark runner. Write `func BenchmarkX(b *testing.B)`, loop `b.N` times, and the runner auto-scales `b.N` until timing is stable:

```go
var sink LogEntry

func BenchmarkParseLineFast(b *testing.B) {
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        e, _ := parseLineFast(benchLine)
        sink = e
    }
}
```

```bash
go test -bench=ParseLine -benchmem ./...
# BenchmarkParseLineRegex-10   1823000   660.1 ns/op   128 B/op   2 allocs/op
# BenchmarkParseLineFast-10    9887379   120.6 ns/op     0 B/op   0 allocs/op
```

`ns/op` is time per call; `B/op` and `allocs/op` (from `b.ReportAllocs()` or `-benchmem`) are bytes and heap allocations per call. **Allocations are often the number that matters most** — each one is future work for the garbage collector.

### Common mistake

**Dead-code elimination.** If the result is unused, the compiler may delete the call and you'll "measure" a bogus `0.3 ns/op`. Assign to a package-level `sink` the compiler can't prove is dead. Suspect this any time a benchmark looks impossibly fast.

---

## Concept 2 — CPU profiling

A benchmark says a function is slow; a profile says *where the whole program spends its time*. The profiler samples the running program ~100×/second and ranks the call tree. **Profile first** — the bottleneck is rarely where you'd guess.

```go
f, _ := os.Create("cpu.prof")
pprof.StartCPUProfile(f)
defer pprof.StopCPUProfile()
// ... workload ...
```

```bash
go test -bench=. -cpuprofile=cpu.prof ./...   # or runtime/pprof in a program
go tool pprof -top -nodecount=10 cpu.prof     # text
go tool pprof -http=:8080 cpu.prof            # flame graph in a browser
```

Profiling the aggregator over 2000 files is illuminating:

```
Total samples = 12050ms
   10080ms 83.65%  syscall.rawsyscalln
     490ms  4.07%  runtime.usleep
```

**84% is syscalls** — opening and reading files. The parser barely shows up; end-to-end, the aggregator is *I/O-bound*, not parse-bound. That's the real lesson: the profile told us the truth, and "optimizing" the parser would not make this workload faster. (Crank `-lines=5000` to make files large and watch parsing climb — then the parser's cost, and its allocations, start to matter.)

### Common mistake

**A run too short to sample.** The profiler samples ~100×/s; a 5ms workload yields ~0 samples and an empty profile. Loop the work for a second or two (`cmd/aggregator-profile -duration=2s`), or profile a benchmark with `-benchtime=5s`.

---

## Concept 3 — Heap / allocation profiling + escape analysis

Every heap allocation costs the allocator now and the GC later. Heap profiling shows *what* allocates; escape analysis shows *why*.

```bash
go test -bench=. -memprofile=heap.prof ./...
go tool pprof -alloc_space heap.prof    # TOTAL bytes allocated (churn)
go tool pprof -inuse_space heap.prof    # bytes LIVE now (leaks)
go build -gcflags=-m ./internal/logparse/ 2>&1 | grep escapes
```

A value **escapes to the heap** when the compiler can't prove it stays within the function (returned, stored in an interface, captured by an outliving closure). Non-escaping values live on the stack — free.

Why does the regex parser allocate 2× per call and the fast one 0×?

```go
m := logLineRE.FindStringSubmatch(s)   // allocates a []string every call → 2 allocs
// vs.
level := rest[:sp]                     // string slice: shares backing array → 0 allocs
```

`s[i:j]` on a string returns a header into the *same* bytes — no copy, no allocation. `FindStringSubmatch` must materialize a slice. That's the 128 B / 2 allocs the benchmark measured, gone.

### Common mistake

**Confusing `alloc_space` with `inuse_space`.** `alloc_space` is cumulative (churn — what to reduce for GC pressure); `inuse_space` is a live snapshot (what to chase for leaks). Reach for the wrong one and you optimize the wrong problem.

---

## Concept 4 — `net/http/pprof`

Profiles from tests are for development. To answer "why is *production* slow right now?" you profile a live process. One blank import exposes the profiles over HTTP:

```go
import _ "net/http/pprof"   // registers /debug/pprof/* on http.DefaultServeMux

func main() {
    go func() { http.ListenAndServe("localhost:6060", nil) }()
    // ... real program ...
}
```

```bash
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30   # live CPU
go tool pprof http://localhost:6060/debug/pprof/heap                 # live heap
curl  http://localhost:6060/debug/pprof/goroutine?debug=1            # goroutine dump
```

`cmd/aggregator-profile -httppprof=localhost:6060` wires exactly this. The `goroutine` endpoint is the first thing to check when goroutine count climbs forever — it lists every stuck goroutine and where it's blocked.

### Common mistake

**Exposing pprof publicly.** The blank import registers on the *default* mux. If your public server uses `http.DefaultServeMux`, you've published `/debug/pprof/` to the internet — internals leak, and `?seconds=600` is a free DoS lever. Bind pprof to `localhost` on its own port; reach it via an SSH tunnel.

---

## Concept 5 — Fuzzing

Table tests check the cases you imagined. Fuzzing checks the ones you didn't — millions of mutated inputs hunting for panics or disagreements with a reference.

```go
func FuzzParseLineFast(f *testing.F) {
    f.Add("2026-01-02T15:04:05 INFO ok")
    f.Add("0")
    f.Fuzz(func(t *testing.T, s string) {
        _, _ = parseLineFast(s)   // property: never panics on ANY input
    })
}
```

```bash
go test -run='^$' -fuzz=FuzzParseLineFast -fuzztime=30s ./...
```

Under plain `go test`, only the `f.Add` seeds run (deterministic regression cases). Under `-fuzz`, the engine mutates inputs guided by coverage; a crasher is saved to `testdata/fuzz/<Name>/<hash>` forever.

**Fuzzing earns its keep in this lesson.** A naive fast parser finds the timestamp with `i := strings.IndexByte(s, ' ')` then `s[:i]`. Table tests pass. Then:

```
--- FAIL: FuzzParseLineFast
    panic: slice bounds out of range [:-1]
        string("0")
```

`IndexByte("0", ' ')` returns `-1`; `s[:-1]` panics. Found in under a second. The fix is a length guard. We also use an **oracle**: whenever the trusted regex parser accepts an input, the fast parser must return the identical result — catching silent wrong answers, not just crashes.

> Subtlety worth knowing: the regex `\s+(.+)$` *backtracks* and accepts a whitespace-only message on raw input (`"TS INFO  "` → message `" "`), while the hand-written parser strips all leading space and rejects it. That divergence only exists on *untrimmed* input — and `Parse` always `TrimSpace`s first. So the fuzz oracle compares on the *trimmed* input (the contract `Parse` actually uses), while the no-panic property still guards every raw byte. Lesson: **fuzz the contract your callers actually exercise.**

### Common mistake

**A non-deterministic fuzz property.** If the property depends on time, randomness, or map order, the fuzzer "finds" irreproducible failures and saves useless corpus entries. A fuzz property must be a pure function of its input — compare to a deterministic oracle or assert an invariant.

---

## Exercise: warm-up — `bench`

Implement `GenLines(n int) string` (n deterministic, well-formed log lines) in `exercises/warmup/bench/bench.go`, and write `BenchmarkParse` over it. This fixed input feeds every later benchmark.

**Time:** 10-15 minutes.

## Exercise: main — `logparse` optimization (+ provided profiler)

1. Implement `parseLineFast` in `internal/logparse/logparse.go` — hand-written, allocation-free, guarding every index — and point `Parse` at it. Fill in `BenchmarkParseLineRegex`/`BenchmarkParseLineFast` and `FuzzParseLineFast` (never panic + agree with the regex oracle on trimmed input).
2. Study `cmd/aggregator-profile/main.go` (provided) — how `runtime/pprof` and `net/http/pprof` wrap `WalkPool`.

**Time:** 45-60 minutes.

---

## Daily habits

After every change:

```bash
gofmt -w ./...
go vet ./...
go test ./...
make test-race    # daily habit since L18
```

## How to run

```bash
# Skeleton (passes vacuously until you implement)
cd lessons/22-profiling-fuzz/exercises
go test ./...

# Reference solution
go test -v ./lessons/22-profiling-fuzz/solutions/...
go test -race ./lessons/22-profiling-fuzz/solutions/...
make test-race

# Benchmark: regex vs fast (5.5× faster, 0 allocs on the reference impl)
go test -bench=ParseLine -benchmem ./lessons/22-profiling-fuzz/solutions/internal/logparse/...

# Fuzz the fixed parser (finds nothing → positive signal)
go test -run='^$' -fuzz=FuzzParseLineFast -fuzztime=30s ./lessons/22-profiling-fuzz/solutions/internal/logparse/

# Profile the aggregator and inspect
go run ./lessons/22-profiling-fuzz/solutions/cmd/aggregator-profile -n=2000 -duration=2s -cpuprofile=/tmp/cpu.prof -memprofile=/tmp/heap.prof
go tool pprof -top -nodecount=8 /tmp/cpu.prof
go tool pprof -alloc_space -top -nodecount=8 /tmp/heap.prof
rm -f /tmp/cpu.prof /tmp/heap.prof

# Live pprof endpoint
go run ./lessons/22-profiling-fuzz/solutions/cmd/aggregator-profile -httppprof=localhost:6060 -duration=30s &
go tool pprof -top http://localhost:6060/debug/pprof/profile?seconds=5
```

## Going further

### Read

- **Go blog — "Profiling Go Programs"**: <https://go.dev/blog/pprof> — the canonical pprof walkthrough.
- **Go docs — testing (Benchmarks + Fuzzing)**: <https://pkg.go.dev/testing> — `B`, `F`, the fuzzing contract, corpus format.
- **Go blog — "Fuzzing is Beta Ready" / fuzzing tutorial**: <https://go.dev/doc/tutorial/fuzz> — `f.Add`, `f.Fuzz`, corpus, minimization.
- **Dave Cheney — "Five things that make Go fast" / escape analysis posts** — intuition for stack vs heap.

### Try

- **Reintroduce the bug.** Rewrite `parseLineFast` with the naive `strings.IndexByte(s, ' ')` + `s[:i]` (no guard), run `go test -fuzz=FuzzParseLineFast`, and watch it crash on `"0"` within a second. Then fix it and confirm the saved crasher passes. This is the whole lesson in 3 minutes.
- **Make it parse-bound.** Run `aggregator-profile -n=50 -lines=20000 -duration=3s -cpuprofile=cpu.prof` and compare the `top` output to the I/O-bound default. Where does `parseLineFast` rank now?
- **Profile `WalkLocked` vs `WalkPool`.** Add a `-mode` flag to `aggregator-profile` and CPU-profile both over the same dataset. Which spends more time in runtime scheduling / lock contention?
- **Add a block profile.** Wire `runtime.SetBlockProfileRate` + a `/debug/pprof/block` analysis and find where goroutines wait on channel sends/receives in the worker pool.
- **Differential-fuzz two implementations.** Keep both `parseLineRegex` and `parseLineFast` and fuzz them as a strict two-way oracle on trimmed input — does the equivalence hold for 5 minutes of fuzzing?

---

> Phase 3 — Concurrency & Systems — is complete. You can now write, reason about, and *measure* concurrent Go. **Phase 4 — Production & Distributed** begins at Lesson 23: HTTP servers (`net/http`, `slog`). The aggregator stays here, measured and hardened; we start building services the outside world talks to.
