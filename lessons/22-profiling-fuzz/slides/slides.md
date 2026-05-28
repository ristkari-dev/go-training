<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">22</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 3 — Concurrency &amp; Systems</div>
<h1>Profiling, benchmarking &amp; fuzzing</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Measure before you optimize. Learn <code>go test -bench</code>, CPU + heap profiling with <code>runtime/pprof</code>, the live <code>net/http/pprof</code> endpoint, escape analysis, and fuzzing. The Phase 3 finale: profile the aggregator, replace its regex parser with an allocation-free one (5.5× faster), then fuzz the replacement to catch the bug it introduced.</p>
</div>
</div>
</div>

---

## What we'll cover

- **Benchmarking** — `go test -bench`, `b.N`, `b.ReportAllocs`; reading ns/op · B/op · allocs/op.
- **CPU profiling** — `runtime/pprof`, `go tool pprof`, flame graphs; where does the time *actually* go?
- **Heap / allocation profiling + escape analysis** — why a value escapes to the heap; `alloc_space` vs `inuse_space`.
- **`net/http/pprof`** — live introspection of a running process.
- **Fuzzing** — `go test -fuzz`, the corpus, oracle fuzzing; catching the bug an optimization slipped in.

---

## The story so far

Phase 3 built the log aggregator across L16-L20: goroutines, channels, select, sync, context, a worker pool. L21 was a networking detour. **L22 returns to the aggregator to MEASURE and HARDEN it** — the Phase 3 capstone by another name.

The thesis of this lesson: **never optimize on a hunch.** Measure first (profiling tells you *where*), isolate the suspect (benchmarking tells you *how much*), change it, measure again — and fuzz the change, because a faster implementation is worthless if it's subtly wrong. We'll live that whole loop on the aggregator's log parser.

---

## Concept 1: Benchmarking

### Motivation

"Is this faster?" is not a question you answer by squinting at code. Go has a benchmark runner built into `go test`. You write a `Benchmark*` function; the runner calls it `b.N` times, auto-scaling `b.N` until the timing is statistically stable, and reports nanoseconds per operation — plus, with one flag, bytes and allocations per operation.

---

### The basics

```go
func BenchmarkParseLineFast(b *testing.B) {
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        e, _ := parseLineFast(benchLine)
        sink = e   // assign to a package var — see the common mistake
    }
}
```

Run it:

```bash
go test -bench=ParseLine -benchmem ./...
```

```
BenchmarkParseLineRegex-10   1823000   660.1 ns/op   128 B/op   2 allocs/op
BenchmarkParseLineFast-10    9887379   120.6 ns/op     0 B/op   0 allocs/op
```

Read the columns:

- **`-10`** — `GOMAXPROCS` (10 cores here).
- **iterations** — how many times `b.N` settled on.
- **`ns/op`** — nanoseconds per call. Lower is faster.
- **`B/op`** — bytes allocated per call.
- **`allocs/op`** — heap allocations per call. Often the number that matters most: allocations feed the garbage collector.

`b.ResetTimer()` discards setup time; `b.ReportAllocs()` turns on the alloc columns (or pass `-benchmem`).

---

### A worked example

The aggregator's regex parser vs the hand-written one:

| Parser | ns/op | B/op | allocs/op |
|---|---|---|---|
| `parseLineRegex` | 660 | 128 | 2 |
| `parseLineFast` | 121 | 0 | 0 |

**5.5× faster, and allocation-free.** The two allocations in the regex version are `FindStringSubmatch` building a `[]string` of submatches, every single call. The fast parser slices the input string in place and allocates nothing.

(Your machine's numbers will differ. The *shape* — fast is several times quicker with zero allocs — won't.)

---

### Common mistake

**Dead-code elimination.** If you don't use the benchmark's result, the compiler may delete the whole call:

