# Plan EE — Lesson 28 (Observability) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax.
>
> **Commit policy:** NO `Co-Authored-By` trailer or "Generated with" line in commit messages or PR bodies. Subject + body only.

**Goal:** Author lesson 28 — the sixth Phase 4 lesson, and the course's **second/final** third-party dependency (OpenTelemetry, after gRPC). Instrument the `logstatsd` daemon with traces, metrics, and trace-correlated logs across both faces (HTTP + gRPC), exported to **stdout** so it runs with no collector.

**Architecture:** Same per-lesson pattern, six tasks. Carries the whole L27 daemon verbatim. New: `warmup/tracemw` (a tracing HTTP middleware — the warmup exercise), `internal/otelx` (`Setup` providers+exporters — the main exercise; plus a provided `Instrument` middleware + meters), and provided wiring into `httpsrv`, `grpcsrv`, and `cmd/logstatsd`. Five-concept slide deck.

**Tech Stack:** Go 1.23 stdlib + carried grpc/protobuf + **OpenTelemetry v1.38.0** (`go.opentelemetry.io/otel`, `otel/sdk`, `otel/sdk/metric`, `otel/exporters/stdout/stdouttrace`, `otel/exporters/stdout/stdoutmetric`; tests also use `otel/sdk/trace/tracetest` + `otel/sdk/metric/metricdata`).

---

## Scope

After Plan EE: lesson 28 complete; `make test` + `make test-race` green; `logstatsd` emits spans + metrics to stdout and trace-correlated logs; both faces instrumented; `28-observability` in the index; **go directive stays `go 1.23`**.

### Design decisions (2 user-approved + plan-recommended)

**User-approved via brainstorming:**

1. **Instrument both faces with core OTel only** — an HTTP otel middleware (span + request counter + latency histogram + trace/log correlation); the carried gRPC logging interceptors gain a span (now logging+tracing); an `ingested_lines_total` counter in both ingest paths. **No `otelgrpc` contrib dep** — manual spans via `tracer.Start`.
2. **Five slide concepts:** three pillars + OTel model · SDK setup + stdout exporters · tracing (spans + propagation) · metrics (counters + histograms) · trace/log correlation + health vs readiness.

**Plan-recommended:**

3. **Pin OTel v1.38.0** — the newest OTel that declares `go 1.23` (later needs go 1.24+), keeping the course module at `go 1.23`. Verified: resolves with the stdout exporters, builds, tests pass, directive unchanged.
4. **`warmup/tracemw` = the warmup exercise** (a focused tracing middleware); **`otelx.Setup` = the main exercise** (build the providers + stdout exporters). The richer `Instrument` middleware, the meters, and all wiring are provided study code.
5. **Carry the whole L27 daemon** (keeps grpc deps live; OTel adds alongside).

### Verified facts (prototyped before writing this plan)

- **OTel v1.38.0** resolves (`otel` + `sdk` + `sdk/metric` + `stdouttrace` + `stdoutmetric`) and a clean tidy keeps the module at **`go 1.23.0`**.
- `otelx.Setup` (TracerProvider via stdouttrace batcher + MeterProvider via stdoutmetric periodic reader + global set + flushing shutdown) builds and works.
- The HTTP tracing middleware (span named `METHOD path` + `Int64Counter` + `Float64Histogram` + trace-ID-into-slog correlation) works; verified via `tracetest.NewInMemoryExporter` (span name + correlated trace ID in the log) and a `ManualReader` + `metricdata.Sum[int64]` (counter value).
- A **gRPC interceptor span** (`tracer.Start(ctx, info.FullMethod)`) is recorded via bufconn + in-memory exporter — core OTel, no contrib.
- gofmt orders OTel imports as: `otel`, `sdkmetric "…/sdk/metric"`, `…/sdk/metric/metricdata`, `sdktrace "…/sdk/trace"`, `…/sdk/trace/tracetest`.

---

## File structure

