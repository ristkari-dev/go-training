<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">19</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 3 — Concurrency &amp; Systems</div>
<h1>context</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>The standard library's cancellation interface. Learn the <code>context.Context</code> type, derivation via <code>WithCancel</code>/<code>WithTimeout</code>/<code>WithDeadline</code>, the <code>Done()</code>+<code>Err()</code> check pattern, and the "first parameter" propagation idiom. The aggregator gains ctx propagation through both Walk variants; the CLI handles SIGINT gracefully via <code>signal.NotifyContext</code>.</p>
</div>
</div>
</div>

---

## What we'll cover

- **`context.Context` interface + first-parameter propagation** — the type, the idiom.
- **Derivation** — `WithCancel`, `WithTimeout`, `WithDeadline`, and the ever-important `defer cancel()`.
- **`Done()` + `Err()` check pattern** — how cancellable goroutines listen for shutdown signals.
- **Common mistakes + `context.Value`** — the anti-patterns, the rarely-correct uses.

---

## The story so far

L17 introduced the **close-of-done-channel** pattern: spawn a goroutine, give it a `done chan struct{}`, close it when you want the goroutine to stop. That works. But every function that wants to be cancellable needs its own done-channel parameter; every caller needs to thread it through; there's no standard place for deadlines, timeouts, or request-scoped values.

L19's `context.Context` is the standard library's answer. **One interface; one parameter; one pattern.** Every cancellable function in Go takes a `ctx context.Context` as its first parameter. The function checks `ctx.Done()`; the caller cancels via `cancel()`. Done.

---

## Concept 1: `context.Context` + first-parameter propagation

### Motivation

You need a way to say "stop what you're doing" to a long-running operation. You need it to be consistent across packages — your code, the stdlib, third-party libraries. You need it to support deadlines, timeouts, and request-scoped cancellation without inventing a new pattern each time.

`context.Context` is that interface. It's tiny, it's universal, and it's the convention every cancellable function in Go follows.

---

### The basics

`context.Context` is an interface with four methods:

```go
type Context interface {
    Done() <-chan struct{}        // signals cancellation
    Err() error                   // why cancelled (or nil if still active)
    Deadline() (time.Time, bool)  // when ctx will fire (if set)
    Value(key any) any            // request-scoped data (use sparingly)
}
```

Two constructors give you root contexts (neither cancellable):

```go
context.Background()  // the canonical root; use in main, init, tests
context.TODO()        // "I'll figure out the right ctx later"
```

**The propagation idiom**: `ctx` is ALWAYS the first parameter (by convention, not enforcement). Functions pass it through to other functions:

```go
func handleRequest(ctx context.Context, req *Request) error {
    user, err := loadUser(ctx, req.UserID)    // ctx propagates
    if err != nil { return err }
    return saveAudit(ctx, user)               // ctx propagates again
}
```

When `ctx` is cancelled (somewhere up the stack), every function that's currently checking `ctx.Done()` learns about it at once.

---

### A worked example

The L19 aggregator's `Walk` propagates ctx through the reduce loop:

```go
func Walk(ctx context.Context, dir string, timeout time.Duration) (WalkResult, error) {
    // ... setup ...
    for len(pending) > 0 {
        select {
        case r := <-resultsCh:
            // process result
        case <-time.After(timeout):
            // file timed out
        case <-ctx.Done():
            return result, ctx.Err()    // ← cancellation
        }
    }
    return result, nil
}
```

When `ctx` is cancelled, the next iteration of the select sees `ctx.Done()` ready, returns the partial result + `ctx.Err()`.

---

### Common mistake

**Storing ctx in a struct.** Tempting but wrong:

```go
// the wrong way
type Worker struct {
    ctx context.Context   // ❌ ctx as struct field
}

func (w *Worker) Process(item Item) { ... }
```

Why wrong: `ctx` represents the lifetime of a CALL, not a long-lived object. A worker that processes multiple requests over its lifetime should accept ctx PER request, not store one ctx forever. The struct-field anti-pattern hides the cancellation propagation and makes lifetimes confusing.

The right shape:

```go
type Worker struct { /* no ctx field */ }

func (w *Worker) Process(ctx context.Context, item Item) { ... }
```

The `golangci-lint` `containedctx` analyzer catches this.

