# Lesson 24: HTTP clients & resilience

## What you'll learn

By the end of this lesson you can:

- Configure an **`http.Client`** with an explicit timeout and reuse it for connection pooling.
- Propagate **per-request `context` deadlines** so in-flight requests abort when the caller gives up.
- Retry transient failures with **exponential backoff + jitter**, honoring `ctx.Done()`.
- Apply **selective retry** — retry network/5xx/429, treat 4xx as permanent — and respect **idempotency**.
- Build a **circuit breaker** (closed/open/half-open) that fails fast when a dependency is down.

This is the second lesson of **Phase 4**. We build the resilient `logstats-client` that ships log lines to the L23 server.

## What's different from L23

L23 built the `logstats` **server**. L24 is the other side of the wire: a **client** that calls it — and survives when the server fails, hangs, or rate-limits. The L23 `logparse`/`logstats`/`logstats-server` are carried forward verbatim (the client's real target); the new work is `warmup/retry`, `internal/breaker`, `internal/shipper`, and `cmd/logstats-client`.

Still **zero third-party dependencies** — `net/http` + `context` + `time` + ~150 lines of our own retry and breaker. (gRPC and OpenTelemetry, Phase 4's first real deps, arrive at L25 and L28.)

## The package layout

```
lessons/24-http-client/{exercises,solutions}/
├── warmup/retry/           ← you implement: Do(ctx, attempts, base, fn) — backoff + jitter + Permanent
├── internal/
│   ├── logparse/           carried from L23 (the server parses with it)
│   ├── logstats/           carried from L23 (the server's accumulator)
│   ├── breaker/            ← you implement: circuit breaker (closed/open/half-open)
│   └── shipper/            ← you implement: resilient client (timeout + selective retry + breaker)
└── cmd/
    ├── logstats-server/    carried from L23 (the client's target)
    └── logstats-client/    provided: reads stdin → shipper → prints summary
```

## The shipper's behavior

| Outcome | Class | Retried? | Trips breaker? |
|---|---|---|---|
| network error (dial/reset/timeout) | transient | ✓ (with backoff) | ✓ |
| 5xx server error | transient | ✓ | ✓ |
| 429 too many requests | transient | ✓ | ✓ |
| 4xx (400/404/422…) | permanent | ✗ (fail at once) | ✗ |
| 2xx | success | — | resets it |
| breaker open | — | ✗ (fail fast, `ErrOpen`) | — |

---

## Concept 1 — `http.Client` + Timeout + Transport reuse

`http.DefaultClient` (and `http.Get`) have **no timeout** — a hung server blocks your goroutine forever. Always set one, and reuse a single client so its `http.Transport` pools TCP connections:

```go
client := &http.Client{Timeout: 10 * time.Second} // whole-request budget: dial + write + read body
resp, err := client.Do(req)
```

The shipper builds one client in `New` and reuses it for every `Ship` — thousands of `/ingest` calls over a handful of pooled connections.

### Common mistake

No timeout (a hung server hangs you forever), or a fresh `&http.Client{}` per request (no pooling, file-descriptor churn → eventually `too many open files`, the L21 FD ceiling). One client, with a timeout, reused.

---

## Concept 2 — Per-request context deadlines

`Client.Timeout` is a fixed ceiling; the *caller* often has its own deadline or gets cancelled. `context` carries that into the request:

```go
req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
resp, err := client.Do(req) // aborts the moment ctx is cancelled / its deadline passes
```

It composes with `Client.Timeout` — whichever fires first wins. The shipper threads `ctx` through every attempt and selects on `ctx.Done()` during backoff, so a SIGINT-cancelled client stops immediately instead of waiting out a sleep.

### Common mistake

Relying only on `Client.Timeout` and building requests with `http.NewRequest` (no context). When the caller is cancelled, the request keeps running to its own timeout — wasted work nobody's waiting for. Always `NewRequestWithContext`.

---

## Concept 3 — Retries + exponential backoff + jitter

A transient blip (a 503, a dropped connection) often succeeds on retry. But retry *carefully*:

```go
backoff := base << n                                // base * 2^n: 100ms, 200ms, 400ms...
d := time.Duration(rand.Int64N(int64(backoff) + 1)) // full jitter: random in [0, backoff]
select {
case <-ctx.Done(): return ctx.Err()                // abort the wait on cancellation
case <-time.After(d):
}
```

- **Exponential** so a persistent problem backs off fast instead of pounding.
- **Jitter** so N clients don't retry in lockstep (a thundering herd that DDoSes your own backend).
- **ctx-aware** so a cancelled caller never waits out the sleep.

The warm-up `retry.Do` is exactly this loop.

### Common mistake

Retrying with no backoff — `for { if call() == nil { break } }` turns one failure into thousands of requests per second, kicking a server while it's down. Always back off + jitter; never `time.Sleep` through a cancellation.

---

## Concept 4 — Selective retry + idempotency

Not every failure is retryable. A 503 is (server overloaded); a **400** is not (your request is malformed — retrying never succeeds). The shipper classifies the outcome and only retries transient classes; a `retry.Permanent(err)` wrapper stops the loop at once:

```go
var p *permanent
if errors.As(err, &p) { return p.err } // terminal — Do returns immediately
```

**Idempotency:** retrying a `GET` or our additive `/ingest` is safe; retrying `POST /charge-card` could double-charge. Only retry operations safe to repeat — or make them idempotent first (the idempotency-key pattern in L29).

### Common mistake

Retrying a 400 (wasted work that never succeeds, amplifying load), or retrying a non-idempotent write (duplicate side effects). Classify before you retry.

---

## Concept 5 — Circuit breaker

Retries handle blips; they make a *hard outage worse* — every client waits out timeouts and retries against a service that needs room to recover (a "retry storm"). The breaker fails fast instead:

- **Closed** — calls pass; count consecutive failures.
- **Open** — after `maxFailures`, calls fail instantly with `ErrOpen` (no I/O) for a `cooldown`.
- **Half-open** — after the cooldown, allow one trial: success → Closed, failure → Open.

```go
func (b *Breaker) Call(fn func() error) error {
    if !b.allow() { return ErrOpen } // open + within cooldown → fail fast
    err := fn()
    b.record(err)                    // success resets; failure may open
    return err
}
```

The shipper wraps each attempt in the breaker but **counts only transient failures** — a 4xx is *our* bad request, not the dependency being down, so it must not trip the circuit. When the breaker opens, the retry loop sees `ErrOpen` and stops immediately (no backoff storm). The breaker takes an injected clock (`WithClock`) so its cooldown transitions test deterministically — no `time.Sleep` in tests.

### Common mistake

No breaker at all — under an outage, retries + backoff across every client become a self-inflicted DDoS that keeps the dependency from recovering. Secondary: tripping the breaker on 4xx (client errors aren't the dependency's fault).

---

## Exercise: warm-up — `retry`

Implement `Do(ctx, attempts, base, fn)` in `exercises/warmup/retry/retry.go` — the exponential-backoff-with-jitter loop that honors `ctx.Done()` and stops early on `retry.Permanent`.

**Time:** 10-15 minutes.

## Exercise: main — `breaker` + `shipper`

1. **`internal/breaker`** — the closed/open/half-open state machine (`Call`/`allow`/`record`); clock injected via `WithClock`.
2. **`internal/shipper`** — `Ship` wiring `retry.Do` + `breaker.Call` + `doOnce`, with selective retry (breaker counts only transient failures; 4xx and `ErrOpen` stop the loop).
3. **`cmd/logstats-client`** (provided) — study how `run` wires stdin → shipper → summary.

**Time:** 45-60 minutes.

---

## Daily habits

```bash
gofmt -w ./...
go vet ./...
go test ./...
make test-race    # daily habit since L18 — the breaker is concurrent shared state
```

## How to run

```bash
# Skeleton (passes vacuously until you implement)
cd lessons/24-http-client/exercises
go test ./...

# Reference solution
go test -v ./lessons/24-http-client/solutions/...
go test -race ./lessons/24-http-client/solutions/...
make test-race

# End-to-end: real server + real client
go build -o /tmp/logstats-server ./lessons/24-http-client/solutions/cmd/logstats-server
go build -o /tmp/logstats-client ./lessons/24-http-client/solutions/cmd/logstats-client
/tmp/logstats-server 127.0.0.1:8282 &
printf '2026-01-02T15:04:05 INFO ok\n2026-01-02T15:04:06 ERROR boom\nbad\n' \
  | /tmp/logstats-client -addr=http://127.0.0.1:8282/ingest
# shipped: accepted=3 parsed=2 failed=1
curl -s 127.0.0.1:8282/stats; echo
# {"counts":{"ERROR":1,"INFO":1},"total":2}
kill %1; rm -f /tmp/logstats-server /tmp/logstats-client
```

## Going further

### Read

- **Go blog — "context" package** + `net/http` client docs: <https://pkg.go.dev/net/http#Client>, <https://pkg.go.dev/context>.
- **AWS Builders' Library — "Timeouts, retries, and backoff with jitter"** (Marc Brooker): the canonical writeup on full vs decorrelated jitter and why it matters.
- **Martin Fowler — "CircuitBreaker"**: <https://martinfowler.com/bliki/CircuitBreaker.html> — the pattern's classic description.
- **`sony/gobreaker`** source — a production circuit breaker; compare its state machine to ours.

### Try

- **Honor `Retry-After` on 429** — parse the header and plumb a delay hint through `retry.Do` (e.g. an error that exposes a `RetryAfter() time.Duration`, used as `max(backoff, retryAfter)`).
- **Max-elapsed-time budget** — add a total deadline to the shipper (`context.WithTimeout`) so retries can't exceed, say, 30s no matter the attempt count.
- **Per-host breakers** — a `map[string]*breaker.Breaker` so one down dependency doesn't open the circuit for healthy ones.
- **Decorrelated jitter** — implement the AWS "decorrelated jitter" variant and compare retry spread against full jitter under simulated load.
- **Streaming/batched client** — instead of one batch, stream stdin in fixed-size batches with bounded in-flight requests (a worker pool from L20).

---

> Phase 4 continues. Next: Lesson 25 — **gRPC**. The `logstats` service gains a typed, streaming RPC surface (client-streaming `Ingest`, unary `GetStats`, interceptors) — and the course takes on its first third-party dependency, `google.golang.org/grpc`, introduced deliberately.
