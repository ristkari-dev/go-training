# Lesson 18: sync & memory model

## What you'll learn

By the end of this lesson you can:

- Use `sync.Mutex` to protect shared state, with `defer mu.Unlock()` as the discipline.
- Reach for `sync.RWMutex` when reads dominate (and recognise when plain Mutex is faster).
- Use `sync.WaitGroup` to wait for N goroutines: `Add` before, `defer Done()`, `Wait()` at the barrier.
- Use `sync.Once` for lazy initialization and `sync/atomic` for lock-free single-word ops (preferring `atomic.Int64` / `atomic.Pointer[T]` over the function-style API).
- Decide deliberately between "share by communicating" (channels) and "share by locking" (mutex) — both are first-class, neither is universally "better."
- Run `make test-race` as a daily habit and read its data-race reports.

## What's different from L17

L17 added per-file timeout via `select` + `time.After`. **L18 adds a second aggregator implementation: `WalkLocked`** — same input/output as L17's `Walk`, but coordinated with `sync.Mutex` + `sync.WaitGroup` instead of channels. Both live in the same package; the CLI gains `-mode=channel|mutex` to pick at runtime.

This is also the lesson where **`make test-race`** becomes part of your daily habits. Phase 3+ lessons run it alongside `make test`; CI for these lessons runs both.

## The package layout

```
lessons/18-sync-memory-model/exercises/
├── warmup/counter/                ← you implement: Counter (mutex-protected)
│   ├── counter.go                  Counter (TODO) + CounterUnsafe (provided)
│   └── counter_test.go             skeleton TestCounter
├── cmd/aggregator/                 ← evolves: -mode flag
│   ├── main.go
│   └── main_test.go                + TestAggregatorMode + TestAggregatorInvalidMode
└── internal/
    ├── logparse/                    verbatim from L17
    └── aggregator/                  ← evolves: WalkLocked alongside Walk
        ├── aggregator.go            Walk (channels) + WalkLocked (mutex)
        └── aggregator_test.go       parameterized via runWalkSuite
```

---

## Concept 1 — `sync.Mutex` + memory model intuition

The canonical pattern:

```go
var mu sync.Mutex
var balance int

func deposit(amount int) {
    mu.Lock()
    defer mu.Unlock()
    balance += amount
}
```

Three things to internalise:

- **Zero value is usable.** No `NewMutex()`; the zero `sync.Mutex` is unlocked and ready.
- **`defer mu.Unlock()` pairs the unlock with the lock** so they can't get out of sync.
- **Mutex protects DATA, not code.** Every code path that touches the data must acquire the same mutex.

### Memory model in one paragraph

Go's memory model is built on **happens-before**. When goroutine A does `mu.Unlock()` and goroutine B then does `mu.Lock()` on the same mutex, every memory write A made before unlock is visible to B after lock. Channels (send happens-before receive), WaitGroup (`Done` happens-before `Wait` returning), and atomic ops on the same address also create happens-before edges. The runtime guarantees visibility across those edges. That's the whole memory model.

### Common mistake

**Copying a Mutex.** `sync.Mutex` contains state; a copy gives you TWO independent mutexes protecting nothing. `go vet` catches this via the `copylocks` analyzer:

```
assignment copies lock value to c2: sync.Mutex
```

Always pass mutex-bearing types by pointer (`func (c *Counter) Inc()`), never by value.

---

## Concept 2 — `sync.RWMutex`

For READ-HEAVY shared state, plain Mutex serializes even concurrent reads. `sync.RWMutex` splits the lock: many readers concurrently OR one writer exclusively.

```go
var mu sync.RWMutex
var cache map[string]string

func get(k string) (string, bool) {
    mu.RLock()
    defer mu.RUnlock()
    v, ok := cache[k]
    return v, ok
}

func set(k, v string) {
    mu.Lock()
    defer mu.Unlock()
    cache[k] = v
}
```

When to reach for it:

- Reads outnumber writes by 10× or more
- Reads are non-trivial (not a single int load)
- Profile shows Mutex contention is the bottleneck

### Common mistake

Reaching for RWMutex when reads aren't truly hot. RWMutex has more bookkeeping than Mutex; for a single-int load with low contention, plain Mutex is faster. **Below ~10 concurrent readers, default to Mutex.** Upgrade later if benchmarks show contention.

