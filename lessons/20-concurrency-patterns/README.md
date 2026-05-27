# Lesson 20: Concurrency patterns

## What you'll learn

By the end of this lesson you can:

- Use **worker pools** to bound parallelism — N goroutines consuming from a shared input channel.
- Recognise **fan-out** (one producer → N consumers) and **fan-in** (N producers → one consumer) as primitives, and bound fan-out in production.
- Compose **pipelines** of streaming stages; propagate ctx through every stage so cancellation reaches the bottom.
- Use **`errgroupx`** (mirrors `golang.org/x/sync/errgroup`) to manage N goroutines with first-error-wins + ctx cancellation.
- Compare three implementations of the same problem side-by-side via `cmd/aggregator -mode=channel|mutex|pool`.

## What's different from L19

L18 introduced `WalkLocked` alongside `Walk` (channels vs mutex). L19 propagated ctx through both. **L20 introduces a third implementation: `WalkPool`** — bounded parallelism via a worker pool, using `errgroupx` for first-error-wins + ctx cancellation. All three Walk variants now coexist; the CLI gains `-mode=pool` + `-workers=N`.

The new `internal/errgroupx/` package is a ~50-line lesson-local errgroup mirroring `golang.org/x/sync/errgroup`'s API. Students implement it to see HOW first-error-wins + ctx cancellation are wired internally. In production code, you import the real `golang.org/x/sync/errgroup` (same API).

## The package layout

```
lessons/20-concurrency-patterns/exercises/
├── warmup/fanin/                       ← you implement: FanIn[T any]
├── cmd/aggregator/                      ← evolves: -mode=pool + -workers flags
│   ├── main.go
│   └── main_test.go                     + TestAggregatorPool
└── internal/
    ├── logparse/                        verbatim from L19
    ├── errgroupx/                       NEW: lesson-local errgroup
    │   ├── errgroupx.go
    │   └── errgroupx_test.go
    └── aggregator/                      ← evolves: WalkPool joins Walk + WalkLocked
        ├── aggregator.go
        └── aggregator_test.go           3-way parameterized via runWalkSuite
```

---

## Concept 1 — Worker pool

Bounded parallelism: spawn exactly N goroutines, feed them work through a shared channel. Replaces "one goroutine per item" (which is fragile at scale) with predictable resource usage.

```go
pathsCh := make(chan string, len(files))
for _, p := range files { pathsCh <- p }
close(pathsCh)

for range workers {
    go func() {
        for path := range pathsCh {
            process(path)
        }
    }()
}
```

Three things:

- **Producer fills + closes the channel.** Workers `range` over it; close signals "no more work."
- **Exactly N goroutines.** Worker count is the tuning knob.
- **Workers don't coordinate directly.** The channel is the coordination.

### Common mistake

"More workers = faster." Not true past saturation: CPU-bound work caps at `runtime.NumCPU()`; I/O-bound work caps at the underlying I/O parallelism (disk seeks, network connections, file descriptors). Profile before tuning.

---

## Concept 2 — Fan-out

The worker pool's **input side**: single producer, N consumers reading the same channel. Unbounded fan-out (one goroutine per item) is the fragile version; bounded fan-out (a worker pool) is the production-safe version.

```go
// Unbounded — risky for large N
for _, item := range items {
    go process(item)
}

// Bounded — a worker pool
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

The L16 `Walk` used unbounded fan-out (fine for the small fixture). L20 `WalkPool` uses bounded fan-out (production-shaped).

### Common mistake

Spawning unbounded goroutines from user input. A web server with one goroutine per request will eventually hit OS process/thread limits. Batch processing one goroutine per file from a user-supplied directory will OOM on a directory with millions of files. **Bound fan-out in production code.**

---

## Concept 3 — Fan-in

The mirror: N producers, single consumer. The L20 warmup `FanIn[T any](chans ...<-chan T) <-chan T` is exactly this pattern:

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

    go func() {        // closer goroutine
        wg.Wait()
        close(out)
    }()

    return out
}
```

Output order is **non-deterministic** — depends on goroutine scheduling. If order matters, use a single producer or a pipeline.

### Common mistake

**Forgetting the closer goroutine.** Without `go func() { wg.Wait(); close(out) }()`, the output channel never closes and downstream `for v := range out` blocks forever. The forwarders exit when their inputs close, but `out` stays open because nobody calls `close(out)`.

---

## Concept 4 — Pipeline

Stages chained by channels. Each stage takes input from the previous, transforms, emits to the next:

```
[source] → [stage1] → [stage2] → [stage3] → [sink]
```

Each stage is a function that takes input channels + returns output channels:

```go
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

Three things:

- **Each stage spawns its own goroutine** and returns the output channel immediately.
- **Each stage `defer close(out)`** — close propagates downstream so consumers know to stop.
- **Stages are independent** — easy to test, compose, add/remove.

### Common mistake

**Not propagating ctx through stages.** Without ctx-aware sends (`select { case out <- v: case <-ctx.Done(): return }`), cancelling the top of the pipeline doesn't stop the middle stages — they block forever trying to send into channels nobody reads. Every stage must accept ctx and check it on every send.

---

## Concept 5 — `errgroupx`

The canonical "manage N goroutines + first-error-wins + ctx cancellation" library. The L20 lesson includes a lesson-local implementation (`internal/errgroupx/`) so you see the pattern from the inside; in production, `import "golang.org/x/sync/errgroup"` (same API).

```go
g, ctx := errgroupx.WithContext(parentCtx)

