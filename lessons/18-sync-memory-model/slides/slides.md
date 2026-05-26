<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">18</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 3 — Concurrency &amp; Systems</div>
<h1>sync &amp; memory model</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Learn the <code>sync</code> package (Mutex, RWMutex, WaitGroup, Once), <code>sync/atomic</code>, and the race detector formally. The aggregator gains a second implementation — <code>WalkLocked</code> with mutex + WaitGroup — alongside L17's channel-based <code>Walk</code>. Concept 5 compares "share by communicating" with "share by locking" side-by-side.</p>
</div>
</div>
</div>

---

## What we'll cover

- **`sync.Mutex` + memory model intuition** — `mu.Lock()` / `defer mu.Unlock()`; happens-before.
- **`sync.RWMutex`** — many readers OR one writer.
- **`sync.WaitGroup`** — count-down barrier for "wait for N goroutines."
- **`sync.Once` + `sync/atomic`** — lazy init + lock-free single-word ops.
- **Share by communicating vs share by locking** — Walk vs WalkLocked, side by side.

---

## Welcome to L18

L16 introduced goroutines + channels. L17 added select + timers. L18 adds the **other** half of Go's concurrency toolkit: explicit synchronization via the `sync` package.

The slogan from L16 ("share memory by communicating, not communicate by sharing memory") is a guideline, not a law. Real Go code uses BOTH approaches. Mutexes are first-class. The race detector catches bugs in either style.

This lesson is also where **`make test-race`** becomes part of your daily habits.

---

## Concept 1: `sync.Mutex` + memory model intuition

### Motivation

When multiple goroutines touch the same memory, you need synchronization — otherwise the Go runtime makes no guarantees about WHICH write wins, or whether reads see consistent state. The race detector flags any such code as "DATA RACE" and tells you exactly which goroutines and which memory location.

`sync.Mutex` is the canonical fix: hold the lock to access shared state; release when done.

---

### The basics

```go
var mu sync.Mutex
var balance int

func deposit(amount int) {
    mu.Lock()
    defer mu.Unlock()
    balance += amount
}

func balanceOf() int {
    mu.Lock()
    defer mu.Unlock()
    return balance
}
```

Three things to internalise:

- **Zero value is usable.** No `NewMutex()`. The zero-value `sync.Mutex` is an unlocked mutex ready to use.
- **`defer mu.Unlock()` is the idiom.** Pairs the unlock with the lock so they can't get out of sync.
- **Mutex protects DATA, not code.** The mutex says "while I hold this lock, only I access these fields." Other code paths must also acquire the same mutex.

### Memory model in one paragraph

Go's memory model is built around **happens-before**. When goroutine A does `mu.Unlock()` and goroutine B then does `mu.Lock()` on the same mutex, everything A wrote before unlock is **visible** to B after lock. Channels work the same way: a send happens-before the corresponding receive. WaitGroup.Done() happens-before the matching Wait() return. That's the whole memory model — channels and sync primitives create happens-before edges, and the runtime guarantees visibility across those edges.

---

### A worked example

The L18 `WalkLocked` aggregator uses Mutex on a shared map:

```go
var mu sync.Mutex
result := WalkResult{Counts: map[string]int{}}

for _, path := range files {
    wg.Add(1)
    go func(p string) {
        defer wg.Done()
        // ... parse the file into per-file counts ...

        mu.Lock()
        for level, count := range counts {
            result.Counts[level] += count
        }
        mu.Unlock()
    }(path)
}

wg.Wait()
// After Wait, happens-before guarantees all worker writes are visible.
// Safe to read result.Counts WITHOUT holding the mutex now.
```

The `wg.Wait()` is itself a happens-before edge: every `wg.Done()` happens-before the return from Wait(). So once main passes Wait, it sees everything every worker wrote.

---

### Common mistake

**Copying a Mutex.**

```go
// the wrong way
type Counter struct {
    mu sync.Mutex
    n  int
}

c := Counter{}
c2 := c   // ❌ copies the Mutex's state
```

