<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">16</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 3 — Concurrency &amp; Systems</div>
<h1>Goroutines &amp; channels</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>First exposure to Go's concurrency primitives. Learn the <code>go</code> keyword and the goroutine model, unbuffered channels (synchronous handoff), buffered channels (capacity), <code>range</code>/<code>close</code> + goroutine lifecycle, and the <strong>aggregator pattern</strong> — multiple producers, one consumer — which carries forward through L17-L20 as the running example evolves.</p>
</div>
</div>
</div>

---

## What we'll cover

- **Goroutines** — `go f()` runs `f` concurrently. No return values. Communicate via channels.
- **Unbuffered channels** — synchronous handoff. Sender and receiver rendezvous.
- **Buffered channels** — capacity. First N sends don't block.
- **`range`, `close`, goroutine lifecycle** — proper shutdown, avoiding leaked goroutines.
- **The aggregator pattern** — multiple producers, single consumer. The running example.

---

## Welcome to Phase 3

Phase 1 taught you to write Go. Phase 2 taught you to write *idiomatic* Go. Phase 3 teaches you to write Go that does **multiple things at once**.

Concurrency is the part of Go most people remember as "the new thing." Goroutines are cheap (a few KB of stack), channels are first-class language constructs, and the runtime's scheduler handles thousands of them without breaking a sweat.

This lesson is the foundation. The next five lessons build on it.

---

## Concept 1: Goroutines

### Motivation

In Go, every program starts with one goroutine — `main`. Until you write `go f()`, that's all you have.

The `go` keyword launches a function as a new concurrent thread of execution. It runs alongside `main`. It does not block. It does not return a value to the caller.

```go
go f()    // f starts running; main continues immediately
```

This is the single most confusing thing for students coming from languages with Futures or Promises. There's no handle. No `await`. No `.then()`. The communication mechanism is **channels** (next concept).

---

### The basics

```go
package main

import "fmt"

func say(msg string) {
    fmt.Println(msg)
}

func main() {
    go say("hello from goroutine")
    say("hello from main")
    // The goroutine may or may not have printed by the time main exits.
}
```

Three things to internalise:

- **`go` returns nothing.** The expression is a statement, not a value. You can't `result := go f()`.
- **Goroutines are cheap.** Each one starts with ~2 KB of stack; the runtime grows them as needed. Thousands of goroutines is fine; millions is plausible.
- **Goroutines outlive their launch site.** When `main` returns, all goroutines die immediately — finished or not. For long-running goroutines, you need a way to know they're done. That's what channels are for.

---

### A worked example

The lesson's `Echo` warmup is the smallest non-trivial use of goroutines:

```go
func Echo(in <-chan string) <-chan string {
    out := make(chan string)
    go func() {
        for v := range in {
            out <- v
        }
        close(out)
    }()
    return out
}
```

`Echo` returns immediately. The goroutine reads from `in`, sends to `out`, and closes `out` when `in` closes. The caller can range over the output channel:

```go
in := make(chan string)
out := Echo(in)

go func() {
    in <- "hello"
    in <- "world"
    close(in)
}()

for v := range out {
    fmt.Println(v)  // "hello", then "world"
}
```

Two goroutines (producer + Echo) plus the main goroutine reading from `out`. Three concurrent threads of execution.

---

### Common mistake

Trying to capture a "return value" from `go`:

```go
// the wrong way
result := go f()   // syntax error: go is a statement, not an expression
```

Or the subtler version — calling a function that returns something and discarding it:

```go
// the wrong way (compiles but does nothing useful)
go calculate()     // calculate returns int; that int is dropped on the floor
```

If the function had a return value you cared about, you can't get it back through `go`. The fix is a channel:

```go
result := make(chan int)
go func() {
    result <- calculate()
}()
value := <-result
```

You launch a goroutine that sends its result on a channel; the caller waits on that channel. That's the pattern.

---

### Recap

- `go f()` launches f as a concurrent goroutine.
- No return values; communicate via channels.
- Cheap: thousands of goroutines is fine.
- `main` returning kills all goroutines immediately.

---

## Concept 2: Unbuffered channels

### Motivation

A channel is a typed conduit through which you can send values. Unbuffered channels are the simplest kind: a send blocks until a matching receive happens (and vice versa). The runtime synchronises sender and receiver at the moment of the handoff.

This "rendezvous" model is how goroutines communicate without locking.

---

### The basics

```go
ch := make(chan int)    // unbuffered channel of int
ch <- 42                // send (blocks until someone is receiving)
v := <-ch               // receive (blocks until someone sends)
```

