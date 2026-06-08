# Lesson 28: Observability

## What you'll learn

By the end of this lesson you can:

- Reason about the **three pillars** (logs, metrics, traces) and the **OpenTelemetry** model (API / SDK / exporters).
- Wire the OTel **SDK** with **stdout exporters** and flush on shutdown — no collector needed.
- Record **spans** and propagate trace context across handlers and RPCs.
- Record **counters** and **histograms**, and keep metric cardinality sane.
- **Correlate** logs with traces (trace ID in `slog`) and distinguish **liveness vs readiness**.

This is the sixth Phase 4 lesson, and the course's **second and final** third-party dependency (OpenTelemetry, after gRPC).

## What's different from L27

L27 made the daemon shippable; L28 makes it **observable**. The `logstatsd` service is carried verbatim and gains telemetry across **both faces**: an HTTP otel middleware, gRPC interceptor spans, request + ingest metrics, and trace-correlated logs — all exported to **stdout** so the lesson runs from a clean checkout with no Jaeger/Prometheus/collector.

## Dependencies (the final boundary shift)

```
go.opentelemetry.io/otel                          v1.38.0
go.opentelemetry.io/otel/sdk                       v1.38.0
go.opentelemetry.io/otel/sdk/metric                v1.38.0
go.opentelemetry.io/otel/exporters/stdout/stdouttrace   v1.38.0
go.opentelemetry.io/otel/exporters/stdout/stdoutmetric  v1.38.0
```

**Why v1.38.0?** It's the newest OpenTelemetry that declares `go 1.23` (later needs go 1.24+), so adding it keeps the course module at `go 1.23`. Only core OTel — no `otelgrpc`/`otelhttp` contrib packages; the gRPC spans are wired manually with `tracer.Start`.

## The package layout

```
lessons/28-observability/{exercises,solutions}/
├── warmup/tracemw/         ← you implement: WithTracing (span per request + status)
├── warmup/{config,buildinfo}/   carried
├── proto/ + logstatspb/    carried
├── internal/
│   ├── logparse, logstats  carried
│   ├── otelx/              ← you implement Setup; Instrument + meters provided
│   ├── httpsrv/            wired to otelx.Instrument + ingested_lines_total
│   └── grpcsrv/            interceptor spans + ingested_lines_total
└── cmd/logstatsd/          calls otelx.Setup at startup; flushes on shutdown
```

---

## Concept 1 — Three pillars + the OTel model

| Pillar | Answers | Example |
|---|---|---|
| Logs | what happened | `"ingest failed: bad line"` |
| Metrics | how much/often | `http_requests_total`, p99 latency |
| Traces | where time went | the span tree for one `/ingest` |

OpenTelemetry = **API** (`go.opentelemetry.io/otel` — what your code calls) + **SDK** (`otel/sdk` — providers/exporters, wired once in `main`) + **exporters** (stdout here; OTLP → Jaeger/Prometheus/Datadog in prod). Instrument once against the API; pick the backend via the exporter.

### Common mistake

Treating the three as separate, uncorrelated tools — during an incident you can't pivot from a slow metric to the trace to the log. OTel emits all three with shared context so you can.

---

## Concept 2 — SDK setup + stdout exporters

```go
func Setup(ctx context.Context, w io.Writer) (func(context.Context) error, error) {
    traceExp, _ := stdouttrace.New(stdouttrace.WithWriter(w))
    tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExp))
    otel.SetTracerProvider(tp)
    metricExp, _ := stdoutmetric.New(stdoutmetric.WithWriter(w))
    mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp)))
    otel.SetMeterProvider(mp)
    return func(ctx context.Context) error { return errors.Join(tp.Shutdown(ctx), mp.Shutdown(ctx)) }, nil
}
```

Providers are set **global** so `otel.Tracer`/`otel.Meter` everywhere return real instruments. The batcher/periodic-reader export in the background; **`Shutdown` forces the final flush**. `logstatsd` calls `Setup(ctx, os.Stdout)` before building the servers and `defer`s the shutdown — on SIGTERM the daemon drains and the spans+metrics flush to stdout.

### Common mistake

Forgetting `Shutdown` — the last (often most interesting) batch is lost on exit. And setting up OTel *after* creating instruments binds them to the no-op provider. Setup first; `defer shutdown`.

---

## Concept 3 — Tracing

```go
ctx, span := tracer.Start(r.Context(), "POST /ingest")
defer span.End()
next.ServeHTTP(w, r.WithContext(ctx))   // ctx carries the span → children attach
span.SetAttributes(attribute.Int("http.status_code", rec.status))
```

A trace is a tree of timed **spans**. `tracer.Start` returns a *new ctx* carrying the span; anything that uses that ctx and starts its own span becomes a child — the tree builds itself. The HTTP middleware spans each request; the gRPC interceptors span each RPC (named by full method) using core OTel.

