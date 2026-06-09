<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">29</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 4 — Production &amp; Distributed</div>
<h1>Distributed Patterns &amp; Course Capstone</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Make ingestion correct across an unreliable network. Learn why exactly-once delivery is a myth, how idempotency keys + server-side dedup turn at-least-once retries into <em>effectively-once</em> processing, and the outbox pattern that survives a crash mid-send — all self-contained, no external broker. Then step back and look at the whole course: from <code>go run hello</code> to a distributed, idempotent service.</p>
</div>
</div>
</div>

---

## What we'll cover

- **At-least-once & the exactly-once myth** — the lost-ack problem; the two honest delivery guarantees.
- **Idempotency keys + dedup** — a stable key per unit of work; the server skips repeats; bounded memory.
- **The outbox pattern** — record intent before sending; retry until acked; prune on ack.
- **Retries + dedup = effectively-once** — the whole pipeline, and the counter-example that double-counts.
- **Course wrap-up** — Phases 1–4 retrospective, the running example's arc, and where to go next.

---

## The story so far — the finale

`logstatsd` is observable, configured, containerized. But it's still **one process**. Real ingestion has a producer somewhere else — an agent on another host tailing logs and shipping lines over a network that drops packets, times out, and duplicates.

The moment a network sits between producer and aggregator, a hard question appears: the forwarder sends a batch, then the connection dies *before the ack comes back*. Did the server process it or not? The forwarder can't tell. If it gives up, it might lose data. If it retries, it might double-count.

**L29 closes that gap.** We add a `forwarder` (the producer) that ships batches to the aggregator's `/ingest`, and we make the pipeline correct under retries — without an external broker, without "exactly-once" magic, using two small ideas: a **stable idempotency key** per batch and **server-side dedup**. This is the same shape as Kafka's idempotent producer or Stripe's `Idempotency-Key` header, in ~100 lines you can read.

---

## Concept 1: At-least-once & why exactly-once is a myth

### Motivation

The forwarder sends, the ack is lost in transit. It cannot distinguish "server never got it" from "server got it, ack dropped." Every retry decision you make is downstream of that one ambiguity — and no protocol removes it.

---

### The basics

There are exactly two honest delivery guarantees over an unreliable network:

| Guarantee | Rule | Failure mode |
|---|---|---|
| **At-most-once** | send, never retry | **loss** (lost ack → data gone) |
| **At-least-once** | retry until acked | **duplicates** (lost ack → re-sent) |

"**Exactly-once *delivery***" is marketing. You cannot, in the general case, deliver a message across a network exactly once — the two-generals problem says so. What systems that claim "exactly-once" actually do is **at-least-once delivery + idempotent processing** = *effectively-once*. The duplicate still arrives; the receiver just makes the second one a no-op.

So the design choice is real and narrow: pick at-least-once (retry), then make duplicates harmless. That's the rest of this lesson.

---

### Common mistake

Assuming TCP or your RPC framework gives you exactly-once. It doesn't — a connection reset after the server commits but before the client reads the ack is indistinguishable from a failure, at every layer. Design for duplicates; don't pretend they can't happen.

---

### Recap

- Two honest guarantees: **at-most-once** (lossy) and **at-least-once** (dup-prone).
- "Exactly-once delivery" is a myth; "effectively-once" = at-least-once + idempotency.
- The lost ack is irreducible — design for the duplicate.

---

## Concept 2: Idempotency keys + server-side dedup

### Motivation

If retries cause duplicates, the receiver needs a way to recognize "I've already processed this exact unit of work." A **stable key** per unit — the same on every retry — lets it.

---

### The basics

```go
// Producer: one stable key per batch, REUSED across retries.
batch{key: fmt.Sprintf("%s-%d", forwarderID, seq), lines: lines}
req.Header.Set("Idempotency-Key", b.key)   // same key, every attempt

// Server: record seen keys; skip repeats (before doing the work).
if key := r.Header.Get("Idempotency-Key"); key != "" && dd.Seen(key) {
    writeJSON(w, 200, ingestResponse{Duplicate: true})
    return
}
```

The key identifies the *work*, not the *attempt*. The server keeps a set of keys it has already applied; a second arrival with the same key is acknowledged (200) but **not re-counted**. The check happens *before* the merge, so the dedup is the gate, not an afterthought.

---

### A worked example

Dedup state can't grow forever. Our `dedup.Store` is **bounded** — a `map` for O(1) lookup plus a FIFO `order` ring; past capacity it evicts the oldest key.

```go
func (s *Store) Seen(key string) bool {
    s.mu.Lock()
    defer s.mu.Unlock()
    if _, ok := s.seen[key]; ok {
        return true               // already processed
    }
    s.seen[key] = struct{}{}
    s.order = append(s.order, key)
    if len(s.order) > s.cap {     // evict oldest
        delete(s.seen, s.order[0])
        s.order = s.order[1:]
    }
    return false                  // first sighting
}
```