`sync.Mutex` is NOT safe to copy. The struct contains internal state (a waiter list, a lock bit, etc.); copying creates two independent mutexes that protect different sets of memory. `go vet` catches this with the `copylocks` analyzer:

```
assignment copies lock value to c2: sync.Mutex
```

Always pass Mutex-bearing types by **pointer**: `func (c *Counter) Inc()`, not `func (c Counter) Inc()`.

---

### Recap

- `mu.Lock()` + `defer mu.Unlock()` — the canonical pattern.
- Zero value is usable; no constructor.
- Mutex protects DATA — every code path touching that data must lock.
- **Don't copy mutex-bearing values** — `go vet` catches it.
- Memory model: mutex/channel ops create happens-before edges.

---

## Concept 2: `sync.RWMutex`

### Motivation

For READ-HEAVY workloads, a Mutex serializes everything — even concurrent reads block each other. `sync.RWMutex` splits the lock: many readers can hold it simultaneously OR one writer can hold it exclusively.

Use this when reads dominate (10×+ more than writes) and reads are non-trivial (more than a single load).

---

### The basics

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

- `RLock()` / `RUnlock()` — readers; multiple can hold simultaneously.
- `Lock()` / `Unlock()` — writers; exclusive.
- Writers block until all readers release; readers block until any active writer releases.

---

### A worked example

A read-heavy cache:

```go
type Cache struct {
    mu sync.RWMutex
    m  map[string]int
}

func (c *Cache) Get(k string) int {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.m[k]
}

func (c *Cache) Set(k string, v int) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.m[k] = v
}
```

If you have 1000 readers per second and 1 writer per second, RWMutex lets the 1000 reads run concurrently. With a plain Mutex they'd serialize.

---

### Common mistake

Using RWMutex when reads aren't truly hot:

```go
// the wrong way (for low read traffic)
type Counter struct {
    mu sync.RWMutex
    n  int
}

func (c *Counter) Value() int {
    c.mu.RLock(); defer c.mu.RUnlock()
    return c.n
}
```

RWMutex has overhead: more bookkeeping than a plain Mutex. For a single-int load, the RWMutex's overhead exceeds the parallelism benefit. **Below ~10 concurrent readers, plain Mutex is usually faster.**

Rule of thumb: profile first. If you're not sure whether reads are hot enough to justify RWMutex, use Mutex. Upgrade later if benchmarks show contention.

---

### Recap

- `sync.RWMutex` — many readers OR one writer.
- `RLock`/`RUnlock` for readers; `Lock`/`Unlock` for writers.
- Worth using when reads ≫ writes AND reads are non-trivial.
- Don't reach for it by default — plain Mutex is faster for low-contention cases.

---

## Concept 3: `sync.WaitGroup`

### Motivation

"Spawn N goroutines, wait for all of them to finish, then continue." This is the most common multi-goroutine coordination pattern. `sync.WaitGroup` is the idiomatic tool.

You met WaitGroup briefly in L16 (mentioned in the worker pool slide). L18 makes it official.

---

### The basics

```go
var wg sync.WaitGroup

for _, item := range items {
    wg.Add(1)                       // BEFORE the goroutine
    go func(item Item) {
        defer wg.Done()             // INSIDE the goroutine
        process(item)
    }(item)
}

wg.Wait()                            // block until counter reaches 0
```

Three things to internalise:

- **`Add(n)` BEFORE the goroutine.** Always call Add from the parent goroutine, before `go func() { ... }()`. Calling Add inside the goroutine races with Wait.
- **`Done()` INSIDE the goroutine.** Equivalent to `Add(-1)`. The `defer wg.Done()` idiom guarantees Done runs even if the goroutine panics or returns early.
- **`Wait()` is the barrier.** Blocks until the counter reaches 0. Returns immediately if Add was never called.

---

### A worked example

The L18 WalkLocked aggregator:

```go
var wg sync.WaitGroup

for _, path := range files {
    wg.Add(1)
    go func(p string) {
        defer wg.Done()
        // ... process p, update result via mu ...
    }(path)
}

wg.Wait()
// All workers finished. Safe to read result without the mutex
// (happens-before guaranteed by WaitGroup.Wait).
```

