# Plan AA — Lesson 24 (HTTP clients & resilience) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author lesson 24 — the second Phase 4 lesson. Build a resilient `logstats-client` that ships log lines to the L23 server's `/ingest`, surviving servers that fail, hang, and rate-limit: explicit timeouts, per-request context deadlines, retry with exponential backoff + jitter, **selective** (transient-only) retry, and a full **circuit breaker** (closed→open→half-open).

**Architecture:** Same per-lesson pattern as Plans D-Z. Seven tasks. Carries forward L23's `logparse` + `logstats` + `cmd/logstats-server` verbatim (the client's real target). New: `warmup/retry` (`Do` + `Permanent`), `internal/breaker` (circuit-breaker state machine, clock injected for deterministic tests), `internal/shipper` (the resilient client wrapping `http.Client` + retry + breaker), and `cmd/logstats-client`. Five-concept slide deck.

**Tech Stack:** Go 1.23 stdlib only (`net/http`, `context`, `time`, `errors`, `encoding/json`, `math/rand/v2`, `bufio`, `io`, `os`, `os/signal`, `sync`, `net/http/httptest` + `sync/atomic` for tests). No third-party deps in L24 (gRPC/OTel arrive at L25/L28).

---

## Scope

After Plan AA: lesson 24 complete; `make test` + `make test-race` green; the `logstats-client` ships lines to a running `logstats-server`; the retry/breaker/shipper unit + integration tests pass deterministically; `24-http-client` lands in the index.

### Design decisions (3 user-approved + plan-recommended)

**User-approved via brainstorming:**

1. **Full circuit-breaker state machine** — a standalone `internal/breaker` with closed/open/half-open, integrated into the shipper and tested across all transitions.
2. **Selective (transient-only) retry** — retry network errors / 5xx / 429; treat 4xx (non-429) as permanent (fail immediately).
3. **Five slide concepts:** `http.Client` + timeout + transport reuse · per-request context deadlines · retries + backoff + jitter · selective retry + idempotency · circuit breaker.

**Plan-recommended:**

4. **Carry forward L23's `logparse` + `logstats` + `cmd/logstats-server` verbatim** (import paths rewritten `23-http-server` → `24-http-client`) so the client has a real target and a manual end-to-end smoke works.
5. **`retry.Do` + `retry.Permanent`.** The warmup retry supports a `Permanent(err)` wrapper (mirrors `cenkalti/backoff.Permanent`) that stops the loop immediately — needed so 4xx and an open breaker short-circuit retry instead of incurring pointless backoff sleeps. (Discovered necessary during prototyping.)
6. **Breaker clock injected via `WithClock` option** so all state transitions test deterministically with a fake clock — no `time.Sleep` in tests (the CI-flake lesson from L17-L22).
7. **Breaker counts only transient failures.** A 4xx is our bad request, not the dependency being down, so the shipper routes permanent errors around the breaker's failure counter (the breaker sees a healthy dependency). Prototype-verified.
8. **429 treated as plain transient** (retried with backoff). Truly honoring `Retry-After` would require plumbing a delay hint through `retry.Do`, bloating the warmup; it's a "going further" exercise instead. (Avoids an unused-field lint failure too.)
9. **The client `run` is provided study code** in both trees (it wires stdin → shipper → summary). The exercise is `retry.Do` + `breaker` + `shipper`; the exercises client test is a skeleton (doesn't invoke `run`, which calls the skeleton `shipper.Ship`).

### Verified facts (prototyped before writing this plan)

- `retry` + `breaker` + `shipper` + a client `Run` were prototyped with full tests and pass **race-clean over 5 runs**, vet-clean, gofmt-clean.
- Verified matrix: retry (eventual success / exhaustion / permanent-stops-at-1-call / ctx-cancel-aborts-backoff); breaker (opens after threshold, half-open→closed on trial success, half-open→open on trial failure, still-open before cooldown — all via fake clock, no sleeps); shipper (success / transient-retry-then-succeed / 4xx-no-retry / 429-retried / breaker-opens-and-fails-fast / ctx-cancel); client (stdin → ship → summary via httptest).
- The breaker correctly is NOT tripped by 4xx (permanent errors bypass its counter).

---

## Plans F-Z lessons-learned applied here

1. Subpackages have their own namespace — no `Warmup*` prefix.
2. Lint exclusion covers nested subpackages (`lessons/.*/exercises/`).
3. Common-mistake content in README + slides (one per concept).
4. Slides + README written inline by the controller.
5. `make test-race` daily-habits continues — the breaker is concurrent shared state.
6. Skeleton tests pass vacuously (`panic("TODO")` impl; skeleton test bodies never invoke the panicking code).
7. Timing tests are deterministic: tiny `base` backoff for retry; injected fake clock for the breaker; `httptest` + `sync/atomic` counters for the shipper. No real-time sleeps in assertions.

---

## File structure

```
lessons/24-http-client/
├── README.md                                       (Task 7)
├── slides/
│   ├── index.html (scaffolded), slides.md, assets/.gitkeep   (Task 6)
├── exercises/
│   ├── warmup/retry/{retry.go, retry_test.go}      (Task 2 — Do SKELETON)
│   ├── internal/
│   │   ├── logparse/                               (Task 1 — carried verbatim)
│   │   ├── logstats/                               (Task 1 — carried verbatim)
│   │   ├── breaker/{breaker.go, breaker_test.go}   (Task 3 — SKELETON)
│   │   └── shipper/{shipper.go, shipper_test.go}   (Task 4 — SKELETON)
│   └── cmd/
│       ├── logstats-server/                        (Task 1 — carried verbatim)
│       └── logstats-client/{main.go, main_test.go} (Task 5 — run provided; test SKELETON)
└── solutions/   (mirrored, full implementations)
```

**File count:** ~28 (much is verbatim carry-forward).

---

## Task 1: Scaffold + carry forward L23 server stack

- [ ] **Step 1:** Scaffold and remove the flat stubs:

```bash
make new-lesson NAME=24-http-client
rm lessons/24-http-client/exercises/main.go \
   lessons/24-http-client/exercises/main_test.go \
   lessons/24-http-client/exercises/warmup.go \
   lessons/24-http-client/exercises/warmup_test.go \
   lessons/24-http-client/solutions/main.go \
   lessons/24-http-client/solutions/main_test.go \
   lessons/24-http-client/solutions/warmup.go \
   lessons/24-http-client/solutions/warmup_test.go
```

- [ ] **Step 2:** Carry forward logparse + logstats + cmd/logstats-server (both trees), rewrite import paths + lesson header:

```bash
SRC=lessons/23-http-server
DST=lessons/24-http-client
for side in exercises solutions; do
  mkdir -p "$DST/$side/internal" "$DST/$side/cmd"
  cp -R "$SRC/$side/internal/logparse" "$DST/$side/internal/logparse"
  cp -R "$SRC/$side/internal/logstats" "$DST/$side/internal/logstats"
  cp -R "$SRC/$side/cmd/logstats-server" "$DST/$side/cmd/logstats-server"
done
grep -rl '23-http-server' "$DST" | while read -r f; do
  sed -i '' 's#lessons/23-http-server#lessons/24-http-client#g' "$f"
done
grep -rl 'lesson 23' "$DST" | while read -r f; do
  sed -i '' 's/lesson 23/lesson 24/g' "$f"
done
```

- [ ] **Step 3:** Verify the carried baseline:

```bash
gofmt -l lessons/24-http-client/
go build ./lessons/24-http-client/...
go test ./lessons/24-http-client/... 2>&1 | tail -10
go vet ./lessons/24-http-client/...
golangci-lint run ./lessons/24-http-client/...
```

Expected: gofmt empty; build clean; carried logparse/logstats/server tests pass (exercises + solutions); vet + lint clean.

- [ ] **Step 4:** Commit:

```bash
git add lessons/24-http-client/exercises lessons/24-http-client/solutions
git commit -m "chore(lesson-24): scaffold + carry forward L23 logparse/logstats/server"
```

---

## Task 2: Warm-up — `retry`

`Do(ctx, attempts, base, fn)` with exponential backoff + full jitter + `Permanent` short-circuit.

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/24-http-client/{exercises,solutions}/warmup/retry`

- [ ] **Step 2:** `exercises/warmup/retry/retry.go` (SKELETON — `Do` panics; `Permanent` + the `permanent` type are provided so the package compiles and the concept is visible):

```go
// Package retry runs a function with bounded retries and exponential
// backoff. It is the lesson 24 warm-up.
package retry

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"
)

