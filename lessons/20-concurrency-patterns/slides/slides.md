<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">20</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 3 — Concurrency &amp; Systems</div>
<h1>Concurrency patterns</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>The "named patterns" lesson. Learn the canonical Go concurrency patterns: <strong>worker pool</strong> (bounded parallelism), <strong>fan-out</strong> (distribute work), <strong>fan-in</strong> (multiplex), <strong>pipeline</strong> (chained stages), and <strong>errgroup</strong> (first-error-wins + ctx cancellation). The aggregator gains a third implementation — <code>WalkPool</code> — joining <code>Walk</code> and <code>WalkLocked</code>. The CLI gets <code>-mode=channel|mutex|pool</code> + <code>-workers=N</code> to compare all three at runtime.</p>
</div>
</div>
</div>

---

## What we'll cover

- **Worker pool** — bounded parallelism. N workers consuming from a shared queue.
- **Fan-out** — distribute work to N consumers from a single producer.
- **Fan-in** — multiplex N producers into a single consumer.
- **Pipeline** — chained stages, each connected by channels.
- **`errgroupx`** — the canonical "manage N goroutines + first-error-wins + ctx cancellation" library.

---

## The story so far

L16 introduced goroutines + channels. L17 added select + timers. L18 added sync primitives + the race detector. L19 wrapped cancellation into the `context.Context` interface.

L20 is the **patterns lesson**. You have all the primitives now; today you learn the named CONFIGURATIONS those primitives commonly take. Each pattern has a name, a shape, and a common-mistake to avoid.

The aggregator running example gets a third implementation — `WalkPool` — alongside the L16-L19 `Walk` and the L18 `WalkLocked`. By the end of the lesson, students can compare all three approaches to the same problem at runtime via a CLI flag.

---

## Concept 1: Worker pool

### Motivation

So far the aggregator has spawned **one goroutine per file**. For 10 files that's fine. For 10,000 files it's wasteful — 10,000 goroutines all contending for the OS's IO queue, each holding a few KB of stack.

The **worker pool** pattern bounds the parallelism. Instead of spawning one goroutine per item, spawn N (typically `runtime.NumCPU()` for CPU-bound work, or higher for I/O-bound work) and feed them work through a shared channel.

---

### The basics

The pattern:

```go
pathsCh := make(chan string, len(files))
for _, p := range files {
    pathsCh <- p
}
close(pathsCh)    // signal: no more work coming

resultsCh := make(chan Result, len(files))

for range N {
    go func() {
        for path := range pathsCh {
            resultsCh <- process(path)
        }
    }()
}
```

Three things to internalise:

- **Producer fills the channel and closes it.** Workers `range` over the channel; the close signals "no more work."
- **Exactly N goroutines exist.** Not "up to N" — exactly N. Worker count is the tuning knob.
- **Workers don't know about each other.** Each just pulls from the queue. Coordination is implicit through the channel.

---

### A worked example

The L20 `WalkPool` (simplified):

```go
func WalkPool(ctx context.Context, dir string, timeout time.Duration, workers int) (WalkResult, error) {
    if workers <= 0 { workers = runtime.NumCPU() }

    files := listFiles(dir)
    pathsCh := make(chan string, len(files))
    for _, p := range files { pathsCh <- p }
    close(pathsCh)

    resultsCh := make(chan fileResult, len(files))

    g, ctx := errgroupx.WithContext(ctx)
    for range workers {
        g.Go(func() error {
            for path := range pathsCh {
                r := process(path)    // per-file timeout etc.
                if r.err != nil { return r.err }
                resultsCh <- r
            }
            return nil
        })
    }

    go func() { _ = g.Wait(); close(resultsCh) }()

    return reduce(resultsCh), g.Wait()
}
```

Bounded parallelism via N workers. errgroupx handles first-error-cancels-rest (concept 5).

---

### Common mistake

**"More workers = faster."** Wrong. Adding workers past a saturation point (typically `NumCPU()` for CPU-bound work, or "I/O parallelism limit" for I/O-bound) doesn't help and starts hurting:

- CPU-bound: workers fight for cores, context-switch overhead dominates
- I/O-bound: workers fight for disk seeks / network connections / file descriptors

**Profile before tuning.** A worker pool of 1000 for processing local files is almost always worse than `NumCPU()`.

---

### Recap

- Worker pool = bounded parallelism via N goroutines + shared input channel.
- Producer closes the channel; workers exit when range loop ends.
- Tune `workers` based on whether work is CPU-bound (NumCPU) or I/O-bound (higher).
- More workers ≠ faster past saturation.

---

