# Lesson 29: Distributed Patterns & Course Capstone

## What you'll learn

By the end of this lesson you can:

- Explain the **two honest delivery guarantees** (at-most-once, at-least-once) and why **exactly-once delivery is a myth**.
- Use **idempotency keys + server-side dedup** to make duplicate deliveries harmless — *effectively-once* processing.
- Implement the **outbox pattern**: record intent before sending, retry until acked, prune on ack.
- Compose **at-least-once retry (same key) + dedup** into a pipeline that doesn't lose *or* double-count.
- Reason about the whole course: the 29-lesson arc, the running example's evolution, and where to go next.

This is the **final lesson** — Phase 4's seventh and the course's capstone. It ties Phases 1–4 together and is self-contained: no external broker, no new third-party dependency (the producer is `net/http`, the dedup is a `map` + a ring).

## What's different — the finale

Every prior Phase 4 lesson made the *single* `logstatsd` process better: it speaks HTTP (L23) and gRPC (L25), is configured (L26), ships in a container (L27), and is observable (L28). L29 adds the thing every real ingestion system has and a single process hides: **a network between the producer and the aggregator.**

We introduce a **`forwarder`** — the producer half — that ships log-line batches to the aggregator's `/ingest` endpoint over HTTP, with retries. And we make the carried aggregator **idempotent**, so a retried batch is counted once, not twice. The result is a two-process pipeline that stays correct when the network drops an ack.

This is a small, readable version of patterns you'll meet everywhere: Kafka's idempotent producer, Stripe's `Idempotency-Key` header, the transactional outbox. The ideas are the point; the ~100 lines are deliberately legible.

## The distributed pipeline

```
   producer                         network                       aggregator
┌───────────────┐                  (unreliable)            ┌────────────────────┐
│  forwarder    │   POST /ingest                           │   logstatsd        │
│               │   Idempotency-Key: fwd1-1   ──────────▶  │   ┌─────────────┐  │
│  outbox:      │                                          │   │ dedup.Seen? │  │
│   [batch k1]  │   ◀── 200 ack ──────────(may be lost)──  │   └──────┬──────┘  │
│               │                                          │      no  │  yes    │
│  Send()       │                                          │   merge  │  skip   │
│   → append    │   ── retry SAME key on lost ack ──────▶  │   record │  200    │
│   → flush     │                                          │   counts │  dup    │
└───────────────┘                                          └────────────────────┘
```

- The **forwarder** records each batch in an in-memory **outbox** with a stable key (`<id>-<seq>`), then `flush`es — delivering in order, pruning each entry only after it's acked.
- `deliver` POSTs with **at-least-once retry**, reusing the batch's key as the `Idempotency-Key` on every attempt. 2xx → done; 5xx/network → retry with ctx-aware backoff; 4xx → permanent.
- The **aggregator** checks the key against a bounded `dedup.Store` *before* merging. First sighting → merge + record. Repeat → acknowledge (200, `{"duplicate":true}`) without re-counting.

## The package layout

```
lessons/29-capstone-final/{exercises,solutions}/
├── warmup/dedup/           ← you implement: Seen (bounded, concurrency-safe seen-key set)
├── warmup/{config,buildinfo,tracemw}/   carried (L26/L27/L28)
├── proto/ + logstatspb/    carried (gRPC, L25)
├── internal/
│   ├── logparse, logstats  carried (the accumulator)
│   ├── otelx               carried (observability, L28)
│   ├── httpsrv/            idempotent /ingest (Idempotency-Key check, provided)
│   ├── grpcsrv             carried (gRPC face)
│   └── forwarder/          ← you implement: deliver (at-least-once retry + key reuse)
└── cmd/
    ├── logstatsd           the aggregator (HTTP + gRPC), now wires dedup.New(100000)
    └── logstats-forward/   the producer binary (provided) — stdin → /ingest
```

You implement **two** things: the `dedup.Seen` warm-up and the `forwarder.deliver` retry loop. Everything else (the outbox, the idempotent server wiring, the runnable binaries, the carried daemon) is provided to study.

---

## Concept 1 — At-least-once & why exactly-once is a myth

The forwarder sends a batch; the ack is lost in transit. It cannot tell "the server never got it" apart from "the server got it, the ack dropped." That single ambiguity is irreducible — no protocol removes it (the two-generals problem).

So there are exactly **two** honest guarantees:

| Guarantee | Rule | Failure mode |
|---|---|---|
| At-most-once | send, never retry | **loss** |
| At-least-once | retry until acked | **duplicates** |

"Exactly-once *delivery*" is marketing. Systems that advertise it actually do **at-least-once delivery + idempotent processing** = *effectively-once*: the duplicate still arrives, the receiver just makes it a no-op.

### Common mistake

Assuming TCP or your RPC framework gives you exactly-once. A connection reset after the server commits but before the client reads the ack is indistinguishable from a real failure, at every layer. Design for the duplicate.