---

### Recap

- `context.Context` is a 4-method interface: `Done`, `Err`, `Deadline`, `Value`.
- `ctx` is ALWAYS the first parameter of a cancellable function.
- `context.Background()` for roots; `context.TODO()` while developing.
- **Don't store ctx in a struct** — it represents a call lifetime, not object lifetime.

---

## Concept 2: Derivation — `WithCancel`, `WithTimeout`, `WithDeadline`

### Motivation

The root contexts (`Background()`, `TODO()`) aren't cancellable — they never expire. To get a CANCELLABLE context, you derive a child from a parent:

```go
ctx, cancel := context.WithCancel(parent)
defer cancel()
```

The child inherits everything from the parent (values, deadlines), gains its own cancellation. Cancelling the parent also cancels the child; cancelling the child does NOT cancel the parent.

---

### The basics

Three derivation functions cover most needs:

```go
// Cancellable on-demand
ctx, cancel := context.WithCancel(parent)

// Cancels after a duration (relative)
ctx, cancel := context.WithTimeout(parent, 5*time.Second)

// Cancels at a specific time (absolute)
ctx, cancel := context.WithDeadline(parent, time.Now().Add(5*time.Second))
```

All three return a `cancel func()` that you should call when done, even if cancellation isn't needed:

```go
defer cancel()
```

The `defer cancel()` releases resources (timers, child-list bookkeeping) when the function returns. If you forget, you leak.

`WithTimeout` is sugar for `WithDeadline(parent, time.Now().Add(d))`.

---

### A worked example

The L19 CLI's main wraps the parent ctx with a signal-driven cancellation:

```go
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
    defer stop()

    if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
        // ...
    }
}
```

`signal.NotifyContext(parent, signals...)` is the modern (Go 1.16+) way to combine a parent ctx with OS signal handling. When SIGINT arrives, the returned ctx is cancelled.

The `defer stop()` removes the signal handler when main returns. Same pattern as `defer cancel()`.

---

### Common mistake

**Forgetting `defer cancel()`.**

```go
// the wrong way
ctx, _ := context.WithTimeout(parent, 5*time.Second)
doSomething(ctx)
// ❌ cancel never called → leak until the 5s timer fires
```

Even if your function returns before the timeout fires, you should `cancel()` so resources free immediately. `defer cancel()` is automatic:

```go
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel()
doSomething(ctx)
```

`go vet`'s `lostcancel` analyzer catches a missing cancel call when the cancel function escapes the scope without being called.

---

### Recap

- Derive cancellable children from parent contexts via `WithCancel`/`WithTimeout`/`WithDeadline`.
- **Always `defer cancel()`** — releases resources, even if cancellation isn't needed.
- `signal.NotifyContext` connects a parent ctx to OS signals.

---

## Concept 3: `Done()` + `Err()` check pattern

### Motivation

A long-running goroutine needs to know WHEN to stop. The `Done()` channel is the signal; `Err()` tells you WHY.

This is L17's done-channel pattern, generalized. Every cancellable function checks `ctx.Done()` periodically — between operations, in select cases, before expensive work.

---

### The basics

```go
select {
case work := <-workCh:
    process(work)
case <-ctx.Done():
    return ctx.Err()    // exit cleanly
}
```

`ctx.Done()` returns a `<-chan struct{}` that's closed when ctx is cancelled. `ctx.Err()` returns the reason:

| Reason | Returned by Err |
|---|---|
| `cancel()` was called | `context.Canceled` |
| Deadline / timeout passed | `context.DeadlineExceeded` |
| ctx not yet cancelled | `nil` |

Callers distinguish via `errors.Is`:

```go
if errors.Is(err, context.Canceled) { /* user cancelled */ }
if errors.Is(err, context.DeadlineExceeded) { /* timed out */ }
```

---

### A worked example

The L19 warmup `SleepWithCtx`:

```go
func SleepWithCtx(ctx context.Context, d time.Duration) error {
    select {
    case <-time.After(d):
        return nil              // slept the full duration
    case <-ctx.Done():
        return ctx.Err()        // cancelled before d elapsed
    }
}
```

Three lines. Returns `nil` on clean completion, `context.Canceled` or `context.DeadlineExceeded` on cancellation. Composes with `errors.Is` perfectly.