// permanent wraps an error to signal that Do must stop immediately
// rather than retry. Mirrors cenkalti/backoff.Permanent.
type permanent struct{ err error }

func (p *permanent) Error() string { return p.err.Error() }
func (p *permanent) Unwrap() error { return p.err }

// Permanent marks err as non-retryable: Do returns it (unwrapped) at
// once instead of retrying.
func Permanent(err error) error { return &permanent{err: err} }

// Do calls fn up to attempts times. On a non-nil error it sleeps with
// exponential backoff (base * 2^n) plus full jitter before the next
// try, returning early if ctx is cancelled. A fn error wrapped with
// Permanent stops the loop immediately. Returns nil on the first
// success, or the last (unwrapped) error.
//
// Hint:
//   for n := 0; n < attempts; n++ {
//       err = fn()
//       if err == nil { return nil }
//       var p *permanent
//       if errors.As(err, &p) { return p.err }      // terminal
//       if n == attempts-1 { break }                // no sleep after last
//       backoff := base << n                        // base * 2^n
//       d := time.Duration(rand.Int64N(int64(backoff) + 1))  // full jitter
//       t := time.NewTimer(d)
//       select {
//       case <-ctx.Done(): t.Stop(); return ctx.Err()
//       case <-t.C:
//       }
//   }
//   return err
func Do(ctx context.Context, attempts int, base time.Duration, fn func() error) error {
	_ = errors.As
	_ = rand.Int64N
	_ = time.NewTimer
	panic("TODO: retry loop with exponential backoff + jitter; honor ctx + Permanent")
}
```

- [ ] **Step 3:** `exercises/warmup/retry/retry_test.go` (SKELETON):

```go
package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestDo is a SKELETON. Cover: eventual success after N failures,
// exhaustion returns the last error, Permanent stops at one call, and
// ctx cancellation aborts the backoff.
func TestDo(t *testing.T) {
	// TODO:
	//   calls := 0
	//   err := Do(context.Background(), 5, time.Microsecond, func() error {
	//       calls++; if calls < 3 { return errors.New("x") }; return nil
	//   })
	//   assert err == nil && calls == 3
	_ = errors.New
	_ = context.Background
	_ = time.Microsecond
}
```

- [ ] **Step 4:** `solutions/warmup/retry/retry.go` — same as the exercises file but with `Do` implemented:

```go
// Package retry runs a function with bounded retries and exponential
// backoff. It is the lesson 24 warm-up reference implementation.
package retry

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"
)

// permanent wraps an error to signal that Do must stop immediately
// rather than retry. Mirrors cenkalti/backoff.Permanent.
type permanent struct{ err error }

func (p *permanent) Error() string { return p.err.Error() }
func (p *permanent) Unwrap() error { return p.err }

// Permanent marks err as non-retryable: Do returns it (unwrapped) at
// once instead of retrying.
func Permanent(err error) error { return &permanent{err: err} }

