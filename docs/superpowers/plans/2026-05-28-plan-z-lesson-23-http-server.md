# Plan Z — Lesson 23 (HTTP servers) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author lesson 23 — the first Phase 4 lesson. Turn the carried-forward log aggregator into the **`logstats` HTTP service**: `POST /ingest` (JSON log lines → parse → accumulate), `GET /stats` (JSON level counts), `GET /healthz`. Teach `net/http` + Go 1.22 `ServeMux` routing, handlers, middleware, and `slog` structured logging.

**Architecture:** Same per-lesson pattern as Plans D-Y. Six tasks. The HTTP server is built around a `sync.Mutex`-guarded accumulator (`internal/logstats`) shared across requests; `logparse` is carried forward and trimmed to its optimized core. Testable `run`/`serve` split (the L21 pattern) + `net/http/httptest` for handler/integration tests. Five-concept slide deck.

**Tech Stack:** Go 1.23 stdlib only (`net/http`, `log/slog`, `encoding/json`, `context`, `net`, `os`, `os/signal`, `sync`, `time`, `bufio`, `strings`, `net/http/httptest` for tests). No third-party deps in L23 (gRPC/OTel arrive at L25/L28).

---

## Scope

After Plan Z: lesson 23 complete; `make test` + `make test-race` green; the `logstats-server` binary runs and serves the three endpoints; handler + integration tests pass; `23-http-server` lands in the index.

### Design decisions (2 user-approved + plan-recommended)

**User-approved via brainstorming:**

1. **JSON in / JSON out.** `POST /ingest` body is `{"lines":["...","..."]}`; `GET /stats` returns `{"counts":{...},"total":N}`. Symmetric, teaches `json.Decoder` on the request, and maps cleanly to L25's protobuf message later.
2. **Five slide concepts:** net/http server + ServeMux 1.22 routing · handlers + request/response · middleware · `slog` · designing the service.

**Plan-recommended:**