```
lessons/28-observability/{exercises,solutions}/
├── warmup/tracemw/{tracemw.go, tracemw_test.go}     (Task 2 — WithTracing SKELETON)
├── warmup/{config,buildinfo}/                        (Task 1 — carried)
├── proto/ + logstatspb/                              (Task 1 — carried, regenerated)
├── internal/
│   ├── logparse, logstats                            (Task 1 — carried)
│   ├── otelx/{otelx.go, otelx_test.go}               (Task 3 — Setup SKELETON; Instrument provided)
│   ├── httpsrv/                                       (Task 4 — wired to otelx.Instrument + ingest counter)
│   └── grpcsrv/                                       (Task 4 — interceptor spans + ingest counter)
└── cmd/logstatsd/{main.go, main_test.go, version_test.go, otel_test.go}   (Task 4 — otelx.Setup wired; gRPC-span test)
```

---

## Task 1: Scaffold + carry forward + OTel deps + regenerate proto (controller-direct)

- [ ] **Step 1:** Scaffold + remove flat stubs (`make new-lesson NAME=28-observability`; rm the 8 flat files).

- [ ] **Step 2:** Carry the whole daemon (both trees), rewrite paths + header:
```bash
SRC=lessons/27-container; DST=lessons/28-observability
for side in exercises solutions; do
  mkdir -p "$DST/$side/internal" "$DST/$side/proto" "$DST/$side/warmup" "$DST/$side/cmd"
  cp -R "$SRC/$side/internal/." "$DST/$side/internal/"
  cp -R "$SRC/$side/warmup/config" "$DST/$side/warmup/config"
  cp -R "$SRC/$side/warmup/buildinfo" "$DST/$side/warmup/buildinfo"
  cp -R "$SRC/$side/cmd/logstatsd" "$DST/$side/cmd/logstatsd"
  cp "$SRC/$side/proto/logstats.proto" "$DST/$side/proto/logstats.proto"
done
grep -rl '27-container' "$DST" | while read -r f; do sed -i '' 's#lessons/27-container#lessons/28-observability#g' "$f"; done
grep -rl 'lesson 27' "$DST" | while read -r f; do sed -i '' 's/lesson 27/lesson 28/g' "$f"; done
```

- [ ] **Step 3:** Add the L28 protos to the `Makefile` `proto` target (append two lines), then regenerate + add OTel deps:
```bash
make proto
go get go.opentelemetry.io/otel@v1.38.0 go.opentelemetry.io/otel/sdk@v1.38.0 \
       go.opentelemetry.io/otel/sdk/metric@v1.38.0 \
       go.opentelemetry.io/otel/exporters/stdout/stdouttrace@v1.38.0 \
       go.opentelemetry.io/otel/exporters/stdout/stdoutmetric@v1.38.0
go mod tidy
grep '^go ' go.mod   # MUST stay go 1.23.0 — if it bumped, the OTel pin is wrong
```

> Note: OTel deps won't be retained by `go mod tidy` until something imports them (the otelx package, Task 3). That's fine — `go get` records them; if tidy drops them before Task 3, re-run `go get` after Task 3, OR do Task 1's `go get` and accept tidy keeps them only once otelx exists. Cleanest: run the `go get` here (records in go.mod), and do NOT `go mod tidy` until otelx imports them in Task 3. So in Step 3, run `go get` + `make proto` but DEFER `go mod tidy` to Task 3 Step 6.

- [ ] **Step 4:** Verify carried baseline builds (`go build ./lessons/28-observability/...`; `go test`; vet/lint/fmt).

- [ ] **Step 5:** Commit:
```bash
git add lessons/28-observability Makefile go.mod go.sum
git commit -m "chore(lesson-28): scaffold + carry forward L27 daemon + add OpenTelemetry v1.38.0 deps"
```

---

## Task 2: Warm-up — `tracemw`

A tracing HTTP middleware: span per request + status captured on the span.

**Files (4 total):**

- [ ] **Step 1:** `exercises/warmup/tracemw/tracemw.go` (SKELETON):