// Do calls fn up to attempts times. On a non-nil error it sleeps with
// exponential backoff (base * 2^n) plus full jitter before the next
// try, returning early if ctx is cancelled. A fn error wrapped with
// Permanent stops the loop immediately. Returns nil on the first
// success, or the last (unwrapped) error.
func Do(ctx context.Context, attempts int, base time.Duration, fn func() error) error {
	var err error
	for n := 0; n < attempts; n++ {
		err = fn()
		if err == nil {
			return nil
		}
		var p *permanent
		if errors.As(err, &p) {
			return p.err // terminal — do not retry
		}
		if n == attempts-1 {
			break // no sleep after the final attempt
		}
		// Full jitter: sleep in [0, base*2^n].
		backoff := base << n
		d := time.Duration(rand.Int64N(int64(backoff) + 1))
		t := time.NewTimer(d)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}
	}
	return err
}
```

- [ ] **Step 5:** `solutions/warmup/retry/retry_test.go`:

```go
package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDoEventualSuccess(t *testing.T) {
	calls := 0
	err := Do(context.Background(), 5, time.Microsecond, func() error {
		calls++
		if calls < 3 {
			return errors.New("transient")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Do = %v, want nil", err)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

func TestDoExhausted(t *testing.T) {
	calls := 0
	want := errors.New("always")
	err := Do(context.Background(), 3, time.Microsecond, func() error {
		calls++
		return want
	})
	if !errors.Is(err, want) {
		t.Errorf("Do = %v, want %v", err, want)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

func TestDoPermanentStopsEarly(t *testing.T) {
	calls := 0
	sentinel := errors.New("bad request")
	err := Do(context.Background(), 5, time.Microsecond, func() error {
		calls++
		return Permanent(sentinel)
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("Do = %v, want %v (unwrapped)", err, sentinel)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (permanent stops immediately)", calls)
	}
}

func TestDoContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := Do(ctx, 10, 50*time.Millisecond, func() error {
		calls++
		cancel() // cancel during the first attempt; backoff should abort
		return errors.New("transient")
	})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Do = %v, want context.Canceled", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (cancel aborts backoff)", calls)
	}
}
```

- [ ] **Step 6:** Verify + commit:

```bash
gofmt -l lessons/24-http-client/
go test ./lessons/24-http-client/exercises/warmup/retry/... 2>&1 | tail -5
go test -v ./lessons/24-http-client/solutions/warmup/retry/... 2>&1 | tail -20
go vet ./lessons/24-http-client/... && golangci-lint run ./lessons/24-http-client/...
make test

git add lessons/24-http-client/exercises/warmup lessons/24-http-client/solutions/warmup
git commit -m "feat(lesson-24): warmup — retry (Do with exp backoff + jitter + Permanent)"
```

Expected: exercises vacuous-pass; solutions 4 tests PASS.

---

## Task 3: `internal/breaker`

The circuit-breaker state machine. Clock injected via `WithClock` for deterministic tests.

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/24-http-client/{exercises,solutions}/internal/breaker`

- [ ] **Step 2:** `exercises/internal/breaker/breaker.go` (SKELETON — the type/states/constructor/options are provided; `Call`/`allow`/`record` panic):

```go
// Package breaker implements a circuit breaker: it fails fast when a
// dependency is unhealthy, giving it room to recover.
package breaker

import (
	"errors"
	"sync"
	"time"
)

// ErrOpen is returned by Call when the breaker is open (failing fast).
var ErrOpen = errors.New("breaker: circuit open")

type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

func (s State) String() string {
	switch s {
	case Closed:
		return "closed"
	case Open:
		return "open"
	case HalfOpen:
		return "half-open"
	}
	return "unknown"
}

// Breaker is a circuit breaker: closed → open (after maxFailures
// consecutive failures) → half-open (after cooldown) → closed (on a
// trial success) or open (on a trial failure).
type Breaker struct {
	maxFailures int
	cooldown    time.Duration
	now         func() time.Time

	mu       sync.Mutex
	state    State
	failures int
	openedAt time.Time
}

// Option configures a Breaker.
type Option func(*Breaker)

// WithClock overrides the time source (for tests).
func WithClock(now func() time.Time) Option {
	return func(b *Breaker) { b.now = now }
}

func New(maxFailures int, cooldown time.Duration, opts ...Option) *Breaker {
	b := &Breaker{
		maxFailures: maxFailures,
		cooldown:    cooldown,
		now:         time.Now,
		state:       Closed,
	}
	for _, o := range opts {
		o(b)
	}
	return b
}

// Call runs fn unless the breaker is open (then returns ErrOpen without
// calling fn). It updates state from the result: a success closes the
// breaker and resets the failure count; a failure increments it and
// opens the breaker at the threshold (or immediately, if half-open).
//
// Hint: gate on allow(); call fn; record(err); return.
func (b *Breaker) Call(fn func() error) error {
	_ = ErrOpen
	panic("TODO: if !allow() return ErrOpen; err := fn(); record(err); return err")
}

// allow reports whether a call may proceed, transitioning Open →
// HalfOpen when the cooldown has elapsed (allowing one trial).
func (b *Breaker) allow() bool {
	panic("TODO: lock; if Open and now-openedAt >= cooldown → HalfOpen, allow one trial; Open → deny; else allow")
}

// record updates state from a call's outcome.
func (b *Breaker) record(err error) {
	panic("TODO: lock; on error → failures++, open if HalfOpen or failures>=max (stamp openedAt); on success → reset to Closed")
}

// State returns the current state (mainly for tests/introspection).
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}
```

- [ ] **Step 3:** `exercises/internal/breaker/breaker_test.go` (SKELETON):

```go
package breaker

import (
	"errors"
	"testing"
	"time"
)

// fakeClock is a manually-advanced time source for deterministic tests.
type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time      { return c.t }
func (c *fakeClock) add(d time.Duration) { c.t = c.t.Add(d) }

// TestBreaker is a SKELETON. Cover: opens after maxFailures; fails fast
// while open (fn not called); half-open→closed on trial success;
// half-open→open on trial failure. Use WithClock(clk.now) + clk.add to
// drive the cooldown deterministically.
func TestBreaker(t *testing.T) {
	// TODO:
	//   clk := &fakeClock{t: time.Unix(0, 0)}
	//   b := New(3, time.Minute, WithClock(clk.now))
	//   ... fail 3 times, assert Open, assert fast-fail, advance clock, trial ...
	_ = errors.New
	_ = New
}
```

- [ ] **Step 4:** `solutions/internal/breaker/breaker.go` — identical to the exercises file but with `Call`/`allow`/`record` implemented:

```go
// Package breaker implements a circuit breaker: it fails fast when a
// dependency is unhealthy, giving it room to recover. Reference impl.
package breaker

import (
	"errors"
	"sync"
	"time"
)

// ErrOpen is returned by Call when the breaker is open (failing fast).
var ErrOpen = errors.New("breaker: circuit open")

type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

func (s State) String() string {
	switch s {
	case Closed:
		return "closed"
	case Open:
		return "open"
	case HalfOpen:
		return "half-open"
	}
	return "unknown"
}

// Breaker is a circuit breaker: closed → open (after maxFailures
// consecutive failures) → half-open (after cooldown) → closed (on a
// trial success) or open (on a trial failure).
type Breaker struct {
	maxFailures int
	cooldown    time.Duration
	now         func() time.Time

	mu       sync.Mutex
	state    State
	failures int
	openedAt time.Time
}

// Option configures a Breaker.
type Option func(*Breaker)

// WithClock overrides the time source (for tests).
func WithClock(now func() time.Time) Option {
	return func(b *Breaker) { b.now = now }
}

func New(maxFailures int, cooldown time.Duration, opts ...Option) *Breaker {
	b := &Breaker{
		maxFailures: maxFailures,
		cooldown:    cooldown,
		now:         time.Now,
		state:       Closed,
	}
	for _, o := range opts {
		o(b)
	}
	return b
}

// Call runs fn unless the breaker is open. It updates state based on
// the result. When open and the cooldown hasn't elapsed it returns
// ErrOpen without calling fn.
func (b *Breaker) Call(fn func() error) error {
	if !b.allow() {
		return ErrOpen
	}
	err := fn()
	b.record(err)
	return err
}

// allow reports whether a call may proceed, transitioning Open →
// HalfOpen when the cooldown has elapsed.
func (b *Breaker) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state == Open {
		if b.now().Sub(b.openedAt) >= b.cooldown {
			b.state = HalfOpen
			return true // allow a single trial
		}
		return false
	}
	return true // Closed or HalfOpen
}

func (b *Breaker) record(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if err != nil {
		b.failures++
		// A failure in HalfOpen, or hitting the threshold in Closed,
		// (re)opens the breaker.
		if b.state == HalfOpen || b.failures >= b.maxFailures {
			b.state = Open
			b.openedAt = b.now()
		}
		return
	}
	// Success resets.
	b.failures = 0
	b.state = Closed
}

// State returns the current state (mainly for tests/introspection).
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}
```

- [ ] **Step 5:** `solutions/internal/breaker/breaker_test.go`:

```go
package breaker

import (
	"errors"
	"testing"
	"time"
)

// fakeClock is a manually-advanced time source for deterministic tests.
type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time      { return c.t }
func (c *fakeClock) add(d time.Duration) { c.t = c.t.Add(d) }

func TestBreakerOpensAfterThreshold(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	b := New(3, time.Minute, WithClock(clk.now))
	boom := errors.New("boom")

	for i := 0; i < 2; i++ {
		if err := b.Call(func() error { return boom }); err != boom {
			t.Fatalf("call %d = %v, want boom", i, err)
		}
		if b.State() != Closed {
			t.Fatalf("after %d failures state = %s, want closed", i+1, b.State())
		}
	}
	_ = b.Call(func() error { return boom }) // 3rd consecutive failure trips it
	if b.State() != Open {
		t.Fatalf("state = %s, want open", b.State())
	}
	called := false
	if err := b.Call(func() error { called = true; return nil }); !errors.Is(err, ErrOpen) {
		t.Errorf("open call = %v, want ErrOpen", err)
	}
	if called {
		t.Error("fn was invoked while breaker open")
	}
}

func TestBreakerHalfOpenToClosed(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	b := New(1, time.Minute, WithClock(clk.now))
	b.Call(func() error { return errors.New("boom") }) // opens (maxFailures=1)
	if b.State() != Open {
		t.Fatalf("state = %s, want open", b.State())
	}
	clk.add(time.Minute) // cooldown elapses
	if err := b.Call(func() error { return nil }); err != nil {
		t.Fatalf("trial call = %v, want nil", err)
	}
	if b.State() != Closed {
		t.Errorf("state = %s, want closed", b.State())
	}
}

func TestBreakerHalfOpenToOpen(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	b := New(1, time.Minute, WithClock(clk.now))
	boom := errors.New("boom")
	b.Call(func() error { return boom }) // opens
	clk.add(time.Minute)                 // cooldown elapses → next call is a trial
	if err := b.Call(func() error { return boom }); err != boom {
		t.Fatalf("trial = %v, want boom", err)
	}
	if b.State() != Open {
		t.Errorf("state = %s, want open (trial failure re-opens)", b.State())
	}
	if err := b.Call(func() error { return nil }); !errors.Is(err, ErrOpen) {
		t.Errorf("call right after re-open = %v, want ErrOpen", err)
	}
}

func TestBreakerStillOpenBeforeCooldown(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	b := New(1, time.Minute, WithClock(clk.now))
	b.Call(func() error { return errors.New("boom") }) // opens
	clk.add(30 * time.Second)                          // less than cooldown
	if err := b.Call(func() error { return nil }); !errors.Is(err, ErrOpen) {
		t.Errorf("call before cooldown = %v, want ErrOpen", err)
	}
}
```

- [ ] **Step 6:** Verify + commit:

```bash
gofmt -l lessons/24-http-client/
go test ./lessons/24-http-client/exercises/internal/breaker/... 2>&1 | tail -5
go test -v ./lessons/24-http-client/solutions/internal/breaker/... 2>&1 | tail -25
go test -race ./lessons/24-http-client/solutions/internal/breaker/... 2>&1 | tail -5
make test-race
go vet ./lessons/24-http-client/... && golangci-lint run ./lessons/24-http-client/...

git add lessons/24-http-client/exercises/internal/breaker lessons/24-http-client/solutions/internal/breaker
git commit -m "feat(lesson-24): internal/breaker (closed/open/half-open, injected clock)"
```

Expected: exercises vacuous-pass; solutions 4 tests PASS; -race clean.

---

## Task 4: `internal/shipper`

The resilient client: `http.Client{Timeout}` + selective retry + breaker. Depends on `retry` + `breaker`.

**Files (4 total).** Solution import paths: `.../solutions/internal/{breaker,retry}` — note `retry` lives under `warmup/`, so the import is `.../solutions/warmup/retry`.

- [ ] **Step 1:** `mkdir -p lessons/24-http-client/{exercises,solutions}/internal/shipper`

- [ ] **Step 2:** `exercises/internal/shipper/shipper.go` (SKELETON — types/options/`New`/`doOnce` provided; `Ship` panics, since the retry+breaker orchestration is the exercise). Use the exercises import paths (`.../exercises/warmup/retry`, `.../exercises/internal/breaker`):

```go
// Package shipper is a resilient HTTP client that POSTs log lines to the
// logstats server's /ingest, with timeouts, selective retry, and a
// circuit breaker.
package shipper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ristkari-dev/go-training/lessons/24-http-client/exercises/internal/breaker"
	"github.com/ristkari-dev/go-training/lessons/24-http-client/exercises/warmup/retry"
)

type ingestRequest struct {
	Lines []string `json:"lines"`
}

// Result mirrors the server's ingest response.
type Result struct {
	Accepted int `json:"accepted"`
	Parsed   int `json:"parsed"`
	Failed   int `json:"failed"`
}

// permanentError marks a non-retryable failure (4xx other than 429).
type permanentError struct{ status int }

func (e *permanentError) Error() string {
	return fmt.Sprintf("shipper: permanent failure: HTTP %d", e.status)
}

// transientError marks a retryable failure (5xx, 429, network).
type transientError struct {
	status int // 0 for a network error
	err    error
}

func (e *transientError) Error() string {
	if e.status != 0 {
		return fmt.Sprintf("shipper: transient failure: HTTP %d", e.status)
	}
	return fmt.Sprintf("shipper: transient failure: %v", e.err)
}

type Shipper struct {
	url         string
	client      *http.Client
	breaker     *breaker.Breaker
	maxAttempts int
	baseBackoff time.Duration
}

type Option func(*Shipper)

func WithBreaker(b *breaker.Breaker) Option  { return func(s *Shipper) { s.breaker = b } }
func WithMaxAttempts(n int) Option           { return func(s *Shipper) { s.maxAttempts = n } }
func WithBaseBackoff(d time.Duration) Option { return func(s *Shipper) { s.baseBackoff = d } }
func WithClient(c *http.Client) Option       { return func(s *Shipper) { s.client = c } }

func New(url string, opts ...Option) *Shipper {
	s := &Shipper{
		url:         url,
		client:      &http.Client{Timeout: 10 * time.Second},
		breaker:     breaker.New(5, 30*time.Second),
		maxAttempts: 3,
		baseBackoff: 100 * time.Millisecond,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Ship POSTs lines to /ingest, retrying transient failures with
// backoff, short-circuiting via the breaker. IMPLEMENT THIS.
//
// Layering you must get right:
//   - doOnce classifies the outcome into success / *transientError /
//     *permanentError.
//   - Wrap each attempt in s.breaker.Call so an open breaker fails fast.
//   - The breaker must count ONLY transient failures — a 4xx is our bad
//     request, not the dependency being down. So inside the breaker's
//     fn: on a *permanentError, capture it and return nil (healthy);
//     on a transient error, return it (breaker counts it).
//   - After breaker.Call: a captured permanent error → retry.Permanent
//     (stop now); breaker.ErrOpen → retry.Permanent (fail fast, no
//     backoff); otherwise return the (nil or transient) error to retry.Do.
//
// Hint: marshal once; var result Result; runErr := retry.Do(...); return result, runErr.
func (s *Shipper) Ship(ctx context.Context, lines []string) (Result, error) {
	_ = json.Marshal
	_ = retry.Permanent
	_ = errors.Is
	panic("TODO: retry.Do wrapping breaker.Call(doOnce); selective retry; breaker counts only transient")
}

// doOnce performs a single POST and classifies the outcome.
func (s *Shipper) doOnce(ctx context.Context, body []byte) (Result, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return Result{}, &transientError{err: err} // network error → transient
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		var r Result
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return Result{}, &transientError{err: err}
		}
		return r, nil
	case resp.StatusCode == http.StatusTooManyRequests:
		return Result{}, &transientError{status: resp.StatusCode} // 429 → retry with backoff
	case resp.StatusCode >= 500:
		return Result{}, &transientError{status: resp.StatusCode}
	default: // 4xx other than 429
		return Result{}, &permanentError{status: resp.StatusCode}
	}
}
```

- [ ] **Step 3:** `exercises/internal/shipper/shipper_test.go` (SKELETON):

```go
package shipper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestShip is a SKELETON. Use httptest servers to cover: success;
// transient (5xx) retried then succeeds; permanent (400) not retried;
// 429 retried; breaker opens after repeated 5xx and then fails fast.
func TestShip(t *testing.T) {
	// TODO: httptest.NewServer(...) returning 200/500/400/429; assert
	// Ship behavior + server hit counts (use sync/atomic counters).
	_ = http.StatusOK
	_ = httptest.NewServer
	_ = context.Background
}
```

- [ ] **Step 4:** `solutions/internal/shipper/shipper.go` — same as exercises but with `Ship` implemented and the `.../solutions/...` import paths:

```go
// Package shipper is a resilient HTTP client that POSTs log lines to the
// logstats server's /ingest, with timeouts, selective retry, and a
// circuit breaker. Reference implementation.
package shipper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ristkari-dev/go-training/lessons/24-http-client/solutions/internal/breaker"
	"github.com/ristkari-dev/go-training/lessons/24-http-client/solutions/warmup/retry"
)

type ingestRequest struct {
	Lines []string `json:"lines"`
}

// Result mirrors the server's ingest response.
type Result struct {
	Accepted int `json:"accepted"`
	Parsed   int `json:"parsed"`
	Failed   int `json:"failed"`
}

// permanentError marks a non-retryable failure (4xx other than 429).
type permanentError struct{ status int }

func (e *permanentError) Error() string {
	return fmt.Sprintf("shipper: permanent failure: HTTP %d", e.status)
}

// transientError marks a retryable failure (5xx, 429, network).
type transientError struct {
	status int // 0 for a network error
	err    error
}

func (e *transientError) Error() string {
	if e.status != 0 {
		return fmt.Sprintf("shipper: transient failure: HTTP %d", e.status)
	}
	return fmt.Sprintf("shipper: transient failure: %v", e.err)
}

type Shipper struct {
	url         string
	client      *http.Client
	breaker     *breaker.Breaker
	maxAttempts int
	baseBackoff time.Duration
}

type Option func(*Shipper)

func WithBreaker(b *breaker.Breaker) Option  { return func(s *Shipper) { s.breaker = b } }
func WithMaxAttempts(n int) Option           { return func(s *Shipper) { s.maxAttempts = n } }
func WithBaseBackoff(d time.Duration) Option { return func(s *Shipper) { s.baseBackoff = d } }
func WithClient(c *http.Client) Option       { return func(s *Shipper) { s.client = c } }

func New(url string, opts ...Option) *Shipper {
	s := &Shipper{
		url:         url,
		client:      &http.Client{Timeout: 10 * time.Second},
		breaker:     breaker.New(5, 30*time.Second),
		maxAttempts: 3,
		baseBackoff: 100 * time.Millisecond,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Ship POSTs lines to the server's /ingest, retrying transient
// failures with backoff, short-circuiting via the breaker. Returns the
// parsed Result on success.
//
// Layering: the breaker counts only TRANSIENT failures (5xx/429/network)
// — a 4xx is our bad request, not the dependency being down, so it must
// not trip the breaker. Permanent (4xx) errors and an open breaker both
// stop the retry loop immediately via retry.Permanent.
func (s *Shipper) Ship(ctx context.Context, lines []string) (Result, error) {
	body, err := json.Marshal(ingestRequest{Lines: lines})
	if err != nil {
		return Result{}, err
	}

	var result Result
	runErr := retry.Do(ctx, s.maxAttempts, s.baseBackoff, func() error {
		var permanent error
		callErr := s.breaker.Call(func() error {
			res, e := s.doOnce(ctx, body)
			if e != nil {
				var p *permanentError
				if errors.As(e, &p) {
					permanent = e // capture; don't let it trip the breaker
					return nil    // breaker sees a healthy dependency
				}
				return e // transient → breaker counts it, retry will retry
			}
			result = res
			return nil
		})
		if permanent != nil {
			return retry.Permanent(permanent) // 4xx → stop now
		}
		if errors.Is(callErr, breaker.ErrOpen) {
			return retry.Permanent(callErr) // open → fail fast, no backoff
		}
		return callErr // nil (success) or transient (retry)
	})
	return result, runErr
}

// doOnce performs a single POST and classifies the outcome.
func (s *Shipper) doOnce(ctx context.Context, body []byte) (Result, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return Result{}, &transientError{err: err} // network error → transient
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		var r Result
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return Result{}, &transientError{err: err}
		}
		return r, nil
	case resp.StatusCode == http.StatusTooManyRequests:
		return Result{}, &transientError{status: resp.StatusCode} // 429 → retry with backoff
	case resp.StatusCode >= 500:
		return Result{}, &transientError{status: resp.StatusCode}
	default: // 4xx other than 429
		return Result{}, &permanentError{status: resp.StatusCode}
	}
}
```

- [ ] **Step 5:** `solutions/internal/shipper/shipper_test.go`:

```go
package shipper

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ristkari-dev/go-training/lessons/24-http-client/solutions/internal/breaker"
)

func testShipper(url string, opts ...Option) *Shipper {
	base := []Option{
		WithMaxAttempts(4),
		WithBaseBackoff(time.Microsecond), // keep tests fast
	}
	return New(url, append(base, opts...)...)
}

func TestShipSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"accepted":2,"parsed":2,"failed":0}`))
	}))
	defer srv.Close()

	res, err := testShipper(srv.URL).Ship(context.Background(), []string{"a", "b"})
	if err != nil {
		t.Fatalf("Ship = %v", err)
	}
	if res.Accepted != 2 || res.Parsed != 2 {
		t.Errorf("res = %+v", res)
	}
}

