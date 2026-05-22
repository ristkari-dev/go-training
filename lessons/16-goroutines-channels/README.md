# Lesson 16: Goroutines & Channels

> **Phase 3 begins.**

## What you'll learn

By the end of this lesson you can:

- Launch a concurrent goroutine with `go f()` and reason about its lifetime.
- Use unbuffered channels for synchronous handoff (sender + receiver rendezvous).
- Use buffered channels for backpressure (first N sends don't block).
- Range over a channel until close, and avoid goroutine leaks.
- Recognise the **aggregator pattern** (multiple producers, single consumer) — the lesson's running example, which carries forward through L17-L20.

## What's different from L15

L15 wrapped Phase 2 with the cmd/+internal/ tracker reorganization. **L16 begins Phase 3** — concurrency. The expense tracker stays where L15 left it; this phase introduces a NEW running example: the log aggregator.

The aggregator walks a directory of log files, parses each in parallel (one goroutine per file), and aggregates the per-file level counts into a combined map. L16 builds the bones; L17 adds per-file timeouts; L18 introduces mutex-protected shared state; L19 wires context cancellation; L20 turns it into a bounded worker pool.

This is also the lesson where the **race detector** enters your toolkit. `go test -race ./...` catches data races at test time with 5-20× runtime overhead. L18 makes `make test-race` a daily habit; for now, treat it as a tool you can reach for when concurrency feels suspicious.

## The package layout

```
lessons/16-goroutines-channels/exercises/
├── warmup/echo/                         ← you implement: Echo(io.Reader-ish channel pattern)
├── cmd/aggregator/                       ← the CLI binary
│   ├── main.go
│   └── main_test.go                     ← full os/exec integration tests
└── internal/
    ├── logparse/                         ← verbatim carry-forward from L14
    └── aggregator/                       ← you implement: Walk(dir) function
        ├── aggregator.go
        └── aggregator_test.go
```

---

## Concept 1 — Goroutines

`go f()` launches `f` as a concurrent thread of execution. The expression is a statement, not a value — there's no return, no handle, no Future. The function runs alongside whatever called it, and communication happens via channels (concept 2).

```go
go say("hello")    // runs concurrently with the caller
say("world")       // runs in the caller's goroutine
```

Three things to internalise:

- **Goroutines are cheap.** Each starts with ~2 KB of stack; the runtime grows them on demand. Thousands is fine.
- **`main` returning kills everything.** When the main goroutine exits, every other goroutine dies instantly. For long-running goroutines you need a way to know they're done (channels).
- **No return values.** If `calculate()` returns an int, `go calculate()` drops the int on the floor. To capture results, send through a channel.

### Common mistake

Trying to assign from `go`:

```go
result := go f()   // syntax error: go is a statement
```

The fix is always a channel — launch a goroutine that sends its result through a channel; the caller waits on that channel.

---

## Concept 2 — Unbuffered channels

A channel is a typed conduit between goroutines. Unbuffered channels (`make(chan T)`) are the simplest kind: a send BLOCKS until a receive happens, and vice versa. The two operations rendezvous at the moment of handoff.

```go
ch := make(chan int)
ch <- 42       // blocks until someone receives
v := <-ch      // blocks until someone sends
```

Direction syntax documents intent:

```go
chan int          // can send AND receive
chan<- int        // send-only
<-chan int        // receive-only
```

You'll see these in function parameters: `Echo` takes a `<-chan string` (read-only) and returns a `<-chan string` (also read-only — the caller can't write to it).

### Common mistake

Sending without a receiver — the runtime detects this and panics with `fatal error: all goroutines are asleep - deadlock!`. The fix is to either add a receiver (in another goroutine), use a buffered channel (concept 3), or restructure the code.

---

## Concept 3 — Buffered channels

`make(chan T, N)` creates a channel with capacity N. The first N sends don't block; the (N+1)-th send waits for a receiver.

```go
ch := make(chan int, 3)
ch <- 1; ch <- 2; ch <- 3    // none block
ch <- 4                       // blocks
```

When to use which:

| Need | Channel |
|---|---|
| "I need to know the receiver got it" | unbuffered |
| "N producers can run ahead of one consumer" | buffered, capacity N |
| "Bursty input; smooth processing" | buffered, capacity = burst size |

The aggregator uses `make(chan fileResult, len(files))` so workers never block on send.

### Common mistake

Thinking "bigger buffer = faster." Buffers are a **backpressure tuning knob**, not a speedup. A million-slot buffer doesn't make the producer or consumer faster; it just lets them decouple temporarily. If the consumer is genuinely too slow, add more consumers (L20 worker pool) or make the consumer faster.

---

## Concept 4 — `range`, `close`, and goroutine lifecycle

`close(ch)` signals "no more sends coming." `for v := range ch` reads until close. Receiving from a closed-and-drained channel returns the zero value immediately (no block); `v, ok := <-ch` returns `ok=false` in that case.

```go
go func() {
    for i := 0; i < 3; i++ { ch <- i }
    close(ch)
}()
for v := range ch { fmt.Println(v) }    // exits when ch closes
```

Rules:

- **Close is the sender's job.** The receiver doesn't know when sending is done.
- **Closing twice panics.** Closing a nil channel panics. Sending on a closed channel panics. Be deliberate.

