# Lesson 19: context

## What you'll learn

By the end of this lesson you can:

- Use `context.Context` as the first parameter of any cancellable function (the propagation idiom).
- Derive cancellable contexts via `context.WithCancel`, `WithTimeout`, `WithDeadline`, and always `defer cancel()`.
- Check `ctx.Done()` in long-running goroutines; return `ctx.Err()` on cancellation; distinguish `context.Canceled` from `context.DeadlineExceeded` via `errors.Is`.
- Recognise when `context.Value` is appropriate (request-scoped data like loggers and trace IDs) and when it's not (parameter shortcuts).
- Wire OS signal handling to ctx cancellation via `signal.NotifyContext`.

## What's different from L18

L18 added `WalkLocked` alongside `Walk` and made `make test-race` a daily habit. **L19 propagates `ctx context.Context` through both Walk variants** and wires SIGINT to ctx cancellation in the CLI.

- Walk's reduce loop's select gains `case <-ctx.Done()`.
- WalkLocked's outer worker goroutine's select gains the same case. After `wg.Wait()`, the function checks `ctx.Err()` FIRST (cancellation trumps file errors).
- `cmd/aggregator`'s `main()` uses `signal.NotifyContext(ctx, os.Interrupt)` so Ctrl-C cancels the walk gracefully. On `errors.Is(err, context.Canceled)`, the CLI prints partial output + "cancelled by user" to stderr and exits 0 (user cancellation isn't a failure).
- `processFile` and `processFileForLocked` are unchanged — they don't take ctx. The file work is fast (~100µs); cancellation happens at the coordination layer (reduce loop / outer worker), where it has meaningful impact.

## The package layout

```
lessons/19-context/exercises/
├── warmup/sleepctx/                 ← you implement: SleepWithCtx
│   ├── sleepctx.go
│   └── sleepctx_test.go
├── cmd/aggregator/                   ← evolves: SIGINT + ctx propagation
│   ├── main.go                        signal.NotifyContext; cancel-as-success
│   └── main_test.go                   + TestAggregatorCancelled
└── internal/
    ├── logparse/                      verbatim from L18
    └── aggregator/                    ← evolves: Walk + WalkLocked both get ctx
        ├── aggregator.go              ctx is new first parameter
        └── aggregator_test.go         + 2 ctx test cases via runWalkSuite
```

---

## Concept 1 — `context.Context` + first-parameter propagation

`context.Context` is a 4-method interface:

```go
type Context interface {
    Done() <-chan struct{}        // signals cancellation
    Err() error                   // why cancelled
    Deadline() (time.Time, bool)  // when ctx will fire (if set)
    Value(key any) any            // request-scoped data
}
```

Two root constructors:

```go
context.Background()  // canonical root; use in main, init, tests
context.TODO()        // "I'll figure out the right ctx later"
```

**The propagation idiom**: ctx is ALWAYS the first parameter:

```go
func handleRequest(ctx context.Context, req *Request) error {
    user, err := loadUser(ctx, req.UserID)    // ctx propagates
    if err != nil { return err }
    return saveAudit(ctx, user)               // ctx propagates again
}
```

When ctx is cancelled (somewhere up the stack), every function checking `ctx.Done()` learns about it at once.

### Common mistake

**Storing ctx in a struct.** Tempting but wrong — ctx represents a CALL lifetime, not an object lifetime. A worker that processes multiple requests should accept ctx PER request:

```go
// the wrong way
type Worker struct { ctx context.Context }     // ❌

// the right way
type Worker struct { /* no ctx field */ }
func (w *Worker) Process(ctx context.Context, item Item) { ... }
```

The `golangci-lint` `containedctx` analyzer catches this.

---

## Concept 2 — Derivation: `WithCancel`, `WithTimeout`, `WithDeadline`

Root contexts (`Background`, `TODO`) aren't cancellable. To get cancellation, derive a child:

```go
// On-demand cancellation
ctx, cancel := context.WithCancel(parent)
defer cancel()

// Cancels after a duration
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel()

// Cancels at a specific time
ctx, cancel := context.WithDeadline(parent, time.Now().Add(5*time.Second))
defer cancel()
```