3. **`logparse` carried + trimmed.** Carry L22's `logparse`, drop the L22 profiling/fuzzing scaffolding (`parseLineRegex`, `logLineRE`, `logparse_bench_test.go`, `fuzz_test.go`), and **export `ParseLine`** (was `parseLineFast`) so the ingest handler can parse lines individually for lenient counting. `Parse` now calls `ParseLine`. Removing the now-unused regex oracle keeps `unused` happy.
4. **Lenient ingest.** Malformed lines are counted (`failed`), not rejected — log ingestion is messy by nature. Only a malformed JSON envelope (or oversized body) → `400`. Response: `{"accepted":N,"parsed":M,"failed":K}`.
5. **`run`/`serve` split for testability.** `run(ctx, addr, stdout)` creates the listener and calls `serve(ctx, ln, stdout)`. Tests pass a `127.0.0.1:0` listener and dial `ln.Addr()` (the L21 ephemeral-port pattern); handler tests use `httptest.NewServer(newRouter(...))`. A light `http.Server.Shutdown` on `ctx.Done()` is included (full graceful drain is L26's topic).
6. **`internal/logstats` accumulator** is the genuine shared concurrent state — `sync.Mutex` + `map[string]int` (the L18 pattern). `make test-race` proves concurrent ingest is safe.

### Verified facts (prototyped before writing this plan)

- The full server (router, handlers, middleware, `serve` lifecycle) + 7 tests pass and are **race-clean over 5 runs**.
- Go 1.22 `ServeMux` `"POST /ingest"` patterns work: a `GET /ingest` correctly returns **405** automatically.
- Lenient ingest verified: body with 4 lines (3 valid, 1 garbage) → `{"accepted":4,"parsed":3,"failed":1}`; `/stats` → `INFO:2, WARN:1, total:3`.
- 50 concurrent `/ingest` requests → `ERROR:50` with no race.
- `serve(ctx, ln, stdout)` returns `nil` on ctx cancel (`http.ErrServerClosed` swallowed).

---

## Plans F-Y lessons-learned applied here

1. Subpackages have their own namespace — no `Warmup*` prefix.
2. Lint exclusion covers nested subpackages (`lessons/.*/exercises/`).
3. Common-mistake content in README + slides (one per concept).
4. Slides + README written inline by the controller.
5. `make test-race` daily-habits continues — the accumulator is shared concurrent state.
6. Skeleton tests pass vacuously (empty case slice / commented body; `panic("TODO")` impl).
7. In-process integration via `httptest` + `net.Listen("127.0.0.1:0")`; no fixed ports.

---

## File structure

```
lessons/23-http-server/
├── README.md                                       (Task 6)
├── slides/
│   ├── index.html (scaffolded), slides.md, assets/.gitkeep   (Task 5)
├── exercises/
│   ├── warmup/middleware/
│   │   ├── middleware.go                            (Task 2 — WithRequestLog SKELETON)
│   │   └── middleware_test.go                       (Task 2 — SKELETON)
│   ├── internal/
│   │   ├── logparse/{logparse.go, logparse_test.go} (Task 1 — carried + trimmed)
│   │   └── logstats/{logstats.go, logstats_test.go} (Task 3 — Store SKELETON)
│   └── cmd/logstats-server/
│       ├── main.go                                  (Task 4 — handlers SKELETON)
│       └── main_test.go                             (Task 4 — SKELETON)
└── solutions/   (mirrored, full implementations)
```

**File count:** ~16.

---

## Task 1: Scaffold + carry-forward & trim `logparse`

- [ ] **Step 1:** Scaffold and remove the flat stubs:

```bash
make new-lesson NAME=23-http-server
rm lessons/23-http-server/exercises/main.go \
   lessons/23-http-server/exercises/main_test.go \
   lessons/23-http-server/exercises/warmup.go \
   lessons/23-http-server/exercises/warmup_test.go \
   lessons/23-http-server/solutions/main.go \
   lessons/23-http-server/solutions/main_test.go \
   lessons/23-http-server/solutions/warmup.go \
   lessons/23-http-server/solutions/warmup_test.go
```

- [ ] **Step 2:** Copy `logparse` from L22 (both trees), rewrite import paths, drop the profiling/fuzzing scaffolding:

```bash
SRC=lessons/22-profiling-fuzz
DST=lessons/23-http-server
for side in exercises solutions; do
  mkdir -p "$DST/$side/internal"
  cp -R "$SRC/$side/internal/logparse" "$DST/$side/internal/logparse"
  rm -f "$DST/$side/internal/logparse/logparse_bench_test.go" \
        "$DST/$side/internal/logparse/fuzz_test.go"
done
grep -rl '22-profiling-fuzz' "$DST" | while read -r f; do
  sed -i '' 's#lessons/22-profiling-fuzz#lessons/23-http-server#g' "$f"
done
grep -rl 'lesson 22' "$DST" | while read -r f; do
  sed -i '' 's/lesson 22/lesson 23/g' "$f"
done
```

- [ ] **Step 3:** Overwrite BOTH `logparse.go` files (exercises + solutions) with the trimmed version below. It removes `parseLineRegex` + `logLineRE` + the `regexp` import, and exports `ParseLine` (was `parseLineFast`); `Parse` calls `ParseLine`. The doc-comment package line should say "lesson 23".

```go
// Package logparse parses application log lines into structured entries.
// Carried forward from L14; L22's optimized hand-written parser is now
// simply the parser (the regex reference + fuzz/bench scaffolding from
// L22 are dropped). ParseLine is exported so callers can parse and
// count lines individually.
package logparse

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"
)

type LogEntry struct {
	Time    time.Time
	Level   string
	Message string
}

const timeLayout = "2006-01-02T15:04:05"

// Parse reads log lines from r and returns parsed entries. Stops at the
// first malformed line, returning the entries parsed so far plus an error.
func Parse(r io.Reader) ([]LogEntry, error) {
	s := bufio.NewScanner(r)
	out := []LogEntry{}
	for line := 1; s.Scan(); line++ {
		text := strings.TrimSpace(s.Text())
		if text == "" {
			continue
		}
		e, err := ParseLine(text)
		if err != nil {
			return out, fmt.Errorf("logparse: line %d: %w", line, err)
		}
		out = append(out, e)
	}
	if err := s.Err(); err != nil {
		return out, fmt.Errorf("logparse: %w", err)
	}
	return out, nil
}

// ParseLine parses a single log line. The allocation-free hand-written
// parser from L22, now exported. Never panics; returns an error for
// malformed input.
func ParseLine(s string) (LogEntry, error) {
	const tsLen = 19 // 2006-01-02T15:04:05
	if len(s) < tsLen {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}
	tsStr := s[:tsLen]
	rest := s[tsLen:]

	rest, ok := cutLeadingSpace(rest)
	if !ok {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}

	sp := -1
	for i := 0; i < len(rest); i++ {
		if isLogSpace(rest[i]) {
			sp = i
			break
		}
	}
	if sp < 0 {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}
	level := rest[:sp]
	if level != "INFO" && level != "WARN" && level != "ERROR" {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}

	msg, ok := cutLeadingSpace(rest[sp:])
	if !ok || msg == "" {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}
	if strings.IndexByte(msg, '\n') >= 0 {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}

	t, err := time.Parse(timeLayout, tsStr)
	if err != nil {
		return LogEntry{}, fmt.Errorf("invalid timestamp %q: %w", tsStr, err)
	}
	return LogEntry{Time: t, Level: level, Message: msg}, nil
}

// isLogSpace matches RE2's \s class: [\t\n\f\r ].
func isLogSpace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\f', '\r':
		return true
	}
	return false
}

// cutLeadingSpace strips a run of >=1 \s chars from the front.
func cutLeadingSpace(s string) (string, bool) {
	i := 0
	for i < len(s) && isLogSpace(s[i]) {
		i++
	}
	return s[i:], i > 0
}

// CountByLevel tallies entries per level.
func CountByLevel(entries []LogEntry) map[string]int {
	out := map[string]int{}
	for _, e := range entries {
		out[e.Level]++
	}
	return out
}
```

- [ ] **Step 4:** Append a `TestParseLine` to BOTH `logparse_test.go` files (the carried test already covers `Parse`/`CountByLevel`; add direct coverage for the newly-exported `ParseLine`):

```go
func TestParseLine(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		e, err := ParseLine("2026-05-21T14:30:00 INFO hello world")
		if err != nil {
			t.Fatalf("ParseLine: %v", err)
		}
		if e.Level != "INFO" || e.Message != "hello world" {
			t.Errorf("ParseLine = %+v", e)
		}
	})
	for _, bad := range []string{"", "0", "garbage", "2026-05-21T14:30:00 DEBUG x", "2026-13-02T15:04:05 INFO x"} {
		if _, err := ParseLine(bad); err == nil {
			t.Errorf("ParseLine(%q) = nil error, want error", bad)
		}
	}
}
```

- [ ] **Step 5:** Verify the carried baseline:

```bash
gofmt -l lessons/23-http-server/
go build ./lessons/23-http-server/...
go test ./lessons/23-http-server/... 2>&1 | tail -10
go vet ./lessons/23-http-server/...
golangci-lint run ./lessons/23-http-server/...
```

Expected: gofmt empty; build clean; logparse tests pass (exercises + solutions); vet + lint clean. (No `regexp`/`unused` complaints — the oracle is gone.)

- [ ] **Step 6:** Commit:

```bash
git add lessons/23-http-server/exercises lessons/23-http-server/solutions
git commit -m "chore(lesson-23): scaffold + carry/trim logparse (export ParseLine)"
```

---

## Task 2: Warm-up — `middleware`

Implement `WithRequestLog` — teaches the `func(http.Handler) http.Handler` wrapper + the ResponseWriter-wrapping trick used to capture the status code.

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/23-http-server/{exercises,solutions}/warmup/middleware`

- [ ] **Step 2:** `exercises/warmup/middleware/middleware.go` (SKELETON):

```go
// Package middleware is the lesson 23 warm-up: an HTTP logging
// middleware. A middleware is a function that wraps an http.Handler
// and returns a new http.Handler — the foundation of request logging,
// auth, recovery, and more.
package middleware

import (
	"log/slog"
	"net/http"
)

// WithRequestLog wraps next so that every request is logged with its
// method, path, status code, and duration. Returns a new handler.
//
// The trick: http.ResponseWriter doesn't expose the status code after
// the fact, so wrap it in a small type that records the code passed to
// WriteHeader (defaulting to 200, which net/http uses when a handler
// writes a body without calling WriteHeader).
//
// Hint:
//   1. return http.HandlerFunc(func(w, r) {
//   2.   start := time.Now()
//   3.   rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
//   4.   next.ServeHTTP(rec, r)
//   5.   logger.Info("request", "method", r.Method, "path", r.URL.Path,
//          "status", rec.status, "duration_ms", time.Since(start).Milliseconds())
//   6. })
//
// You'll need a statusRecorder type embedding http.ResponseWriter with
// an overridden WriteHeader that stores the code.
func WithRequestLog(next http.Handler, logger *slog.Logger) http.Handler {
	_ = slog.LevelInfo
	panic("TODO: wrap next; record status via a ResponseWriter wrapper; log method/path/status/duration")
}
```

- [ ] **Step 3:** `exercises/warmup/middleware/middleware_test.go` (SKELETON — must not call the panicking WithRequestLog):

```go
package middleware

import (
	"log/slog"
	"net/http"
	"testing"
)

// TestWithRequestLog is a SKELETON. Wrap a handler that writes a known
// status; assert the wrapped handler passes the status through AND that
// the logger recorded method/path/status.
func TestWithRequestLog(t *testing.T) {
	// TODO:
	//   var buf bytes.Buffer
	//   logger := slog.New(slog.NewJSONHandler(&buf, nil))
	//   h := WithRequestLog(http.HandlerFunc(func(w, r){ w.WriteHeader(418); w.Write([]byte("hi")) }), logger)
	//   rec := httptest.NewRecorder(); h.ServeHTTP(rec, httptest.NewRequest("GET", "/foo", nil))
	//   assert rec.Code == 418
	//   assert buf contains "method":"GET", "path":"/foo", "status":418
	_ = slog.New
	_ = http.HandlerFunc(nil)
}
```

- [ ] **Step 4:** `solutions/warmup/middleware/middleware.go`:

```go
// Package middleware is the lesson 23 warm-up reference implementation.
package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// statusRecorder wraps http.ResponseWriter to capture the status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// WithRequestLog logs method, path, status, and duration for each request.
func WithRequestLog(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}
```

- [ ] **Step 5:** `solutions/warmup/middleware/middleware_test.go`:

```go
package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithRequestLog(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	h := WithRequestLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("hi"))
	}), logger)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/foo", nil))

	if rec.Code != http.StatusTeapot {
		t.Errorf("status passthrough = %d, want 418", rec.Code)
	}
	out := buf.String()
	for _, want := range []string{`"method":"GET"`, `"path":"/foo"`, `"status":418`} {
		if !strings.Contains(out, want) {
			t.Errorf("log missing %s in: %s", want, out)
		}
	}
}