func TestShipRetriesTransientThenSucceeds(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable) // 503 → transient
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"accepted":1,"parsed":1,"failed":0}`))
	}))
	defer srv.Close()

	res, err := testShipper(srv.URL).Ship(context.Background(), []string{"x"})
	if err != nil {
		t.Fatalf("Ship = %v", err)
	}
	if res.Parsed != 1 {
		t.Errorf("res = %+v", res)
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Errorf("server calls = %d, want 3", calls)
	}
}

func TestShipPermanentNoRetry(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadRequest) // 400 → permanent
	}))
	defer srv.Close()

	_, err := testShipper(srv.URL).Ship(context.Background(), []string{"x"})
	if err == nil {
		t.Fatal("Ship = nil, want permanent error")
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("server calls = %d, want 1 (no retry on 4xx)", calls)
	}
}

func TestShip429IsTransient(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 2 {
			w.WriteHeader(http.StatusTooManyRequests) // 429 → transient
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"accepted":1,"parsed":1,"failed":0}`))
	}))
	defer srv.Close()

	_, err := testShipper(srv.URL).Ship(context.Background(), []string{"x"})
	if err != nil {
		t.Fatalf("Ship = %v", err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Errorf("server calls = %d, want 2 (429 retried)", calls)
	}
}

func TestShipBreakerOpensAndFailsFast(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError) // always 500
	}))
	defer srv.Close()

	b := breaker.New(3, time.Minute) // opens after 3 consecutive failures
	s := testShipper(srv.URL, WithBreaker(b), WithMaxAttempts(10))

	_, err := s.Ship(context.Background(), []string{"x"})
	if err == nil {
		t.Fatal("Ship = nil, want error")
	}
	if got := atomic.LoadInt32(&calls); got > 3 {
		t.Errorf("server calls = %d, want <= 3 (breaker should stop the storm)", got)
	}
	if b.State() != breaker.Open {
		t.Errorf("breaker state = %s, want open", b.State())
	}
	before := atomic.LoadInt32(&calls)
	_, err = s.Ship(context.Background(), []string{"y"})
	if !errors.Is(err, breaker.ErrOpen) {
		t.Errorf("second Ship = %v, want ErrOpen", err)
	}
	if atomic.LoadInt32(&calls) != before {
		t.Errorf("server was hit while breaker open")
	}
}

func TestShipContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled
	_, err := testShipper(srv.URL, WithBaseBackoff(time.Second)).Ship(ctx, []string{"x"})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Ship = %v, want context.Canceled", err)
	}
}
```

- [ ] **Step 6:** Verify + commit:

```bash
gofmt -l lessons/24-http-client/
go test ./lessons/24-http-client/exercises/internal/shipper/... 2>&1 | tail -5
go test -v ./lessons/24-http-client/solutions/internal/shipper/... 2>&1 | tail -25
go test -race ./lessons/24-http-client/solutions/internal/shipper/... 2>&1 | tail -5
make test-race
go vet ./lessons/24-http-client/... && golangci-lint run ./lessons/24-http-client/...

git add lessons/24-http-client/exercises/internal/shipper lessons/24-http-client/solutions/internal/shipper
git commit -m "feat(lesson-24): internal/shipper (timeout + selective retry + breaker)"
```

Expected: exercises vacuous-pass; solutions 6 tests PASS; -race clean.

---

## Task 5: `cmd/logstats-client`

Provided study code (`run` wires stdin → shipper → summary) in BOTH trees; exercises test is a skeleton (doesn't invoke `run`, which calls the skeleton `shipper.Ship`); solutions test is real.

**Files (4 total).** Use the per-tree import path for `shipper`.

- [ ] **Step 1:** `mkdir -p lessons/24-http-client/{exercises,solutions}/cmd/logstats-client`

- [ ] **Step 2:** `exercises/cmd/logstats-client/main.go` — provided `run` (exercises import path):

```go
// Package main is the lesson 24 logstats client: it reads log lines
// from stdin (or -file) and ships them to a logstats server's /ingest
// using the resilient shipper (timeouts, selective retry, breaker).
//
// Usage:
//   logstats-client -addr=http://localhost:8080/ingest < app.log
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/ristkari-dev/go-training/lessons/24-http-client/exercises/internal/shipper"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	addr := flag.String("addr", "http://localhost:8080/ingest", "ingest endpoint URL")
	flag.Parse()

	if err := run(ctx, *addr, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run reads all lines from stdin, ships them in one batch, and prints
// the server's summary. (A production client would batch by size and
// stream; one batch keeps the lesson focused on resilience.)
func run(ctx context.Context, addr string, stdin io.Reader, stdout io.Writer) error {
	var lines []string
	sc := bufio.NewScanner(stdin)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return err
	}

	res, err := shipper.New(addr).Ship(ctx, lines)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "shipped: accepted=%d parsed=%d failed=%d\n", res.Accepted, res.Parsed, res.Failed)
	return nil
}
```

- [ ] **Step 3:** `exercises/cmd/logstats-client/main_test.go` (SKELETON — does not call `run`):

```go
package main

