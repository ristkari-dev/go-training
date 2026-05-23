<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">17</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 3 — Concurrency &amp; Systems</div>
<h1>Select &amp; timers</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Learn <code>select</code> for multiplexing channels, <code>time.After</code> + <code>time.NewTicker</code> for time-bounded operations, the <code>default</code> branch for non-blocking select, and cancellation via close-of-done-channel — the foundation for <code>context</code> in L19. The aggregator gains per-file timeouts so one slow file no longer blocks the rest.</p>
</div>
</div>
</div>

---

## What we'll cover

- **`select`** — multiplex N channels in one statement.
- **`time.After` + timeouts** — the canonical "time-bound this operation" pattern.
- **`time.NewTicker`** — periodic events. The heartbeat printer.
- **`default` branch** — non-blocking select; try-send and try-receive.
- **Cancellation via close-of-done-channel** — foundation for L19 `context`.

---

## Concept 1: `select` basics

### Motivation

In L16 you wrote single-channel reads: `r := <-resultsCh`. That blocks until a value arrives. But what if you're listening to MULTIPLE channels and want to react to whichever one is ready first?

That's what `select` is for. It's like a `switch` for channels.

---

### The basics

```go
select {
case v := <-chA:
    // chA had a value ready; v is it
case v := <-chB:
    // chB had a value ready first
case chOut <- 42:
    // chOut had a receiver ready, and we sent 42
}
```

Three rules to internalise:

- **Blocks until at least one case is ready.** If no channel is ready, select blocks (like a single `<-ch` would).
- **Randomly picks one if multiple are ready.** No priority. If both chA and chB have values, the runtime picks one at random. Don't rely on order.
- **Each case must be a channel operation.** Send (`ch <- v`), receive (`v := <-ch`), or receive-with-ok (`v, ok := <-ch`). No function calls, no arbitrary expressions.

---

### A worked example

Pick from two workers, whichever finishes first:

```go
chA := make(chan int, 1)
chB := make(chan int, 1)

go func() { chA <- expensiveA() }()
go func() { chB <- expensiveB() }()

select {
case v := <-chA:
    fmt.Println("A finished first:", v)
case v := <-chB:
    fmt.Println("B finished first:", v)
}
```

Whichever goroutine finishes first sends, the select fires that case, the other result is ignored (and the unread goroutine keeps holding its buffered value — that's a small leak we'll handle in concept 5 via cancellation).

---

### Common mistake

Assuming select cases have priority:

```go
// the wrong way
select {
case <-fastCh:
    handleFast()
case <-slowCh:
    handleSlow()
}
// Don't assume fastCh "wins" if both are ready — runtime picks randomly.
```

If you genuinely need priority, use TWO selects:

```go
// poll fast first; only fall through if it's not ready
select {
case <-fastCh:
    handleFast()
default:
    // (default makes this non-blocking — concept 4)
    select {
    case <-fastCh:
        handleFast()
    case <-slowCh:
        handleSlow()
    }
}
```

---

### Recap

- `select` multiplexes channel operations in one statement.
- Blocks until at least one case is ready.
- Randomly picks one if multiple ready. No priority.
- Each case must be a channel send or receive.

---

## Concept 2: `time.After` + timeouts

### Motivation

The single most common reason to use `select`: time-bound an operation. "Wait up to 5 seconds for this channel, then give up." That's the timeout pattern. It's how the L17 aggregator avoids being blocked forever by one slow file.

---

### The basics

`time.After(d)` returns a `<-chan time.Time` that fires once after duration `d`. Combine with `select`:

```go
select {
case v := <-workCh:
    handle(v)
case <-time.After(5 * time.Second):
    fmt.Println("timed out")
}
```

After 5 seconds with no value on `workCh`, the timeout case fires. The select returns; we move on.

---

### A worked example

The L17 aggregator's reduce loop:

```go
for len(pending) > 0 {
    var timeoutCh <-chan time.Time
    if timeout > 0 {
        timeoutCh = time.After(timeout)
    }
    select {
    case r := <-resultsCh:
        delete(pending, r.path)
        // ... merge r.counts
    case <-timeoutCh:
        // Pick any pending path, mark timed out, move on.
        for p := range pending {
            result.TimedOut = append(result.TimedOut, p)
            delete(pending, p)
            break
        }
    }
}
```

The `var timeoutCh <-chan time.Time` (nil) + conditional `if timeout > 0` trick: a nil channel in a select **never fires**. So `timeout=0` means "no timeout" — the select only has the result-receive case to choose from, effectively unconditional.

This is the textbook idiomatic way to do an optional timeout in Go.

---

### Common mistake

`time.After` allocates a fresh `*time.Timer` and channel every call. In a hot loop, that's GC pressure:

```go
// the wrong way (in a hot loop)
for {
    select {
    case v := <-ch:
        handle(v)
    case <-time.After(time.Second):  // new timer every iteration
        log("timed out")
    }
}
```

For hot loops, allocate once with `time.NewTimer` and reset:

```go
t := time.NewTimer(time.Second)
defer t.Stop()
for {
    t.Reset(time.Second)
    select {
    case v := <-ch:
        handle(v)
    case <-t.C:
        log("timed out")
    }
}
```

For non-hot paths (one-shot use), `time.After` is fine and idiomatic.

---

### Recap

- `time.After(d)` returns a `<-chan time.Time` firing once after d.
- Combined with `select`, gives timeouts.
- Nil-channel trick: `var ch <-chan time.Time` left nil means "no timeout" in select (nil channels never fire).
- For hot loops, prefer `time.NewTimer` + `Reset` to avoid allocations.

---

## Concept 3: `time.NewTicker`

### Motivation

Periodic events — heartbeats, polling, "every N seconds do X." Tickers fire repeatedly on a channel.

---

### The basics

```go
t := time.NewTicker(time.Second)
defer t.Stop()

for {
    select {
    case <-t.C:
        heartbeat()
    case <-done:
        return
    }
}
```

`t.C` is the channel; receives a `time.Time` every interval. **Always call `t.Stop()`** when done — otherwise the ticker keeps firing into an unread channel and the goroutine driving it leaks.

---

### A worked example

A hypothetical heartbeat printer for a long-running aggregator:

```go
go func() {
    t := time.NewTicker(2 * time.Second)
    defer t.Stop()
    for {
        select {
        case <-t.C:
            fmt.Printf("... processed %d / %d files\n", done, total)
        case <-cancelCh:
            return
        }
    }
}()
```

We don't add this to the L17 aggregator (scope creep) but it's the natural pattern for "give the user feedback during a long operation." L20's worker pool would be a good place to add it.

---

### Common mistake

Forgetting `t.Stop()`:

```go
// the wrong way
func work() {
    t := time.NewTicker(time.Second)   // never stopped
    for {
        <-t.C
        if done() { return }
    }
}
```

When `work()` returns, the ticker keeps running. It fires into `t.C` (which has a tiny buffer) and then sits there forever. Memory leak; goroutine leak. **Always `defer t.Stop()` right after creation.**

---

### Recap

- `time.NewTicker(d)` fires every d on `t.C`.
- **Always `defer t.Stop()`** — otherwise the ticker leaks.
- Use for periodic work: heartbeats, polling, status reports.
- For one-shot timing, prefer `time.After` (cleaner; ticker is overkill).

---

## Concept 4: `default` branch (non-blocking select)

### Motivation

Sometimes you want to TRY a channel operation — receive if a value is ready, send if a receiver is ready — but not block waiting. The `default` branch in select makes it non-blocking.

---

### The basics

```go
select {
case v := <-ch:
    handle(v)
default:
    // ch had nothing to give right now
}
```

`default` fires when no other case is immediately ready. The select returns instantly instead of blocking.

Common patterns:

- **Try-receive**: `select { case v := <-ch: ...; default: ... }` — receive if ready, else fall through.
- **Try-send**: `select { case ch <- v: ...; default: ... }` — send if receiver ready, else drop (or queue elsewhere).

---

### A worked example

Dropping log messages when the queue is full (back-pressure):

```go
func logAsync(msg string) {
    select {
    case logCh <- msg:
        // queued
    default:
        // queue full, drop the message
        atomic.AddInt64(&dropped, 1)
    }
}
```

When `logCh` is a buffered channel that's full, the send would normally block. With `default`, it falls through instead — we drop the message rather than blocking the caller.

---

### Common mistake

Using `default` in a busy loop:

```go
// the wrong way
for {
    select {
    case v := <-ch:
        handle(v)
    default:
        // burns CPU spinning
    }
}
```

Without `default`, the select would block efficiently until `ch` has a value. With `default`, the loop spins at 100% CPU. **If you mean to wait, omit `default`.**

`default` is for "try, don't wait" — single-shot, not in a loop.

---

### Recap

- `default` in select makes it non-blocking.
- Try-receive / try-send patterns.
- **Don't busy-loop** — if you want to wait, omit default.

---

## Concept 5: Cancellation via close-of-done-channel

### Motivation

When you spawn a long-running goroutine, you need a way to tell it to stop. The Go idiom: pass it a `done` channel; close the channel to signal cancellation.

This is the foundation for `context` (L19). `context.Context` is essentially a `done channel + cancellation API + value bag`. Today we learn the underlying pattern.

---

### The basics

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