### Common mistake

Not threading the returned ctx (`_, span := ...` or passing the original ctx downstream) → orphaned spans instead of a tree. Always use and pass the ctx `Start` returns.

---

## Concept 4 — Metrics

```go
reqs, _ := meter.Int64Counter("http_requests_total")
dur, _  := meter.Float64Histogram("http_request_duration_seconds")
reqs.Add(ctx, 1, metric.WithAttributes(attribute.Int("status", status)))
dur.Record(ctx, time.Since(start).Seconds())
```

Counters (monotonic totals → rates) and histograms (distributions → p50/p95/p99). `ingested_lines_total` is recorded where parsing happens — in **both** the HTTP handler and gRPC `Ingest`, into one counter via the global meter — so it's correct regardless of transport (the metric lives at the domain event, not the transport).

### Common mistake

High-cardinality attributes (user ID, request ID, raw path) — each distinct value is a new time series, which explodes the metrics backend. Keep metric attributes low-cardinality (status, method, route template); put high-cardinality detail on **spans**. (Our counter uses `status`, not the raw path — though the span *name* uses the path.)

---

## Concept 5 — Trace/log correlation + health vs readiness

```go
reqLogger := logger.With("trace_id", span.SpanContext().TraceID().String())
reqLogger.Info("request handled", "status", rec.status)
// {"msg":"request handled","trace_id":"4bf92f...","status":200}
```

The trace ID on every request log lets you pivot logs↔traces. And two distinct probes: **liveness** (`/healthz`) → fail means *restart me*; **readiness** (`/readyz`) → fail means *stop routing to me* (return 503 during warm-up or while draining). Conflating them causes needless restarts or routing to broken instances.

### Common mistake

Uncorrelated logs (no trace ID — you find the error but can't reach the trace), or conflating liveness with readiness (restarting a not-yet-ready pod / routing to an unhealthy one).

---

## Exercise: warm-up — `tracemw`

Implement `WithTracing(next http.Handler, tracer trace.Tracer) http.Handler` in `exercises/warmup/tracemw/tracemw.go` — start a span named `METHOD path`, thread the ctx into the request, record the status as a span attribute (a `statusRecorder` is provided).

**Time:** 10-15 minutes.

## Exercise: main — `otelx.Setup`

Implement `Setup(ctx, w)` in `exercises/internal/otelx/otelx.go` — build the `TracerProvider` (stdouttrace) + `MeterProvider` (stdoutmetric) + a flushing `shutdown`. The `Instrument` middleware, the meters, and the wiring into httpsrv/grpcsrv/cmd are provided — study how they compose.

**Time:** 30-45 minutes.

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
# Skeleton (passes vacuously until you implement)
cd lessons/28-observability/exercises
go test ./...

# Reference solution
go test -v ./lessons/28-observability/solutions/...
go test -race ./lessons/28-observability/solutions/...
make test-race

# See telemetry on stdout: run the daemon, hit it, then SIGTERM it
go build -o /tmp/logstatsd ./lessons/28-observability/solutions/cmd/logstatsd
/tmp/logstatsd -http-addr=127.0.0.1:8080 -grpc-addr=127.0.0.1:9090 &
sleep 0.5
curl -s -XPOST 127.0.0.1:8080/ingest -d '{"lines":["2026-01-02T15:04:05 INFO ok","bad"]}' >/dev/null
curl -s 127.0.0.1:8080/stats >/dev/null
kill -TERM %1   # span + metric JSON flushes to stdout (GET /stats, POST /ingest,
                # http_requests_total, ingested_lines_total), logs carry trace_id
rm -f /tmp/logstatsd
```

## Going further

- **OTLP exporter** — swap `stdouttrace`/`stdoutmetric` for `otlptracegrpc`/`otlpmetricgrpc` pointing at a local Jaeger/Tempo + Prometheus (via docker-compose); the instrumentation code doesn't change.
- **Contrib instrumentation** — replace the manual middleware/interceptor with `otelhttp` + `otelgrpc` (auto-instrumentation); compare span detail.
- **Resource attributes** — set `service.name`, `service.version` (from `buildinfo`!) on the providers so telemetry is attributable.
- **Exemplars** — link a histogram bucket to an example trace ID, so a slow-latency data point jumps to the trace.
- **Sampling** — configure `sdktrace.WithSampler` (e.g. `TraceIDRatioBased`) and reason about head vs tail sampling under load.

---

> The finale is next. Lesson 29 — **Distributed patterns & course capstone**: two `logstatsd` instances and the patterns that make ingestion correct across them — idempotency keys, at-least-once delivery, the outbox pattern — self-contained, no external broker. From `go run hello` to an observable, idempotent, distributed service.