func TestWithRequestLogDefaultStatus(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	// Handler that writes a body without calling WriteHeader → 200.
	h := WithRequestLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}), logger)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if !strings.Contains(buf.String(), `"status":200`) {
		t.Errorf("default status not logged as 200: %s", buf.String())
	}
}
```

- [ ] **Step 6:** Verify + commit:

```bash
gofmt -l lessons/23-http-server/
go test ./lessons/23-http-server/exercises/warmup/middleware/... 2>&1 | tail -5
go test -v ./lessons/23-http-server/solutions/warmup/middleware/... 2>&1 | tail -15
go vet ./lessons/23-http-server/...
golangci-lint run ./lessons/23-http-server/...
make test

git add lessons/23-http-server/exercises/warmup lessons/23-http-server/solutions/warmup
git commit -m "feat(lesson-23): warmup — middleware (WithRequestLog + statusRecorder)"
```

Expected: exercises vacuous-pass; solutions tests PASS.

---

## Task 3: `internal/logstats` accumulator

The concurrent shared state: a mutex-guarded level-count store.

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/23-http-server/{exercises,solutions}/internal/logstats`

- [ ] **Step 2:** `exercises/internal/logstats/logstats.go` (SKELETON):

```go
// Package logstats is the in-memory level-count accumulator shared
// across all ingest requests of the logstats service. It is safe for
// concurrent use.
package logstats

import "sync"

// Store accumulates per-level log counts under a mutex.
type Store struct {
	mu     sync.Mutex
	counts map[string]int
}

// NewStore returns an empty, ready-to-use Store.
func NewStore() *Store {
	return &Store{counts: map[string]int{}}
}

// Merge adds the per-level deltas into the store.
//
// Hint: lock, then `for level, n := range delta { s.counts[level] += n }`.
func (s *Store) Merge(delta map[string]int) {
	_ = sync.Mutex{}
	panic("TODO: lock, add each delta into s.counts")
}

// Snapshot returns a COPY of the counts plus the grand total. Returning
// a copy (not the internal map) keeps callers from racing on it.
//
// Hint: lock, copy counts into a new map, sum the values.
func (s *Store) Snapshot() (map[string]int, int) {
	panic("TODO: lock, copy counts to a new map, return (copy, total)")
}
```