### Common mistake

**Goroutine leaks.** A goroutine that blocks forever because the protocol never completes:

```go
go func() { ch <- expensiveResult() }()  // blocks forever if nobody reads ch
```

The function returns, the channel becomes unreachable from the caller's view, but the goroutine still holds a reference. It sits there forever, holding memory. The race detector doesn't catch this; you find it via `runtime.NumGoroutine()` or code review.

The fix: every goroutine needs a path to exit. Someone reads the channel, OR you use `select` with a timeout (L17), OR you pass a `context` for cancellation (L19).

---

## Concept 5 — The aggregator pattern

The lesson's running example, named explicitly: **multiple producers, single consumer**.

```
[workerA] ──┐
            │
[workerB] ──┼──► [resultsCh] ──► [main]
            │
[workerC] ──┘
```

The shape:

```go
resultsCh := make(chan fileResult, len(files))

for _, path := range files {
    go processFile(path, resultsCh)    // fan-out
}

for i := 0; i < len(files); i++ {
    r := <-resultsCh                    // reduce
    // ... merge r into combined
}
```

Three pieces: fan-out (N goroutines), shared channel (buffered when N is known), reduce (one consumer).

You'll see this pattern again as the aggregator evolves: L17 adds per-file timeouts via `select`; L18 swaps "send maps through channels" for "shared map with `sync.Mutex`"; L19 propagates `context` for cancellation; L20 bounds parallelism with a worker pool.

### Common mistake

Forgetting the buffered channel:

```go
resultsCh := make(chan fileResult)    // unbuffered
for _, path := range files {
    go processFile(path, resultsCh)
}
```

The first N-1 workers send and immediately block waiting for main. Only the last worker runs concurrently with the reducer. You lose parallelism without realising it. For known-size fan-out, **buffer with capacity = N**.

---

## Exercise: warm-up — `echo`

Implement `Echo(in <-chan string) <-chan string` in `exercises/warmup/echo/echo.go`. The function MUST return immediately via a spawned goroutine; the goroutine reads from `in` and forwards to a new output channel, closing the output when `in` closes.

**Time:** 5-10 minutes.

## Exercise: main — aggregator + cmd

Three pieces, all in `exercises/`:

1. **`internal/aggregator/Walk(dir string) (map[string]int, error)`** — list files in dir, spawn one goroutine per file, parse each with `logparse.Parse`, send per-file counts through a buffered channel, reduce. On any file error, return partial map + wrapped error mentioning the file.
2. **`cmd/aggregator/main.go`** — parse `-dir=<path>`, call Walk, print sorted level counts to stdout. Exit 1 on errors.
3. **`cmd/aggregator/main_test.go`** — integration tests build the binary in `t.TempDir`, exec it against fixture log files, assert stdout. Already provided as skeleton.

**Time:** 30-45 minutes.

---

## Daily habits

After every change:

```bash
gofmt -w ./...
go vet ./...
go test ./...
```

NEW in Phase 3 — when concurrency is involved:

```bash
go test -race ./...        # catch data races (5-20× slower)
```

L18 will make `make test-race` a Makefile target you run regularly. For now, reach for `-race` whenever you're touching channel code or suspect a goroutine bug.

## How to run

```bash
# Skeleton (passes vacuously until you implement)
cd lessons/16-goroutines-channels/exercises
go test ./...

# Reference solution
go test -v ./lessons/16-goroutines-channels/solutions/...

# Race detector
go test -race ./lessons/16-goroutines-channels/solutions/...

# The binary
mkdir /tmp/logs
echo "2026-05-21T14:30:00 INFO server started" > /tmp/logs/a.log
echo "2026-05-21T14:31:00 WARN slow" > /tmp/logs/b.log
go run ./lessons/16-goroutines-channels/solutions/cmd/aggregator -dir=/tmp/logs
# Prints: INFO: 1
#         WARN: 1
```

## Going further

### Read

- **Effective Go — "Concurrency"**: <https://go.dev/doc/effective_go#concurrency> — the canonical motivation; everything in this lesson is in there too.
- **Go blog — "Share memory by communicating"**: <https://go.dev/blog/codelab-share> — the slogan, explained.
- **"Visualizing Go concurrency" (Ivan Daniluk, 2016)** — old but still illustrative: <https://divan.dev/posts/go_concurrency_visualize/>

### Try

- **Run the aggregator on `/var/log`** (or any directory full of files). What happens with hundreds of goroutines? Watch CPU usage. The runtime handles it without breaking a sweat.
- **Benchmark with vs without goroutines.** Write a serial version of `Walk` that processes files one at a time (no `go`); compare ns/op to the concurrent version on 10/100/1000 files. The benefit depends on whether file I/O or parsing dominates.
- **Sneak preview of L20's worker pool.** Add a `-workers=N` flag to `cmd/aggregator`. Implement a bounded version of Walk where at most N goroutines run at once (hint: use a `sem := make(chan struct{}, N)` token bucket). Compare to the unbounded version when there are 10,000 files.

---

> Phase 3 begins. Next stop: `select` and timers (L17) — handling the slow file that would otherwise block everything else.