The shape is universal for "fan-out and wait." You'll see it in worker pools (L20), HTTP handlers spawning sub-requests, anywhere you need parallel work with a synchronization point.

---

### Common mistake

**`Add(1)` inside the goroutine.**

```go
// the wrong way
for _, item := range items {
    go func(item Item) {
        wg.Add(1)             // ❌ races with wg.Wait()
        defer wg.Done()
        process(item)
    }(item)
}
wg.Wait()
```

The race: main might reach `wg.Wait()` BEFORE all goroutines have called `Add(1)`. Wait sees counter=0 (or partial), returns immediately, main continues, then the late goroutines call Add(1)... too late.

**Always call Add from the parent, before spawning the goroutine.**

The other classic mistake: forgetting `defer wg.Done()` and using a plain `wg.Done()` at the end. If the goroutine panics or returns early, Done is skipped → Wait hangs forever.

---

### Recap

- `wg.Add(n)` BEFORE the goroutine (in the parent).
- `defer wg.Done()` INSIDE the goroutine.
- `wg.Wait()` is the barrier.
- Add-before-goroutine + defer-Done is non-negotiable.

---

## Concept 4: `sync.Once` + `sync/atomic`

### Motivation

Two more sync primitives, both for narrow but real use cases:

- **`sync.Once`** — "run this function exactly once across all goroutines." Lazy initialization without race conditions.
- **`sync/atomic`** — lock-free reads and writes on single-word types (int32, int64, uintptr, pointers). Faster than a mutex when you only need a single load or store.

---

### The basics

**`sync.Once`:**

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

`once.Do(fn)` guarantees `fn` runs exactly ONE TIME no matter how many goroutines call `loadConfig` concurrently. The first caller actually runs fn; everyone else blocks until fn returns, then proceeds. After that, all calls to `once.Do(fn)` return immediately without running fn.

**`sync/atomic`:**

```go
import "sync/atomic"

var requestCount int64

func handle() {
    atomic.AddInt64(&requestCount, 1)
    // ... handle request ...
}

func getCount() int64 {
    return atomic.LoadInt64(&requestCount)
}
```

`atomic.AddInt64` and `atomic.LoadInt64` operate on a `*int64` without locking. They're faster than Mutex for single-word operations, and the race detector understands them (concurrent atomic ops are NOT a race).

Modern alternative (Go 1.19+):

```go
var requestCount atomic.Int64

func handle() {
    requestCount.Add(1)
}

func getCount() int64 {
    return requestCount.Load()
}
```

The generic types (`atomic.Int64`, `atomic.Int32`, `atomic.Pointer[T]`, etc.) are cleaner — you can't accidentally do a non-atomic load on a value that's meant to be atomic.

---

### A worked example

The L18 WalkLocked uses `atomic.Pointer[error]` to capture the first error from any worker:

```go
var firstErr atomic.Pointer[error]

// In each worker goroutine:
if r.err != nil {
    wrapped := fmt.Errorf("aggregator: %s: %w", r.path, r.err)
    firstErr.CompareAndSwap(nil, &wrapped)
    return
}

// After wg.Wait():
if perr := firstErr.Load(); perr != nil {
    return result, *perr
}
```

`CompareAndSwap(nil, &wrapped)` atomically sets firstErr to `&wrapped` IF it's currently nil. Only the first worker to fail actually sets the pointer; subsequent failures are dropped. No mutex needed for this one piece of state.

---

### Common mistake

**Mixing atomic and non-atomic access on the same memory.**

```go
// the wrong way
var counter int64

go func() {
    atomic.AddInt64(&counter, 1)
}()

go func() {
    v := counter           // ❌ non-atomic read of atomic counter
    fmt.Println(v)
}()
```

The atomic Add provides synchronization with other ATOMIC operations on the same memory. A plain read might see a torn or stale value. **If you use atomic to write, use atomic to read.** The generic types (`atomic.Int64.Load()`) prevent this mistake.