## Concept 2: Fan-out

### Motivation

The worker pool's **input side** — one producer, N consumers reading from the same channel. This is **fan-out**: work is distributed (fanned out) to multiple workers.

Fan-out alone, without a pool, means "spawn a goroutine per item." Bounded fan-out (= worker pool) is the safer default.

---

### The basics

Unbounded fan-out:

```go
for _, item := range items {
    go process(item)    // one goroutine per item
}
```

Bounded fan-out (= worker pool):

```go
itemsCh := make(chan Item, len(items))
for _, item := range items { itemsCh <- item }
close(itemsCh)

for range workers {
    go func() {
        for item := range itemsCh {
            process(item)
        }
    }()
}
```

The difference: unbounded fan-out spawns `len(items)` goroutines; bounded fan-out spawns `workers` goroutines. For small N, identical. For large N, bounded protects you.

---

### A worked example

The L20 `WalkPool` IS bounded fan-out. The L16 `Walk` was unbounded fan-out:

```go
// L16 Walk: unbounded fan-out
for _, path := range files {
    go processFile(path, resultsCh)
}
// 10,000 files → 10,000 goroutines

// L20 WalkPool: bounded fan-out (a worker pool)
for range workers {
    go func() {
        for path := range pathsCh {
            resultsCh <- processFile(path)
        }
    }()
}
// 10,000 files → `workers` goroutines (typically 4-16)
```

Both produce identical results. The bounded version has predictable resource usage.

---

### Common mistake

**Unbounded fan-out in production code.** A web server that spawns one goroutine per request without limits will eventually hit the OS's process/thread limits or run out of memory. Same for batch processing one goroutine per file when the file count comes from user input.

The fix is always the same: introduce a worker pool. Bounded fan-out is the canonical pattern.

---

### Recap

- Fan-out = single producer, N consumers reading the same channel.
- Unbounded fan-out (one goroutine per item) is fragile at scale.
- Bounded fan-out IS a worker pool.
- Always bound fan-out in production.

---

## Concept 3: Fan-in

### Motivation

The mirror of fan-out: **N producers, single consumer**. Many goroutines each producing values on their own channel; a single channel collects them all for downstream consumption.

The L20 warmup `FanIn[T any](chans ...<-chan T) <-chan T` is exactly this pattern.

---

### The basics

The shape:

```go
func FanIn[T any](chans ...<-chan T) <-chan T {
    out := make(chan T)
    var wg sync.WaitGroup

    for _, ch := range chans {
        wg.Add(1)
        go func(input <-chan T) {
            defer wg.Done()
            for v := range input {
                out <- v
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}
```

Two pieces:

- **Forwarder goroutines** — one per input, each copies values to the shared output.
- **Closer goroutine** — waits for all forwarders to finish, then closes the output. Without this, the output never closes and downstream consumers block forever.

---

### A worked example

The warmup test:

```go
a := make(chan int, 3); a <- 1; a <- 2; a <- 3; close(a)
b := make(chan int, 3); b <- 4; b <- 5; b <- 6; close(b)

out := FanIn[int](a, b)
for v := range out {
    fmt.Println(v)  // 1..6 in some order
}
// range exits when a AND b are both closed
```

The output order is **non-deterministic** — depends on goroutine scheduling. If order matters, fan-in is the wrong pattern; use a pipeline (concept 4) or a single producer.

---

### Common mistake

**Forgetting the closer goroutine.** Without it, `out` is never closed, and the downstream `for v := range out` blocks forever:

```go
// the wrong way
func FanInBroken[T any](chans ...<-chan T) <-chan T {
    out := make(chan T)
    for _, ch := range chans {
        go func(input <-chan T) {
            for v := range input {
                out <- v
            }
        }(ch)
    }
    return out    // ❌ no closer
}
```

The forwarders exit when their inputs close, but `out` stays open because nothing calls `close(out)`. The fix is the closer goroutine — `go func() { wg.Wait(); close(out) }()`.

---

### Recap

- Fan-in = N producers, single consumer; merge channels via forwarders + WaitGroup + closer goroutine.
- Output order is non-deterministic.
- **Closer goroutine is essential** — without it, downstream `range` blocks forever.

---

## Concept 4: Pipeline

### Motivation

Stages chained by channels. Each stage takes input from the previous one, transforms, and emits to the next. The classic shape:

```
[source] → [stage1] → [stage2] → [stage3] → [sink]
```

Pipelines are how you compose streaming work. The L13 CSV importer is a 3-stage pipeline (read → parse → save). The L18 aggregator is a 2-stage pipeline (read → reduce).

---