All three return a `cancel func()`. **Always `defer cancel()`** to release resources (timers, child-list bookkeeping), even if cancellation isn't needed in your specific flow.

The CLI uses `signal.NotifyContext` (Go 1.16+) to combine ctx with OS signals:

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
defer stop()
```

When SIGINT arrives, ctx is cancelled. `defer stop()` removes the signal handler.

### Common mistake

**Forgetting `defer cancel()`.** Even if your function returns before the timeout fires, you should `cancel()` so resources free immediately:

```go
// the wrong way
ctx, _ := context.WithTimeout(parent, 5*time.Second)   // ❌ cancel dropped
doSomething(ctx)

// the right way
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel()
doSomething(ctx)
```

`go vet`'s `lostcancel` analyzer catches this.

---

## Concept 3 — `Done()` + `Err()` check pattern

Long-running goroutines listen for cancellation via the `Done()` channel:

```go
select {
case work := <-workCh:
    process(work)
case <-ctx.Done():
    return ctx.Err()     // exit cleanly
}
```

`ctx.Done()` returns a `<-chan struct{}` that's closed when ctx is cancelled. `ctx.Err()` returns the reason:

| Reason | `ctx.Err()` returns |
|---|---|
| `cancel()` was called | `context.Canceled` |
| Timeout / deadline passed | `context.DeadlineExceeded` |
| ctx not yet cancelled | `nil` |

Callers distinguish via `errors.Is`:

```go
if errors.Is(err, context.Canceled) { /* user cancelled */ }
if errors.Is(err, context.DeadlineExceeded) { /* timed out */ }
```

The L19 warmup `SleepWithCtx` is the textbook example — three lines:

```go
func SleepWithCtx(ctx context.Context, d time.Duration) error {
    select {
    case <-time.After(d):
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

### Common mistake

**Ignoring Done() in a hot loop.** A function that takes ctx but doesn't check it defeats the entire point. For long loops, even a non-blocking poll works:

```go
for _, item := range items {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    expensiveWork(item)
}
```

The `default` branch makes the select non-blocking — polls for cancellation between iterations.

---

## Concept 4 — Common mistakes + `context.Value`

**`context.Value`** stores key/value pairs on a context. Children inherit values from parents. Useful for genuinely request-scoped data:

```go
type loggerKey struct{}

func WithLogger(ctx context.Context, log *slog.Logger) context.Context {
    return context.WithValue(ctx, loggerKey{}, log)
}

func LoggerFrom(ctx context.Context) *slog.Logger {
    if log, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
        return log
    }
    return slog.Default()
}
```

Three rules:

1. **Typed keys.** Use a package-private type (`type fooKey struct{}`) to avoid collisions across packages.
2. **For DATA, not parameters.** Loggers, trace IDs, authenticated user — fine. Database connection, business inputs — pass as real parameters.
3. **Untyped at the call site.** `ctx.Value(...).(*Logger)` is a runtime type assertion; typos panic at runtime, not compile time.

### When to use Value (correctly)

The bar: if you'd add a parameter to EVERY function in your call tree (logger, trace ID, request-scoped auth), Value is justified. If it's data 2-3 functions care about, just pass it as a parameter.

### Common mistakes

- **Passing nil ctx:** `doSomething(nil)` panics on first `ctx.Done()` call. Use `context.Background()` or `context.TODO()`.
- **Storing ctx in a struct** (see concept 1).
- **Cancelling parent from child:** children can only cancel themselves. Call the parent's `cancel()` directly.
- **Forgetting `defer cancel()`** (see concept 2).
- **Using `context.Value` as a parameter shortcut:** prefer real parameters when the data isn't request-scoped.

The `golangci-lint` analyzers `containedctx`, `lostcancel`, and `noctx` catch many of these.

---

## Exercise: warm-up — `sleepctx`

Implement `SleepWithCtx(ctx context.Context, d time.Duration) error` in `exercises/warmup/sleepctx/sleepctx.go`. Three-line select: `time.After(d)` for clean completion, `ctx.Done()` for cancellation, return `ctx.Err()` on cancellation. Tests cover clean completion, pre-cancelled ctx, cancellation mid-sleep, and deadline expiration.

**Time:** 5-10 minutes.

## Exercise: main — aggregator with ctx + CLI SIGINT handler

Three pieces:

1. **`internal/aggregator/Walk(ctx, dir, timeout)`** — same channel-based logic from L18 but reduce loop's select adds `case <-ctx.Done(): return result, ctx.Err()`.
2. **`internal/aggregator/WalkLocked(ctx, dir, timeout)`** — same mutex+WaitGroup logic but outer goroutine's select adds the same case. After `wg.Wait()`, check `ctx.Err()` FIRST (cancellation has priority over `firstErr`).
3. **`cmd/aggregator/main.go`** — wrap call in `signal.NotifyContext(ctx, os.Interrupt)`. `run(ctx, args, stdout, stderr) error`. On `errors.Is(err, context.Canceled)`, print "cancelled by user" to stderr and exit 0 (not 1).

The aggregator tests parameterize via `runWalkSuite(t, fn)` (carried from L18) plus 2 new ctx sub-tests: `ctx-cancelled-before-walk` (pre-cancelled → context.Canceled) and `ctx-deadline-exceeded` (1µs timeout via context.WithTimeout → context.DeadlineExceeded).

**Time:** 30-45 minutes.

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
cd lessons/19-context/exercises
go test ./...

# Reference solution
go test -v ./lessons/19-context/solutions/...

# Race detector
go test -race ./lessons/19-context/solutions/...
make test-race

# The binary
mkdir /tmp/logs
echo "2026-05-21T14:30:00 INFO ok" > /tmp/logs/a.log
echo "2026-05-21T14:31:00 WARN slow" > /tmp/logs/b.log

# Normal run (channel mode)
go run ./lessons/19-context/solutions/cmd/aggregator -dir=/tmp/logs

# Mutex mode
go run ./lessons/19-context/solutions/cmd/aggregator -dir=/tmp/logs -mode=mutex

# SIGINT demo (Ctrl-C during a long walk)
mkdir /tmp/big-logs
for i in $(seq 1 1000); do
    echo "2026-05-21T14:30:00 INFO ok" > /tmp/big-logs/f$i.log
done
go run ./lessons/19-context/solutions/cmd/aggregator -dir=/tmp/big-logs
# Press Ctrl-C
# Output: partial counts on stdout, "cancelled by user" on stderr, exit 0
```

## Going further

### Read

- **Go blog — "Go Concurrency Patterns: Context"** (Sameer Ajmani, 2014): <https://go.dev/blog/context> — the canonical introduction. Still authoritative.
- **`context` package docs**: <https://pkg.go.dev/context> — short, dense, worth a careful read.
- **`signal.NotifyContext` docs**: <https://pkg.go.dev/os/signal#NotifyContext> — Go 1.16+ pattern for signal handling.

### Try

- **Whole-Walk timeout via ctx:** instead of (or in addition to) the per-file `-timeout` flag, add `-deadline=<dur>` that creates a `context.WithTimeout(ctx, dur)` wrapping the call. What changes? When would a user prefer per-file timeout vs whole-walk deadline?
- **errgroup preview:** in WalkLocked, the `atomic.Pointer[error]` first-error-wins pattern + ctx propagation is exactly what `golang.org/x/sync/errgroup` provides. Read its source (or implement a tiny local version) and rewrite WalkLocked using it. L20 formalises this.
- **Trace IDs via context.Value:** add a request-ID generator at the CLI entry point. Store a unique ID in ctx via `context.WithValue`. Print it in each `timeout:` line and error message. Demonstrates the one legitimate use of Value.

---

> Next stop: Lesson 20 — Concurrency patterns. Worker pools, fan-in/fan-out, pipelines, and the canonical errgroup pattern (first-error-wins + ctx cancellation). We'll generalise L18's atomic.Pointer[error] and L19's ctx propagation into a tiny lesson-local `errgroupx` package.
