<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">24</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 4 — Production &amp; Distributed</div>
<h1>HTTP clients &amp; resilience</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Call services that fail, hang, and rate-limit — and survive. Learn the <code>http.Client</code> (timeouts, transport reuse), per-request <code>context</code> deadlines, retries with exponential backoff + jitter, <strong>selective</strong> (transient-only) retry, and the <strong>circuit breaker</strong>. Build a <code>logstats-client</code> that resiliently ships log lines to the L23 server.</p>
</div>
</div>
</div>

---

## What we'll cover

- **`http.Client` + Timeout + Transport reuse** — never call out with no timeout; reuse the client for connection pooling.
- **Per-request context deadlines** — `NewRequestWithContext`; cancellation propagates to the in-flight request.
- **Retries + exponential backoff + jitter** — back off so you don't hammer a struggling server; jitter to avoid thundering herds.
- **Selective retry + idempotency** — retry transient failures, never 4xx; only retry idempotent operations.
- **Circuit breaker** — fail fast when a dependency is down so it can recover.

---

## The story so far — the other side of the wire

L23 built the `logstats` **server**: `POST /ingest`, `GET /stats`. L24 builds the **client** — a `logstats-client` that ships log lines to that server.

Writing a client sounds trivial: `http.Get`, done. But production clients call services across a network that drops packets, servers that restart mid-request, load balancers that return 503, and rate limiters that return 429. A naive client hangs forever, retries a doomed request into a storm, or turns a one-second blip into a cascading outage.

This lesson is the discipline of calling things that fail. Every concept is a layer of armor: timeouts so you don't hang, retries so a transient blip recovers, backoff so retries don't pile on, selectivity so you don't retry the un-retryable, and a circuit breaker so a down dependency fails fast instead of dragging you down with it.

---

## Concept 1: `http.Client` + Timeout + Transport reuse

### Motivation

The single most common Go production bug: `http.Get(url)` with no timeout. `http.DefaultClient` has **no timeout** — if the server accepts the connection and then hangs, your goroutine blocks forever, leaking memory and file descriptors until the process dies. Every real client sets a timeout.

---

### The basics

```go
client := &http.Client{
    Timeout: 10 * time.Second, // whole-request budget: dial + write + read
}
resp, err := client.Do(req)
```

`Client.Timeout` covers the entire exchange — connection, sending the request, and reading the *whole* body. When it fires, the in-flight request is cancelled and `Do` returns an error.

**Reuse the client.** An `http.Client` wraps an `http.Transport`, which pools and reuses TCP connections (keep-alive). Create one client and share it:

```go
// ❌ a fresh client per request — no pooling, a new TCP+TLS handshake
// every time, and file-descriptor churn under load.
for _, req := range reqs {
    (&http.Client{}).Do(req)
}

// ✓ one shared client — connections are pooled and reused.
client := &http.Client{Timeout: 10 * time.Second}
for _, req := range reqs {
    client.Do(req)
}
```

---

### A worked example

The shipper builds one client at construction and reuses it for every `Ship`:

```go
func New(url string, opts ...Option) *Shipper {
    s := &Shipper{
        url:    url,
        client: &http.Client{Timeout: 10 * time.Second},
        // ...
    }
    for _, o := range opts {
        o(s)
    }
    return s
}
```

A long-running shipper makes thousands of `/ingest` calls over a handful of pooled connections — no per-request handshake tax.

---

### Common mistake

**No timeout** (`http.Get`, `http.DefaultClient`, or `&http.Client{}` with `Timeout` unset). A hung server hangs your client forever. Always set `Timeout`. **And: a fresh client per request** — you lose connection pooling and churn file descriptors; under load you'll hit `too many open files` (the FD ceiling from L21). One client, reused.

---

### Recap

- `http.DefaultClient` has no timeout — never use it for external calls.
- `Client.Timeout` is the whole-request budget (dial → write → read body).
- Reuse one `http.Client`/`Transport` for connection pooling.

---

## Concept 2: Per-request context deadlines

### Motivation