### The basics

Each stage is a function: takes input channel(s), returns output channel(s):

```go
func parse(in <-chan []byte) <-chan Record {
    out := make(chan Record)
    go func() {
        defer close(out)
        for raw := range in {
            out <- parseOne(raw)
        }
    }()
    return out
}

func filter(in <-chan Record, pred func(Record) bool) <-chan Record {
    out := make(chan Record)
    go func() {
        defer close(out)
        for r := range in {
            if pred(r) {
                out <- r
            }
        }
    }()
    return out
}

// Compose:
records := parse(read("data.csv"))
filtered := filter(records, isValid)
for r := range filtered {
    save(r)
}
```

Three things to internalise:

- **Each stage spawns its own goroutine** and returns the output channel immediately.
- **Each stage `defer close(out)`** — when the stage's `range in` ends, the close propagates downstream so consumers know to stop.
- **Stages are independent** — easy to test in isolation; easy to compose; easy to add/remove stages.

---

### A worked example

A read-CSV → parse → filter → write pipeline:

```go
raws := readLines("data.csv")              // <-chan []byte
records := parse(raws)                      // <-chan Record
highValue := filter(records, isHighValue)   // <-chan Record
for r := range highValue {
    save(r)
}
```

Each stage runs in its own goroutine; they all execute concurrently. Producer/consumer rates differ — buffered channels between stages absorb the variance.

---

### Common mistake

**Not propagating ctx through pipeline stages.** Each stage should accept ctx and check `ctx.Done()`. Otherwise cancelling the top doesn't stop the bottom:

```go
// the wrong way
func parse(in <-chan []byte) <-chan Record {
    out := make(chan Record)
    go func() {
        defer close(out)
        for raw := range in {
            out <- parseOne(raw)    // ❌ blocks on out forever if downstream stops
        }
    }()
    return out
}

// the right way
func parse(ctx context.Context, in <-chan []byte) <-chan Record {
    out := make(chan Record)
    go func() {
        defer close(out)
        for raw := range in {
            select {
            case out <- parseOne(raw):
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}
```

Every stage takes ctx; every send through a channel happens via select with ctx.Done(). Cancellation propagates instantly.

---

### Recap

- Pipeline = stages chained by channels, each running in its own goroutine.
- Each stage `defer close(out)` — close propagates downstream.
- **Propagate ctx through every stage** — cancellation must reach the bottom.

---

## Concept 5: `errgroupx` (and `golang.org/x/sync/errgroup`)

### Motivation

You've now seen the "manage N goroutines + first-error-cancels-rest" pattern several times:

- L18 `WalkLocked` uses `atomic.Pointer[error]` + `sync.WaitGroup`
- L19 propagates ctx through the aggregator
- L20 `WalkPool` orchestrates N workers

This pattern is so common it has a canonical library: `golang.org/x/sync/errgroup`. The L20 lesson teaches a lesson-local equivalent (`internal/errgroupx`) so you see the implementation from the inside.

---

### The basics

The API:

```go
g, ctx := errgroupx.WithContext(parentCtx)

for _, item := range items {
    item := item    // capture
    g.Go(func() error {
        return process(ctx, item)   // ctx cancels if any sibling fails
    })
}

if err := g.Wait(); err != nil {
    return err    // first error from any goroutine
}
```

Three things to internalise:

- **`WithContext` returns a Group + derived ctx.** The ctx is cancelled when the first error occurs or when Wait returns.
- **`Go(func() error)` spawns** — same as `go`, but the group tracks completion + first error.
- **`Wait()` blocks** until all spawned goroutines return; returns the first non-nil error (if any).

Internally, errgroupx is ~50 lines: `sync.WaitGroup` for completion + `sync.Once` for first-error capture + `context.CancelFunc` for cancellation. The L20 main exercise has you read this implementation; in real code, you import the real errgroup.

---

### A worked example

The L20 `WalkPool` uses errgroupx for its worker coordination:

```go
g, ctx := errgroupx.WithContext(ctx)
for range workers {
    g.Go(func() error {
        for path := range pathsCh {
            r := processFile(ctx, path)
            if r.err != nil {
                return fmt.Errorf("aggregator: %s: %w", r.path, r.err)
            }
            resultsCh <- r
        }
        return nil
    })
}

go func() {
    _ = g.Wait()
    close(resultsCh)
}()

// ... reduce resultsCh ...

return result, g.Wait()
```

The first worker to error cancels the rest via ctx. The outer goroutine wait + close completes the reducer's range loop. The final `g.Wait()` extracts the error (or nil).

---

### Common mistake

