# Plan FF — Lesson 29 (Distributed patterns & course capstone) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax.
>
> **Commit policy:** NO `Co-Authored-By` trailer or "Generated with" line in commit messages or PR bodies. Subject + body only.

**Goal:** Author lesson 29 — the **course finale**. Make ingestion correct across two `logstatsd` instances: a forwarder ships log batches to an aggregator; idempotency keys + server-side dedup mean retried/duplicated batches count **once**. Self-contained — no external broker. Tie Phases 1-4 together.

**Architecture:** Same per-lesson pattern, six tasks. Carries the entire L28 daemon verbatim. New: `warmup/dedup` (bounded, concurrency-safe seen-key set — the warmup exercise), idempotent `/ingest` wiring in `httpsrv` (provided), `internal/forwarder` (outbox + at-least-once retry + stable idempotency keys — the main exercise), and `cmd/logstats-forward` (provided). Five-concept slide deck (incl. course wrap-up).

**Tech Stack:** Go 1.23 stdlib (`net/http`, `context`, `encoding/json`, `sync`, `time`, `fmt`) + carried grpc/protobuf/otel. **No new deps.**

---

## Scope

After Plan FF (and its merge): lesson 29 complete; the course's **29 lessons are done**; `make test` + `make test-race` green; the forwarder→aggregator pipeline counts duplicates once; `29-capstone-final` in the index; go directive stays `go 1.23`.

### Design decisions (2 user-approved + plan-recommended)

**User-approved via brainstorming:**

1. **`internal/forwarder` + `cmd/logstats-forward`** — core logic in the package (testable in-process), plus a runnable binary shipping stdin lines to an aggregator `/ingest`. A real two-process demo.
2. **Five slide concepts**, the fifth being a **course wrap-up** (Phases 1-4 retrospective).

**Plan-recommended:**

3. **Idempotency via an `Idempotency-Key` HTTP header + `dedup.Store`.** The handler skips the merge (returns `{duplicate:true}`) when the key was seen; no key → processed every time (back-compat).
4. **Stable deterministic keys `<forwarderID>-<seq>`** (no `rand`/time) — reused across retries so a redelivery dedups.
5. **`warmup/dedup` + `internal/forwarder` are the exercises**; the idempotent-ingest wiring (httpsrv/cmd) and `cmd/logstats-forward` are provided study code.
6. **Carry the whole L28 daemon** (keeps grpc/otel deps live; the directive stays `go 1.23`).

### Verified facts (prototyped before writing this plan)