Both operations BLOCK until the other side is ready. If a sender has nobody to give the value to, it waits. If a receiver has nobody to take from, it waits. They synchronise at the moment of handoff.

Direction syntax:

```go
chan int           // can send AND receive
chan<- int         // send-only (write-only)
<-chan int         // receive-only (read-only)
```

You'll see these in function signatures. A function parameter typed `chan<- int` can only send; `<-chan int` can only receive. This is a documentation feature — you can't accidentally read from a channel that's supposed to be write-only.

---

### A worked example

Two goroutines synchronizing via an unbuffered channel:

```go
done := make(chan bool)

go func() {
    fmt.Println("worker: starting")
    time.Sleep(time.Second)
    fmt.Println("worker: done")
    done <- true   // signal completion
}()

<-done   // wait for the worker
fmt.Println("main: worker finished")
```

The main goroutine blocks on `<-done` until the worker sends. The worker's `done <- true` blocks until main is reading. They rendezvous; both proceed.

This is the foundation of "wait for another goroutine to finish." Lesson 18 (sync) introduces `sync.WaitGroup` which generalises this pattern to N workers.

---

### Common mistake

Sending without a receiver — the classic "all goroutines are asleep" deadlock:

```go
// the wrong way
ch := make(chan int)
ch <- 42   // blocks forever; nobody is receiving
```

The Go runtime detects this and panics:

```
fatal error: all goroutines are asleep - deadlock!
```

Don't let this scare you — it's a CORRECT detection. The runtime is telling you the program can't make progress. Either add a receiver (in another goroutine), make the channel buffered (concept 3), or restructure the logic.

The same deadlock applies to `<-ch` when no sender exists.

---

### Recap

- `make(chan T)` creates an unbuffered channel.
- Send `ch <- v` and receive `v := <-ch` both BLOCK until the other side is ready.
- Channel direction syntax (`chan<- T`, `<-chan T`) documents intent.
- Deadlock if there's no counterpart goroutine.

---

## Concept 3: Buffered channels

### Motivation

Sometimes you want senders to run ahead of receivers. A web crawler that fetches faster than it processes; a log collector that accepts bursts; a producer that knows the consumer will catch up. Buffered channels give you that headroom.

A buffered channel holds up to N values in flight. First N sends don't block. The (N+1)-th send waits for a receiver.

---

### The basics

```go
ch := make(chan int, 3)   // capacity 3

ch <- 1   // doesn't block
ch <- 2   // doesn't block
ch <- 3   // doesn't block
ch <- 4   // blocks until someone receives
```

Inspect a channel's state:

```go
len(ch)   // number of values currently in the buffer
cap(ch)   // capacity (the N you passed to make)
```

When to use which:

| Use case | Channel kind |
|---|---|
| "I need to know the receiver got it" | unbuffered |
| "N producers can run ahead of one consumer" | buffered, capacity N |
| "Bursty input; smooth processing" | buffered, capacity = burst size |
| "Pipeline stage" | buffered or unbuffered depending on latency vs throughput |

Bigger buffer does NOT mean faster. The buffer is a **backpressure tuning knob** — it lets producers temporarily outpace consumers, but it doesn't actually do work faster.

---

### A worked example

The L16 aggregator uses a buffered channel sized to the number of files:

```go
files := []string{"a.log", "b.log", "c.log"}
resultsCh := make(chan fileResult, len(files))   // capacity = 3

for _, path := range files {
    go processFile(path, resultsCh)   // workers never block on send
}

for i := 0; i < len(files); i++ {
    r := <-resultsCh
    // ... reduce r into combined map
}
```

Each worker sends one `fileResult` and exits. With a buffer of `len(files)`, workers never wait for the main goroutine — they send and finish. The main goroutine collects results at its own pace.

If we'd used `make(chan fileResult)` (unbuffered), the first N-1 workers would block on send until main read their results. Correct, just less parallel.

---

### Common mistake

Thinking "bigger buffer = faster":

```go
// the wrong way — thinking through capacity
ch := make(chan int, 1000000)   // huge buffer
```

A million-slot buffer doesn't make your producer or consumer any faster. It just lets the producer race far ahead. If the consumer is slow, you've moved the bottleneck without fixing it — you've just hidden it behind a queue.

Buffers are for **smoothing**, not **speedup**. If your consumer is genuinely too slow, add more consumers (worker pool — L20) or make the consumer faster.

---

### Recap