**Forgetting to call Wait.** Goroutines spawned via `g.Go(...)` continue running until they exit naturally. If you `return` from your function without calling Wait, the goroutines leak — and you never learn about any errors they encountered.

```go
// the wrong way
func doWork(ctx context.Context, items []Item) error {
    g, ctx := errgroupx.WithContext(ctx)
    for _, item := range items {
        g.Go(func() error { return process(ctx, item) })
    }
    return nil    // ❌ never called g.Wait() — goroutines run unsupervised
}
```

`g.Wait()` is mandatory. Treat errgroup the same way you treat `defer wg.Done()` — paired with creation.

---

### Recap

- `errgroupx` (mirrors `golang.org/x/sync/errgroup`) wraps the "N goroutines + first-error + cancel" pattern.
- `Go(func() error)` spawns; `Wait() error` blocks + returns first error.
- ctx is cancelled on first error or when Wait returns.
- **Always Wait** — otherwise goroutines leak.
- In production: `import "golang.org/x/sync/errgroup"`. The API is identical.

---

## Practice

### Warm-up

Implement `FanIn[T any](chans ...<-chan T) <-chan T` in `exercises/warmup/fanin/fanin.go`. Spawn a forwarder goroutine per input + a closer goroutine that waits for all forwarders via `sync.WaitGroup` then closes the output.

```bash
cd lessons/20-concurrency-patterns/exercises/warmup/fanin
go test -v
go test -race
```

---

### `internal/errgroupx/`

Implement the lesson-local errgroup in `exercises/internal/errgroupx/errgroupx.go`. Three methods: `WithContext`, `Go`, `Wait`. Uses `sync.WaitGroup` + `sync.Once` + `context.WithCancelCause`. ~30 lines of body.

```bash
cd lessons/20-concurrency-patterns/exercises/internal/errgroupx
go test -v
```

---

### Main — aggregator gains WalkPool

Add `WalkPool(ctx, dir, timeout, workers)` to `internal/aggregator/aggregator.go`. Uses `errgroupx` for the worker pool; `runtime.NumCPU()` default when `workers=0`. Same `WalkResult` / same semantics as `Walk` and `WalkLocked`.

Add `-mode=pool` + `-workers=N` flags to `cmd/aggregator/main.go`.

```bash
cd lessons/20-concurrency-patterns/exercises
go test ./...
go test -race ./...
make test-race
```

Manual smoke test (all three modes produce identical output):

```bash
mkdir /tmp/logs
echo "2026-05-21T14:30:00 INFO ok" > /tmp/logs/a.log
echo "2026-05-21T14:31:00 WARN slow" > /tmp/logs/b.log

go run ./cmd/aggregator -dir=/tmp/logs                         # channel (default)
go run ./cmd/aggregator -dir=/tmp/logs -mode=mutex             # mutex
go run ./cmd/aggregator -dir=/tmp/logs -mode=pool              # pool, default workers
go run ./cmd/aggregator -dir=/tmp/logs -mode=pool -workers=8   # pool, 8 workers
```

Note:
The benchmark comparison (which mode is faster) belongs to L22 (profiling). For now: observe that they all produce the same answer. The differences are about resource usage and code structure.

---

## Closing thought

You now have the canonical Go concurrency vocabulary:

- **Worker pool** — bounded parallelism (this lesson)
- **Fan-out / fan-in** — distribute / collect (this lesson)
- **Pipeline** — chained stages (this lesson)
- **errgroup** — manage N goroutines (this lesson)
- **Mutex / WaitGroup / atomic / Once** — sync primitives (L18)
- **Context** — cancellation propagation (L19)
- **Select / timers** — multiplexing (L17)
- **Goroutines / channels** — primitives (L16)

Real Go concurrent programs are compositions of these patterns. The L20 WalkPool is fan-out + worker pool + errgroup + per-file timeout + ctx propagation — five patterns in ~80 lines of business logic.

L21 (networking) builds on these patterns for TCP servers. L22 (profiling) measures them.

---

## What we learned

- **Worker pool**: bounded parallelism via N goroutines + shared input channel.
- **Fan-out**: distribute work to N consumers. Always bound in production.
- **Fan-in**: multiplex N producers; **closer goroutine essential**.
- **Pipeline**: chained stages; propagate ctx through every stage.
- **`errgroupx`**: canonical "manage N goroutines + first-error + cancel"; always Wait.

---

## Up next

Lesson 21 — **Networking & syscalls**. TCP listen/accept loop, signals, file descriptors. We'll build a tiny line-echo server + client — standalone this time; the aggregator stays where L20 left it.