`Client.Timeout` is a fixed per-request ceiling. But the *caller* often has its own deadline — an incoming request with 2 seconds left, a user who hit Ctrl-C, a parent operation that was cancelled. `context` carries that deadline/cancellation down into the HTTP call so the in-flight request aborts the moment the caller gives up.

---

### The basics

```go
req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
if err != nil { return err }
resp, err := client.Do(req)
```

When `ctx` is cancelled or its deadline passes, the in-flight request is torn down and `Do` returns `ctx.Err()` (wrapped). This composes with `Client.Timeout` — whichever fires first wins.

```go
ctx, cancel := context.WithTimeout(parentCtx, 2*time.Second)
defer cancel()
// this request can take at most 2s, regardless of the client's 10s timeout
```

---

### A worked example

The shipper threads the caller's `ctx` through every attempt, so cancelling the client's `run` ctx (e.g. on SIGINT) immediately stops an in-flight ship — and aborts any pending backoff sleep:

```go
func (s *Shipper) doOnce(ctx context.Context, body []byte) (Result, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(body))
    // ...
    resp, err := s.client.Do(req)
    // ...
}
```

And the retry loop selects on `ctx.Done()` during backoff — so a cancellation doesn't wait out the sleep.

---

### Common mistake

**Relying only on `Client.Timeout` and ignoring caller cancellation.** If you build requests without a context (`http.NewRequest`, no `...WithContext`), then when the caller is cancelled your request keeps running to its own 10s timeout — wasting work nobody's waiting for. Always use `NewRequestWithContext` and propagate the caller's `ctx`.

---

### Recap

- `http.NewRequestWithContext(ctx, ...)` ties the request to the caller's deadline/cancellation.
- It composes with `Client.Timeout` — first to fire wins.
- Propagate `ctx` everywhere, including across backoff waits.

---

## Concept 3: Retries + exponential backoff + jitter

### Motivation

Networks blip. A single 503 or dropped connection often succeeds on the next try. So retry — but *how* you retry matters enormously. Retry instantly in a tight loop and you hammer a server that's already struggling. Retry on a fixed 1-second interval across a thousand clients and they all retry in lockstep — a **thundering herd** that DDoSes your own backend. The fix is exponential backoff with jitter.

---

### The basics

```go
for n := 0; n < attempts; n++ {
    err = fn()
    if err == nil { return nil }
    backoff := base << n                 // base * 2^n: 100ms, 200ms, 400ms, ...
    d := time.Duration(rand.Int64N(int64(backoff) + 1))  // full jitter: [0, backoff]
    select {
    case <-ctx.Done(): return ctx.Err()  // abort the wait if cancelled
    case <-time.After(d):
    }
}
```

- **Exponential** — each wait doubles, so a persistent problem backs off fast instead of pounding.
- **Jitter** — randomizing the wait spreads retries across time so N clients don't synchronize. "Full jitter" (sleep a random duration in `[0, backoff]`) is the simplest effective form.
- **ctx-aware** — never sleep through a cancellation.

---

### A worked example

The L24 warm-up is exactly this loop, with a `Permanent` escape hatch (Concept 4):

```go
func Do(ctx context.Context, attempts int, base time.Duration, fn func() error) error {
    var err error
    for n := 0; n < attempts; n++ {
        err = fn()
        if err == nil { return nil }
        var p *permanent
        if errors.As(err, &p) { return p.err } // don't retry permanent failures
        if n == attempts-1 { break }
        backoff := base << n
        d := time.Duration(rand.Int64N(int64(backoff) + 1))
        t := time.NewTimer(d)
        select {
        case <-ctx.Done(): t.Stop(); return ctx.Err()
        case <-t.C:
        }
    }
    return err
}
```

---

### Common mistake

**Retrying with no backoff** — a `for { if err := call(); err == nil { break } }` loop turns one failed request into thousands per second, kicking a server while it's down. Always back off, and add jitter. (Secondary mistake: sleeping with `time.Sleep` instead of a ctx-aware `select`, so a cancelled caller still waits out the backoff.)