```go
func BenchmarkParse(b *testing.B) {
    for i := 0; i < b.N; i++ {
        parseLineFast(benchLine)   // ❌ result unused → may be optimized away
    }
}
// Result: a suspiciously fast "0.3 ns/op" — you benchmarked nothing.
```

Fix: assign the result to a package-level `sink` variable the compiler can't prove is unused:

```go
var sink LogEntry

func BenchmarkParse(b *testing.B) {
    for i := 0; i < b.N; i++ {
        e, _ := parseLineFast(benchLine)
        sink = e   // ✓ keeps the call alive
    }
}
```

If a benchmark reports sub-nanosecond times, suspect dead-code elimination first.

---

### Recap

- `go test -bench=. -benchmem` runs `Benchmark*` functions and reports ns/op · B/op · allocs/op.
- `b.ResetTimer()` excludes setup; `b.ReportAllocs()` shows allocations.
- **Allocations drive GC pressure** — `allocs/op` is often the number to watch.
- Consume the result (`sink`) or the compiler may benchmark nothing.

---

## Concept 2: CPU profiling

### Motivation

A benchmark tells you a function is slow. A **profile** tells you *where the program spends its time* — across the whole call tree, ranked. You don't guess which function is hot; the profiler samples the running program (≈100×/sec) and shows you.

The first rule of optimization: **profile first.** The bottleneck is almost never where you think.

---

### The basics

Two ways to collect a CPU profile:

```bash
# 1. From a test/benchmark:
go test -bench=. -cpuprofile=cpu.prof ./...

# 2. From a program, via runtime/pprof:
```
```go
f, _ := os.Create("cpu.prof")
pprof.StartCPUProfile(f)
defer pprof.StopCPUProfile()
// ... the workload you want to profile ...
```

Then analyze:

```bash
go tool pprof -top -nodecount=10 cpu.prof     # text: hottest functions
go tool pprof -http=:8080 cpu.prof            # browser: flame graph, call graph
```

In the interactive tool: `top` (ranked list), `list parseLineFast` (per-line annotated source), `web` (call graph).

---

### A worked example

Profiling the aggregator end-to-end (`cmd/aggregator-profile -duration=2s`) over 2000 files:

```
Duration: 2.21s, Total samples = 12050ms
      flat  flat%   sum%        cum   cum%
   10080ms 83.65% 83.65%    10080ms 83.65%  syscall.rawsyscalln
     490ms  4.07% 87.72%      490ms  4.07%  runtime.usleep
     480ms  3.98% 91.70%      480ms  3.98%  runtime.pthread_cond_wait
```

The surprise: **84% is syscalls** — opening and reading 2000 files. The parser barely registers. The bottleneck is *file I/O*, not parsing.

This is the lesson within the lesson: **the profile told us the truth.** If we'd "optimized" the parser to make the aggregator faster, we'd have wasted our time — I/O dominates this workload. To actually speed up the aggregator end-to-end you'd batch reads, use fewer/larger files, or parallelize I/O (which the worker pool already does).

So why optimize the parser at all? Two reasons: (1) when files are *large* (thousands of lines each), parsing becomes a real fraction — bump `-lines=5000` and re-profile to see it climb; (2) the parser's *allocations* create GC pressure that hurts the whole process. Which is exactly what the next concept measures.

---

### Common mistake

**Profiling a run too short to sample.** The CPU profiler samples ~100×/second. A workload that finishes in 5ms yields ~0 samples and a useless, empty-looking profile:

```
(pprof) top
Total samples = 0
```

Fix: profile enough work. Loop the operation for a second or two (our `cmd/aggregator-profile -duration=2s` does exactly this), or profile a benchmark (`-cpuprofile` with `-benchtime=5s`).

---

### Recap

- **Profile before optimizing** — the hot spot is rarely where you'd guess.
- Collect via `go test -cpuprofile` or `runtime/pprof.StartCPUProfile`.
- Analyze with `go tool pprof`: `top`, `list`, `-http` flame graph.
- The aggregator is I/O-bound end-to-end — profiling proved it.

