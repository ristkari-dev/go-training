# Lesson 17: Select & timers

## What you'll learn

By the end of this lesson you can:

- Use `select` to multiplex N channels in one statement, knowing the runtime picks randomly when multiple cases are ready.
- Time-bound operations with `time.After(d)` + select, including the nil-channel-never-fires trick for "no timeout."
- Use `time.NewTicker(d)` for periodic events — and always `defer t.Stop()`.
- Use `default` in select for non-blocking try-receive / try-send patterns.
- Build cancellation via `close(done)` on a `chan struct{}` — the foundation for L19's `context.Context`.

## What's different from L16

L16 built the aggregator's bones: one goroutine per file, results via channel, single reducer. L17 evolves it:

- **`Walk` signature changes**: `Walk(dir, timeout)` returns a `WalkResult` struct with `Counts` AND `TimedOut []string`. Library stays I/O-free; CLI handles printing.
- **Per-file timeout**: the reduce loop's `select` includes a `time.After(timeout)` case. `timeout=0` means "no timeout" via the nil-channel trick.
- **CLI gains `-timeout=<dur>`**: prints timed-out paths to stderr, aggregated counts to stdout.

The aggregator still uses the L16 "share by communicating" pattern. L18 will introduce the alternative — "share memory by locking" with `sync.Mutex` — and we'll compare them.

## The package layout

```
lessons/17-select-timers/exercises/
├── warmup/wait/                      ← you implement: WaitWithTimeout[T any]
├── cmd/aggregator/                    ← evolves: -timeout flag, stderr printing
│   ├── main.go
│   └── main_test.go                  ← golden + malformed + missing flag + timeout
└── internal/
    ├── logparse/                      ← verbatim from L16
    └── aggregator/                    ← evolves: WalkResult + timeout
        ├── aggregator.go
        └── aggregator_test.go
```

---

## Concept 1 — `select`

`select` multiplexes channel operations. Blocks until at least one case is ready; if multiple are ready, the runtime picks one at random.

```go
select {
case v := <-chA:
    // chA had a value ready
case v := <-chB:
    // chB was ready first (or both — runtime picks randomly)
case chOut <- 42:
    // chOut had a receiver ready, and we sent 42
}
```

Each case must be a channel operation (send, receive, or receive-with-ok). No function calls, no arbitrary expressions.

### Common mistake

Assuming cases have priority based on declaration order. They don't. If you need priority, use two selects: a `default`-guarded poll of the high-priority channel, falling through to a multi-case select if it wasn't ready.

---

## Concept 2 — `time.After` + timeouts

`time.After(d)` returns a `<-chan time.Time` that fires once after duration `d`. The canonical timeout pattern:

```go
select {
case v := <-workCh:
    handle(v)
case <-time.After(5 * time.Second):
    log("timed out")
}
```

The L17 aggregator uses this in its reduce loop with one extra trick — the **nil-channel-never-fires** pattern for optional timeouts:

```go
var timeoutCh <-chan time.Time
if timeout > 0 {
    timeoutCh = time.After(timeout)
}
select {
case r := <-resultsCh: /* ... */
case <-timeoutCh:        // nil channel never fires — "no timeout" path
    /* ... */
}
```

A nil channel in a select never gets selected. So `timeout=0` falls through to the unconditional receive case.

### Common mistake

`time.After` allocates a fresh `*time.Timer` per call. In a hot loop, that's GC pressure. For hot paths, allocate once with `time.NewTimer` and call `Reset` per iteration. For one-shot use (like the aggregator's per-file timeout), `time.After` is idiomatic.

---

## Concept 3 — `time.NewTicker`

Periodic events:

```go
t := time.NewTicker(time.Second)
defer t.Stop()    // critical
for {
    select {
    case <-t.C:
        heartbeat()
    case <-done:
        return
    }
}
```

`t.C` is the channel; receives a `time.Time` every interval.

The L17 aggregator doesn't use a ticker (scope creep — we don't have a long-running operation to heartbeat). The hypothetical "print progress every 2 seconds" use case fits L20's worker pool better.

### Common mistake

Forgetting `t.Stop()`. The ticker keeps firing into `t.C` (small buffer) after your function returns, and the goroutine driving it leaks. **Always `defer t.Stop()` immediately after creation.**

---

## Concept 4 — `default` branch (non-blocking select)

`default` fires when no other case is immediately ready. Makes the select return instantly instead of blocking.

```go
// Try-receive: receive if a value is waiting, else fall through
select {
case v := <-ch:
    handle(v)
default:
    // ch had nothing right now
}

// Try-send: send if a receiver is ready, else drop
select {
case ch <- v:
    // queued
default:
    // queue full, drop or escalate
}
```

The classic use: back-pressure in async logging — drop messages when the queue is full rather than blocking the caller.

### Common mistake

Using `default` in a loop:

```go
for {
    select {
    case v := <-ch: /* ... */
    default: /* burns CPU */
    }
}
```

Without `default`, the select blocks efficiently until `ch` has a value. With `default`, the loop spins at 100% CPU. **`default` is for "try once," not "wait."**