```go
// Package tracemw is the lesson 28 warm-up: an HTTP middleware that
// records a tracing span for each request. It's the tracing half of the
// production otelx.Instrument middleware.
package tracemw

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// statusRecorder captures the response status code (the L23 trick).
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// WithTracing wraps next so each request runs inside a span named
// "<METHOD> <path>", with the response status recorded as a span
// attribute. The span's context is threaded into the request so handlers
// can create child spans. IMPLEMENT THIS.
//
// Hint:
//   return http.HandlerFunc(func(w, r) {
//       ctx, span := tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
//       defer span.End()
//       rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
//       next.ServeHTTP(rec, r.WithContext(ctx))
//       span.SetAttributes(attribute.Int("http.status_code", rec.status))
//   })
func WithTracing(next http.Handler, tracer trace.Tracer) http.Handler {
	_ = attribute.Int
	_ = statusRecorder{}
	panic("TODO: start a span named METHOD+path, thread ctx, record status attr on End")
}
```

- [ ] **Step 2:** `exercises/warmup/tracemw/tracemw_test.go` (SKELETON):

```go
package tracemw

import "testing"

// TestWithTracing is a SKELETON. Set an in-memory tracer provider, wrap
// a handler, serve a request, and assert one span named "GET /x" with
// the status attribute. See the solution for the tracetest pattern.
func TestWithTracing(t *testing.T) {
	// TODO
}
```

- [ ] **Step 3:** `solutions/warmup/tracemw/tracemw.go` (same but `WithTracing` implemented):

```go
// Package tracemw is the lesson 28 warm-up reference implementation.
package tracemw

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// WithTracing wraps next so each request runs inside a span named
// "<METHOD> <path>", recording the response status as a span attribute.
func WithTracing(next http.Handler, tracer trace.Tracer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
		defer span.End()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r.WithContext(ctx))
		span.SetAttributes(attribute.Int("http.status_code", rec.status))
	})
}
```

- [ ] **Step 4:** `solutions/warmup/tracemw/tracemw_test.go`:

```go
package tracemw

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestWithTracing(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp)))
	tracer := otel.Tracer("test")

	h := WithTracing(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}), tracer)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/x", nil))

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if spans[0].Name != "GET /x" {
		t.Errorf("span name = %q, want %q", spans[0].Name, "GET /x")
	}
	var found bool
	for _, a := range spans[0].Attributes {
		if string(a.Key) == "http.status_code" && a.Value.AsInt64() == http.StatusTeapot {
			found = true
		}
	}
	if !found {
		t.Errorf("span missing http.status_code=418 attribute: %+v", spans[0].Attributes)
	}
}
```

- [ ] **Step 5:** Verify + commit:
```bash
gofmt -l lessons/28-observability/
go test ./lessons/28-observability/exercises/warmup/tracemw/... 2>&1 | tail -5
go test -v ./lessons/28-observability/solutions/warmup/tracemw/... 2>&1 | tail -10
go vet ./lessons/28-observability/... && golangci-lint run ./lessons/28-observability/...

git add lessons/28-observability/exercises/warmup/tracemw lessons/28-observability/solutions/warmup/tracemw
git commit -m "feat(lesson-28): warmup — tracemw (tracing HTTP middleware: span per request + status)"
```

Expected: exercises vacuous-pass; solutions TestWithTracing PASS.

---

## Task 3: `internal/otelx` — Setup (exercise) + Instrument + meters (provided)

**Files (4 total).** Per-tree import paths where logparse/logstats aren't needed (otelx is self-contained OTel).

- [ ] **Step 1:** `exercises/internal/otelx/otelx.go` (SKELETON — `Setup` panics; `Instrument` + `statusRecorder` provided working):

