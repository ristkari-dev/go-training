<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">28</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 4 — Production &amp; Distributed</div>
<h1>Observability</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Make the daemon observable. Learn the three pillars (logs, metrics, traces), the OpenTelemetry SDK, spans + context propagation, counters + histograms, and trace/log correlation — instrumenting both the HTTP and gRPC faces, exported to stdout so it runs with no collector. The course's second and final third-party dependency, after gRPC.</p>
</div>
</div>
</div>

---

## What we'll cover

- **Three pillars + the OTel model** — logs, metrics, traces as one correlated picture.
- **SDK setup + stdout exporters** — providers, exporters, flush-on-shutdown; no collector needed.
- **Tracing** — spans + context propagation across handlers and RPCs.
- **Metrics** — counters and histograms; attributes and cardinality.
- **Trace/log correlation + health vs readiness** — pivot logs↔traces; liveness vs readiness.

---

## The story so far — you can't fix what you can't see

The daemon runs, is configured, ships in a container. But when it's slow or wrong in production, how do you know *why*? Logs alone tell you events; they don't tell you how long the ingest took, how many requests are erroring, or which request a log line belongs to. **L28 makes `logstatsd` observable** — it emits traces, metrics, and trace-correlated logs across both faces.

Observability is the difference between "the service is down" and "the p99 ingest latency tripled at 14:03 when the WARN rate spiked, here's the exact slow request's trace." It's not an afterthought you bolt on during an incident; it's instrumentation you build in, so the data is already there when the pager goes off.

We use **OpenTelemetry** — the vendor-neutral standard for all three signals. And we export to **stdout**, so the whole thing runs from a clean checkout with no Jaeger, no Prometheus, no collector. In production you'd swap the stdout exporter for an OTLP one pointing at your backend; the instrumentation code doesn't change.

---

## Concept 1: Three pillars + the OTel model

### Motivation

"Observability" is three kinds of telemetry, each answering a different question. Logs: *what happened* (discrete events). Metrics: *how much / how often* (aggregatable numbers). Traces: *where the time went* (the path of one request across functions/services). Their power is **together** — and OpenTelemetry is the one SDK that emits all three, correlated, to any backend.

---

### The basics

| Pillar | Answers | Example |
|---|---|---|
| **Logs** | what happened | `"ingest failed: bad line"` |
| **Metrics** | how much/often | `http_requests_total`, p99 latency |
| **Traces** | where time went | a span tree for one `/ingest` call |

OpenTelemetry splits into:

- **API** (`go.opentelemetry.io/otel`) — what your code calls (`Tracer`, `Meter`). Stable, dependency-light.
- **SDK** (`otel/sdk`, `otel/sdk/metric`) — the implementation: providers, samplers, exporters. Wired once at startup.
- **Exporters** — where the data goes: stdout, OTLP (→ Jaeger/Tempo/Prometheus/Datadog/…).

Your instrumentation calls the API; `main` wires the SDK. Swap exporters without touching handlers.

---

### A worked example

`logstatsd` instruments both faces against the same API, so traces/metrics/logs line up regardless of transport: an HTTP `/ingest` and a gRPC `Ingest` both produce a span, both bump `ingested_lines_total`, both log with a trace ID. One mental model, two doors.

---

### Common mistake

**Treating the three as separate tools** — logs in one system, metrics in another, traces in a third, none correlated. When an incident hits you're tab-hopping, unable to jump from a slow metric to the trace to the log. OTel's value is one SDK emitting all three with shared context, so you *can* pivot.

---

### Recap

- Three pillars: logs (events), metrics (aggregates), traces (request paths) — strongest together.
- OTel = API (your code) + SDK (wired in main) + exporters (where data goes).
- Instrument once against the API; choose the backend via the exporter.

---

## Concept 2: SDK setup + stdout exporters

### Motivation

Your handlers call `otel.Tracer(...)`/`otel.Meter(...)` — but those are no-ops until `main` installs real *providers* backed by *exporters*. Setup is a one-time wiring step, and it owns a critical detail: **flushing on shutdown**, or batched telemetry is lost when the process exits.

---

### The basics

```go
func Setup(ctx context.Context, w io.Writer) (func(context.Context) error, error) {
    traceExp, err := stdouttrace.New(stdouttrace.WithWriter(w))
    // ...
    tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExp))
    otel.SetTracerProvider(tp)

    metricExp, err := stdoutmetric.New(stdoutmetric.WithWriter(w))
    // ...
    mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp)))
    otel.SetMeterProvider(mp)

    return func(ctx context.Context) error {
        return errors.Join(tp.Shutdown(ctx), mp.Shutdown(ctx))  // flush both
    }, nil
}
```

- **Providers** (`TracerProvider`, `MeterProvider`) are set *global* via `otel.Set...`, so `otel.Tracer`/`otel.Meter` anywhere return real instruments.
- **Batcher / PeriodicReader** buffer and export in the background (efficient); they only emit on their interval — or when you flush.
- **`Shutdown` flushes.** The returned closure is `defer`-ed in the daemon's `run` so spans+metrics are written before exit.