---

## Concept 5 — Cancellation via close-of-done-channel

The Go idiom for signaling "stop" to one or more goroutines:

```go
done := make(chan struct{})

go func() {
    for {
        select {
        case <-done:
            return
        case work := <-workCh:
            process(work)
        }
    }
}()

// When ready to cancel:
close(done)
```

Three things to internalise:

- **Closing is the signal.** Receiving from a closed channel returns the zero value immediately — every receiver sees it at once. Perfect for broadcast cancellation.
- **`chan struct{}`, not `chan bool`.** The value doesn't matter; we just need "is it closed." `struct{}` takes zero bytes.
- **Closer is whoever decides to cancel.** Not the receiver. The closer is the goroutine that wants others to stop.

L19 generalizes this pattern via `context.Context` — the canonical Go cancellation API. When you read that lesson, notice that `ctx.Done()` returns exactly the kind of channel we built here.

### Common mistake

Sending on `done` instead of closing it. `done <- struct{}{}` only reaches ONE receiver; with N workers you'd need N sends. `close(done)` broadcasts to ALL receivers atomically. **For cancellation, always close.**

Secondary mistake: closing twice panics. If you might cancel from multiple places, wrap in `sync.Once` (L18 preview): `var once sync.Once; cancel := func() { once.Do(func() { close(done) }) }`.

---

## Exercise: warm-up — `wait`

Implement `WaitWithTimeout[T any](ch <-chan T, d time.Duration) (T, error)` in `exercises/warmup/wait/wait.go`. Returns the first value from ch, or `(zero T, ErrTimeout)` after d. Uses `select { case v := <-ch: ...; case <-time.After(d): ... }`.

Sentinel `var ErrTimeout = errors.New("wait: timeout")` lets callers distinguish via `errors.Is(err, wait.ErrTimeout)`.

**Time:** 5-10 minutes.

## Exercise: main — aggregator with timeout

Evolve the aggregator:

1. **`WalkResult` struct** with `Counts map[string]int` and `TimedOut []string`.
2. **`Walk(dir, timeout time.Duration) (WalkResult, error)`** — reduce loop uses `select` with `time.After(timeout)`. `timeout=0` uses the nil-channel trick for "no timeout."
3. **`cmd/aggregator -timeout=<dur>` flag** — parses with `time.ParseDuration`; prints timed-out paths to stderr.

**Time:** 30-45 minutes.

---

## Daily habits

After every change:

```bash
gofmt -w ./...
go vet ./...
go test ./...
```

For concurrency code (Phase 3 onward):

```bash
go test -race ./...     # catch data races (5-20× slower)
```

The race detector is invaluable for concurrency. L18 makes `make test-race` a Makefile target you run regularly — for now, reach for `-race` whenever you're working with channels or goroutines.

## How to run

```bash
# Skeleton (passes vacuously until you implement)
cd lessons/17-select-timers/exercises
go test ./...

# Reference solution
go test -v ./lessons/17-select-timers/solutions/...

# Race detector
go test -race ./lessons/17-select-timers/solutions/...

# The binary — normal run
mkdir /tmp/logs
echo "2026-05-21T14:30:00 INFO server started" > /tmp/logs/a.log
echo "2026-05-21T14:31:00 WARN slow" > /tmp/logs/b.log
go run ./lessons/17-select-timers/solutions/cmd/aggregator -dir=/tmp/logs

# The binary — with timeout (everything times out at 1 nanosecond)
go run ./lessons/17-select-timers/solutions/cmd/aggregator -dir=/tmp/logs -timeout=1ns
# stderr: "timeout: /tmp/logs/a.log\ntimeout: /tmp/logs/b.log"
```

## Going further

### Read

- **Go blog — "Go Concurrency Patterns: Timing out, moving on"**: <https://go.dev/blog/concurrency-timeouts> — Andrew Gerrand walks through select + timeout patterns.
- **`time` package docs**: <https://pkg.go.dev/time> — Timer vs Ticker vs After. The "Stop" sections are essential reading.
- **Dave Cheney — "Curious Channels"** (2013, still relevant): <https://dave.cheney.net/2013/04/30/curious-channels> — nil channels, closed channels, and how they behave in select.

### Try

- **Whole-Walk timeout instead of per-file**: hoist the `time.After(timeout)` call OUT of the reduce loop. What's the semantic difference? When would you want each?
- **Ticker-based heartbeat in cmd/aggregator**: add a goroutine that prints "... still processing" every 2 seconds during the walk. Use the `done` channel pattern from concept 5 to stop it when Walk returns.
- **Try-send back-pressure**: extend the aggregator with a buffered `logCh chan string` for verbose-mode logging. Use `select { case logCh <- msg: ...; default: ... }` to drop log messages when the queue is full, counting drops via `sync/atomic` (preview of L18).

---

> Next stop: Lesson 18 — `sync` & memory model. The aggregator's shared state moves from "communicate via channels" to "share with `sync.Mutex`" — and we'll compare them side-by-side. `make test-race` becomes a daily habit.