---

## Concept 3 — `sync.WaitGroup`

"Spawn N goroutines, wait for all of them." The canonical fan-out-and-wait pattern.

```go
var wg sync.WaitGroup

for _, item := range items {
    wg.Add(1)                        // BEFORE the goroutine, in the parent
    go func(it Item) {
        defer wg.Done()              // INSIDE the goroutine
        process(it)
    }(item)
}

wg.Wait()                             // barrier
```

Rules:

- **`Add(n)` BEFORE the goroutine, in the parent.** Calling Add inside the goroutine races with Wait.
- **`defer wg.Done()` INSIDE the goroutine.** Guarantees Done runs even if the goroutine panics or returns early.
- **`Wait()` is the barrier.** Returns when the counter reaches 0.

### Common mistake

**`Add(1)` inside the goroutine.** The race: main might reach `wg.Wait()` before all goroutines have called Add(1); Wait sees counter=0 and returns; main continues prematurely. Always call Add from the parent.

The other classic mistake: forgetting `defer wg.Done()` and using a bare `wg.Done()` at the end. If the goroutine panics or returns early, Done is skipped and `Wait()` hangs forever.

---

## Concept 4 — `sync.Once` + `sync/atomic`

Two narrow but real primitives.

**`sync.Once`** — run a function exactly once across all goroutines:

```go
var (
    once   sync.Once
    config *Config
)

func loadConfig() *Config {
    once.Do(func() {
        config = readConfigFromDisk()
    })
    return config
}
```

Concurrent callers block until the first call's `fn` returns; subsequent calls return immediately.

**`sync/atomic`** — lock-free reads/writes on single-word types:

```go
var requestCount atomic.Int64    // generic type, Go 1.19+ — prefer this

func handle() {
    requestCount.Add(1)
}

func getCount() int64 {
    return requestCount.Load()
}
```

The function-style API (`atomic.AddInt64(&counter, 1)`) also works but the generic type is harder to misuse — you can't accidentally do a non-atomic load on it.

### Common mistake

**Mixing atomic and non-atomic access on the same memory:**

```go
atomic.AddInt64(&counter, 1)   // in one goroutine
v := counter                    // in another — non-atomic read, RACES
```

If you use atomic to write, use atomic to read. The generic types prevent this entirely.

For `sync.Once`: don't roll your own with `bool initialized` — it races. The whole point of Once is the race-free flag check.

---

## Concept 5 — Share by communicating vs share by locking

The lesson's centerpiece. You now have TWO ways to coordinate goroutines. The L18 aggregator demonstrates both — same input, same output, different internals:

**`Walk` (channels, L17):**
```go
resultsCh := make(chan fileResult, len(files))
for _, path := range files { go processFile(path, resultsCh) }
for i := 0; i < len(files); i++ { /* reduce */ }
```

**`WalkLocked` (mutex + WaitGroup, L18):**
```go
var mu sync.Mutex
var wg sync.WaitGroup
for _, path := range files {
    wg.Add(1)
    go func(p string) { defer wg.Done(); /* parse + mu.Lock + merge + mu.Unlock */ }(path)
}
wg.Wait()
```

Same `WalkResult` out. Pick via `-mode=channel|mutex`.

When to use which:

| Situation | Reach for |
|---|---|
| Pipelined work, multiple stages | Channels |
| Cancellation, timeouts, deadlines | Channels (select) |
| Producer-consumer with backpressure | Buffered channels |
| Simple shared counter | atomic |
| Simple shared map (low contention) | Mutex |
| Read-heavy shared state | RWMutex |
| Run N goroutines, wait for all | WaitGroup |
| First-error-wins across goroutines | atomic.Pointer + WaitGroup |

### Common mistake

Assuming one style is "better." Both are correct. The community slogan "share memory by communicating" is a guideline, not a rule — real Go code uses both, sometimes in the same package. **Pick whichever is simpler for the problem.**

---

## Exercise: warm-up — `counter`

Implement `Counter` (mutex-protected) in `exercises/warmup/counter/counter.go`. `CounterUnsafe` is already there (deliberately buggy, no mutex). The unit test exercises Counter only.