- `make(chan T, N)` creates a buffered channel with capacity N.
- First N sends don't block; (N+1)-th send waits for a receive.
- `len(ch)` and `cap(ch)` inspect state.
- Buffer = backpressure tuning, not speedup.

---

## Concept 4: `range`, `close`, and goroutine lifecycle

### Motivation

So far we've sent values one at a time. The natural extension is "send a sequence, then signal we're done." Go uses `close(ch)` as the "no more sends coming" signal, and `for v := range ch` reads until close.

This is also where the most common bug hides: **goroutine leaks**. If you start a goroutine that's waiting to send on a channel nobody reads, it blocks forever. The program keeps running but the goroutine is permanently stuck.

---

### The basics

```go
ch := make(chan int)

go func() {
    for i := 1; i <= 3; i++ {
        ch <- i
    }
    close(ch)   // signal: no more sends
}()

for v := range ch {
    fmt.Println(v)   // 1, 2, 3
}
// range exits when ch is closed
```

Three pieces to internalise:

- **`close(ch)` is the sender's job.** Whoever sends on a channel decides when to close it. Closing from the receiver side is wrong (the sender might still want to send).
- **Receiving from a closed channel returns the zero value immediately, no block.** Useful for "drain remaining items":
  ```go
  v, ok := <-ch   // ok is false if ch is closed and empty
  ```
- **Closing a channel twice panics.** Closing a nil channel panics. Sending on a closed channel panics. Be deliberate about close.

---

### A worked example

The aggregator's worker functions don't close `resultsCh` — main reads exactly N values and exits the loop. But if you wanted to use `for r := range resultsCh` instead of a counted loop, you'd need someone to close after the last worker finishes:

```go
resultsCh := make(chan fileResult, len(files))
var wg sync.WaitGroup       // L18 preview

for _, path := range files {
    wg.Add(1)
    go func(p string) {
        defer wg.Done()
        processFile(p, resultsCh)
    }(path)
}

// Closer goroutine: wait for all workers, then close.
go func() {
    wg.Wait()
    close(resultsCh)
}()

// Main reduces via range.
combined := map[string]int{}
for r := range resultsCh {
    if r.err != nil { /* ... */ }
    for level, count := range r.counts { combined[level] += count }
}
```

This shape uses `sync.WaitGroup` (L18) — preview. The L16 version uses a counted loop instead, which is simpler.

---

### Common mistake

**Goroutine leaks.** Starting a goroutine that blocks forever because nobody completes the protocol:

```go
// the wrong way
func leaky() {
    ch := make(chan int)
    go func() {
        ch <- expensiveCompute()   // blocks forever; nobody reads ch
    }()
    // function returns; goroutine still alive, blocked on send
}
```

When `leaky` returns, the channel goes out of scope from the caller's view, but the GOROUTINE still holds a reference to it. The goroutine sits there forever, holding memory and a runtime slot.

The fix: ensure every goroutine has a path to exit. Either someone reads from `ch`, or you use `select` with a timeout (L17), or you pass a `context` for cancellation (L19).

Goroutine leaks don't crash your program. They slowly accumulate. The race detector won't catch them — they're not data races. Look for them by instrumenting (`runtime.NumGoroutine()`) or just being careful in code review.

---

### Recap

- `close(ch)` signals "no more sends." Sender's responsibility.
- `for v := range ch` reads until close.
- `v, ok := <-ch` — `ok` is false on closed-and-drained.
- Goroutine leaks: goroutines blocked forever because the protocol never completes. Always have an exit path.

---

## Concept 5: The aggregator pattern

### Motivation

You now know the primitives: `go`, channels (unbuffered and buffered), `range`, `close`. Time to put them together into a NAMED pattern.

The **aggregator pattern** is the simplest non-trivial concurrent program in Go: **multiple producers, one consumer**. N goroutines do work in parallel; they send results through a shared channel; one main goroutine reduces.

This pattern is the lesson's running example. You'll see it again in L17 (with timeouts), L18 (with shared state instead of channels), L19 (with cancellation), and L20 (with a worker pool that bounds parallelism).

---

### The basics

The shape:

```
[workerA] ──┐
            │
[workerB] ──┼──► [resultsCh] ──► [main]
            │
[workerC] ──┘
```

N workers → one channel → one reducer. The workers don't know about each other; they only know how to do their job and send a result. The reducer doesn't know how many workers ran; it just counts results.

---

### A worked example

The L16 aggregator, in full:

```go
func Walk(dir string) (map[string]int, error) {
    entries, err := os.ReadDir(dir)
    if err != nil {
        return nil, fmt.Errorf("aggregator: read dir %s: %w", dir, err)
    }

    files := []string{}
    for _, e := range entries {
        if e.IsDir() { continue }
        files = append(files, filepath.Join(dir, e.Name()))
    }

    resultsCh := make(chan fileResult, len(files))

    for _, path := range files {
        go processFile(path, resultsCh)
    }

    combined := map[string]int{}
    for i := 0; i < len(files); i++ {
        r := <-resultsCh
        if r.err != nil {
            return combined, fmt.Errorf("aggregator: %s: %w", r.path, r.err)
        }
        for level, count := range r.counts {
            combined[level] += count
        }
    }
    return combined, nil
}
```

Three pieces:

1. **Fan-out** (`for _, path := range files { go processFile(path, resultsCh) }`) — spawn N workers.
2. **Reduce** (`for i := 0; i < len(files); i++ { r := <-resultsCh; ... }`) — read exactly N results.
3. **Buffered channel** (`make(chan fileResult, len(files))`) — workers never block.

The error handling is intentionally simple: the first error returned by any worker aborts the reduce loop with a partial result. We'll do better in L20 (errgroup-style first-error-wins with cancellation).

---

### Common mistake

Forgetting the buffered channel and getting tighter coupling than you wanted:

```go
// the wrong way (works, just less parallel)
resultsCh := make(chan fileResult)   // unbuffered
for _, path := range files {
    go processFile(path, resultsCh)
}
```

With an unbuffered channel, the first N-1 workers send and immediately BLOCK waiting for main to receive. The N-th worker is the only one running while main is reducing the first. You get goroutines pinned around the channel, doing nothing.

For a fixed-size fan-out where you know N up front, **buffer with capacity = N**. Workers send, exit, free their resources. The channel acts as a parking lot for completed work.

---

### Recap

- Aggregator pattern = multiple producers, one consumer.
- Fan-out: spawn N goroutines, each does one piece.
- Reduce: read N results from a shared channel.
- Buffer with capacity N when N is known up front.
- You'll see this pattern through L17-L20 as we add timeouts, sync, context, and worker pools.

---

## Practice

### Warm-up

Implement `Echo(in <-chan string) <-chan string`. Read everything from `in`, send through a new output channel, close output when `in` closes. The function MUST return immediately via a spawned goroutine.

```bash
cd lessons/16-goroutines-channels/exercises/warmup/echo
go test -v
go test -race    # verify no data races
```

---

### Main

Build `aggregator.Walk(dir string) (map[string]int, error)` — the lesson's running example. Walks a directory, spawns one goroutine per file, parses each with `logparse.Parse`, collects per-file level counts via a buffered channel, reduces into a combined map. Plus `cmd/aggregator/main.go` — the CLI binary with full os/exec integration tests.

```bash
cd lessons/16-goroutines-channels/exercises
go test ./...
go test -race ./...    # the race detector is your friend
```

Manual smoke test:

```bash
mkdir /tmp/logs
echo "2026-05-21T14:30:00 INFO server started" > /tmp/logs/a.log
echo "2026-05-21T14:31:00 WARN slow query" > /tmp/logs/b.log
echo "2026-05-21T14:32:00 ERROR conn refused" > /tmp/logs/c.log
go run ./cmd/aggregator -dir=/tmp/logs
# ERROR: 1
# INFO: 1
# WARN: 1
```

Note:
The race detector (`go test -race ./...`) catches data races at test time with 5-20× runtime overhead. L18 makes `make test-race` a daily habit; for now, treat it as a tool you can reach for when you suspect concurrency issues.

---

## Closing thought

Go's concurrency model has a slogan: **"Don't communicate by sharing memory; share memory by communicating."** Other languages reach for mutexes and locks by default; Go's idiomatic move is to send values through channels instead.

That's not a rule — sometimes a mutex IS the right tool (L18). But for THIS lesson's running example, channels do everything: workers send results, main reads them, no shared state. The aggregator pattern is the slogan in code.

---

## What we learned

- `go f()` launches a concurrent goroutine. No return values; communicate via channels.
- Unbuffered channels: synchronous handoff. Send and receive rendezvous.
- Buffered channels: capacity N. First N sends don't block.
- `range` reads until close. `close(ch)` is the sender's job.
- The aggregator pattern: multiple producers, one consumer. See it through L17-L20.

---

## Up next

Lesson 17 — **Select & timers**. The `select` statement multiplexes channels; `time.After` gives you timeouts. We'll add a per-file timeout to the aggregator so one slow file doesn't block the rest.
