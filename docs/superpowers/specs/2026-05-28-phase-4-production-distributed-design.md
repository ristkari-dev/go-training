# Phase 4 — Production & Distributed (lessons 23-29) Design

**Status:** Approved (brainstorming complete, awaiting per-lesson implementation plans)
**Date:** 2026-05-28
**Owner:** Aki Ristkari
**Master design:** [`2026-05-05-go-course-design.md`](2026-05-05-go-course-design.md)
**Phase 3 design (predecessor):** [`2026-05-22-phase-3-concurrency-design.md`](2026-05-22-phase-3-concurrency-design.md)

## Summary

Phase 4 takes students from "I can write and reason about concurrent Go" (end of Phase 3) to "I can build, observe, and operate a production service." Seven lessons (23-29) covering HTTP servers, resilient HTTP clients, gRPC, configuration & graceful shutdown, containerization, observability, and distributed patterns.

The running example evolves the Phase 3 **log aggregator** into the **`logstats` service** — a push-ingest HTTP/gRPC service. Clients POST log lines; the server parses them with the carried-forward `logparse` package and accumulates per-level counts in shared in-memory state (the `sync.Mutex`-guarded map from L18). `GET /stats` returns the counts; `GET /healthz` reports liveness. Every Phase 4 lesson grows this one service: L23 builds the HTTP API, L24 a resilient client that ships logs to it, L25 a gRPC variant, L26 config + graceful shutdown, L27 containerizes it, L28 instruments it, L29 makes ingestion idempotent and at-least-once across two instances.

Phase 4 is where the course's **zero-third-party-deps** invariant finally relaxes — deliberately and narrowly. The master design always anticipated this ("gRPC packages in lesson 25", "traces (OpenTelemetry)" in lesson 28). Real dependencies are introduced **only in the two lessons that teach them**; everything else stays stdlib.

## Audience and pacing

- Phase 4 students have completed lessons 1-22. They can write idiomatic, concurrent Go: interfaces, errors, generics, goroutines, channels, `sync`, `context`, worker pools, networking, profiling.
- Each lesson: ~90 minutes live + ~90 minutes self-study (unchanged from Phases 2-3).
- 7 lessons total. Roughly half a semester of weekly sessions.
- L25 (gRPC) and L28 (observability) are the heaviest — they introduce new toolchains (protobuf codegen, OpenTelemetry SDK). Expect them to run long.

## Core decisions (locked in during brainstorming)

| Decision | Choice |
|---|---|
| Running example | Evolve the log aggregator into the **`logstats` push-ingest service** |
| Service interaction model | **Ingest API (push)**: clients POST log lines → parse + accumulate shared counts → `GET /stats` |
| Third-party deps | **Introduced only where taught** — gRPC/protobuf (L25), OpenTelemetry (L28); everything else stdlib |
| L29 distributed infra | **Self-contained** — idempotency / at-least-once / outbox taught with stdlib + in-process/HTTP queue; **no external broker** |
| Phase entry | **Phase-level design doc first** (this document), its own PR, then per-lesson plans — mirrors Phases 1-3 |
| Web framework | None — `net/http` only (Gin/Echo/Fiber out of scope, per master design) |
| Carry-forward | Each lesson re-hosts the evolving `logstats` service with import paths rewritten (same as Phases 1-3) |
| Test scaffolding | Skeleton tests in warmup + main (continued) |
| Slide rhythm | 4-5 concepts per lesson (continued) |
| `make test-race` | Continues — the service has concurrent shared state |

### Why each decision

**Running example: evolve the aggregator into `logstats` (vs a fresh service vs all-standalone)**
The aggregator has been the through-line since L16; carrying it into Phase 4 preserves the "watch one system grow" pedagogy that worked across Phases 1-3. The Phase 3 design doc already flagged the path ("HTTP-based aggregator … would extend the aggregator with a `/stats` endpoint"). Reframing batch directory-walking as a push-ingest service is a small conceptual step that naturally exposes every Phase 4 topic: HTTP handlers (ingest/stats), a resilient client (ship logs), gRPC (stream log lines), config (where to bind, log level), containerization (ship the binary), observability (request/ingest rates), and distribution (idempotent ingest across instances).

**Ingest API (push) vs on-demand walk vs file-upload**
A push API is the canonical service shape and the only one that exercises all Phase 4 topics well. "On-demand walk" keeps a batch shape (awkward over gRPC streaming, and a server-side directory path is a poor/ unsafe API). "File upload" is stateless per request — no interesting shared concurrency state and a weak fit for streaming/observability. Push ingest reuses L18's mutex-guarded accumulator as genuine shared state, gives L24's client something to retry, gives L25 a natural client-streaming RPC, and gives L28 meaningful metrics (ingest rate, parse-error rate).