import (
	"net/http/httptest"
	"testing"
)

// TestRun is a SKELETON. Spin up an httptest server returning an ingest
// response; call run with a strings.Reader of log lines; assert the
// printed summary. (Until shipper.Ship is implemented, run panics — so
// this skeleton only references httptest to keep the import alive.)
func TestRun(t *testing.T) {
	// TODO: httptest server → run(ctx, srv.URL, strings.NewReader(...), &buf)
	//       assert buf contains "accepted=... parsed=... failed=..."
	_ = httptest.NewServer
}
```

- [ ] **Step 4:** `solutions/cmd/logstats-client/main.go` — identical to the exercises file but with the `.../solutions/...` import path.

- [ ] **Step 5:** `solutions/cmd/logstats-client/main_test.go`:

```go
package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"accepted":3,"parsed":2,"failed":1}`))
	}))
	defer srv.Close()

	var out bytes.Buffer
	in := strings.NewReader("2026-01-02T15:04:05 INFO ok\n2026-01-02T15:04:06 ERROR boom\nbad\n")
	if err := run(context.Background(), srv.URL, in, &out); err != nil {
		t.Fatalf("run = %v", err)
	}
	if !strings.Contains(out.String(), "accepted=3 parsed=2 failed=1") {
		t.Errorf("out = %q", out.String())
	}
}
```