// Later, when we want the goroutine to exit:
close(done)
```

Three things to internalise:

- **Closing is the signal.** A closed channel returns the zero value immediately on receive — every receiver sees it at once. Perfect for broadcast cancellation.
- **`chan struct{}`, not `chan bool`.** The value doesn't matter; we just need "is it closed." `struct{}` takes zero bytes.
- **Closer is whoever decides to cancel.** Not the receiver. The closer of `done` is the goroutine that wants others to stop.

---

### A worked example

A worker that exits cleanly when cancelled:

```go
func startWorker(done <-chan struct{}, in <-chan job) {
    go func() {
        for {
            select {
            case <-done:
                return    // cancellation signal received
            case j := <-in:
                process(j)
            }
        }
    }()
}

// Driver:
done := make(chan struct{})
startWorker(done, jobCh)
// ... do stuff
close(done)   // signals: all workers exit
```

If you have N workers reading from `done`, **closing it cancels them all simultaneously**. That's the broadcast property of close.

---

### Common mistake

**Sending on `done` instead of closing it.**

```go
// the wrong way
done := make(chan struct{})

// Send to one worker:
done <- struct{}{}  // only ONE worker receives

// With N workers, you'd need N sends — and each send blocks until
// some worker is ready. Race-prone, error-prone.
```

`close(done)` broadcasts to all receivers atomically. `done <- struct{}{}` only reaches one. **For cancellation, always close.**

The other common mistake: **closing twice panics**. If you might need to cancel from multiple places, wrap in `sync.Once` (L18 preview):

```go
var once sync.Once
cancel := func() { once.Do(func() { close(done) }) }
```

---

### Recap

- Done-channel pattern: pass `<-chan struct{}`; `close(done)` to signal cancel.
- `chan struct{}` because the value doesn't matter — only the close.
- Close broadcasts; send only reaches one receiver.
- Closing twice panics — wrap in `sync.Once` if multiple cancellers.
- **L19 generalizes this via `context.Context`** — the canonical Go cancellation API.

---

## Practice

### Warm-up

Implement `WaitWithTimeout[T any](ch <-chan T, d time.Duration) (T, error)` — block on ch, or return `(zero, ErrTimeout)` after d. Uses `select { case v := <-ch: ...; case <-time.After(d): ... }`.

```bash
cd lessons/17-select-timers/exercises/warmup/wait
go test -v
go test -race
```

---

### Main

The aggregator EVOLVES:

1. **`WalkResult` struct**: `{Counts map[string]int; TimedOut []string}` — caller sees both aggregate AND which files timed out.
2. **`Walk(dir string, timeout time.Duration) (WalkResult, error)`**: reduce loop uses `select` with `time.After(timeout)`. `timeout=0` means "no timeout" via the nil-channel trick.
3. **CLI gains `-timeout=<dur>` flag**: print timed-out paths to stderr; aggregated counts to stdout.

```bash
cd lessons/17-select-timers/exercises
go test ./...
go test -race ./...
```

Manual smoke test:

```bash
mkdir /tmp/logs
echo "2026-05-21T14:30:00 INFO ok" > /tmp/logs/a.log
go run ./cmd/aggregator -dir=/tmp/logs              # no timeout
go run ./cmd/aggregator -dir=/tmp/logs -timeout=1ns # everything times out
# stderr: "timeout: /tmp/logs/a.log"
```

Note:
Concept 5 (cancellation via close-of-done) is the foundation for L19's `context.Context`. When you read that lesson, notice that `context.Context.Done()` returns exactly the kind of channel we built here — the pattern lifts directly to the stdlib.

---

## Closing thought

`select` is a small statement with surprisingly large reach. Once you can write `select { case x: ...; case timeout: ... }`, you've covered the bulk of real-world concurrency: timeouts, cancellation, multiplexing, back-pressure. Lesson 19's `context.Context` packages the cancellation half into a standard interface; everything else stays as plain `select`.

---

## What we learned

- `select` multiplexes channel operations; blocks until one case is ready; random pick when multiple.
- `time.After(d)` for timeouts; nil-channel trick for "no timeout."
- `time.NewTicker(d)` for periodic events; always `defer t.Stop()`.
- `default` makes select non-blocking; don't busy-loop.
- Done-channel cancellation: `chan struct{}`; `close(done)` broadcasts to all receivers.

---

## Up next

Lesson 18 — **`sync` & memory model**. `sync.Mutex`, `sync.RWMutex`, `sync.WaitGroup`, `sync.Once`, `sync/atomic`. The aggregator's shared map gets refactored from channel-based to mutex-based — we'll compare the two approaches side-by-side. The race detector becomes a daily-habits tool (`make test-race`).