Using stdout exporters means the telemetry is JSON on the daemon's stdout — no collector to stand up for the lesson.

---

### A worked example

`logstatsd`'s `run` calls `otelx.Setup(ctx, os.Stdout)` *before* building the servers (so their meters/tracers come from the now-installed global providers), and `defer`s the shutdown. On SIGTERM the daemon drains, `run` returns, and the deferred `shutdown` flushes — you see the span and metric JSON appear on stdout right before exit. (Verified: a `/ingest` + `/stats` then SIGTERM emits the `GET /stats`/`POST /ingest` spans and the `http_requests_total`/`ingested_lines_total` metrics.)

---

### Common mistake

**Forgetting `Shutdown`.** The batch span processor and periodic metric reader buffer data; if the process exits without flushing, the last batch — often the most interesting, right before a crash — is lost. Always wire a `defer shutdown(ctx)`. (Order matters too: set up OTel *before* creating instruments, or they bind to the no-op provider.)

---

### Recap

- `main` wires providers + exporters once and sets them global; handlers just call the API.
- Batcher/PeriodicReader export in the background; **`Shutdown` forces the final flush** — always `defer` it.
- stdout exporters = runnable with no collector; swap for OTLP in prod.

---

## Concept 3: Tracing

### Motivation

A trace is the story of one request: a tree of **spans**, each a timed operation, linked parent→child. It's how you see that a slow `/ingest` spent 90% of its time in parsing, or that a request fanned out to a gRPC call that timed out. The magic is **context propagation** — the active span travels in `context.Context`, so child operations attach automatically.

---

### The basics

```go
ctx, span := tracer.Start(r.Context(), "POST /ingest")
defer span.End()
// ... do work; pass ctx down so child spans attach to this one ...
next.ServeHTTP(w, r.WithContext(ctx))   // ctx carries the span
span.SetAttributes(attribute.Int("http.status_code", rec.status))
```

- **`tracer.Start(ctx, name)`** begins a span and returns a *new ctx* carrying it. `defer span.End()` closes it (records the duration).
- **Propagation:** anything that uses the returned `ctx` and starts its own span becomes a child — the tree builds itself.
- **Attributes** (`span.SetAttributes`) annotate a span (status, route). Across process boundaries, OTel propagates trace context over HTTP/gRPC headers so the trace spans services.

---

### A worked example

The HTTP middleware starts a span per request and threads the ctx into the handler; the gRPC interceptors start a span per RPC named by the full method (`/logstats.v1.LogStats/GetStats`) — core OTel, no contrib library. Both wrap the request the way L23's middleware did, but now producing a span. A handler that called another service with the threaded ctx would extend the same trace.

---

### Common mistake

**Not threading the returned ctx.** `_, span := tracer.Start(ctx, ...)` (discarding the new ctx) or passing the *original* ctx downstream means child spans attach to nothing — you get orphaned, disconnected spans instead of a tree. Always use the ctx that `Start` returns, and pass it down (`r.WithContext(ctx)`, `handler(ctx, ...)`).

---

### Recap

- A trace is a tree of timed spans; `tracer.Start` opens one, `span.End()` closes it.
- The returned **ctx carries the span** — thread it so children attach automatically.
- Attributes annotate spans; trace context propagates across services via headers.

---

## Concept 4: Metrics

### Motivation

Traces sample individual requests; metrics aggregate *all* of them into numbers you alert and dashboard on: request rate, error rate, latency percentiles, throughput. Two instruments cover most needs — **counters** (monotonic totals) and **histograms** (distributions).

---

### The basics

```go
meter := otel.Meter("logstatsd/http")
reqs, _ := meter.Int64Counter("http_requests_total")
dur, _  := meter.Float64Histogram("http_request_duration_seconds")

reqs.Add(ctx, 1, metric.WithAttributes(attribute.Int("status", status)))
dur.Record(ctx, time.Since(start).Seconds())
```

- **Counter** — only goes up (requests, lines ingested, errors). Backends compute rates from it.
- **Histogram** — records a distribution (latency, payload size); backends compute p50/p95/p99.
- **Attributes** split a metric by dimension (status, method) — but each distinct combination is a separate time series.

---

### A worked example

`logstatsd` records `http_requests_total{status}` + `http_request_duration_seconds` in the HTTP middleware, and `ingested_lines_total` where parsing happens — in **both** the HTTP handler and the gRPC `Ingest`, into one counter via the global meter. So "lines ingested" is correct regardless of which face the client used: the metric lives at the domain event, not the transport.

---

### Common mistake