For `sync.Once`: don't try to "skip Once" by adding your own boolean flag. The whole point of Once is that the flag check + initialization are race-free; rolling your own with `bool initialized` is a classic data race.

---

### Recap

- **`sync.Once`**: run a function exactly once across all goroutines. Lazy init pattern.
- **`sync/atomic`**: lock-free reads/writes on single-word types.
- **Modern**: prefer `atomic.Int64`, `atomic.Pointer[T]` over the function-style API.
- **Don't mix** atomic and non-atomic access on the same memory.

---

## Concept 5: Share by communicating vs share by locking

### Motivation

This is the lesson's centerpiece. You now have TWO ways to coordinate goroutines: channels (L16-17) and shared state with locks (L18). Both are first-class in Go. Neither is "better." The community's slogan — "share memory by communicating, not communicate by sharing memory" — is a guideline, not a rule.

The L18 aggregator has BOTH implementations side-by-side: `Walk` (channels) and `WalkLocked` (mutex + WaitGroup). Same input, same output, different internals.

---

### The basics

**`Walk` (channel-based):**

```go
resultsCh := make(chan fileResult, len(files))
for _, path := range files {
    go processFile(path, resultsCh)
}
for i := 0; i < len(files); i++ {
    r := <-resultsCh
    // merge r.counts into result.Counts
}
```

Each worker sends a typed message; main reduces. The channel provides both data transfer and synchronization.

**`WalkLocked` (mutex + WaitGroup):**

```go
var mu sync.Mutex
var wg sync.WaitGroup
result := WalkResult{Counts: map[string]int{}}

for _, path := range files {
    wg.Add(1)
    go func(p string) {
        defer wg.Done()
        // ... parse p into per-file counts ...
        mu.Lock()
        for level, count := range counts {
            result.Counts[level] += count
        }
        mu.Unlock()
    }(path)
}
wg.Wait()
```

Workers update the shared map directly under mu; WaitGroup tracks completion. The mutex provides synchronization on the shared state.

---

### A worked example

Both implementations produce identical output for the same input — the lesson's tests prove it via `t.Run("Walk", ...) / t.Run("WalkLocked", ...)` parameterization. The CLI lets students pick at runtime:

```bash
aggregator -dir=./logs                    # default: channel
aggregator -dir=./logs -mode=channel      # explicit
aggregator -dir=./logs -mode=mutex        # WalkLocked
```

Same files, same counts. Different internals.

---

### When to use which

| Situation | Reach for |
|---|---|
| Pipelined work, multiple stages | Channels |
| Cancellation, timeouts, deadlines | Channels (select) |
| Producer-consumer with backpressure | Channels (buffered) |
| Simple shared counter | atomic |
| Simple shared map (low contention) | Mutex |
| Read-heavy shared state | RWMutex |
| Run N goroutines, wait for all | WaitGroup (always) |
| First-error-wins across goroutines | atomic.Pointer + WaitGroup |

The honest answer: **whichever is simpler for the problem.** Channels feel natural when work flows between stages. Mutex feels natural when there's a single piece of shared state. Don't force one style.

---

### Common mistake

Assuming one style is "better." Both are correct. Both can have bugs. The race detector catches data races regardless of style.

Real Go code uses BOTH. The standard library's `net/http` server uses mutexes for connection state AND channels for shutdown signaling. The Go runtime itself uses both. Be flexible.

---

### Recap

- Two styles: share by communicating (channels) vs share by locking (mutex + WaitGroup).
- Both are first-class. Both are correct when used right.
- Walk + WalkLocked produce identical output — proof that the choice is style, not correctness.
- Race detector catches bugs in either style.
- The honest rule: **pick whichever is simpler for the problem.**

---

## The race detector — formally introduced

`go test -race ./...` enables Go's race detector. It instruments memory accesses; any unsynchronized read + write OR write + write on the same memory from different goroutines is flagged at runtime.

```
WARNING: DATA RACE
Read at 0x00c00001a0a8 by goroutine 7:
  ...
Previous write at 0x00c00001a0a8 by goroutine 6:
  ...
```