---

## Concept 3: Heap / allocation profiling + escape analysis

### Motivation

CPU time isn't the only cost. Every heap allocation is work for the allocator now and the garbage collector later. A function that allocates on every call quietly taxes the whole program. Heap profiling shows *what allocates*; escape analysis shows *why*.

---

### The basics

Collect a heap profile:

```bash
go test -bench=. -memprofile=heap.prof ./...
```
```go
f, _ := os.Create("heap.prof")
runtime.GC()                  // get up-to-date stats
pprof.WriteHeapProfile(f)
```

Two views of the same profile:

```bash
go tool pprof -alloc_space heap.prof    # TOTAL bytes allocated over time (cumulative)
go tool pprof -inuse_space heap.prof    # bytes LIVE right now (snapshot)
```

**Escape analysis** explains why a value lands on the heap. Ask the compiler:

```bash
go build -gcflags=-m ./internal/logparse/ 2>&1 | grep escapes
```
```
./logparse.go:45:30: m escapes to heap
```

A value "escapes" when the compiler can't prove it stays within the function — e.g. it's returned, stored in an interface, or captured by a closure that outlives the call. Escaped values are heap-allocated; non-escaping ones live on the stack (free).

---

### A worked example

Why does `parseLineRegex` allocate 2× and `parseLineFast` 0×?

```go
// regex: FindStringSubmatch allocates a []string of submatches, every call
m := logLineRE.FindStringSubmatch(s)   // 1 slice header + 1 backing array → 2 allocs
return LogEntry{Time: t, Level: m[2], Message: m[3]}, nil

// fast: slices the input in place — substrings share the input's backing array
level := rest[:sp]                     // no copy, no alloc
msg := rest[sp+...:]                   // no copy, no alloc
return LogEntry{Time: t, Level: level, Message: msg}, nil
```

`s[i:j]` on a string returns a new string *header* pointing into the same backing bytes — **no allocation**. The regex API, by contrast, must materialize a `[]string` to hand you the submatches. That's the 128 B / 2 allocs per call the benchmark measured.

---

### Common mistake

**Confusing `alloc_space` with `inuse_space`.** `alloc_space` is cumulative — every byte ever allocated, including freed ones; it's what you want for "what's churning the allocator." `inuse_space` is a live snapshot — what's resident *now*; it's what you want for "what's leaking / holding memory." Reaching for the wrong one sends you optimizing the wrong thing (e.g. chasing a "leak" that's really just high churn the GC handles fine).

---

### Recap

- Heap profiling (`-memprofile`) shows what allocates; `alloc_space` (churn) vs `inuse_space` (live).
- Escape analysis (`go build -gcflags=-m`) shows *why* a value is heap-allocated.
- Slicing a string allocates nothing (shared backing array); `FindStringSubmatch` allocates a slice every call.
- Fewer allocations → less GC pressure → a faster, smoother whole program.

---

## Concept 4: `net/http/pprof`

### Motivation

Profiles from tests are great in development. But the interesting questions — why is *production* slow at 3am? what's allocating under real traffic? — need profiling on a *live, running* process. `net/http/pprof` exposes the same profiles over HTTP, on demand, with no redeploy.

---

### The basics

A single blank import registers the handlers:

```go
import _ "net/http/pprof"   // registers /debug/pprof/* on http.DefaultServeMux

func main() {
    go func() { http.ListenAndServe("localhost:6060", nil) }()
    // ... your real program ...
}
```

Now, against the running process:

```bash
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30   # 30s CPU profile
go tool pprof http://localhost:6060/debug/pprof/heap                 # heap snapshot
curl http://localhost:6060/debug/pprof/goroutine?debug=1             # all goroutine stacks
```

Our `cmd/aggregator-profile -httppprof=:6060` does exactly this — start it, then point `go tool pprof` at the live endpoint while the walk loops.

---

### A worked example