- [ ] **Step 3:** `exercises/internal/logstats/logstats_test.go` (SKELETON):

```go
package logstats

import "testing"

// TestStore is a SKELETON. Merge some deltas, Snapshot, assert counts +
// total. Then exercise concurrent Merge under `go test -race`.
func TestStore(t *testing.T) {
	cases := []struct {
		name   string
		deltas []map[string]int
		want   map[string]int
	}{
		// TODO: e.g. {"single", []map[string]int{{"INFO":2}}, map[string]int{"INFO":2}}
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewStore()
			for _, d := range tc.deltas {
				s.Merge(d)
			}
			got, _ := s.Snapshot()
			_ = got
			_ = tc.want
		})
	}
}
```

- [ ] **Step 4:** `solutions/internal/logstats/logstats.go`:

```go
// Package logstats is the lesson 23 reference accumulator.
package logstats

import "sync"

// Store accumulates per-level log counts under a mutex. Safe for
// concurrent use by multiple request handlers.
type Store struct {
	mu     sync.Mutex
	counts map[string]int
}

func NewStore() *Store {
	return &Store{counts: map[string]int{}}
}

// Merge adds the per-level deltas into the store.
func (s *Store) Merge(delta map[string]int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for level, n := range delta {
		s.counts[level] += n
	}
}

// Snapshot returns a copy of the counts plus the grand total.
func (s *Store) Snapshot() (map[string]int, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]int, len(s.counts))
	total := 0
	for level, n := range s.counts {
		out[level] = n
		total += n
	}
	return out, total
}
```

- [ ] **Step 5:** `solutions/internal/logstats/logstats_test.go`:

```go
package logstats

import (
	"sync"
	"testing"
)

func TestStore(t *testing.T) {
	s := NewStore()
	s.Merge(map[string]int{"INFO": 2, "WARN": 1})
	s.Merge(map[string]int{"INFO": 3, "ERROR": 1})

	counts, total := s.Snapshot()
	if counts["INFO"] != 5 || counts["WARN"] != 1 || counts["ERROR"] != 1 {
		t.Errorf("counts = %v", counts)
	}
	if total != 7 {
		t.Errorf("total = %d, want 7", total)
	}
}

func TestSnapshotIsCopy(t *testing.T) {
	s := NewStore()
	s.Merge(map[string]int{"INFO": 1})
	snap, _ := s.Snapshot()
	snap["INFO"] = 999 // mutating the copy must not affect the store
	again, _ := s.Snapshot()
	if again["INFO"] != 1 {
		t.Errorf("Snapshot returned the internal map, not a copy: %v", again)
	}
}

func TestConcurrentMerge(t *testing.T) {
	s := NewStore()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Merge(map[string]int{"INFO": 1})
		}()
	}
	wg.Wait()
	counts, _ := s.Snapshot()
	if counts["INFO"] != 100 {
		t.Errorf("INFO = %d, want 100", counts["INFO"])
	}
}
```