---

### Recap

- Retry transient failures, but back off **exponentially** (`base·2ⁿ`).
- Add **jitter** so clients don't synchronize into a thundering herd.
- Make the backoff **ctx-aware** — abort the wait on cancellation.

---

## Concept 4: Selective retry + idempotency

### Motivation

Not every failure should be retried. A 503 is worth retrying — the server is momentarily overloaded. A **400 Bad Request** is not — your request is malformed; retrying it a thousand times just wastes effort and never succeeds. And even for retryable failures, you may only retry **idempotent** operations — ones safe to repeat.

---

### The basics

Classify the outcome, then decide:

| Outcome | Class | Retry? |
|---|---|---|
| network error (dial/reset/timeout) | transient | ✓ |
| 5xx (server error) | transient | ✓ |
| 429 (too many requests) | transient | ✓ |
| 4xx other than 429 (400, 404, 422…) | permanent | ✗ |
| 2xx | success | — |

The `retry.Permanent(err)` wrapper signals "stop now" to the retry loop:

```go
var p *permanent
if errors.As(err, &p) { return p.err } // terminal — Do returns immediately
```

**Idempotency:** retrying a `GET` or our additive `/ingest` is safe. Retrying a non-idempotent `POST /charge-card` could double-charge. Only retry operations whose repetition is harmless — or make them idempotent first (the idempotency-key pattern, which L29 builds).

---

### A worked example

The shipper classifies in `doOnce` and acts in `Ship`:

```go
switch {
case resp.StatusCode >= 200 && resp.StatusCode < 300:
    return r, nil
case resp.StatusCode == http.StatusTooManyRequests:
    return Result{}, &transientError{status: resp.StatusCode} // 429 → retry
case resp.StatusCode >= 500:
    return Result{}, &transientError{status: resp.StatusCode} // 5xx → retry
default: // 4xx
    return Result{}, &permanentError{status: resp.StatusCode} // → stop
}
```

`Ship` turns a `*permanentError` into `retry.Permanent(...)` so the loop stops at one attempt — a 400 hits the server **once**, not four times.

---

### Common mistake

**Retrying a 400** (or any 4xx). It will never succeed and just amplifies load. Retry only transient classes. **And: retrying a non-idempotent write** — if `POST` creates a resource and you retry after a timeout, you may create it twice. Know whether your operation is safe to repeat before you wrap it in retries.

---

### Recap

- Retry transient failures (network/5xx/429); treat 4xx as permanent.
- `retry.Permanent` stops the loop immediately.
- Only retry **idempotent** operations (or add an idempotency key — L29).

---

## Concept 5: Circuit breaker

### Motivation

Retries help with brief blips. But what if a dependency is *hard down* — a database outage, a crashed service? Every request now waits for its full timeout, then retries with backoff, across every client and goroutine. You've turned a fast failure into a slow one, exhausted your own resources waiting, and you keep pounding a service that needs room to recover. The circuit breaker is the fix: after enough consecutive failures, **stop calling** and fail instantly.

---

### The basics

Three states, like an electrical breaker:

- **Closed** (normal) — calls pass through; count consecutive failures.
- **Open** (tripped) — after `maxFailures`, calls fail **instantly** with `ErrOpen` (no network I/O) for a `cooldown` period.
- **Half-open** (testing) — after the cooldown, allow **one** trial call. Success → Closed (recovered); failure → Open again (still down).

```go
func (b *Breaker) Call(fn func() error) error {
    if !b.allow() {        // open + within cooldown → fail fast
        return ErrOpen
    }
    err := fn()
    b.record(err)          // success closes + resets; failure may open
    return err
}
```

---

### A worked example

The shipper wraps each attempt in the breaker — and crucially, **counts only transient failures**:

```go
callErr := s.breaker.Call(func() error {
    res, e := s.doOnce(ctx, body)
    if e != nil {
        var p *permanentError
        if errors.As(e, &p) {
            permanent = e
            return nil // a 4xx is OUR bad request, not the dependency being down
        }
        return e // transient → the breaker counts it
    }
    result = res
    return nil
})
```