```bash
# Terminal 1: run the aggregator with the live endpoint
go run ./.../cmd/aggregator-profile -httppprof=localhost:6060 -duration=60s

# Terminal 2: grab a 10s CPU profile from the running process
go tool pprof -top -nodecount=5 http://localhost:6060/debug/pprof/profile?seconds=10
```

The `/debug/pprof/goroutine` endpoint is gold for diagnosing leaks — if your goroutine count climbs forever, this lists every stuck goroutine and where it's blocked.

---

### Common mistake

**Exposing the pprof endpoint publicly.** The blank import registers handlers on the *default* mux. If your application's public HTTP server uses `http.DefaultServeMux`, you've just published `/debug/pprof/` to the internet — leaking internals and handing attackers a CPU-burning DoS lever (`?seconds=600`).

Fix: serve pprof on a **separate, bound-to-localhost** listener (`localhost:6060`), never the public one, and firewall it. In production, reach it via an SSH tunnel or an internal-only port.

---

### Recap

- `import _ "net/http/pprof"` + an HTTP server exposes live profiles at `/debug/pprof/`.
- `go tool pprof http://host/debug/pprof/{profile,heap,goroutine}` pulls them on demand.
- `/debug/pprof/goroutine` is the go-to for leak diagnosis.
- **Never expose pprof on a public interface** — bind it to localhost.

---

## Concept 5: Fuzzing

### Motivation

You replaced a battle-tested regex parser with hand-written index math for a 5.5× win. Are you *sure* it's correct? Table tests check the cases you thought of. **Fuzzing checks the cases you didn't** — it throws millions of mutated inputs at your function looking for panics or for disagreements with a reference.

---

### The basics

A fuzz target looks like a test with an `f *testing.F`:

```go
func FuzzParseLineFast(f *testing.F) {
    f.Add("2026-01-02T15:04:05 INFO ok")   // seed corpus
    f.Add("0")
    f.Fuzz(func(t *testing.T, s string) {
        // property that must hold for ALL s:
        _, _ = parseLineFast(s)            // must never panic
    })
}
```

- **`f.Add(...)`** seeds the corpus with starting inputs.
- **`f.Fuzz(func(t, in))`** is the property under test; the runner mutates the seeds and calls it millions of times.
- Under plain `go test`, only the seeds run (fast, deterministic — a regression test).
- Under `go test -fuzz=FuzzParseLineFast`, the engine mutates inputs continuously, guided by code coverage.

```bash
go test -run='^$' -fuzz=FuzzParseLineFast -fuzztime=30s ./...
```

A crasher is written to `testdata/fuzz/FuzzParseLineFast/<hash>` and becomes a permanent regression case.

---

### A worked example — fuzzing catches our bug

A first cut of the fast parser finds the timestamp by splitting on a space:

```go
i := strings.IndexByte(s, ' ')
tsStr := s[:i]                 // looks fine...
```

Table tests pass. Then:

```bash
go test -fuzz=FuzzParseLineFast
--- FAIL: FuzzParseLineFast
    panic: runtime error: slice bounds out of range [:-1]
    Failing input written to testdata/fuzz/FuzzParseLineFast/...
        string("0")
```

`IndexByte("0", ' ')` returns **-1** (no space), so `s[:-1]` panics. The fuzzer found it in *under a second* with the seed `"0"`. The fix is a guard:

```go
if len(s) < tsLen { return LogEntry{}, errMalformed }   // never index blindly
```

We also use an **oracle**: whenever the trusted regex parser accepts an input, the fast parser must produce the identical result. That catches not just panics but silent *wrong answers* — the optimization preserving behavior, not just avoiding crashes.

---

### Common mistake

**A non-deterministic fuzz target.** If your property depends on time, randomness, or map iteration order, the fuzzer "finds" failures that don't reproduce, and the saved corpus entries are useless:

```go
f.Fuzz(func(t *testing.T, s string) {
    if parse(s) != time.Now().Second()%2 { t.Fail() }   // ❌ flaky, irreproducible
})
```