The aggregator's `Walk` does the same shape inside its reduce loop. Every cancellable function in your code should look similar.

---

### Common mistake

**Ignoring `Done()` in a hot loop:**

```go
// the wrong way
func process(ctx context.Context, items []Item) error {
    for _, item := range items {
        expensiveWork(item)   // ❌ never checks ctx.Done()
    }
    return nil
}
```

The function takes ctx but doesn't check it. The work continues even after ctx is cancelled — defeats the entire point. **Long loops should check Done() periodically**:

```go
func process(ctx context.Context, items []Item) error {
    for _, item := range items {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        expensiveWork(item)
    }
    return nil
}
```

The `default` branch makes the select non-blocking — it polls for cancellation but doesn't block waiting for it. Use this pattern when you can't naturally combine `ctx.Done()` with another channel in a blocking select.

---

### Recap

- `ctx.Done()` returns a channel that's closed when ctx is cancelled.
- `ctx.Err()` returns `context.Canceled` or `context.DeadlineExceeded`.
- Callers use `errors.Is` to distinguish reasons.
- **Always check Done() in long loops** — even via a non-blocking default-branch select if needed.

---

## Concept 4: Common mistakes + `context.Value`

### Motivation

Two more topics — both about what NOT to do, with one exception.

`context.Value` exists for request-scoped data (trace IDs, request loggers, authenticated user). Most uses are wrong. We'll cover when it's actually correct.

Plus a survey of the other anti-patterns Go developers hit when they first encounter context.

---

### The basics

**`context.Value(key any) any`** stores key/value pairs on a context. Children inherit values from parents:

```go
ctx := context.WithValue(parent, traceIDKey, "abc-123")
// Anywhere downstream:
id := ctx.Value(traceIDKey).(string)
```

Three rules for Value:

1. **Keys must be typed.** Use a package-private type to avoid collisions across packages:
   ```go
   type traceIDKey struct{}
   ctx = context.WithValue(parent, traceIDKey{}, "abc-123")
   ```
2. **Value is for request-scoped DATA, not parameters.** Trace IDs, request loggers, authenticated user — fine. Database connection, config, business inputs — pass as real parameters.
3. **It's untyped at the call site.** `ctx.Value(...).(string)` is a runtime type assertion; typos compile silently and panic at runtime.

---

### A worked example

The canonical correct use — a request-scoped logger:

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

Now any function with `ctx` can fetch the request's logger:

```go
func handleRequest(ctx context.Context, req *Request) error {
    log := LoggerFrom(ctx)
    log.Info("received request", "id", req.ID)
    // ...
}
```

The logger isn't a parameter to every function in your call tree, but it propagates anyway via ctx.

This works because logger is genuinely request-scoped (each request gets its own) and the alternative (logger parameter on every function) is much worse.

---

### The wrong way to use context.Value

```go
// ❌ Using ctx as a parameter-passing shortcut
type userIDKey struct{}
ctx = context.WithValue(parent, userIDKey{}, req.UserID)
// ... 5 calls later ...
id := ctx.Value(userIDKey{}).(string)
processUser(id)
```

`userID` is a parameter, not request-scoped data. Pass it as a real parameter to `processUser`. The Value detour:
- Loses type safety (silent runtime assertion)
- Hides the data flow (you have to grep for the key type)
- Makes refactoring brittle

**Rule of thumb**: if you'd write a regular parameter, write a regular parameter. `context.Value` is for things that propagate through DOZENS of unrelated functions (loggers, trace IDs) where adding a parameter would pollute every signature.

---

### Other common mistakes

**Passing nil context:**

```go
doSomething(nil)   // ❌ panics on first ctx.Done() call
```

Use `context.Background()` if you have no ctx; `context.TODO()` if you're not sure yet.

**Storing ctx in a struct** (from concept 1): ctx represents call lifetime, not object lifetime. Pass it as a method parameter instead.

**Cancelling parent from child:** child contexts can ONLY cancel themselves; cancelling the child does NOT cancel the parent. If you want parent cancellation, call the parent's `cancel()` directly.

**Forgetting `defer cancel()`:** every `WithCancel`/`WithTimeout`/`WithDeadline` call returns a cancel function. Always defer it.

---

### Recap