In production you'd bound by **time** instead — a TTL window comfortably longer than your max retry horizon — so a legitimate slow retry still dedups.

---

### Common mistake

A **fresh key per retry** — then every retry looks like new work and dedup never fires (you've rebuilt at-least-once-with-duplicates). And **unbounded dedup state** — a `map` that never evicts is a memory leak that OOMs the aggregator. Stable key; bounded store.

---

### Recap

- The key identifies the **work**, not the attempt — reused across retries.
- Check **before** applying; a repeat is acked but not re-counted.
- Bound the dedup state (size here; TTL in production).

---

## Concept 3: The outbox pattern

### Motivation

The forwarder is told to send lines. If it POSTs immediately and the process crashes mid-send, those lines are gone — they only ever lived in that one in-flight request. The fix is to **write down the intent first**.

---

### The basics

```go
func (f *Forwarder) Send(ctx context.Context, lines []string) error {
    f.seq++
    f.outbox = append(f.outbox, batch{key: f.key(), lines: lines})  // record intent
    return f.flush(ctx)                                             // then try to send
}
```

You append to the outbox **before** sending, so the intent outlives any single attempt. The key is assigned at record time, so every later retry of that entry reuses it (Concept 2).

---

### A worked example

```go
func (f *Forwarder) flush(ctx context.Context) error {
    for len(f.outbox) > 0 {
        if err := f.deliver(ctx, f.outbox[0]); err != nil {
            return err              // keep this entry + the rest; retry next flush
        }
        f.outbox = f.outbox[1:]     // prune ONLY on ack
    }
    return nil
}
```

`flush` delivers in order and removes an entry **only after it's acked**. A failed delivery leaves the entry in place for a later `flush` to retry. In production the outbox is a database table written *in the same transaction* as the business change — here it's an in-memory slice to keep the idea visible.

---

### Common mistake

**Send-then-record** (or never recording): a crash between sending and noting success loses the work, or double-sends with no stable key. And **never pruning** the outbox — it grows without bound. Record before send; prune on ack.

---

### Recap

- Record intent **before** sending — survives a crash mid-flight.
- Deliver in order; prune an entry **only on ack**.
- The key is assigned at record time, so retries reuse it.

---

## Concept 4: Retries + dedup together = effectively-once

### Motivation

Neither half is enough alone. Retries without dedup double-count; dedup without retries still loses data on the first failure. Composed, they give the property everyone actually wants.

---

### The basics

```
forwarder.Send(lines)
   └─ outbox.append(batch{key:"fwd1-1", lines})
   └─ deliver attempt 1 ─POST key=fwd1-1─▶ server: Seen? no → MERGE, record key
                          ◀─ ack LOST ──── (connection dies)
   └─ deliver attempt 2 ─POST key=fwd1-1─▶ server: Seen? YES → skip, 200 duplicate
                          ◀─ 200 ───────── outbox.prune
result: lines counted ONCE, despite two deliveries
```

**At-least-once retry (same key) + server dedup = effectively-once.** The retry guarantees the work eventually lands; the key guarantees it lands *once*. That is the entire pipeline.

---

### A worked example

The capstone test (`TestEffectivelyOnce`) proves it. The aggregator processes a batch, then "loses" the ack — it returns 503 the first time it sees a key (after the real handler has already merged + recorded it), replaying the real response on the retry. The forwarder retries the **same** key; the server dedups; the store's total is **2**, not 4:

```go
if _, total := store.Snapshot(); total != 2 {
    t.Errorf("total = %d, want 2 (effectively-once despite redelivery)", total)
}
```

The counter-example is one line different: give each retry a **new** key and the same test counts **4** — every redelivery is fresh work. The key reuse is the whole trick.

---

### Common mistake

Retrying with a **new key** (double-counts), or classifying failures wrong — retrying a **4xx** (a bad request will fail forever; that's a poison-pill retry loop) or giving up on a **5xx/network** error (transient; should retry with backoff). Transient → retry, same key; permanent → stop.

---

### Recap

- At-least-once retry (same key) + dedup = **effectively-once**.
- Classify failures: 5xx/network → retry; 4xx → permanent.
- A fresh key per retry breaks the property — proven by the counter-example.

---

## Concept 5: The course, end to end

### Four phases, one through-line

| Phase | Lessons | What you can do |
|---|---|---|
| **1 — Foundations** | 01–08 | `go run hello` → types, functions, structs, slices/maps, errors, a CLI |
| **2 — Idiomatic Go** | 09–14 | pointers, interfaces, generics, errors-as-values, testing, tooling |
| **3 — Concurrency** | 15–22 | goroutines, channels, `sync`, `context`, pipelines, profiling/fuzzing |
| **4 — Production** | 23–29 | HTTP, gRPC, config, containers, observability, **distributed** |

The running example grew with you: an **expense tracker** CLI (P1–2) → a **log aggregator** (P3, when concurrency arrived) → the **`logstats` service** with HTTP + gRPC faces (P4) → a **distributed pipeline** (L29). Same domain, escalating ambition — so every new concept landed on code you already understood.

---

### Two dependencies, on purpose

The whole course used the **standard library** — HTTP server, JSON, flags, testing, profiling, `log/slog`, `context` — and added exactly **two** third-party modules, each where the lesson was *about* it: **gRPC/protobuf** (L25) and **OpenTelemetry** (L28). That restraint is the lesson: Go's stdlib carries you remarkably far, and every dependency is a boundary you choose deliberately, not a reflex.

---

### Where to go next

We deliberately left out the things that need real infrastructure:

- **Databases** — `database/sql`, migrations, the outbox as a real table written transactionally.
- **Message brokers** — swap the in-memory outbox for NATS/Kafka; revisit "effectively-once" with a real log.
- **Orchestration** — deploy the distroless image to Kubernetes; wire liveness/readiness, HPA, rolling updates.
- **Backpressure & rate limiting** — what happens when the producer outruns the aggregator.
- **Persistent dedup** — TTL windows, sharding the dedup store, surviving restarts.

You now have the foundation to learn all of them: you can read the docs, reason about concurrency and failure, write tests that prove a property, and ship a static binary. That's the goal.

---

## Practice

### Warm-up — `dedup`

In `exercises/warmup/dedup/`, implement `Seen(key string) bool` — first sighting records the key (evicting the oldest past capacity) and returns `false`; a repeat returns `true`. Concurrency-safe.

```bash
cd lessons/29-capstone-final/exercises
go test ./warmup/dedup/...
```

### Main — `forwarder.deliver`

In `internal/forwarder/`, implement `deliver(ctx, b)` — POST the batch with **at-least-once retry**, reusing `b.key` as the `Idempotency-Key` on **every** attempt; 2xx → done, 5xx/network → retry with ctx-aware backoff, 4xx → permanent failure. The outbox, `Send`, `flush`, and the idempotent server are provided — study how they compose.

```bash
cd lessons/29-capstone-final/exercises
go test ./...
make test-race
```

Note:
The two-process demo is the payoff slide — run it live if you can. Start the daemon, pipe two lines through the forwarder, show `total:2`. Then (optional) flip the forwarder to a fresh key per retry and show the same test counting 4 — the key reuse is the whole trick, and seeing the counter-example fail makes it stick.

---

## The two-process capstone

```bash
# aggregator
go build -o /tmp/logstatsd ./lessons/29-capstone-final/solutions/cmd/logstatsd
/tmp/logstatsd -http-addr=127.0.0.1:8080 -grpc-addr=127.0.0.1:9090 &

# forwarder ships stdin lines to the aggregator's /ingest
go build -o /tmp/fwd ./lessons/29-capstone-final/solutions/cmd/logstats-forward
printf '2026-01-02T15:04:05 INFO a\n2026-01-02T15:04:06 WARN b\n' \
  | /tmp/fwd -addr=http://127.0.0.1:8080/ingest -id=fwd1

curl -s 127.0.0.1:8080/stats; echo   # {"counts":{"INFO":1,"WARN":1},"total":2}
kill -TERM %1
```

Two processes, a network between them, retries that don't double-count. That's the finale.

---

## Closing thought

Distributed correctness isn't built from exotic primitives — it's built from accepting one uncomfortable truth (the network will duplicate your messages) and designing so the duplicate doesn't matter. A stable key, a set of seen keys, an outbox you prune on ack. The same pattern scales from these 100 lines to Kafka's idempotent producer; the principle is identical.

And that mirrors the whole course. Every phase took something that looked hard — concurrency, RPC, observability, distribution — and reduced it to a few honest ideas applied to code you already understood. The expense tracker became a distributed service one comprehensible step at a time. That's how real systems get built: not in a leap, but as a sequence of changes each small enough to reason about.

You started at `go run hello`. You've finished at an observable, idempotent, distributed daemon you could explain line by line. **That's the course.**

---

## What we learned

- **Exactly-once is a myth**: the honest guarantees are at-most-once (lossy) and at-least-once (dup-prone).
- **Idempotency keys + dedup**: a stable key per unit of work; the server skips repeats; **bound** the state.
- **Outbox**: record intent before sending, deliver in order, prune **only on ack** — survives a crash mid-send.
- **Effectively-once**: at-least-once retry (same key) + dedup; a fresh key per retry double-counts.
- **The arc**: stdlib-first, two deliberate dependencies, `go run hello` → a distributed service.

---

## Up next

There is no Lesson 30 — **you've finished the course.** From here the path leads *outward*: pick a piece of real infrastructure (a database, a broker, Kubernetes) and apply what you have. You can read the docs, reason about failure, prove a property with a test, and ship a binary. Go build something.