To see CounterUnsafe fail under the race detector, write a quick test using the snippet in `counter.go`'s doc comment and run `go test -race`. You'll get a DATA RACE warning pointing at `c.n` in CounterUnsafe.Inc and Value.

**Time:** 10 minutes.

## Exercise: main — aggregator with WalkLocked + CLI -mode flag

Two pieces:

1. **`internal/aggregator/WalkLocked`** — same signature as `Walk(dir, timeout) (WalkResult, error)`, but uses `sync.Mutex` + `sync.WaitGroup` + `atomic.Pointer[error]` (for first-error-wins) instead of channel coordination. The existing `processFile` is unchanged; add a sibling `processFileForLocked` helper that returns `fileResult` directly.
2. **`cmd/aggregator -mode=channel|mutex`** — picks at runtime. Default is `channel` (backward-compatible with L17). Invalid mode → error.

The aggregator tests parameterize over both Walk and WalkLocked via `t.Run("Walk", ...)` / `t.Run("WalkLocked", ...)`. The CLI's `TestAggregatorMode` runs the binary three times (default, channel, mutex) and asserts identical stdout.

**Time:** 45-60 minutes.

---

## Daily habits

After every change:

```bash
gofmt -w ./...
go vet ./...
go test ./...
```

**NEW from L18 onward:**

```bash
make test-race
```

The race detector is the canonical tool for catching concurrency bugs. 5-20× slower than `make test`, but worth it before committing concurrency-touching code OR opening a PR. Phase 3+ CI runs both.

## How to run

```bash
# Skeleton (passes vacuously until you implement)
cd lessons/18-sync-memory-model/exercises
go test ./...

# Reference solution
go test -v ./lessons/18-sync-memory-model/solutions/...

# Race detector (the new daily habit)
make test-race
# or scoped to L18:
go test -race ./lessons/18-sync-memory-model/...

# The binary
mkdir /tmp/logs
echo "2026-05-21T14:30:00 INFO ok" > /tmp/logs/a.log
echo "2026-05-21T14:31:00 WARN slow" > /tmp/logs/b.log

# Default (channel mode)
go run ./lessons/18-sync-memory-model/solutions/cmd/aggregator -dir=/tmp/logs

# Mutex mode — same output
go run ./lessons/18-sync-memory-model/solutions/cmd/aggregator -dir=/tmp/logs -mode=mutex

# Combine with timeout
go run ./lessons/18-sync-memory-model/solutions/cmd/aggregator -dir=/tmp/logs -mode=mutex -timeout=1ns
# stderr: "timeout: /tmp/logs/a.log\ntimeout: /tmp/logs/b.log"
```

## Going further

### Read

- **The Go Memory Model** (formal spec): <https://go.dev/ref/mem> — short, dense, worth one careful read.
- **Dmitry Vyukov — "Data Race Detector"** (Go blog): <https://go.dev/blog/race-detector> — how `-race` works internally.
- **Effective Go — "Concurrency"** (revisit): <https://go.dev/doc/effective_go#concurrency>
- **Russ Cox — "Hardware Memory Models" + "Programming Language Memory Models"** (2021): <https://research.swtch.com/hwmm> — deep but illuminating background.

### Try

- **Benchmark Walk vs WalkLocked.** Add a `BenchmarkWalk` and `BenchmarkWalkLocked` over a tempdir of 1000 synthetic log files. Compare ns/op + allocs/op. Which wins on your machine? Why? (L20 preview: worker pools win for large N — see L22 for profiling.)
- **Sabotage WalkLocked.** Remove `mu.Lock()` / `mu.Unlock()` around the map merge. Run `go test -race ./lessons/18-sync-memory-model/solutions/internal/aggregator/...` — you'll see a DATA RACE report. Restore the mutex; verify it passes. Internalise the race detector's output format.
- **Switch to `atomic.Int64`** for the L18 warmup. Rewrite `Counter` using `atomic.Int64` instead of Mutex. Compare: which is shorter? Faster? More readable?

---

> Next stop: Lesson 19 — `context`. The done-channel cancellation pattern becomes a standard library interface. The aggregator gets `ctx context.Context` propagation; the CLI's SIGINT handler cancels mid-walk gracefully.