- [ ] **Step 6:** Verify + commit:

```bash
gofmt -l lessons/23-http-server/
go test ./lessons/23-http-server/exercises/internal/logstats/... 2>&1 | tail -5
go test -v ./lessons/23-http-server/solutions/internal/logstats/... 2>&1 | tail -15
go test -race ./lessons/23-http-server/solutions/internal/logstats/... 2>&1 | tail -5
make test-race
golangci-lint run ./lessons/23-http-server/...

git add lessons/23-http-server/exercises/internal/logstats lessons/23-http-server/solutions/internal/logstats
git commit -m "feat(lesson-23): logstats accumulator (mutex-guarded level counts)"
```

Expected: exercises vacuous-pass; solutions tests PASS; -race clean.

---

## Task 4: `cmd/logstats-server`

The HTTP service: router (ServeMux 1.22), the three handlers, middleware, `slog`, `run`/`serve` lifecycle.

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/23-http-server/{exercises,solutions}/cmd/logstats-server`

- [ ] **Step 2:** `exercises/cmd/logstats-server/main.go` (SKELETON — `newRouter` panics; `main`/`run`/`serve`/middleware/`writeJSON` are provided working so the package compiles and students fill in the handlers + router):

```go
// Package main is the lesson 23 logstats HTTP service.
//
// Endpoints:
//   POST /ingest   {"lines":["<log line>", ...]}  → parse + accumulate
//   GET  /stats    → {"counts":{...},"total":N}
//   GET  /healthz  → {"status":"ok"}
//
// You implement newRouter + the three handlers. The server lifecycle
// (run/serve), middleware, and writeJSON are provided.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ristkari-dev/go-training/lessons/23-http-server/exercises/internal/logparse"
	"github.com/ristkari-dev/go-training/lessons/23-http-server/exercises/internal/logstats"
)

const maxIngestBytes = 1 << 20 // 1 MiB request-body cap

type ingestRequest struct {
	Lines []string `json:"lines"`
}

type ingestResponse struct {
	Accepted int `json:"accepted"`
	Parsed   int `json:"parsed"`
	Failed   int `json:"failed"`
}

type statsResponse struct {
	Counts map[string]int `json:"counts"`
	Total  int            `json:"total"`
}

// newRouter builds the mux + middleware. IMPLEMENT THIS.
//
// Hint:
//   mux := http.NewServeMux()
//   mux.HandleFunc("POST /ingest", ingestHandler(store))
//   mux.HandleFunc("GET /stats", statsHandler(store))
//   mux.HandleFunc("GET /healthz", healthHandler)
//   return withRequestLog(withRecovery(mux, logger), logger)
func newRouter(store *logstats.Store, logger *slog.Logger) http.Handler {
	_ = store
	_ = logger
	_ = logparse.ParseLine
	panic("TODO: build mux with POST /ingest, GET /stats, GET /healthz; wrap in middleware")
}

// ingestHandler decodes {"lines":[...]}, parses each line with
// logparse.ParseLine (lenient: count failures, don't reject), merges
// valid levels into the store, and returns the accepted/parsed/failed
// breakdown. Bad JSON or an oversized body → 400. IMPLEMENT THIS.
func ingestHandler(store *logstats.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = maxIngestBytes
		panic("TODO: decode (MaxBytesReader), ParseLine each, store.Merge, writeJSON breakdown")
	}
}

// statsHandler returns the current counts + total. IMPLEMENT THIS.
func statsHandler(store *logstats.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		panic("TODO: store.Snapshot(); writeJSON statsResponse")
	}
}

// healthHandler reports liveness. IMPLEMENT THIS.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	panic("TODO: writeJSON 200 {\"status\":\"ok\"}")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func withRequestLog(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		logger.Info("request", "method", r.Method, "path", r.URL.Path,
			"status", rec.status, "duration_ms", time.Since(start).Milliseconds())
	})
}

func withRecovery(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				logger.Error("panic recovered", "value", v, "path", r.URL.Path)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func run(ctx context.Context, addr string, stdout io.Writer) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return serve(ctx, ln, stdout)
}