**Third-party deps: introduce only where taught**
You cannot honestly teach gRPC without `google.golang.org/grpc` + `google.golang.org/protobuf`, nor OpenTelemetry without its SDK — a hand-rolled mini-version (the `errgroupx` approach from L20) would teach "how it works" but not "the tool you'll actually use," which defeats the purpose of a *production* phase. The master design sanctioned exactly this. The discipline: deps enter `go.mod` only at L25 and L28; HTTP (L23/24), config (L26), container (L27), and distributed (L29) lessons remain pure stdlib. The single-module repo means `go.mod` gains a few `require` lines once L25 lands; only L25/L28 code imports them. This is itself a teaching moment — "minimal, deliberate dependencies" is a production value.

**L29 self-contained (no external broker)**
The course's defining trait is that everything runs with `go test` / `go run` and no external infrastructure. A real broker (NATS/Redis) + docker-compose would make L29 the only lesson you can't run from a clean checkout. The distributed *patterns* (idempotency keys, at-least-once delivery, the outbox pattern, dedup) are fully teachable with the service talking to itself over HTTP or an in-process queue — the patterns are the lesson, not the broker.

## Per-lesson breakdown

### Lesson 23 — HTTP servers

- **Concepts:** `net/http` server; `http.ServeMux` with Go 1.22+ method+pattern routing (`POST /ingest`, `GET /stats/{level}`); `http.Handler`/`HandlerFunc`; middleware (logging, panic recovery) as handler wrappers; structured logging with `slog` (JSON handler, levels, `With`); JSON request/response with `encoding/json`.
- **Per-lesson example:** Build `cmd/logstats-server` exposing `POST /ingest` (body = log lines; parse with `logparse`, accumulate counts), `GET /stats` (JSON level counts), `GET /healthz`. Wrap handlers in logging + recovery middleware emitting `slog` JSON.
- **Warm-up:** Implement a `middleware` wrapper, e.g. `WithRequestLog(next http.Handler, log *slog.Logger) http.Handler` that logs method/path/status/duration.
- **Main:** `internal/logstats` accumulator (mutex-guarded counts, carried from L18's pattern) + `cmd/logstats-server` HTTP API. Tests use `net/http/httptest`.
- **New imports:** `net/http`, `log/slog`, `net/http/httptest` (test).
- **Running-example contribution:** the aggregator becomes a live service.

### Lesson 24 — HTTP clients & resilience

- **Concepts:** `http.Client` with explicit `Timeout`; per-request `context` deadlines; retry with exponential backoff + jitter; idempotency of retried requests; the circuit-breaker concept (hand-rolled, simple state machine); `http.Transport` connection reuse.
- **Per-lesson example:** `cmd/logstats-client` reads log lines (file/stdin) and ships them to a `logstats-server` `/ingest` endpoint, with timeouts, bounded retries, and ctx cancellation.
- **Warm-up:** Implement `retry(ctx, attempts, func() error) error` with exponential backoff honoring `ctx.Done()`.
- **Main:** resilient `internal/shipper` client + `cmd/logstats-client`. Tests use `httptest` servers that fail N times then succeed.
- **New imports:** none net-new (`net/http`, `context`, `time`).
- **Running-example contribution:** a real client feeds the service.

### Lesson 25 — gRPC

- **Concepts:** protobuf IDL; `protoc` + `protoc-gen-go`/`protoc-gen-go-grpc` codegen (generated code committed); unary RPC (`GetStats`); client-streaming RPC (`Ingest(stream LogLine)`); interceptors (the gRPC analogue of HTTP middleware); when gRPC beats REST.
- **Per-lesson example:** a `LogStats` gRPC service mirroring the HTTP API — `Ingest` client-streaming (ship many lines, get a summary) + `GetStats` unary. `cmd/logstats-grpc-server` + a gRPC client.
- **Warm-up:** define the `.proto` and regenerate; implement the unary `GetStats` handler against the carried accumulator.
- **Main:** `proto/logstats.proto` + committed generated code + gRPC server/client + a logging interceptor.
- **New imports:** **`google.golang.org/grpc`, `google.golang.org/protobuf`** (first third-party deps). `go.mod` gains requires; generated `*.pb.go` committed so `go build` needs no `protoc`. README documents the regen command + plugin install.
- **Running-example contribution:** the service gains a typed, streaming RPC surface.

### Lesson 26 — Configuration, secrets & graceful shutdown

- **Concepts:** config precedence (flags > env > file > defaults); `flag` package; `os.Getenv`; a small JSON/env config loader (no third-party — no Viper); secrets hygiene (never log them; read from env/file); graceful shutdown with `signal.NotifyContext` + `http.Server.Shutdown(ctx)` draining in-flight requests.
- **Per-lesson example:** the server takes `-addr`, `-log-level`, `-config=file.json`, env overrides; SIGINT/SIGTERM triggers a drain-then-exit.
- **Warm-up:** implement `Load(args, env) (Config, error)` with the precedence rules + a table test.
- **Main:** `internal/config` + wire `http.Server.Shutdown` into `cmd/logstats-server`.
- **New imports:** `os/signal` (revisit), `flag`, `encoding/json`.
- **Running-example contribution:** the service becomes configurable and shuts down cleanly.

### Lesson 27 — Build, release & containerization

- **Concepts:** `go build` flags (`-ldflags` for version stamping), build tags, `CGO_ENABLED=0` static binaries, cross-compilation (`GOOS`/`GOARCH`); multi-stage Dockerfile; distroless/scratch base images; image size + security hygiene; `.dockerignore`.
- **Per-lesson example:** a multi-stage `Dockerfile` that builds a static `logstats-server` and ships it on distroless; a `Makefile`/script target for versioned builds.
- **Warm-up:** add a `-version` flag stamped via `-ldflags -X`; a test asserting the default and stamped values.
- **Main:** `Dockerfile` + `.dockerignore` + build script; an integration test that builds the binary (skipped if Docker unavailable) and a unit test for version stamping.
- **New imports:** none (Docker is external tooling, not a Go dep).
- **Running-example contribution:** the service is shippable as a container.

### Lesson 28 — Observability

- **Concepts:** the three pillars (logs, metrics, traces); OpenTelemetry SDK; spans + context propagation across handlers; metrics (request counter, ingest counter, latency histogram); trace/log correlation (trace ID in `slog`); health vs readiness; **stdout exporters** so the lesson runs with no collector.
- **Per-lesson example:** instrument `logstats-server` — every request is a span; ingest/stat counts are metrics; logs carry the trace ID; spans/metrics print to stdout via the OTel stdout exporters.
- **Warm-up:** wrap a handler in a tracing middleware that starts a span and records status.
- **Main:** `internal/otelx` setup (tracer/meter providers + stdout exporters) wired into the server; tests assert spans/metrics are recorded via an in-memory exporter.
- **New imports:** **`go.opentelemetry.io/otel`** + SDK + stdout/stdouttrace/stdoutmetric exporters (second third-party dep set).
- **Running-example contribution:** the service is observable.

### Lesson 29 — Distributed patterns & course capstone

- **Concepts:** at-least-once delivery and why exactly-once is a myth; idempotency keys + server-side dedup; the outbox pattern; retries + dedup working together; a self-contained queue (in-process channel-backed or HTTP between two `logstats` instances) — **no external broker**. Course wrap-up tying Phases 1-4 together.
- **Per-lesson example:** two `logstats` instances — a "forwarder" ships batches with idempotency keys to an "aggregator" instance that dedups, so retried/duplicated batches don't double-count.
- **Warm-up:** implement an idempotency-key dedup store (`Seen(key) bool` with bounded memory) + table test.
- **Main:** idempotent `/ingest` (dedup by key) + a forwarder using an outbox + at-least-once retry; an integration test proving duplicate batches are counted once.
- **New imports:** none (stdlib only).
- **Running-example contribution:** the capstone — a small distributed, idempotent ingest pipeline built entirely from the course's own service.

## File layout per lesson (Phases 2-3 pattern continues)

```
lessons/NN-name/
├── README.md
├── slides/
│   ├── index.html, slides.md, assets/.gitkeep
├── exercises/
│   ├── warmup/<topic>/
│   ├── cmd/<binary>/            (logstats-server, logstats-client, …)
│   ├── proto/                   (L25 only: .proto + generated *.pb.go)
│   └── internal/
│       ├── logparse/            (carried from L14)
│       ├── logstats/            (the accumulator — the Phase 4 core)
│       └── <other-subpkgs>/     (shipper, config, otelx, dedup, …)
└── solutions/  (mirrored)
```

### Running-example carry-forward

Each lesson L24-L29 carries forward the PREVIOUS lesson's `logstats` service and extends it, exactly as the aggregator evolved across L17-L20. `internal/logparse` rides along from L14. By L29 the service has an HTTP API, a resilient client, a gRPC surface, config + graceful shutdown, a container, observability, and idempotent distributed ingest.

## Lesson naming (matches `tools/build-index/main.go` master list)

| # | Topic | Slug |
|---|---|---|
| 23 | HTTP servers | `http-server` |
| 24 | HTTP clients & resilience | `http-client` |
| 25 | gRPC | `grpc` |
| 26 | Config & graceful shutdown | `config` |
| 27 | Build & containerization | `container` |
| 28 | Observability | `observability` |
| 29 | Distributed patterns & capstone | `capstone-final` |

Slugs already match the build-index master list; each per-lesson plan confirms accuracy (precedent: Plan Q fixed L14's slug, Plan Y updated L22's).

## Cross-lesson invariants

- **Skeleton tests in BOTH warmup and main** — continued.
- **Subpackage layout** — no flat `warmup.go`/`main.go` at the lesson root.
- **No `Warmup*` prefix** — subpackages namespace.
- **Slides + README written inline by the controller** — subagents handle code-bearing tasks.
- **`make test-race`** in daily-habits READMEs — the service has shared concurrent state.
- **In-process integration tests** — `net/http/httptest` + `net.Listen("127.0.0.1:0")` ephemeral ports (proven in L21); no fixed ports, no external services.
- **4-5 concept slide decks.**
- **Lesson-local subpackages over carry-forward import** — each lesson copies `logparse`/`logstats`/etc. with rewritten import paths.
- **Conventional Commits + GPG signing.**
- **Third-party deps are committed to `go.mod`/`go.sum`** at L25 and L28; generated protobuf code is committed so `go build` works without `protoc`.

## Dependency-management specifics

- **Single module** (`github.com/ristkari-dev/go-training`). Adding grpc/otel `require`s affects the whole module's `go.mod`; only L25/L28 packages import them. Lessons 1-24, 26, 27, 29 import zero third-party packages.
- **L25 gRPC:** install `protoc` + `protoc-gen-go` + `protoc-gen-go-grpc` is a *developer* step; the generated `*.pb.go` is committed, so students and CI build without `protoc`. The README documents the install + `protoc` regen command and pins plugin versions.
- **L28 OTel:** use the **stdout** trace/metric exporters (`go.opentelemetry.io/otel/exporters/stdout/...`) so the lesson is fully runnable with no collector. Tests use the in-memory exporter (`tracetest`/`metricdata` test helpers).
- **CI:** the existing `build` job runs `make test` over the module; once L25 lands it downloads grpc/protobuf, and once L28 lands it downloads OTel. `golangci-lint` config may need exclusions for generated protobuf code (`.*\.pb\.go`) — confirmed per-lesson.
- **`.golangci.yml`:** add a generated-code exclusion for `*.pb.go` when L25 lands (mechanical; noted in the L25 plan).

## Non-scope (Phase 4)

- Web frameworks (Gin/Echo/Fiber) — `net/http` only.
- Databases / ORMs / connection pooling — out of scope (the accumulator is in-memory by design).
- Real Kubernetes manifests / Helm / cloud deploys — mentioned conceptually only.
- Deep TLS / mTLS / certificate management — mentioned; not built.
- AuthN/AuthZ (OAuth2/OIDC) — out of scope (touched conceptually if at all).
- External message brokers (NATS/Kafka/Redis) — L29 stays self-contained.
- Service mesh, API gateways — out of scope.

## Open issues (to revisit per-lesson)

- **L25 plugin versions:** pin `protoc-gen-go` / `protoc-gen-go-grpc` versions in the README and verify the committed generated code matches; decide whether to add a `make proto` target.
- **L28 exporter choice:** stdout exporters keep it self-contained, but the output is verbose; the lesson may pretty-trim or sample. Per-lesson brainstorming confirms.
- **L27 Docker in CI:** the containerization test must skip gracefully when Docker isn't present (CI may not have a daemon) — assert the Dockerfile builds only when `docker` is available; otherwise unit-test the version stamping path.
- **L29 queue shape:** in-process channel-backed vs HTTP-between-two-instances — the per-lesson brainstorm picks the one that best demonstrates idempotency without external infra.
- **`logstats` accumulator reset semantics:** whether `/stats` is cumulative-since-start or windowed — default cumulative; revisit if L28 metrics want windows.

## Execution after approval

Mirror the Phase 1/2/3 design-PR process:

1. User reviews this design doc → approves or requests changes.
2. PR for the design doc only (not the implementation).
3. Once merged, per-lesson brainstorming begins for L23 (next plan letter after Plan Y → **Plan Z**; subsequent lessons continue the sequence).
4. Each lesson follows the established subagent-driven flow: brainstorm → plan → execute → review → PR → merge.