for _, item := range items {
    item := item
    g.Go(func() error {
        return process(ctx, item)
    })
}

if err := g.Wait(); err != nil {
    return err    // first error from any goroutine
}
```

Three things:

- **`WithContext` returns Group + ctx.** The ctx is cancelled on first error or when Wait returns.
- **`Go(func() error)` spawns** — like `go` but the group tracks completion + first error.
- **`Wait() error` blocks** until all spawned goroutines return; returns the first non-nil error.

Internally: `sync.WaitGroup` for completion + `sync.Once` for first-error capture + `context.WithCancelCause` for ctx cancellation. ~50 lines total.

### Common mistake

**Forgetting to call `Wait`.** Goroutines spawned via `g.Go(...)` run until they exit naturally. If you return from your function without calling Wait, the goroutines leak and you never see their errors. **Always Wait** — treat it the same way you treat `defer wg.Done()`.

---

## Exercise: warm-up — `fanin`

Implement `FanIn[T any](chans ...<-chan T) <-chan T` in `exercises/warmup/fanin/fanin.go`. Spawn a forwarder goroutine per input + a closer goroutine using `sync.WaitGroup`. Tests cover two inputs, three inputs, empty input list, and the closer-fires-when-all-inputs-close case.

**Time:** 10-15 minutes.

## Exercise: errgroupx implementation

Implement the lesson-local errgroup in `exercises/internal/errgroupx/errgroupx.go`. Three methods: `WithContext`, `Go`, `Wait`. Uses `sync.WaitGroup` + `sync.Once` + `context.WithCancelCause`. ~30 lines of body. Tests verify success path, first-error-wins, ctx-cancelled-on-error, ctx-cancelled-on-wait.

**Time:** 15-20 minutes.

## Exercise: main — aggregator gains WalkPool

Add `WalkPool(ctx, dir, timeout, workers)` to `internal/aggregator/aggregator.go`. Uses `errgroupx` for the worker pool; `runtime.NumCPU()` default when `workers=0`. Same `WalkResult` / same semantics as `Walk` and `WalkLocked`. Then extend `cmd/aggregator/main.go` with `-mode=pool` + `-workers=N` flags.

Tests parameterize over all three Walk variants via `runWalkSuite(t, fn)` — same 9 sub-tests run against Walk, WalkLocked, and a walkPoolAdapter. Plus `TestWalkPoolWorkers` verifies the `workers` parameter is respected (workers=1/2/8/0 all produce the same output).

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
cd lessons/20-concurrency-patterns/exercises
go test ./...

# Reference solution
go test -v ./lessons/20-concurrency-patterns/solutions/...

# Race detector
go test -race ./lessons/20-concurrency-patterns/solutions/...
make test-race

# All three Walk variants produce identical output
mkdir /tmp/logs
echo "2026-05-21T14:30:00 INFO server started" > /tmp/logs/a.log
echo "2026-05-21T14:31:00 WARN slow query" > /tmp/logs/b.log

go run ./lessons/20-concurrency-patterns/solutions/cmd/aggregator -dir=/tmp/logs                # channel (default)
go run ./lessons/20-concurrency-patterns/solutions/cmd/aggregator -dir=/tmp/logs -mode=mutex    # mutex
go run ./lessons/20-concurrency-patterns/solutions/cmd/aggregator -dir=/tmp/logs -mode=pool     # pool, default workers
go run ./lessons/20-concurrency-patterns/solutions/cmd/aggregator -dir=/tmp/logs -mode=pool -workers=8
```

## Going further

### Read

- **`golang.org/x/sync/errgroup` source** — read the real thing: <https://pkg.go.dev/golang.org/x/sync/errgroup> + source. Our `errgroupx` is ~99% identical (we skip `SetLimit`).
- **Go blog — "Go Concurrency Patterns: Pipelines and cancellation"** (Sameer Ajmani, 2014): <https://go.dev/blog/pipelines> — the canonical reference. Most of L20 is in there.
- **Go blog — "Advanced Go Concurrency Patterns"** (Sameer Ajmani, 2013): <https://go.dev/blog/io2013-talk-concurrency> — fan-out / fan-in / select tricks.

### Try

- **Benchmark the three modes.** Add `BenchmarkWalk`, `BenchmarkWalkLocked`, `BenchmarkWalkPool` over a fixture of 100 files. Compare ns/op + allocs/op. Which wins on your machine? Why? (L22 dives deep into profiling.)
- **Sweep the workers parameter.** Run `WalkPool` with `workers=1, 2, 4, 8, 16, 32, 64, 128` over a fixture of 1000 files. Plot time-to-completion. Find your machine's saturation point.
- **Pipeline the CSV importer.** Rewrite L13's `csvimport` as a 3-stage pipeline: read raw lines → parse rows → filter valid → return. Each stage takes ctx and its own input channel. Compare to the original single-function version.
- **Add `SetLimit` to errgroupx.** The real `golang.org/x/sync/errgroup` has a `SetLimit(n)` method that bounds in-flight goroutines (so `Go` blocks until capacity is available). Implement it. Hint: a `chan struct{}` semaphore.

---

> Phase 3 is almost done. Next stop: Lesson 21 — Networking & syscalls. We'll build a tiny TCP echo server + client. The aggregator stays where L20 left it.