```go
// Package otelx wires OpenTelemetry for logstatsd: tracer + meter
// providers exporting to stdout (no collector needed), plus an HTTP
// middleware that records a span, request metrics, and a
// trace-correlated logger.
package otelx

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Setup builds global tracer + meter providers that export to w (stdout
// in production) and returns a shutdown func that flushes both. IMPLEMENT THIS.
//
// Hint:
//   traceExp, _ := stdouttrace.New(stdouttrace.WithWriter(w))
//   tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExp)); otel.SetTracerProvider(tp)
//   metricExp, _ := stdoutmetric.New(stdoutmetric.WithWriter(w))
//   mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp))); otel.SetMeterProvider(mp)
//   return func(ctx) error { return errors.Join(tp.Shutdown(ctx), mp.Shutdown(ctx)) }, nil
//
// (Imports you'll add: errors, the stdouttrace/stdoutmetric exporters,
// sdktrace "go.opentelemetry.io/otel/sdk/trace", sdkmetric "go.opentelemetry.io/otel/sdk/metric".)
func Setup(ctx context.Context, w io.Writer) (func(context.Context) error, error) {
	_ = w
	panic("TODO: build TracerProvider (stdouttrace) + MeterProvider (stdoutmetric), set global, return flushing shutdown")
}

// statusRecorder captures the response status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// Instrument wraps next with a per-request span, request counter +
// latency histogram, and a trace-correlated request logger. Provided.
func Instrument(next http.Handler, logger *slog.Logger) http.Handler {
	tracer := otel.Tracer("logstatsd/http")
	meter := otel.Meter("logstatsd/http")
	reqs, _ := meter.Int64Counter("http_requests_total")
	dur, _ := meter.Float64Histogram("http_request_duration_seconds")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx, span := tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
		defer span.End()

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		// Trace/log correlation: stamp the trace ID on a request logger.
		reqLogger := logger.With("trace_id", span.SpanContext().TraceID().String(),
			"method", r.Method, "path", r.URL.Path)

		next.ServeHTTP(rec, r.WithContext(ctx))

		span.SetAttributes(attribute.Int("http.status_code", rec.status))
		// status is low-cardinality — safe as a metric attribute (the raw
		// path is NOT, so we don't use it here).
		reqs.Add(ctx, 1, metric.WithAttributes(attribute.Int("status", rec.status)))
		dur.Record(ctx, time.Since(start).Seconds())
		reqLogger.Info("request handled", "status", rec.status,
			"duration_ms", time.Since(start).Milliseconds())
	})
}
```

- [ ] **Step 2:** `exercises/internal/otelx/otelx_test.go` (SKELETON):

```go
package otelx

import "testing"

// TestOtelx is a SKELETON. Cover Setup (build providers, shutdown flushes
// without error) and Instrument (a request records a span + increments
// the request counter + correlates the trace ID into the log). See the
// solution for the tracetest + ManualReader patterns.
func TestOtelx(t *testing.T) {
	// TODO
}
```

- [ ] **Step 3:** `solutions/internal/otelx/otelx.go` — same Instrument/statusRecorder; `Setup` implemented:

```go
// Package otelx wires OpenTelemetry for logstatsd. Reference impl.
package otelx

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Setup builds global tracer + meter providers exporting to w and
// returns a shutdown func that flushes both.
func Setup(ctx context.Context, w io.Writer) (func(context.Context) error, error) {
	traceExp, err := stdouttrace.New(stdouttrace.WithWriter(w))
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExp))
	otel.SetTracerProvider(tp)

	metricExp, err := stdoutmetric.New(stdoutmetric.WithWriter(w))
	if err != nil {
		return nil, err
	}
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp)))
	otel.SetMeterProvider(mp)

	return func(ctx context.Context) error {
		return errors.Join(tp.Shutdown(ctx), mp.Shutdown(ctx))
	}, nil
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// Instrument wraps next with a per-request span, request counter +
// latency histogram, and a trace-correlated request logger.
func Instrument(next http.Handler, logger *slog.Logger) http.Handler {
	tracer := otel.Tracer("logstatsd/http")
	meter := otel.Meter("logstatsd/http")
	reqs, _ := meter.Int64Counter("http_requests_total")
	dur, _ := meter.Float64Histogram("http_request_duration_seconds")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx, span := tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
		defer span.End()

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		reqLogger := logger.With("trace_id", span.SpanContext().TraceID().String(),
			"method", r.Method, "path", r.URL.Path)

		next.ServeHTTP(rec, r.WithContext(ctx))

		span.SetAttributes(attribute.Int("http.status_code", rec.status))
		reqs.Add(ctx, 1, metric.WithAttributes(attribute.Int("status", rec.status)))
		dur.Record(ctx, time.Since(start).Seconds())
		reqLogger.Info("request handled", "status", rec.status,
			"duration_ms", time.Since(start).Milliseconds())
	})
}
```