---

## Concept 2 — Idempotency keys + server-side dedup

```go
// Producer: one stable key per batch, reused on every retry.
req.Header.Set("Idempotency-Key", b.key)   // b.key == "<id>-<seq>"

// Server: check before doing the work.
if key := r.Header.Get("Idempotency-Key"); key != "" && dd.Seen(key) {
    writeJSON(w, http.StatusOK, ingestResponse{Duplicate: true})
    return
}
```

The key identifies the **work**, not the attempt. The server keeps a set of applied keys; a second arrival with the same key is acked but not re-counted. The check is the **gate** — it runs before the merge.

The dedup state is **bounded** — a `map` for O(1) lookup plus a FIFO `order` ring that evicts the oldest key past capacity:

```go
func (s *Store) Seen(key string) bool {
    s.mu.Lock()
    defer s.mu.Unlock()
    if _, ok := s.seen[key]; ok {
        return true
    }
    s.seen[key] = struct{}{}
    s.order = append(s.order, key)
    if len(s.order) > s.cap {
        delete(s.seen, s.order[0])
        s.order = s.order[1:]
    }
    return false
}
```

### Common mistake

A **fresh key per retry** — dedup never fires, and you're back to duplicates. And an **unbounded** dedup map — a memory leak that OOMs the aggregator. Stable key; bounded store. (Production bounds by *time*, not count: a TTL window longer than your max retry horizon, so a slow legitimate retry still dedups.)

---

## Concept 3 — The outbox pattern

```go
func (f *Forwarder) Send(ctx context.Context, lines []string) error {
    f.seq++
    f.outbox = append(f.outbox, batch{key: f.key(), lines: lines})  // record intent
    return f.flush(ctx)                                             // then send
}

func (f *Forwarder) flush(ctx context.Context) error {
    for len(f.outbox) > 0 {
        if err := f.deliver(ctx, f.outbox[0]); err != nil {
            return err              // keep this entry + the rest
        }
        f.outbox = f.outbox[1:]     // prune ONLY on ack
    }
    return nil
}
```

Record intent **before** sending, so the work outlives any single in-flight attempt. Deliver in order; remove an entry **only after it's acked**. The key is assigned at record time, so a later retry of that entry reuses it (Concept 2). In production the outbox is a DB table written in the *same transaction* as the business change; here it's a slice to keep the idea visible.

### Common mistake

**Send-then-record** loses the work on a crash between send and success. **Never pruning** grows the outbox without bound. Record before send; prune on ack.

---

## Concept 4 — Retries + dedup = effectively-once

```
deliver attempt 1 ─POST key=fwd1-1─▶ server: Seen? no  → MERGE + record
                   ◀─ ack LOST ───── (connection dies)
deliver attempt 2 ─POST key=fwd1-1─▶ server: Seen? YES → skip, 200 duplicate
                   ◀─ 200 ────────── outbox.prune
result: counted ONCE despite two deliveries
```

At-least-once retry (same key) + server dedup = **effectively-once**. The retry guarantees the work eventually lands; the key guarantees it lands once. The `deliver` retry loop you implement classifies failures:

```go
switch {
case resp.StatusCode >= 200 && resp.StatusCode < 300:
    return nil                                   // acked
case resp.StatusCode >= 500:
    lastErr = fmt.Errorf("server %d", resp.StatusCode)
    continue                                     // transient → retry (same key)
default:
    return fmt.Errorf("permanent %d", resp.StatusCode)  // 4xx → give up
}
```

The capstone test `TestEffectivelyOnce` proves the property: the aggregator processes a batch, then "loses" the ack (returns 503 the first time it sees a key, replaying the real response on retry). The forwarder retries the same key; the store total is **2**, not 4. Swap to a fresh key per retry and the same test counts **4** — the key reuse is the whole trick.

### Common mistake

Retrying with a **new key** (double-counts), retrying a **4xx** (poison-pill loop — a bad request fails forever), or giving up on a **5xx/network** error (transient — should retry with backoff).

---

## Concept 5 — The course, end to end

| Phase | Lessons | What you can do |
|---|---|---|
| 1 — Foundations | 01–08 | `go run hello` → types, functions, structs, slices/maps, errors, a CLI |
| 2 — Idiomatic Go | 09–14 | pointers, interfaces, generics, errors-as-values, testing, tooling |
| 3 — Concurrency | 15–22 | goroutines, channels, `sync`, `context`, pipelines, profiling/fuzzing |
| 4 — Production | 23–29 | HTTP, gRPC, config, containers, observability, distributed |

The running example grew with you: an **expense tracker** CLI (P1–2) → a **log aggregator** when concurrency arrived (P3) → the **`logstats` service** with HTTP + gRPC faces (P4) → a **distributed pipeline** (L29). Same domain, escalating ambition — every new concept landed on code you already understood.