- `dedup.Store` (map + FIFO `order` ring, mutex): `Seen` returns false-then-true; bounded eviction works; concurrent `Seen` race-clean. (A subtle TEST bug — probing an evicted key re-adds it and evicts the next — was found and the tests written to avoid it.)
- Idempotent `/ingest`: same key twice → counted once (`duplicate:true` on the 2nd); different keys → twice; no key → each time.
- `forwarder` (outbox + at-least-once retry, key reused across attempts): retries transient (503×2 then 200) delivering once.
- **Capstone end-to-end:** an aggregator that PROCESSES then "loses the ack" (wrapper returns 503 the first time it sees a key, replays the real response on the retry) → the forwarder retries the same key → the aggregator dedups → the lines are merged **exactly once**. Verified race-clean.
- The forwarder POSTs to the full endpoint URL (caller appends `/ingest`, like L24's shipper).

---

## File structure

```
lessons/29-capstone-final/{exercises,solutions}/
├── warmup/dedup/{dedup.go, dedup_test.go}        (Task 2 — Seen SKELETON)
├── warmup/{config,buildinfo,tracemw}/            (Task 1 — carried)
├── proto/ + logstatspb/                          (Task 1 — carried, regenerated)
├── internal/
│   ├── logparse, logstats, otelx, grpcsrv        (Task 1 — carried)
│   ├── httpsrv/                                   (Task 3 — idempotent /ingest, provided)
│   └── forwarder/{forwarder.go, forwarder_test.go}   (Task 4 — Send SKELETON; capstone test in solutions)
└── cmd/
    ├── logstatsd/                                 (Task 3 — dedup wired in, provided)
    └── logstats-forward/{main.go, main_test.go}  (Task 4 — provided; test SKELETON in exercises)
```

---

## Task 1: Scaffold + carry forward + regenerate proto (controller-direct)

- [ ] **Step 1:** Scaffold + remove the 8 flat stubs (`make new-lesson NAME=29-capstone-final`; rm flat main/warmup files).

- [ ] **Step 2:** Carry the whole daemon (both trees), rewrite paths + header:
```bash
SRC=lessons/28-observability; DST=lessons/29-capstone-final
for side in exercises solutions; do
  mkdir -p "$DST/$side/internal" "$DST/$side/proto" "$DST/$side/warmup" "$DST/$side/cmd"
  cp -R "$SRC/$side/internal/." "$DST/$side/internal/"
  cp -R "$SRC/$side/warmup/." "$DST/$side/warmup/"
  cp -R "$SRC/$side/cmd/logstatsd" "$DST/$side/cmd/logstatsd"
  cp "$SRC/$side/proto/logstats.proto" "$DST/$side/proto/logstats.proto"
done
grep -rl '28-observability' "$DST" | while read -r f; do sed -i '' 's#lessons/28-observability#lessons/29-capstone-final#g' "$f"; done
grep -rl 'lesson 28' "$DST" | while read -r f; do sed -i '' 's/lesson 28/lesson 29/g' "$f"; done
```

- [ ] **Step 3:** Add the L29 protos to the `Makefile` `proto` target (append two lines), then regenerate + tidy:
```bash
make proto && go mod tidy && grep '^go ' go.mod   # stays go 1.23.0
```

- [ ] **Step 4:** Verify carried baseline (`go build`; `go test`; vet/lint/fmt).

- [ ] **Step 5:** Commit:
```bash
git add lessons/29-capstone-final Makefile go.mod go.sum
git commit -m "chore(lesson-29): scaffold + carry forward the L28 logstatsd daemon"
```

---

## Task 2: Warm-up — `dedup`

**Files (4 total):**

- [ ] **Step 1:** `exercises/warmup/dedup/dedup.go` (SKELETON):

```go
// Package dedup is a bounded, concurrency-safe set of seen idempotency
// keys — the heart of server-side deduplication.
package dedup

import "sync"

// Store remembers up to `cap` recently-seen keys (FIFO eviction).
type Store struct {
	mu    sync.Mutex
	cap   int
	seen  map[string]struct{}
	order []string // insertion order, for bounded eviction
}

func New(capacity int) *Store {
	if capacity < 1 {
		capacity = 1
	}
	return &Store{cap: capacity, seen: make(map[string]struct{}, capacity)}
}

// Seen reports whether key was already recorded. On first sight it
// records the key (evicting the oldest if over capacity) and returns
// false; on a repeat it returns true. Safe for concurrent use.
//
// Hint:
//   lock; if _, ok := s.seen[key]; ok { return true }
//   s.seen[key] = struct{}{}; s.order = append(s.order, key)
//   if len(s.order) > s.cap { evict s.order[0] from both order and seen }
//   return false
func (s *Store) Seen(key string) bool {
	_ = sync.Mutex{}
	panic("TODO: bounded, concurrency-safe seen-set; first sight false (record), repeat true")
}
```

- [ ] **Step 2:** `exercises/warmup/dedup/dedup_test.go` (SKELETON):

```go
package dedup

import "testing"

// TestSeen is a SKELETON. Cover: first sight false then repeat true;
// bounded eviction (oldest evicted past capacity — probe ONE key, since
// Seen mutates state); concurrent Seen under -race.
func TestSeen(t *testing.T) {
	// TODO
	_ = New
}
```

- [ ] **Step 3:** `solutions/warmup/dedup/dedup.go` (full `Seen` — VERIFIED):

```go
// Package dedup is a bounded, concurrency-safe set of seen idempotency
// keys — the heart of server-side deduplication. Reference impl.
package dedup

import "sync"

type Store struct {
	mu    sync.Mutex
	cap   int
	seen  map[string]struct{}
	order []string
}

func New(capacity int) *Store {
	if capacity < 1 {
		capacity = 1
	}
	return &Store{cap: capacity, seen: make(map[string]struct{}, capacity)}
}

func (s *Store) Seen(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.seen[key]; ok {
		return true
	}
	s.seen[key] = struct{}{}
	s.order = append(s.order, key)
	if len(s.order) > s.cap {
		oldest := s.order[0]
		s.order = s.order[1:]
		delete(s.seen, oldest)
	}
	return false
}
```

- [ ] **Step 4:** `solutions/warmup/dedup/dedup_test.go` (VERIFIED — note the careful eviction probing):

```go
package dedup

import (
	"fmt"
	"sync"
	"testing"
)

func TestSeenFirstThenRepeat(t *testing.T) {
	d := New(10)
	if d.Seen("a") {
		t.Error("first sight should be false")
	}
	if !d.Seen("a") {
		t.Error("second sight should be true")
	}
}

func TestBoundedEviction(t *testing.T) {
	d := New(2)
	d.Seen("a")
	d.Seen("b")
	d.Seen("c") // over cap 2 → evicts oldest "a"; holds {b, c}
	// "a" was evicted, so it reads as new. (This probe re-adds "a" and
	// evicts "b" — Seen mutates, so probe only the one key under test.)
	if d.Seen("a") {
		t.Error("a should have been evicted (read as new)")
	}
}

func TestConcurrent(t *testing.T) {
	d := New(1000)
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(n int) { defer wg.Done(); d.Seen(fmt.Sprintf("k%d", n)) }(i)
	}
	wg.Wait()
}
```

- [ ] **Step 5:** Verify + commit:
```bash
gofmt -l lessons/29-capstone-final/
go test ./lessons/29-capstone-final/exercises/warmup/dedup/... 2>&1 | tail -5
go test -v -race ./lessons/29-capstone-final/solutions/warmup/dedup/... 2>&1 | tail -10
make test && go vet ./lessons/29-capstone-final/... && golangci-lint run ./lessons/29-capstone-final/...

git add lessons/29-capstone-final/exercises/warmup/dedup lessons/29-capstone-final/solutions/warmup/dedup
git commit -m "feat(lesson-29): warmup — dedup (bounded, concurrency-safe seen-key set)"
```

Expected: exercises vacuous-pass; solutions 3 tests PASS; -race clean.

---

## Task 3: Idempotent `/ingest` (httpsrv + cmd wiring, provided)

Add the `Idempotency-Key` dedup check to the carried `httpsrv` and wire a `dedup.Store` through the daemon. PROVIDED (study code) in both trees.

- [ ] **Step 1:** Edit `httpsrv/httpsrv.go` (both trees):
  - Add an `ingestResponse.Duplicate bool` field: `Duplicate bool \`json:"duplicate,omitempty"\``.
  - Change `Router(store, logger)` → `Router(store *logstats.Store, dd *dedup.Store, logger *slog.Logger)`; pass `dd` to `ingestHandler`.
  - Change `ingestHandler(store, ingested)` → `ingestHandler(store *logstats.Store, dd *dedup.Store, ingested metric.Int64Counter)`; at the very top of the returned handler:
    ```go
    if key := r.Header.Get("Idempotency-Key"); key != "" && dd.Seen(key) {
        writeJSON(w, http.StatusOK, ingestResponse{Duplicate: true})
        return
    }
    ```
  - Add the import `.../internal/dedup` (per tree). (`dedup` lives at `warmup/dedup` — import `.../warmup/dedup`.)

- [ ] **Step 2:** Update `httpsrv/httpsrv_test.go` (both trees): `testRouter()` now passes a `dedup.New(1000)`. Solutions: ADD idempotency tests (same key once, different keys twice, no key each — see below). Exercises: stays a vacuous skeleton.

Solutions `httpsrv_test.go` additions (idempotency); note the vet-clean `total` helper (check the error):

```go
func TestIdempotentSameKeyCountedOnce(t *testing.T) {
	srv := httptest.NewServer(Router(logstats.NewStore(), dedup.New(1000), discardLogger()))
	defer srv.Close()
	body := `{"lines":["2026-01-02T15:04:05 INFO a","2026-01-02T15:04:06 WARN b"]}`
	r1 := postKey(t, srv.URL, "k1", body)
	r2 := postKey(t, srv.URL, "k1", body)
	if r1.Duplicate || !r2.Duplicate {
		t.Errorf("dup flags: r1=%v r2=%v (want false,true)", r1.Duplicate, r2.Duplicate)
	}
	if got := totalStat(t, srv.URL); got != 2 {
		t.Errorf("total = %d, want 2 (counted once)", got)
	}
}
// + TestDifferentKeysCountedTwice, TestNoKeyCountedEachTime,
// + helpers postKey (sets Idempotency-Key) and totalStat (checks the
//   http.Get error — go vet flags using resp before the err check).
```

- [ ] **Step 3:** Edit `cmd/logstatsd/main.go` (both trees): in `run`, create `dd := dedup.New(100000)` and pass it to `httpsrv.Router(store, dd, logger)`. Import `.../warmup/dedup`. (Provided.)

- [ ] **Step 4:** Verify + commit:
```bash
gofmt -l lessons/29-capstone-final/
go test ./lessons/29-capstone-final/exercises/internal/httpsrv/... 2>&1 | tail -3
go test -v ./lessons/29-capstone-final/solutions/internal/httpsrv/... 2>&1 | tail -20
go test -race ./lessons/29-capstone-final/solutions/cmd/logstatsd/... 2>&1 | tail -3
make test && make test-race && go vet ./lessons/29-capstone-final/... && golangci-lint run ./lessons/29-capstone-final/...

git add lessons/29-capstone-final
git commit -m "feat(lesson-29): idempotent /ingest (Idempotency-Key + dedup) wired through the daemon"
```

Expected: exercises vacuous-pass; solutions httpsrv idempotency tests PASS; carried logstatsd TestServe + TestGRPCSpan still pass; -race clean.

> Note: the exercises `dedup.Seen` is a SKELETON (panics); the exercises `httpsrv` ingest calls it but the exercises `httpsrv` test is vacuous, so nothing invokes the panic. Consistent with the carried convention.

---

## Task 4: `internal/forwarder` + `cmd/logstats-forward`

The distributed half: outbox + at-least-once retry + stable keys (the main exercise), plus a runnable forwarder binary (provided) and the capstone effectively-once test (solutions).

- [ ] **Step 1:** `exercises/internal/forwarder/forwarder.go` (SKELETON — types/`New`/`flush`/`deliver` provided; `Send` panics? No — make `deliver` the core but keep it simple: `Send` is the exercise, `deliver` provided). Actually: provide the struct + `New` + `deliver` (the retry/key mechanics) and make **`Send`** (enqueue + flush) the skeleton — OR make `deliver` the skeleton. Decision: **`deliver` (the at-least-once retry + key reuse) is the exercise**; `New`/`Send`/`flush`/outbox are provided. That puts the meaty distributed logic (retry, key reuse, transient-vs-permanent) in the student's hands.

Provide everything except `deliver`:

```go
// Package forwarder ships log-line batches to an aggregator's /ingest
// with an outbox + at-least-once retry + stable idempotency keys.
package forwarder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ingestRequest struct {
	Lines []string `json:"lines"`
}

// batch is an outbox entry: a unit of work with a stable idempotency key
// reused across retries (so a redelivery dedups, not double-counts).
type batch struct {
	key   string
	lines []string
}

type Forwarder struct {
	url         string
	id          string
	client      *http.Client
	maxAttempts int
	baseBackoff time.Duration

	seq    int
	outbox []batch
}

// New returns a Forwarder posting to url (the full /ingest endpoint),
// tagging keys with id (so multiple forwarders don't collide).
func New(url, id string) *Forwarder {
	return &Forwarder{
		url: url, id: id,
		client:      &http.Client{Timeout: 5 * time.Second},
		maxAttempts: 5,
		baseBackoff: time.Microsecond,
	}
}

// Send enqueues lines into the outbox with a fresh stable key, then
// flushes the outbox (delivering pending batches in order).
func (f *Forwarder) Send(ctx context.Context, lines []string) error {
	f.seq++
	f.outbox = append(f.outbox, batch{key: fmt.Sprintf("%s-%d", f.id, f.seq), lines: lines})
	return f.flush(ctx)
}

// flush delivers outbox batches in order, stopping (and keeping the rest)
// on the first delivery error.
func (f *Forwarder) flush(ctx context.Context) error {
	for len(f.outbox) > 0 {
		if err := f.deliver(ctx, f.outbox[0]); err != nil {
			return err
		}
		f.outbox = f.outbox[1:]
	}
	return nil
}

// deliver POSTs one batch with at-least-once retry, REUSING the batch's
// idempotency key on every attempt (so a redelivery after a lost ack is
// deduped by the server). Retries transient failures (network, 5xx);
// gives up on a permanent (4xx) failure. IMPLEMENT THIS.
//
// Hint:
//   body, _ := json.Marshal(ingestRequest{Lines: b.lines})
//   for attempt := 0; attempt < f.maxAttempts; attempt++ {
//       if attempt > 0 { backoff via time.NewTimer(f.baseBackoff << (attempt-1)) honoring ctx }
//       req, _ := http.NewRequestWithContext(ctx, POST, f.url, bytes.NewReader(body))
//       req.Header.Set("Content-Type","application/json")
//       req.Header.Set("Idempotency-Key", b.key)   // SAME key every attempt
//       resp, err := f.client.Do(req)
//       if err != nil { lastErr = err; continue }   // network → retry
//       drain+close body
//       2xx → return nil; 5xx → retry; 4xx → return permanent error
//   }
//   return lastErr
func (f *Forwarder) deliver(ctx context.Context, b batch) error {
	_ = bytes.NewReader
	_ = io.Discard
	panic("TODO: POST with retry, reusing b.key each attempt; 2xx ok, 5xx/network retry, 4xx give up")
}
```

- [ ] **Step 2:** `exercises/internal/forwarder/forwarder_test.go` (SKELETON, no calls).

- [ ] **Step 3:** `solutions/internal/forwarder/forwarder.go` — same but `deliver` implemented (VERIFIED):

```go
func (f *Forwarder) deliver(ctx context.Context, b batch) error {
	body, _ := json.Marshal(ingestRequest{Lines: b.lines})
	var lastErr error
	for attempt := 0; attempt < f.maxAttempts; attempt++ {
		if attempt > 0 {
			t := time.NewTimer(f.baseBackoff << (attempt - 1))
			select {
			case <-ctx.Done():
				t.Stop()
				return ctx.Err()
			case <-t.C:
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.url, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", b.key) // SAME key every attempt
		resp, err := f.client.Do(req)
		if err != nil {
			lastErr = err
			continue // network → transient
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		switch {
		case resp.StatusCode >= 200 && resp.StatusCode < 300:
			return nil // acked
		case resp.StatusCode >= 500:
			lastErr = fmt.Errorf("forwarder: server %d", resp.StatusCode)
			continue // transient → retry (same key)
		default:
			return fmt.Errorf("forwarder: permanent %d", resp.StatusCode) // 4xx
		}
	}
	return lastErr
}
```

(The rest of `solutions/internal/forwarder/forwarder.go` is identical to the exercises file with `deliver` implemented.)

- [ ] **Step 4:** `solutions/internal/forwarder/forwarder_test.go` (VERIFIED — retry + the capstone effectively-once; uses `sync.Mutex`, no unused imports):

```go
package forwarder

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/29-capstone-final/solutions/internal/httpsrv"
	"github.com/ristkari-dev/go-training/lessons/29-capstone-final/solutions/internal/logstats"
	"github.com/ristkari-dev/go-training/lessons/29-capstone-final/solutions/warmup/dedup"
	// note: a slog discard logger is needed for httpsrv.Router; import log/slog + io
)

func TestRetriesTransient(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	f := New(srv.URL+"/ingest", "fwd1")
	if err := f.Send(context.Background(), []string{"2026-01-02T15:04:05 INFO a"}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

// TestEffectivelyOnce — the capstone. The aggregator PROCESSES a batch
// but its ack is "lost" (the wrapper returns 503 the first time it sees
// a key, replaying the real response on the retry). The forwarder
// retries the SAME key → the aggregator dedups → lines counted ONCE.
func TestEffectivelyOnce(t *testing.T) {
	store := logstats.NewStore()
	real := httpsrv.Router(store, dedup.New(1000), discardLogger())

	var mu sync.Mutex
	failed := map[string]bool{}
	agg := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		rec := httptest.NewRecorder()
		real.ServeHTTP(rec, r) // real handler merges + records the key (first time)
		mu.Lock()
		first := !failed[key]
		failed[key] = true
		mu.Unlock()
		if first {
			w.WriteHeader(http.StatusServiceUnavailable) // "lose the ack"
			return
		}
		for k, vs := range rec.Header() { // replay the real response on retry
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(rec.Code)
		_, _ = w.Write(rec.Body.Bytes())
	}))
	defer agg.Close()

	f := New(agg.URL+"/ingest", "fwd1")
	if err := f.Send(context.Background(), []string{
		"2026-01-02T15:04:05 INFO a", "2026-01-02T15:04:06 WARN b",
	}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if _, total := store.Snapshot(); total != 2 {
		t.Errorf("total = %d, want 2 (effectively-once despite redelivery)", total)
	}
}
```

(Add a `discardLogger()` helper — `slog.New(slog.NewJSONHandler(io.Discard, nil))` — and the `io`/`log/slog` imports.)

- [ ] **Step 5:** `cmd/logstats-forward/main.go` (both trees, PROVIDED): `run(ctx, addr, id, stdin, stdout)` reads stdin lines, `forwarder.New(addr, id).Send(ctx, lines)`, prints a summary. `main` wires `-addr` (full `/ingest` URL) + `-id` flags + `signal.NotifyContext`. `exercises/cmd/logstats-forward/main_test.go` is a SKELETON; `solutions/...main_test.go` is a small in-process test (forward to an httptest aggregator, assert delivered).

- [ ] **Step 6:** Verify + commit:
```bash
gofmt -l lessons/29-capstone-final/
go test ./lessons/29-capstone-final/exercises/... 2>&1 | tail -8
go test -v -race ./lessons/29-capstone-final/solutions/internal/forwarder/... 2>&1 | tail -15
make test && make test-race && go vet ./lessons/29-capstone-final/... && golangci-lint run ./lessons/29-capstone-final/...

# Manual two-process demo
go build -o /tmp/logstatsd ./lessons/29-capstone-final/solutions/cmd/logstatsd
go build -o /tmp/fwd ./lessons/29-capstone-final/solutions/cmd/logstats-forward
/tmp/logstatsd -http-addr=127.0.0.1:8080 -grpc-addr=127.0.0.1:9090 >/dev/null 2>&1 &
LSD=$!; sleep 0.6
printf '2026-01-02T15:04:05 INFO a\n2026-01-02T15:04:06 WARN b\n' | /tmp/fwd -addr=http://127.0.0.1:8080/ingest -id=fwd1
curl -s 127.0.0.1:8080/stats; echo
kill -TERM $LSD; rm -f /tmp/logstatsd /tmp/fwd

git add lessons/29-capstone-final/exercises/internal/forwarder lessons/29-capstone-final/exercises/cmd/logstats-forward lessons/29-capstone-final/solutions/internal/forwarder lessons/29-capstone-final/solutions/cmd/logstats-forward
git commit -m "feat(lesson-29): internal/forwarder (outbox + at-least-once retry) + cmd/logstats-forward"
```

Expected: exercises vacuous-pass; solutions TestRetriesTransient + TestEffectivelyOnce PASS; -race clean; manual demo shows `{"counts":{"INFO":1,"WARN":1},"total":2}`.

---

## Task 5: Slide deck — 5 concepts incl. course wrap-up (controller inline)

**File:** `lessons/29-capstone-final/slides/slides.md`. Concepts:

1. **At-least-once & why exactly-once is a myth** — the two-generals/lost-ack problem; you get at-most-once (lossy) or at-least-once (dup-prone); "exactly-once delivery" is marketing — what you build is at-least-once + idempotency. *Mistake:* assuming the network delivers exactly once.
2. **Idempotency keys + server-side dedup** — a stable key per unit of work; the server records seen keys and skips duplicates; bounded memory. *Mistake:* a fresh key per retry (no dedup); unbounded dedup state.
3. **The outbox pattern** — record intent before sending; retry from the outbox until acked; remove on ack. *Mistake:* send-then-record (crash loses it); never pruning.
4. **Retries + dedup together** — at-least-once retry (same key) + dedup = effectively-once; the whole pipeline; the counter-example (fresh key → double-count). *Mistake:* retrying with a new key.
5. **Course wrap-up** — Phases 1-4 retrospective: L1 `go run hello` → types/interfaces/errors/generics (P1-2) → goroutines/channels/sync/context/concurrency (P3) → HTTP/gRPC/config/containers/observability/distributed (P4). The running example's arc: expense tracker → log aggregator → `logstats` service → distributed pipeline. What to learn next (databases, queues, k8s, the things we deliberately left out). Zero→two third-party deps, on purpose.

- [ ] Author (title-slide-grid, "What we'll cover", "The story so far — the finale", 5 concepts, Practice, a Closing thought celebrating the whole course, What we learned, and an "Up next" that points OUTWARD — the course is done, here's where to go). Then build + commit.

---

## Task 6: README + verify + final review + PR (controller)

- [ ] **Step 1:** Confirm build-index L29 (`Slug:"capstone-final"`) — matches; no change.
- [ ] **Step 2:** Write `lessons/29-capstone-final/README.md` (~300 lines): "What's different / the finale"; the distributed pipeline (forwarder → aggregator, idempotency keys, dedup); the at-least-once + idempotency = effectively-once principle; the two-process run demo; a course-completion section (the 29-lesson arc, the running examples, what's intentionally out of scope + where to go next); "going further" (persistent outbox + crash recovery; a real broker (NATS/Kafka) swap; exactly-once semantics caveats; dedup TTL/sharding; backpressure).
- [ ] **Step 3:** Full sweep (`make test` + `make test-race`; solutions `-v`; `-race`; vet/lint/fmt; `grep '^go ' go.mod` → 1.23.0; slides-build + index; the two-process manual demo).
- [ ] **Step 4:** Commit README.
- [ ] **Step 5:** Dispatch `feature-dev:code-reviewer` over `git diff main...HEAD`. Focus: dedup correctness (bounded eviction, concurrency); idempotent ingest (key check before merge, no-key back-compat, Duplicate field); forwarder (key reused across retries — the crux; transient vs permanent classification; outbox order; ctx-aware backoff); the effectively-once test actually proves the property (merge-once despite redelivery); deps/go-directive unchanged; test determinism (httptest, atomic, tiny backoff, no rand); exercises skeletons vacuous; slide/README accuracy + the course-wrap-up claims. Apply fixes.
- [ ] **Step 6:** Push + open PR (no co-author trailer); watch CI green. **This is the final lesson — the PR body notes the course (29 lessons) is complete.**

---

## Verification (after Task 6)

```bash
make test && make test-race
go test ./lessons/29-capstone-final/exercises/...          # vacuous-pass
go test -v ./lessons/29-capstone-final/solutions/...       # dedup + idempotent ingest + forwarder + capstone + carried
go test -race ./lessons/29-capstone-final/...
go vet ./... && golangci-lint run ./...
gofmt -l lessons/29-capstone-final/ && grep '^go ' go.mod  # go 1.23.0
make slides-build && grep -q "29-capstone-final" dist/index.html && rm -rf dist

# Two-process capstone demo: forwarder → aggregator, duplicate counted once
go build -o /tmp/logstatsd ./lessons/29-capstone-final/solutions/cmd/logstatsd
go build -o /tmp/fwd ./lessons/29-capstone-final/solutions/cmd/logstats-forward
/tmp/logstatsd -http-addr=127.0.0.1:8080 -grpc-addr=127.0.0.1:9090 >/dev/null 2>&1 &
LSD=$!; sleep 0.6
printf '2026-01-02T15:04:05 INFO a\n2026-01-02T15:04:06 WARN b\n' | /tmp/fwd -addr=http://127.0.0.1:8080/ingest -id=fwd1
curl -s 127.0.0.1:8080/stats; echo   # {"counts":{"INFO":1,"WARN":1},"total":2}
kill -TERM $LSD; rm -f /tmp/logstatsd /tmp/fwd
```

## Critical file paths

To create: `lessons/29-capstone-final/` tree (README, slides, warmup/dedup, internal/forwarder, cmd/logstats-forward, idempotent httpsrv, + carried daemon), mirrored.
To modify: `Makefile` (proto target += L29 protos), `go.mod`/`go.sum` (tidy), the carried `httpsrv`/`cmd/logstatsd` (idempotency wiring).
To reference: `lessons/28-observability/...` (carry-forward source), `tools/build-index/main.go` (slug `capstone-final` correct).

## Execution after approval

1. (Branch `feature/plan-ff-lesson-29-capstone` already created off main.)
2. Commit this plan doc.
3. Execute: Task 1 controller-direct (carry + regen); subagents for Tasks 2-4 (dedup, idempotent-ingest wiring, forwarder); controller inline for Tasks 5-6 (slides, README); controller runs all verification incl. the two-process demo.
4. Final code review subagent. Push + PR (no co-author trailer); watch CI green. The course is complete after merge.