- [ ] **Step 4:** `solutions/internal/otelx/otelx_test.go`:

```go
package otelx

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestSetupShutdown(t *testing.T) {
	shutdown, err := Setup(context.Background(), io.Discard)
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Errorf("shutdown: %v", err)
	}
}

func TestInstrument(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp)))
	reader := sdkmetric.NewManualReader()
	otel.SetMeterProvider(sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)))

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	h := Instrument(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), logger)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/stats", nil))

	// Span recorded with the right name.
	spans := exp.GetSpans()
	if len(spans) != 1 || spans[0].Name != "GET /stats" {
		t.Fatalf("spans = %+v", spans)
	}
	// Trace/log correlation: the span's trace ID appears in the log.
	if tid := spans[0].SpanContext.TraceID().String(); !strings.Contains(buf.String(), tid) {
		t.Errorf("log missing correlated trace_id %s: %s", tid, buf.String())
	}
	// Request counter incremented.
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatal(err)
	}
	var total int64
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "http_requests_total" {
				if d, ok := m.Data.(metricdata.Sum[int64]); ok {
					for _, dp := range d.DataPoints {
						total += dp.Value
					}
				}
			}
		}
	}
	if total != 1 {
		t.Errorf("http_requests_total = %d, want 1", total)
	}
}
```

- [ ] **Step 5:** Run `go mod tidy` now (otelx imports the OTel deps, so tidy retains them):
```bash
go mod tidy && grep '^go ' go.mod   # still go 1.23.0
```

- [ ] **Step 6:** Verify + commit:
```bash
gofmt -l lessons/28-observability/
go test ./lessons/28-observability/exercises/internal/otelx/... 2>&1 | tail -5
go test -v ./lessons/28-observability/solutions/internal/otelx/... 2>&1 | tail -15
go test -race ./lessons/28-observability/solutions/internal/otelx/... 2>&1 | tail -3
make test && go vet ./lessons/28-observability/... && golangci-lint run ./lessons/28-observability/...

git add lessons/28-observability/exercises/internal/otelx lessons/28-observability/solutions/internal/otelx go.mod go.sum
git commit -m "feat(lesson-28): internal/otelx (Setup providers+stdout exporters; Instrument middleware + metrics)"
```

Expected: exercises vacuous-pass; solutions TestSetupShutdown + TestInstrument PASS; -race clean.

---

## Task 4: Wire both faces (provided) + gRPC-span test

Wire `otelx` into the daemon: HTTP via `otelx.Instrument`, gRPC interceptor spans, an ingest counter in both paths, and `otelx.Setup` at startup. All PROVIDED (study code); the exercises tree's otelx.Setup panics, so the exercises otel test is a skeleton.

- [ ] **Step 1:** `internal/httpsrv` (both trees): in `Router`, wrap the mux with `otelx.Instrument(..., logger)` (replacing or composing with the existing `withRequestLog` — keep recovery; otelx.Instrument supersedes the request logging). Add an `ingested_lines_total` counter: create it once (e.g. `otel.Meter("logstatsd").Int64Counter("ingested_lines_total")` in `Router`, closed over by `ingestHandler`) and `Add(ctx, int64(parsed+failed))` after parsing. Import `otelx` + `go.opentelemetry.io/otel`. (Per-tree import paths.)

- [ ] **Step 2:** `internal/grpcsrv` (both trees): in the carried `LoggingUnaryInterceptor` + `LoggingStreamInterceptor`, start a span around `handler` (`tracer := otel.Tracer("logstatsd/grpc")`; `ctx, span := tracer.Start(ctx, info.FullMethod)`; for the stream interceptor wrap with a context — start a span on `ss.Context()` and log; for streams the span just brackets the handler). Add an `ingested_lines_total` `Add` in `Ingest` after counting. Import `otel`.

- [ ] **Step 3:** `cmd/logstatsd/main.go` (both trees): in `run`, before building servers, call `shutdown, err := otelx.Setup(ctx, stdout)` (handle err), and `defer shutdown(context.Background())` (or call it in the drain). Pass `stdout` (already a param). Import `otelx`. (Provided.)