And it stayed **stdlib-first**: HTTP server, JSON, flags, testing, profiling, `log/slog`, `context` — all standard library. Exactly **two** third-party modules were added, each where the lesson was *about* it: gRPC/protobuf (L25) and OpenTelemetry (L28). Restraint is the lesson: the stdlib carries you far, and every dependency is a boundary you choose deliberately.

---

## Exercise: warm-up — `dedup`

Implement `Seen(key string) bool` in `exercises/warmup/dedup/dedup.go` — first sighting records the key (evicting the oldest past capacity) and returns `false`; a repeat returns `true`. Must be concurrency-safe (a mutex guards the map + ring). The struct and `New(capacity int)` are provided.

**Time:** 10–15 minutes.

## Exercise: main — `forwarder.deliver`

Implement `deliver(ctx, b)` in `exercises/internal/forwarder/forwarder.go` — POST the batch with at-least-once retry, reusing `b.key` as the `Idempotency-Key` on **every** attempt. 2xx → `nil`; 5xx/network → retry with ctx-aware backoff (`time.NewTimer(baseBackoff << (attempt-1))`, honoring `ctx.Done()`); 4xx → permanent error. The outbox (`Send`/`flush`), the `Forwarder` struct, and the idempotent server are provided — study how they compose, then write the retry loop.

**Time:** 30–45 minutes.

---

## Daily habits

```bash
gofmt -w ./...
go vet ./...
go test ./...
make test-race
```

## How to run

```bash
# Skeleton (passes vacuously until you implement Seen + deliver)
cd lessons/29-capstone-final/exercises
go test ./...

# Reference solution: dedup + idempotent ingest + forwarder + the capstone test
go test -v ./lessons/29-capstone-final/solutions/...
go test -race ./lessons/29-capstone-final/solutions/...
make test-race
```

### The two-process capstone demo

```bash
go build -o /tmp/logstatsd ./lessons/29-capstone-final/solutions/cmd/logstatsd
go build -o /tmp/fwd       ./lessons/29-capstone-final/solutions/cmd/logstats-forward

/tmp/logstatsd -http-addr=127.0.0.1:8080 -grpc-addr=127.0.0.1:9090 >/dev/null 2>&1 &
sleep 0.6

# ship two lines through the forwarder to the aggregator's /ingest
printf '2026-01-02T15:04:05 INFO a\n2026-01-02T15:04:06 WARN b\n' \
  | /tmp/fwd -addr=http://127.0.0.1:8080/ingest -id=fwd1

curl -s 127.0.0.1:8080/stats; echo   # {"counts":{"INFO":1,"WARN":1},"total":2}

kill -TERM %1
rm -f /tmp/logstatsd /tmp/fwd
```

Two processes, a network between them, retries that don't double-count.

---

## You've finished the course

This is the end of the 29-lesson path. You started at `go run hello`; you've finished at an observable, idempotent, distributed daemon you could explain line by line. Along the way you wrote a CLI, learned Go's type system and interfaces, mastered goroutines/channels/`context`, profiled and fuzzed real code, and built a service that speaks two protocols, is configured and containerized, emits telemetry, and stays correct across a network.

### What we deliberately left out — and where to go next

The course stopped at the edge of *external infrastructure*, on purpose: those topics need a database, a broker, or a cluster, and each deserves its own course. With the foundation you now have, you can take them on:

- **Databases** — `database/sql`, migrations, and the outbox as a real table written transactionally with the business change.
- **Message brokers** — replace the in-memory outbox with NATS or Kafka; revisit "effectively-once" with a real, replayable log.
- **Orchestration** — deploy the distroless image (L27) to Kubernetes; wire liveness/readiness (L28), autoscaling, rolling updates.
- **Backpressure & rate limiting** — what happens when the producer outruns the aggregator; bounded queues, load shedding.
- **Persistent, sharded dedup** — TTL windows, sharding the dedup store across instances, surviving restarts.

You can read the docs, reason about concurrency and failure, prove a property with a test, and ship a static binary. That's the whole goal. Go build something.

## Going further (in this lesson)

- **Persist the outbox** — back it with a file or SQLite and reload on startup, so a crash mid-send doesn't lose unacked batches.
- **A real broker** — swap `deliver`'s HTTP POST for a NATS/Kafka publish; the outbox + key reuse stay identical.
- **Dedup TTL & sharding** — replace the FIFO ring with a time-windowed store; shard by key prefix across aggregator instances.
- **Backpressure** — bound the outbox; have `Send` block or shed when it's full, and surface that to the producer.
- **Exactly-once caveats** — read why "exactly-once" still requires idempotency even with transactional brokers; the duplicate never truly disappears, you just absorb it.

---

> The end. From `go run hello` to a distributed, idempotent service — stdlib-first, two deliberate dependencies, one running example carried the whole way. Thanks for building it.