func serve(ctx context.Context, ln net.Listener, stdout io.Writer) error {
	logger := slog.New(slog.NewJSONHandler(stdout, nil))
	store := logstats.NewStore()
	srv := &http.Server{Handler: newRouter(store, logger)}

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()

	logger.Info("listening", "addr", ln.Addr().String())
	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	addr := ":8080"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	if err := run(ctx, addr, os.Stdout); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 3:** `exercises/cmd/logstats-server/main_test.go` (SKELETON — must not call the panicking router):

```go
package main

import (
	"net/http/httptest"
	"testing"
)

// TestIngestThenStats is a SKELETON. Build the router, POST /ingest with
// some lines, GET /stats, assert the counts.
func TestIngestThenStats(t *testing.T) {
	// TODO:
	//   srv := httptest.NewServer(newRouter(logstats.NewStore(), discardLogger))
	//   defer srv.Close()
	//   POST {"lines":[...]} to srv.URL+"/ingest"; assert accepted/parsed/failed
	//   GET srv.URL+"/stats"; assert counts + total
	_ = httptest.NewServer
}
```

- [ ] **Step 4:** `solutions/cmd/logstats-server/main.go` — same as the exercises file but with `newRouter` + the three handlers IMPLEMENTED, and the import path `.../solutions/...`:

```go
// Package main is the lesson 23 logstats HTTP service (reference impl).
//
// Endpoints:
//   POST /ingest   {"lines":["<log line>", ...]}  → parse + accumulate
//   GET  /stats    → {"counts":{...},"total":N}
//   GET  /healthz  → {"status":"ok"}
package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ristkari-dev/go-training/lessons/23-http-server/solutions/internal/logparse"
	"github.com/ristkari-dev/go-training/lessons/23-http-server/solutions/internal/logstats"
)

const maxIngestBytes = 1 << 20

type ingestRequest struct {
	Lines []string `json:"lines"`
}

type ingestResponse struct {
	Accepted int `json:"accepted"`
	Parsed   int `json:"parsed"`
	Failed   int `json:"failed"`
}

type statsResponse struct {
	Counts map[string]int `json:"counts"`
	Total  int            `json:"total"`
}

func newRouter(store *logstats.Store, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /ingest", ingestHandler(store))
	mux.HandleFunc("GET /stats", statsHandler(store))
	mux.HandleFunc("GET /healthz", healthHandler)
	return withRequestLog(withRecovery(mux, logger), logger)
}

func ingestHandler(store *logstats.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ingestRequest
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxIngestBytes))
		if err := dec.Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		delta := map[string]int{}
		parsed, failed := 0, 0
		for _, line := range req.Lines {
			e, err := logparse.ParseLine(line)
			if err != nil {
				failed++
				continue
			}
			delta[e.Level]++
			parsed++
		}
		store.Merge(delta)
		writeJSON(w, http.StatusOK, ingestResponse{
			Accepted: len(req.Lines), Parsed: parsed, Failed: failed,
		})
	}
}

func statsHandler(store *logstats.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		counts, total := store.Snapshot()
		writeJSON(w, http.StatusOK, statsResponse{Counts: counts, Total: total})
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func withRequestLog(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		logger.Info("request", "method", r.Method, "path", r.URL.Path,
			"status", rec.status, "duration_ms", time.Since(start).Milliseconds())
	})
}

func withRecovery(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				logger.Error("panic recovered", "value", v, "path", r.URL.Path)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func run(ctx context.Context, addr string, stdout io.Writer) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return serve(ctx, ln, stdout)
}

func serve(ctx context.Context, ln net.Listener, stdout io.Writer) error {
	logger := slog.New(slog.NewJSONHandler(stdout, nil))
	store := logstats.NewStore()
	srv := &http.Server{Handler: newRouter(store, logger)}

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()

	logger.Info("listening", "addr", ln.Addr().String())
	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	addr := ":8080"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	if err := run(ctx, addr, os.Stdout); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 5:** `solutions/cmd/logstats-server/main_test.go`:

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ristkari-dev/go-training/lessons/23-http-server/solutions/internal/logstats"
)

func testRouter() http.Handler {
	return newRouter(logstats.NewStore(), slog.New(slog.NewJSONHandler(io.Discard, nil)))
}

func TestIngestThenStats(t *testing.T) {
	srv := httptest.NewServer(testRouter())
	defer srv.Close()

	body := `{"lines":[
		"2026-01-02T15:04:05 INFO ok",
		"2026-01-02T15:04:06 INFO again",
		"2026-01-02T15:04:07 WARN slow",
		"garbage line"
	]}`
	resp, err := http.Post(srv.URL+"/ingest", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("ingest status = %d", resp.StatusCode)
	}
	var ir ingestResponse
	_ = json.NewDecoder(resp.Body).Decode(&ir)
	resp.Body.Close()
	if ir.Accepted != 4 || ir.Parsed != 3 || ir.Failed != 1 {
		t.Errorf("ingest resp = %+v, want accepted=4 parsed=3 failed=1", ir)
	}

	resp2, err := http.Get(srv.URL + "/stats")
	if err != nil {
		t.Fatalf("get stats: %v", err)
	}
	var sr statsResponse
	_ = json.NewDecoder(resp2.Body).Decode(&sr)
	resp2.Body.Close()
	if sr.Total != 3 || sr.Counts["INFO"] != 2 || sr.Counts["WARN"] != 1 {
		t.Errorf("stats = %+v, want total=3 INFO=2 WARN=1", sr)
	}
}

func TestIngestBadJSON(t *testing.T) {
	srv := httptest.NewServer(testRouter())
	defer srv.Close()
	resp, err := http.Post(srv.URL+"/ingest", "application/json", strings.NewReader("{not json"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestHealthz(t *testing.T) {
	rec := httptest.NewRecorder()
	testRouter().ServeHTTP(rec, httptest.NewRequest("GET", "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("healthz = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type = %q", ct)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	rec := httptest.NewRecorder()
	testRouter().ServeHTTP(rec, httptest.NewRequest("GET", "/ingest", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /ingest = %d, want 405", rec.Code)
	}
}

func TestConcurrentIngestRace(t *testing.T) {
	srv := httptest.NewServer(testRouter())
	defer srv.Close()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Post(srv.URL+"/ingest", "application/json",
				strings.NewReader(`{"lines":["2026-01-02T15:04:05 ERROR e"]}`))
			if err == nil {
				resp.Body.Close()
			}
		}()
	}
	wg.Wait()
	resp, _ := http.Get(srv.URL + "/stats")
	var sr statsResponse
	_ = json.NewDecoder(resp.Body).Decode(&sr)
	resp.Body.Close()
	if sr.Counts["ERROR"] != 50 {
		t.Errorf("ERROR = %d, want 50", sr.Counts["ERROR"])
	}
}

func TestServeLifecycle(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- serve(ctx, ln, io.Discard) }()

	url := fmt.Sprintf("http://%s/healthz", ln.Addr().String())
	ready := false
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if resp, err := http.Get(url); err == nil {
			resp.Body.Close()
			ready = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !ready {
		t.Fatal("server never became ready")
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("serve returned %v, want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("serve did not return after cancel")
	}
}
```

- [ ] **Step 6:** Verify + commit:

```bash
gofmt -l lessons/23-http-server/
go test ./lessons/23-http-server/exercises/cmd/logstats-server/... 2>&1 | tail -5
go test -v ./lessons/23-http-server/solutions/cmd/logstats-server/... 2>&1 | tail -25
go test -race ./lessons/23-http-server/solutions/cmd/logstats-server/... 2>&1 | tail -5
make test
make test-race
go vet ./lessons/23-http-server/...
golangci-lint run ./lessons/23-http-server/...

# Manual smoke
go build -o /tmp/logstats-server ./lessons/23-http-server/solutions/cmd/logstats-server
/tmp/logstats-server 127.0.0.1:8181 &
sleep 0.5
curl -s -XPOST localhost:8181/ingest -d '{"lines":["2026-01-02T15:04:05 INFO ok","bad"]}'; echo
curl -s localhost:8181/stats; echo
curl -s localhost:8181/healthz; echo
kill %1; rm -f /tmp/logstats-server

git add lessons/23-http-server/exercises/cmd lessons/23-http-server/solutions/cmd
git commit -m "feat(lesson-23): cmd/logstats-server (net/http + ServeMux 1.22 + slog + middleware)"
```

Expected: exercises vacuous-pass; solutions 6 tests PASS; -race clean; smoke shows `{"accepted":2,"parsed":1,"failed":1}` then `{"counts":{"INFO":1},"total":1}` then `{"status":"ok"}`.

---

## Task 5: Slide deck — 5 concepts

Heavy-explanatory pattern per concept (motivation → basics → worked example → common mistake → recap). ~30 slides.

**File:** `lessons/23-http-server/slides/slides.md`

Concept order:

1. **`net/http` server + ServeMux 1.22 routing** — `http.Server`, `ListenAndServe`/`Serve`, `http.NewServeMux`, method+pattern patterns (`POST /ingest`, `GET /stats`), `{wildcard}` paths, automatic 405. *Common mistake:* the pre-1.22 "match path prefix, switch on `r.Method` by hand" idiom (verbose, error-prone).
2. **Handlers + request/response** — `http.Handler`/`HandlerFunc`, reading the body, `json.Decoder`/`json.Encoder`, `http.MaxBytesReader`, status codes, `http.Error`. *Common mistake:* decoding unbounded request bodies (DoS); forgetting to set status/Content-Type.
3. **Middleware** — `func(http.Handler) http.Handler`, chaining, the `statusRecorder` ResponseWriter wrapper, panic recovery. *Common mistake:* not calling `next`; not capturing the status; recovering but still writing a 200.
4. **Structured logging with `slog`** — `slog.New(slog.NewJSONHandler(...))`, levels, `With`, attributes, request-scoped loggers. *Common mistake:* unstructured `fmt`-style logs; logging secrets/PII.
5. **Designing the service** — the ingest/stats/healthz surface, the mutex-guarded accumulator as shared state, `run`/`serve` testability + `httptest`, a light `Shutdown` preview. *Common mistake:* an unguarded shared map across concurrent requests (race — ties back to L18).

- [ ] **Step 1-7:** Author the deck following the L19/L21/L22 format (title-slide-grid header, "What we'll cover", "The story so far" — opening Phase 4, the service framing, 5 concept sections, Practice, Closing thought, What we learned, Up next → Lesson 24 HTTP clients).
- [ ] **Step 8:** Verify slides build + commit:

```bash
make slides-build
grep -q "23-http-server" dist/index.html && echo "✓ in index"
rm -rf dist

git add lessons/23-http-server/slides/
git commit -m "feat(lesson-23): slides — HTTP servers (5 concepts)"
```

---

## Task 6: README + verify + final review + PR

- [ ] **Step 1:** Confirm `tools/build-index/main.go` already lists lesson 23 as `{Number:"23", Slug:"http-server", Title:"HTTP servers", Blurb:"net/http · slog", Phase:4}` — the slug matches the directory, so **no build-index change is needed** (verify with `grep '23' tools/build-index/main.go`).

- [ ] **Step 2:** Write `lessons/23-http-server/README.md` (~280 lines). Mirror the 5 concepts with per-concept Common-mistake paragraphs. Include:
  - "What's different / Phase 4 begins": the aggregator becomes the `logstats` service; the running example is now a long-lived process serving requests.
  - API reference (the three endpoints with example `curl`s + JSON shapes).
  - "How to run": `go run ... cmd/logstats-server :8080`, `curl` examples, `make test-race`.
  - Note that ingest is lenient (malformed lines counted, not rejected) and stats are cumulative-since-start.
  - "Going further": add `GET /stats/{level}` using a path wildcard; add a `-addr` flag (preview of L26); rate-limit middleware; serve `/healthz` readiness vs liveness.

- [ ] **Step 3:** Full verification sweep:

```bash
make test
make test-race
go test -v ./lessons/23-http-server/solutions/...
go test -race ./lessons/23-http-server/...
go vet ./...
golangci-lint run ./...
gofmt -l lessons/23-http-server/
make slides-build && grep -q "23-http-server" dist/index.html && echo "✓ index" && rm -rf dist
```

- [ ] **Step 4:** Commit README:

```bash
git add lessons/23-http-server/README.md
git commit -m "docs(lesson-23): README — HTTP servers self-study"
```

- [ ] **Step 5:** Dispatch the `feature-dev:code-reviewer` over `git diff main...HEAD`. Focus: handler correctness + status codes, the `statusRecorder`/recovery middleware, race-safety of the accumulator under concurrent requests, `serve` lifecycle (no goroutine leak, clean shutdown), slide/README/code consistency, and that the exercises skeletons pass vacuously without invoking panicking code. Apply fixes.

- [ ] **Step 6:** Push + open PR:

```bash
git push -u origin feature/plan-z-lesson-23-http-server
gh pr create --title "Lesson 23 — HTTP servers" --body "..."
```

---

## Verification (after Task 6)

```bash
make test
make test-race
go test ./lessons/23-http-server/exercises/...                      # vacuous-pass
go test -v ./lessons/23-http-server/solutions/...                   # warmup + logstats + logparse + server
go test -race ./lessons/23-http-server/...
go vet ./...
golangci-lint run ./...
gofmt -l lessons/23-http-server/
make slides-build && grep -q "23-http-server" dist/index.html && rm -rf dist

# Manual end-to-end
go build -o /tmp/logstats-server ./lessons/23-http-server/solutions/cmd/logstats-server
/tmp/logstats-server 127.0.0.1:8181 &
sleep 0.5
curl -s -XPOST localhost:8181/ingest -d '{"lines":["2026-01-02T15:04:05 INFO ok","2026-01-02T15:04:06 ERROR boom","bad"]}'; echo
curl -s localhost:8181/stats; echo
curl -s localhost:8181/healthz; echo
kill %1; rm -f /tmp/logstats-server
```

## Critical file paths

To be created:
- `lessons/23-http-server/` (directory)
- `lessons/23-http-server/README.md`
- `lessons/23-http-server/slides/{index.html (scaffolded), slides.md, assets/.gitkeep}`
- `lessons/23-http-server/exercises/warmup/middleware/{middleware.go, middleware_test.go}`
- `lessons/23-http-server/exercises/internal/logparse/{logparse.go, logparse_test.go}` (carried + trimmed)
- `lessons/23-http-server/exercises/internal/logstats/{logstats.go, logstats_test.go}`
- `lessons/23-http-server/exercises/cmd/logstats-server/{main.go, main_test.go}`
- `lessons/23-http-server/solutions/...` (mirrored)

To be referenced (not modified):
- `lessons/22-profiling-fuzz/solutions/internal/logparse/` (carry-forward source)
- `tools/build-index/main.go` (slug `http-server` already correct — no change)
- `.golangci.yml` (no changes; no third-party deps in L23)

## Execution after approval

1. (Branch `feature/plan-z-lesson-23-http-server` already created off main.)
2. Commit this plan doc.
3. Execute the 6 tasks via subagent-driven development (controller carries/trims logparse + verifies Task 1 directly; dispatches subagents for Tasks 2-4; writes slides/README inline for Tasks 5-6; runs verification directly).
4. Final code review subagent.
5. Push and open PR.