Cost: 5-20× runtime overhead. Worth it during development.

**`make test-race`** is now your daily-habits target:

```bash
make test         # fast smoke (no -race)
make test-race    # full sweep with race detector (5-20× slower)
```

Run `make test` constantly. Run `make test-race` before committing concurrency-touching code AND before opening a PR. CI runs both on Phase 3+ branches.

The race detector catches REAL bugs. It does NOT catch:
- Goroutine leaks
- Deadlocks
- Logic errors that happen to not race
- Races on memory that isn't accessed in this test run

But for what it does catch, it's gold. Use it.

---

## Practice

### Warm-up

Implement `Counter` (mutex-protected) in `exercises/warmup/counter/counter.go`. The `CounterUnsafe` type is already there (deliberately buggy — no mutex). Run the unit test to verify Counter works; then **manually** try a CounterUnsafe race demo:

```bash
cd lessons/18-sync-memory-model/exercises/warmup/counter
go test -v
go test -race    # Counter passes; CounterUnsafe isn't tested
```

To see CounterUnsafe fail under -race, write a quick test (see counter.go's doc comment for the snippet) and run with -race.

---

### Main

The aggregator gains a SECOND implementation:

1. **`internal/aggregator/WalkLocked`** — same input/output as `Walk`, but uses `sync.Mutex` + `sync.WaitGroup` + `atomic.Pointer[error]` instead of channel coordination.
2. **`cmd/aggregator -mode=channel|mutex`** — picks at runtime. Default is channel (backward-compatible with L17).

Both Walk and WalkLocked produce identical results for the same input. The lesson's tests prove this via parameterization.

```bash
cd lessons/18-sync-memory-model/exercises
go test ./...
go test -race ./...    # both modes race-clean
make test-race          # whole-repo, your new daily habit
```

Manual smoke test:

```bash
mkdir /tmp/logs
echo "2026-05-21T14:30:00 INFO server started" > /tmp/logs/a.log
echo "2026-05-21T14:31:00 WARN slow query" > /tmp/logs/b.log
go run ./cmd/aggregator -dir=/tmp/logs                 # channel mode
go run ./cmd/aggregator -dir=/tmp/logs -mode=mutex     # SAME output
```

Note:
The `-mode=mutex` mode demonstrates "share by locking." When you run it under `-race`, the test must still pass — which proves the mutex coordination is correct. Try removing the `mu.Lock()` call from WalkLocked and re-running with `-race`; you'll see the race detector flag the unsynchronized map writes.

---

## Closing thought

Phase 3 so far has been three lessons of building up the concurrency toolkit:

- L16: goroutines + channels (the basics).
- L17: select + timers (coordination).
- L18: sync + memory model (shared state).

You've now seen Go's full primitive set. Lesson 19 wraps cancellation into the standard `context.Context` interface — the pattern we've been building toward since L17's done-channel concept. Lesson 20 puts it all together into the worker pool pattern.

The race detector is your friend. `make test-race` is your daily habit. Trust the slogan but don't follow it blindly — both styles are correct.

---

## What we learned

- `sync.Mutex` + memory model: zero value usable; `defer mu.Unlock()`; happens-before via mutex/channel/WaitGroup.
- `sync.RWMutex` for read-heavy state; plain Mutex below ~10 concurrent readers.
- `sync.WaitGroup`: Add BEFORE goroutine, `defer Done()`, `Wait()` is the barrier.
- `sync.Once` for lazy init; `sync/atomic` for lock-free single-word ops; prefer the generic `atomic.Int64`/`atomic.Pointer[T]` form.
- Share by communicating vs share by locking: both correct, both have race-detector coverage, pick what's simpler.
- `make test-race` joins daily habits.

---

## Up next

Lesson 19 — **`context`**. The done-channel cancellation pattern from L17 becomes a standard library interface. `context.Context` carries cancellation signals, deadlines, and values through goroutine boundaries. The aggregator gets ctx propagation; the CLI's SIGINT handler cancels mid-walk gracefully.