**High-cardinality attributes.** Adding a user ID, request ID, or raw URL path as a metric attribute creates a new time series per distinct value — millions of them — which explodes memory and bankrupts your metrics backend. Keep metric attributes low-cardinality (status code, method, route *template*). High-cardinality detail belongs on **spans** (where each is individual), not metrics. (That's why our counter uses `status`, not the raw path — even though the *span name* uses the path.)

---

### Recap

- Counters (monotonic) for totals/rates; histograms for latency/size distributions.
- Record at the domain event (`ingested_lines_total` in both ingest paths) so it's transport-independent.
- Keep metric attributes **low-cardinality**; put high-cardinality detail on spans.

---

## Concept 5: Trace/log correlation + health vs readiness

### Motivation

You have three signals — now make them work together, and make the daemon tell the orchestrator how it's doing. Correlation lets you jump from a log line to its trace. Health/readiness let a load balancer route only to instances that can actually serve.

---

### The basics — correlation

Stamp the **trace ID** onto every log line for the request:

```go
reqLogger := logger.With("trace_id", span.SpanContext().TraceID().String(),
    "method", r.Method, "path", r.URL.Path)
reqLogger.Info("request handled", "status", rec.status)
// {"level":"INFO","msg":"request handled","trace_id":"4bf92f...","status":200}
```

Now a log search that finds an error gives you a `trace_id` you paste into your tracing UI to see the whole request — and vice versa. One field turns three siloed signals into one navigable picture.

### The basics — health vs readiness

Two distinct probes (Kubernetes treats them differently):

- **Liveness** (`/healthz`) — "am I alive?" Fail → the orchestrator *restarts* me.
- **Readiness** (`/readyz`) — "can I serve traffic *right now*?" Fail → the orchestrator stops *routing* to me (but doesn't restart). Return 503 during startup warm-up or while draining on shutdown.

Conflating them is a classic outage: a readiness blip that's wired to liveness triggers needless restarts; a liveness check that's really readiness keeps routing to a broken instance.

---

### A worked example

Every `logstatsd` request logs with its `trace_id`, so the JSON logs and the stdout spans share an ID you can grep across. `/healthz` answers liveness; readiness (`/readyz`, a "going further" extension) would return 503 the moment graceful shutdown begins, so the load balancer drains it before the in-flight requests finish (tying back to L26's graceful shutdown).

---

### Common mistake

**Uncorrelated logs** (no trace ID) — you find the error log but can't get to the trace, so you're back to guessing. And **conflating liveness with readiness** — restarting a pod that's merely not-yet-ready, or routing to one that's unhealthy. Correlate the signals; keep the two probes distinct.

---

### Recap

- Stamp the **trace ID** on request logs (`slog.With`) to pivot logs↔traces.
- **Liveness** (`/healthz`) → restart on failure; **readiness** (`/readyz`) → stop routing on failure.
- Correlated signals + distinct probes = an operable service.

---

## Practice

### Warm-up

In `exercises/warmup/tracemw/`, implement `WithTracing(next, tracer)` — start a span named `METHOD path`, thread the ctx into the request, record the response status as a span attribute.

```bash
cd lessons/28-observability/exercises
go test ./warmup/tracemw/...
```

### Main

In `internal/otelx/`, implement `Setup(ctx, w)` — build the `TracerProvider` (stdouttrace) + `MeterProvider` (stdoutmetric) + a flushing `shutdown`. (The `Instrument` middleware, the meters, and the wiring into both faces are provided — study how they fit.)

```bash
cd lessons/28-observability/exercises
go test ./...
make test-race

# see telemetry on stdout: run the daemon, hit it, then SIGTERM it
go run ./cmd/logstatsd -http-addr=127.0.0.1:8080 -grpc-addr=127.0.0.1:9090 &
curl -s -XPOST 127.0.0.1:8080/ingest -d '{"lines":["2026-01-02T15:04:05 INFO ok","bad"]}'; echo
curl -s 127.0.0.1:8080/stats; echo
kill -TERM %1   # spans + metrics flush to stdout
```

---

## Closing thought

Observability is the property that makes everything else operable. Config, graceful shutdown, containers — they get the service running; telemetry is how you keep it running, because production *will* surprise you, and the only question is whether you have the data to understand it at 3am.

OpenTelemetry's bet is decoupling: instrument once against a vendor-neutral API, and choose your backend at deploy time via an exporter. We exported to stdout to keep the lesson self-contained, but the spans, metrics, and correlation you wired are exactly what flows to Jaeger or Prometheus or Datadog in production — change one line in `Setup`, not a single handler. That's the same lesson as L25's transports and L26's config: keep the core clean, and the edges become swappable.

---

## What we learned

- **Three pillars**: logs/metrics/traces, strongest correlated; OTel = API + SDK + exporters.
- **SDK setup**: wire providers+exporters in `main`, set global, **`defer shutdown`** to flush; stdout = no collector.
- **Tracing**: `tracer.Start` returns a ctx-carrying span; thread it so children attach; attributes annotate.
- **Metrics**: counters + histograms at the domain event; keep attributes low-cardinality.
- **Correlation + probes**: trace ID into `slog`; liveness (`/healthz`, restart) vs readiness (`/readyz`, route).

---

## Up next

Lesson 29 — **Distributed patterns & course capstone**. The finale: two `logstatsd` instances and the patterns that make ingestion correct across them — idempotency keys, at-least-once delivery, the outbox pattern — all self-contained (no external broker). We tie Phases 1-4 together: from `go run hello` to an observable, idempotent, distributed service.