- `context.Value` is for REQUEST-SCOPED data (loggers, trace IDs), not parameter shortcuts.
- Use typed keys (`type fooKey struct{}`) to avoid collisions.
- Never pass nil ctx — use `context.Background()` or `context.TODO()`.
- Don't store ctx in a struct; pass it as a parameter.
- The `golangci-lint` analyzers `containedctx`, `lostcancel`, and `noctx` catch many of these.

---

## Practice

### Warm-up

Implement `SleepWithCtx(ctx context.Context, d time.Duration) error` in `exercises/warmup/sleepctx/sleepctx.go`. Three-line select: `time.After(d)` for clean completion, `ctx.Done()` for cancellation, `return ctx.Err()` in the cancellation case.

```bash
cd lessons/19-context/exercises/warmup/sleepctx
go test -v
go test -race
```

---

### Main

The aggregator EVOLVES again:

1. **`Walk(ctx, dir, timeout)`** — channel-based; reduce loop's select gains `case <-ctx.Done()`. Same for **`WalkLocked(ctx, dir, timeout)`** — outer goroutine's select gains the same case. After `wg.Wait()`, WalkLocked checks `ctx.Err()` FIRST (cancellation trumps file errors).
2. **`cmd/aggregator`** — `main()` uses `signal.NotifyContext(ctx, os.Interrupt)`. On `errors.Is(err, context.Canceled)`, print "cancelled by user" to stderr and exit 0 (NOT 1 — user cancellation isn't a failure).

```bash
cd lessons/19-context/exercises
go test ./...
go test -race ./...
make test-race      # daily habit (since L18)
```

Manual smoke test (run in a terminal where you can Ctrl-C):

```bash
mkdir /tmp/big-logs
for i in $(seq 1 1000); do
    echo "2026-05-21T14:30:00 INFO ok" > "/tmp/big-logs/f$i.log"
done

# Start the walk, press Ctrl-C mid-flight:
go run ./lessons/19-context/solutions/cmd/aggregator -dir=/tmp/big-logs
# ^C
# Output: partial counts to stdout, "cancelled by user" to stderr, exit 0
```

Note:
The aggregator now has THREE ways to stop early:
- Real error from a file (open or parse) — exit 1
- Per-file timeout (`-timeout=`) — skip that file, continue
- ctx cancellation (SIGINT or deadline) — return partial result + ctx.Err(), exit 0

All three coexist. The CLI distinguishes "user cancelled" (exit 0) from "something went wrong" (exit 1).

---

## Closing thought

Phase 3 has been building toward this lesson. L16 introduced goroutines. L17 added select + timers and the done-channel pattern. L18 added sync primitives + the race detector. L19 wraps cancellation into the standard library's `context.Context` interface — the canonical answer to "how do I tell a goroutine to stop?"

Every cancellable function in real Go code takes a ctx. Every standard library function that does I/O accepts one (`http.Request.Context()`, `database/sql.QueryContext`, `os/exec.CommandContext`). Once you've internalised the pattern, the entire stdlib reveals itself as ctx-propagating from end to end.

L20 (concurrency patterns) takes ctx as a given and builds on top of it — worker pools, fan-in/out, and the `errgroup` pattern (which combines ctx cancellation with first-error-wins, generalising our L18 `atomic.Pointer[error]` approach).

---

## What we learned

- `context.Context` is the standard cancellation interface — 4 methods, always the first parameter.
- Derive cancellable children via `WithCancel`/`WithTimeout`/`WithDeadline`. Always `defer cancel()`.
- Check `ctx.Done()` in long-running loops; return `ctx.Err()` on cancellation.
- `errors.Is(err, context.Canceled)` and `errors.Is(err, context.DeadlineExceeded)` distinguish reasons.
- `context.Value` for request-scoped DATA (logger, trace ID); not for parameters.
- Don't store ctx in a struct. Don't pass nil ctx. Always `defer cancel()`.

---

## Up next

Lesson 20 — **Concurrency patterns**. Worker pools, fan-in/fan-out, pipelines, and the **`errgroup`** pattern (first-error-wins with ctx cancellation). We'll generalise L18's `atomic.Pointer[error]` and L19's ctx propagation into the standard "manage N goroutines, cancel on first error" idiom.