A fuzz property must be a pure function of its input. Compare against a deterministic oracle or assert an invariant (never panics, round-trips, output is sorted) — never against a moving target.

---

### Recap

- Fuzzing mutates inputs to find panics and oracle disagreements you'd never enumerate by hand.
- `f.Add` seeds; `f.Fuzz` is the property; `go test -fuzz` runs the engine; crashers persist in `testdata/fuzz`.
- Use an **oracle** (a trusted reference) to catch wrong answers, not just crashes.
- Keep the property **deterministic** — pure function of the input.

---

## Practice

### Warm-up

In `exercises/warmup/bench/`, implement `GenLines(n int) string` — n deterministic, well-formed synthetic log lines — and write `BenchmarkParse` over it. This is the fixed input every later benchmark reuses.

```bash
cd lessons/22-profiling-fuzz/exercises
go test ./warmup/bench/...
go test -bench=. -benchmem ./warmup/bench/...
```

---

### Main

1. **`internal/logparse`** — implement `parseLineFast` (hand-written, allocation-free), then point `Parse` at it. Add `BenchmarkParseLineRegex` vs `BenchmarkParseLineFast`, and `FuzzParseLineFast` (never panic + agree with the regex oracle on trimmed input).
2. **`cmd/aggregator-profile`** (provided) — study how it wires `runtime/pprof` + `net/http/pprof` around `WalkPool`.

```bash
cd lessons/22-profiling-fuzz/exercises
go test ./...
go test -race ./...
make test-race      # daily habit (since L18)

# the whole loop:
go test -bench=ParseLine -benchmem ./internal/logparse/...
go test -run='^$' -fuzz=FuzzParseLineFast -fuzztime=30s ./internal/logparse/...
go run ../solutions/cmd/aggregator-profile -n=2000 -duration=2s -cpuprofile=/tmp/cpu.prof
go tool pprof -top -nodecount=8 /tmp/cpu.prof
```

---

## Closing thought

The discipline is a loop, and every step is a different tool:

1. **Benchmark** to get a number you can compare against.
2. **Profile** to find where the time and the allocations actually go.
3. **Optimize** the thing the profile pointed at — not the thing you assumed.
4. **Benchmark again** to prove the win is real.
5. **Fuzz** to prove the optimization didn't break correctness.

We did all five on the aggregator's parser: benchmarked it (660 ns, 2 allocs), profiled the aggregator (found I/O dominates, parsing's cost is its allocations), optimized (hand-written parser: 121 ns, 0 allocs), re-benchmarked (5.5×), and fuzzed — which immediately found the `s[:-1]` panic a naive parser introduces. *That last step is the point:* a faster wrong answer is still wrong, and fuzzing is how you find out before production does.

---

## What we learned

- **Benchmarking**: `go test -bench -benchmem`; read ns/op · B/op · allocs/op; feed a `sink` so the compiler can't elide the work.
- **CPU profiling**: `runtime/pprof` + `go tool pprof`; profile first — the aggregator is I/O-bound, not parse-bound.
- **Heap/alloc + escape analysis**: `-memprofile`, `alloc_space` vs `inuse_space`, `-gcflags=-m`; string slicing allocates nothing, `FindStringSubmatch` allocates every call.
- **`net/http/pprof`**: live profiles at `/debug/pprof/`; bind to localhost, never the public interface.
- **Fuzzing**: `f.Add` + `f.Fuzz` + `-fuzz`; use an oracle + a deterministic property; it caught the optimization's panic in under a second.

---

## Up next

That's Phase 3 — **Concurrency & Systems** — done. You can now write, reason about, and *measure* concurrent Go. **Phase 4 — Production & Distributed** starts at Lesson 23: HTTP servers with `net/http` and structured logging with `slog`. We leave the aggregator here, measured and hardened, and start building services the outside world talks to.