- [ ] **Step 6:** Verify + commit:

```bash
gofmt -l lessons/24-http-client/
go test ./lessons/24-http-client/exercises/cmd/logstats-client/... 2>&1 | tail -5
go test -v ./lessons/24-http-client/solutions/cmd/logstats-client/... 2>&1 | tail -10
make test
make test-race
go vet ./lessons/24-http-client/... && golangci-lint run ./lessons/24-http-client/...

# Manual end-to-end: real server + real client
go build -o /tmp/logstats-server ./lessons/24-http-client/solutions/cmd/logstats-server
go build -o /tmp/logstats-client ./lessons/24-http-client/solutions/cmd/logstats-client
/tmp/logstats-server 127.0.0.1:8282 &
sleep 0.5
printf '2026-01-02T15:04:05 INFO ok\n2026-01-02T15:04:06 ERROR boom\nbad\n' | /tmp/logstats-client -addr=http://127.0.0.1:8282/ingest
curl -s 127.0.0.1:8282/stats; echo
kill %1; rm -f /tmp/logstats-server /tmp/logstats-client

git add lessons/24-http-client/exercises/cmd/logstats-client lessons/24-http-client/solutions/cmd/logstats-client
git commit -m "feat(lesson-24): cmd/logstats-client (stdin → resilient shipper → summary)"
```

Expected: exercises vacuous-pass; solutions test PASS; smoke shows `shipped: accepted=3 parsed=2 failed=1` then `{"counts":{"ERROR":1,"INFO":1},"total":2}`.

---

## Task 6: Slide deck — 5 concepts

Heavy-explanatory per concept (motivation → basics → worked example → common mistake → recap). ~30 slides.

**File:** `lessons/24-http-client/slides/slides.md`

Concept order:

1. **`http.Client` + Timeout + Transport reuse** — `&http.Client{Timeout: ...}`; reuse one client (connection pooling via `http.Transport`). *Common mistake:* `http.DefaultClient` / `http.Get` with no timeout (hang forever); a fresh client per request (no pooling, fd churn).
2. **Per-request context deadlines** — `http.NewRequestWithContext`; ctx cancellation aborts the in-flight request; deadlines compose with `Client.Timeout`. *Common mistake:* relying only on `Client.Timeout` and ignoring caller cancellation.
3. **Retries + exponential backoff + jitter** — `base·2ⁿ`; why fixed-interval retries cause thundering herds; full jitter; honor `ctx.Done()` during the wait. *Common mistake:* retry with no backoff (hammering a struggling server).
4. **Selective retry + idempotency** — retry transient (network/5xx/429), never 4xx; only retry idempotent operations; `retry.Permanent`. *Common mistake:* retrying a 400 (wasted work); retrying a non-idempotent write (duplicate side effects).
5. **Circuit breaker** — closed/open/half-open; fail fast when a dependency is down so it can recover; the breaker counts only transient failures. *Common mistake:* no breaker → retry storms turn a brief blip into a full outage.

- [ ] **Step 1-7:** Author the deck following the L21/L22/L23 format (title-slide-grid header, "What we'll cover", "The story so far" — the other side of the wire, 5 concept sections, Practice, Closing thought, What we learned, Up next → Lesson 25 gRPC).
- [ ] **Step 8:** Verify slides build + commit:

```bash
make slides-build
grep -q "24-http-client" dist/index.html && echo "✓ in index"
rm -rf dist

git add lessons/24-http-client/slides/
git commit -m "feat(lesson-24): slides — HTTP clients & resilience (5 concepts)"
```

---

## Task 7: README + verify + final review + PR

- [ ] **Step 1:** Confirm `tools/build-index/main.go` lists lesson 24 as `{Number:"24", Slug:"http-client", Title:"HTTP clients", Blurb:"retries · timeouts", Phase:4}` — slug matches the directory, so **no build-index change is needed** (verify with `grep '"24"' tools/build-index/main.go`).

- [ ] **Step 2:** Write `lessons/24-http-client/README.md` (~280 lines). Mirror the 5 concepts with per-concept Common-mistake paragraphs. Include:
  - "What's different from L23": the other side of the wire — a resilient client for the L23 service.
  - The shipper's behavior table (network/5xx/429 → retry; 4xx → permanent; breaker fail-fast).
  - "How to run": build server + client, pipe lines through, check `/stats`; the warmup; `make test-race`.
  - "Going further": honor `Retry-After` on 429 (plumb a delay hint through retry); add a max-elapsed-time budget; per-host breakers; jittered backoff variants (full vs decorrelated); a streaming/batched client.

- [ ] **Step 3:** Full verification sweep:

```bash
make test
make test-race
go test -v ./lessons/24-http-client/solutions/...
go test -race ./lessons/24-http-client/...
go vet ./...
golangci-lint run ./...
gofmt -l lessons/24-http-client/
make slides-build && grep -q "24-http-client" dist/index.html && echo "✓ index" && rm -rf dist
```

- [ ] **Step 4:** Commit README:

```bash
git add lessons/24-http-client/README.md
git commit -m "docs(lesson-24): README — HTTP clients & resilience self-study"
```

- [ ] **Step 5:** Dispatch the `feature-dev:code-reviewer` over `git diff main...HEAD`. Focus: the shipper's retry/breaker layering (breaker not tripped by 4xx; permanent + ErrOpen short-circuit retry); breaker state-machine correctness + race-safety; retry backoff honoring ctx; determinism of all timing tests (fake clock, tiny backoff, atomic counters); slide/README/code consistency; exercises skeletons pass vacuously without invoking panicking code. Apply fixes.

- [ ] **Step 6:** Push + open PR:

```bash
git push -u origin feature/plan-aa-lesson-24-http-client
gh pr create --title "Lesson 24 — HTTP clients & resilience" --body "..."
```

Then watch CI to green (the repo has a history of timing flakes; verify before reporting done).

---

## Verification (after Task 7)

```bash
make test
make test-race
go test ./lessons/24-http-client/exercises/...                      # vacuous-pass
go test -v ./lessons/24-http-client/solutions/...                   # retry + breaker + shipper + client + carried
go test -race ./lessons/24-http-client/...
go vet ./... && golangci-lint run ./...
gofmt -l lessons/24-http-client/
make slides-build && grep -q "24-http-client" dist/index.html && rm -rf dist

# Manual end-to-end
go build -o /tmp/logstats-server ./lessons/24-http-client/solutions/cmd/logstats-server
go build -o /tmp/logstats-client ./lessons/24-http-client/solutions/cmd/logstats-client
/tmp/logstats-server 127.0.0.1:8282 &
sleep 0.5
printf '2026-01-02T15:04:05 INFO ok\n2026-01-02T15:04:06 ERROR boom\nbad\n' | /tmp/logstats-client -addr=http://127.0.0.1:8282/ingest
curl -s 127.0.0.1:8282/stats; echo
kill %1; rm -f /tmp/logstats-server /tmp/logstats-client
```

## Critical file paths

To be created:
- `lessons/24-http-client/` (directory)
- `lessons/24-http-client/README.md`
- `lessons/24-http-client/slides/{index.html (scaffolded), slides.md, assets/.gitkeep}`
- `lessons/24-http-client/exercises/warmup/retry/{retry.go, retry_test.go}`
- `lessons/24-http-client/exercises/internal/breaker/{breaker.go, breaker_test.go}`
- `lessons/24-http-client/exercises/internal/shipper/{shipper.go, shipper_test.go}`
- `lessons/24-http-client/exercises/cmd/logstats-client/{main.go, main_test.go}`
- `lessons/24-http-client/exercises/internal/logparse/`, `internal/logstats/`, `cmd/logstats-server/` (carried)
- `lessons/24-http-client/solutions/...` (mirrored)

To be referenced (not modified):
- `lessons/23-http-server/solutions/...` + `exercises/...` (carry-forward source)
- `tools/build-index/main.go` (slug `http-client` already correct — no change)
- `.golangci.yml` (no changes; no third-party deps in L24)

## Execution after approval

1. (Branch `feature/plan-aa-lesson-24-http-client` already created off main.)
2. Commit this plan doc.
3. Execute the 7 tasks via subagent-driven development (controller carries forward + verifies Task 1 directly; dispatches subagents for Tasks 2-5; writes slides/README inline for Tasks 6-7; runs verification directly).
4. Final code review subagent.
5. Push and open PR; watch CI to green.