A 400 must not trip the breaker (the dependency is healthy; *we* sent garbage). Only 5xx/429/network failures count toward opening it. When the breaker opens, the retry loop sees `ErrOpen` and stops immediately — no backoff storm against a down service.

For testing, the breaker takes an **injected clock** (`WithClock`) so cooldown transitions are deterministic — no `time.Sleep` in tests.

---

### Common mistake

**No breaker at all.** Under a real outage, retries + backoff across every client become a self-inflicted DDoS that prevents the dependency from recovering ("retry storm" / "metastable failure"). A breaker turns the outage into instant local failures, sheds load off the struggling service, and probes for recovery. **Secondary mistake:** tripping the breaker on 4xx — client errors aren't the dependency's fault and shouldn't open the circuit.

---

### Recap

- Closed → Open (after N consecutive failures) → Half-open (after cooldown) → Closed/Open.
- Open = fail fast with `ErrOpen`, no I/O — gives the dependency room to recover.
- Count only **transient** failures; never trip on 4xx.
- Inject the clock for deterministic tests.

---

## Practice

### Warm-up

In `exercises/warmup/retry/`, implement `Do(ctx, attempts, base, fn)` — the exponential-backoff-with-jitter retry loop that honors `ctx.Done()` and stops early on `retry.Permanent`.

```bash
cd lessons/24-http-client/exercises
go test ./warmup/retry/...
```

### Main

1. **`internal/breaker`** — the closed/open/half-open state machine (`Call`/`allow`/`record`), clock injected via `WithClock`.
2. **`internal/shipper`** — `Ship` wiring `retry.Do` + `breaker.Call` + `doOnce` with selective retry (the breaker counts only transient failures; 4xx and `ErrOpen` stop the loop).
3. **`cmd/logstats-client`** (provided) — study how `run` wires stdin → shipper → summary.

```bash
cd lessons/24-http-client/exercises
go test ./...
go test -race ./...
make test-race      # daily habit (since L18)

# end-to-end against the carried server
go run ./cmd/logstats-server :8080 &
printf '2026-01-02T15:04:05 INFO ok\nbad\n' | go run ./cmd/logstats-client -addr=http://localhost:8080/ingest
curl -s localhost:8080/stats; echo
```

---

## Closing thought

A client that *works* is easy. A client that **keeps working when the network and the server don't** is the job. Every layer here exists because someone's pager went off: timeouts (a hung dependency took down the fleet), backoff + jitter (synchronized retries DDoSed the backend), selective retry (a 400 retried a million times), the circuit breaker (a database outage cascaded into total failure because everyone kept waiting and retrying).

None of it needed a library — `net/http`, `context`, and `time` gave us everything, plus ~150 lines of our own retry and breaker. The real libraries you'll use in production (`sony/gobreaker`, cloud SDKs' built-in retryers) implement exactly these patterns. Now you know what's inside them, and why each piece is there.

---

## What we learned

- **`http.Client`:** always set `Timeout`; reuse one client for connection pooling. `http.DefaultClient` has no timeout.
- **Context deadlines:** `NewRequestWithContext` ties requests to the caller's cancellation; compose with `Client.Timeout`.
- **Retries:** exponential backoff (`base·2ⁿ`) + jitter, ctx-aware waits — no thundering herds.
- **Selective retry:** transient (network/5xx/429) → retry; 4xx → permanent; only retry idempotent ops.
- **Circuit breaker:** closed/open/half-open; fail fast when a dependency is down; count only transient failures; inject the clock to test.

---

## Up next

Lesson 25 — **gRPC**. We give the `logstats` service a typed, streaming RPC surface: protobuf message + service definitions, a client-streaming `Ingest`, a unary `GetStats`, and interceptors (gRPC's middleware). It's also the course's **first third-party dependency** — `google.golang.org/grpc` + `google.golang.org/protobuf` — introduced deliberately, because you can't honestly teach gRPC without it.