- [ ] **Step 4:** `solutions/cmd/logstatsd/otel_test.go` (real — gRPC span via bufconn; VERIFIED pattern):

```go
package main

import (
	"context"
	"net"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ristkari-dev/go-training/lessons/28-observability/solutions/internal/grpcsrv"
	"github.com/ristkari-dev/go-training/lessons/28-observability/solutions/internal/logstats"
	pb "github.com/ristkari-dev/go-training/lessons/28-observability/solutions/proto/logstatspb"
)

// TestGRPCSpan asserts the gRPC interceptor records a span per RPC.
func TestGRPCSpan(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp)))

	lis := bufconn.Listen(1 << 20)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(grpcsrv.LoggingUnaryInterceptor(logger)),
		grpc.StreamInterceptor(grpcsrv.LoggingStreamInterceptor(logger)),
	)
	pb.RegisterLogStatsServer(srv, grpcsrv.New(logstats.NewStore()))
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if _, err := pb.NewLogStatsClient(conn).GetStats(context.Background(), &pb.StatsRequest{}); err != nil {
		t.Fatalf("GetStats: %v", err)
	}

	spans := exp.GetSpans()
	if len(spans) == 0 {
		t.Fatal("no spans recorded for the gRPC call")
	}
	var found bool
	for _, s := range spans {
		if s.Name == "/logstats.v1.LogStats/GetStats" {
			found = true
		}
	}
	if !found {
		t.Errorf("no span named for GetStats; got %v", spanNames(spans))
	}
}
```

(Add the needed imports: `io`, `log/slog`, `google.golang.org/grpc/test/bufconn`, and a small `spanNames` helper. The IMPLEMENTER must reconcile imports + helper.) The exercises tree gets a skeleton `otel_test.go` (no calls).

- [ ] **Step 5:** Verify + commit:
```bash
gofmt -l lessons/28-observability/
go test ./lessons/28-observability/exercises/... 2>&1 | tail -8
go test -v ./lessons/28-observability/solutions/... 2>&1 | tail -25
go test -race ./lessons/28-observability/solutions/... 2>&1 | tail -5
make test && make test-race && go vet ./lessons/28-observability/... && golangci-lint run ./lessons/28-observability/...

# Manual: see spans + metrics on stdout
go build -o /tmp/logstatsd ./lessons/28-observability/solutions/cmd/logstatsd
/tmp/logstatsd -http-addr=127.0.0.1:8080 -grpc-addr=127.0.0.1:9090 &
sleep 0.5
curl -s -XPOST 127.0.0.1:8080/ingest -d '{"lines":["2026-01-02T15:04:05 INFO ok","bad"]}' >/dev/null
curl -s 127.0.0.1:8080/stats >/dev/null
kill -TERM %1; wait %1 2>/dev/null   # spans + metrics flush to stdout on shutdown
rm -f /tmp/logstatsd

git add lessons/28-observability
git commit -m "feat(lesson-28): instrument both faces (HTTP otel middleware + gRPC interceptor spans + ingest counter)"
```

Expected: exercises vacuous-pass; solutions TestGRPCSpan + TestServe + otelx tests PASS; -race clean; manual run prints span/metric JSON to stdout.

---

## Task 5: Slide deck — 5 concepts (controller inline)

**File:** `lessons/28-observability/slides/slides.md`. Concepts:

1. **Three pillars + the OTel model** — logs/metrics/traces as one correlated picture; OTel API vs SDK; providers → exporters; vendor-neutral. *Mistake:* siloed signals you can't pivot between.
2. **SDK setup + stdout exporters** — `TracerProvider`/`MeterProvider`, exporters, the `Shutdown` flush; stdout = no collector. *Mistake:* no `Shutdown` → batched data lost on exit.
3. **Tracing** — spans, `tracer.Start`, context propagation through handlers + RPCs; parent/child. *Mistake:* not threading the returned ctx → orphaned spans.
4. **Metrics** — counters (requests/ingest) vs histograms (latency); attributes + cardinality. *Mistake:* high-cardinality labels (user ID, raw path) → metric explosion.
5. **Trace/log correlation + health vs readiness** — trace ID into `slog` (pivot logs↔traces); `/healthz` liveness vs `/readyz` readiness. *Mistake:* uncorrelated logs; conflating liveness with readiness.

- [ ] Author + build + commit (`feat(lesson-28): slides — observability (5 concepts)`).

---

## Task 6: README + verify + final review + PR (controller)

- [ ] **Step 1:** Confirm build-index L28 (`Slug:"observability"`) — matches; no change.
- [ ] **Step 2:** Write `lessons/28-observability/README.md` (~300 lines): "What's different from L27" (the daemon gains telemetry across both faces); the OTel v1.38.0 pin + why (keeps go 1.23); the three signals + how to read the stdout output; trace/log correlation; "going further" (an OTLP exporter to a real collector; `otelgrpc`/`otelhttp` contrib; exemplars; resource attributes/service.name; sampling).
- [ ] **Step 3:** Full sweep (`make test` + `make test-race`; solutions `-v`; `-race`; vet/lint/fmt; `grep '^go ' go.mod` → 1.23.0; slides-build + index; the manual stdout smoke).
- [ ] **Step 4:** Commit README.
- [ ] **Step 5:** Dispatch `feature-dev:code-reviewer` over `git diff main...HEAD`. Focus: otelx Setup correctness (providers set global, shutdown flushes via errors.Join); Instrument (ctx threaded, status captured, low-cardinality metric attrs, correlation); gRPC interceptor spans (ctx propagation, stream interceptor correctness); ingest counter placement; cmd Setup + shutdown ordering (flush before exit); deps/go-directive unchanged (go 1.23, OTel v1.38.0); test determinism (in-memory/manual exporters, no stdout noise in tests); exercises skeletons vacuous; slide/README accuracy. Apply fixes.
- [ ] **Step 6:** Push + open PR (no co-author trailer); watch CI green (CI now downloads OTel — confirm resolves).

---

## Verification (after Task 6)

```bash
make test && make test-race
go test ./lessons/28-observability/exercises/...          # vacuous-pass
go test -v ./lessons/28-observability/solutions/...       # tracemw + otelx + gRPC span + carried daemon
go test -race ./lessons/28-observability/...
go vet ./... && golangci-lint run ./...
gofmt -l lessons/28-observability/ && grep '^go ' go.mod  # go 1.23.0
make slides-build && grep -q "28-observability" dist/index.html && rm -rf dist

# Manual: telemetry to stdout
go build -o /tmp/logstatsd ./lessons/28-observability/solutions/cmd/logstatsd
/tmp/logstatsd -http-addr=127.0.0.1:8080 -grpc-addr=127.0.0.1:9090 &
sleep 0.5
curl -s -XPOST 127.0.0.1:8080/ingest -d '{"lines":["2026-01-02T15:04:05 INFO ok","bad"]}'; echo
curl -s 127.0.0.1:8080/stats; echo
kill -TERM %1; wait %1 2>/dev/null     # span + metric JSON flushes to stdout
rm -f /tmp/logstatsd
```

## Critical file paths

To create: `lessons/28-observability/` tree (README, slides, warmup/tracemw, internal/otelx, otel_test.go, + carried daemon), mirrored.
To modify: `Makefile` (proto target += L28 protos), `go.mod`/`go.sum` (OTel v1.38.0 + indirects), and the carried `httpsrv`/`grpcsrv`/`cmd/logstatsd` (provided OTel wiring).
To reference: `lessons/27-container/...` (carry-forward source), `tools/build-index/main.go` (slug `observability` correct).

## Execution after approval

1. (Branch `feature/plan-ee-lesson-28-observability` already created off main.)
2. Commit this plan doc.
3. Execute: Task 1 controller-direct (carry + deps + regen — environment-sensitive); subagents for Tasks 2-4 (tracemw, otelx, wiring); controller inline for Tasks 5-6 (slides, README); controller runs all verification incl. the stdout smoke.
4. Final code review subagent. Push + PR (no co-author trailer); watch CI green.
